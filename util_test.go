package sync

import "testing"

func TestSessionID(t *testing.T) {
	sid := sessionID()
	t.Logf("sid: %s", sid)
}
