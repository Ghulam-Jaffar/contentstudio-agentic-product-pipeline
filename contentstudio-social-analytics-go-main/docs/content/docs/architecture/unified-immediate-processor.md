---
title: Unified Immediate Processor Architecture
description: Queue-based architecture for processing immediate analytics requests across all social platforms
---

## Overview

The Unified Immediate Processor is a consolidated service that handles real-time analytics processing requests for all social platforms (Facebook, Instagram, LinkedIn) using a two-tier queue architecture with parallel processing.

## Queue Architecture

```
                              ┌─────────────────────────────────┐
                              │       GLOBAL QUEUE              │
                              │   Capacity: 100,000             │
                              │                                 │
                              │   Admission control for all     │
                              │   platforms combined            │
                              │                                 │
                              │   If full → Reject immediately  │
                              │   "System busy, try later"      │
                              └───────────────┬─────────────────┘
                                              │
                                              │ Admitted requests
                                              │ (routed by platform)
                                              ▼
                    ┌─────────────────────────┼─────────────────────────┐
                    │                         │                         │
                    ▼                         ▼                         ▼
          ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
          │  Facebook Queue │       │ Instagram Queue │       │  LinkedIn Queue │
          │  Capacity: 24K  │       │  Capacity: 16K  │       │  Capacity: 8K   │
          │  Workers: 40    │       │  Workers: 30    │       │  Workers: 15    │
          └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
                   │                         │                         │
                   │ Parallel                │ Parallel                │ Parallel
                   │ goroutines              │ goroutines              │ goroutines
                   ▼                         ▼                         ▼
             40 Workers                 30 Workers                15 Workers
             (concurrent)              (concurrent)              (concurrent)
```

## Queue Specifications

### Global Queue

| Property | Value | Description |
|----------|-------|-------------|
| **Capacity** | 100,000 | Maximum concurrent requests across all platforms |
| **Purpose** | Admission control | First-level check before platform routing |
| **Rejection** | Immediate | Returns "system busy" when full |
| **Future Ready** | Yes | 52K reserved for Twitter, TikTok, Pinterest, YouTube |

### Platform Queues

| Platform | Queue Size | Workers | Max Capacity | Rationale |
|----------|------------|---------|--------------|-----------|
| **Facebook** | 500 | 40 | 24,000 | Highest volume, most complex API |
| **Instagram** | 400 | 30 | 16,000 | Medium volume, media-heavy |
| **LinkedIn** | 200 | 15 | 8,000 | Lower volume, rate-limited API |
| *Twitter* | *300* | *25* | *12,000* | *Future* |
| *TikTok* | *250* | *20* | *10,000* | *Future* |
| *Pinterest* | *200* | *15* | *8,000* | *Future* |
| *YouTube* | *250* | *20* | *10,000* | *Future* |

## Request Flow

### Step 1: Global Queue Admission
```
Kafka Message → Check global queue capacity
                      │
                      ├── Capacity available → Admit (increment counter)
                      │                             │
                      │                             ▼
                      │                        Route to platform
                      │
                      └── Queue full → REJECT immediately
                                       Log: "SYSTEM BUSY"
                                       User notified to retry later
```

### Step 2: Platform Queue Routing
```
Admitted Request → Route to platform queue by topic/platform field
                         │
                         ├── Platform queue has space → Enqueue
                         │                                  │
                         │                                  ▼
                         │                             Worker picks up
                         │
                         └── Platform queue full → Release global slot
                                                   Drop request
                                                   Log warning
```

### Step 3: Parallel Processing
```
Platform Queue → Worker Pool (goroutines)
                      │
                      ├── Worker 1 ─┐
                      ├── Worker 2 ─┼── Process concurrently
                      ├── Worker 3 ─┤   (don't block each other)
                      │    ...      │
                      └── Worker N ─┘
                              │
                              ▼
                         On completion:
                         - Release global queue slot
                         - Increment processed counter
                         - Log success/failure
```

## Parallel Processing Model

### Key Characteristics

1. **Platform Independence**: Facebook, Instagram, and LinkedIn process independently
   - Facebook processing does NOT block Instagram
   - Each platform has dedicated goroutines

2. **Worker Concurrency**: Within each platform, workers process simultaneously
   - 40 Facebook accounts can be processed at the same time
   - Workers read from shared channel (Go's channel synchronization)

3. **No Cross-Platform Blocking**: Slow LinkedIn API doesn't affect Facebook throughput

### Example Concurrent State

```
Time T1:
├── Facebook: Workers 1-40 all processing different accounts
├── Instagram: Workers 1-30 all processing different accounts  
└── LinkedIn: Workers 1-15 all processing different accounts

Total concurrent: 85 accounts being processed simultaneously
```

## Deployment with Replicas

### 3-Replica Configuration

```
                    Load Balancer / Kafka Consumer Group
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
    ┌─────────┐          ┌─────────┐          ┌─────────┐
    │Replica 1│          │Replica 2│          │Replica 3│
    │         │          │         │          │         │
    │Global:  │          │Global:  │          │Global:  │
    │ 33K cap │          │ 33K cap │          │ 33K cap │
    │         │          │         │          │         │
    │FB:  8K  │          │FB:  8K  │          │FB:  8K  │
    │IG: 5.3K │          │IG: 5.3K │          │IG: 5.3K │
    │LI: 2.7K │          │LI: 2.7K │          │LI: 2.7K │
    │         │          │         │          │         │
    │Workers: │          │Workers: │          │Workers: │
    │ FB: 40  │          │ FB: 40  │          │ FB: 40  │
    │ IG: 30  │          │ IG: 30  │          │ IG: 30  │
    │ LI: 15  │          │ LI: 15  │          │ LI: 15  │
    └─────────┘          └─────────┘          └─────────┘
    
    Total System Capacity:
    - Global: 100K requests
    - Workers: 255 concurrent (85 × 3 replicas)
```

### Kafka Consumer Group

- All replicas share consumer group: `unified-immediate-processor-group`
- Kafka distributes partitions across replicas
- Each message processed by exactly one replica

## Statistics and Monitoring

### Logged Every 30 Seconds

#### Global Queue Stats
```json
{
  "global_current": 45000,
  "global_capacity": 100000,
  "global_admitted": 1250000,
  "global_rejected": 5000,
  "global_utilization_pct": 45.0
}
```

#### Per-Platform Stats
```json
{
  "platform": "facebook",
  "queue_depth": 350,
  "queue_capacity": 500,
  "processed": 450000,
  "dropped": 100,
  "utilization_pct": 70.0
}
```

### Key Metrics to Monitor

| Metric | Warning Threshold | Critical Threshold |
|--------|-------------------|-------------------|
| Global utilization | > 70% | > 90% |
| Platform queue depth | > 80% capacity | > 95% capacity |
| Rejected requests | > 100/min | > 1000/min |
| Processing duration | > 30s avg | > 60s avg |

## Data Persistence Considerations

### Current Behavior (In-Memory Queues)

- **Queues**: Go channels (in-memory)
- **On Crash**: All queued work orders are LOST
- **Kafka Offset**: Committed BEFORE processing (at-most-once)
- **Recovery**: Lost messages are NOT re-delivered

### Why This Is Acceptable

1. **ClickHouse Idempotency**: `ReplacingMergeTree` handles duplicates
2. **User Retry**: Users can manually retry failed requests
3. **Simplicity**: No external queue dependency (Redis, RabbitMQ)
4. **Performance**: In-memory is fastest

### Alternative Approaches (Not Implemented)

| Approach | Pros | Cons |
|----------|------|------|
| Commit after processing | Zero loss | Slower, needs idempotency |
| Redis-backed queue | Persistent | External dependency |
| Kafka as queue (no buffer) | Native persistence | Less control |

## Adding New Platforms

### Step 1: Add Platform Config

```go
var PlatformSettings = map[string]PlatformConfig{
    "facebook":  {Workers: 40, QueueSize: 500, MaxCapacity: 24000},
    "instagram": {Workers: 30, QueueSize: 400, MaxCapacity: 16000},
    "linkedin":  {Workers: 15, QueueSize: 200, MaxCapacity: 8000},
    // Add new platform:
    "twitter":   {Workers: 25, QueueSize: 300, MaxCapacity: 12000},
}
```

### Step 2: Import Processor Package

```go
import (
    twprocessor "github.com/.../twitter-immediate-processor/processor"
)
```

### Step 3: Initialize Processor

```go
processor := &UnifiedProcessor{
    // ... existing processors
    twitterProcessor: twprocessor.New(mongoRepo, sink, producer, notifier, pusherClient, log, cfg),
}
```

### Step 4: Add Processing Case

```go
case "twitter":
    err = p.twitterProcessor.ProcessAccount(ctx, twprocessor.WorkOrder{
        ID:          wo.ID,
        AccountID:   wo.AccountID,
        AccessToken: wo.AccessToken,
        // ... other fields
    })
```

### Step 5: Add Kafka Topic

```go
topics := []string{
    "immediate-work-order-facebook",
    "immediate-work-order-instagram",
    "immediate-work-order-linkedin",
    "immediate-work-order-twitter",  // Add new topic
}
```

## Configuration

### Command-Line Flags

```bash
# Scale workers (multiply all worker counts)
./unified_immediate_processor -workerMultiplier=1.5

# Result:
# Facebook: 40 × 1.5 = 60 workers
# Instagram: 30 × 1.5 = 45 workers
# LinkedIn: 15 × 1.5 = 22 workers
```

### Environment Variables

Same as other services (see main configuration docs):
- `APP_MONGO_*` - MongoDB connection
- `APP_KAFKA_*` - Kafka brokers
- `APP_CLICKHOUSE_*` - ClickHouse connection
- `APP_LOG_LEVEL` - Logging verbosity

## Related Services

### Processor Packages (Reusable Logic)

| Package | Location | Description |
|---------|----------|-------------|
| `fbprocessor` | `services/facebook/facebook-immediate-processor/processor/` | Facebook processing logic |
| `igprocessor` | `services/instagram/instagram-immediate-processor/processor/` | Instagram processing logic |
| `liprocessor` | `services/linkedin/linkedin-immediate-processor/processor/` | LinkedIn processing logic |

### Standalone Services (Also Available)

For running single-platform processors independently:

- `services/facebook/facebook-immediate-processor/main.go`
- `services/instagram/instagram-immediate-processor/main.go`
- `services/linkedin/linkedin-immediate-processor/main.go`

These use the same processor packages but run as separate services.

## Build and Run

### Build

```bash
cd src
make unified_immediate_processor
```

### Run

```bash
./bin/unified_immediate_processor

# With worker scaling
./bin/unified_immediate_processor -workerMultiplier=2.0
```

### Docker

```bash
docker build -t unified-immediate-processor .
docker run -e APP_MONGO_URI=... unified-immediate-processor unified_immediate_processor
```
