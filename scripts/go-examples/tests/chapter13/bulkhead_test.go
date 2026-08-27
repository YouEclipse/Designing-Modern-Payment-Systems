package bulkhead

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCanceledWaitersNeverAcquireOrLeakAPermit(t *testing.T) {
	if _, err := New(0); !errors.Is(err, ErrInvalidLimit) {
		t.Fatal(err)
	}
	l, _ := New(1)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for i := 0; i < 100; i++ {
		if release, err := l.Acquire(canceled); !errors.Is(err, context.Canceled) || release != nil {
			if release != nil {
				release()
			}
			t.Fatalf("pre-canceled request acquired: %v", err)
		}
	}
	release, err := l.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancelWait := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := l.Acquire(ctx); done <- err }()
	cancelWait()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	release()
	release() // Idempotent release must not consume somebody else's slot.
	retry, err := l.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	retry()
}

func TestConcurrentInFlightIsBoundedAndQueryHasIndependentCapacity(t *testing.T) {
	const limit = 8
	sale, _ := New(limit)
	query, _ := New(1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	holders := make([]func(), limit)
	for i := range holders {
		var err error
		holders[i], err = sale.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
	queryRelease, err := query.Acquire(ctx)
	if err != nil {
		t.Fatal("SALE saturation blocked independent QUERY capacity")
	}
	queryRelease()
	for _, release := range holders {
		release()
	}
	var active, peak atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			release, err := sale.Acquire(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			defer release()
			current := active.Add(1)
			for old := peak.Load(); current > old && !peak.CompareAndSwap(old, current); old = peak.Load() {
			}
			runtime.Gosched()
			active.Add(-1)
		}()
	}
	close(start)
	wg.Wait()
	if peak.Load() > limit || peak.Load() < 1 || active.Load() != 0 {
		t.Fatalf("in-flight peak=%d current=%d", peak.Load(), active.Load())
	}
}
