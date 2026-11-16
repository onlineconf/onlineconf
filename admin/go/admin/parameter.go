package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

const (
	maxPathLen      = 512 // defined in admin/etc/schema.sql
	maxSymlinkDepth = 5   // recursion limit for checkDeletedParamSymlinks
)

var (
	ErrAccessDenied    = errors.New("Access denied")
	ErrAlreadyExists   = errors.New("Parameter already exists")
	ErrVersionNotMatch = errors.New("Version not match")
	ErrCommentRequired = errors.New("Comment required")
	ErrInvalidValue    = errors.New("Invalid value")
	ErrInvalidPath     = errors.New("Invalid path")
	ErrNotEmpty        = errors.New("Parameter has children")
	ErrNotFound        = errors.New("Parameter not found")
	ErrParentNotFound  = errors.New("Parent not found")
	ErrDeletedSymlink  = errors.New("Parameter being deleted is symlinked")
	ErrDeletedTmpl     = errors.New("Parameter being deleted is referred from template(s)")
)

const selectFields string = `
	SELECT t.ID, t.Name, t.ParentID, t.Path, t.Value, t.ContentType, t.Encrypted, t.Summary, t.Description, t.Version, t.MTime, t.Deleted,
	(SELECT count(*) FROM my_config_tree c WHERE c.ParentID = t.ID AND NOT c.Deleted) AS NumChildren,
	(SELECT count(*) <> 0 FROM my_config_tree_group g WHERE g.NodeID = t.ID) AS AccessModified,
		my_config_tree_access(t.ID, ?) AS RW,
		my_config_tree_notification(t.ID) AS Notification,
		t.Notification IS NOT NULL AS NotificationModified
`

const selectFromConfig string = selectFields + `
	FROM my_config_tree t
`

type Parameter struct {
	ID                   int           `json:"-"`
	Name                 string        `json:"name"`
	ParentID             sql.NullInt64 `json:"-"`
	Path                 string        `json:"path"`
	Value                NullString    `json:"data"`
	ContentType          string        `json:"mime"`
	Encrypted            bool          `json:"encrypted"`
	CanReveal            bool          `json:"can_reveal,omitempty"`
	Summary              string        `json:"summary"`
	Description          string        `json:"description"`
	Version              int           `json:"version"`
	MTime                string        `json:"mtime"`
	Deleted              bool          `json:"-"`
	NumChildren          int           `json:"num_children"`
	AccessModified       bool          `json:"access_modified"`
	RW                   NullBool      `json:"rw"`
	Notification         string        `json:"notification"`
	NotificationModified bool          `json:"notification_modified"`
	storedValue          NullString    // the stored form of Value, kept aside for a reveal
}

// rowScanner is what sql.Row and sql.Rows have in common.
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanParameter reads one row of selectFields as stored. Every reader but
// the write path's locked read then applies finishScan.
func scanParameter(row rowScanner) (*Parameter, error) {
	p := Parameter{}
	err := row.Scan(
		&p.ID, &p.Name, &p.ParentID, &p.Path, &p.Value, &p.ContentType, &p.Encrypted,
		&p.Summary, &p.Description, &p.Version, &p.MTime, &p.Deleted,
		&p.NumChildren, &p.AccessModified, &p.RW, &p.Notification, &p.NotificationModified,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// finishScan applies the read-side rules to a freshly scanned row: no value
// without read access, never a stored ciphertext — an encrypted value is
// masked (null, marker kept) until an explicit reveal, for every caller —
// and can_reveal from the reveal policy and the caller's access.
func (p *Parameter) finishScan(ctx context.Context) error {
	if !p.RW.Valid {
		p.Value = NullString{}
		return nil
	}
	p.storedValue = p.Value
	if p.Encrypted {
		p.Value = NullString{}
		requireWrite, err := revealRequiresWrite(ctx)
		if err != nil {
			return err
		}
		p.CanReveal = canReveal(requireWrite, p.RW)
	}
	return nil
}

func SelectParameter(ctx context.Context, path string) (*Parameter, error) {
	query := selectFromConfig + "WHERE Path = ? AND NOT Deleted\n"

	p, err := scanParameter(DB.QueryRowContext(ctx, query, Username(ctx), path))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := p.finishScan(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

func SelectParameterFollowingSymlink(ctx context.Context, path string) (*Parameter, error) {
	return selectParameterFollowingSymlink(ctx, path, map[string]bool{})
}

func selectParameterFollowingSymlink(ctx context.Context, path string, seen map[string]bool) (*Parameter, error) {
	if seen[path] {
		return nil, nil
	}
	seen[path] = true
	defer delete(seen, path)

	parameter, err := SelectParameter(ctx, path)
	if err != nil {
		return nil, err
	}
	if parameter != nil {
		return parameter, nil
	}

	realPath := ""
	segments := strings.Split(path, "/")[1:]
	for i, s := range segments {
		realPath += "/" + s
		if i < len(segments)-1 {
			parameter, err = selectParameterResolvingSymlink(ctx, realPath, seen)
		} else {
			parameter, err = selectParameterFollowingSymlink(ctx, realPath, seen)
		}
		if err != nil {
			return nil, err
		}
		if parameter == nil {
			return nil, nil
		}
		realPath = parameter.Path
	}
	return parameter, nil
}

func SelectParameterResolvingSymlink(ctx context.Context, path string) (*Parameter, error) {
	return selectParameterResolvingSymlink(ctx, path, map[string]bool{})
}

func selectParameterResolvingSymlink(ctx context.Context, path string, seen map[string]bool) (*Parameter, error) {
	parameter, err := selectParameterFollowingSymlink(ctx, path, seen)
	if err != nil {
		return nil, err
	}
	if parameter == nil {
		return nil, nil
	}
	if parameter.ContentType == ContentTypeSymlink {
		seen[path] = true
		defer delete(seen, path)
		return selectParameterResolvingSymlink(ctx, parameter.Value.String, seen)
	}
	return parameter, nil
}

type ParameterWithDescendants interface {
	GetParameter() *Parameter
	AddChild(Parameter)
}

func selectChildren(ctx context.Context, params []ParameterWithDescendants) error {
	paramsMap := make(map[int]ParameterWithDescendants, len(params))
	ids := make([]interface{}, 0, len(params))
	for _, param := range params {
		p := param.GetParameter()
		if p.NumChildren != 0 {
			ids = append(ids, p.ID)
			paramsMap[p.ID] = param
		}
	}
	if len(ids) == 0 {
		return nil
	}
	marks := make([]string, len(ids))
	for i := range marks {
		marks[i] = "?"
	}
	marksStr := strings.Join(marks, ", ")
	query := selectFromConfig + `
		WHERE ParentID IN (` + marksStr + `) AND NOT Deleted
		ORDER BY Name
	`
	binds := append([]interface{}{Username(ctx)}, ids...)
	rows, err := DB.QueryContext(ctx, query, binds...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanParameter(rows)
		if err != nil {
			return err
		}
		if err := c.finishScan(ctx); err != nil {
			return err
		}
		parent := paramsMap[int(c.ParentID.Int64)]
		parent.AddChild(*c)
	}
	return rows.Err()
}

type ParameterWithChildren struct {
	Parameter
	Children []Parameter `json:"children"`
}

func (p *ParameterWithChildren) GetParameter() *Parameter {
	return &p.Parameter
}

func (p *ParameterWithChildren) AddChild(child Parameter) {
	p.Children = append(p.Children, child)
}

// SelectChildren fills in the children of the given nodes.
func SelectChildren(ctx context.Context, params []*ParameterWithChildren) error {
	paramsIf := make([]ParameterWithDescendants, 0, len(params))
	for _, p := range params {
		p.Children = []Parameter{}
		paramsIf = append(paramsIf, p)
	}
	return selectChildren(ctx, paramsIf)
}

func (p *Parameter) WithChildren(ctx context.Context) (*ParameterWithChildren, error) {
	pc := ParameterWithChildren{Parameter: *p}
	err := SelectChildren(ctx, []*ParameterWithChildren{&pc})
	if err != nil {
		return nil, err
	}
	return &pc, nil
}

type ParameterWithSubtree struct {
	Parameter
	Children []ParameterWithSubtree `json:"children"`
}

func (p *ParameterWithSubtree) GetParameter() *Parameter {
	return &p.Parameter
}

func (p *ParameterWithSubtree) AddChild(child Parameter) {
	p.Children = append(p.Children, ParameterWithSubtree{Parameter: child})
}

func SelectSubtree(ctx context.Context, params []*ParameterWithSubtree) error {
	paramsIf := make([]ParameterWithDescendants, 0, len(params))
	for _, p := range params {
		p.Children = []ParameterWithSubtree{}
		paramsIf = append(paramsIf, p)
	}
	err := selectChildren(ctx, paramsIf)
	if err != nil {
		return err
	}
	children := make([]*ParameterWithSubtree, 0)
	for _, p := range params {
		for i := range p.Children {
			children = append(children, &p.Children[i])
		}
	}
	if len(children) == 0 {
		return nil
	}
	return SelectSubtree(ctx, children)
}

func (p *Parameter) WithSubtree(ctx context.Context) (*ParameterWithSubtree, error) {
	ps := ParameterWithSubtree{Parameter: *p}
	err := SelectSubtree(ctx, []*ParameterWithSubtree{&ps})
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

func SelectWithChildrenMulti(ctx context.Context, paths []string) (map[string]*ParameterWithChildren, error) {
	if len(paths) == 0 {
		return map[string]*ParameterWithChildren{}, nil
	}
	pathMap := make(map[string]bool, len(paths))
	binds := []interface{}{Username(ctx)}
	for _, path := range paths {
		binds = append(binds, path)
		pathMap[path] = true
	}
	for _, path := range paths {
		binds = append(binds, path)
	}
	marks := make([]string, len(paths))
	for i := range marks {
		marks[i] = "?"
	}
	marksStr := strings.Join(marks, ", ")
	rows, err := DB.QueryContext(ctx, selectFields+`
		FROM (
			SELECT t.*
			FROM my_config_tree t
			WHERE t.Path IN (`+marksStr+`)
			UNION
			SELECT t.*
			FROM my_config_tree t
			JOIN my_config_tree p ON p.ID = t.ParentID
			WHERE p.Path IN (`+marksStr+`)
		) t
		WHERE NOT t.Deleted
		ORDER BY t.Path
	`, binds...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]*ParameterWithChildren)
	for rows.Next() {
		p, err := scanParameter(rows)
		if err != nil {
			return nil, err
		}
		if err := p.finishScan(ctx); err != nil {
			return nil, err
		}
		if pathMap[p.Path] {
			result[p.Path] = &ParameterWithChildren{
				Parameter: *p,
				Children:  []Parameter{},
			}
		}
		if p.Path != "/" {
			parentPath, _, err := splitPath(p.Path)
			if err != nil {
				return nil, err
			}
			if _, ok := result[parentPath]; ok {
				result[parentPath].Children = append(result[parentPath].Children, *p)
			}
		}
	}
	return result, rows.Err()
}

func SearchParameters(ctx context.Context, term string) ([]Parameter, error) {
	like := "%" + likeEscape(term) + "%"
	publicSearch := `Summary LIKE ? OR Description LIKE ?`
	bind := []interface{}{Username(ctx), like, like, like, like, like}
	isASCII := strings.IndexFunc(term, func(r rune) bool { return r > unicode.MaxASCII }) == -1
	if isASCII {
		publicSearch = `Name COLLATE ascii_general_ci LIKE ? OR ` + publicSearch
		bind = append(bind, like, like)
	}
	rows, err := DB.QueryContext(ctx, `
		SELECT * FROM (
			`+selectFromConfig+`
			WHERE NOT Deleted
			AND (`+publicSearch+` OR (NOT Encrypted AND Value COLLATE utf8mb4_unicode_ci LIKE ?))
			ORDER BY Path
		) x
		WHERE RW IS NOT NULL
		OR (`+publicSearch+`)
	`, bind...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]Parameter, 0)
	for rows.Next() {
		p, err := scanParameter(rows)
		if err != nil {
			return nil, err
		}
		if err := p.finishScan(ctx); err != nil {
			return nil, err
		}
		list = append(list, *p)
	}
	return list, rows.Err()
}

func CreateParameter(ctx context.Context, path, contentType string, encrypted bool, value string, optSummary, optDescription, optNotification NullString, comment string) error {
	ctx, sink := withNotifications(ctx)
	var nullValue NullString
	if contentType != ContentTypeNull {
		nullValue.Valid = true
		nullValue.String = value
	}

	if err := validateParameter(ctx, contentType, encrypted, nullValue); err != nil {
		return err
	}

	var notification NullString
	if optNotification.Valid {
		if err := validateNotification(ctx, optNotification.String); err != nil {
			return err
		}
		if optNotification.String != "" {
			notification.Valid = true
			notification.String = optNotification.String
		}
	}

	parentPath, name, err := splitPath(path)
	if err != nil {
		return err
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	parent, err := selectParameterForUpdate(ctx, tx, parentPath)
	if err != nil {
		tx.Rollback()
		if err == ErrNotFound {
			return ErrParentNotFound
		}
		return err
	}

	var deleted bool
	err = tx.QueryRowContext(ctx, "SELECT Deleted FROM my_config_tree WHERE Path = ? FOR UPDATE", path).Scan(&deleted)
	exists := err == nil
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}
	if exists && !deleted {
		tx.Rollback()
		return ErrAlreadyExists
	}
	// the plaintext is validated; from here on only the stored form is
	// written — encrypted after the row locks above (see activeKeyForWrite)
	storedValue, keyID, err := encryptForStorage(ctx, tx, encrypted, nullValue)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		_, err = tx.ExecContext(ctx, "INSERT INTO my_config_tree (ParentID, Name, ContentType, Encrypted, Value, Summary, Description, Notification) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			parent.ID, name, contentType, encrypted, storedValue, optSummary.String, optDescription.String, notification)
	} else {
		// Encrypted is set unconditionally: a resurrected row must not inherit
		// the marker from its previous, unrelated life as a soft-deleted zombie
		fields := "ContentType = ?, Encrypted = ?, Value = ?, "
		bind := []interface{}{contentType, encrypted, storedValue}
		if optSummary.Valid {
			fields += "Summary = ?, "
			bind = append(bind, optSummary.String)
		}
		if optDescription.Valid {
			fields += "Description = ?, "
			bind = append(bind, optDescription.String)
		}
		if optNotification.Valid {
			fields += "Notification = ?, "
			bind = append(bind, notification)
		}
		bind = append(bind, path)
		_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET "+fields+"Version = Version + 1, MTime = now(), Deleted = false WHERE Path = ?", bind...)
	}
	if err == nil && encrypted {
		// a resurrected zombie may have plaintext history to seal
		err = sealHistoryByPath(ctx, tx, path, keyID)
	}
	if err == nil {
		err = LogLastVersion(ctx, tx, path, comment)
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	return commitAndNotify(ctx, tx, sink)
}

// SetParameter writes a new version. The encrypted field is tri-state: on a
// plaintext parameter an absent field keeps it plaintext; on an encrypted
// parameter the field is mandatory — a client that does not know it cannot
// have obtained the plaintext (it is posting back the masked null it
// received), so writing would replace the secret with an empty value.
func SetParameter(ctx context.Context, path string, version int, contentType string, encrypted NullBool, value string, comment string) error {
	ctx, sink := withNotifications(ctx)
	var nullValue NullString
	if contentType != ContentTypeNull {
		nullValue.Valid = true
		nullValue.String = value
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	p, err := selectParameterForUpdate(ctx, tx, path)
	if err != nil {
		tx.Rollback()
		return err
	}
	if p.Version != version {
		tx.Rollback()
		return ErrVersionNotMatch
	}

	var storeEncrypted bool
	switch {
	case encrypted.Valid:
		storeEncrypted = encrypted.Bool
	case p.Encrypted:
		tx.Rollback()
		return ErrEncryptedFieldRequired
	}
	if err := validateParameter(ctx, contentType, storeEncrypted, nullValue); err != nil {
		tx.Rollback()
		return err
	}
	storedValue, keyID, err := encryptForStorage(ctx, tx, storeEncrypted, nullValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Encrypted = ?, Value = ?, Version = Version + 1, MTime = now() WHERE Path = ?", contentType, storeEncrypted, storedValue, path)
	if err == nil && storeEncrypted {
		err = sealHistory(ctx, tx, p.ID, keyID)
	}
	if err == nil {
		err = LogLastVersion(ctx, tx, path, comment)
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	return commitAndNotify(ctx, tx, sink)
}

// RestoreParameter writes a historical version as the parameter's new
// version, server-side: the UI cannot re-post a value it only sees masked,
// and a rollback is not a reveal. The version is identified by its log
// entry, which must have been written at this path — what the history of a
// path lists. With a version given the path holds a live parameter at that
// version, and the entry may be a former occupant's (a node that lived at
// the path before the current one): the caller then needs read access to
// that node and, when the entry holds a secret, the right to reveal it —
// the rights a reveal-and-post would have demanded. Without a version the
// path holds no live parameter, a soft-deleted one or none, and the entry
// brings the parameter back as a create would, under the parent's write
// access. Encryption never downgrades: an encrypted parameter stays
// encrypted whenever the restored value can be. The restored plaintext is
// re-validated (a symlink target may be gone).
func RestoreParameter(ctx context.Context, path string, version *int, entryID int, comment string) error {
	ctx, sink := withNotifications(ctx)
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := restoreParameter(ctx, tx, path, version, entryID, comment); err != nil {
		tx.Rollback()
		return err
	}
	return commitAndNotify(ctx, tx, sink)
}

func restoreParameter(ctx context.Context, tx *sql.Tx, path string, version *int, entryID int, comment string) error {
	var entry struct {
		nodeID      int
		path        string
		version     int
		contentType string
		encrypted   bool
		value       NullString
		deleted     bool
	}
	err := tx.QueryRowContext(ctx, "SELECT NodeID, Path, Version, ContentType, Encrypted, Value, Deleted FROM my_config_tree_log WHERE ID = ?", entryID).
		Scan(&entry.nodeID, &entry.path, &entry.version, &entry.contentType, &entry.encrypted, &entry.value, &entry.deleted)
	if err == sql.ErrNoRows || (err == nil && (entry.path != path || entry.deleted)) {
		return ErrInvalidValue
	} else if err != nil {
		return err
	}
	plaintext, err := decryptStored(entry.encrypted, entry.value)
	if err != nil {
		return fmt.Errorf("%w: version %d: %v", ErrDecryptionFailed, entry.version, err)
	}
	// the version's own marker: only a scalar of an eligible type ever
	// carries one (validateParameter, sealHistory)
	encrypted := entry.encrypted

	if version != nil {
		// a live parameter at the path, at that version
		p, err := selectParameterForUpdate(ctx, tx, path)
		if err != nil {
			return err
		}
		if p.Version != *version {
			return ErrVersionNotMatch
		}
		if entry.nodeID != p.ID {
			// a former occupant's entry
			if err := checkEntryAccess(ctx, tx, entry.nodeID, entry.encrypted); err != nil {
				return err
			}
		}
		// encryption never downgrades: a version of an eligible type restored
		// onto an encrypted parameter is stored encrypted; a null, symlink,
		// case or template version comes back as the plaintext it is
		if p.Encrypted && EncryptionEligible(entry.contentType) && plaintext.Valid {
			encrypted = true
		}
		if err := validateParameter(ctx, entry.contentType, encrypted, plaintext); err != nil {
			return err
		}
		storedValue, keyID, err := encryptForStorage(ctx, tx, encrypted, plaintext)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Encrypted = ?, Value = ?, Version = Version + 1, MTime = now() WHERE Path = ?", entry.contentType, encrypted, storedValue, path); err != nil {
			return err
		}
		if encrypted {
			if err := sealHistory(ctx, tx, p.ID, keyID); err != nil {
				return err
			}
		}
		return LogLastVersion(ctx, tx, path, comment)
	}

	// no live parameter at the path: the entry brings one back as a create
	// would — under the parent's write access, into the soft-deleted row or
	// a new one — with the marker of the version: there is no current value
	// to never-downgrade from, and a resurrected row never inherits the
	// marker of its previous life
	parentPath, name, err := splitPath(path)
	if err != nil {
		return err
	}
	parent, err := selectParameterForUpdate(ctx, tx, parentPath)
	if err != nil {
		if err == ErrNotFound {
			return ErrParentNotFound
		}
		return err
	}
	var deleted bool
	err = tx.QueryRowContext(ctx, "SELECT Deleted FROM my_config_tree WHERE Path = ? FOR UPDATE", path).Scan(&deleted)
	exists := err == nil
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if exists && !deleted {
		return ErrAlreadyExists
	}
	// the entry's own node decides who may bring it back — the rights the
	// history listing needs to show the entry's value, and the reveal right
	// for a secret: a deleted node keeps its access groups, and a node that
	// left the path alive has its own, which is not the parent's
	if err := checkEntryAccess(ctx, tx, entry.nodeID, entry.encrypted); err != nil {
		return err
	}
	if err := validateParameter(ctx, entry.contentType, encrypted, plaintext); err != nil {
		return err
	}
	storedValue, keyID, err := encryptForStorage(ctx, tx, encrypted, plaintext)
	if err != nil {
		return err
	}
	if exists {
		_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Encrypted = ?, Value = ?, Version = Version + 1, MTime = now(), Deleted = false WHERE Path = ?", entry.contentType, encrypted, storedValue, path)
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO my_config_tree (ParentID, Name, ContentType, Encrypted, Value) VALUES (?, ?, ?, ?, ?)", parent.ID, name, entry.contentType, encrypted, storedValue)
	}
	if err != nil {
		return err
	}
	if encrypted {
		// a resurrected row may have plaintext history to seal
		if err := sealHistoryByPath(ctx, tx, path, keyID); err != nil {
			return err
		}
	}
	return LogLastVersion(ctx, tx, path, comment)
}

// checkEntryAccess checks the caller's rights on a history entry of a node
// other than the one being written — a former occupant of the path, or a
// deleted node being brought back: read access to that node, and when the
// entry holds a secret the right to reveal it under the policy.
func checkEntryAccess(ctx context.Context, tx *sql.Tx, nodeID int, encrypted bool) error {
	var rw NullBool
	if err := tx.QueryRowContext(ctx, "SELECT my_config_tree_access(?, ?)", nodeID, Username(ctx)).Scan(&rw); err != nil {
		return err
	}
	if !rw.Valid {
		return ErrAccessDenied
	}
	if encrypted {
		requireWrite, err := revealRequiresWrite(ctx)
		if err != nil {
			return err
		}
		if !canReveal(requireWrite, rw) {
			return ErrRevealDenied
		}
	}
	return nil
}

func MoveParameter(ctx context.Context, path string, newPath string, symlink bool, version int, comment string) error {
	ctx, sink := withNotifications(ctx)
	newParentPath, newName, err := splitPath(newPath)
	if err != nil {
		return err
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	p, err := selectParameterForUpdate(ctx, tx, path)
	if err != nil {
		tx.Rollback()
		return err
	}

	newParent, err := selectParameterForUpdate(ctx, tx, newParentPath)
	if err != nil {
		tx.Rollback()
		if err == ErrNotFound {
			return ErrParentNotFound
		}
		return err
	}
	if p.Version != version {
		tx.Rollback()
		return ErrVersionNotMatch
	}

	if err = freeDeletedPath(ctx, tx, newPath); err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ParentID = ?, Name = ?, Version = Version + 1, MTime = now() WHERE Path = ?", newParent.ID, newName, path)
	if err == nil {
		likePath := likeEscape(p.Path) + "/%"
		_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Path = '' WHERE Path LIKE ?", likePath)
	}
	if err == nil {
		// the rewrite above re-stamps descendants' paths via the trigger but
		// bumps no version; bump the live ones so each can be logged at its new
		// path (see LogMovedDescendants)
		_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Version = Version + 1, MTime = now() WHERE Path LIKE ? AND NOT Deleted", likeEscape(newPath)+"/%")
	}

	moveComment := fmt.Sprintf("Moved from %s. %s", path, comment)
	if err == nil {
		err = LogLastVersion(ctx, tx, newPath, moveComment)
	}
	if err == nil {
		err = LogMovedDescendants(ctx, tx, newPath, moveComment)
	}

	if err == nil && symlink {
		var nullNotification NullString
		if p.NotificationModified {
			nullNotification.Valid = true
			nullNotification.String = p.Notification
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO my_config_tree (ParentID, Name, ContentType, Value, Notification) VALUES (?, ?, ?, ?, ?)",
			p.ParentID, p.Name, ContentTypeSymlink, newPath, nullNotification)
		if err == nil {
			err = LogLastVersion(ctx, tx, path, fmt.Sprintf("Moved to %s. %s", newPath, comment))
		}
	}

	if err != nil {
		tx.Rollback()
		return err
	}
	return commitAndNotify(ctx, tx, sink)
}

// freeDeletedPath frees up a Path occupied by a soft-deleted row so it can be
// reused by another node moved there. Soft-deletion keeps the row with its
// original Path intact, and Path has a UNIQUE constraint, so without this any
// move whose destination matches a previously-deleted node would fail with a
// duplicate-key error. The row's Name is suffixed with its ID (guaranteed
// unique) and the my_config_tree_move trigger recomputes Path; descendants
// are then forced to recompute their paths through the same trigger by
// setting their Path to ”. If the path is occupied by a live node,
// ErrAlreadyExists is returned.
func freeDeletedPath(ctx context.Context, tx *sql.Tx, path string) error {
	row := tx.QueryRowContext(ctx, "SELECT ID, Deleted FROM my_config_tree WHERE Path = ? FOR UPDATE", path)
	var id int
	var deleted bool
	err := row.Scan(&id, &deleted)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	if !deleted {
		return ErrAlreadyExists
	}
	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Name = CONCAT(Name, '#', ID) WHERE ID = ?", id)
	if err != nil {
		return err
	}
	likePath := likeEscape(path) + "/%"
	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Path = '' WHERE Path LIKE ?", likePath)
	return err
}

func SetParameterDescription(ctx context.Context, path string, summary, description string) error {
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = selectParameterForUpdate(ctx, tx, path)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Summary = ?, Description = ? WHERE Path = ?", summary, description, path)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func SetParameterNotification(ctx context.Context, path string, notification string) error {
	if err := validateNotification(ctx, notification); err != nil {
		return err
	}

	var nullNotification NullString
	if notification != "" {
		nullNotification.Valid = true
		nullNotification.String = notification
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = selectParameterForUpdate(ctx, tx, path)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Notification = ? WHERE Path = ?", nullNotification, path)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func DeleteParameter(ctx context.Context, path string, version int, comment string) error {
	ctx, sink := withNotifications(ctx)
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	p, err := selectParameterForUpdate(ctx, tx, path)
	if err != nil {
		tx.Rollback()
		return err
	}
	if p.Version != version {
		tx.Rollback()
		return ErrVersionNotMatch
	}

	row := tx.QueryRowContext(ctx, "SELECT count(*) FROM my_config_tree WHERE ParentID = ? AND Deleted = false", p.ID)
	var count int
	err = row.Scan(&count)
	if err != nil {
		tx.Rollback()
		return err
	} else if count != 0 {
		tx.Rollback()
		return ErrNotEmpty
	}

	if err = checkDeletedParamSymlinks(ctx, tx, p); err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Deleted = true, Version = Version + 1, MTime = now() WHERE Path = ?", path)
	if err == nil {
		err = ClearAccess(ctx, tx, p.ID)
	}
	if err == nil {
		err = LogLastVersion(ctx, tx, path, comment)
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	return commitAndNotify(ctx, tx, sink)
}

type symlink struct {
	from string // my_config_tree.Path
	to   string // my_config_tree.Value
}

func checkDeletedParamSymlinks(ctx context.Context, tx *sql.Tx, param *Parameter) error {
	if param.Path == "/" {
		return nil
	}

	if disabled, err := deletedSymlinksCheckDisabled.enabled(ctx); err != nil {
		return err
	} else if disabled {
		return nil
	}

	return checkDeletedParamSymlinksRecursive(ctx, tx, param.Path, isSymlink(param), 0)
}

func isSymlink(param *Parameter) bool {
	switch param.ContentType {
	case ContentTypeSymlink:
		return true
	case ContentTypeCase:
		var branches []CaseBranch
		if err := json.Unmarshal([]byte(param.Value.String), &branches); err != nil {
			return false
		}
		for _, b := range branches {
			if b.ContentType == ContentTypeSymlink {
				return true
			}
		}
	}

	return false
}

func checkDeletedParamSymlinksRecursive(ctx context.Context, tx *sql.Tx, linkTarget string, isSymlink bool, depth int) error {
	// isSymlink means "is path being deleted a symlink", so it must be set to false when this func is called recursively
	query, targets := formatSymlinkCheckQuery(linkTarget, isSymlink)

	rows, err := tx.QueryContext(ctx, query, targets...)
	if err != nil {
		return err
	}

	defer rows.Close()

	parentLinks := make([]symlink, 0, len(targets)/2) // symlinks to one of parent subdirs
	symlinked := make([]string, 0, len(targets)/2)    // symlinks to the target's exact path
	expanded := make([]string, 0, len(targets)/2)     // expanded in a template

	for rows.Next() {
		var link symlink

		if err = rows.Scan(&link.from, &link.to); err != nil {
			return err
		}

		if link.to == "" { // template (see sqlSymlinksCheck)
			expanded = append(expanded, link.from)
		} else if link.to == linkTarget || len(link.to) > len(linkTarget) { // exact symlink match, or the nested symlink is used (sqlSymlinksCheck1s/sqlSymlinksCheck2s)
			symlinked = append(symlinked, link.from)
		} else { // symlink to a parent folder
			parentLinks = append(parentLinks, link)
		}
	}

	// symlinked and expanded alices contain unique strings because of UNION
	if len(symlinked) != 0 {
		return fmt.Errorf("%w: %s", ErrDeletedSymlink, strings.Join(symlinked, ", "))
	}

	if len(expanded) != 0 {
		return fmt.Errorf("%w: %s", ErrDeletedTmpl, strings.Join(expanded, ", "))
	}

	if len(parentLinks) == 0 {
		return nil
	}

	// check parentLinks recursively

	depth++
	if depth >= maxSymlinkDepth {
		return nil
	}

	rows.Close() // rows from the tx should be closed when calling this func recursively

	for _, link := range parentLinks {
		recursiveLink := link.from + linkTarget[len(link.to):]
		if len(recursiveLink) > maxPathLen {
			continue
		}

		if err = checkDeletedParamSymlinksRecursive(ctx, tx, recursiveLink, false, depth); err != nil {
			return err
		}
	}

	return nil
}

// we can use Value == ” for a template since symlink's Values are always non-empty because of IN conditions
const sqlSymlinksCheck1 = `
    SELECT Path, Value
      FROM my_config_tree
     WHERE ContentType = '` + ContentTypeSymlink + `'
       AND NOT Deleted
       AND (Value IN (`

// sqlSymlinksCheck1s and sqlCheckSymlinks2s are used when the value being deleted is a symlink itself
const sqlSymlinksCheck1s = `) OR Value LIKE ?`

const sqlSymlinksCheck2 = `)
    UNION
    SELECT t.Path, j.value
      FROM my_config_tree AS t
      JOIN JSON_TABLE(t.Value, "$[*]" COLUMNS (mime VARCHAR(32) PATH "$.mime", value VARCHAR(512) PATH "$.value")) AS j
     WHERE t.ContentType='` + ContentTypeCase + `'
       AND NOT t.Deleted
       AND j.mime='` + ContentTypeSymlink + `'
       AND (j.value IN (`

const sqlSymlinksCheck2s = `) OR j.value LIKE ?`

const sqlSymlinksCheck3 = `)
    UNION
    SELECT Path, ''
      FROM my_config_tree
     WHERE ContentType = '` + ContentTypeTemplate + `'
       AND NOT Deleted
       AND Value LIKE ?
    UNION
    SELECT t.Path, ''
      FROM my_config_tree AS t
      JOIN JSON_TABLE(t.Value, "$[*]" COLUMNS (mime VARCHAR(32) PATH "$.mime", value VARCHAR(512) PATH "$.value")) AS j
     WHERE t.ContentType='` + ContentTypeCase + `'
       AND NOT t.Deleted
       AND j.mime='` + ContentTypeTemplate + `'
       AND j.value LIKE ?
`

func formatSymlinkCheckQuery(target string, isSymlink bool) (query string, params []any) {
	targetSl := strings.Split(target, "/")
	targets := make([]any, 0, 2*len(targetSl)+2) // len(targetSl) == target components + an empty str
	parent := &strings.Builder{}
	placeholders := &strings.Builder{}

	for _, t := range targetSl {
		if t == "" {
			continue
		}

		parent.WriteRune('/')
		parent.WriteString(t)
		targets = append(targets, parent.String())

		if placeholders.Len() == 0 {
			placeholders.WriteRune('?')
		} else {
			placeholders.WriteString(",?")
		}
	}

	if isSymlink {
		targets = append(targets, target+"/%")
	}

	likeTmpl := "%${" + target + "}%"
	targets = append(targets, targets...)         // x-symlink and x-case parameters
	targets = append(targets, likeTmpl, likeTmpl) // x-template and x-case parameters
	pl := placeholders.String()

	if isSymlink {
		return sqlSymlinksCheck1 + pl + sqlSymlinksCheck1s + sqlSymlinksCheck2 + pl + sqlSymlinksCheck2s + sqlSymlinksCheck3, targets
	}

	return sqlSymlinksCheck1 + pl + ")" + sqlSymlinksCheck2 + pl + ")" + sqlSymlinksCheck3, targets
}

// selectParameterForUpdate locks the row for a write and returns it as
// stored (no masking: the write path compares and re-encrypts stored
// values).
func selectParameterForUpdate(ctx context.Context, tx *sql.Tx, path string) (*Parameter, error) {
	p, err := scanParameter(tx.QueryRowContext(ctx, selectFromConfig+"WHERE Path = ? AND NOT Deleted FOR UPDATE", Username(ctx), path))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !(p.RW.Valid && p.RW.Bool) {
		return nil, ErrAccessDenied
	}
	return p, nil
}

// encryptForStorage turns a plaintext value into its stored form: an envelope
// under the key in use when the value is to be encrypted, the value itself
// otherwise. It is the one place every encrypted write passes — a create, a
// set, a restore — so the feature switch gates it here: while the switch is
// off, or no keyring is loaded, nothing encrypted is written (a restore
// whose result would hold a secret is refused like any other encrypted
// write), while reveals and declassification stay unaffected. The key is
// read in the write's transaction, after its row locks (activeKeyForWrite),
// and returned for the sealing of the history that follows the write.
func encryptForStorage(ctx context.Context, tx *sql.Tx, encrypted bool, value NullString) (stored NullString, keyID string, err error) {
	if !encrypted || !value.Valid {
		return value, "", nil
	}
	if enabled, err := isValueEncryptionEnabled(ctx); err != nil {
		return value, "", err
	} else if !enabled {
		return value, "", ErrEncryptionDisabled
	}
	keyID, err = activeKeyForWrite(ctx, tx)
	if err != nil {
		return value, "", err
	}
	envelope, err := EncryptWith(keyID, value.String)
	if err != nil {
		return value, "", err
	}
	return NullStringOf(envelope), keyID, nil
}

// decryptStored turns a stored value into its plaintext: an encrypted value
// is opened with the key its envelope names, and a value marked encrypted
// that is not an envelope is a DecryptionError, never plaintext.
func decryptStored(encrypted bool, value NullString) (NullString, error) {
	if !encrypted {
		return value, nil
	}
	plaintext, err := Decrypt(value.String)
	if err != nil {
		return value, err
	}
	return NullStringOf(plaintext), nil
}

// validateParameter checks a plaintext value against its content type and
// marker — before encryption, which is the only place the structure of a
// secret can be checked.
func validateParameter(ctx context.Context, contentType string, encrypted bool, value NullString) error {
	if encrypted && !EncryptionEligible(contentType) {
		return ErrEncryptionNotAllowed
	}
	switch contentType {
	case ContentTypeNull:
		if value.Valid {
			return ErrInvalidValue
		} else {
			return nil
		}
	case ContentTypeText, ContentTypeList, ContentTypeServer, ContentTypeServer2, ContentTypeTemplate:
		if value.Valid {
			return nil
		} else {
			return ErrInvalidValue
		}
	case ContentTypeJSON:
		if !value.Valid {
			return ErrInvalidValue
		}
		// the decoder's error is not returned: it quotes the value, which
		// may be a secret, and a 500 would log it
		var i interface{}
		if err := json.Unmarshal([]byte(value.String), &i); err != nil {
			return ErrInvalidValue
		}
		return nil
	case ContentTypeYAML:
		if !value.Valid {
			return ErrInvalidValue
		}
		var i interface{}
		if err := yaml.Unmarshal([]byte(value.String), &i); err != nil {
			return ErrInvalidValue
		}
		return nil
	case ContentTypeSymlink:
		if !value.Valid {
			return ErrInvalidValue
		}
		target, err := SelectParameterFollowingSymlink(ctx, value.String)
		if err != nil {
			return err
		} else if target == nil {
			return ErrInvalidValue
		} else {
			return nil
		}
	case ContentTypeCase:
		if !value.Valid {
			return ErrInvalidValue
		}
		var cases []CaseBranch
		if err := json.Unmarshal([]byte(value.String), &cases); err != nil {
			return err
		}
		def := 0
		for _, c := range cases {
			i := 0
			for _, cond := range []NullString{c.Server, c.Group, c.Datacenter, c.Service} {
				if cond.Valid {
					i++
					if cond.String == "" {
						return ErrInvalidValue
					}
				}
			}
			if i == 0 {
				def++
			} else if i > 1 {
				return ErrInvalidValue
			}
			// a branch holds no secret: it reaches one through a symlink to
			// an encrypted parameter, and a stray marker on it is an unknown
			// key like any other
			if err := validateParameter(ctx, c.ContentType, false, c.Value); err != nil {
				return err
			}
		}
		if def > 1 {
			return ErrInvalidValue
		}
		return nil
	default:
		return ErrInvalidValue
	}
}

func validateNotification(ctx context.Context, notification string) error {
	if notification == "none" {
		root, err := UserIsRoot(ctx, Username(ctx))
		if err != nil {
			return err
		} else if !root {
			return ErrAccessDenied
		}
	}
	return nil
}
