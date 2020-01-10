package invasion

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFromValidMap(t *testing.T) {
	text := `
Foo south=Baz north=Tot-H
Tot-H east=Bar
Baz
Bar
`
	buf := bytes.NewBuffer([]byte(text))

	expected := NewMap()
	expected.AddCity(NewCity("Foo"))
	expected.AddCity(NewCity("Tot-H"))
	expected.AddCity(NewCity("Baz"))
	expected.AddCity(NewCity("Bar"))
	expected.AddRoute("foo", "baz", south)
	expected.AddRoute("foo", "tot-h", north)
	expected.AddRoute("tot-h", "bar", east)

	received := NewMap()
	n, err := received.ReadFrom(buf)
	require.NoError(t, err)
	require.Equal(t, len(text), n)

	require.Equal(t, expected, received)
}

func TestReadFromUnexpectedFormat(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`
Foo south=Baz north=Tot=sothe=Bar
Tot-H east=Bar
`))

	received := NewMap()
	_, err := received.ReadFrom(buf)
	require.True(t, errors.Is(err, ErrUnexpectedFormat), "error is %v", err)
}

func TestReadFromManyDirections(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`
Foo south=Baz north=Tot east=Another west=Something north-west=SomethingElse
`))

	received := NewMap()
	_, err := received.ReadFrom(buf)
	require.True(t, errors.Is(err, ErrUnexpectedFormat), "error is %v", err)
}

func TestReadFromInvalidRoutes(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`
Foo south=Bar
Baz north=Foo
`))

	received := NewMap()
	_, err := received.ReadFrom(buf)
	require.Error(t, err)
}

func TestWriteToConsistentWithOriginal(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 50))

	// note that order of insertion follows lexic order of the nodes, simplifies test verification
	original := NewMap()
	original.AddCity(NewCity("Bar"))
	original.AddRoute("tot-h", "bar", east)

	original.AddCity(NewCity("Baz"))
	original.AddRoute("foo", "baz", south)

	original.AddCity(NewCity("Foo"))
	original.AddCity(NewCity("Tot-H"))
	original.AddRoute("foo", "tot-h", north)

	wn, err := original.WriteTo(buf)
	require.NoError(t, err)
	require.Equal(t, buf.Len(), wn)

	recovered := NewMap()
	rn, err := recovered.ReadFrom(buf)
	require.NoError(t, err)
	require.Equal(t, wn, rn)

	require.Equal(t, original, recovered)
}

func TestMapGetRandomCity(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	m := NewMap()
	m.AddCity(NewCity("Bar"))
	m.AddCity(NewCity("Baz"))
	m.AddRoute("bar", "baz", east)

	baz := m.GetRandomCityFrom(rng, "bar")
	require.NotNil(t, baz)
	require.Equal(t, "Baz", baz.Name)

	m.DeleteCity("baz")
	baz = m.GetRandomCityFrom(rng, "bar")
	require.Nil(t, baz)
}

func TestMapCityDeletionPreservesConsistency(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	m := NewMap()
	m.AddCity(NewCity("Bar"))
	m.AddCity(NewCity("Baz"))
	m.AddCity(NewCity("Foo"))
	m.AddRoute("bar", "baz", east)
	m.AddRoute("foo", "bar", north)

	expect := `Bar east=Baz
Baz west=Bar
`
	m.DeleteCity("foo")

	_, err := m.WriteTo(buf)
	require.NoError(t, err)

	require.Equal(t, expect, buf.String())
}
