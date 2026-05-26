BINARY_NAME 	= kivi
BIN_DIR 		= bin
OUTPUT 			= $(BIN_DIR)/$(BINARY_NAME)
CMD_PATH 		= ./cmd/kivi

# Go flags
GO				= go
GO_FLAGS 		= -v 
# -v 			: VERBOSE: enables detailed output logging across various commands
TEST_FLAGS 		= -count=1 -timeout=60s
# -count=1 		: run the test once and do not cache the result
# 				  count flag defines the number of times to run the test
#	 		      it overrides the cache result because Cache.Disable is set to true 
# 				  internally by Go's logic when count flag is set
# -timeout=60s	: sets a timeout for tests to run till 60s and not indefinately
BENCH_FLAGS 	= -bench=. -benchmem -benchtime=3s -run=^$
# -bench=.		: run all benchmarks
# -benchmem		: report memory allocations
# -benchtime=3s	: run each benchmark for 3 seconds to get more stable results
# -run=^$		: do not run any tests when running benchmarks 
# 				  ^$ means empty string (no test function name could be empty string), 
#			      but it cannot filter out benchmark tests as they are governed by the -bench flag

all: build

## build: compile the binary into bin/
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GO_FLAGS) -o $(OUTPUT) $(CMD_PATH)

## test: run all tests (no race detector)
test:
	$(GO) test $(TEST_FLAGS) ./...

## bench: run all benchmarks, skip unit tests
bench:
	$(GO) test $(BENCH_FLAGS) ./...

## lint: run golangci-lint (install: brew install golangci-lint)
lint:
	golangci-lint run ./...

## race: run all tests with the race detector enabled
race:
	$(GO) test $(TEST_FLAGS) -race ./...

## clean: remove compiled binaries
clean:
	@rm -rf $(BIN_DIR)

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  make /'