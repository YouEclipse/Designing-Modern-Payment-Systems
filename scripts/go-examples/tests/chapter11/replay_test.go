package replay

import (
	"errors"
	"testing"
)

func TestOnlyUniqueConstraintWinnerMayExecute(t *testing.T) {
	in := Incoming{CommandScope: "confirm", Fingerprint: []byte("hash_1")}
	record := Record{CommandScope: "confirm", Fingerprint: []byte("hash_1"), State: StateInProgress, ResourceID: "intent_1"}
	if action, err := Resolve(true, in, Record{}); err != nil || action != ActionExecute {
		t.Fatalf("owner=%s,%v", action, err)
	}
	for i := 0; i < 20; i++ { // The classifier consumes the DB winner; it does not simulate a DB lock.
		if action, err := Resolve(false, in, record); err != nil || action != ActionReadResource {
			t.Fatalf("loser executed: %s,%v", action, err)
		}
	}
	record.State = StateCompleted
	if action, err := Resolve(false, in, record); err != nil || action != ActionReplayResult {
		t.Fatalf("completed=%s,%v", action, err)
	}
	for _, conflict := range []Incoming{{CommandScope: "create", Fingerprint: []byte("hash_1")}, {CommandScope: "confirm", Fingerprint: []byte("hash_2")}} {
		if action, err := Resolve(false, conflict, record); !errors.Is(err, ErrKeyConflict) || action != "" {
			t.Fatalf("conflict=%s,%v", action, err)
		}
	}
}
