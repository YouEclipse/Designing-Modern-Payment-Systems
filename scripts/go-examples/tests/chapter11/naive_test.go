package naive

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type simultaneousMissCache struct {
	entered chan struct{}
	release chan struct{}
}

func (c *simultaneousMissCache) Get(context.Context, string) ([]byte, bool, error) {
	c.entered <- struct{}{}
	<-c.release
	return nil, false, nil
}
func (*simultaneousMissCache) Set(context.Context, string, []byte, time.Duration) error { return nil }

type countingProvider struct{ calls atomic.Int32 }

func (p *countingProvider) Sale(context.Context, int64, string) ([]byte, error) {
	p.calls.Add(1)
	return []byte("captured"), nil
}

func TestCacheMissRaceAllowsTwoSales(t *testing.T) {
	cache := &simultaneousMissCache{entered: make(chan struct{}, 2), release: make(chan struct{})}
	p := &countingProvider{}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Confirm(context.Background(), cache, p, "same_key", 12900, "USD"); err != nil {
				t.Error(err)
			}
		}()
	}
	<-cache.entered
	<-cache.entered // Both callers have observed the same miss.
	close(cache.release)
	wg.Wait()
	if p.calls.Load() != 2 {
		t.Fatalf("expected duplicate remote effects, got %d", p.calls.Load())
	}
}

type unavailableCache struct{}

func (unavailableCache) Get(context.Context, string) ([]byte, bool, error) {
	return nil, false, errors.New("cache unavailable")
}
func (unavailableCache) Set(context.Context, string, []byte, time.Duration) error {
	return errors.New("cache unavailable")
}

func TestCacheFailureAfterCaptureMakesBlindReplayUnsafe(t *testing.T) {
	p := &countingProvider{}
	for i := 0; i < 2; i++ {
		if _, err := Confirm(context.Background(), unavailableCache{}, p, "same_key", 12900, "USD"); err == nil {
			t.Fatal("cache write failure expected")
		}
	}
	if p.calls.Load() != 2 {
		t.Fatal("counterexample no longer demonstrates the lost-cache window")
	}
}
