//go:build !integration

package admin

import (
	"context"
	"testing"
)

// emitNotification must buffer onto the context sink rather than write to
// notifyDB immediately, so a later rollback cannot leak a phantom notification.
// (notifyDB is nil here; a non-buffered write would dereference it and panic.)
func TestNotificationsAreBufferedUntilCommit(t *testing.T) {
	ctx, sink := withNotifications(context.Background())
	for _, m := range []string{"a", "b", "c"} {
		if err := emitNotification(ctx, m); err != nil {
			t.Fatalf("emitNotification: %v", err)
		}
	}
	if got := len(sink.messages); got != 3 {
		t.Fatalf("buffered %d messages, want 3", got)
	}
}
