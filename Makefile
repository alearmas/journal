.PHONY: test race coverage coverage-check fmt lint install-hooks

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
