all: clean build

TIME=$(shell date +'%Y-%m-%d_%T')
GITVER=$(shell git rev-parse HEAD)
GO=go
GOFLAGS=-ldflags="-X 'github.com/equres/sec/pkg/server.GlobalSHA1Ver=$(GITVER)' -X 'github.com/equres/sec/pkg/server.GlobalBuildTime=$(TIME)'"

build:
	gofmt -w *.go
	$(GO) build $(GOFLAGS) -o sec *.go
	GOARCH=amd64 GOOS=linux $(GO) build $(GOFLAGS) -o sec.linux

run:
	./sec $(action)

clean:
	rm -rf sec sec.linux

# this should be called "stylecheck"
#	gofmt -w -l -e .

lint:
	golangci-lint run

createdb:
	docker exec -it postgres12 createdb --username=test_postgres --owner=test_postgres sec_project

migrateup:
	./sec migrate up
