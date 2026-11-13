package resolver

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var (
	ErrParentNotFound = errors.New("Parent not found")
	ErrDuplicateRoot  = errors.New("Duplicate root detected")
)

type Param struct {
	ID           int
	Name         string
	Path         string
	MTime        string
	Version      int
	ContentType  string
	Encrypted    bool // holds (or, after resolution, carries) a secret: never log its value
	Value        NullString
	Children     map[string]**Param
	common       bool
	deepResolved bool
	seen         bool
	serialized   []byte
}

func (node *Param) Strings() ([]string, error) {
	switch node.ContentType {
	case ContentTypeNull:
		return nil, nil
	case ContentTypeList, ContentTypeText:
		untrimmed := strings.Split(node.Value.String, ",")
		list := make([]string, 0, len(untrimmed))
		for _, item := range untrimmed {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				list = append(list, trimmed)
			}
		}
		return list, nil
	case ContentTypeJSON:
		var list []string
		err := json.Unmarshal([]byte(node.Value.String), &list)
		return list, err
	case ContentTypeYAML:
		var list []string
		err := yaml.Unmarshal([]byte(node.Value.String), &list)
		return list, err
	default:
		return nil, errors.New("Can't read []string from " + node.ContentType)
	}
}

func (node *Param) GetStruct(ptr interface{}) error {
	switch node.ContentType {
	case ContentTypeYAML:
		return yaml.Unmarshal([]byte(node.Value.String), ptr)
	case ContentTypeJSON:
		return json.Unmarshal([]byte(node.Value.String), ptr)
	default: // ignore any other type
		return nil
	}
}

// logValue is what a log statement may print of a parameter's value.
func (node *Param) logValue() string {
	if node.Encrypted {
		return "<encrypted>"
	}
	return node.Value.String
}

func (node *Param) deepMarkCommon(ctx context.Context) {
	ok := true
	for _, childPtr := range node.Children {
		child := *childPtr
		child.deepMarkCommon(ctx)
		if !child.common {
			ok = false
		}
	}
	switch node.ContentType {
	case ContentTypeSymlink, ContentTypeCase, ContentTypeTemplate:
		ok = false
	}
	node.common = ok
	node.deepResolved = ok
	if ok {
		if err := serializeParam(node); err != nil {
			log.Ctx(ctx).Warn().Str("param", node.Path).Msg("failed to serialize common param")
		}
	}
}

func (node *Param) deepCreateChildList(ctx context.Context) {
	if len(node.Children) == 0 {
		return
	}

	childrenNames := make([]string, 0, len(node.Children))
	for name, childPtr := range node.Children {
		childrenNames = append(childrenNames, name)
		(*childPtr).deepCreateChildList(ctx)
	}

	sort.Strings(childrenNames)

	chilrenNamesJson, err := json.Marshal(childrenNames)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("error encoding children names list") // should not happen
	}

	listPath := node.Path
	if !strings.HasSuffix(listPath, "/") {
		listPath += "/"
	}

	param := &Param{
		ID:          -node.ID,
		Name:        "", // Name field is used by perl legacy updater only
		Path:        listPath,
		Version:     1,
		ContentType: ContentTypeJSON,
		Value: NullString{NullString: sql.NullString{
			String: string(chilrenNamesJson),
			Valid:  true,
		}},
	}

	node.Children[""] = &param
}

func (node *Param) deepClone() *Param {
	if node.deepResolved {
		return node
	}
	copy := *node
	copy.Children = make(map[string]**Param)
	for name, childPtr := range node.Children {
		childCopy := (*childPtr).deepClone()
		copy.Children[name] = &childCopy
	}
	return &copy
}

type tree struct {
	root         *Param
	mtime        string
	rw           sync.RWMutex
	ephemeralIPs []net.IPNet
	datacenters  []datacenter
	groups       []group
	services     map[string][]byte
	synchronized []synchronizedEntry
	syncMutex    sync.Mutex
}

// treeLogMarginSeconds keeps the change log watermark behind the clock. A
// write is stamped at its UPDATE and committed later, so a write committed
// after a reload with a stamp that reload already covered would never make
// max(MTime) grow and would stay out of the tree until the next unrelated
// write. Leaving the newest seconds out of the watermark lets the next tick
// see it, at the price of a change reaching the tree up to the margin
// later; a transaction longer than the margin can still slip through when
// another write lands meanwhile.
// TODO: the real fix is a tree version counter incremented inside every
// write transaction, which makes this check exact.
const treeLogMarginSeconds = 5

// getTreeMTime returns the watermark: the newest MTime older than the
// margin, or "" when there is none yet.
func getTreeMTime(ctx context.Context) (string, error) {
	query := fmt.Sprintf("SELECT max(MTime) FROM my_config_tree_log WHERE MTime < now() - interval %d second", treeLogMarginSeconds)
	row := DB.QueryRowContext(ctx, query)
	var mtime sql.NullString
	err := row.Scan(&mtime)
	return mtime.String, err
}

func selectTree(ctx context.Context) (*Param, error) {
	query := `
		SELECT ID, Name, Path, MTime, Version, ContentType, Encrypted, Value, ParentID
		FROM my_config_tree
		WHERE NOT Deleted
		ORDER BY Path
	`
	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paramByID := make(map[int]*Param)
	var root *Param
	for rows.Next() {
		param := Param{}
		param.Children = make(map[string]**Param)
		var parentID sql.NullInt64
		err = rows.Scan(&param.ID, &param.Name, &param.Path, &param.MTime, &param.Version, &param.ContentType, &param.Encrypted, &param.Value, &parentID)
		if err != nil {
			return nil, err
		}
		// the resolver works on plaintext: decrypt at load, and fail closed
		// on a value the keyring cannot open (the caller keeps the last good
		// tree, or refuses to start)
		param.Value, err = DecryptStoredValue(param.ContentType, param.Encrypted, param.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", param.Path, err)
		}
		// a case with an encrypted branch is as secret as its branches
		param.Encrypted = HasEncryptedValue(param.ContentType, param.Encrypted, param.Value)
		paramByID[param.ID] = &param
		if parentID.Valid {
			parent, ok := paramByID[int(parentID.Int64)]
			if !ok {
				log.Error().Str("path", param.Path).Int64("parent_id", parentID.Int64).Msg("parent not found")
				return nil, ErrParentNotFound
			}
			paramPtr := &param
			parent.Children[param.Name] = &paramPtr
		} else {
			if root != nil {
				log.Error().Str("path", param.Path).Msg("duplicate root detected")
				return nil, ErrDuplicateRoot
			}
			root = &param
		}
	}

	root.deepCreateChildList(ctx)
	root.deepMarkCommon(ctx)

	return root, nil
}

func (t *tree) update(ctx context.Context) error {
	isReadyToUpdate, mtime, err := t.readyToUpdate(ctx)
	if err != nil {
		return err
	}
	if !isReadyToUpdate {
		return nil
	}
	log.Ctx(ctx).Info().Str("mtime", mtime).Msg("updating graph")

	var root *Param
	var ephemeralIPs []net.IPNet
	var datacenters []datacenter
	var groups []group
	var services map[string][]byte

	root, err = selectTree(ctx)
	if err != nil {
		return err
	}

	cg := newCommonGraph(root.deepClone())
	if ephemeralIPs, err = cg.readEphemeralIPs(ctx); err != nil {
		return err
	}
	if datacenters, err = cg.readDatacenters(ctx); err != nil {
		return err
	}
	if groups, err = cg.readGroups(ctx); err != nil {
		return err
	}
	if services, err = cg.readServices(ctx); err != nil {
		return err
	}

	t.syncMutex.Lock()
	for _, entry := range t.synchronized {
		cg.synchronize(ctx, entry.path, entry.target)
	}
	t.syncMutex.Unlock()

	t.rw.Lock()
	t.root = root
	t.ephemeralIPs = ephemeralIPs
	t.datacenters = datacenters
	t.groups = groups
	t.services = services
	t.mtime = mtime
	t.rw.Unlock()
	return nil
}

func (t *tree) readyToUpdate(ctx context.Context) (res bool, mtime string, err error) {
	if mtime, err = getTreeMTime(ctx); err != nil {
		return false, mtime, err
	}
	t.rw.RLock()
	res = mtime > t.mtime
	t.rw.RUnlock()
	return
}

func (t *tree) getServicePwdHash(username string) (pwdHash []byte, exists bool) {
	t.rw.RLock()
	pwdHash, exists = t.services[username]
	t.rw.RUnlock()
	return
}

func (t *tree) getEphemeralIPs() []net.IPNet {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.ephemeralIPs
}

type Synchronized interface {
	Update(context.Context, *Param)
}

type synchronizedEntry struct {
	path   string
	target Synchronized
}

func (t *tree) synchronize(ctx context.Context, path string, target Synchronized) {
	t.syncMutex.Lock()
	t.synchronized = append(t.synchronized, synchronizedEntry{path: path, target: target})
	t.syncMutex.Unlock()

	t.rw.RLock()
	defer t.rw.RUnlock()
	if t.root != nil {
		cg := newCommonGraph(t.root.deepClone())
		cg.synchronize(ctx, path, target)
	}
}
