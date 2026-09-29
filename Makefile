default: lint

lint:
	golangci-lint run --config .golangci.yml -v ./...

test:
	go test -v -race -cover ./...

test-docker:
	docker build -f test.dockerfile -t bc-wallet-healthcheck:test . \
		&& docker run --rm bc-wallet-healthcheck:test go test -v -race -cover ./...

.PHONY: lint test-docker