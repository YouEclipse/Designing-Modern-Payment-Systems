package configapproval

import (
	"errors"
	"testing"
	"time"
)

func TestApprovalRejectsExpiredOrUnboundEvidence(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	base := Review{ChangeID: "change-1", Revision: 4, State: "in_review", Scope: "prod/eu",
		SubjectDigest: [32]byte{1}, ValidatedDigest: [32]byte{1}, ReportDigest: [32]byte{2},
		ValidatorVersion: "validator-v1", ValidationPassed: true,
		ValidationExpiresAt: now.Add(time.Minute), AuthorPersonIDs: []string{"author-1"}}
	actor := Actor{PersonID: "reviewer-1", Scope: base.Scope, CanApprove: true}
	for name, mutate := range map[string]func(*Review, *Actor, *uint64){
		"revision":    func(_ *Review, _ *Actor, version *uint64) { *version = 3 },
		"scope":       func(_ *Review, a *Actor, _ *uint64) { a.Scope = "prod/us" },
		"permission":  func(_ *Review, a *Actor, _ *uint64) { a.CanApprove = false },
		"expiry":      func(r *Review, _ *Actor, _ *uint64) { r.ValidationExpiresAt = now },
		"report":      func(r *Review, _ *Actor, _ *uint64) { r.ReportDigest = [32]byte{} },
		"validator":   func(r *Review, _ *Actor, _ *uint64) { r.ValidatorVersion = "" },
		"failed gate": func(r *Review, _ *Actor, _ *uint64) { r.ValidationPassed = false },
		"self review": func(_ *Review, a *Actor, _ *uint64) { a.PersonID = "author-1" },
	} {
		t.Run(name, func(t *testing.T) {
			review, reviewer, revision := base, actor, base.Revision
			mutate(&review, &reviewer, &revision)
			if approval, err := Approve(review, reviewer, revision, now); !errors.Is(err, ErrApprovalDenied) || approval.ChangeID != "" {
				t.Fatalf("invalid evidence obtained approval: %+v, %v", approval, err)
			}
		})
	}
	approval, err := Approve(base, actor, base.Revision, now)
	if err != nil || approval.SubjectDigest != base.SubjectDigest || approval.ReportDigest != base.ReportDigest ||
		approval.PersonID != actor.PersonID || !approval.ApprovedAt.Equal(now) {
		t.Fatalf("approval lost its evidence binding: %+v, %v", approval, err)
	}
}
