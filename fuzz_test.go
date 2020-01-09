package invasion

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"
)

var (
	fuzzSeed = flag.Int64("fuzz", 0, "seed to use with fuzz")
)

func TestFuzzMapInvasion(t *testing.T) {
	if testing.Short() {
		t.Skip("fuzz is skipped")
		return
	}
	seed := time.Now().UnixNano()
	if *fuzzSeed != 0 {
		seed = *fuzzSeed
	}
	t.Logf("fuzz using seed %d", seed)

	r := rand.New(rand.NewSource(seed))
	m := GenerateMap(r, rand.Intn(10000), rand.Intn(10000))

	inv := NewSerialInvasion(m, r, ioutil.Discard, rand.Intn(100), rand.Intn(10000))
	inv.Run()
}