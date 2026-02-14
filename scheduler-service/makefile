buf:
	buf def update

generate:
	buf generate --template buf.gen.yaml 

service:
	go run main.go server

worker:
	go run main.go scheduler_worker

test:
	go test -v -race -short ./... -coverprofile=coverage.out
# 	go tool cover -html=coverage.out
