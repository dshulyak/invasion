package invasion

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
)

const (
	north = "north"
	south = "south"
	east  = "east"
	west  = "west"

	maxRoutes = 4
)

var (
	// ErrUnexpectedFormat returned if map has items with unexpected format.
	ErrUnexpectedFormat = errors.New("Unexpected format")

	emptySpace = []byte(" ")
	equalSign  = []byte("=")
	newLine    = []byte("\n")
)

// reverseDirection returns a reverse direction for any valid direction.
// might be used to test validity.
func reverseDirection(direction string) string {
	switch direction {
	case north:
		return south
	case south:
		return north
	case east:
		return west
	case west:
		return east
	default:
		panic("unknown direction: " + direction)
	}
}

func NewCity(name string) *City {
	return &City{
		Name: name,
		ID:   strings.ToLower(name),
	}
}

// City is a representation of a point on the map.
type City struct {
	ID string
	// Name of the city.
	Name string
	// Invaded is true if another alien already invaded this city.
	// TODO can be a counter for more flexible simulation.
	Invaded bool
	Invader int
	// Destroyed is true if two alliens fought in this city.
	Destroyed bool
}

// Route represents route from a city to another city using cardinal direction.
type Route struct {
	To        string
	Direction string
}

// NewMap returns new instance of the map.
func NewMap() *Map {
	return &Map{
		cities: map[string]*City{},
		routes: map[string][]Route{},
	}
}

// NewMapFromString creates from a pregenerated string.
// Provided data must be a valid map, otherwise function will panic.
func NewMapFromString(data string) *Map {
	m := NewMap()
	n, err := m.ReadFrom(bytes.NewBuffer([]byte(data)))
	if err != nil {
		panic(fmt.Sprintf("not a valid map: %v", err))
	}
	if n != len(data) {
		panic(fmt.Sprintf("unable to read whole map, data length is %d, read %d", len(data), n))
	}
	return m
}

// Map is a representation of geographical map, that keeps track of Routes between cities.
type Map struct {
	cities map[string]*City
	// routes are represented as city id => slice of routes
	// slice is used to simplify determnistic simulation
	// note that small slice is equal or better in term of performance then map for key/value sets/gets
	routes map[string][]Route
}

// Size returns number of cities on the map.
func (m *Map) Size() int {
	return len(m.cities)
}

// AddCity add city to a map.
func (m *Map) AddCity(city *City) {
	m.cities[city.ID] = city
}

// MustAddRoute same as AddRoute but panics is route already exists.
func (m *Map) MustAddRoute(from, to, direction string) {
	if err := m.AddRoute(from, to, direction); err != nil {
		panic(err.Error())
	}
}

// AddRoute from a city to another city via cardinal direction.
func (m *Map) AddRoute(from, to, direction string) error {
	if err := m.addRoute(from, to, direction); err != nil {
		return err
	}
	if err := m.addRoute(to, from, reverseDirection(direction)); err != nil {
		return err
	}
	return nil
}

func (m *Map) addRoute(from, to, direction string) error {
	cRoutes, exist := m.routes[from]
	if !exist {
		cRoutes = make([]Route, 0, maxRoutes)
		m.routes[from] = cRoutes
	}
	for i := range cRoutes {
		r := &cRoutes[i]
		if r.Direction == direction {
			if r.To == to {
				return nil // route already in the table
			}
			return fmt.Errorf("adding conflicting route: from %v to %v via %v", from, to, direction)
		}
	}
	m.routes[from] = append(cRoutes, Route{To: to, Direction: direction})
	return nil
}

// DeleteCity removes city from list of cities and removes all routes.
func (m *Map) DeleteCity(name string) {
	delete(m.cities, name)
	m.DeleteRoutes(name)
}

// DeleteRoutes removes all routes from a city, and restores correctness of the routing table.
func (m *Map) DeleteRoutes(from string) {
	cRoutes, exist := m.routes[from]
	if !exist {
		return
	}
	delete(m.routes, from)
	for i := range cRoutes {
		r := &cRoutes[i]
		m.deleteRoute(r.To, from)
	}
}

// deleteRoute deletes route to a specified city. doesn't restore correctness
// of the routing table.
func (m *Map) deleteRoute(from, to string) {
	cRoutes, exist := m.routes[from]
	if !exist {
		panic(fmt.Sprintf("routing table for %v is out of date", from))
	}
	idx := -1
	for i := range cRoutes {
		r := &cRoutes[i]
		if r.To == to {
			idx = i
		}
	}
	last := len(cRoutes) - 1
	if idx != -1 {
		copy(cRoutes[idx:], cRoutes[idx+1:])
		m.routes[from] = cRoutes[:last]
	}
}

// GetCity queries map for a city using city id.
func (m *Map) GetCity(id string) *City {
	return m.cities[id]
}

func (m *Map) iterateCities(f func(*City, []Route) bool) {
	ids := m.GetCitiesIDs()
	sort.SliceStable(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	for _, id := range ids {
		routes := m.routes[id]
		if !f(m.GetCity(id), routes) {
			return
		}
	}
}

// GetCitiesIDs returns slice with all cities identifies that are currently on the map.
func (m *Map) GetCitiesIDs() []string {
	ids := make([]string, 0, len(m.cities))
	for id := range m.cities {
		ids = append(ids, id)
	}
	return ids
}

// GetRandomCityFrom picks a random city based on existing routs from a specified city.
func (m *Map) GetRandomCityFrom(r *rand.Rand, from string) *City {
	cRoutes, exist := m.routes[from]
	if !exist {
		return nil
	}
	if len(cRoutes) == 0 {
		return nil
	}
	route := cRoutes[r.Intn(len(cRoutes))]
	return m.cities[route.To]
}

// ReadFrom reads from r until io.EOF and adds all cities and routes found.
// Any error except io.EOF will be returned.
func (m *Map) ReadFrom(r io.Reader) (int, error) {
	sr := bufio.NewScanner(r)
	total := 0
	for sr.Scan() {
		total += len(sr.Bytes())
		total++ // scanner splits based on new line byte
		if len(sr.Bytes()) == 0 {
			continue
		}
		parts := strings.Split(sr.Text(), " ")
		if len(parts) == 0 {
			continue
		}
		// keep original name to use it for priting, etc
		// but normalize the id to avoid duplicates on city map
		id := strings.ToLower(parts[0])
		m.AddCity(NewCity(parts[0]))
		for i, r := range parts[1:] {
			parts = strings.Split(r, "=")
			if len(parts) != 2 {
				return total, fmt.Errorf("%w: route %d in %v is in unexpected format", ErrUnexpectedFormat, i+1, sr.Text())
			}
			m.MustAddRoute(id, strings.ToLower(parts[1]), strings.ToLower(parts[0]))
		}
	}
	return total, sr.Err()
}

// WriteTo writes Map to w in the same format as received. Format:
// Bar south=Baz north=Foo
// Foo north=Bat
//
// Order of the output is deterministic, and will be the same in every execution.
// Any error returned by w.Write will be back-propagated.
// Caller SHOULD use buffered writer, as Map.WriteTo performs many small writes.
func (m *Map) WriteTo(w io.Writer) (int, error) {
	var (
		total, n int
		err      error
	)
	m.iterateCities(func(city *City, routes []Route) bool {
		// TODO consider counting required number of bytes and allocating slice ones
		n, err = w.Write([]byte(city.Name))
		if err != nil {
			return false
		}
		total += n
		for _, r := range routes {
			n, err = w.Write(emptySpace)
			if err != nil {
				return false
			}
			total += n
			n, err = w.Write([]byte(r.Direction))
			if err != nil {
				return false
			}
			total += n
			n, err = w.Write(equalSign)
			if err != nil {
				return false
			}
			total += n
			city := m.GetCity(r.To)
			n, err = w.Write([]byte(city.Name))
			if err != nil {
				return false
			}
			total += n
		}
		n, err = w.Write(newLine)
		if err != nil {
			return false
		}
		total += n
		return true
	})
	return total, err
}
