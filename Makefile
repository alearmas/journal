.PHONY: test race coverage coverage-check fmt lint install-hooks swag server

test:
	./scripts/test.sh

race:
	go test ./... -race -count=1

coverage:
	./scripts/coverage.sh

coverage-check:
	./scripts/coverage.sh
	./scripts/coverage_threshold.sh 70

fmt:
	go fmt ./...

lint:
	golangci-lint run

install-hooks:
	./scripts/install-hooks.sh

swag:
	swag init -g cmd/server/main.go -o docs

server:
	go run ./cmd/server/...
