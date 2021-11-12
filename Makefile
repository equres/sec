all: clean build

build:
	go build
	GOARCH=amd64 GOOS=linux go build -o sec.linux

run:
	./sec $(action)

clean:
	gofmt -w -l -e .

lint:
	golangci-lint run

createdb:
	docker exec -it postgres12 createdb --username=test_postgres --owner=test_postgres sec_project

migrateup:
	./sec migrate up
