all: clean sec sec.linux

TIME=$(shell date +'%Y-%m-%d_%T')
GITVER=$(shell git rev-parse HEAD)
GO=go
GOFLAGS=-ldflags="-X 'github.com/equres/sec/pkg/server.GlobalSHA1Ver=$(GITVER)' -X 'github.com/equres/sec/pkg/server.GlobalBuildTime=$(TIME)'"
DEPLOY_HOST=$(cat .host.conf)

sec: fmt
	$(GO) build $(GOFLAGS) -o sec *.go

sec.linux: fmt
	GOARCH=amd64 GOOS=linux $(GO) build $(GOFLAGS) -o sec.linux

fmt: lint
	gofmt -w *.go

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

run:
	./sec $(action)

clean:
	rm -rf sec sec.linux

migrateup:
	./sec migrate up

deploy:
	set -x
	set -e
	uname -a
	pwd
	ls -la
	whoami

	echo "New program MD5"
	openssl md5 sec.linux
	echo "Old program MD5"
	ssh sec@equres.com 'if [ -e sec ]; then openssl md5 sec; fi'

	date
	echo "Uploading the new binary file in progress"
	xz - < ./sec.linux | ssh sec@equres.com 'unxz - > /home/sec/sec.new--inprogress'
	date

	echo "Rotating binaries (sec -> sec.old, sec.new--inprogress -> sec.new)"
	ssh sec@equres.com 'mv sec sec.old; mv sec.new--inprogress sec.new; chmod 755 sec.new sec.old'

