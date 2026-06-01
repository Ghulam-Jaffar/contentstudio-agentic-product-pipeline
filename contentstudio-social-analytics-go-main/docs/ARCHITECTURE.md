# ContentStudio Social Analytics - System Architecture

## Overview

ContentStudio Social Analytics is a high-performance microservices-based data pipeline for extracting, processing, and storing social media analytics data from multiple platforms into ClickHouse for real-time analytics.

## Supported Platforms

| Platform | Status | Services | Documentation |
|----------|--------|----------|---------------|
| Facebook | Production | 6 microservices | [facebook.md](platforms/facebook.md) |
| Instagram | Production | 6 microservices | [instagram.md](platforms/instagram.md) |
| LinkedIn | Production | 5 microservices | [linkedin.md](platforms/linkedin.md) |
| TikTok | Implemented | 2 microservices | [tiktok.md](platforms/tiktok.md) |
| YouTube | Implemented | 3 microservices | [youtube.md](platforms/youtube.md) |
| GMB | Implemented | 2 microservices | [gmb.md](platforms/gmb.md) |

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              MONGODB                                             │
│                    (social_integrations collection)                              │
│         Facebook | Instagram | LinkedIn | TikTok | YouTube accounts              │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           UNIFIED ACCOUNT FETCHER                                │
│                    (Reads accounts, creates work orders)                         │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
            ┌───────────┐     ┌───────────┐     ┌───────────┐
            │work-order-│     │work-order-│     │work-order-│
            │ facebook  │     │ linkedin  │     │  youtube  │
            └─────┬─────┘     └─────┬─────┘     └─────┬─────┘
                  │                 │                 │
                  ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              KAFKA TOPICS                                        │
│                                                                                  │
│  work-order-* ──▶ raw-* ──▶ parsed-* ──▶ *-db (sink topics)                    │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
            ┌───────────┐     ┌───────────┐     ┌───────────┐
            │  Fetcher  │     │  Parser   │     │   Sink    │
            │ (API Call)│     │(Transform)│     │(ClickHouse│
            └───────────┘     └───────────┘     └───────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLICKHOUSE                                          │
│                                                                                  │
│  facebook_posts | instagram_posts | linkedin_posts | tiktok_posts | youtube_*   │
│  facebook_insights | instagram_insights | linkedin_insights | tiktok_insights   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Five-Stage Pipeline

### Stage 1: Scheduler (Account Fetcher)
- Reads accounts from MongoDB `social_integrations` collection
- Groups accounts by platform and type
- Creates batch work orders
- Publishes to `work-order-{platform}` topics

### Stage 2: Fetcher
- Consumes work orders from Kafka
- Calls platform-specific APIs
- Handles authentication and rate limiting
- Produces raw data to `raw-{platform}-*` topics

### Stage 3: Parser
- Consumes raw data from Kafka
- Transforms API responses to analytics-ready format
- Normalizes data across platforms
- Produces parsed data to `parsed-{platform}-*` topics

### Stage 4: Processor (Optional)
- Enriches parsed data
- Calculates derived metrics
- Handles immediate/real-time processing

### Stage 5: Sink
- Consumes parsed/processed data
- Batch inserts to ClickHouse
- Handles deduplication
- Manages backpressure

## Graceful Shutdown Architecture

All services implement a consistent graceful shutdown pattern to ensure zero data loss during shutdowns or deployments.

### Shutdown Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `idleTimeout` | 5 minutes | Time to wait with no activity before shutdown |
| `idleCheckInterval` | 30 seconds | How often to check for idle state |

### Fetcher Service Shutdown Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         FETCHER GRACEFUL SHUTDOWN                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. Kafka Consumer receives message                                          │
│     └── atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())          │
│                                                                              │
│  2. Work order dispatched to worker                                          │
│     └── atomic.AddInt64(&activeJobs, 1)  // Increment                       │
│                                                                              │
│  3. Worker processes work order                                              │
│     ├── Calls platform API                                                   │
│     ├── Publishes raw data to Kafka                                          │
│     └── Updates MongoDB                                                      │
│                                                                              │
│  4. Work order completed                                                     │
│     ├── atomic.AddInt64(&activeJobs, -1)  // Decrement                      │
│     └── atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())          │
│                                                                              │
│  5. Idle checker runs every 30 seconds                                       │
│     └── IF len(workChan) == 0 AND activeJobs == 0                           │
│         AND time.Since(lastMessageTime) >= 5 minutes                         │
│         THEN trigger graceful shutdown                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### ClickHouse Sink Shutdown Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      CLICKHOUSE SINK GRACEFUL SHUTDOWN                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. Kafka Consumer receives message                                          │
│     └── atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())          │
│                                                                              │
│  2. Message routed to parser workers                                         │
│     └── Parsed data sent to batch collector channels                         │
│                                                                              │
│  3. Batch processors collect data                                            │
│     ├── Collect until maxBatchSize (10,000) OR batchTimeout (10s)           │
│     └── Execute batch INSERT to ClickHouse                                   │
│                                                                              │
│  4. Idle checker runs every 30 seconds                                       │
│     └── IF all batch channels empty                                          │
│         AND no pending batches                                               │
│         AND time.Since(lastMessageTime) >= 5 minutes                         │
│         THEN trigger graceful shutdown                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Implementation Details

#### Active Job Tracking (Fetchers)
```go
// processorWithTracking wraps worker with job tracking
func processorWithTracking(ctx context.Context, workerID int, workChan <-chan WorkOrderMessage,
    ..., activeJobs *int64, lastMessageTime *int64) {
    for {
        select {
        case msg, ok := <-workChan:
            if !ok {
                return
            }
            atomic.AddInt64(activeJobs, 1)           // Start tracking
            processWorkOrder(msg)
            atomic.AddInt64(activeJobs, -1)          // Stop tracking
            atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())
        case <-ctx.Done():
            return
        }
    }
}
```

#### Idle State Check
```go
// Idle checker goroutine
ticker := time.NewTicker(idleCheckInterval)
for {
    select {
    case <-ticker.C:
        lastTime := time.Unix(0, atomic.LoadInt64(&lastMessageTime))
        idleDuration := time.Since(lastTime)
        queueEmpty := len(workChan) == 0
        noActiveJobs := atomic.LoadInt64(&activeJobs) == 0

        if queueEmpty && noActiveJobs && idleDuration >= idleTimeout {
            log.Info().Msg("Idle timeout reached, initiating graceful shutdown")
            cancel()
        }
    case <-ctx.Done():
        return
    }
}
```

### Shutdown Guarantees

1. **Zero Data Loss**: All in-flight work orders complete before shutdown
2. **Queue Drainage**: All queued messages are processed before exit
3. **Batch Completion**: All pending ClickHouse batches are flushed
4. **Clean Exit**: Kafka consumers commit offsets, connections closed properly
5. **Timeout Safety**: 5-minute idle window prevents premature shutdown

### Platform-Specific Implementations

| Platform | Fetcher | Analytics Sink |
|----------|---------|----------------|
| Facebook | `processorWithTracking` | Batch channel drain |
| Instagram | `processorWithTracking` | Batch channel drain |
| LinkedIn | `processorWithTracking` | Batch channel drain |
| TikTok | `processorWithTracking` | Batch channel drain |
| YouTube | `processorWithTracking` | Batch channel drain |

## Data Flow

### Scheduled (Daily) Processing
```
MongoDB ──▶ Account Fetcher ──▶ work-order-* ──▶ Fetcher ──▶ raw-*
       ──▶ Parser ──▶ parsed-* ──▶ Sink ──▶ ClickHouse
```

### Immediate (Real-time) Processing
```
API Request ──▶ immediate-work-order-* ──▶ Immediate Processor ──▶ ClickHouse
```

## MongoDB Schema

### social_integrations Collection
```javascript
{
  "_id": ObjectId,
  "platform_type": "facebook|instagram|linkedin|tiktok|youtube",
  "platform_identifier": "platform_specific_id",
  "platform_name": "Display Name",
  "type": "Page|Profile|Business|Creator",
  "state": "Added|Syncing|Processed|Deleted",
  "validity": "valid|invalid|expired",

  // Authentication
  "access_token": "encrypted_token",
  "refresh_token": "encrypted_refresh_token",
  "long_access_token": "facebook_long_lived_token",

  // Tracking
  "workspace_id": ObjectId,
  "user_id": ObjectId,
  "last_analytics_updated_at": "timestamp_string",

  // Platform-specific fields
  "facebook_id": "for_facebook",
  "instagram_id": "for_instagram",
  "linkedin_id": "for_linkedin",
  "tiktok_id": "for_tiktok",
  "channel_id": "for_youtube"
}
```

## Kafka Topic Naming Convention

| Pattern | Example | Purpose |
|---------|---------|---------|
| `work-order-{platform}` | `work-order-facebook` | Batch work orders |
| `immediate-work-order-{platform}` | `immediate-work-order-facebook` | Real-time requests |
| `raw-{platform}-{type}` | `raw-facebook-posts` | Raw API responses |
| `parsed-{platform}-{type}` | `parsed-facebook-posts` | Normalized data |
| `{platform}-{type}-db` | `facebook-posts-db` | Sink consumption |

## ClickHouse Table Naming Convention

| Pattern | Example | Purpose |
|---------|---------|---------|
| `{platform}_posts` | `facebook_posts` | Post/media analytics |
| `{platform}_insights` | `facebook_insights` | Account-level metrics |
| `{platform}_competitors_*` | `facebook_competitors_posts` | Competitor data |
| `{platform}_{special}` | `youtube_traffic_insights` | Platform-specific |

## Consumer Group Naming Convention

| Pattern | Example |
|---------|---------|
| `{platform}-fetcher-group` | `facebook-fetcher-group` |
| `{platform}-parser-group` | `facebook-posts-parser-group` |
| `{platform}-immediate-processor-group` | `facebook-immediate-processor-group` |
| `{platform}-clickhouse-sink-group` | `facebook-clickhouse-sink-group` |

## Configuration

### Environment Variables
```bash
# MongoDB
APP_MONGO_HOST=localhost
APP_MONGO_PORT=27017
APP_MONGO_DATABASE=contentstudio
APP_MONGO_USERNAME=user
APP_MONGO_PASSWORD=pass

# Kafka
APP_KAFKA_BROKERS=localhost:9092
APP_KAFKA_SASL_AUTH=true
APP_KAFKA_SASL_USERNAME=user
APP_KAFKA_SASL_PASSWORD=pass

# ClickHouse
APP_CLICKHOUSE_HOST=localhost
APP_CLICKHOUSE_PORT=9000
APP_CLICKHOUSE_DATABASE=contentstudiobackend
APP_CLICKHOUSE_USERNAME=default
APP_CLICKHOUSE_PASSWORD=pass

# Platform Credentials
APP_FACEBOOK_APP_ID=xxx
APP_FACEBOOK_APP_SECRET=xxx
APP_YOUTUBE_CLIENT_ID=xxx
APP_YOUTUBE_CLIENT_SECRET=xxx
APP_TIKTOK_CLIENT_KEY=xxx
APP_TIKTOK_CLIENT_SECRET=xxx

# Security
APP_DECRYPTION_KEY=xxx

# Monitoring
APP_LOG_LEVEL=info
APP_SENTRY_DSN=https://xxx@sentry.io/xxx
```

## Rate Limiting by Platform

| Platform | Per-Token | Global | Strategy |
|----------|-----------|--------|----------|
| Facebook | 4 RPS | 12 RPS | Token bucket |
| Instagram | 4 RPS | 12 RPS | Token bucket |
| LinkedIn | - | HTTP timeout | Pagination delay |
| TikTok | - | 100ms delay | Inter-page delay |
| YouTube | 5 RPS | 10 burst | Exponential backoff |

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Max Workers per Service | 10-15 |
| Channel Buffer Size | 50,000 messages |
| ClickHouse Batch Size | 1,000-5,000 items |
| Idle Timeout | 5 minutes |
| Idle Check Interval | 30 seconds |
| HTTP Timeout | 30 seconds |

## Error Handling

### Token Expiration
- Facebook: Long-lived token refresh
- Instagram: Re-authentication required
- LinkedIn: OAuth refresh flow
- TikTok: Refresh token grant
- YouTube: OAuth refresh flow

### Rate Limiting
- Exponential backoff with jitter
- Per-token and global limits
- Automatic retry with delay

### API Errors
- Logged to Sentry with context
- Non-fatal errors: skip and continue
- Fatal errors: stop processing account

## Monitoring

### Metrics Logged
- Messages picked (from Kafka)
- Messages parsed (transformed)
- Messages inserted (to ClickHouse)
- Channel utilization
- Processing duration
- Error rates

### Health Checks
- Kafka consumer lag
- ClickHouse connection
- MongoDB connection
- API response times

## Directory Structure

```
src/
├── clients/
│   └── social/           # Platform API clients
├── services/
│   ├── facebook/         # Facebook microservices
│   ├── instagram/        # Instagram microservices
│   ├── linkedin/         # LinkedIn microservices
│   ├── tiktok/           # TikTok microservices
│   ├── youtube/          # YouTube microservices
│   └── unified/          # Shared services
├── models/
│   ├── kafka/            # Kafka message models
│   └── db/
│       ├── clickhouse/   # ClickHouse schemas
│       └── mongo/        # MongoDB models
├── utils/
│   └── parsing/          # Platform parsers
├── kafka/                # Kafka client wrapper
├── db/
│   ├── clickhouse/       # ClickHouse client
│   └── mongodb/          # MongoDB repository
├── config/               # Configuration management
├── logger/               # Structured logging
└── docs/                 # Documentation
    └── platforms/        # Platform-specific docs
```

## Deployment

### Prerequisites
- Go 1.21+
- Kafka cluster
- MongoDB cluster
- ClickHouse cluster
- Sentry (optional)

### Build
```bash
make build              # Build all services
make test               # Run tests
make clean              # Clean artifacts
```

### Run
```bash
# Start all Facebook services
./scripts/run_facebook_services.sh start

# Start specific service
./bin/facebook-fetcher
```

## Security Considerations

1. **Token Encryption**: All tokens encrypted at rest
2. **HMAC Verification**: Facebook uses appsecret_proof
3. **SASL Authentication**: Kafka with SASL
4. **TLS**: All external connections use TLS
5. **No PII in Logs**: Sensitive data excluded from logging
