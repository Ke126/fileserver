start:
	go run cmd/main.go

test:
	go test ./...

lint:
	go vet ./...
