package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hash/crc32"
	"regexp"
	"strings"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

type LogEntry struct {
	ID          int        `json:"id"`
	NodeID      int        `json:"-"`
	Path        string     `json:"path"`
	Version     int        `json:"version"`
	ContentType string     `json:"mime"`
	Value       NullString `json:"data"`
	MTime       string     `json:"mtime"`
	Author      string     `json:"author"`
	Comment     NullString `json:"comment"`
	Deleted     bool       `json:"deleted"`
	RW          NullBool   `json:"rw"`
	Same        bool       `json:"same"`
}

type LogFilter struct {
	Path   string
	Author string
	Branch string
	From   string
	Till   string
	All    bool
}

type logNotifyEntry struct {
	Version     int        `json:"version"`
	ContentType string     `json:"mime"`
	Value       NullString `json:"data"`
	Author      string     `json:"author"`
	Comment     NullString `json:"comment"`
	Deleted     bool       `json:"deleted"`
}

var notifyDB *sql.DB

var ErrNoSuchVersion = errors.New("no such version")

var tillRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

var avatars = []rune("🐀🐁🐂🐃🐄🐅🐆🐇🐈🐉🐊🐋🐌🐍🐎🐏🐐🐑🐒🐓🐕🐖🐗🐘🐙🐛🐜🐝🐞🐟🐠🐡🐢🐥🐨🐩🐪🐫🐬🐭🐮🐯🐰🐱🐲🐳🐴🐵🐶🐷🐸🐹🐺🐻🐼" +
	"🐿🦀🦁🦂🦃🦄🦅🦆🦇🦈🦉🦊🦋🦌🦍🦎🦏🦐🦑🦒🦓🦔🦕🦖🦗🦘🦙🦚🦛🦜🦝🦞🦟🦠🦡🦢🦥🦦🦧🦨🦩")

func SelectLog(ctx context.Context, filter LogFilter, lastID int) ([]LogEntry, error) {
	condition := make([]string, 0)
	bind := []interface{}{Username(ctx)}

	if filter.Path != "" {
		condition = append(condition, "t.Path = ?")
		bind = append(bind, filter.Path)
	}
	if filter.Author != "" {
		condition = append(condition, "l.Author = ?")
		bind = append(bind, filter.Author)
	}
	if filter.Branch != "" {
		condition = append(condition, "t.Path LIKE ?")
		bind = append(bind, likeEscape(filter.Branch)+"%")
	}
	if filter.From != "" {
		condition = append(condition, "l.MTime >= ?")
		bind = append(bind, filter.From)
	}
	if filter.Till != "" {
		if tillRe.MatchString(filter.Till) {
			condition = append(condition, "l.MTime < ? + interval 1 day")
		} else {
			condition = append(condition, "l.MTime < ?")
		}
		bind = append(bind, filter.Till)
	}
	if !filter.All {
		condition = append(condition, "my_config_tree_notification(t.ID) <> 'none'")
	}
	if lastID != 0 {
		condition = append(condition, "l.ID < ?")
		bind = append(bind, lastID)
	}

	query := `
		SELECT
			l.ID, l.NodeID, t.Path, l.Version, l.ContentType, l.Value, l.MTime, l.Author, l.Comment, l.Deleted,
			my_config_tree_access(t.ID, ?) AS RW,
			l.ContentType = t.ContentType AND ((l.Value IS NULL AND t.Value IS NULL) OR l.Value = t.Value) AS Same
		FROM my_config_tree_log l JOIN my_config_tree t ON t.ID = l.NodeID
	`
	if len(condition) > 0 {
		query += "WHERE " + strings.Join(condition, " AND ") + "\n"
	}
	query += "ORDER BY l.ID DESC\nLIMIT 50\n"

	rows, err := DB.QueryContext(ctx, query, bind...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]LogEntry, 0)
	for rows.Next() {
		var l LogEntry
		err := rows.Scan(&l.ID, &l.NodeID, &l.Path, &l.Version, &l.ContentType, &l.Value, &l.MTime, &l.Author, &l.Comment, &l.Deleted, &l.RW, &l.Same)
		if err != nil {
			return nil, err
		}
		if !l.RW.Valid {
			l.Value = NullString{}
		}
		list = append(list, l)
	}
	return list, nil
}

func LogLastVersion(ctx context.Context, tx *sql.Tx, path, comment string) error {
	if comment == "" {
		return ErrCommentRequired
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO my_config_tree_log (NodeID, Version, ContentType, Value, Author, MTime, Comment, Deleted)
		SELECT ID, Version, ContentType, Value, ?, MTime, ?, Deleted
		FROM my_config_tree
		WHERE Path = ?
	`, Username(ctx), comment, path)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	return notify(ctx, tx, id)
}

func notify(ctx context.Context, tx *sql.Tx, versionId int64) error {
	var path, notification string
	row := tx.QueryRowContext(ctx, `
		SELECT t.Path, my_config_tree_notification(t.ID) AS Notification
		FROM my_config_tree t
		JOIN my_config_tree_log l ON l.NodeID = t.ID
		WHERE l.ID = ?
	`, versionId)
	err := row.Scan(&path, &notification)
	if err != nil {
		return err
	}
	if notification == "none" {
		return nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT l.Version, l.ContentType, l.Value, l.Author, l.Comment, l.Deleted
		FROM (SELECT NodeID, Version FROM my_config_tree_log WHERE ID = ?) s
		JOIN my_config_tree_log l ON l.NodeID = s.NodeID AND (l.Version BETWEEN s.Version - 1 AND s.Version)
		ORDER BY l.ID DESC
	`, versionId)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return ErrNoSuchVersion
	}
	var new logNotifyEntry
	err = rows.Scan(&new.Version, &new.ContentType, &new.Value, &new.Author, &new.Comment, &new.Deleted)
	if err != nil {
		return err
	}
	message := string(avatars[int(crc32.ChecksumIEEE([]byte(new.Author)))%len(avatars)]) + " " + new.Author + "\n"
	if rows.Next() {
		var old logNotifyEntry
		err := rows.Scan(&old.Version, &old.ContentType, &old.Value, &old.Author, &old.Comment, &old.Deleted)
		if err != nil {
			return err
		}
		if new.Deleted {
			message += "❌️"
		} else if old.Deleted {
			message += "🆕️"
		} else {
			message += "✏️"
		}
	} else if new.Version == 1 {
		message += "🆕️"
	} else {
		return ErrNoSuchVersion
	}
	message += " " + path
	if !new.Deleted && notification == "with-value" {
		message += " ➾ "
		if new.Value.Valid {
			message += contentTypeSymbol(new.ContentType)
			if new.ContentType == "application/x-case" {
				var data []map[string]string
				err = json.Unmarshal([]byte(new.Value.String), &data)
				if err == nil {
					for _, c := range data {
						message += "\n"
						if s, ok := c["server"]; ok {
							message += "ⓗ«" + s + "»: "
						} else if g, ok := c["group"]; ok {
							message += "ⓖ«" + g + "»: "
						} else if d, ok := c["datacenter"]; ok {
							message += "ⓓ«" + d + "»: "
						} else if s, ok := c["service"]; ok {
							message += "ⓢ«" + s + "»: "
						} else {
							message += "☆️: "
						}
						if v, ok := c["value"]; ok && c["mime"] != "application/x-null" {
							message += contentTypeSymbol(c["mime"]) + "«" + v + "»"
						} else {
							message += "∅"
						}
					}
				} else {
					message += "«" + new.Value.String + "»"
				}
			} else {
				message += "«" + new.Value.String + "»"
			}
		} else {
			message += "∅"
		}
	}
	if new.Comment.Valid {
		message += "\n🗒 " + new.Comment.String
	}
	return insertNotification(ctx, message)
}

func contentTypeSymbol(contentType string) string {
	switch contentType {
	case "application/x-symlink":
		return "➦"
	case "application/x-case":
		return "☰"
	case "application/x-template":
		return "✄"
	case "application/json":
		return "ⓙ"
	case "application/x-yaml":
		return "ⓨ"
	default:
		return ""
	}
}

func insertNotification(ctx context.Context, message string) error {
	_, err := notifyDB.ExecContext(ctx, "INSERT INTO my_change_notification (Origin, Message) VALUES ('onlineconf', ?)", message)
	return err
}
