---
title: Social Media Integration Workflow
description: Complete guide for social media platform integration and data processing
---

Moved to `docs/architecture/social-media-integration-workflow.md`.

## Critical Implementation Note: Python to Go Migration

### Reference Implementation Approach

**This is the most important aspect of implementing new social media workflows:** We are migrating from Python to Go, and the Python codebase serves as the authoritative reference for business logic and API interactions. Every new platform implementation must carefully study and understand the existing Python code before writing the Go version.

### Migration Principles

1. **Preserve All API Functionality**

   - Every API call made in the Python code must have an equivalent in Go
   - Request parameters, headers, and authentication methods must match exactly
   - Response parsing logic should capture all the same fields
   - Rate limiting and retry logic must be maintained

2. **Improve Code Quality**

   - Use more descriptive and Go-idiomatic naming conventions
   - Structure code with better separation of concerns
   - Add comprehensive error handling that Python might have lacked
   - Implement proper context cancellation and timeouts

3. **Study Before Implementation**

   ```go
   // Example: If Python code has:
   // def fetch_posts(self, account_id, since_date):
   //     url = f"{self.base_url}/v1/accounts/{account_id}/posts"
   //     params = {"since": since_date, "limit": 100}
   //     headers = {"Authorization": f"Bearer {self.token}"}

   // Go implementation should preserve the logic:
   func (c *PlatformClient) FetchPosts(ctx context.Context, accountID string, sinceDate time.Time) ([]Post, error) {
       url := fmt.Sprintf("%s/v1/accounts/%s/posts", c.baseURL, accountID)
       params := url.Values{
           "since": {sinceDate.Format(time.RFC3339)},
           "limit": {"100"},
       }
       // Maintain exact same API contract
   }
   ```

4. **Ask Questions When Uncertain**
   - If Python code logic is unclear, ask for clarification before proceeding
   - Document any assumptions made during migration
   - Flag any Python patterns that don't translate well to Go

### Common Python to Go Patterns

| Python Pattern              | Go Equivalent                         | Notes                               |
| --------------------------- | ------------------------------------- | ----------------------------------- |
| `requests.get()` with retry | HTTP client with exponential backoff  | Implement proper retry logic        |
| Dictionary comprehensions   | Map initialization or loops           | Use explicit loops for clarity      |
| Dynamic typing              | Strong interfaces and type assertions | Define clear structs for all data   |
| `try/except` blocks         | Error returns and checking            | Always check and wrap errors        |
| Class inheritance           | Composition and interfaces            | Favor composition over inheritance  |
| Decorators                  | Middleware functions                  | Use function wrappers or middleware |

### Example Migration Process

When migrating a Python fetcher to Go:

1. **Analyze the Python code structure:**

   - Identify all API endpoints used
   - Document request/response formats
   - Note error handling patterns
   - List all data transformations

2. **Map Python concepts to Go:**

   - Convert classes to structs with methods
   - Replace dynamic configs with typed structs
   - Transform generators to channels or slices
   - Convert async/await to goroutines if needed

3. **Preserve business logic exactly:**

   - Keep the same pagination strategies
   - Maintain identical data filtering rules
   - Preserve all calculated fields
   - Keep the same error recovery behavior

4. **Enhance where appropriate:**
   - Add context support for cancellation
   - Implement proper connection pooling
   - Add structured logging throughout
   - Include comprehensive metrics

### Questions to Ask Before Implementation

Before starting any new platform migration, ensure you understand:

- What are all the API endpoints this platform uses?
- Are there any special authentication flows or token refresh mechanisms?
- What rate limits exist and how does the Python code handle them?
- Are there any platform-specific data transformations or calculations?
- What error scenarios does the Python code handle?
- Are there any undocumented behaviors or workarounds in the Python code?

Remember: The goal is not just to translate Python to Go, but to create a more robust, maintainable, and performant implementation while preserving 100% of the functionality.

## Architecture Overview

### Core Pipeline Pattern

Every social media workflow follows this exact five-stage pipeline:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Account        │    │  Platform        │    │  Platform       │    │  Platform       │    │  Platform       │
│  Fetcher        │───▶│  Fetcher         │───▶│  Parser         │───▶│  Immediate      │───▶│  ClickHouse     │
│                 │    │                  │    │                 │    │  Processor      │    │  Sink           │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │                       │                       │
         ▼                       ▼                       ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ work-order-     │    │ raw-platform-*   │    │ parsed-platform │    │ processed-       │    │ ClickHouse      │
│ platform        │    │ topics           │    │ topics          │    │ platform-topics  │    │ Database        │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Key Principles

1. **Unidirectional Data Flow**: Data flows in one direction only, from MongoDB → Kafka → ClickHouse
2. **Service Independence**: Services communicate only through Kafka, never directly
3. **Consistent Patterns**: Each platform follows identical structural patterns
4. **Type Safety**: Strong typing throughout with Go structs
5. **Production Ready**: Built-in observability, error handling, and scalability

## Directory Structure

For each new platform, create the following structure:

```
src/
├── cmd/
│   └── services/
│       ├── {platform}-fetcher/          # API data extraction
│       ├── {platform}-parser/           # Data parsing and transformation
│       ├── {platform}-immediate-processor/  # Real-time processing
│       └── {platform}-clickhouse-sink/  # Batch storage
├── internal/
│   ├── clients/
│   │   └── social/
│   │       └── {platform}.go           # Platform API client
│   └── parsing/
│       └── {platform}_parser.go        # Parsing logic
├── pkg/
│   ├── models/
│   │   ├── kafka/                    # Add platform Kafka message types
│   │   │   └── {platform}.go
│   │   ├── clickhouse/               # Add platform ClickHouse models
│   │   │   └── {platform}.go
│   │   └── mongo/                    # Add platform MongoDB models
│   │       └── {platform}.go
│   └── sinks/
│       └── {platform}_clickhouse.go    # ClickHouse sink conversion logic
└── sql/
    └── clickhouse/
        └── {platform}_tables.sql       # ClickHouse table schemas
```

## Step-by-Step Implementation Guide

### Step 1: Define Data Models

#### 1.1 MongoDB Models (`src/pkg/models/mongo/`)

Add platform-specific account model following the FacebookAccount pattern:

```go
// {Platform}Account represents a document in the '{platform}_accounts' collection
type {Platform}Account struct {
    ID                               primitive.ObjectID `bson:"_id,omitempty"`
    {Platform}ID                     string             `bson:"{platform}_id,omitempty"`
    Type                             string             `bson:"type,omitempty"`
    Validity                         string             `bson:"validity,omitempty"`
    State                            string             `bson:"state,omitempty"`
    AccessToken                      string             `bson:"access_token,omitempty"`
    RefreshToken                     string             `bson:"refresh_token,omitempty"`
    TokenExpiry                      *MongoTime         `bson:"token_expiry,omitempty"`

    // Analytics update timestamps
    LastAnalyticsUpdatedAt           *MongoTime         `bson:"last_analytics_updated_at,omitempty"`
    LastInsightsAnalyticsUpdatedAt   *MongoTime         `bson:"last_insights_analytics_updated_at,omitempty"`
    LastVideoAnalyticsUpdatedAt      *MongoTime         `bson:"last_video_analytics_updated_at,omitempty"`
    CreatedAt                        *MongoTime         `bson:"created_at,omitempty"`

    // Dynamic fields for platform-specific data
    ExtraData map[string]interface{} `bson:",inline"`
}
```

#### 1.2 Kafka Message Models (`src/pkg/models/kafka/`)

Define all message types for the platform:

```go
// Raw{Platform}Post represents raw post data from the platform API
type Raw{Platform}Post struct {
    ID          string    `json:"id"`
    Content     string    `json:"content"`
    AuthorID    string    `json:"author_id"`
    AuthorName  string    `json:"author_name"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    // Platform-specific fields
    Metrics     {Platform}Metrics `json:"metrics"`
    MediaItems  []MediaItem       `json:"media_items"`
    // ... additional fields
}

// Parsed{Platform}Post represents processed post data ready for analytics
type Parsed{Platform}Post struct {
    // Standard fields across all platforms
    PageID          string    `json:"page_id"`
    PageName        string    `json:"page_name"`
    PostID          string    `json:"post_id"`
    Content         string    `json:"content"`
    MediaType       string    `json:"media_type"`
    CreatedTime     time.Time `json:"created_time"`
    SavingTime      time.Time `json:"saving_time"`

    // Engagement metrics
    Likes           int64     `json:"likes"`
    Comments        int64     `json:"comments"`
    Shares          int64     `json:"shares"`
    Views           int64     `json:"views"`
    TotalEngagement int64     `json:"total_engagement"`

    // Platform-specific metrics
    // ... additional fields
}
```

#### 1.3 ClickHouse Models (`src/pkg/models/clickhouse/`)

Define ClickHouse storage models:

```go
// {Platform}Posts represents the ClickHouse table structure
type {Platform}Posts struct {
    PageName        string    `ch:"page_name"`
    PageID          string    `ch:"page_id"`
    PostID          string    `ch:"post_id"`
    MediaType       string    `ch:"media_type"`
    Content         string    `ch:"content"`
    CreatedTime     time.Time `ch:"created_time"`
    UpdatedTime     time.Time `ch:"updated_time"`
    SavingTime      time.Time `ch:"saving_time"`

    // Engagement metrics
    Likes           int64     `ch:"likes"`
    Comments        int64     `ch:"comments"`
    Shares          int64     `ch:"shares"`
    Views           int64     `ch:"views"`
    TotalEngagement int64     `ch:"total_engagement"`

    // Analytics dimensions
    DayOfWeek       string    `ch:"day_of_week"`
    HourOfDay       int32     `ch:"hour_of_day"`

    // Platform-specific fields
    // ... additional fields
}
```

### Step 2: Create Platform API Client

Create `src/internal/clients/social/{platform}.go`:

```go
package social

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/logger"
    kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/pkg/models/kafka"
)

const (
    {platform}APIVersion = "v1"
    {platform}BaseURL    = "https://api-server.{platform}.com/"
    maxPagesToFetch     = 100
)

// {Platform}Client handles all interactions with the {Platform} API
type {Platform}Client struct {
    httpClient  *http.Client
    baseURL     string
    apiKey      string
    apiSecret   string
    log         *logger.Logger
}

// New{Platform}Client creates a new {Platform} API client
func New{Platform}Client(apiKey, apiSecret string) *{Platform}Client {
    return &{Platform}Client{
        httpClient: &http.Client{Timeout: 45 * time.Second},
        baseURL:    {platform}BaseURL,
        apiKey:     apiKey,
        apiSecret:  apiSecret,
        log:        logger.New("info"),
    }
}

// FetchPosts retrieves posts for a given account
func (c *{Platform}Client) FetchPosts(ctx context.Context, accountID, accessToken string) ([]kafkamodels.Raw{Platform}Post, error) {
    // Implementation following Facebook pattern:
    // 1. Build API URL with proper endpoint
    // 2. Add authentication headers
    // 3. Handle pagination
    // 4. Parse responses
    // 5. Handle errors gracefully
    // 6. Return accumulated results
}

// Additional methods for videos, insights, etc.
```

### Step 3: Implement Fetcher Service

Create `src/cmd/services/{platform}-fetcher/main.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "os"
    "os/signal"
    "sync"
    "syscall"

    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/config"
    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/kafka"
    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/logger"
    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/clients/social"
)

const (
    maxWorkers        = 5
    workOrderChanSize = 100
)

// {Platform}AccountWorkOrder represents work order structure
type {Platform}AccountWorkOrder struct {
    ID           string `json:"id"`
    {Platform}ID string `json:"{platform}_id"`
    Type         string `json:"type"`
    AccessToken  string `json:"access_token"`
    WorkspaceID  string `json:"workspace_id"`
    SyncType     string `json:"sync_type"`
}

func main() {
    // 1. Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        panic("Failed to load configuration: " + err.Error())
    }

    // 2. Initialize logger
    log := logger.New(cfg.LogLevel)
    log.Info().Msg("Starting {Platform} Fetcher service")

    // 3. Create platform client
    client := social.New{Platform}Client(cfg.{Platform}.APIKey, cfg.{Platform}.APISecret)

    // 4. Create Kafka consumer and producer
    consumer, err := kafka.NewConsumer(cfg.Kafka, "{platform}-fetcher-group", log.Logger)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
    }
    defer consumer.Close()

    producer, err := kafka.NewProducer(cfg.Kafka, log.Logger)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to create Kafka producer")
    }
    defer producer.Close()

    // 5. Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 6. Create worker pool
    workOrderChan := make(chan WorkOrderMessage, workOrderChanSize)
    var wg sync.WaitGroup

    for i := 0; i < maxWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            workOrderProcessor(ctx, workerID, workOrderChan, client, producer, cfg)
        }(i)
    }

    // 7. Start consuming work orders
    go func() {
        topics := []string{"work-order-{platform}"}
        consumer.Consume(ctx, topics, messageHandler(workOrderChan))
    }()

    // 8. Wait for shutdown
    <-sigChan
    log.Info().Msg("Shutdown signal received")
    cancel()
    close(workOrderChan)
    wg.Wait()

    log.Info().Msg("{Platform} Fetcher service stopped")
}

// Process work orders and fetch data from platform API
func workOrderProcessor(ctx context.Context, workerID int, workOrderChan <-chan WorkOrderMessage,
    client *social.{Platform}Client, producer kafka.Producer, cfg *config.Config) {
    // Implementation following Facebook pattern
}
```

### Step 4: Create Parser Service

Create `src/internal/parsing/{platform}_parser.go`:

```go
package parsing

import (
    "crypto/md5"
    "fmt"
    "time"

    kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/pkg/models/kafka"
)

// {Platform}Parser handles parsing of raw {Platform} data
type {Platform}Parser struct {
    MediaTypeMapping map[string]string
    MetricFields     []string
}

// New{Platform}Parser creates a new parser instance
func New{Platform}Parser() *{Platform}Parser {
    return &{Platform}Parser{
        MediaTypeMapping: map[string]string{
            // Platform-specific mappings
        },
        MetricFields: []string{
            // Platform-specific metrics
        },
    }
}

// ParsePost converts raw post to parsed format
func (p *{Platform}Parser) ParsePost(rawPost kafkamodels.Raw{Platform}Post, accountID, accountName string) (*kafkamodels.Parsed{Platform}Post, []kafkamodels.Parsed{Platform}MediaAsset, error) {
    // Implementation following Facebook pattern:
    // 1. Initialize parsed post structure
    // 2. Extract standard fields
    // 3. Parse engagement metrics
    // 4. Extract media assets
    // 5. Calculate derived metrics
    // 6. Return parsed data
}
```

### Step 5: Implement Parser Service

Create `src/cmd/services/{platform}-parser/main.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "strings"
    "sync"

    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/config"
    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/kafka"
    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/logger"
    "github.com/d4interactive/contentstudio-social-analytics-go/src/internal/parsing"
    kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/pkg/models/kafka"
)

const (
    maxWorkers      = 5
    messageChanSize = 100
)

func main() {
    // Setup similar to fetcher
    // Main difference: consume from raw_{platform}_* topics
    // and produce to parsed_{platform}_* topics
}
```

### Step 6: Create ClickHouse Sink Converter

Create `src/pkg/sinks/{platform}_clickhouse.go`:

```go
package sinks

import (
    "context"
    clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/pkg/models/schema"
)

// Convert{Platform}Post converts parsed post to ClickHouse model
func (s *ClickHouseSink) Convert{Platform}Post(parsed *clickhousemodels.{Platform}Posts) *clickhousemodels.{Platform}Posts {
    if parsed == nil {
        return nil
    }

    return &clickhousemodels.{Platform}Posts{
        // Map all fields from parsed to ClickHouse model
        PageName:        parsed.PageName,
        PageID:          parsed.PageID,
        PostID:          parsed.PostID,
        MediaType:       parsed.MediaType,
        Content:         parsed.Content,
        CreatedTime:     parsed.CreatedTime,
        UpdatedTime:     parsed.UpdatedTime,
        SavingTime:      parsed.SavingTime,
        // ... map all fields
    }
}

// BulkInsert{Platform}Posts performs batch insert to ClickHouse
func (s *ClickHouseSink) BulkInsert{Platform}Posts(ctx context.Context, posts []*clickhousemodels.{Platform}Posts) error {
    // Implementation following Facebook pattern
}
```

### Step 7: Update ClickHouse Client

Add platform-specific methods to `src/internal/clients/clickhouse/{platform}.go`:

```go
package clickhouse

import (
    "context"
    "fmt"

    "github.com/d4interactive/contentstudio-social-analytics-go/src/pkg/models/clickhouse"
)

// BulkInsert{Platform}Posts inserts {Platform} posts into ClickHouse
func (c *Client) BulkInsert{Platform}Posts(ctx context.Context, posts []*clickhousemodels.{Platform}Posts) error {
    if len(posts) == 0 {
        return nil
    }

    c.logger.Info().
        Int("count", len(posts)).
        Msg("Bulk inserting {Platform} posts to ClickHouse")

    batch, err := c.conn.PrepareBatch(ctx, `
        INSERT INTO {platform}_posts (
            page_name, page_id, post_id, media_type, content,
            created_time, updated_time, saving_time,
            likes, comments, shares, views, total_engagement,
            day_of_week, hour_of_day
            -- Add all fields
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to prepare batch: %w", err)
    }

    // Append all posts to batch
    for _, post := range posts {
        err = batch.Append(
            post.PageName, post.PageID, post.PostID, post.MediaType, post.Content,
            post.CreatedTime, post.UpdatedTime, post.SavingTime,
            post.Likes, post.Comments, post.Shares, post.Views, post.TotalEngagement,
            post.DayOfWeek, post.HourOfDay,
            // ... all fields
        )
        if err != nil {
            return fmt.Errorf("failed to append post to batch: %w", err)
        }
    }

    return batch.Send()
}
```

### Step 8: Create MongoDB Repository

Create `src/internal/repository/mongodb/{platform}_account_repository.go`:

```go
package mongodb

import (
    "context"
    mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/pkg/models/mongo"
    "github.com/rs/zerolog"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

// {Platform}AccountRepository interface
type {Platform}AccountRepository interface {
    FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.{Platform}Account, error)
    GetBy{Platform}ID(ctx context.Context, {platform}ID string) (*mongomodels.{Platform}Account, error)
    GetValidAccounts(ctx context.Context, accountType string) ([]mongomodels.{Platform}Account, error)
    UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error
    UpdateLastAnalyticsUpdatedAt(ctx context.Context, id primitive.ObjectID, timestamp time.Time) error
    // ... other methods
}

// Implementation following FacebookAccountRepository pattern
```

### Step 9: Update Configuration

#### 10.1 Add to `src/internal/config/config.go`:

```go
// {Platform}Config holds {Platform} specific configuration
type {Platform}Config struct {
    APIKey       string `mapstructure:"api_key"`
    APISecret    string `mapstructure:"api_secret"`
    APIEndpoint  string `mapstructure:"api_endpoint"`
    RateLimit    int    `mapstructure:"rate_limit"`
}

// Add to main Config struct:
{Platform} {Platform}Config `mapstructure:"{platform}"`

// Add defaults in LoadConfig():
viper.SetDefault("{PLATFORM}.API_KEY", "")
viper.SetDefault("{PLATFORM}.API_SECRET", "")
viper.SetDefault("{PLATFORM}.API_ENDPOINT", "https://api-server.{platform}.com")
viper.SetDefault("{PLATFORM}.RATE_LIMIT", 100)
```

#### 10.2 Update `.env.example`:

```bash
# {Platform} API Configuration
APP_{PLATFORM}_API_KEY=""
APP_{PLATFORM}_API_SECRET=""
APP_{PLATFORM}_API_ENDPOINT="https://api.{platform}.com"
APP_{PLATFORM}_RATE_LIMIT=100
```

### Step 10: Update Build System

Add to `src/Makefile`:

```makefile
# Add new service variables
APP_SERVICE_{PLATFORM}_FETCHER := {platform}_fetcher
APP_SERVICE_{PLATFORM}_PARSER := {platform}_parser
APP_SERVICE_{PLATFORM}_IMMEDIATE_PROCESSOR := {platform}_immediate_processor
APP_SERVICE_{PLATFORM}_CLICKHOUSE_SINK := {platform}_clickhouse_sink

# Add build targets
SERVICE_{PLATFORM}_FETCHER_BIN := $(BIN_DIR)/$(APP_SERVICE_{PLATFORM}_FETCHER)
SERVICE_{PLATFORM}_PARSER_BIN := $(BIN_DIR)/$(APP_SERVICE_{PLATFORM}_PARSER)
# ... etc

# Add to build dependencies
build: ... $(SERVICE_{PLATFORM}_FETCHER_BIN) $(SERVICE_{PLATFORM}_PARSER_BIN) ...

# Add individual build rules
$(SERVICE_{PLATFORM}_FETCHER_BIN): $(GO_SOURCE_FILES)
	@echo "==> Building $(APP_SERVICE_{PLATFORM}_FETCHER)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(SERVICE_{PLATFORM}_FETCHER_BIN) ./cmd/services/{platform}-fetcher
	@echo "==> $(APP_SERVICE_{PLATFORM}_FETCHER) built: $(SERVICE_{PLATFORM}_FETCHER_BIN)"
```

## Testing and Validation

### Unit Tests

For each component, create corresponding test files:

```go
// {platform}_parser_test.go
func TestParse{Platform}Post(t *testing.T) {
    parser := New{Platform}Parser()

    rawPost := kafkamodels.Raw{Platform}Post{
        // Test data
    }

    parsed, assets, err := parser.ParsePost(rawPost, "test_id", "test_name")

    assert.NoError(t, err)
    assert.NotNil(t, parsed)
    assert.Equal(t, "expected_value", parsed.SomeField)
}
```

### Integration Tests

Test the complete pipeline:

1. Mock API responses
2. Test data flow through Kafka
3. Verify ClickHouse storage
4. Check error handling

### Performance Testing

1. Load test API client with rate limiting
2. Benchmark parsing performance
3. Test batch insert efficiency
4. Monitor memory usage

## Common Patterns and Best Practices

### Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to fetch {platform} posts for account %s: %w", accountID, err)
}

// Log errors with structured context
log.Error().
    Err(err).
    Str("account_id", accountID).
    Str("platform", "{platform}").
    Msg("Failed to process work order")
```

### Logging

```go
// Use structured logging consistently
log.Info().
    Str("platform", "{platform}").
    Str("account_id", accountID).
    Int("posts_fetched", len(posts)).
    Dur("duration", time.Since(startTime)).
    Msg("Successfully fetched posts")
```

### Metrics

```go
// Define Prometheus metrics
var (
    {platform}PostsFetched = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "{platform}_posts_fetched_total",
            Help: "Total number of {platform} posts fetched",
        },
        []string{"account_type", "status"},
    )
)
```

### Graceful Shutdown

```go
// Always implement graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Info().Msg("Shutdown signal received")

    // 1. Stop accepting new work
    cancel()

    // 2. Wait for in-flight work to complete
    wg.Wait()

    // 3. Close connections
    consumer.Close()
    producer.Close()

    // 4. Exit
    log.Info().Msg("Service stopped gracefully")
}()
```

## Platform-Specific Considerations

### Instagram

- Handle both personal and business accounts
- Support Stories, Reels, and IGTV
- Implement Instagram Basic Display API and Graph API
- Handle media carousel posts

### Twitter/X

- Support v2 API endpoints
- Handle tweet threads
- Implement OAuth 2.0 authentication
- Track retweets vs quotes

### LinkedIn

- Support both personal profiles and company pages
- Handle article posts differently
- Implement OAuth 2.0 with refresh tokens
- Track professional engagement metrics

### TikTok

- Handle short-form video content
- Track trending metrics
- Implement TikTok Marketing API
- Support effects and sounds analytics

### YouTube

- Support channels and videos
- Track detailed video analytics
- Handle YouTube Data API v3
- Support live stream metrics

### Pinterest

- Handle boards and pins
- Track save and click metrics
- Implement Pinterest API v5
- Support shopping features

## Conclusion

This guide provides a complete blueprint for implementing new social media platform integrations. By following these patterns and conventions, each new platform will maintain consistency with the existing codebase while leveraging proven architectural decisions and best practices.

Remember: consistency is key. When in doubt, refer to the Facebook implementation as the reference architecture.
