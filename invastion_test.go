package invasion

import (
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSerialInvasionReproducible(t *testing.T) {
	seed := time.Now().UnixNano()

	data := `
Foo south=Baz north=Bam
Fam east=Baz north=Cap
Cap east=En west=Foo
En
Baz
Bam
`

	inv := NewSerialInvasion(NewMapFromString(data), rand.New(rand.NewSource(seed)), ioutil.Discard, 3, 10)
	first := []Event{}
	for inv.Valid() {
		for _, ev := range inv.Next() {
			first = append(first, ev)
		}
	}

	inv = NewSerialInvasion(NewMapFromString(data), rand.New(rand.NewSource(seed)), ioutil.Discard, 3, 10)
	second := []Event{}
	for inv.Valid() {
		for _, ev := range inv.Next() {
			second = append(second, ev)
		}
	}

	require.Equal(t, first, second)
}

func TestSerialInvasionTrappedAliens(t *testing.T) {
	data := `
En
Baz
Bam
`

	inv := NewSerialInvasion(NewMapFromString(data), rand.New(rand.NewSource(time.Now().UnixNano())),
		ioutil.Discard, 3, 10)
	inv.Run()

	aliens := inv.Aliens()

	// one or three alien will be trapped, 2 aliens may get destroyed and gc'ed if they invade same city initially
	require.True(t, len(aliens) > 0, "all aliens were destroyed")
	for _, a := range aliens {
		require.True(t, a.Trapped)
	}
}

func TestSerialInvasionAlienMaxMoves(t *testing.T) {
	data := `
En north=Baz
Baz east=Bam
Bam east=En
`

	moves := 10
	inv := NewSerialInvasion(
		NewMapFromString(data), rand.New(rand.NewSource(time.Now().UnixNano())),
		ioutil.Discard, 1, moves)
	inv.Run()

	aliens := inv.Aliens()

	require.Len(t, aliens, 1, "allien dissapeared somehow")
	require.False(t, aliens[0].Trapped)
	require.False(t, aliens[0].Dead)
	require.Equal(t, aliens[0].Moves, moves)
}
