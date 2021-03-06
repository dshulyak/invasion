package invasion

import (
	"bytes"
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
`

	inv := NewSerialInvasion(NewMapFromString(data), rand.New(rand.NewSource(time.Now().UnixNano())),
		ioutil.Discard, 1, 10)
	inv.Next()
	inv.Next()
	aliens := inv.Aliens()

	// one or three alien will be trapped, 2 aliens may get destroyed and gc'ed if they invade same city initially
	require.Len(t, aliens, 1, "all aliens were destroyed")
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

	require.Empty(t, aliens)
}

func BenchmarkSerialInvasion100(b *testing.B) {
	r := rand.New(rand.NewSource(100))
	m := GenerateMap(r, 1000, 750)

	buf := bytes.NewBuffer(nil)
	_, err := m.WriteTo(buf)
	require.NoError(b, err)
	data := buf.String()
	buf.Reset()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv := NewSerialInvasion(m, r, ioutil.Discard, 100, 10000)
		inv.Run()

		b.StopTimer()
		m = NewMapFromString(data)
		b.StartTimer()
	}
}
