package srm

import (
	"errors"
	"math"
	"testing"
)

func TestSRMDetectsAssignmentLossNotCandidateExposure(t *testing.T) {
	balanced, err := Check(900, 100, 0.1, 0.001)
	if err != nil || math.Abs(balanced.ChiSquare) > 1e-12 || math.Abs(balanced.PValue-1) > 1e-12 || balanced.Flagged || balanced.MethodVersion != "pearson-chi-square-df1-v1" {
		t.Fatalf("plan matched: %+v,%v", balanced, err)
	}
	missing, err := Check(950, 50, 0.1, 0.001)
	if err != nil || math.Abs(missing.ChiSquare-250.0/9) > 1e-10 || missing.PValue >= 0.001 || !missing.Flagged {
		t.Fatalf("lost assignments not flagged: %+v,%v", missing, err)
	}
	symmetric, err := Check(50, 950, 0.9, 0.001)
	if err != nil || math.Abs(symmetric.ChiSquare-missing.ChiSquare) > 1e-10 {
		t.Fatalf("label symmetry: %+v,%v", symmetric, err)
	}
}

func TestSRMRejectsSparseAndInvalidPlans(t *testing.T) {
	for _, counts := range [][2]uint64{{0, 0}, {1, 1}, {40, 4}} {
		if _, err := Check(counts[0], counts[1], 0.1, 0.001); !errors.Is(err, ErrSparseCount) {
			t.Fatalf("sparse=%v,%v", counts, err)
		}
	}
	for _, values := range [][2]float64{{0, 0.05}, {1, 0.05}, {math.NaN(), 0.05}, {math.Inf(1), 0.05}, {0.1, 0}, {0.1, 1}, {0.1, math.NaN()}, {0.1, math.Inf(1)}} {
		if _, err := Check(900, 100, values[0], values[1]); !errors.Is(err, ErrInvalidPlan) {
			t.Fatalf("invalid plan=%v,%v", values, err)
		}
	}
	if got, err := Check(^uint64(0), ^uint64(0), 0.5, 0.001); err != nil || math.IsInf(got.ChiSquare, 0) || got.Flagged {
		t.Fatalf("count addition overflow: %+v,%v", got, err)
	}
}
