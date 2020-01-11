package invasion

import (
	"encoding/hex"
	"math/rand"
)

// GenerateMap creates random map with defined numbers of cities and routes between them.
// We count routes globally and uniquely per map. For example, route from Baz to Bam and Bam to Baz is a single route.
func GenerateMap(r *rand.Rand, cities, routes int) *Map {
	m := NewMap()

	ids := []string{}
	for i := 0; i < cities; {
		rname := make([]byte, 10)
		r.Read(rname)

		// encode to hex for some human readability
		city := NewCity(hex.EncodeToString(rname))
		if exists := m.GetCity(city.ID); exists != nil {
			// city with this name already exists
			continue
		}
		m.AddCity(city)
		ids = append(ids, city.ID)
		i++
	}

	full := map[string]struct{}{} // number of cities with all routes set
	directions := []string{east, west, north, south}
	for i := 0; i < routes; {
		from := ids[r.Intn(cities)]
		to := ids[r.Intn(cities)]
		if from == to {
			continue
		}
		direction := directions[r.Intn(4)]

		if err := m.AddRoute(from, to, direction); err == nil {
			i++
		}

		if m.RoutesSize(from) == 4 {
			full[from] = struct{}{}
		}
		if len(full) == cities {
			return m
		}
	}
	return m
}
