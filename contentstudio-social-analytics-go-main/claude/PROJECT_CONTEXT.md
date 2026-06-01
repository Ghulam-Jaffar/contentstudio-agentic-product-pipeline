# ContentStudio Social Analytics Go — Project Context

## What This Project Does

High-performance microservices pipeline that pulls social media analytics data from platform APIs, processes it, and stores it in ClickHouse for the ContentStudio analytics dashboard.

---

## Pipeline Architecture

```
MongoDB (accounts)
    → Kafka work orders
    → Fetcher (calls platform API)
    → Kafka raw data
    → Parser (transforms to analytics shape)
    → Kafka parsed data
    → Processor (enrichment)
    → Kafka processed data
    → ClickHouse Sink (batch insert)
    → ClickHouse
```

Five stages per platform: **Scheduler → Fetcher → Parser → Processor → Sink**

---

## Platforms & Status

| Platform  | Status      | Fetcher | Parser | Sink | Notes |
|-----------|-------------|---------|--------|------|-------|
| Facebook  | Production  | ✅      | ✅     | ✅   | Reference implementation |
| Instagram | Production  | ✅      | ✅     | ✅   |       |
| LinkedIn  | Production  | ✅      | ✅     | ✅   |       |
| TikTok    | Production  | ✅      | ✅     | ✅   |       |
| Twitter   | In progress | ✅      | —      | —    |       |
| YouTube   | Planned     | —       | —      | —    |       |
| Pinterest | Planned     | —       | —      | —    |       |

---

## Key Directories

```
src/
  cmd/services/               # Service entry points (main.go per service)
  clients/social/             # HTTP API clients: facebook.go, instagram.go, linkedin.go, tiktok.go, twitter.go
  internal/
    config/config.go          # Central config via Viper + env vars
    parsing/                  # Data transformation logic per platform
  pkg/models/
    kafka/                    # Kafka message structs
    clickhouse/               # ClickHouse table structs
    mongo/                    # MongoDB account models
  services/
    facebook/
    instagram/
    linkedin/
    tiktok/
```

---

## ClickHouse Tables

| Platform  | Posts Table          | Insights Table       | Other                                      |
|-----------|----------------------|----------------------|--------------------------------------------|
| Facebook  | facebook_posts       | facebook_insights    | facebook_competitor_posts                  |
| Instagram | instagram_posts      | instagram_insights   | instagram_competitor_posts                 |
| LinkedIn  | linkedin_posts       | linkedin_insights    | —                                          |
| TikTok    | tiktok_posts         | tiktok_insights      | —                                          |
| Twitter   | twitter_posts        | twitter_insights     | —                                          |
| YouTube   | youtube_videos       | youtube_insights     | youtube_channel_demographics, youtube_traffic_sources, youtube_sharing_services |
| Pinterest | pinterest_pins       | pinterest_pin_insights | pinterest_users, pinterest_user_insights, pinterest_boards |
| Overview  | mv_social_daily_metrics (materialized view) | — | — |

---

## Analytics API Server

- Binary: `cmd/api-server/main.go`, port **8080**
- Routes use **GET** (not POST like the Laravel backend)
- camelCase URL segments
- 3-layer architecture: Handler → Service → Repository
- All data served from ClickHouse

### Implemented endpoints (on `features` branch)
Facebook ✅, Instagram ✅, LinkedIn ✅, YouTube ✅, Pinterest ✅, GMB ✅, Overview V2 ✅

---

## HTTP Client Patterns

### Retry (`doWithRetry`)
Every social API client wraps HTTP calls in a retry helper:
- **3 attempts**, exponential backoff: `attempt * 500ms`
- **Never retry** 401, 403 (auth errors), 429 (rate limit)
- **Retry** 5xx and network errors

### POST Body Retry — Factory Pattern
POST requests must use a factory function so the body reader is recreated fresh on each attempt:
```go
doWithRetry(ctx, "MethodName", func() (*http.Request, error) {
    return http.NewRequestWithContext(ctx, http.MethodPost, url,
        bytes.NewReader(bodyBytes))  // fresh reader each attempt
})
```

### Twitter OAuth Re-signing
Twitter OAuth 1.0a embeds timestamp+nonce in the signature. The factory calls `c.signRequest` on every attempt so each retry gets a valid fresh signature.

### Facebook `waitRate`
Facebook's `doWithRetry` calls `waitRate` internally. Never call `waitRate` explicitly before `doWithRetry` — it will double-count.

---

## Configuration

Environment variables via `.env` or shell:

| Prefix             | Purpose                        |
|--------------------|--------------------------------|
| `APP_MONGO_*`      | MongoDB connection              |
| `APP_KAFKA_*`      | Kafka brokers + SASL auth      |
| `APP_CLICKHOUSE_*` | ClickHouse connection          |
| `APP_FACEBOOK_*`   | Facebook Graph API credentials |
| `APP_DECRYPTION_KEY` | Token AES decryption key     |

---

## Logging Standards

```go
log.Info().
    Str("service", "instagram-fetcher").
    Str("instagram_id", igID).
    Int("media_count", n).
    Msg("Processed media job")
```

- Use Zerolog structured logging everywhere
- No PII in logs, never log tokens
- Log `Error` for recoverable failures that skip work
- Log `Warn` for degraded-but-continuing situations (e.g. channel full)

---

## Performance Targets

- ClickHouse sink: **15,000 items/second** via 1000-item batches
- Parallel workers: **15 concurrent processors** (3 per data type)
- Channel buffers: **50K messages** to absorb bursts

---

## Token Handling

- Tokens stored encrypted in MongoDB (AES)
- Decryption key from `APP_DECRYPTION_KEY` env var
- Plain tokens: Instagram starts with `IGAA`, Facebook starts with `EAA`
- If decryption fails → return `""` → skip account with MongoDB error record (do NOT send ciphertext to API)
