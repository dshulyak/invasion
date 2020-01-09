package invasion

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
)

func NewAliens(n int) map[int]*Alien {
	rst := make(map[int]*Alien, n)
	for i := 0; i < n; i++ {
		rst[i] = &Alien{ID: i}
	}
	return rst
}

type Alien struct {
	ID int

	Moves int

	Dead bool

	Location string
	Trapped  bool
}

func (a *Alien) Leave(city *City) {
	city.Invaded = false
	city.Invader = -1
	a.Location = ""
}

func (a *Alien) Invade(city *City) {
	a.Location = city.ID
	city.Invaded = true
	city.Invader = a.ID
}

func (a *Alien) FightAt(contender *Alien, city *City) {
	a.Dead = true
	contender.Dead = true
	city.Destroyed = true
}

func NewSerialInvasion(m *Map, r *rand.Rand, notifier io.Writer, aliensCount, moves int) *SerialInvasion {
	aliens := NewAliens(aliensCount)

	// sort both so that we don't depend on the implementaton of aliens and maps
	order := make([]int, 0, len(aliens))
	for i := range aliens {
		order = append(order, i)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return order[i] < order[j]
	})

	citiesOrder := m.GetCitiesIDs()
	sort.SliceStable(citiesOrder, func(i, j int) bool {
		return citiesOrder[i] < citiesOrder[j]
	})

	return &SerialInvasion{
		r:           r,
		notifier:    notifier,
		m:           m,
		aliens:      aliens,
		aliensOrder: order,
		citiesOrder: citiesOrder,
		maxMoves:    moves,
	}
}

type SerialInvasion struct {
	r        *rand.Rand
	notifier io.Writer

	aliensOrder []int
	aliens      map[int]*Alien

	citiesOrder []string
	m           *Map

	maxMoves int
}

func (si *SerialInvasion) Run() {
	for si.Valid() {
		evs := si.Next()
		for _, ev := range evs {
			if ev.Important {
				_, _ = fmt.Fprintln(si.notifier, ev.Data)
			}
		}
	}
}

func (si *SerialInvasion) Next() (evs []Event) {
	var (
		idx   = si.r.Intn(len(si.aliensOrder))
		alien = si.aliens[si.aliensOrder[idx]]
	)
	alien.Moves++

	if len(alien.Location) == 0 && si.m.Size() != 0 {

		// alien starts in random city
		cidx := si.r.Intn(len(si.citiesOrder))
		city := si.m.GetCity(si.citiesOrder[cidx])
		// another alien can start at the same city, so we check for that from the start
		evs = append(evs, NewEvent("alien %d invades %v", alien.ID, city.Name))
		evs = si.invadeCity(alien, city, evs)

	} else if !alien.Trapped && !alien.Dead {

		// if alien already invaded a city, pick a random one based on existing routes
		city := si.m.GetRandomCityFrom(si.r, alien.Location)
		if city == nil {
			evs = append(evs, NewEvent("alien %d trapped at %v", alien.ID, alien.Location))
			// if there are no cities reachable from current location then alien is trapped
			alien.Trapped = true
		} else {
			evs = append(evs, NewEvent("alien %d invades %v from %v", alien.ID, city.Name, alien.Location))
			// otherwise try to invade new city
			alien.Leave(si.m.GetCity(alien.Location))
			evs = si.invadeCity(alien, city, evs)
		}
	}

	// gc alien whenever simulation observed his death
	if alien.Dead {
		si.deleteAlienFromOrder(idx)
	}
	return evs
}

func (si *SerialInvasion) deleteCityFromOrder(requested string) {
	idx := -1
	for i, id := range si.citiesOrder {
		if requested == id {
			idx = i
		}
	}
	if idx != -1 {
		last := len(si.citiesOrder) - 1
		// FIXME copy for last element is unnecessary
		copy(si.citiesOrder[idx:], si.citiesOrder[idx+1:])
		si.citiesOrder = si.citiesOrder[:last]
	}
}

func (si *SerialInvasion) Aliens() []*Alien {
	rst := make([]*Alien, 0, len(si.aliensOrder))
	for _, idx := range si.aliensOrder {
		rst = append(rst, si.aliens[idx])
	}
	return rst
}

func (si *SerialInvasion) deleteAlienFromOrder(idx int) {
	delete(si.aliens, si.aliensOrder[idx])
	last := len(si.aliensOrder) - 1
	// FIXME copy for last element is unnecessary
	copy(si.aliensOrder[idx:], si.aliensOrder[idx+1:])
	si.aliensOrder = si.aliensOrder[:last]
}

func (si *SerialInvasion) invadeCity(alien *Alien, city *City, evs []Event) []Event {
	if !city.Invaded {
		alien.Invade(city)
	} else {
		// if city already invaded two aliens will fight, both die and city is destroyed
		contender := si.aliens[city.Invader]
		alien.FightAt(si.aliens[city.Invader], city)
		si.m.DeleteCity(city.ID)
		si.deleteCityFromOrder(city.ID)
		evs = append(evs, NewImportantEvent(
			"%s has been destroyted by alien %d and alien %d!",
			city.Name, alien.ID, contender.ID))
	}
	return evs
}

// Valid if any alien can move.
func (si *SerialInvasion) Valid() bool {
	if si.m.Size() == 0 {
		return false
	}
	for i := range si.aliens {
		if si.aliens[i].Moves < si.maxMoves {
			return true
		}
	}
	return false
}