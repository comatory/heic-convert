build:
  go build -o bin/heic-to-jpeg ./cmd/heic-to-jpeg/main.go

lint:
  golangci-lint run ./...

format:
  go fmt ./...
