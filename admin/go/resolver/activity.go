package resolver

import (
	"context"
	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
)

func updateServerActivity(ctx context.Context, server *Server, mtime string, version string) error {
	_, err := DB.ExecContext(ctx, `
		REPLACE INTO my_config_activity (Host, Time, Online, Package)
		SELECT n.Host, n.Time, n.Online, n.Package
		FROM (SELECT ? AS Host, ? AS Time, NOW() AS Online, ? AS Package) n
		WHERE NOT EXISTS (
			SELECT 1
			FROM my_config_activity o
			WHERE o.Host = n.Host
			AND o.Time = n.Time
			AND o.Package = n.Package
			AND o.Online > SUBTIME(n.Online, '00:01:00')
		)
	`, server.Host, mtime, version)
	return err
}
