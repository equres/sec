all: clean build

build:
	go build

run:
	./sec $(action)

clean:
	gofmt -w -l -e .