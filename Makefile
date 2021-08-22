all: clean build

build:
	go build -o secapp

run:
	./secapp $(action)

clean:
	gofmt -w -l -e .

lint:
	golangci-lint run

createdb:
	docker exec -it postgres12 createdb --username=test_postgres --owner=test_postgres sec_project

migrateup:
	./secapp migrate up --config=ci