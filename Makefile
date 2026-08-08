.PHONY: build run dev test lint fmt vet clean docker docker-up docker-down

BINARY := karthub
MAIN := ./cmd/server

build:
	CGO_ENABLED=1 go build -o bin/$(BINARY) $(MAIN)

run: build
	./bin/$(BINARY)

dev:
	air

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/ data/

docker:
	docker build -t karthub .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

seed:
	go run ./cmd/seed
