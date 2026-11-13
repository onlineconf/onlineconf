package admin

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

// The audit log (my_config_audit) records the acts worth telling a
// parameter's owners about that are not changes of the parameter itself —
// today the reveals of encrypted values. A row is written before the act
// (a refused row refuses the act), served to bots through the botapi audit
// feed, and echoed to the legacy notification table.

// Audit kinds: the members of the my_config_audit.Kind enum. A new kind is
// appended to the enum by its update-N.sql — at the end of the list, which
// MySQL applies as a metadata-only change — and gets its wording in
// auditMessage.
const (
	AuditRevealView    = "reveal_view"    // the current value was revealed, to view or to edit it
	AuditRevealHistory = "reveal_history" // the parameter's history was revealed
)

// auditMessage is the legacy notification wording of an audit event.
func auditMessage(kind, user, path string) string {
	switch kind {
	case AuditRevealView:
		return fmt.Sprintf("🔓 %s revealed %s", user, path)
	case AuditRevealHistory:
		return fmt.Sprintf("🔓 %s revealed the history of %s", user, path)
	default:
		return fmt.Sprintf("%s: %s on %s", user, kind, path)
	}
}

// recordAudit writes the audit row; the act is refused if that fails. The
// legacy message follows and never fails the act.
func recordAudit(ctx context.Context, nodeID int, path, kind string) error {
	_, err := DB.ExecContext(ctx, "INSERT INTO my_config_audit (NodeID, Path, User, Kind) VALUES (?, ?, ?, ?)",
		nodeID, path, Username(ctx), kind)
	if err != nil {
		return err
	}
	notifyAudit(ctx, nodeID, path, kind)
	return nil
}

// notifyAudit emits the legacy notifyDB message of an audit event to the
// node's notification subscribers; failures are logged, not returned — the
// audit row is the record, the message a courtesy.
func notifyAudit(ctx context.Context, nodeID int, path, kind string) {
	if notifyDB == nil {
		return
	}
	var notification string
	err := DB.QueryRowContext(ctx, "SELECT my_config_tree_notification(?)", nodeID).Scan(&notification)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("path", path).Msg("failed to read the notification setting for an audit event")
		return
	}
	if notification == "none" {
		return
	}
	if err := insertNotification(ctx, auditMessage(kind, Username(ctx), path)); err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("path", path).Str("kind", kind).Msg("failed to deliver the audit notification")
	}
}
