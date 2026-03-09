.PHONEY: build run

build:
	@go build -o bin/bfs ./cmd/bfs

run: build
	@./bin/bfs
