package invasion

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	fuzzSeed = flag.Int64("fuzz", time.Now().UnixNano(), "seed to use with fuzz")
)

func TestFuzzMapInvasion(t *testing.T) {
	if testing.Short() {
		t.Skip("fuzz is skipped")
		return
	}
	t.Logf("fuzz using seed %d", *fuzzSeed)

	r := rand.New(rand.NewSource(*fuzzSeed))
	m := GenerateMap(r, rand.Intn(10000), rand.Intn(10000))

	inv := NewSerialInvasion(m, r, ioutil.Discard, rand.Intn(100), rand.Intn(10000))

	period := 100
	for i := 0; inv.Valid(); i++ {
		inv.Next()
		if i%period == 0 {
			require.NoError(t, VerifyInvariants(m, inv.Aliens()))
		}
	}
	require.NoError(t, VerifyInvariants(m, inv.Aliens()))
}
