default: lint

lint:
	golangci-lint run --config .golangci.yml -v ./...

.PHONY: lint