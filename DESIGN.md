Simulation design
=================

## Goal

The goal of the implementation was to see how aliens will make progress if they would be
acting concurrently, but make sure that simulation is repeatable if the user wants it.

To achieve the first part of the goal, simulation picks each alien randomly whenever it makes progress.
For example in a simulation with 2 aliens, the first alien may make 10 moves before the second alien will
invade the first city, all depends on numbers produced by randomness source.

We want the simulation to be repeatable, therefore all randomness must be coming from a single source.
In our case, this source is `math.Rand` pseudorandomness. Particularly we need to exclude non-determinism caused
by the scheduler or not ordered data structures (e.g. hash tables).

## State

#### Map

Map implements an undirected graph, with known maximum number of edges. Nodes in the graph are `City` and the edge is `Route`.
Each edge has a label, which is one of the cardinal directions (north, south, east, west). Note that is an edge to the city
hash a label `north`, same edge from the city will have a label `south`.

```go
type Map struct {
        cities map[string]*City
        routes map[string][]Route
}
```

#### Routes

Routes, edges, are implemented as `map[string][]Route` instead of a more common `map[string]map[string]Route` to guarantee determinism when a random route is retrieved.

It allows to make random city retrieval to be very simple:

```go
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
```

An important part of the simulation is to maintain correctness of the routing table, for this `DeleteRoutes` must not only
remove all outgoing edges, but also remove incoming edges as well.


#### City

The city itself has the following definition in golang:

```go
type City struct {
        ID string
        Name string
        Invaded bool
        Invader int
        Destroyed bool
}
```

ID is lowercased Name, to simplify validation, but still ensure that we won't have duplicates on the map with minor differences in the Name.

Invader is a back-reference to alien ID that currently invades the city. Only valid if Invaded is true.

Cities are stored as a regular map, and therefore additional measures will be needed to ensure that we can work with
them deterministically (described later).

#### Alien

Defined in golang as:

```go
type Alien struct {
        ID int
        Moves int
        Dead bool
        Location string
        Trapped  bool
}
```

Valid state transitions for aliens and related city are:
- the alien can't invade a city, before leaving currently invaded city
- if the alien is trapped city should have no routes
- for the alien to die, another alien must die and the city be destroyed
- moves can't grow larger than a global `maxMoves` parameter

Collections of aliens is represented as `map[int]*Alien`.

## Simulation

#### Additional state

As was noted earlier cities and aliens are stored in non-ordered maps. Thus to guarantee that maps non-determinism
won't interfere with simulation randomness two auxiliary slices were created. Slices must be sorted, so that the initial order is the same between different executions of the simulation.

#### Algorithm

Simulation runs until there is an alien with moves lower than defined `maxMoves` parameter and the map is not empty.
The simulation will have at most `aliensNumber*maxMoves` steps.

Each step follows the next sequence of instructions:

- Pick a random alien from an ordered pool of aliens (pick index from the pool, and then get alien from the map).
- Increment number of aliens moves.
- If an alien invades the world for the first time (`Location` is empty), pick a city from ordered cities pool.
  Go to Invade city routine (below).
- If an alien is not yet trapped or dead, and the location is not empty - pick a random city based on existing routes.
  If there are no routes mark alien as trapped, so he will be ignored in the future.
  Otherwise, leave the current city, and go to Invade city routine.

Invade city routine, needs to verify if the city already invaded or not. If it is then the new alien will fight with an alien
that currently is city invader (both will die in the process, and the city will get destroyed).
If the city is not invaded - the alien will set himself as an invader.

- At the end of the step, we will remove the alien if we observed his death in this step, in this case, alien
  is removed both from aliens collection and aliens ordered a slice.
  If alien reached max moves we will remove an alien from the ordered slice, so that we won't pick him anymore,
  but Alien object may still be useful, e.g. if another alien invades city where original alien, with exhausted moves,
  ended up.
