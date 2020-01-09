.PHONY: build
build:
	mkdir -p ./build/
	go build -mod=readonly -o ./build/mapgen ./cmd/mapgen
	go build -mod=readonly -o ./build/invasion ./cmd/sim

maps:
	mkdir -p ./_assets/
	./build/mapgen -c 1000 -r 500 -out=./_assets/1000-500.out
	./build/mapgen -c 10000 -r 5000 -out=./_assets/10000-5000.out
	./build/mapgen -c 100000 -r 50000 -out=./_assets/100000-50000.out

run-maps: maps
	./build/invasion -n 100 -m 10000 -out=./_assets/rst-1000-500.out ./_assets/1000-500.out
	./build/invasion -n 100 -m 10000 -out=./_assets/rst-10000-5000.out ./_assets/10000-5000.out
	./build/invasion -n 1000 -m 10000 -out=./_assets/rst-100000-50000.out ./_assets/100000-50000.out

test:
	go test ./ -cover
