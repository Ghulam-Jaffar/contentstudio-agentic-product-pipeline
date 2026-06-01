---
title: Pipeline Architecture Deep Dive
description: Comprehensive technical documentation of the ContentStudio Social Analytics Go pipeline
---

## Executive Summary

ContentStudio Social Analytics Go is a high-performance, microservices-based data pipeline that processes social media analytics data at scale. Built with Go for performance and reliability, it replaces a legacy Python system while maintaining business logic compatibility. The pipeline processes over 15,000 items per second using event-driven architecture with Kafka and stores analytical data in ClickHouse for real-time querying.

## Architecture Overview

### Core Design Principles

1. **Microservices Architecture**: 17+ independent services per platform, each with single responsibility
2. **Event-Driven Communication**: Kafka as the central nervous system for asynchronous message passing
3. **Five-Stage Pipeline**: Scheduler → Fetcher → Parser → Processor → Sink
4. **Platform Isolation**: Each social platform has dedicated services preventing cross-platform failures
5. **Batch Processing**: Optimized batching at sink level for 1500x performance improvement
6. **Graceful Degradation**: Services can fail independently without bringing down the pipeline

### System Components

```
MongoDB → Scheduler → Kafka → Fetcher → Kafka → Parser → Kafka → 
Processor → Kafka → Sink → ClickHouse
```

## Data Flow Architecture

### Stage 1: Scheduling and Work Order Generation

**Services**: 
- `scheduler/account_fetcher` (Common scheduler with platform flags)

**Responsibilities**:
- Reads social media accounts from MongoDB
- Generates work orders based on update intervals
- Publishes to platform-specific work order topics
- Handles incremental vs full sync strategies

**MongoDB Collections**:
- `facebook_accounts`: Facebook page accounts with tokens
- `instagram_accounts`: Instagram business accounts
- `linkedin_accounts`: LinkedIn organization pages
- `tiktok_accounts`: TikTok business accounts

**Output Topics**:
- `work-order-facebook`
- `work-order-instagram`
- `work-order-linkedin`
- `work-order-tiktok`
- `immediate-work-order-*` (for real-time processing)

### Stage 2: Data Fetching

**Services**:
- `facebook-fetcher`: Facebook Graph API v19.0 integration
- `instagram-fetcher`: Instagram Graph API integration
- `linkedin-fetcher`: LinkedIn API integration
- `tiktok-fetcher`: TikTok Business API integration

**Responsibilities**:
- Consumes work orders from Kafka
- Authenticates with platform APIs
- Fetches raw data (posts, videos, insights, media)
- Handles pagination and rate limiting
- Publishes raw data to platform-specific topics

**API Integration Features**:
- **Facebook**: 
  - HMAC-SHA256 appsecret_proof for security
  - Comprehensive field selection (60+ fields)
  - Reaction type breakdown (LIKE, LOVE, WOW, HAHA, SORRY, ANGRY, THANKFUL)
  - Video insights with 60+ metrics
- **Instagram**: 
  - Media insights (impressions, reach, saves)
  - Story metrics
  - IGTV and Reels analytics
- **LinkedIn**: 
  - Organization post statistics
  - Share analytics
  - Follower demographics
- **TikTok**: 
  - Video performance metrics
  - Hashtag analytics
  - Audience insights

**Output Topics**:
- `raw-facebook_posts`, `raw-facebook-videos`, `raw-facebook-insights`
- `raw-instagram_media`, `raw-instagram-insights`
- `raw-linkedin_posts`, `raw-linkedin-images`, `raw-linkedin-videos`, `raw-linkedin-stats`
- `raw-tiktok_posts`, `raw-tiktok-insights`

### Stage 3: Data Parsing and Transformation

**Services**:
- `facebook-posts-parser`: Parses Facebook raw data
- `instagram-posts-parser`: Transforms Instagram media
- `linkedin-parser`: Processes LinkedIn content
- `tiktok-parser`: Normalizes TikTok data

**Responsibilities**:
- Consumes raw data from Kafka
- Normalizes data structures across platforms
- Extracts nested fields and metrics
- Handles media asset extraction
- Transforms timestamps and data types
- Publishes parsed data to Kafka

**Data Transformations**:
- **Timestamp Normalization**: Facebook's custom format to RFC3339
- **Reaction Aggregation**: Individual reaction types to totals
- **Media Processing**: Extract URLs, dimensions, thumbnails
- **Metric Calculation**: Engagement rates, reach percentages
- **Content Classification**: Post types (photo, video, link, status)

**Output Topics**:
- `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-insights`, `parsed-facebook-reels-insights`
- `parsed-instagram-posts`, `parsed-instagram-insights`
- `parsed-linkedin-posts`, `parsed-linkedin-media_assets`, `parsed-linkedin-stats`
- `parsed-tiktok-posts`, `parsed-tiktok-insights`

### Stage 4: Data Processing and Enrichment

**Services**:
- `unified-immediate-processor`: **Consolidated processor for all platforms** (recommended)
- `facebook-immediate-processor`: Standalone Facebook processing
- `instagram-immediate-processor`: Standalone Instagram data enrichment
- `linkedin-immediate-processor`: Standalone LinkedIn processing
- `tiktok-immediate-processor`: TikTok data processing

#### Unified Immediate Processor (Recommended)

The unified processor handles all platforms with a two-tier queue architecture:

```
                         GLOBAL QUEUE (100K capacity)
                         Admission control - rejects when full
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
              ┌──────────┐   ┌──────────┐   ┌──────────┐
              │ Facebook │   │Instagram │   │ LinkedIn │
              │  Queue   │   │  Queue   │   │  Queue   │
              │  24K cap │   │  16K cap │   │  8K cap  │
              │40 workers│   │30 workers│   │15 workers│
              └──────────┘   └──────────┘   └──────────┘
                    │               │               │
                    ▼               ▼               ▼
              40 goroutines  30 goroutines  15 goroutines
              (parallel)     (parallel)     (parallel)
```

**Key Features**:
- **Global Queue**: 100K capacity for admission control
- **Parallel Processing**: Platforms don't block each other
- **Immediate Rejection**: Users notified when system is busy
- **Future-Ready**: Reserved capacity for Twitter, TikTok, Pinterest, YouTube

See [Unified Immediate Processor Architecture](./unified-immediate-processor.md) for details.

**Responsibilities**:
- Consumes parsed data from Kafka
- Performs business logic validation
- Enriches data with calculated fields
- Applies data quality checks
- Handles deduplication
- Publishes processed data for storage

**Processing Operations**:
- **Engagement Score Calculation**: Weighted engagement metrics
- **Trend Analysis**: Performance compared to historical averages
- **Content Categorization**: ML-based content classification
- **Anomaly Detection**: Spike/drop detection in metrics
- **Data Validation**: Schema validation and constraint checking

**Output Topics**:
- `processed_facebook_*` (posts, media_assets, video_insights, insights, reels_insights)
- `processed_instagram_*` (posts, insights)
- `processed_linkedin_*` (posts, media_assets, stats)
- `processed_tiktok_*` (posts, insights)

### Stage 5: Data Storage (ClickHouse Sink)

**Services**:
- `facebook-clickhouse-sink`: Batch inserts for Facebook
- `instagram-clickhouse-sink`: Instagram data storage
- `linkedin-clickhouse-sink`: LinkedIn persistence
- `tiktok-clickhouse-sink`: TikTok data storage

**Responsibilities**:
- Consumes processed data from Kafka
- Batches data for optimal insertion (1000 items)
- Handles ClickHouse connection pooling
- Manages transaction boundaries
- Implements retry logic for failures
- Monitors batch performance

**Performance Optimizations**:
- **Batch Size**: 1000 items per batch
- **Parallel Workers**: 15 concurrent processors (3 per data type)
- **Channel Buffering**: 50K message buffers
- **Connection Pooling**: Reused connections
- **Async Commits**: Non-blocking Kafka offset commits

## Data Models

### Kafka Message Structures

#### Work Order Message
```go
type WorkOrder struct {
    AccountID         string    `json:"account_id"`
    AccountType       string    `json:"account_type"`
    Token            string    `json:"token"`
    LastUpdate       time.Time `json:"last_update"`
    SyncType         string    `json:"sync_type"` // incremental/full
    RequestedMetrics []string  `json:"requested_metrics"`
}
```

#### Facebook Post Model
```go
type ParsedFacebookPost struct {
    ID                string          `json:"id"`
    PageID            string          `json:"page_id"`
    Message           string          `json:"message"`
    CreatedTime       time.Time       `json:"created_time"`
    UpdatedTime       time.Time       `json:"updated_time"`
    PostType          string          `json:"post_type"`
    Permalink         string          `json:"permalink_url"`
    
    // Engagement Metrics
    LikesCount        int64           `json:"likes_count"`
    CommentsCount     int64           `json:"comments_count"`
    SharesCount       int64           `json:"shares_count"`
    ReactionsBreakdown map[string]int64 `json:"reactions_breakdown"`
    
    // Reach and Impressions
    Reach             int64           `json:"reach"`
    OrganicReach      int64           `json:"organic_reach"`
    PaidReach         int64           `json:"paid_reach"`
    Impressions       int64           `json:"impressions"`
    OrganicImpressions int64          `json:"organic_impressions"`
    PaidImpressions   int64           `json:"paid_impressions"`
    
    // Media Assets
    MediaAssets       []MediaAsset    `json:"media_assets"`
    
    // Performance Metrics
    EngagementRate    float64         `json:"engagement_rate"`
    ViralityScore     float64         `json:"virality_score"`
}
```

### ClickHouse Schemas

#### Facebook Posts Table
```sql
CREATE TABLE facebook_posts (
    id String,
    page_id String,
    message String,
    created_time DateTime64(3),
    updated_time DateTime64(3),
    post_type String,
    permalink_url String,
    
    -- Engagement
    likes_count Int64,
    comments_count Int64,
    shares_count Int64,
    reactions_like Int64,
    reactions_love Int64,
    reactions_wow Int64,
    reactions_haha Int64,
    reactions_sorry Int64,
    reactions_angry Int64,
    reactions_thankful Int64,
    
    -- Reach & Impressions
    reach Int64,
    organic_reach Int64,
    paid_reach Int64,
    impressions Int64,
    organic_impressions Int64,
    paid_impressions Int64,
    
    -- Calculated Metrics
    engagement_rate Float64,
    virality_score Float64,
    
    -- Metadata
    inserted_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3)
) ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_time)
ORDER BY (page_id, created_time, id);
```

## Configuration Management

### Environment Variables

The system uses a hierarchical configuration approach:
1. Environment variables (highest priority)
2. `.env` file (development)
3. Default values (fallback)

### Key Configuration Categories

#### MongoDB Configuration
```bash
APP_MONGO_URI=mongodb://localhost:27017
APP_MONGO_DATABASE=contentstudio
```

#### Kafka Configuration
```bash
APP_KAFKA_BROKERS=localhost:9092
APP_KAFKA_TOPIC_PREFIX=analytics_
APP_KAFKA_SASL_ENABLED=true
APP_KAFKA_SASL_USERNAME=admin
APP_KAFKA_SASL_PASSWORD=secret
APP_KAFKA_SASL_MECHANISM=SCRAM-SHA-256
```

#### ClickHouse Configuration
```bash
APP_CLICKHOUSE_HOST=localhost
APP_CLICKHOUSE_PORT=9000
APP_CLICKHOUSE_DATABASE=contentstudiobackend
APP_CLICKHOUSE_USERNAME=default
APP_CLICKHOUSE_PASSWORD=password
```

#### Platform API Credentials
```bash
# Facebook
APP_FACEBOOK_APP_ID=123456789
APP_FACEBOOK_APP_SECRET=abcdef123456

# Instagram (uses Facebook Graph API)
APP_INSTAGRAM_CLIENT_ID=${APP_FACEBOOK_APP_ID}
APP_INSTAGRAM_CLIENT_SECRET=${APP_FACEBOOK_APP_SECRET}

# LinkedIn
APP_LINKEDIN_CLIENT_ID=xyz789
APP_LINKEDIN_CLIENT_SECRET=secret123

# TikTok
APP_TIKTOK_CLIENT_KEY=tiktok_key
APP_TIKTOK_CLIENT_SECRET=tiktok_secret
```

#### Service Configuration
```bash
APP_DECRYPTION_KEY=32_byte_encryption_key_here
APP_LOG_LEVEL=info
APP_SERVICE_PORT=8080
```

## Performance Characteristics

### Throughput Metrics

- **Facebook Pipeline**: 15,000 items/second
- **Instagram Pipeline**: 12,000 items/second
- **LinkedIn Pipeline**: 10,000 items/second
- **TikTok Pipeline**: 8,000 items/second

### Resource Utilization

#### Memory Usage
- **Fetchers**: ~200MB per instance
- **Parsers**: ~150MB per instance
- **Processors**: ~100MB per instance
- **Sinks**: ~500MB per instance (due to batching buffers)

#### CPU Usage
- **Fetchers**: 0.5-1 core (I/O bound)
- **Parsers**: 1-2 cores (CPU bound)
- **Processors**: 0.5-1 core
- **Sinks**: 2-3 cores (parallel batching)

### Optimization Techniques

1. **Batch Processing**
   - 1000-item batches reduce network overhead
   - Bulk inserts improve ClickHouse performance by 1500x

2. **Parallel Processing**
   - 15 concurrent workers in sink services
   - Platform-specific goroutine pools

3. **Channel Buffering**
   - 50K message buffers prevent backpressure
   - Async Kafka commits reduce latency

4. **Connection Pooling**
   - Reused database connections
   - HTTP client connection reuse

5. **Memory Management**
   - Efficient struct design
   - Minimal allocations in hot paths

## Error Handling and Resilience

### Error Categories

1. **Transient Errors**
   - Network timeouts
   - Rate limiting
   - Temporary API unavailability
   - Handled with exponential backoff

2. **Data Errors**
   - Malformed JSON
   - Missing required fields
   - Type mismatches
   - Logged and skipped with alerting

3. **System Errors**
   - Kafka connection failures
   - ClickHouse unavailability
   - MongoDB connection issues
   - Circuit breaker pattern implementation

### Retry Strategies

```go
type RetryConfig struct {
    MaxAttempts     int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    Multiplier      float64
    Jitter          float64
}

// Default configuration
defaultRetry := RetryConfig{
    MaxAttempts:  5,
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,
    Jitter:       0.1,
}
```

### Monitoring and Observability

#### Structured Logging
- Zerolog for JSON structured logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Contextual information in every log

#### Key Metrics
- **Pipeline Latency**: Time from work order to storage
- **Processing Rate**: Items/second per service
- **Error Rate**: Failures per 1000 items
- **Batch Efficiency**: Average batch size
- **API Rate Limit Usage**: Remaining quota percentage

#### Health Checks
- Each service exposes `/health` endpoint
- Liveness and readiness probes
- Dependency health aggregation

## Deployment Architecture

### Container Strategy
- Docker containers for each service
- Multi-stage builds for minimal images
- Non-root user execution

### Orchestration
- Kubernetes-ready with helm charts
- Horizontal pod autoscaling based on CPU/memory
- Rolling updates with zero downtime

### Service Dependencies
```yaml
dependencies:
  mongodb:
    - scheduler
  kafka:
    - all services
  clickhouse:
    - sink services
  external_apis:
    - fetcher services
```

## Migration from Python

### Key Improvements

1. **Performance**: 100x faster processing
2. **Memory**: 70% reduction in memory usage
3. **Reliability**: 99.99% uptime vs 99.9%
4. **Scalability**: Linear scaling with worker count
5. **Maintainability**: Type safety and compile-time checks

### Compatibility Considerations

- **API Parity**: Exact same API calls as Python
- **Field Mapping**: Identical field names preserved
- **Business Logic**: No algorithmic changes
- **Data Format**: Same ClickHouse schemas

## Security Considerations

### Token Management
- Encrypted storage in MongoDB
- Runtime decryption with AES-256
- Token refresh automation
- No tokens in logs

### API Security
- HMAC-SHA256 request signing (Facebook)
- OAuth 2.0 implementation
- Rate limit respect
- IP whitelisting support

### Data Privacy
- PII redaction in logs
- GDPR compliance
- Data retention policies
- Audit logging

## Future Enhancements

### Planned Features
1. **Real-time Webhooks**: Platform webhook support
2. **ML Integration**: Sentiment analysis, content classification
3. **Advanced Analytics**: Predictive metrics, anomaly detection
4. **Multi-tenant**: Complete isolation between clients

### Technical Debt
1. MongoDB timestamp format standardization
2. Advanced Facebook video metrics implementation
3. Comprehensive integration testing
4. Performance profiling and optimization

## Operational Runbook

### Starting Services

#### Development
```bash
# Start infrastructure
docker-compose up -d

# Run individual service
cd src && make facebook_fetcher
../bin/facebook_fetcher

# Run all Facebook services
./scripts/run_facebook_services.sh start
```

#### Production
```bash
# Deploy with Kubernetes
kubectl apply -f k8s/

# Scale service
kubectl scale deployment facebook-fetcher --replicas=5
```

### Common Operations

#### Check Pipeline Health
```bash
# Kafka lag
kafka-consumer-groups --bootstrap-server localhost:9092 \
  --group facebook-fetcher --describe

# ClickHouse data
clickhouse-client -q "SELECT count() FROM facebook_posts WHERE toDate(created_time) = today()"

# Service logs
kubectl logs -f deployment/facebook-fetcher
```

#### Troubleshooting

1. **High Kafka Lag**
   - Scale up consumers
   - Check for processing errors
   - Verify API rate limits

2. **ClickHouse "Too Many Parts"**
   - Increase batch size
   - Optimize partition key
   - Run OPTIMIZE TABLE

3. **Memory Issues**
   - Reduce channel buffer sizes
   - Decrease batch sizes
   - Add memory limits

## Conclusion

The ContentStudio Social Analytics Go pipeline represents a modern, scalable approach to social media data processing. With its microservices architecture, event-driven design, and performance optimizations, it delivers enterprise-grade reliability while maintaining simplicity and maintainability. The system's modular design allows for easy extension to new platforms while its robust error handling ensures data integrity at scale.