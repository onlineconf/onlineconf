package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
	"gopkg.in/yaml.v2"
	"strings"
)

var (
	ErrAccessDenied    = errors.New("Access denied")
	ErrAlreadyExists   = errors.New("Parameter already exists")
	ErrVersionNotMatch = errors.New("Version not match")
	ErrCommentRequired = errors.New("Comment required")
	ErrInvalidValue    = errors.New("Invalid value")
)

const selectFromConfig string = `
	SELECT ID, Name, ParentID, Path, Value, ContentType, Summary, Description, Version, MTime, Deleted,
	(SELECT count(*) FROM my_config_tree c WHERE c.ParentID = t.ID AND NOT c.Deleted) AS NumChildren,
	(SELECT count(*) <> 0 FROM my_config_tree_group g WHERE g.NodeID = t.ID) AS AccessModified,
		my_config_tree_access(t.ID, ?) AS RW,
		my_config_tree_notification(t.ID) AS Notification,
		t.Notification IS NOT NULL AS NotificationModified
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
	query := selectFromConfig + "WHERE Path = ?\n"

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

func (p *Parameter) Children(ctx context.Context) ([]Parameter, error) {
	children := make([]Parameter, p.NumChildren)
	query := selectFromConfig + `
		WHERE ParentID = ? AND NOT Deleted
		ORDER BY Name
	`
	rows, err := DB.QueryContext(ctx, query, Username(ctx), p.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		c := &children[i]
		err := rows.Scan(
			&c.ID, &c.Name, &c.ParentID, &c.Path, &c.Value, &c.ContentType,
			&c.Summary, &c.Description, &c.Version, &c.MTime, &c.Deleted,
			&c.NumChildren, &c.AccessModified, &c.RW, &c.Notification, &c.NotificationModified,
		)
		if err != nil {
			return nil, err
		}
		if !c.RW.Valid {
			c.Value = NullString{}
		}
		i += 1
	}
	return children, nil
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

type ParameterWithChildren struct {
	Parameter
	Children []Parameter `json:"children"`
}

func (p *Parameter) WithChildren(ctx context.Context) (*ParameterWithChildren, error) {
	var err error
	pc := ParameterWithChildren{Parameter: *p}
	pc.Children, err = p.Children(ctx)
	if err != nil {
		return nil, err
	}
	return &pc, nil
}

func SearchParameters(ctx context.Context, term string) ([]Parameter, error) {
	like := likeEscape(term)
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
	row := tx.QueryRowContext(ctx, selectFromConfig+"WHERE Path = ? FOR UPDATE", Username(ctx), path)
	p := Parameter{}
	err := row.Scan(
		&p.ID, &p.Name, &p.ParentID, &p.Path, &p.Value, &p.ContentType,
		&p.Summary, &p.Description, &p.Version, &p.MTime, &p.Deleted,
		&p.NumChildren, &p.AccessModified, &p.RW, &p.Notification, &p.NotificationModified,
	)
	if err != nil {
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
	case "text/plain", "application/x-list", "application/x-server", "application/x-template":
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
