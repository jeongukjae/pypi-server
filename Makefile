all: help

.PHONY: generate
## generate: generate code
generate:
	go generate ./...

.PHONY: format
## format: format the source code
format:
	go tool golang.org/x/tools/cmd/goimports -local github.com/jeongukjae/pypi-server -w .

.PHONY: lint
## lint: lint the source code
lint:
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run -c .golangci.yaml

.PHONY: test
## test: run tests
test:
	go tool github.com/rakyll/gotest ./...

.PHONY: dev-env-up
## dev-env-up: bring up the development environment using Docker Compose
dev-env-up:
	docker-compose up -d

.PHONY: dev-env-down
## dev-env-down: bring down the development environment using Docker Compose
dev-env-down:
	docker-compose down -v

.PHONY: new-migration
## new-migration: create a new database migration
new-migration:
	go tool github.com/jackc/tern/v2 new --migrations=./queries/migrations new-migration

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'
