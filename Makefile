BINARY_NAME   ?= myapp
GO            ?= go
SQLC          ?= sqlc
GO_TEST_FLAGS ?= -v -race -coverprofile=coverage.out
GO_BUILD_FLAGS?= -ldflags="-s -w"

.PHONY: all
all: build

.PHONY: build
build: generate
	$(GO) build $(GO_BUILD_FLAGS) -o $(BINARY_NAME) .

.PHONY: install
install: generate
	$(GO) install .

.PHONY: run
run: generate
	$(GO) run .

.PHONY: test
test: generate
	$(GO) test $(GO_TEST_FLAGS) ./...

.PHONY: generate
generate: sqlc-generate

.PHONY: sqlc-generate
sqlc-generate:
	$(SQLC) generate --file script/sqlc.yaml


.PHONY: sqlc-compile
sqlc-compile:
	$(SQLC) compile --file script/sqlc.yaml

.PHONY: migrate-install
migrate-install:
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

.PHONY: deps
deps:
	$(GO) mod download
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: docker-db-run 
docker-db-run :
	chmod +x script/setup_local_postgres.sh
	./script/setup_local_postgres.sh

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              Build the project (default)"
	@echo "  build            Build the binary"
	@echo "  install          Install the binary to GOPATH/bin"
	@echo "  run              Run the application (generates code first)"
	@echo "  test             Run tests with coverage"
	@echo "  docker-db-run      Run postgres docker "
	@echo "  sqlc-generate    Run sqlc generate"
	@echo "  sqlc-compile     Compile Postgres"
	@echo "  clean            Remove build artifacts"
	@echo "  deps             Download go modules and install sqlc"
	@echo "  fmt              Format Go code"
	@echo "  lint             Run golangci-lint"