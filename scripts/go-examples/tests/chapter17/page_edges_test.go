package reconciliation

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type pageSourceFunc func(context.Context, Run, string) (Page, error)

func (f pageSourceFunc) Fetch(ctx context.Context, run Run, cursor string) (Page, error) {
	return f(ctx, run, cursor)
}

type pageStoreFunc func(context.Context, Run, Checkpoint, Page) (Checkpoint, error)

func (f pageStoreFunc) CommitPage(ctx context.Context, run Run, checkpoint Checkpoint, page Page) (Checkpoint, error) {
	return f(ctx, run, checkpoint, page)
}

// These spies prove call ordering and contract validation, not database durability.
func TestScanRejectsMismatchedPageBeforeCommitting(t *testing.T) {
	run := Run{ID: "run-1", ScopeDigest: "scope-1", AdapterVersion: "adapter-v1", LeaseEpoch: 7}
	checkpoint := Checkpoint{RunID: run.ID, Version: 3, Cursor: "cursor-a"}
	base := Page{ScopeDigest: run.ScopeDigest, AdapterVersion: run.AdapterVersion,
		FromCursor: checkpoint.Cursor, NextCursor: "cursor-b"}
	for name, mutate := range map[string]func(*Page){
		"scope":     func(p *Page) { p.ScopeDigest = "other-scope" },
		"adapter":   func(p *Page) { p.AdapterVersion = "adapter-v2" },
		"cursor":    func(p *Page) { p.FromCursor = "another-page" },
		"stalled":   func(p *Page) { p.NextCursor = p.FromCursor },
		"no cursor": func(p *Page) { p.NextCursor = "" },
	} {
		t.Run(name, func(t *testing.T) {
			page := base
			mutate(&page)
			committed := false
			source := pageSourceFunc(func(context.Context, Run, string) (Page, error) { return page, nil })
			store := pageStoreFunc(func(context.Context, Run, Checkpoint, Page) (Checkpoint, error) {
				committed = true
				return Checkpoint{}, nil
			})
			got, err := ScanPage(context.Background(), source, store, run, checkpoint)
			if err == nil || got != checkpoint || committed {
				t.Fatalf("invalid page reached commit: %+v, %v", got, err)
			}
		})
	}
}

func TestScanPassesFixedIdentityToSingleCommitAndPropagatesFailure(t *testing.T) {
	run := Run{ID: "run-1", ScopeDigest: "scope-1", AdapterVersion: "adapter-v1", LeaseEpoch: 7}
	checkpoint := Checkpoint{RunID: run.ID, Version: 3, Cursor: "cursor-a"}
	page := Page{ScopeDigest: run.ScopeDigest, AdapterVersion: run.AdapterVersion,
		FromCursor: checkpoint.Cursor, Last: true,
		Records: []SourceRecord{{RecordID: "record-1", RecordRevision: "v1", PayloadDigest: [32]byte{1}}}}
	var calls []string
	source := pageSourceFunc(func(_ context.Context, received Run, cursor string) (Page, error) {
		calls = append(calls, "fetch")
		if received != run || cursor != checkpoint.Cursor {
			t.Fatal("fetch identity changed")
		}
		return page, nil
	})
	failure := errors.New("checkpoint CAS rejected")
	store := pageStoreFunc(func(_ context.Context, received Run, old Checkpoint, receivedPage Page) (Checkpoint, error) {
		calls = append(calls, "commit")
		if received != run || old != checkpoint || !reflect.DeepEqual(receivedPage, page) {
			t.Fatal("commit did not receive the fixed scan contract")
		}
		return old, failure
	})
	got, err := ScanPage(context.Background(), source, store, run, checkpoint)
	if !errors.Is(err, failure) || got != checkpoint || !reflect.DeepEqual(calls, []string{"fetch", "commit"}) {
		t.Fatalf("commit failure or order changed: %+v, %v, %v", got, err, calls)
	}
}
