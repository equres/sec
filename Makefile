all: clean build

build:
	go build -o .

run:
	go run . $(action)

clean:
	gofmt .