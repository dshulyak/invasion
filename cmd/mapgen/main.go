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
	cities = flag.Int("c", 100, "number of cities in the random map")
	routes = flag.Int("r", 50, "number of unique routes in the random map")
	out    = flag.String("out", "", "if provided, map will be saved to a file, otherwise printed to stdout. file will be truncated.")
	seed   = flag.Int64("seed", 0, "if non zero seed will be used for map generation")
)

func main() {
	flag.Parse()

	use := time.Now().UnixNano()
	if *seed != 0 {
		use = *seed
	}
	log.Printf("using seed %d", use)

	m := invasion.GenerateMap(rand.New(rand.NewSource(use)), *cities, *routes)

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
