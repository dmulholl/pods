.PHONY: help
help: ## Prints available make commands.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / \
	{printf "\033[1;36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Builds the application binary.
	go build -o ./build/pods ./cmd/pods

.PHONY: run
run: ## Builds and runs the application binary.
	go run ./cmd/pods

.PHONY: install
install: ## Builds and installs the application binary.
	go install ./cmd/pods

.PHONY: test
test: ## Runs unit tests.
	go test ./...

.PHONY: test-verbose
test-verbose: ## Runs unit tests verbosely.
	go test ./... -v

.PHONY: clean
clean: ## Deletes all build artifacts.
	@rm -rf ./build
