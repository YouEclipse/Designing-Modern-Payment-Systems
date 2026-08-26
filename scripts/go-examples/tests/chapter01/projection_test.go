package lifecycle

import (
	"context"
	"errors"
	"testing"
)

func TestProjectionRetainsSuccessAndFlagsConflicts(t *testing.T) {
	op := Operation{ID: "op_1", AmountMinor: 5200, DispatchState: OperationUnknown}
	success := Evidence{EventID: "evt_1", OperationID: op.ID, Outcome: OperationSucceeded, AmountMinor: 5200, ProviderOperationRef: "ref_1"}
	failed := Evidence{EventID: "evt_2", OperationID: op.ID, Outcome: OperationFailed, AmountMinor: 5200, ProviderOperationRef: "ref_1"}
	for _, facts := range [][]Evidence{{success, failed}, {failed, success}, {success, success, failed}} {
		got := ProjectOperation(op, facts)
		if !got.Succeeded || !got.Failed || !got.ReviewRequired || got.Unknown || len(got.EvidenceIDs) != 2 {
			t.Fatalf("conflict was discarded: %+v", got)
		}
	}
	changed := success
	changed.AmountMinor++
	if got := ProjectOperation(op, []Evidence{success, changed}); !got.ReviewRequired || len(got.EvidenceIDs) != 1 {
		t.Fatalf("conflicting event ID treated as replay: %+v", got)
	}
	unknown := Evidence{EventID: "evt_unknown", OperationID: op.ID, Outcome: OperationUnknown, AmountMinor: 5200}
	for _, facts := range [][]Evidence{{unknown, success}, {success, unknown}} {
		if got := ProjectOperation(op, facts); !got.Succeeded || got.Unknown || got.ReviewRequired {
			t.Fatalf("late UNKNOWN regressed success: %+v", got)
		}
	}
}

type responseLostGateway struct{}

func (responseLostGateway) Sale(context.Context, SaleCommand) (SaleResult, error) {
	return SaleResult{}, context.DeadlineExceeded
}

func TestNaiveAntiExampleMisclassifiesLostResponse(t *testing.T) {
	marker := errors.New("incorrectly marked failed")
	err := naiveConfirm(context.Background(), responseLostGateway{}, SaleCommand{PaymentID: "pi_1"}, func(string) error { return marker }, func(string, string) error { t.Fatal("unexpected success"); return nil })
	if !errors.Is(err, marker) {
		t.Fatal("anti-example no longer demonstrates its stated bug")
	}
}
