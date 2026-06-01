# ContentStudio Social Analytics Go - AI Assistant Guide

This guide helps AI assistants understand and work effectively with the ContentStudio Social Analytics Go codebase.

## Project Overview

ContentStudio Social Analytics Go is a high-performance microservices-based data pipeline for extracting, processing, and storing social media analytics data from multiple platforms into ClickHouse for real-time analytics.

### Key Architecture Principles
- **Microservices Architecture**: 17+ independent services per platform
- **Event-Driven Design**: Kafka-based asynchronous communication
- **Five-Stage Pipeline**: Scheduler → Fetcher → Parser → Processor → Sink
- **Multi-Platform Support**: Facebook, Instagram, LinkedIn, TikTok (with Twitter, Pinterest, YouTube planned)
- **Python-to-Go Migration**: Preserving exact business logic while improving performance

## Quick Start for AI Assistants

### Understanding the Codebase

1. **Start with the pipeline flow**:
   ```
   MongoDB (accounts) → Kafka (work orders) → API Fetchers → Kafka (raw data) → 
   Parsers → Kafka (parsed data) → Processors → Kafka (processed data) → 
   ClickHouse Sink → ClickHouse (analytics)
   ```

2. **Key directories to understand**:
   - `/src/cmd/services/` - All microservices entry points
   - `/src/internal/clients/` - API clients for external services
   - `/src/pkg/models/` - Data models for Kafka, ClickHouse, MongoDB
   - `/src/internal/parsing/` - Data transformation logic

3. **Platform implementation pattern** (using Facebook as reference):
   - `facebook-fetcher/` - Fetches data from Graph API
   - `facebook-posts-parser/` - Parses posts and videos
   - `facebook-immediate-processor/` - Real-time processing
   - `facebook-clickhouse-sink/` - Batch storage to ClickHouse

### Critical Implementation Rules

#### 1. Python-to-Go Migration Guidelines
- **ALWAYS** study the Python implementation first if it exists
- **PRESERVE** exact API calls, parameters, and response parsing
- **MAINTAIN** business logic while improving code quality
- **NEVER** add new features without explicit approval

#### 2. Data Model Consistency
- **MongoDB Models**: Handle legacy timestamp formats with `MongoTime`
- **Kafka Messages**: Use platform-specific message types
- **ClickHouse Tables**: Follow existing schema patterns
- **Custom Types**: Use `FacebookTime` for Facebook timestamps

#### 3. Service Implementation Pattern
Each service must:
- Use Zerolog for structured logging
- Implement graceful shutdown with context
- Handle Kafka consumer groups properly
- Follow the existing configuration pattern with Viper
- Include comprehensive error handling

## Common Tasks

### Adding a New Platform

1. **Create MongoDB Model** in `/src/pkg/models/mongo/`:
   ```go
   type PlatformAccount struct {
       ID                    primitive.ObjectID     `bson:"_id,omitempty"`
       AccountID             string                 `bson:"account_id"`
       Token                 string                 `bson:"token"`
       LastAnalyticsUpdate   *MongoTime            `bson:"last_analytics_update"`
       // ... platform-specific fields
   }
   ```

2. **Define Kafka Messages** in `/src/pkg/models/kafka/`:
   - `RawPlatformPost` - API response structure
   - `ParsedPlatformPost` - Analytics-ready structure
   - Work order messages for scheduling

3. **Create ClickHouse Schema** in `/src/sql/clickhouse/`:
   - Posts table with engagement metrics
   - Insights table for account-level metrics
   - Media assets table if applicable

4. **Implement Services** following the 5-stage pattern:
   - Account fetcher integration in scheduler
   - Platform fetcher with API client
   - Parser for data transformation
   - Immediate processor for enrichment
   - ClickHouse sink for batch storage

5. **Update Build System**:
   - Add service entries to `/src/Makefile`
   - Update Dockerfile if needed
   - Create service runner script in `/scripts/`

### Working with Existing Services

#### Facebook Implementation (Production-Ready Reference)
- **API Client**: `/src/internal/clients/social/facebook.go`
  - Graph API v19.0 with appsecret_proof
  - Comprehensive field selection
  - Pagination and rate limiting
- **Parser**: `/src/internal/parsing/facebook_parser.go`
  - 60+ video metrics processing
  - Reaction type aggregation (including THANKFUL)
  - Media asset extraction
- **Performance**: 15,000 items/second with batching

#### Key Performance Optimizations
1. **Batch Processing**: 1000-item batches in ClickHouse sink
2. **Parallel Workers**: 15 concurrent processors (3 per data type)
3. **Channel Buffering**: 50K message buffers prevent data loss
4. **Backpressure Handling**: Zero data loss guarantees

### Configuration Management

All services use environment-based configuration with this hierarchy:
1. Environment variables (highest priority)
2. `.env` file
3. Default values

Key configuration sections:
- `APP_MONGO_*` - MongoDB connection
- `APP_KAFKA_*` - Kafka brokers and SASL auth
- `APP_CLICKHOUSE_*` - ClickHouse connection
- `APP_FACEBOOK_*` - Facebook API credentials
- `APP_DECRYPTION_KEY` - Token encryption key

### Testing and Quality

#### Current State
- **No automated tests** - Manual testing only
- **No CI/CD pipeline** - Manual deployment
- Focus on production stability over test coverage

#### When Adding Code
1. **Follow existing patterns** exactly
2. **Use structured logging** with appropriate levels
3. **Handle all errors** with context wrapping
4. **Validate data** at service boundaries
5. **Monitor performance** with channel utilization logs

### Database Operations

#### ClickHouse Best Practices
```go
// Always use batch operations
batch, err := conn.PrepareBatch(ctx, `INSERT INTO table_name`)
for _, item := range items {
    err = batch.Append(item.Field1, item.Field2, ...)
}
err = batch.Send()
```

#### MongoDB Repository Pattern
```go
// Use repository pattern for data access
repo := repository.NewMongoRepository(client, database)
accounts, err := repo.GetFacebookAccounts(ctx, filter)
```

#### Kafka Consumer Pattern
```go
// Standard consumer with error handling
consumer := kafka.NewConsumer(config)
for {
    records := consumer.Fetch(ctx)
    for _, record := range records {
        // Process message
        consumer.MarkOffset(record)
    }
}
```

## Performance Characteristics

### Expected Throughput
- **Fetcher**: ~500 posts + ~100 videos per work order
- **Parser**: High-throughput concurrent processing
- **ClickHouse Sink**: 15,000 items/second (1500x improvement over individual inserts)

### Resource Usage
- **Memory**: High channel buffers (50K messages)
- **CPU**: 15 parallel processors per sink service
- **Network**: Batch operations reduce network calls

## Debugging and Monitoring

### Logging Standards
```go
log.Info().
    Str("service", "facebook-fetcher").
    Str("account_id", accountID).
    Int("posts_fetched", len(posts)).
    Msg("Fetched Facebook posts")
```

### Key Metrics to Monitor
1. **Channel Utilization**: Shows pipeline bottlenecks
2. **Batch Sizes**: Optimal at 1000 items
3. **Processing Time**: Per-stage latency
4. **Error Rates**: API failures, parsing errors

### Common Issues
1. **ClickHouse "too many parts"**: Solved with batching
2. **Kafka lag**: Increase parallel processors
3. **API rate limits**: Implement exponential backoff
4. **Memory pressure**: Reduce channel buffer sizes

## Security Considerations

1. **Token Management**: 
   - Encrypted storage in MongoDB
   - Decryption key in environment
   - Never log tokens

2. **API Security**:
   - Facebook: HMAC-SHA256 appsecret_proof
   - SASL authentication for Kafka
   - TLS for all external connections

3. **Data Privacy**:
   - No PII in logs
   - Secure configuration management
   - Proper access controls

## Development Workflow

### Local Development
1. Copy `.env.example` to `.env` and configure
2. Start dependencies: `docker-compose up -d`
3. Run migrations: Apply ClickHouse schemas
4. Build services: `make build`
5. Run services: `./scripts/run_facebook_services.sh`

### Making Changes
1. **Research Phase**: Understand existing implementation
2. **Implementation**: Follow patterns exactly
3. **Testing**: Manual testing with real data
4. **Performance**: Verify throughput meets expectations
5. **Deployment**: Build Docker image and deploy

### Code Style Guidelines
- **No comments** unless absolutely necessary
- **Descriptive variable names** over comments
- **Consistent error handling** with wrapping
- **Structured logging** at appropriate levels
- **Follow Go idioms** and best practices

## Migration from Python

### Key Differences to Preserve
1. **API Parameters**: Must match Python exactly
2. **Field Mappings**: Preserve all field names
3. **Business Logic**: No modifications allowed
4. **Error Handling**: Improve but maintain behavior

### Performance Improvements
1. **Concurrency**: Go routines for parallel processing
2. **Batching**: Bulk operations everywhere
3. **Memory Management**: Efficient channel usage
4. **Connection Pooling**: Reuse database connections

## Future Considerations

### Planned Improvements
1. **Automated Testing**: Unit and integration tests
2. **CI/CD Pipeline**: GitHub Actions setup
3. **Monitoring**: Prometheus metrics and Grafana
4. **Service Mesh**: Kubernetes deployment

### Technical Debt
1. MongoDB timestamp inconsistencies
2. Missing advanced Facebook metrics
3. No webhook support for real-time updates
4. Limited error recovery mechanisms

## Commands Reference

### Build Commands
```bash
make build              # Build all services
make clean             # Clean build artifacts
make tidy              # Update Go modules
```

### Service Management
```bash
./scripts/run_facebook_services.sh start    # Start Facebook pipeline
./scripts/run_facebook_services.sh stop     # Stop Facebook pipeline
./scripts/run_facebook_services.sh status   # Check service status
```

### Common Operations
```bash
# Check Kafka topics
kafka-topics --list --bootstrap-server localhost:9092

# Monitor ClickHouse
clickhouse-client -q "SELECT count() FROM facebook_posts"

# View service logs
tail -f logs/facebook-fetcher.log
```

## Important Files Reference

### Configuration
- `/src/internal/config/config.go` - Central configuration management
- `.env` - Environment configuration

### Models
- `/src/pkg/models/kafka/` - Message definitions
- `/src/pkg/models/clickhouse/` - Table schemas
- `/src/pkg/models/mongo/` - Account models

### Services
- `/src/cmd/services/*/main.go` - Service entry points
- `/src/internal/clients/` - External API clients
- `/src/internal/parsing/` - Data transformation

### Build & Deploy
- `/src/Makefile` - Build configuration
- `/Dockerfile` - Container definition
- `/scripts/` - Operational scripts

## Contact and Support

This is an internal ContentStudio project. For questions:
1. Check existing documentation in `/docs/`
2. Review Python implementation for business logic
3. Study Facebook implementation as reference
4. Follow established patterns exactly