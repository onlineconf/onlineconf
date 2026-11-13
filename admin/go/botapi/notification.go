package botapi

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

const maxNotificationsLimit = 100

var ErrLimitTooLarge = fmt.Errorf("Limit is too large (max allowed: %d)", maxNotificationsLimit)

type Notification struct {
	ValueFlags
	ID           int               `json:"id"`
	NodeID       int               `json:"-"`
	Path         string            `json:"path"`
	Version      int               `json:"version"`
	ContentType  string            `json:"type"`
	Value        NullString        `json:"value"`
	MTime        string            `json:"mtime"`
	Author       string            `json:"author"`
	Comment      NullString        `json:"comment"`
	Action       string            `json:"action"`
	Notification string            `json:"notification"`
	Users        map[string]string `json:"users"`
}

func SelectNotifications(ctx context.Context, lastID, limit int) ([]Notification, int, error) {
	maxID, err := selectLogLastID(ctx)
	if err != nil {
		return nil, 0, err
	}

	if lastID >= maxID || limit == 0 {
		return []Notification{}, maxID, nil
	}

	notifications, err := selectNotifications(ctx, lastID, limit)
	if err != nil {
		return nil, 0, err
	}

	for i := range notifications {
		notifications[i].Users, err = selectNotificationUsers(ctx, notifications[i].NodeID)
		if err != nil {
			return nil, 0, err
		}
	}

	if len(notifications) == limit {
		maxID = notifications[len(notifications)-1].ID
	} else if len(notifications) > 0 {
		if last := notifications[len(notifications)-1].ID; last > maxID {
			maxID = last
		}
	}

	return notifications, maxID, nil
}

func selectLogLastID(ctx context.Context) (int, error) {
	var maxID int
	row := DB.QueryRowContext(ctx, "SELECT max(ID) FROM my_config_tree_log")
	err := row.Scan(&maxID)
	return maxID, err
}

func selectNotifications(ctx context.Context, lastID, limit int) ([]Notification, error) {
	if limit > maxNotificationsLimit {
		return nil, ErrLimitTooLarge
	}
	query := `
		SELECT l.ID, l.NodeID, l.Path, l.Version, l.Value, l.ContentType, l.Encrypted, l.Author, l.MTime, l.Comment,
			IF(l.Deleted, 'delete', IFNULL((SELECT 'modify' FROM my_config_tree_log WHERE NodeID = l.NodeID AND Version = l.Version - 1 AND NOT Deleted), 'create')) AS Action,
			my_config_tree_notification(NodeID) AS Notification
		FROM my_config_tree_log l
		WHERE l.ID > ? AND NOT l.Silent
		HAVING Notification <> 'none'
		ORDER BY l.ID
		LIMIT ` + strconv.Itoa(limit)
	rows, err := DB.QueryContext(ctx, query, lastID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]Notification, 0)
	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.NodeID, &n.Path, &n.Version, &n.Value, &n.ContentType, &n.Encrypted, &n.Author, &n.MTime, &n.Comment, &n.Action, &n.Notification)
		if err != nil {
			return nil, err
		}
		if n.Notification != "with-value" {
			n.Value = NullString{}
		} else {
			// botapi never decrypts: encrypted leaves are served as null
			n.Value = MaskStoredValue(n.ContentType, n.Encrypted, n.Value)
		}
		list = append(list, n)
	}
	return list, nil
}

func selectNotificationUsers(ctx context.Context, nodeID int) (map[string]string, error) {
	query := `
		SELECT u.User, g.RW
		FROM (
			SELECT ID, my_config_tree_group_access(?, ID) AS RW
			FROM my_config_group
		) g
		JOIN my_config_user_group u ON u.GroupID = g.ID
		WHERE g.RW IS NOT NULL
		ORDER BY 1
	`
	rows, err := DB.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make(map[string]string)
	for rows.Next() {
		var user string
		var rw bool
		err := rows.Scan(&user, &rw)
		if err != nil {
			return nil, err
		}
		var access string
		if rw {
			access = "rw"
		} else {
			access = "ro"
		}
		users[user] = access
	}
	return users, nil
}
