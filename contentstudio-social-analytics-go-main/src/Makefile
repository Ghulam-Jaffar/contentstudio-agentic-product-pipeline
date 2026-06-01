# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTIDY=$(GOCMD) mod tidy

# Directories
BIN_DIR := ./bin

# Application names
APP_SCHEDULER_ACCOUNT_FETCHER := account_fetcher
APP_SCHEDULER_URL_REFERSHER := url_refersher
APP_SERVICE_FACEBOOK_FETCHER := facebook_fetcher
APP_SERVICE_FACEBOOK_POSTS_PARSER := facebook_posts_videos_parser
APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR := facebook_immediate_processor
APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK := facebook_clickhouse_sink
APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK := instagram_clickhouse_sink
APP_SERVICE_INSTAGRAM_FETCHER := instagram_fetcher
APP_SERVICE_INSTAGRAM_POSTS_PARSER := instagram_posts_parser
APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR := instagram_immediate_processor
APP_SERVICE_LINKEDIN_FETCHER := linkedin_fetcher
APP_SERVICE_LINKEDIN_PARSER := linkedin_parser
APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK := linkedin_clickhouse_sink
APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR := linkedin_immediate_processor
APP_SERVICE_TIKTOK_FETCHER := tiktok_fetcher
APP_SERVICE_TIKTOK_PARSER := tiktok_parser
APP_SERVICE_TIKTOK_CLICKHOUSE_SINK := tiktok_clickhouse_sink
APP_SERVICE_TIKTOK_IMMEDIATE_PROCESSOR := tiktok_immediate_processor
APP_SERVICE_TWITTER_FETCHER := twitter_fetcher
APP_SERVICE_TWITTER_ANALYTICS_SINK := twitter_analytics_sink
APP_SERVICE_IMMEDIATE_WORK_API := immediate_work_api

# Unified services (in services/unified)
APP_SERVICE_UNIFIED_ACCOUNT_FETCHER := unified_account_fetcher
APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR := unified_immediate_processor

# Merged parser+sink services (replaces separate parser and sink)
APP_SERVICE_INSTAGRAM_ANALYTICS_SINK := instagram_analytics_sink
APP_SERVICE_FACEBOOK_ANALYTICS_SINK := facebook_analytics_sink
APP_SERVICE_LINKEDIN_ANALYTICS_SINK := linkedin_analytics_sink

# YouTube services
APP_SERVICE_YOUTUBE_FETCHER := youtube_fetcher
APP_SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR := youtube_immediate_processor
APP_SERVICE_YOUTUBE_ANALYTICS_SINK := youtube_analytics_sink
APP_SERVICE_TIKTOK_ANALYTICS_SINK := tiktok_analytics_sink

# Pinterest services
APP_SERVICE_PINTEREST_FETCHER := pinterest_fetcher
APP_SERVICE_PINTEREST_PARSER := pinterest_parser
APP_SERVICE_PINTEREST_IMMEDIATE_PROCESSOR := pinterest_immediate_processor
APP_SERVICE_PINTEREST_ANALYTICS_SINK := pinterest_analytics_sink

# GMB services
APP_SERVICE_GMB_FETCHER := gmb_fetcher
APP_SERVICE_GMB_ANALYTICS_SINK := gmb_analytics_sink


# ---- Linting (golangci-lint) -----------------------------------------------
GOLANGCI_LINT_VERSION ?= v1.64.8
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)
LINT_CONFIG := $(firstword $(wildcard .golangci.yml ../.golangci.yml))
LINT ?= 1

.PHONY: lint lint-tools vet check
lint-tools:
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "==> Installing golangci-lint $(GOLANGCI_LINT_VERSION)"; \
		GO111MODULE=on $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

lint: lint-tools
ifeq ($(LINT),1)
	@echo "==> Running golangci-lint"
	@if [ -n "$(LINT_CONFIG)" ]; then \
		golangci-lint run --config "$(LINT_CONFIG)" ./... ; \
	else \
		echo "!! No .golangci.yml found (checked . and ..); running with defaults"; \
		golangci-lint run ./... ; \
	fi
	@echo "==> Lint passed"
else
	@echo "==> Lint skipped (LINT=$(LINT))"
endif

# ---- Vetting (go vet) ------------------------------------------------------
vet:
	@echo "==> Running go vet..."
	$(GOCMD) vet ./...
	@echo "==> go vet passed"

# Run both vet + lint
check: vet lint
	@echo "==> All checks passed ✅"

# ---- Testing ---------------------------------------------------------------
GOTEST=$(GOCMD) test
COVERAGE_DIR := ./coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html
MIN_COVERAGE ?= 60

.PHONY: test test-cover test-cover-html test-race test-verbose test-short
.PHONY: test-facebook test-instagram test-linkedin test-tiktok test-twitter test-pinterest test-unified
.PHONY: test-clients test-models test-kafka test-db

# Run all tests
test:
	@echo "==> Running all tests..."
	$(GOTEST) -count=1 ./...
	@echo "==> All tests passed ✅"

# Run tests with coverage
test-cover:
	@echo "==> Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -count=1 -coverprofile=$(COVERAGE_FILE) ./...
	@echo "==> Coverage report generated: $(COVERAGE_FILE)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | tail -1

# Run tests with coverage and generate HTML report
test-cover-html: test-cover
	@echo "==> Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "==> HTML coverage report: $(COVERAGE_HTML)"

# Run tests with race detector
test-race:
	@echo "==> Running tests with race detector..."
	$(GOTEST) -count=1 -race ./...
	@echo "==> Race detection tests passed ✅"

# Run tests with verbose output
test-verbose:
	@echo "==> Running tests (verbose)..."
	$(GOTEST) -count=1 -v ./...

# Run short tests only (skip long-running tests)
test-short:
	@echo "==> Running short tests..."
	$(GOTEST) -count=1 -short ./...
	@echo "==> Short tests passed ✅"

# ---- Platform-specific tests ------------------------------------------------

# Test Facebook services
test-facebook:
	@echo "==> Running Facebook service tests..."
	$(GOTEST) -count=1 -cover ./services/facebook/...
	@echo "==> Facebook tests passed ✅"

# Test Instagram services
test-instagram:
	@echo "==> Running Instagram service tests..."
	$(GOTEST) -count=1 -cover ./services/instagram/...
	@echo "==> Instagram tests passed ✅"

# Test LinkedIn services
test-linkedin:
	@echo "==> Running LinkedIn service tests..."
	$(GOTEST) -count=1 -cover ./services/linkedin/...
	@echo "==> LinkedIn tests passed ✅"

# Test TikTok services
test-tiktok:
	@echo "==> Running TikTok service tests..."
	$(GOTEST) -count=1 -cover ./services/tiktok/...
	@echo "==> TikTok tests passed ✅"

# Test Twitter services
test-twitter:
	@echo "==> Running Twitter service tests..."
	$(GOTEST) -count=1 -cover ./services/twitter/...
	@echo "==> Twitter tests passed ✅"

# Test unified services
test-unified:
	@echo "==> Running unified service tests..."
	$(GOTEST) -count=1 -cover ./services/unified/...
	@echo "==> Unified tests passed ✅"

# Test Pinterest services
test-pinterest:
	@echo "==> Running Pinterest service tests..."
	$(GOTEST) -count=1 -cover ./services/pinterest/...
	@echo "==> Pinterest tests passed ✅"

# Test GMB services
test-gmb:
	@echo "==> Running GMB service tests..."
	$(GOTEST) -count=1 -cover ./services/gmb/...
	@echo "==> GMB tests passed ✅"

# ---- Component tests --------------------------------------------------------

# Test API clients
test-clients:
	@echo "==> Running client tests..."
	$(GOTEST) -count=1 -cover ./clients/...
	@echo "==> Client tests passed ✅"

# Test models
test-models:
	@echo "==> Running model tests..."
	$(GOTEST) -count=1 -cover ./models/...
	@echo "==> Model tests passed ✅"

# Test Kafka components
test-kafka:
	@echo "==> Running Kafka tests..."
	$(GOTEST) -count=1 -cover ./kafka/...
	@echo "==> Kafka tests passed ✅"

# Test database components
test-db:
	@echo "==> Running database tests..."
	$(GOTEST) -count=1 -cover ./db/...
	@echo "==> Database tests passed ✅"

# ---- CI/CD target -----------------------------------------------------------

# Run full CI checks: vet, lint, and tests with coverage
ci: check test-cover
	@echo "==> CI checks completed ✅"

# Run full CI with race detection (slower but thorough)
ci-full: check test-race test-cover
	@echo "==> Full CI checks completed ✅"

# ---------------------------------------------------------------------------

# Find all .go files
GO_SOURCE_FILES := $(shell find . -name '*.go' ! -path './vendor/*')

# Default target
all: build

# Build all applications (runs vet + lint first)
build: check $(SCHEDULER_ACCOUNT_FETCHER_BIN) $(SCHEDULER_URL_REFERSHER_BIN) $(SERVICE_FACEBOOK_FETCHER_BIN) $(SERVICE_INSTAGRAM_FETCHER_BIN) $(SERVICE_LINKEDIN_FETCHER_BIN) $(SERVICE_FACEBOOK_POSTS_PARSER_BIN) $(SERVICE_INSTAGRAM_POSTS_PARSER_BIN) $(SERVICE_LINKEDIN_PARSER_BIN) $(SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_BIN) $(SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_BIN) $(SERVICE_FACEBOOK_CLICKHOUSE_SINK_BIN) $(SERVICE_INSTAGRAM_CLICKHOUSE_SINK_BIN) $(SERVICE_LINKEDIN_CLICKHOUSE_SINK_BIN) $(SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_BIN) $(SERVICE_TIKTOK_FETCHER_BIN) $(SERVICE_TIKTOK_PARSER_BIN) $(SERVICE_TIKTOK_CLICKHOUSE_SINK_BIN) $(SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_BIN) $(SERVICE_TWITTER_FETCHER_BIN) $(SERVICE_TWITTER_ANALYTICS_SINK_BIN) $(SERVICE_IMMEDIATE_WORK_API_BIN) $(SERVICE_LISTENING_SCHEDULER_BIN) $(SERVICE_LISTENING_CONSUMER_BIN) $(SERVICE_LISTENING_API_BIN)
	@echo "All specified binaries built successfully in $(BIN_DIR)/"

# Full paths to application main.go files and output binaries
SCHEDULER_URL_REFERSHER_MAIN := cmd/url-refresher/main.go
SCHEDULER_URL_REFERSHER_BIN := $(BIN_DIR)/$(APP_SCHEDULER_URL_REFERSHER)

SCHEDULER_ACCOUNT_FETCHER_MAIN := cmd/main.go
SCHEDULER_ACCOUNT_FETCHER_BIN := $(BIN_DIR)/$(APP_SCHEDULER_ACCOUNT_FETCHER)

SERVICE_FACEBOOK_FETCHER_MAIN := cmd/services/facebook-fetcher/main.go
SERVICE_FACEBOOK_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_FACEBOOK_FETCHER)

SERVICE_FACEBOOK_POSTS_PARSER_MAIN := cmd/services/facebook-posts-parser/main.go
SERVICE_FACEBOOK_POSTS_PARSER_BIN := $(BIN_DIR)/$(APP_SERVICE_FACEBOOK_POSTS_PARSER)

SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_MAIN := cmd/services/facebook-immediate-processor
SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/$(APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR)

SERVICE_FACEBOOK_CLICKHOUSE_SINK_MAIN := cmd/services/facebook-clickhouse-sink
SERVICE_FACEBOOK_CLICKHOUSE_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK)

SERVICE_INSTAGRAM_CLICKHOUSE_SINK_MAIN := cmd/services/instagram-clickhouse-sink
SERVICE_INSTAGRAM_CLICKHOUSE_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK)

SERVICE_INSTAGRAM_FETCHER_MAIN := cmd/services/instagram-fetcher/main.go
SERVICE_INSTAGRAM_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_INSTAGRAM_FETCHER)

SERVICE_INSTAGRAM_POSTS_PARSER_MAIN := cmd/services/instagram-posts-parser/main.go
SERVICE_INSTAGRAM_POSTS_PARSER_BIN := $(BIN_DIR)/$(APP_SERVICE_INSTAGRAM_POSTS_PARSER)

SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_MAIN := cmd/services/instagram-immediate-processor
SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/$(APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR)

SERVICE_LINKEDIN_FETCHER_MAIN := cmd/services/linkedin-fetcher/main.go
SERVICE_LINKEDIN_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_LINKEDIN_FETCHER)

SERVICE_LINKEDIN_PARSER_MAIN := cmd/services/linkedin-parser/main.go
SERVICE_LINKEDIN_PARSER_BIN := $(BIN_DIR)/$(APP_SERVICE_LINKEDIN_PARSER)

SERVICE_LINKEDIN_CLICKHOUSE_SINK_MAIN := cmd/services/linkedin-clickhouse-sink
SERVICE_LINKEDIN_CLICKHOUSE_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK)

SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_MAIN := cmd/services/linkedin-immediate-processor
SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/$(APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR)

SERVICE_TIKTOK_FETCHER_BIN := $(BIN_DIR)/tiktok_fetcher
SERVICE_TIKTOK_PARSER_BIN := $(BIN_DIR)/tiktok_parser
SERVICE_TIKTOK_CLICKHOUSE_SINK_BIN := $(BIN_DIR)/tiktok_clickhouse_sink
SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/tiktok_immediate_processor
SERVICE_TWITTER_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_TWITTER_FETCHER)
SERVICE_TWITTER_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_TWITTER_ANALYTICS_SINK)
SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_MAIN := cmd/services/tiktok-immediate-processor

SERVICE_IMMEDIATE_WORK_API_MAIN := cmd/services/immediate-work-api
SERVICE_IMMEDIATE_WORK_API_BIN := $(BIN_DIR)/$(APP_SERVICE_IMMEDIATE_WORK_API)

# Build the account_fetcher scheduler
$(SCHEDULER_ACCOUNT_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SCHEDULER_ACCOUNT_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SCHEDULER_ACCOUNT_FETCHER_BIN) ./cmd/jobs
	@echo "==> $(APP_SCHEDULER_ACCOUNT_FETCHER) built: $(SCHEDULER_ACCOUNT_FETCHER_BIN)"

# Build the url_refresher scheduler
$(SCHEDULER_URL_REFERSHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SCHEDULER_URL_REFERSHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SCHEDULER_URL_REFERSHER_BIN) ./cmd/jobs
	@echo "==> $(SCHEDULER_URL_REFERSHER) built: $(SCHEDULER_URL_REFERSHER_BIN)"


# Build the facebook_fetcher service
$(SERVICE_FACEBOOK_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_FACEBOOK_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_FACEBOOK_FETCHER_BIN) ./services/facebook/facebook-fetcher
	@echo "==> $(APP_SERVICE_FACEBOOK_FETCHER) built: $(SERVICE_FACEBOOK_FETCHER_BIN)"

# Build the instagram_fetcher service
$(SERVICE_INSTAGRAM_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_INSTAGRAM_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_INSTAGRAM_FETCHER_BIN) ./services/instagram/instagram-fetcher
	@echo "==> $(APP_SERVICE_INSTAGRAM_FETCHER) built: $(SERVICE_INSTAGRAM_FETCHER_BIN)"

# Build the linkedin_fetcher service
$(SERVICE_LINKEDIN_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_LINKEDIN_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_LINKEDIN_FETCHER_BIN) ./services/linkedin/linkedin-fetcher
	@echo "==> $(APP_SERVICE_LINKEDIN_FETCHER) built: $(SERVICE_LINKEDIN_FETCHER_BIN)"

# Build the facebook_posts_parser service
$(SERVICE_FACEBOOK_POSTS_PARSER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_FACEBOOK_POSTS_PARSER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_FACEBOOK_POSTS_PARSER_BIN) ./services/facebook/facebook-posts-parser/main.go
	@echo "==> $(APP_SERVICE_FACEBOOK_POSTS_PARSER) built: $(SERVICE_FACEBOOK_POSTS_PARSER_BIN)"

# Build the instagram_posts_parser service
$(SERVICE_INSTAGRAM_POSTS_PARSER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_INSTAGRAM_POSTS_PARSER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_INSTAGRAM_POSTS_PARSER_BIN) ./services/instagram/instagram-posts-parser
	@echo "==> $(APP_SERVICE_INSTAGRAM_POSTS_PARSER) built: $(SERVICE_INSTAGRAM_POSTS_PARSER_BIN)"

# Build the linkedin_parser service
$(SERVICE_LINKEDIN_PARSER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_LINKEDIN_PARSER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_LINKEDIN_PARSER_BIN) ./services/linkedin/linkedin-parser
	@echo "==> $(APP_SERVICE_LINKEDIN_PARSER) built: $(SERVICE_LINKEDIN_PARSER_BIN)"

# Build the facebook_immediate_processor service
$(SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_BIN) ./services/facebook/facebook-immediate-processor
	@echo "==> $(APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR) built: $(SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_BIN)"

# Build the instagram_immediate_processor service
$(SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_BIN) ./services/instagram/instagram-immediate-processor
	@echo "==> $(APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR) built: $(SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_BIN)"

# Build the linkedin_immediate_processor service
$(SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_BIN) ./services/linkedin/linkedin-immediate-processor
	@echo "==> $(APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR) built: $(SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_BIN)"

# Build the facebook_clickhouse_sink service
$(SERVICE_FACEBOOK_CLICKHOUSE_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_FACEBOOK_CLICKHOUSE_SINK_BIN) ./services/facebook/facebook-clickhouse-sink
	@echo "==> $(APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK) built: $(SERVICE_FACEBOOK_CLICKHOUSE_SINK_BIN)"

# Build the instagram_clickhouse_sink service
$(SERVICE_INSTAGRAM_CLICKHOUSE_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_INSTAGRAM_CLICKHOUSE_SINK_BIN) ./services/instagram/instagram-clickhouse-sink
	@echo "==> $(APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK) built: $(SERVICE_INSTAGRAM_CLICKHOUSE_SINK_BIN)"

# Build the linkedin_clickhouse_sink service
$(SERVICE_LINKEDIN_CLICKHOUSE_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_LINKEDIN_CLICKHOUSE_SINK_BIN) ./services/linkedin/linkedin-clickhouse-sink
	@echo "==> $(APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK) built: $(SERVICE_LINKEDIN_CLICKHOUSE_SINK_BIN)"

# Build the tiktok_fetcher service
$(SERVICE_TIKTOK_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TIKTOK_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_TIKTOK_FETCHER_BIN) ./services/tiktok/tiktok-fetcher
	@echo "==> $(APP_SERVICE_TIKTOK_FETCHER) built: $(SERVICE_TIKTOK_FETCHER_BIN)"

# Build the tiktok_parser service
$(SERVICE_TIKTOK_PARSER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TIKTOK_PARSER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_TIKTOK_PARSER_BIN) ./services/tiktok/tiktok-parser
	@echo "==> $(APP_SERVICE_TIKTOK_PARSER) built: $(SERVICE_TIKTOK_PARSER_BIN)"

# Build the tiktok_immediate_processor service
$(SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TIKTOK_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_BIN) ./services/tiktok/tiktok-immediate-processor
	@echo "==> $(APP_SERVICE_TIKTOK_IMMEDIATE_PROCESSOR) built: $(SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_BIN)"

# Build the tiktok_clickhouse_sink service
$(SERVICE_TIKTOK_CLICKHOUSE_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TIKTOK_CLICKHOUSE_SINK)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_TIKTOK_CLICKHOUSE_SINK_BIN) ./services/tiktok/tiktok-clickhouse-sink
	@echo "==> $(APP_SERVICE_TIKTOK_CLICKHOUSE_SINK) built: $(SERVICE_TIKTOK_CLICKHOUSE_SINK_BIN)"

# Build the immediate_work_api service
$(SERVICE_IMMEDIATE_WORK_API_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_IMMEDIATE_WORK_API)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_IMMEDIATE_WORK_API_BIN) ./cmd/api-server
	@echo "==> $(APP_SERVICE_IMMEDIATE_WORK_API) built: $(SERVICE_IMMEDIATE_WORK_API_BIN)"

# Build the twitter_fetcher service
$(SERVICE_TWITTER_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TWITTER_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_TWITTER_FETCHER_BIN) ./services/twitter/twitter-fetcher
	@echo "==> $(APP_SERVICE_TWITTER_FETCHER) built: $(SERVICE_TWITTER_FETCHER_BIN)"

# Build the twitter_analytics_sink service
$(SERVICE_TWITTER_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TWITTER_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	@if [ -f .env ]; then echo "==> Sourcing .env file..."; set -a; source .env; set +a; fi
	$(GOBUILD) -o $(SERVICE_TWITTER_ANALYTICS_SINK_BIN) ./services/twitter/twitter-analytics-sink
	@echo "==> $(APP_SERVICE_TWITTER_ANALYTICS_SINK) built: $(SERVICE_TWITTER_ANALYTICS_SINK_BIN)"

# Tidy module dependencies
tidy:
	@echo "==> Tidying module dependencies..."
	$(GOTIDY)
	@echo "==> Modules tidied."

# Clean build artifacts
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -rf $(COVERAGE_DIR)
	@echo "==> Clean complete."

# Help target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  all                  - Build all applications (default)"
	@echo "  build                - Build all applications"
	@echo "  build-unified        - Build unified services"
	@echo "  build-analytics-sink - Build analytics sink services"
	@echo "  build-twitter        - Build Twitter services"
	@echo "  build-youtube        - Build YouTube services"
	@echo "  build-pinterest      - Build Pinterest services"
	@echo "  build-analytics-api  - Build Analytics API server"
	@echo ""
	@echo "Quality targets:"
	@echo "  check                - Run vet and lint"
	@echo "  vet                  - Run go vet"
	@echo "  lint                 - Run golangci-lint"
	@echo ""
	@echo "Test targets:"
	@echo "  test                 - Run all tests"
	@echo "  test-cover           - Run tests with coverage report"
	@echo "  test-cover-html      - Generate HTML coverage report"
	@echo "  test-race            - Run tests with race detector"
	@echo "  test-verbose         - Run tests with verbose output"
	@echo "  test-short           - Run short tests only"
	@echo ""
	@echo "Platform test targets:"
	@echo "  test-facebook        - Run Facebook service tests"
	@echo "  test-instagram       - Run Instagram service tests"
	@echo "  test-linkedin        - Run LinkedIn service tests"
	@echo "  test-tiktok          - Run TikTok service tests"
	@echo "  test-twitter         - Run Twitter service tests"
	@echo "  test-youtube         - Run YouTube service tests"
	@echo "  test-pinterest       - Run Pinterest service tests"
	@echo "  test-unified         - Run unified service tests"
	@echo ""
	@echo "Component test targets:"
	@echo "  test-clients         - Run API client tests"
	@echo "  test-models          - Run model tests"
	@echo "  test-kafka           - Run Kafka tests"
	@echo "  test-db              - Run database tests"
	@echo ""
	@echo "CI/CD targets:"
	@echo "  ci                   - Run vet, lint, and tests with coverage"
	@echo "  ci-full              - Run full CI with race detection"
	@echo ""
	@echo "Utility targets:"
	@echo "  tidy                 - Tidy Go module dependencies"
	@echo "  clean                - Remove build artifacts (bin directory)"
	@echo "  help                 - Show this help message"
	@echo ""
	@echo "Individual service build targets:"
	@echo "  $(APP_SCHEDULER_ACCOUNT_FETCHER)"
	@echo "  $(APP_SERVICE_FACEBOOK_FETCHER)"
	@echo "  $(APP_SERVICE_FACEBOOK_POSTS_PARSER)"
	@echo "  $(APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK)"
	@echo "  $(APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK)"
	@echo "  $(APP_SERVICE_INSTAGRAM_POSTS_PARSER)"
	@echo "  $(APP_SERVICE_INSTAGRAM_FETCHER)"
	@echo "  $(APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_LINKEDIN_FETCHER)"
	@echo "  $(APP_SERVICE_LINKEDIN_PARSER)"
	@echo "  $(APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK)"
	@echo "  $(APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_TIKTOK_FETCHER)"
	@echo "  $(APP_SERVICE_TIKTOK_PARSER)"
	@echo "  $(APP_SERVICE_TIKTOK_CLICKHOUSE_SINK)"
	@echo "  $(APP_SERVICE_TIKTOK_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_TWITTER_FETCHER)"
	@echo "  $(APP_SERVICE_TWITTER_ANALYTICS_SINK)"
	@echo "  $(APP_SERVICE_IMMEDIATE_WORK_API)"
	@echo "  $(APP_SERVICE_UNIFIED_ACCOUNT_FETCHER)"
	@echo "  $(APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_YOUTUBE_FETCHER)"
	@echo "  $(APP_SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_YOUTUBE_ANALYTICS_SINK)"
	@echo "  $(APP_SERVICE_PINTEREST_FETCHER)"
	@echo "  $(APP_SERVICE_PINTEREST_PARSER)"
	@echo "  $(APP_SERVICE_PINTEREST_IMMEDIATE_PROCESSOR)"
	@echo "  $(APP_SERVICE_PINTEREST_ANALYTICS_SINK)"
	@echo "  $(APP_SERVICE_GMB_FETCHER)"
	@echo "  $(APP_SERVICE_GMB_ANALYTICS_SINK)"

# Explicit targets for building applications directly
$(APP_SCHEDULER_ACCOUNT_FETCHER): $(SCHEDULER_ACCOUNT_FETCHER_BIN)
$(APP_SERVICE_FACEBOOK_FETCHER): $(SERVICE_FACEBOOK_FETCHER_BIN)
$(APP_SERVICE_INSTAGRAM_FETCHER): $(SERVICE_INSTAGRAM_FETCHER_BIN)
$(APP_SERVICE_LINKEDIN_FETCHER): $(SERVICE_LINKEDIN_FETCHER_BIN)
$(APP_SERVICE_FACEBOOK_POSTS_PARSER): $(SERVICE_FACEBOOK_POSTS_PARSER_BIN)
$(APP_SERVICE_INSTAGRAM_POSTS_PARSER): $(SERVICE_INSTAGRAM_POSTS_PARSER_BIN)
$(APP_SERVICE_LINKEDIN_PARSER): $(SERVICE_LINKEDIN_PARSER_BIN)
$(APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR): $(SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR_BIN)
$(APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR): $(SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR_BIN)
$(APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK): $(SERVICE_FACEBOOK_CLICKHOUSE_SINK_BIN)
$(APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK): $(SERVICE_INSTAGRAM_CLICKHOUSE_SINK_BIN)
$(APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK): $(SERVICE_LINKEDIN_CLICKHOUSE_SINK_BIN)
$(APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR): $(SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR_BIN)
$(APP_SERVICE_TIKTOK_FETCHER): $(SERVICE_TIKTOK_FETCHER_BIN)
$(APP_SERVICE_TIKTOK_PARSER): $(SERVICE_TIKTOK_PARSER_BIN)
$(APP_SERVICE_TIKTOK_CLICKHOUSE_SINK): $(SERVICE_TIKTOK_CLICKHOUSE_SINK_BIN)
$(APP_SERVICE_TIKTOK_IMMEDIATE_PROCESSOR): $(SERVICE_TIKTOK_IMMEDIATE_PROCESSOR_BIN)
$(APP_SERVICE_TWITTER_FETCHER): $(SERVICE_TWITTER_FETCHER_BIN)
$(APP_SERVICE_TWITTER_ANALYTICS_SINK): $(SERVICE_TWITTER_ANALYTICS_SINK_BIN)
$(APP_SERVICE_IMMEDIATE_WORK_API): $(SERVICE_IMMEDIATE_WORK_API_BIN)

# =====================================================
# Unified Services (services/unified)
# =====================================================

SERVICE_UNIFIED_ACCOUNT_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_UNIFIED_ACCOUNT_FETCHER)
SERVICE_UNIFIED_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/$(APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR)

# Build unified account fetcher
$(SERVICE_UNIFIED_ACCOUNT_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_UNIFIED_ACCOUNT_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_UNIFIED_ACCOUNT_FETCHER_BIN) ./services/unified/account-fetcher
	@echo "==> $(APP_SERVICE_UNIFIED_ACCOUNT_FETCHER) built: $(SERVICE_UNIFIED_ACCOUNT_FETCHER_BIN)"

# Build unified immediate processor
$(SERVICE_UNIFIED_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_UNIFIED_IMMEDIATE_PROCESSOR_BIN) ./services/unified/immediate-processor
	@echo "==> $(APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR) built: $(SERVICE_UNIFIED_IMMEDIATE_PROCESSOR_BIN)"

# Explicit targets for unified services
$(APP_SERVICE_UNIFIED_ACCOUNT_FETCHER): $(SERVICE_UNIFIED_ACCOUNT_FETCHER_BIN)
$(APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR): $(SERVICE_UNIFIED_IMMEDIATE_PROCESSOR_BIN)

# Build all unified services
build-unified: $(SERVICE_UNIFIED_ACCOUNT_FETCHER_BIN) $(SERVICE_UNIFIED_IMMEDIATE_PROCESSOR_BIN)
	@echo "All unified services built successfully"

# =====================================================
# Merged Parser+Sink Services (analytics-sink)
# =====================================================

SERVICE_INSTAGRAM_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_INSTAGRAM_ANALYTICS_SINK)
SERVICE_FACEBOOK_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_FACEBOOK_ANALYTICS_SINK)
SERVICE_LINKEDIN_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_LINKEDIN_ANALYTICS_SINK)
SERVICE_TIKTOK_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_TIKTOK_ANALYTICS_SINK)

# Build Instagram analytics sink (merged parser+sink)
$(SERVICE_INSTAGRAM_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_INSTAGRAM_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_INSTAGRAM_ANALYTICS_SINK_BIN) ./services/instagram/instagram-analytics-sink
	@echo "==> $(APP_SERVICE_INSTAGRAM_ANALYTICS_SINK) built: $(SERVICE_INSTAGRAM_ANALYTICS_SINK_BIN)"

# Build Facebook analytics sink (merged parser+sink)
$(SERVICE_FACEBOOK_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_FACEBOOK_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_FACEBOOK_ANALYTICS_SINK_BIN) ./services/facebook/facebook-analytics-sink
	@echo "==> $(APP_SERVICE_FACEBOOK_ANALYTICS_SINK) built: $(SERVICE_FACEBOOK_ANALYTICS_SINK_BIN)"

# Build LinkedIn analytics sink (merged parser+sink)
$(SERVICE_LINKEDIN_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_LINKEDIN_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_LINKEDIN_ANALYTICS_SINK_BIN) ./services/linkedin/linkedin-analytics-sink
	@echo "==> $(APP_SERVICE_LINKEDIN_ANALYTICS_SINK) built: $(SERVICE_LINKEDIN_ANALYTICS_SINK_BIN)"

# Build TikTok analytics sink (merged parser+sink)
$(SERVICE_TIKTOK_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_TIKTOK_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_TIKTOK_ANALYTICS_SINK_BIN) ./services/tiktok/tiktok-analytics-sink
	@echo "==> $(APP_SERVICE_TIKTOK_ANALYTICS_SINK) built: $(SERVICE_TIKTOK_ANALYTICS_SINK_BIN)"

# Explicit targets for merged services
$(APP_SERVICE_INSTAGRAM_ANALYTICS_SINK): $(SERVICE_INSTAGRAM_ANALYTICS_SINK_BIN)
$(APP_SERVICE_FACEBOOK_ANALYTICS_SINK): $(SERVICE_FACEBOOK_ANALYTICS_SINK_BIN)
$(APP_SERVICE_LINKEDIN_ANALYTICS_SINK): $(SERVICE_LINKEDIN_ANALYTICS_SINK_BIN)
$(APP_SERVICE_TIKTOK_ANALYTICS_SINK): $(SERVICE_TIKTOK_ANALYTICS_SINK_BIN)

# Build all merged analytics-sink services
build-analytics-sink: $(SERVICE_INSTAGRAM_ANALYTICS_SINK_BIN) $(SERVICE_FACEBOOK_ANALYTICS_SINK_BIN) $(SERVICE_LINKEDIN_ANALYTICS_SINK_BIN) $(SERVICE_TIKTOK_ANALYTICS_SINK_BIN) $(SERVICE_TWITTER_ANALYTICS_SINK_BIN)
	@echo "All analytics-sink services built successfully"

# Build all Twitter services
build-twitter: $(SERVICE_TWITTER_FETCHER_BIN) $(SERVICE_TWITTER_ANALYTICS_SINK_BIN)
	@echo "All Twitter services built successfully"

# =====================================================
# YouTube Services
# =====================================================

SERVICE_YOUTUBE_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_YOUTUBE_FETCHER)
SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/$(APP_SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR)
SERVICE_YOUTUBE_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_YOUTUBE_ANALYTICS_SINK)

# Build YouTube fetcher
$(SERVICE_YOUTUBE_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_YOUTUBE_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_YOUTUBE_FETCHER_BIN) ./services/youtube/youtube-fetcher
	@echo "==> $(APP_SERVICE_YOUTUBE_FETCHER) built: $(SERVICE_YOUTUBE_FETCHER_BIN)"

# Build YouTube immediate processor
$(SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR_BIN) ./services/youtube/youtube-immediate-processor
	@echo "==> $(APP_SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR) built: $(SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR_BIN)"

# Build YouTube analytics sink
$(SERVICE_YOUTUBE_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_YOUTUBE_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_YOUTUBE_ANALYTICS_SINK_BIN) ./services/youtube/youtube-analytics-sink
	@echo "==> $(APP_SERVICE_YOUTUBE_ANALYTICS_SINK) built: $(SERVICE_YOUTUBE_ANALYTICS_SINK_BIN)"

# Explicit targets for YouTube services
$(APP_SERVICE_YOUTUBE_FETCHER): $(SERVICE_YOUTUBE_FETCHER_BIN)
$(APP_SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR): $(SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR_BIN)
$(APP_SERVICE_YOUTUBE_ANALYTICS_SINK): $(SERVICE_YOUTUBE_ANALYTICS_SINK_BIN)

# Build all YouTube services
build-youtube: $(SERVICE_YOUTUBE_FETCHER_BIN) $(SERVICE_YOUTUBE_IMMEDIATE_PROCESSOR_BIN) $(SERVICE_YOUTUBE_ANALYTICS_SINK_BIN)
	@echo "All YouTube services built successfully"

# Test YouTube services
test-youtube:
	@echo "==> Running YouTube service tests..."
	$(GOTEST) -count=1 -cover ./services/youtube/...
	@echo "==> YouTube tests passed ✅"

# =====================================================
# Pinterest Services
# =====================================================

SERVICE_PINTEREST_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_PINTEREST_FETCHER)
SERVICE_PINTEREST_PARSER_BIN := $(BIN_DIR)/$(APP_SERVICE_PINTEREST_PARSER)
SERVICE_PINTEREST_IMMEDIATE_PROCESSOR_BIN := $(BIN_DIR)/$(APP_SERVICE_PINTEREST_IMMEDIATE_PROCESSOR)
SERVICE_PINTEREST_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_PINTEREST_ANALYTICS_SINK)

# Build Pinterest fetcher
$(SERVICE_PINTEREST_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_PINTEREST_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_PINTEREST_FETCHER_BIN) ./services/pinterest/pinterest-fetcher
	@echo "==> $(APP_SERVICE_PINTEREST_FETCHER) built: $(SERVICE_PINTEREST_FETCHER_BIN)"

# Build Pinterest parser
$(SERVICE_PINTEREST_PARSER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_PINTEREST_PARSER)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_PINTEREST_PARSER_BIN) ./services/pinterest/pinterest-parser
	@echo "==> $(APP_SERVICE_PINTEREST_PARSER) built: $(SERVICE_PINTEREST_PARSER_BIN)"

# Build Pinterest immediate processor
$(SERVICE_PINTEREST_IMMEDIATE_PROCESSOR_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_PINTEREST_IMMEDIATE_PROCESSOR)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_PINTEREST_IMMEDIATE_PROCESSOR_BIN) ./services/pinterest/pinterest-immediate-processor
	@echo "==> $(APP_SERVICE_PINTEREST_IMMEDIATE_PROCESSOR) built: $(SERVICE_PINTEREST_IMMEDIATE_PROCESSOR_BIN)"

# Build Pinterest analytics sink
$(SERVICE_PINTEREST_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_PINTEREST_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_PINTEREST_ANALYTICS_SINK_BIN) ./services/pinterest/pinterest-analytics-sink
	@echo "==> $(APP_SERVICE_PINTEREST_ANALYTICS_SINK) built: $(SERVICE_PINTEREST_ANALYTICS_SINK_BIN)"

# Explicit targets for Pinterest services
$(APP_SERVICE_PINTEREST_FETCHER): $(SERVICE_PINTEREST_FETCHER_BIN)
$(APP_SERVICE_PINTEREST_PARSER): $(SERVICE_PINTEREST_PARSER_BIN)
$(APP_SERVICE_PINTEREST_IMMEDIATE_PROCESSOR): $(SERVICE_PINTEREST_IMMEDIATE_PROCESSOR_BIN)
$(APP_SERVICE_PINTEREST_ANALYTICS_SINK): $(SERVICE_PINTEREST_ANALYTICS_SINK_BIN)

# Build all Pinterest services
build-pinterest: $(SERVICE_PINTEREST_FETCHER_BIN) $(SERVICE_PINTEREST_PARSER_BIN) $(SERVICE_PINTEREST_IMMEDIATE_PROCESSOR_BIN) $(SERVICE_PINTEREST_ANALYTICS_SINK_BIN)
	@echo "All Pinterest services built successfully"

# =====================================================
# GMB Services
# =====================================================

SERVICE_GMB_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_GMB_FETCHER)
SERVICE_GMB_ANALYTICS_SINK_BIN := $(BIN_DIR)/$(APP_SERVICE_GMB_ANALYTICS_SINK)

# Build GMB fetcher
$(SERVICE_GMB_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_GMB_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_GMB_FETCHER_BIN) ./services/gmb/gmb-fetcher
	@echo "==> $(APP_SERVICE_GMB_FETCHER) built: $(SERVICE_GMB_FETCHER_BIN)"

# Build GMB analytics sink
$(SERVICE_GMB_ANALYTICS_SINK_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_GMB_ANALYTICS_SINK)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_GMB_ANALYTICS_SINK_BIN) ./services/gmb/gmb-analytics-sink
	@echo "==> $(APP_SERVICE_GMB_ANALYTICS_SINK) built: $(SERVICE_GMB_ANALYTICS_SINK_BIN)"

# Explicit targets for GMB services
$(APP_SERVICE_GMB_FETCHER): $(SERVICE_GMB_FETCHER_BIN)
$(APP_SERVICE_GMB_ANALYTICS_SINK): $(SERVICE_GMB_ANALYTICS_SINK_BIN)

# Build all GMB services
build-gmb: $(SERVICE_GMB_FETCHER_BIN) $(SERVICE_GMB_ANALYTICS_SINK_BIN)
	@echo "All GMB services built successfully"

# Phony targets: these are not files
.PHONY: all build build-unified build-analytics-sink build-twitter clean tidy help
.PHONY: ci ci-full
.PHONY: $(APP_SCHEDULER_ACCOUNT_FETCHER) $(APP_SERVICE_FACEBOOK_FETCHER)
.PHONY: $(APP_SERVICE_INSTAGRAM_FETCHER) $(APP_SERVICE_LINKEDIN_FETCHER)
.PHONY: $(APP_SERVICE_FACEBOOK_POSTS_PARSER) $(APP_SERVICE_INSTAGRAM_POSTS_PARSER)
.PHONY: $(APP_SERVICE_LINKEDIN_PARSER) $(APP_SERVICE_FACEBOOK_IMMEDIATE_PROCESSOR)
.PHONY: $(APP_SERVICE_INSTAGRAM_IMMEDIATE_PROCESSOR) $(APP_SERVICE_FACEBOOK_CLICKHOUSE_SINK)
.PHONY: $(APP_SERVICE_INSTAGRAM_CLICKHOUSE_SINK) $(APP_SERVICE_LINKEDIN_CLICKHOUSE_SINK)
.PHONY: $(APP_SERVICE_LINKEDIN_IMMEDIATE_PROCESSOR) $(APP_SERVICE_TIKTOK_FETCHER)
.PHONY: $(APP_SERVICE_TIKTOK_PARSER) $(APP_SERVICE_TIKTOK_CLICKHOUSE_SINK)
.PHONY: $(APP_SERVICE_TIKTOK_IMMEDIATE_PROCESSOR) $(APP_SERVICE_IMMEDIATE_WORK_API)
.PHONY: $(APP_SERVICE_TWITTER_FETCHER) $(APP_SERVICE_TWITTER_ANALYTICS_SINK)
.PHONY: $(APP_SERVICE_UNIFIED_ACCOUNT_FETCHER) $(APP_SERVICE_UNIFIED_IMMEDIATE_PROCESSOR)
.PHONY: $(APP_SERVICE_INSTAGRAM_ANALYTICS_SINK) $(APP_SERVICE_FACEBOOK_ANALYTICS_SINK)
.PHONY: $(APP_SERVICE_LINKEDIN_ANALYTICS_SINK) $(APP_SERVICE_TIKTOK_ANALYTICS_SINK)
.PHONY: $(APP_SERVICE_YOUTUBE_FETCHER) $(APP_SERVICE_YOUTUBE_ANALYTICS_SINK) build-youtube test-youtube
.PHONY: $(APP_SERVICE_PINTEREST_FETCHER) $(APP_SERVICE_PINTEREST_ANALYTICS_SINK) build-pinterest test-pinterest
.PHONY: $(APP_SERVICE_GMB_FETCHER) $(APP_SERVICE_GMB_ANALYTICS_SINK) build-gmb test-gmb

# =====================================================
# Listening Services
# =====================================================

SERVICE_LISTENING_SCHEDULER_BIN := $(BIN_DIR)/listening_scheduler
SERVICE_LISTENING_CONSUMER_BIN := $(BIN_DIR)/listening_consumer
SERVICE_LISTENING_API_BIN := $(BIN_DIR)/listening_api

# Build listening scheduler
$(SERVICE_LISTENING_SCHEDULER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building listening_scheduler..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_LISTENING_SCHEDULER_BIN) ./services/listening/listening-scheduler
	@echo "==> listening_scheduler built: $(SERVICE_LISTENING_SCHEDULER_BIN)"

# Build listening consumer (fetcher -> parser -> sink + async enrichment)
$(SERVICE_LISTENING_CONSUMER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building listening_consumer..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_LISTENING_CONSUMER_BIN) ./services/listening/listening-consumer
	@echo "==> listening_consumer built: $(SERVICE_LISTENING_CONSUMER_BIN)"

# Build listening API (HTTP server for listening endpoints)
$(SERVICE_LISTENING_API_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building listening_api..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_LISTENING_API_BIN) ./services/listening/listening-api
	@echo "==> listening_api built: $(SERVICE_LISTENING_API_BIN)"

# Explicit targets
listening_scheduler: $(SERVICE_LISTENING_SCHEDULER_BIN)
listening_consumer: $(SERVICE_LISTENING_CONSUMER_BIN)
listening_api: $(SERVICE_LISTENING_API_BIN)
listening: listening_scheduler listening_consumer listening_api

.PHONY: listening_scheduler listening_consumer listening_api listening
