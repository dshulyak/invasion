package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dshulyak/invasion"
)

var (
	aliens = flag.Int("n", 100, "number of aliens that invade the world")
	moves  = flag.Int("m", 10000, "max number of moves every alien can make")
	seed   = flag.Int64("seed", time.Now().UnixNano(), "provided seed will be used for simulation")
	// TODO replace with positional
	out = flag.String("out", "", "after simulation updated map will be saved to this file, otherwise printed to stdout. file will be truncated.")

	usage = `Run invasion simulation. Requires map file with the following format:

Foo123 south=Baz north=Tot-H
Tot-H east=Bar
Baz north=Foo123
Bar

Each line should start with a city as a word without empty spaces, any characters except '=' are allowed.
At most four directions should follow the city name, zero is fine too.
Each direction should be in <key>=<value> format without empty spaces in the middle.
Directions should be symmetric, e.g. if Foo123 has a Baz in the south, Baz should have Foo123 in the north. Such relationships
doesn't have to be specified for every pair, the program will restore them automatically.
If Bar defines direction to Foo123 - it can't be north, as it will conflict with Baz.

Usage:

sim <your.map>

Examples:
sim -out=./_assets/rst-1000-500.out ./_assets/1000-500.out
sim -seed=777 ./_assets/1000-500.out

Defaults:`
)

func main() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatalf("program expects first positional argument to be a file")
	}

	in := flag.Arg(0)
	f, err := os.OpenFile(in, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatalf("failed to open a file %s: %v", in, err)
	}
	defer f.Close()
	buf := bufio.NewReader(f)

	m := invasion.NewMap()
	_, err = m.ReadFrom(buf)
	if err != nil {
		log.Fatalf("failed to fill the map: %v", err)
	}

	invasion := invasion.NewSerialInvasion(
		m, rand.New(rand.NewSource(*seed)),
		os.Stdout, *aliens, *moves,
	)
	invasion.Run()

	// TODO deduplicate this code and code in invasion cmd
	if len(*out) > 0 {
		f, err := os.OpenFile(*out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("%v", err)
		}
		defer f.Close()
		buf := bufio.NewWriter(f) // 4mb will be allocated by default
		_, err = m.WriteTo(buf)
		if err != nil {
			log.Fatalf("failed to write map: %v", err)
		}
		if err := buf.Flush(); err != nil {
			log.Fatalf("failed to flush buffer: %v", err)
		}
		if err := f.Sync(); err != nil {
			log.Fatalf("failed to fsync: %v", err)
		}
	} else {
		_, err := m.WriteTo(os.Stdout)
		if err != nil {
			log.Fatalf("failed to print to stdout: %v", err)
		}
	}
}
