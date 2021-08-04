build:
	got build .

run:
	go run . $(action)

clean:
	gofmt -w