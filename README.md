Alien invasion challenge
========================

If interested in rationale behind impementation see [design document](./DESIGN.md).

This repository provides a program to execute a repeatable simulation of world destruction.
The world map must be defined in the following format:

```
Foo123 south=Baz north=Tot-H
Tot-H east=Bar
Baz north=Foo123
Bar
```

Format restrictions:
- Each line should start with a city as a word without empty spaces, any characters execpt `=` are allowed.
- At most four directions should follow the city name, zero is fine too. Each direction should be in `key=value` format without empty spaces in the middle.
- All routes should be symmetric, in the example above if Foo123 has a Baz in the south, Baz should have Foo123 in the north. Such relationships doesn't have to be defined for every pair, the program will restore them automatically.
- There should be no conflicting routes, if Bar defines direction to Foo123 - it can't be north, as it will conflict with Baz.
- There should be no routes that

Additionally, there is a tool to generate random maps, of required size and connectivity.

How to build?
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

For example, lets take a map above and put it into `your.map` file.

Also, by default simulation will spawn 100 aliens, with 10000 moves each, but because the map is small lets
reduce that number, e.g. lets run `./build/invasion your.map -n 4 -m 7`, where `n` is number of aliens and `m` is number of moves available for each of them.

We will get the following output, in this case two cities were destroyed by 4 aliens, and 2 cities are still on the map
without routes between them.

```
Baz has been destroyted by alien 2 and alien 3!
Tot-H has been destroyted by alien 1 and alien 0!
Bar
Foo123
```

In case you want to get repeatable result from simulation you can provide `-seed` argument, accepts any 64 bit signed integer.
Like `/build/invasion -n 4 -m 7 -seed=1 rst.out`, will output:

```
Tot-H has been destroyted by alien 3 and alien 1!
Bar
Baz north=Foo123
Foo123 south=Baz
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
