package resolver

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"sync"

	"github.com/rs/zerolog/log"

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
	Value        NullString
	Children     map[string]**Param
	common       bool
	deepResolved bool
	seen         bool
	serialized   []byte
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
	case "application/x-symlink", "application/x-case", "application/x-template":
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
}

func getTreeMTime(ctx context.Context) (string, error) {
	query := "SELECT max(MTime) FROM my_config_tree_log"
	row := DB.QueryRowContext(ctx, query)
	var mtime string
	err := row.Scan(&mtime)
	return mtime, err
}

func selectTree(ctx context.Context) (*Param, error) {
	query := `
		SELECT ID, Name, Path, MTime, Version, ContentType, Value, ParentID
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
		err = rows.Scan(&param.ID, &param.Name, &param.Path, &param.MTime, &param.Version, &param.ContentType, &param.Value, &parentID)
		if err != nil {
			return nil, err
		}
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
