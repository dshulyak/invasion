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

// NewCity creates instance of the city with a give name and using lowecased name as id.
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
// All operations on the map are performed with city.ID.
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

// RoutesSize returns number of available routes from a city.
func (m *Map) RoutesSize(from string) int {
	if routes, exists := m.routes[from]; exists {
		return len(routes)
	}
	return 0
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
	if from == to {
		return fmt.Errorf("%w: %s adds a route to self", ErrUnexpectedFormat, from)
	}
	selfExists, err := m.verifyRoute(from, to, direction)
	if err != nil {
		return err
	}
	peerExists, err := m.verifyRoute(to, from, reverseDirection(direction))
	if err != nil {
		return err
	}

	if !selfExists {
		m.addRoute(from, to, direction)
	}
	if !peerExists {
		m.addRoute(to, from, reverseDirection(direction))
	}
	return nil
}

func (m *Map) verifyRoute(from, to, direction string) (bool, error) {
	routes, exist := m.routes[from]
	if !exist {
		return false, nil
	}
	for _, r := range routes {
		if r.Direction == direction {
			if r.To == to {
				return true, nil // route already in the table
			}
			return true, fmt.Errorf(
				"adding conflicting route: %v(%v->%v) conflicts with %v(%v->%v)",
				from, to, direction,
				from, r.To, direction,
			)
		}
	}
	return false, nil
}

func (m *Map) addRoute(from, to, direction string) {
	routes, exist := m.routes[from]
	if !exist {
		routes = make([]Route, 0, maxRoutes)
		m.routes[from] = routes
	}
	m.routes[from] = append(routes, Route{To: to, Direction: direction})
}

// DeleteCity removes city from list of cities and removes all routes.
func (m *Map) DeleteCity(name string) {
	delete(m.cities, name)
	m.DeleteRoutes(name)
}

// DeleteRoutes removes all routes from a city, and restores correctness of the routing table.
func (m *Map) DeleteRoutes(from string) {
	routes, exist := m.routes[from]
	if !exist {
		return
	}
	delete(m.routes, from)
	for _, r := range routes {
		m.deleteRoute(r.To, from, reverseDirection(r.Direction))
	}
}

// deleteRoute deletes route to a specified city. doesn't restore correctness
// of the routing table.
func (m *Map) deleteRoute(from, to, direction string) {
	routes, exist := m.routes[from]
	if !exist {
		panic(fmt.Sprintf("routing table for %v is out of date", from))
	}
	idx := -1
	for i, r := range routes {
		if r.To == to && r.Direction == direction {
			idx = i
		}
	}
	last := len(routes) - 1
	if idx != -1 {
		// FIXME copy for last element is unnecessary
		copy(routes[idx:], routes[idx+1:])
		m.routes[from] = routes[:last]
	}
}

// GetCity queries map for a city using city id.
func (m *Map) GetCity(id string) *City {
	return m.cities[id]
}

// IterateCities loops through cities and associated routes. Iteration function should return true to continue.
func (m *Map) IterateCities(f func(*City, []Route) bool) {
	ids := m.GetCitiesIDs()
	sort.Slice(ids, func(i, j int) bool {
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
	routes, exist := m.routes[from]
	if !exist {
		return nil
	}
	if len(routes) == 0 {
		return nil
	}
	route := routes[r.Intn(len(routes))]
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

		// FIXME city names like New York should be valid
		parts := strings.Split(sr.Text(), " ")
		if len(parts) == 0 {
			continue
		}
		if len(parts) > 5 {
			return total, fmt.Errorf("%w: expect to received one city and at most 4 directions per line. got %v", ErrUnexpectedFormat, sr.Text())
		}

		// keep original name to use it for priting, etc
		// but normalize the id to avoid duplicates on city map
		city := NewCity(parts[0])
		m.AddCity(city)
		for i, r := range parts[1:] {
			parts = strings.Split(r, "=")
			if len(parts) != 2 {
				return total, fmt.Errorf("%w: route %d in %v is in unexpected format", ErrUnexpectedFormat, i+1, sr.Text())
			}
			peer := NewCity(parts[1])
			// FIXME check if city exists, update may overwrite some state
			m.AddCity(peer)
			if err := m.AddRoute(city.ID, peer.ID, strings.ToLower(parts[0])); err != nil {
				return total, err
			}
		}
	}
	return total, sr.Err()
}

// WriteTo writes Map to w in the same format as received. Format:
// Bar south=Baz north=Foo
// Foo north=Bat
//
// Order of the output is deterministic, and will be the same in every execution.
// Any error returned by w.Write will be returned to the caller.
// Caller SHOULD use buffered writer, as Map.WriteTo performs many small writes.
func (m *Map) WriteTo(w io.Writer) (int, error) {
	var (
		total, n int
		err      error
	)
	m.IterateCities(func(city *City, routes []Route) bool {
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
