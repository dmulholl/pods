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

.PHONY: release
release: ## Builds a set of release binaries.
release: clean
	@termline grey
	GOOS=linux GOARCH=amd64 go build -o build/release/pods-linux-amd64/pods ./cmd/pods
	GOOS=linux GOARCH=arm64 go build -o build/release/pods-linux-arm64/pods ./cmd/pods
	GOOS=darwin GOARCH=amd64 go build -o build/release/pods-mac-amd64/pods ./cmd/pods
	GOOS=darwin GOARCH=arm64 go build -o build/release/pods-mac-arm64/pods ./cmd/pods
	GOOS=windows GOARCH=amd64 go build -o build/release/pods-windows-amd64/pods.exe ./cmd/pods
	GOOS=windows GOARCH=arm64 go build -o build/release/pods-windows-arm64/pods.exe ./cmd/pods
	@termline grey
	@tree build
	@termline grey
	@shasum -a 256 build/release/*/*
	@termline grey
	@mkdir -p build/zipped
	@cd build/release && zip -r ../zipped/pods-linux-amd64.zip pods-linux-amd64 > /dev/null
	@cd build/release && zip -r ../zipped/pods-linux-arm64.zip pods-linux-arm64 > /dev/null
	@cd build/release && zip -r ../zipped/pods-mac-amd64.zip pods-mac-amd64 > /dev/null
	@cd build/release && zip -r ../zipped/pods-mac-arm64.zip pods-mac-arm64 > /dev/null
	@cd build/release && zip -r ../zipped/pods-windows-amd64.zip pods-windows-amd64 > /dev/null
	@cd build/release && zip -r ../zipped/pods-windows-arm64.zip pods-windows-arm64 > /dev/null
	@tree build/zipped
	@termline grey
