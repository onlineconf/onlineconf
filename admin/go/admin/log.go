package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hash/crc32"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"

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
	condition := make([]string, 0, 7) // at most one per LogFilter field plus lastID
	bind := make([]interface{}, 0, 7) // Username plus one per filter field

	// Each entry is shown at the path it was actually written at (l.Path), so a
	// path filter naturally returns the history of everything that ever lived
	// there — the current occupant and any soft-deleted former occupants alike.
	bind = append(bind, Username(ctx))
	if filter.Path != "" {
		condition = append(condition, "l.Path = ?")
		bind = append(bind, filter.Path)
	}
	if filter.Author != "" {
		condition = append(condition, "l.Author = ?")
		bind = append(bind, filter.Author)
	}
	if filter.Branch != "" {
		condition = append(condition, "l.Path LIKE ?")
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
			l.ID, l.NodeID, l.Path, l.Version, l.Value, l.ContentType, l.Author, l.MTime, l.Comment, l.Deleted,
			my_config_tree_access(t.ID, ?) AS RW,
			l.ContentType = t.ContentType AND ((l.Value IS NULL AND t.Value IS NULL) OR l.Value = t.Value) AND l.Deleted = t.Deleted AS Same
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
		err := rows.Scan(&l.ID, &l.NodeID, &l.Path, &l.Version, &l.Value, &l.ContentType, &l.Author, &l.MTime, &l.Comment, &l.Deleted, &l.RW, &l.Same)
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
		INSERT INTO my_config_tree_log (NodeID, Path, Version, Value, ContentType, Author, MTime, Comment, Deleted)
		SELECT ID, Path, Version, Value, ContentType, ?, MTime, ?, Deleted
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

// LogMovedDescendants records a change-log entry for every live descendant
// relocated by a subtree move, stamping each with its new path so its history
// stays visible there. A subtree move only re-stamps the root's path with a
// version bump; the descendants' paths are rewritten in bulk without a version
// bump or a log row, so without this their existing log rows would keep their
// pre-move paths and their history would disappear from the new location. Each
// gets the same comment as the root move (its own new path is in Path).
//
// Descendant entries deliberately produce no change notifications: the root
// move already emitted one, and a bulk move must not flood the feed with a
// message per descendant. For the legacy notifyDB delivery that is achieved
// by not calling notify here; for the botapi-based delivery (onlineconf-bot)
// the rows are marked Silent and the notification feed skips them. Callers
// must bump the descendants' Version first so the new (NodeID, Version) rows
// do not collide with the existing ones. Soft-deleted descendants are
// skipped — they are tombstones and keep their pre-move path.
func LogMovedDescendants(ctx context.Context, tx *sql.Tx, newPath, comment string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO my_config_tree_log (NodeID, Path, Version, Value, ContentType, Author, MTime, Comment, Deleted, Silent)
		SELECT ID, Path, Version, Value, ContentType, ?, MTime, ?, Deleted, true
		FROM my_config_tree
		WHERE Path LIKE ? AND NOT Deleted
	`, Username(ctx), comment, likeEscape(newPath)+"/%")
	return err
}

func notify(ctx context.Context, tx *sql.Tx, versionId int64) error {
	if notifyDB == nil {
		return nil
	}

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
		SELECT l.Version, l.Value, l.ContentType, l.Author, l.Comment, l.Deleted
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
	err = rows.Scan(&new.Version, &new.Value, &new.ContentType, &new.Author, &new.Comment, &new.Deleted)
	if err != nil {
		return err
	}
	message := string(avatars[int(crc32.ChecksumIEEE([]byte(new.Author)))%len(avatars)]) + " " + new.Author + "\n"
	if rows.Next() {
		var old logNotifyEntry
		err := rows.Scan(&old.Version, &old.Value, &old.ContentType, &old.Author, &old.Comment, &old.Deleted)
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
			if new.ContentType == ContentTypeCase {
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
						if v, ok := c["value"]; ok && c["mime"] != ContentTypeNull {
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
	return emitNotification(ctx, message)
}

func contentTypeSymbol(contentType string) string {
	switch contentType {
	case ContentTypeSymlink:
		return "➦"
	case ContentTypeCase:
		return "☰"
	case ContentTypeTemplate:
		return "✄"
	case ContentTypeJSON:
		return "ⓙ"
	case ContentTypeYAML:
		return "ⓨ"
	default:
		return ""
	}
}

func insertNotification(ctx context.Context, message string) error {
	_, err := notifyDB.ExecContext(ctx, "INSERT INTO my_change_notification (Origin, Message) VALUES ('onlineconf', ?)", message)
	return err
}

// notificationSink buffers change notifications produced while a transaction is
// open so they are delivered only after it commits. Notifications live in a
// separate database (notifyDB) that cannot enlist in the transaction, so
// writing them inline leaks a phantom notification whenever the transaction
// later rolls back. Buffer during the tx, flush after commit.
//
// A sink is created per write operation (withNotifications), lives only in that
// request's context, and is touched solely by the goroutine that created it —
// synchronously, via emitNotification then commitAndNotify. It is therefore NOT
// safe for concurrent use: do not share it across goroutines (e.g. by fanning
// out emitNotification calls), or the unsynchronized messages slice will race.
type notificationSink struct {
	messages []string
}

type notificationSinkKey struct{}

// withNotifications attaches a notificationSink to the context and returns it.
// Call before opening the transaction; flush the returned sink after commit
// (commitAndNotify does both).
func withNotifications(ctx context.Context) (context.Context, *notificationSink) {
	sink := &notificationSink{}
	return context.WithValue(ctx, notificationSinkKey{}, sink), sink
}

// emitNotification buffers the message for post-commit delivery when a sink is
// present on the context; otherwise it delivers immediately.
func emitNotification(ctx context.Context, message string) error {
	if sink, ok := ctx.Value(notificationSinkKey{}).(*notificationSink); ok {
		sink.messages = append(sink.messages, message)
		return nil
	}
	return insertNotification(ctx, message)
}

// commitAndNotify commits the transaction and, only on success, delivers the
// notifications buffered during it.
func commitAndNotify(ctx context.Context, tx *sql.Tx, sink *notificationSink) error {
	if err := tx.Commit(); err != nil {
		return err
	}
	// The change is durable now; a delivery failure is logged rather than
	// returned so we never report failure for an operation that succeeded.
	for _, message := range sink.messages {
		if err := insertNotification(ctx, message); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to deliver change notification")
		}
	}
	return nil
}
