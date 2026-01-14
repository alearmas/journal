.PHONY: test race coverage coverage-check fmt lint

test:
	./scripts/test.sh

race:
	go test ./... -race -count=1

coverage:
	./scripts/coverage.sh

coverage-check:
	./scripts/coverage.sh
	./scripts/coverage_threshold.sh 80

fmt:
	go fmt ./...

# Lint opcional (si usás golangci-lint)
lint:
	golangci-lint run
