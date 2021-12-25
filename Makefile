build:
	go build -o bin/main main.go

test:
	go test -v ./...

run:
	./bin/main