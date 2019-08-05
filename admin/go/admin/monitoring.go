package admin

import (
	"context"

	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
)

type ServerStatus struct {
	Host        string     `json:"host"`
	MTime       NullString `json:"mtime"`
	Online      string     `json:"online"`
	Package     string     `json:"package"`
	MTimeAlert  bool       `json:"mtime_alert"`
	OnlineAlert bool       `json:"online_alert"`
}

var sortColumns map[string]string = map[string]string{"host": "Host", "mtime": "Time", "online": "Online", "package": "Package"}

func SelectServerStatus(ctx context.Context, sort string) ([]ServerStatus, error) {
	row := DB.QueryRowContext(ctx, "SELECT MAX(MTime) AS MTime FROM my_config_tree_log")
	var mtime string
	err := row.Scan(&mtime)
	if err != nil {
		return nil, err
	}

	order, ok := sortColumns[sort]
	if !ok {
		order = "Host"
	}

	rows, err := DB.QueryContext(ctx, `
		SELECT Host, Time, Online, Package,
			Time IS NOT NULL AND Time <> ? AND ? < now() - INTERVAL 30 MINUTE AS TimeAlert,
			Online < now() - INTERVAL 30 MINUTE AS OnlineAlert
		FROM my_config_activity
		ORDER BY `+order, mtime, mtime)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	list := make([]ServerStatus, 0)
	for rows.Next() {
		var s ServerStatus
		err := rows.Scan(&s.Host, &s.MTime, &s.Online, &s.Package, &s.MTimeAlert, &s.OnlineAlert)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func DeleteServerStatus(ctx context.Context, host string) error {
	_, err := DB.ExecContext(ctx, "DELETE FROM my_config_activity WHERE Host = ?", host)
	return err
}
