package admin

import (
	"context"
	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
)

func SelectGroups(ctx context.Context) ([]string, error) {
	rows, err := DB.QueryContext(ctx, "SELECT Name FROM my_config_group ORDER BY Name")
	if err != nil {
		return nil, err
	}
	return ReadStrings(rows)
}

func CreateGroup(ctx context.Context, name string) error {
	_, err := DB.ExecContext(ctx, "INSERT INTO my_config_group (Name) VALUES (?)", name)
	return err
}

func DeleteGroup(ctx context.Context, name string) error {
	_, err := DB.ExecContext(ctx, "DELETE FROM my_config_group WHERE Name = ?", name)
	return err
}

func SelectGroupUsers(ctx context.Context, group string) ([]string, error) {
	rows, err := DB.QueryContext(ctx, "SELECT User FROM my_config_user_group WHERE GroupID = (SELECT ID FROM my_config_group WHERE Name = ?) ORDER BY User", group)
	if err != nil {
		return nil, err
	}
	return ReadStrings(rows)
}

func AddUserToGroup(ctx context.Context, group, user string) error {
	_, err := DB.ExecContext(ctx, "INSERT INTO my_config_user_group (GroupID, User) VALUES ((SELECT ID from my_config_group WHERE Name = ?), ?)", group, user)
	return err
}

func DeleteUserFromGroup(ctx context.Context, group, user string) error {
	_, err := DB.ExecContext(ctx, "DELETE FROM my_config_user_group WHERE GroupID = (SELECT ID FROM my_config_group WHERE Name = ?) AND User = ?", group, user)
	return err
}

func UserIsRoot(ctx context.Context, user string) (bool, error) {
	row := DB.QueryRowContext(ctx, `
		SELECT count(*)
		FROM my_config_group g
		JOIN my_config_user_group ug ON ug.GroupID = g.ID
		WHERE g.Name = 'root' AND ug.User = ?
	`, user)
	var count int
	err := row.Scan(&count)
	return count > 0, err
}
