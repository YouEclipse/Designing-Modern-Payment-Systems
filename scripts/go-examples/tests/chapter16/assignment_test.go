package experiment

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func assignmentFixture() (Unit, Artifact, []byte) {
	return Unit{MerchantID: "merchant_demo", PaymentIntentID: "intent_42"}, Artifact{
		ExperimentID: "exp_cost_v43", AllocationVersion: "allocation_v1", BucketKeyVersion: "bucket_key_v1",
		ControlPolicyVersion: "score_v42", TreatmentPolicyVersion: "score_v43", TreatmentBuckets: 6554,
	}, []byte("test-only-bucket-key-32-bytes-long")
}

func TestOneThousandStableIntentAssignmentsAndRampPreservesBindings(t *testing.T) {
	unit, artifact, key := assignmentFixture()
	winners := make([]Assignment, 1000)
	vector := sha256.New()
	treatment := 0
	for i := range winners {
		unit.PaymentIntentID = fmt.Sprintf("intent_%d", i)
		got, err := Resolve(unit, nil, artifact, key, true, true)
		if err != nil || got.Assignment == nil {
			t.Fatalf("assignment=%+v,%v", got, err)
		}
		winners[i] = *got.Assignment
		var bucket [2]byte
		binary.BigEndian.PutUint16(bucket[:], uint16(got.Assignment.Bucket))
		_, _ = vector.Write(bucket[:])
		if got.Assignment.Variant == "treatment" {
			treatment++
		}
		if got.Assignment.Bucket >= BucketSpace || got.Assignment.ProbabilityDenominator != BucketSpace {
			t.Fatal(got.Assignment)
		}
	}
	// Independently generated with Node's HMAC-SHA256 and length-prefixed UTF-8.
	// Exact vector coverage catches a constant-bucket implementation without a
	// flaky statistical assertion about random population balance.
	if treatment != 84 || fmt.Sprintf("%x", vector.Sum(nil)) != "52ee47c72c7638600b4abbeff352068a66969314e4d58d1dce3112be6cab2854" {
		t.Fatalf("bucket golden drift: treatment=%d digest=%x", treatment, vector.Sum(nil))
	}
	ramped := artifact
	ramped.AllocationVersion, ramped.TreatmentBuckets, ramped.TreatmentPolicyVersion = "allocation_v2", 20000, "score_v44"
	for i := len(winners) - 1; i >= 0; i-- {
		want := winners[i]
		fresh, err := Resolve(want.Unit, nil, artifact, key, true, true)
		if err != nil || !reflect.DeepEqual(*fresh.Assignment, want) {
			t.Fatalf("restart/order changed assignment: %+v,%v", fresh, err)
		}
		bound, err := Resolve(want.Unit, &want, ramped, nil, false, false)
		if err != nil || !reflect.DeepEqual(*bound.Assignment, want) || bound.Reason != "existing_assignment" {
			t.Fatalf("kill/ramp rewrote assignment: %+v,%v", bound, err)
		}
		bound.Assignment.Variant = "changed"
		if want.Variant == "changed" {
			t.Fatal("replay aliases immutable binding")
		}
		freshRamp, err := Resolve(want.Unit, nil, ramped, key, true, true)
		if err != nil || freshRamp.Assignment.Bucket != want.Bucket {
			t.Fatal("allocation version changed bucket space")
		}
	}
}

func TestAssignerHasNoMutableSharedState(t *testing.T) {
	unit, artifact, key := assignmentFixture()
	want, err := Resolve(unit, nil, artifact, key, true, true)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := Resolve(unit, nil, artifact, key, true, true)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Errorf("concurrent assignment drift: %+v,%v", got, err)
			}
		}()
	}
	wg.Wait()
}

func TestKillSwitchScopeAndMalformedBindingFailExplicitly(t *testing.T) {
	unit, artifact, key := assignmentFixture()
	for _, tc := range []struct {
		eligible, open bool
		reason         string
	}{{false, true, "not_enrolled_scope"}, {true, false, "not_enrolled_disabled"}} {
		got, err := Resolve(unit, nil, artifact, nil, tc.eligible, tc.open)
		if err != nil || got.Assignment != nil || got.Reason != tc.reason {
			t.Fatalf("enrollment=%+v,%v", got, err)
		}
	}
	winner, _ := Resolve(unit, nil, artifact, key, true, true)
	for _, change := range []func(*Assignment){
		func(a *Assignment) { a.Unit.MerchantID = "other" }, func(a *Assignment) { a.ExperimentID = "other" },
		func(a *Assignment) { a.Version = "" }, func(a *Assignment) { a.Variant = "unknown" },
		func(a *Assignment) { a.ProbabilityNumerator = 0 }, func(a *Assignment) { a.ProbabilityDenominator = 100 },
		func(a *Assignment) { a.Bucket = BucketSpace },
		func(a *Assignment) { a.Variant, a.Bucket, a.ProbabilityNumerator = "treatment", 60000, 6554 },
		func(a *Assignment) { a.Variant, a.Bucket, a.ProbabilityNumerator = "control", 1, BucketSpace-6554 },
	} {
		bad := *winner.Assignment
		change(&bad)
		if _, err := Resolve(unit, &bad, artifact, key, true, true); !errors.Is(err, ErrInvalidAssignment) {
			t.Fatalf("bad binding accepted: %+v,%v", bad, err)
		}
	}
	if _, err := Resolve(unit, nil, artifact, []byte("short"), true, true); !errors.Is(err, ErrInvalidAssignment) {
		t.Fatal("weak key accepted")
	}
	unit.PaymentIntentID = strings.Repeat("x", 1025)
	if _, err := Resolve(unit, nil, artifact, key, true, true); !errors.Is(err, ErrInvalidAssignment) {
		t.Fatal("unbounded unit accepted")
	}
}
