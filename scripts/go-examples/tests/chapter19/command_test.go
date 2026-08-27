package fencedcommand

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type ownershipStoreFunc func(context.Context, Permit, func(OwnedTx) error) error

func (f ownershipStoreFunc) WithActiveOwnership(ctx context.Context, permit Permit, fn func(OwnedTx) error) error {
	return f(ctx, permit, fn)
}

type commandSpy struct {
	want                 Command
	calls                []string
	insertErr, appendErr error
	t                    *testing.T
}

func (s *commandSpy) InsertOnce(_ context.Context, command Command) error {
	s.calls = append(s.calls, "insert")
	if command != s.want {
		s.t.Fatal("command identity changed before persistence")
	}
	return s.insertErr
}

func (s *commandSpy) AppendDispatchEvent(_ context.Context, command Command) error {
	s.calls = append(s.calls, "outbox")
	if command != s.want {
		s.t.Fatal("outbox changed the immutable dispatch identity")
	}
	return s.appendErr
}

func commandFixture() (Permit, Command) {
	permit := Permit{IntentID: "intent-1", Region: "eu", Epoch: 42}
	command := Command{OperationID: "operation-1", IntentID: permit.IntentID,
		ConnectionID: "candidate-eu", ConnectionRevision: "revision-7", NamespaceID: "account-eu",
		PayloadDigest: "fixed-payload", ProviderKey: "fixed-provider-key"}
	return permit, command
}

// Spies validate the orchestration contract; they do not prove SQL rollback,
// multi-region exclusion, outbox uniqueness, or external Provider deduplication.
func TestPrepareChecksIdentityBeforeOwnershipTransaction(t *testing.T) {
	permit, base := commandFixture()
	for name, mutate := range map[string]func(*Command){
		"intent":     func(c *Command) { c.IntentID = "other-intent" },
		"operation":  func(c *Command) { c.OperationID = "" },
		"connection": func(c *Command) { c.ConnectionID = "" },
		"revision":   func(c *Command) { c.ConnectionRevision = "" },
		"namespace":  func(c *Command) { c.NamespaceID = "" },
		"payload":    func(c *Command) { c.PayloadDigest = "" },
		"key":        func(c *Command) { c.ProviderKey = "" },
	} {
		t.Run(name, func(t *testing.T) {
			command := base
			mutate(&command)
			called := false
			store := ownershipStoreFunc(func(context.Context, Permit, func(OwnedTx) error) error {
				called = true
				return nil
			})
			if err := Prepare(context.Background(), store, permit, command); err == nil || called {
				t.Fatalf("incomplete identity reached transaction: %v", err)
			}
		})
	}
}

func TestPreparePreservesAtomicCallbackOrderAndErrors(t *testing.T) {
	permit, command := commandFixture()
	failure := errors.New("injected store failure")
	for _, stage := range []string{"none", "ownership", "insert", "outbox"} {
		t.Run(stage, func(t *testing.T) {
			spy := &commandSpy{want: command, t: t}
			if stage == "insert" {
				spy.insertErr = failure
			}
			if stage == "outbox" {
				spy.appendErr = failure
			}
			store := ownershipStoreFunc(func(_ context.Context, received Permit, fn func(OwnedTx) error) error {
				if received != permit {
					t.Fatal("ownership permit changed")
				}
				if stage == "ownership" {
					return failure
				}
				return fn(spy)
			})
			err := Prepare(context.Background(), store, permit, command)
			if (stage == "none" && err != nil) || (stage != "none" && !errors.Is(err, failure)) {
				t.Fatalf("callback error was hidden: %v", err)
			}
			var expected []string
			if stage != "ownership" {
				expected = []string{"insert"}
				if stage != "insert" {
					expected = append(expected, "outbox")
				}
			}
			if !reflect.DeepEqual(spy.calls, expected) {
				t.Fatalf("unexpected command order: %v, want %v", spy.calls, expected)
			}
		})
	}
}
