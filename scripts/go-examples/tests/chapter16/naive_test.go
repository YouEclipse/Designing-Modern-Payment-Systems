package naiveexperiment

import (
	"math/rand"
	"testing"
)

func TestPerRequestRandomnessCanChangeVariantForOneIntent(t *testing.T) {
	// Go 1.25 makes rand.Seed a no-op by default. Re-enable the documented
	// legacy switch only inside this negative example to fix its random vector.
	t.Setenv("GODEBUG", "randseednop=0")
	rand.Seed(1)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		variant := PickStrategy() // The same conceptual Intent has no input here.
		if variant != "champion" && variant != "challenger" {
			t.Fatalf("unknown variant %q", variant)
		}
		seen[variant] = true
	}
	if len(seen) != 2 {
		t.Fatal("fixed random sequence should expose unstable assignment")
	}
}
