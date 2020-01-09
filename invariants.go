package invasion

import "fmt"

// invariants are a series of conditions that should be maintained at the each step of simulation

// VerifyInvariants verifies every known invariant.
func VerifyInvariants(m *Map, aliens []*Alien) error {
	for _, a := range aliens {
		if err := verifyAlienValidState(m, a); err != nil {
			return err
		}
	}
	if err := verifyCitiesState(m); err != nil {
		return err
	}
	return nil
}

func verifyAlienValidState(m *Map, a *Alien) error {
	// to be dead alien needs to find a path to someone else
	if a.Trapped && a.Dead {
		return fmt.Errorf("alien %d is dead and trapped at the same time", a.ID)
	}

	// if alien is trapped city can't be destroyed and it should have no routes
	if a.Trapped {
		city := m.GetCity(a.Location)
		if city == nil || city.Destroyed {
			return fmt.Errorf("city %v destroyed while alien %d trapped in it", a.Location, a.ID)
		}
		if rsize := m.RoutesSize(a.Location); rsize > 0 {
			return fmt.Errorf("alien %d trapped in the city %v with routes (%d)", a.ID, a.Location, rsize)
		}
	}

	// if dead city should not exists or atleast be destroyed
	if a.Dead {
		city := m.GetCity(a.Location)
		if city != nil && !city.Destroyed {
			return fmt.Errorf("city %v is not destroyed while alien %d died in it", a.Location, a.ID)
		}
	}

	return nil
}

func verifyCitiesState(m *Map) (err error) {

	invaders := map[int]string{}

	// alien can't invade without leaving previous city
	m.IterateCities(func(c *City, _ []Route) bool {
		if c.Invaded {
			if name, exist := invaders[c.Invader]; exist {
				err = fmt.Errorf("alien %d invaded two cities %s and %s", c.Invader, name, c.Name)
				return false
			}
			invaders[c.Invader] = c.Name
		}

		// TODO test that routes are symmetric
		return true
	})
	return
}
