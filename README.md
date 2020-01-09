Alien invasion challenge
========================

This repository provides a program to execute a repeatable simulation of world destruction.
The world must be defined in the next format:

```
Foo123 south=Baz north=Tot-H
Tot-H east=Bar
Baz north=Foo123
Bar
```

Each line should start with a city as a word without empty spaces, any characters except `=` are allowed.
At most four directions should follow the city name, zero is fine too.
Each direction should be in `key=value` format without empty spaces in the middle.
Directions should be symmetric, e.g. if Foo123 has a Baz in the south, Baz should have Foo123 in the north. Such relationships
doesn't have to be defined for every pair, the program will restore them automatically.
If Bar defines direction to Foo123 - it can't be north, as it will conflict with Baz.


Additionally, there is a tool to generate random maps, with required parameters.

How to build
---

Install [go 1.13](https://golang.org/dl/).

Run `make build`. It will create two binaries in `build` directory.

How to run a simulation?
---

Execute `./build/invasion --help` to see all available options. But if you just want to simulate with a world,
and print updated world to stdout, you can simply run:

```
./build/invasion your.map
```

How to generate a map?
---

To generate any random map simply use `./build/mapgen &> any.map`. You can also explore all options with `-help`.
In general, it allows for generating a map of the desired size and connectivity.

Tests
---

Unit tests can be executed with `make test`.

To generate maps of different sizes and run a simulation with them use `make run-maps`.
It will create files in `_assets` directory,  `rst` prefix will be used for versions after simulation.


Possible extensions
---

The very practical extension is to allow aliens to make progress concurrently and this is something that should be supported. While it is quite possible to achieve by locking shared state and spawning multiple workers, we will lose repeatability of simulation.
