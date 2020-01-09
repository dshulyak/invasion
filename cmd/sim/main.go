package main

import (
	"bufio"
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dshulyak/invasion"
)

var (
	aliens = flag.Int("n", 100, "number of aliens that invade the world")
	moves  = flag.Int("m", 10000, "max number of moves every alien can make")
	seed   = flag.Int64("seed", 0, "if not zero will be used for simulation")
	out    = flag.String("out", "", "after simulation map will be saved to this fail, otherwise printed to stdout. file will be truncated.")
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatalf("program expect second argument to be a file")
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

	use := time.Now().UnixNano()
	if *seed != 0 {
		use = *seed
	}

	invasion := invasion.NewSerialInvasion(
		m, rand.New(rand.NewSource(use)),
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
