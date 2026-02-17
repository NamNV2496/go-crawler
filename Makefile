.PHONY: test test-coverage test-unit test-integration mock-gen lint fmt docker-up docker-down run-scheduler run-crawler clean

test:
	cd scheduler-service && go test -v -race -short ./...  -coverprofile=coverage.out || true
	cd crawler-service && go test -v -race -short ./...  -coverprofile=coverage.out || true

mock-gen:
	cd scheduler-service && go generate ./... || true
	cd crawler-service && go generate ./... || true

lint: 
	cd scheduler-service && golangci-lint run ./... || true;
	cd crawler-service && golangci-lint run ./... || true;

fmt:
	cd scheduler-service && go fmt ./...
	cd crawler-service && go fmt ./...
	gofmt -s -w .

vet:
	cd scheduler-service && go vet ./...
	cd crawler-service && go vet ./...

docker-build-scheduler:
	docker build -t go-crawler/scheduler:latest ./scheduler-service

docker-build-crawler:
	docker build -t go-crawler/crawler:latest ./crawler-service

docker-build-all: docker-build-scheduler docker-build-crawler

docker-compose-build:
	docker-compose -f docker-compose.services.yml build

docker-compose-logs:
	docker-compose -f docker-compose.yml -f docker-compose.services.yml logs -f

build-scheduler:
	cd scheduler-service && go build -o bin/scheduler main.go

build-crawler:
	cd crawler-service && go build -o bin/crawler main.go

build-all: build-scheduler build-crawler
