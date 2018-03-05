package admin

import (
	"context"
	"database/sql"
	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
)

type Access struct {
	Group      string   `json:"group"`
	Overridden bool     `json:"overridden"`
	RW         NullBool `json:"rw"`
}

func SelectAccess(ctx context.Context, path string) ([]Access, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT g.Name,
			tg.GroupID IS NOT NULL AS Overridden,
			IF(tg.GroupID IS NOT NULL, tg.RW, my_config_tree_group_access(t.ID, g.ID)) AS RW
		FROM my_config_group g
		LEFT JOIN my_config_tree t ON t.Path = ?
		LEFT JOIN my_config_tree_group tg ON tg.GroupID = g.ID AND tg.NodeID = t.ID
		ORDER BY g.Name
	`, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]Access, 0)
	for rows.Next() {
		var access Access
		err := rows.Scan(&access.Group, &access.Overridden, &access.RW)
		if err != nil {
			return nil, err
		}
		list = append(list, access)
	}
	return list, nil
}

func UserCanEditAccess(ctx context.Context, user, path string) (bool, error) {
	row := DB.QueryRowContext(ctx, "SELECT my_config_tree_access(t.ID, ?) AS RW FROM my_config_tree t WHERE t.Path = ?", user, path)
	var access NullBool
	err := row.Scan(&access)
	if err != nil {
		return false, err
	}
	if access.Valid && access.Bool {
		return true, nil
	}
	return UserIsRoot(ctx, user)
}

func SetAccess(ctx context.Context, path, group string, rw NullBool) error {
	_, err := DB.ExecContext(ctx, `
		INSERT INTO my_config_tree_group (NodeID, GroupID, RW)
		VALUES ((SELECT ID FROM my_config_tree WHERE Path = ?), (SELECT ID FROM my_config_group WHERE Name = ?), ?)
		ON DUPLICATE KEY UPDATE RW = VALUES(RW)
	`, path, group, rw)
	return err
}

func DeleteAccess(ctx context.Context, path, group string) error {
	_, err := DB.ExecContext(ctx, `
		DELETE FROM my_config_tree_group
		WHERE NodeID = (SELECT ID FROM my_config_tree WHERE Path = ?)
		AND GroupID = (SELECT ID FROM my_config_group WHERE Name = ?)
	`, path, group)
	return err
}

func ClearAccess(ctx context.Context, tx *sql.Tx, nodeID int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM my_config_tree_group WHERE NodeID = ?", nodeID)
	return err
}
