package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
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
	ValueFlags
	ID                   int           `json:"-"`
	Name                 string        `json:"name"`
	ParentID             sql.NullInt64 `json:"-"`
	Path                 string        `json:"path"`
	Value                NullString    `json:"data"`
	ContentType          string        `json:"mime"`
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
// without read access, never a stored ciphertext — encrypted leaves are
// masked (null, marker kept) until an explicit reveal, for every caller —
// and can_reveal from the reveal policy and the caller's access.
func (p *Parameter) finishScan(ctx context.Context) error {
	if !p.RW.Valid {
		p.Value = NullString{}
		return nil
	}
	p.storedValue = p.Value
	p.Value = MaskStoredValue(p.ContentType, p.Encrypted, p.Value)
	if HasEncryptedValue(p.ContentType, p.Encrypted, p.Value) {
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

// parameters lists the node and its children, for read-side post-processing.
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
			AND (`+publicSearch+` OR Value COLLATE utf8mb4_unicode_ci LIKE ?)
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

func CreateParameter(ctx context.Context, path, contentType string, flags ValueFlags, value string, optSummary, optDescription, optNotification NullString, comment string) error {
	ctx, sink := withNotifications(ctx)
	var nullValue NullString
	if contentType != ContentTypeNull {
		nullValue.Valid = true
		nullValue.String = value
	}

	if err := validateParameter(ctx, contentType, flags, nullValue); err != nil {
		return err
	}
	// the plaintext is validated; from here on only the stored form is written
	storedValue, err := encryptForStorage(ctx, contentType, flags.Encrypted, nullValue)
	if err != nil {
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

	row := tx.QueryRowContext(ctx, "SELECT Deleted FROM my_config_tree WHERE Path = ? FOR UPDATE", path)
	var deleted bool
	err = row.Scan(&deleted)
	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx, "INSERT INTO my_config_tree (ParentID, Name, ContentType, Encrypted, Value, Summary, Description, Notification) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			parent.ID, name, contentType, flags.Encrypted, storedValue, optSummary.String, optDescription.String, notification)
	} else if err == nil {
		if !deleted {
			tx.Rollback()
			return ErrAlreadyExists
		}
		// Encrypted is set unconditionally: a resurrected row must not inherit
		// the marker from its previous, unrelated life as a soft-deleted zombie
		fields := "ContentType = ?, Encrypted = ?, Value = ?, "
		bind := []interface{}{contentType, flags.Encrypted, storedValue}
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
	if err == nil && HasEncryptedValue(contentType, flags.Encrypted, nullValue) {
		// a resurrected zombie may have plaintext history to seal
		err = sealHistoryByPath(ctx, tx, path)
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

	var flags ValueFlags
	switch {
	case encrypted.Valid:
		flags.Encrypted = encrypted.Bool
	case p.Encrypted:
		tx.Rollback()
		return ErrEncryptedFieldRequired
	}
	if err := validateParameter(ctx, contentType, flags, nullValue); err != nil {
		tx.Rollback()
		return err
	}
	storedValue, err := encryptForStorage(ctx, contentType, flags.Encrypted, nullValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Encrypted = ?, Value = ?, Version = Version + 1, MTime = now() WHERE Path = ?", contentType, flags.Encrypted, storedValue, path)
	if err == nil && HasEncryptedValue(contentType, flags.Encrypted, nullValue) {
		err = sealHistory(ctx, tx, p.ID)
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

// RestoreParameter writes a historical version of the parameter as its new
// version, server-side: the UI cannot re-post a value it only sees masked,
// and a rollback is not a reveal. Encryption never downgrades at parameter
// level — a parameter holding any encrypted value keeps holding one — and
// the restored plaintext is re-validated (a symlink target may be gone).
func RestoreParameter(ctx context.Context, path string, version, restoreVersion int, comment string) error {
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

	var contentType string
	var encrypted, deleted bool
	var stored NullString
	row := tx.QueryRowContext(ctx, "SELECT ContentType, Encrypted, Value, Deleted FROM my_config_tree_log WHERE NodeID = ? AND Version = ?", p.ID, restoreVersion)
	err = row.Scan(&contentType, &encrypted, &stored, &deleted)
	if err == sql.ErrNoRows || (err == nil && deleted) {
		tx.Rollback()
		return ErrInvalidValue
	} else if err != nil {
		tx.Rollback()
		return err
	}
	plaintext, err := DecryptStoredValue(contentType, encrypted, stored)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%w: version %d: %v", ErrDecryptionFailed, restoreVersion, err)
	}
	// a sealed history row carries the marker whatever its type (see sealHistory); the
	// node-level marker applies to eligible scalars only
	flags := ValueFlags{Encrypted: encrypted && EncryptionEligible(contentType)}
	if HasEncryptedValue(p.ContentType, p.Encrypted, p.Value) && !HasEncryptedValue(contentType, flags.Encrypted, plaintext) {
		plaintext, flags.Encrypted, err = MarkAllEncrypted(contentType, plaintext)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := validateParameter(ctx, contentType, flags, plaintext); err != nil {
		tx.Rollback()
		return err
	}
	storedValue, err := encryptForStorage(ctx, contentType, flags.Encrypted, plaintext)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Encrypted = ?, Value = ?, Version = Version + 1, MTime = now() WHERE Path = ?", contentType, flags.Encrypted, storedValue, path)
	if err == nil && HasEncryptedValue(contentType, flags.Encrypted, plaintext) {
		err = sealHistory(ctx, tx, p.ID)
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

// sealHistory rewrites the plaintext history rows of a node that holds an
// encrypted value to envelopes under the active key — in the same
// transaction as the write that made it encrypted, so no version of a secret
// waits in plaintext for a tool. Case rows get every eligible branch marked;
// rows of other types are enveloped whole and marked at node level. Rows
// already sealed are excluded by the marker, which makes this idempotent and
// cheap on a parameter whose history is already sealed. Only the given node's
// rows are touched: a former occupant of the path is another node.
func sealHistory(ctx context.Context, tx *sql.Tx, nodeID int) error {
	rows, err := tx.QueryContext(ctx, "SELECT ID, ContentType, Value FROM my_config_tree_log WHERE NodeID = ? AND NOT Encrypted AND Value IS NOT NULL", nodeID)
	if err != nil {
		return err
	}
	type logRow struct {
		id          int64
		contentType string
		value       string
	}
	var pending []logRow
	for rows.Next() {
		var r logRow
		if err := rows.Scan(&r.id, &r.contentType, &r.value); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close() // before reusing the connection below
	if len(pending) > 0 {
		if err := adoptRegisteredKey(ctx); err != nil {
			return err
		}
	}
	for _, r := range pending {
		value := NullStringOf(r.value)
		var sealed NullString
		if r.contentType == ContentTypeCase {
			marked, _, err := MarkAllEncrypted(r.contentType, value)
			if err != nil {
				return err
			}
			sealed, err = EncryptStoredValue(r.contentType, false, marked)
			if err != nil {
				return err
			}
		} else {
			sealed, err = EncryptStoredValue(r.contentType, true, value)
			if err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, "UPDATE my_config_tree_log SET Value = ?, Encrypted = true WHERE ID = ?", sealed, r.id); err != nil {
			return err
		}
	}
	return nil
}

func sealHistoryByPath(ctx context.Context, tx *sql.Tx, path string) error {
	var id int
	if err := tx.QueryRowContext(ctx, "SELECT ID FROM my_config_tree WHERE Path = ?", path).Scan(&id); err != nil {
		return err
	}
	return sealHistory(ctx, tx, id)
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

	refs, err := encryptedTemplateRefs(ctx, tx)
	if err != nil {
		return err
	}
	return checkDeletedParamSymlinksRecursive(ctx, tx, param.Path, isSymlink(param), 0, refs)
}

// encryptedTemplateVarRe matches the ${/path} variables of a template.
var encryptedTemplateVarRe = regexp.MustCompile(`\$\{(/[^}]*)\}`)

// encryptedTemplateRefs collects the ${/path} references of templates stored
// encrypted — x-template rows and template branches of cases, nested ones
// included — which the SQL arms of the deletion check cannot see through
// the ciphertext. Keyed by referenced path, valued by referrer paths. Built
// once per deletion; a value the keyring cannot open fails the deletion
// (closed), as serving it would.
func encryptedTemplateRefs(ctx context.Context, tx *sql.Tx) (map[string][]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT Path, ContentType, Encrypted, Value
		  FROM my_config_tree
		 WHERE NOT Deleted
		   AND ((Encrypted AND ContentType = '`+ContentTypeTemplate+`')
		        OR (ContentType = '`+ContentTypeCase+`' AND Value LIKE '%"encrypted":true%'))
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refs := map[string][]string{}
	for rows.Next() {
		var path, contentType string
		var encrypted bool
		var value NullString
		if err := rows.Scan(&path, &contentType, &encrypted, &value); err != nil {
			return nil, err
		}
		plaintext, err := DecryptStoredValue(contentType, encrypted, value)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrDecryptionFailed, path, err)
		}
		seen := map[string]bool{}
		for _, text := range templateTexts(contentType, plaintext) {
			for _, m := range encryptedTemplateVarRe.FindAllStringSubmatch(text, -1) {
				if !seen[m[1]] {
					seen[m[1]] = true
					refs[m[1]] = append(refs[m[1]], path)
				}
			}
		}
	}
	return refs, rows.Err()
}

// templateTexts returns the template texts of a value: the value of an
// x-template, the template branches of a (nested) case.
func templateTexts(contentType string, value NullString) []string {
	if !value.Valid {
		return nil
	}
	switch contentType {
	case ContentTypeTemplate:
		return []string{value.String}
	case ContentTypeCase:
		var branches []caseBranch
		if err := json.Unmarshal([]byte(value.String), &branches); err != nil {
			return nil
		}
		var texts []string
		for _, b := range branches {
			texts = append(texts, templateTexts(b.ContentType.String, b.Value)...)
		}
		return texts
	}
	return nil
}

func isSymlink(param *Parameter) bool {
	switch param.ContentType {
	case ContentTypeSymlink:
		return true
	case ContentTypeCase:
		var caseType []struct {
			Type string `json:"mime"`
		}

		if err := json.Unmarshal([]byte(param.Value.String), &caseType); err != nil {
			return false
		}

		for _, ct := range caseType {
			if ct.Type == ContentTypeSymlink {
				return true
			}
		}
	}

	return false
}

func checkDeletedParamSymlinksRecursive(ctx context.Context, tx *sql.Tx, linkTarget string, isSymlink bool, depth int, refs map[string][]string) error {
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

	// encrypted templates are invisible to the SQL above; their references
	// were extracted from the plaintext once per deletion
	expanded = append(expanded, refs[linkTarget]...)

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

		if err = checkDeletedParamSymlinksRecursive(ctx, tx, recursiveLink, false, depth, refs); err != nil {
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

// encryptForStorage is EncryptStoredValue for the write path: when a leaf is
// to be encrypted, new encryptions first adopt the registry's active key.
func encryptForStorage(ctx context.Context, contentType string, encrypted bool, value NullString) (NullString, error) {
	if HasEncryptedValue(contentType, encrypted, value) {
		if err := adoptRegisteredKey(ctx); err != nil {
			return value, err
		}
	}
	return EncryptStoredValue(contentType, encrypted, value)
}

// caseBranch is one entry of a case value. Absent condition keys stay invalid
// NullStrings, mirroring the resolver's selection semantics. Unknown keys are
// ignored, as they always were.
type caseBranch struct {
	Server      NullString `json:"server"`
	Group       NullString `json:"group"`
	Datacenter  NullString `json:"datacenter"`
	Service     NullString `json:"service"`
	ContentType NullString `json:"mime"`
	Value       NullString `json:"value"`
	Encrypted   bool       `json:"encrypted"`
}

// validateParameter checks a plaintext value against its content type and
// flags — before encryption, which is the only place the structure of a
// secret can be checked.
func validateParameter(ctx context.Context, contentType string, flags ValueFlags, value NullString) error {
	if flags.Encrypted && !EncryptionEligible(contentType) {
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
		var i interface{}
		return json.Unmarshal([]byte(value.String), &i)
	case ContentTypeYAML:
		if !value.Valid {
			return ErrInvalidValue
		}
		var i interface{}
		return yaml.Unmarshal([]byte(value.String), &i)
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
		var cases []caseBranch
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
			if err := validateParameter(ctx, c.ContentType.String, ValueFlags{Encrypted: c.Encrypted}, c.Value); err != nil {
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
