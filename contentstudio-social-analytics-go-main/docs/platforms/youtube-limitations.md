# YouTube API Limitations & Constraints

## API Rate Limits

### Quota System (Daily)
| Resource | Quota Cost | Notes |
|----------|------------|-------|
| Read operations | 1 unit | Most GET requests |
| Write operations | 50 units | POST/PUT/DELETE |
| Video upload | 1600 units | Per video |
| Search | 100 units | Per search request |

### Daily Quota Limits
| Limit Type | Value |
|------------|-------|
| Default quota | 10,000 units/day |
| With approval | Up to 1,000,000 units/day |

### Our Implementation Limits
| Limit Type | Value |
|------------|-------|
| Per-client | 5 RPS |
| Burst | 10 requests |
| Backoff base | 500ms |
| Backoff max | 8 seconds |

## Dual API Architecture

### YouTube Data API v3
| Feature | Purpose | Quota Cost |
|---------|---------|------------|
| Channels | Channel metadata | 1 unit |
| Videos | Video metadata | 1 unit |
| Activities | Upload activities | 1 unit |
| Search | Find content | 100 units |

### YouTube Analytics API v2
| Feature | Purpose | Quota Cost |
|---------|---------|------------|
| Reports | Analytics data | 1 unit |
| Group reports | Grouped analytics | 1 unit |

### API Differences
| Aspect | Data API | Analytics API |
|--------|----------|---------------|
| Data freshness | Real-time | 2-3 day delay |
| Historical data | Limited | Up to 18 months |
| Metrics available | Basic stats | Detailed analytics |
| Dimensions | None | Multiple |

## Data Delay Limitation

### YouTube Analytics API Delay
```
┌─────────────────────────────────────────────────────────┐
│  Today: Jan 30                                          │
│  ├── Data available: up to Jan 27 (3 days ago)         │
│  ├── Jan 28-30: Data incomplete/unavailable            │
│  └── Best practice: Use endDate = now - 3 days         │
└─────────────────────────────────────────────────────────┘
```

### Delay by Metric Type
| Metric Type | Typical Delay |
|-------------|---------------|
| Views | 24-48 hours |
| Watch time | 24-48 hours |
| Revenue | 2-3 days |
| Demographics | 2-3 days |
| Traffic sources | 24-48 hours |

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Historical Access | Notes |
|-----------|-------------------|-------|
| Video analytics | 18 months | Rolling window |
| Channel analytics | 18 months | Rolling window |
| Lifetime stats | All time | Via Data API |
| Revenue data | 18 months | Partner-only |

### Date Range Limits
| Query Type | Max Range |
|------------|-----------|
| Single query | 365 days |
| Dimensions query | Varies by dimension |
| Daily granularity | 365 days |

## Content Type Limitations

### Supported Content
| Content Type | Data API | Analytics API |
|--------------|----------|---------------|
| Regular videos | ✓ | ✓ |
| Shorts | ✓ | Partial |
| Live streams | ✓ | Limited |
| Premieres | ✓ | Limited |
| Private videos | Owner only | Owner only |
| Unlisted videos | Owner only | Owner only |

### Shorts Detection Challenge
```go
// No direct API field for Shorts
// Must detect via:
// 1. Duration <= 60 seconds
// 2. URL redirect check (expensive)

func IsShort(video Video) bool {
    duration := parseDuration(video.ContentDetails.Duration)
    return duration <= 60
}
```

### Live Stream Limitations
- Real-time analytics limited
- Concurrent viewers not in reports
- Chat data not available via Analytics API
- Post-live becomes regular video

## Metrics Limitations

### Available vs Partner-Only Metrics
| Metric | All Channels | Partner-Only |
|--------|--------------|--------------|
| Views | ✓ | ✓ |
| Watch time | ✓ | ✓ |
| Subscribers | ✓ | ✓ |
| Revenue | ✗ | ✓ |
| Red views | ✗ | ✓ |
| Red watch time | ✗ | ✓ |
| CPM/RPM | ✗ | ✓ |

### Metric Accuracy Issues
| Issue | Impact |
|-------|--------|
| View count differences | Data API vs Analytics API can differ |
| Watch time rounding | May be rounded to minutes |
| Subscriber count | Public count may be abbreviated |
| Real-time vs final | Numbers finalize after 48-72 hours |

## Dimension Limitations

### Traffic Source Dimensions
```go
const (
    TrafficSourceYTSearch     = "YT_SEARCH"
    TrafficSourceExtURL       = "EXT_URL"
    TrafficSourceRelatedVideo = "RELATED_VIDEO"
    TrafficSourcePlaylist     = "PLAYLIST"
    TrafficSourceSubscriber   = "SUBSCRIBER"
    TrafficSourceNotification = "NOTIFICATION"
    TrafficSourceShorts       = "SHORTS"
    // ... many more
)
```

### Dimension Combination Limits
- Not all dimensions can be combined
- Some combinations return errors
- Query must be valid for date range

## Authentication Limitations

### OAuth2 Token Lifecycle
| Token Type | Lifespan | Refresh |
|------------|----------|---------|
| Access token | 1 hour | Via refresh token |
| Refresh token | 6 months (inactive) | Re-authenticate |

### Token Issues
| Issue | Impact |
|-------|--------|
| 1-hour expiration | Must refresh frequently |
| Refresh token revocation | User must re-authorize |
| Scope changes | May invalidate tokens |

### Required Scopes
```
https://www.googleapis.com/auth/youtube.readonly
https://www.googleapis.com/auth/yt-analytics.readonly
https://www.googleapis.com/auth/yt-analytics-monetary.readonly  # Partner only
```

## Pagination Limitations

### Data API Pagination
```go
type PageInfo struct {
    TotalResults   int    `json:"totalResults"`
    ResultsPerPage int    `json:"resultsPerPage"`
}
// Max 50 results per page for most endpoints
```

### Analytics API Pagination
- No pagination for reports
- Must use date ranges to chunk
- Large channels may timeout

## Error Handling

### Common Error Codes
| Code | Meaning | Action |
|------|---------|--------|
| 401 | Unauthorized | Refresh token |
| 403 | Forbidden/Quota | Check quota/permissions |
| 404 | Not found | Resource deleted |
| 429 | Rate limited | Backoff and retry |
| 500 | Server error | Retry with backoff |

### YouTube-Specific Errors
| Error | Cause | Solution |
|-------|-------|----------|
| "quotaExceeded" | Daily quota hit | Wait for reset |
| "forbidden" | No access to channel | Check ownership |
| "invalidParameter" | Bad dimension combo | Fix query |

## Deduplication Challenges

### ClickHouse Deduplication
```sql
-- ReplacingMergeTree for deduplication
ENGINE = ReplacingMergeTree(inserted_at)
ORDER BY (record_id, created_at)

-- record_id = MD5(channel_id + date)
-- Ensures one row per channel per day
```

### Deduplication Issues
| Issue | Solution |
|-------|----------|
| Duplicate inserts | ReplacingMergeTree |
| Merge timing | OPTIMIZE TABLE FINAL |
| Query duplicates | Use FINAL keyword |

## Channel-Specific Limitations

### Small Channel Limits
| Requirement | Limitation |
|-------------|------------|
| Subscribers | Some metrics require 1000+ |
| Watch hours | 4000 hours for analytics |
| Verification | Some features require verification |

### Large Channel Issues
- API timeouts for channels with many videos
- Pagination required
- Higher quota consumption

## Competitor Analysis Limitations

### Public Data Only
| Available | Not Available |
|-----------|---------------|
| Public video stats | Analytics data |
| Subscriber count (public) | Exact count (if hidden) |
| Video titles/descriptions | Revenue data |
| View counts | Watch time |

## Known Issues & Quirks

### API Behavior Issues
| Issue | Description |
|-------|-------------|
| Shorts in activities | May not appear in activities feed |
| Duration format | ISO 8601 (PT1H2M3S) parsing needed |
| Thumbnail URLs | Multiple sizes, may be unavailable |
| Country restrictions | Some videos have limited data |

### Data Consistency Issues
- Analytics vs Data API counts differ
- Real-time counts vs final counts
- Timezone handling (UTC only)

## Recommendations

### Best Practices
1. **Use 3-day offset** for Analytics API end date
2. **Cache video metadata** to reduce quota
3. **Batch video detail requests** (max 50 IDs)
4. **Implement exponential backoff** for rate limits
5. **Monitor quota usage** daily

### Quota Optimization
| Strategy | Savings |
|----------|---------|
| Batch requests | Significant |
| Cache responses | Significant |
| Incremental sync | Moderate |
| Reduce search | 100 units per search |

### Monitoring
- Track daily quota usage
- Alert at 80% quota
- Log API errors with context
- Monitor token refresh success

## Summary of Key Limitations

1. **2-3 day data delay** - Analytics not real-time
2. **Daily quota limits** - 10,000 units default
3. **1-hour token expiration** - Frequent refreshes
4. **No Shorts API** - Must detect via duration
5. **Partner-only metrics** - Revenue, Red views
6. **18-month data retention** - Historical limit
7. **Dimension combinations** - Not all valid
8. **Deduplication required** - Multiple fetch = duplicates
