package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var (
	ErrAccessDenied    = errors.New("Access denied")
	ErrAlreadyExists   = errors.New("Parameter already exists")
	ErrVersionNotMatch = errors.New("Version not match")
	ErrCommentRequired = errors.New("Comment required")
	ErrInvalidValue    = errors.New("Invalid value")
	ErrNotEmpty        = errors.New("Parameter has children")
	ErrNotFound        = errors.New("Parameter not found")
	ErrParentNotFound  = errors.New("Parent not found")
)

const selectFields string = `
	SELECT t.ID, t.Name, t.ParentID, t.Path, t.Value, t.ContentType, t.Summary, t.Description, t.Version, t.MTime, t.Deleted,
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
}

func SelectParameter(ctx context.Context, path string) (*Parameter, error) {
	query := selectFromConfig + "WHERE Path = ? AND NOT Deleted\n"

	row := DB.QueryRowContext(ctx, query, Username(ctx), path)
	p := Parameter{}
	err := row.Scan(
		&p.ID, &p.Name, &p.ParentID, &p.Path, &p.Value, &p.ContentType,
		&p.Summary, &p.Description, &p.Version, &p.MTime, &p.Deleted,
		&p.NumChildren, &p.AccessModified, &p.RW, &p.Notification, &p.NotificationModified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !p.RW.Valid {
		p.Value = NullString{}
	}
	return &p, nil
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
	if parameter.ContentType == "application/x-symlink" {
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
		c := Parameter{}
		err := rows.Scan(
			&c.ID, &c.Name, &c.ParentID, &c.Path, &c.Value, &c.ContentType,
			&c.Summary, &c.Description, &c.Version, &c.MTime, &c.Deleted,
			&c.NumChildren, &c.AccessModified, &c.RW, &c.Notification, &c.NotificationModified,
		)
		if err != nil {
			return err
		}
		if !c.RW.Valid {
			c.Value = NullString{}
		}
		parent := paramsMap[int(c.ParentID.Int64)]
		parent.AddChild(c)
	}
	return nil
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
		var p Parameter
		err := rows.Scan(
			&p.ID, &p.Name, &p.ParentID, &p.Path, &p.Value, &p.ContentType,
			&p.Summary, &p.Description, &p.Version, &p.MTime, &p.Deleted,
			&p.NumChildren, &p.AccessModified, &p.RW, &p.Notification, &p.NotificationModified,
		)
		if err != nil {
			return nil, err
		}
		if !p.RW.Valid {
			p.Value = NullString{}
		}
		if pathMap[p.Path] {
			result[p.Path] = &ParameterWithChildren{
				Parameter: p,
				Children:  []Parameter{},
			}
		}
		if p.Path != "/" {
			parentPath, _ := splitPath(p.Path)
			if _, ok := result[parentPath]; ok {
				result[parentPath].Children = append(result[parentPath].Children, p)
			}
		}
	}
	return result, nil
}

func SearchParameters(ctx context.Context, term string) ([]Parameter, error) {
	like := "%" + likeEscape(term) + "%"
	rows, err := DB.QueryContext(ctx, `
		SELECT * FROM (
			`+selectFromConfig+`
			WHERE NOT Deleted
			AND (Name COLLATE ascii_general_ci LIKE ? OR Value COLLATE utf8mb4_unicode_ci LIKE ? OR Summary LIKE ? OR Description LIKE ?)
			ORDER BY Path
		) x
		WHERE RW IS NOT NULL
		OR (Name COLLATE ascii_general_ci LIKE ? OR Summary LIKE ? OR Description LIKE ?)
	`, Username(ctx), like, like, like, like, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]Parameter, 0)
	for rows.Next() {
		var p Parameter
		err := rows.Scan(
			&p.ID, &p.Name, &p.ParentID, &p.Path, &p.Value, &p.ContentType,
			&p.Summary, &p.Description, &p.Version, &p.MTime, &p.Deleted,
			&p.NumChildren, &p.AccessModified, &p.RW, &p.Notification, &p.NotificationModified,
		)
		if err != nil {
			return nil, err
		}
		if !p.RW.Valid {
			p.Value = NullString{}
		}
		list = append(list, p)
	}
	return list, nil
}

func CreateParameter(ctx context.Context, path, contentType, value, summary, description, notification, comment string) error {
	var nullValue NullString
	if contentType != "application/x-null" {
		nullValue.Valid = true
		nullValue.String = value
	}

	if err := validateParameter(ctx, contentType, nullValue); err != nil {
		return err
	}

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

	parentPath, name := splitPath(path)
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
		_, err = tx.ExecContext(ctx, "INSERT INTO my_config_tree (ParentID, Name, ContentType, Value, Summary, Description, Notification) VALUES (?, ?, ?, ?, ?, ?, ?)",
			parent.ID, name, contentType, nullValue, summary, description, nullNotification)
	} else if err == nil {
		if !deleted {
			tx.Rollback()
			return ErrAlreadyExists
		}
		_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Value = ?, Summary = ?, Description = ?, Notification = ?, Version = Version + 1, MTime = now(), Deleted = false WHERE Path = ?",
			contentType, nullValue, summary, description, nullNotification, path)
	}
	if err == nil {
		err = LogLastVersion(ctx, tx, path, comment)
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func SetParameter(ctx context.Context, path string, version int, contentType string, value string, comment string) error {
	var nullValue NullString
	if contentType != "application/x-null" {
		nullValue.Valid = true
		nullValue.String = value
	}

	if err := validateParameter(ctx, contentType, nullValue); err != nil {
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
	if p.Version != version {
		tx.Rollback()
		return ErrVersionNotMatch
	}
	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ContentType = ?, Value = ?, Version = Version + 1, MTime = now() WHERE Path = ?", contentType, nullValue, path)
	if err == nil {
		err = LogLastVersion(ctx, tx, path, comment)
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func MoveParameter(ctx context.Context, path string, newPath string, symlink bool, version int, comment string) error {
	newParentPath, newName := splitPath(newPath)

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

	_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET ParentID = ?, Name = ?, Version = Version + 1, MTime = now() WHERE Path = ?", newParent.ID, newName, path)
	if err == nil {
		likePath := likeEscape(p.Path) + "/%"
		_, err = tx.ExecContext(ctx, "UPDATE my_config_tree SET Path = NULL WHERE Path LIKE ?", likePath)
	}

	if err == nil {
		err = LogLastVersion(ctx, tx, newPath, fmt.Sprintf("Moved from %s. %s", path, comment))
	}

	if err == nil && symlink {
		var nullNotification NullString
		if p.NotificationModified {
			nullNotification.Valid = true
			nullNotification.String = p.Notification
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO my_config_tree (ParentID, Name, ContentType, Value, Notification) VALUES (?, ?, ?, ?, ?)",
			p.ParentID, p.Name, "application/x-symlink", newPath, nullNotification)
		if err == nil {
			err = LogLastVersion(ctx, tx, path, fmt.Sprintf("Moved to %s. %s", newPath, comment))
		}
	}

	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
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
	return tx.Commit()
}

func selectParameterForUpdate(ctx context.Context, tx *sql.Tx, path string) (*Parameter, error) {
	row := tx.QueryRowContext(ctx, selectFromConfig+"WHERE Path = ? AND NOT Deleted FOR UPDATE", Username(ctx), path)
	p := Parameter{}
	err := row.Scan(
		&p.ID, &p.Name, &p.ParentID, &p.Path, &p.Value, &p.ContentType,
		&p.Summary, &p.Description, &p.Version, &p.MTime, &p.Deleted,
		&p.NumChildren, &p.AccessModified, &p.RW, &p.Notification, &p.NotificationModified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !(p.RW.Valid && p.RW.Bool) {
		return nil, ErrAccessDenied
	}
	return &p, nil
}

func validateParameter(ctx context.Context, contentType string, value NullString) error {
	switch contentType {
	case "application/x-null":
		if value.Valid {
			return ErrInvalidValue
		} else {
			return nil
		}
	case "text/plain", "application/x-list", "application/x-server", "application/x-server2", "application/x-template":
		if value.Valid {
			return nil
		} else {
			return ErrInvalidValue
		}
	case "application/json":
		if !value.Valid {
			return ErrInvalidValue
		}
		var i interface{}
		return json.Unmarshal([]byte(value.String), &i)
	case "application/x-yaml":
		if !value.Valid {
			return ErrInvalidValue
		}
		var i interface{}
		return yaml.Unmarshal([]byte(value.String), &i)
	case "application/x-symlink":
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
	case "application/x-case":
		if !value.Valid {
			return ErrInvalidValue
		}
		var cases []map[string]NullString
		if err := json.Unmarshal([]byte(value.String), &cases); err != nil {
			return err
		}
		def := 0
		for _, c := range cases {
			i := 0
			if c["server"].Valid {
				i++
			}
			if c["group"].Valid {
				i++
			}
			if c["datacenter"].Valid {
				i++
			}
			if c["service"].Valid {
				i++
			}
			if i == 0 {
				def++
			} else if i > 1 {
				return ErrInvalidValue
			}
			if err := validateParameter(ctx, c["mime"].String, c["value"]); err != nil {
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
