# Pinterest API Limitations & Constraints

## API Rate Limits

### Pinterest API v5 Rate Limits
| Limit Type | Value | Notes |
|------------|-------|-------|
| Standard | 1000 requests/minute | Per access token |
| Write operations | 100 requests/minute | Creates/updates |
| Analytics | 500 requests/minute | Analytics endpoints |

### Our Implementation Limits
| Configuration | Value |
|---------------|-------|
| Requests per second | 1 RPS |
| Burst | 1 request |
| HTTP Timeout | 30 seconds |
| Max Retries | 3 |
| Base Backoff | 1 second |
| Max Backoff | 10 seconds |

## API Access Limitations

### Available vs Not Available
| Feature | Available | Notes |
|---------|-----------|-------|
| User account info | ✓ | Profile, followers, monthly views |
| User analytics | ✓ | Daily metrics for account |
| Board list | ✓ | All user boards |
| Board details | ✓ | Single board metadata |
| Board pins | ✓ | Paginated with bookmark |
| Pin analytics | ✓ | Daily per-pin metrics |
| Multi-pin analytics | ✓ | Batch of 25 pins max |
| Comment data | ✗ | Not available via API |
| Follower list | ✗ | Not available |
| Pin impressions by source | ✗ | Not broken down |
| Audience demographics | ✗ | Not available via v5 API |

### Critical Missing Features
1. **No comment data** — Cannot analyze sentiment or engagement quality
2. **No audience demographics** — Age, gender, location breakdowns not available
3. **No follower list** — Cannot analyze follower profiles
4. **No impression source breakdown** — Cannot distinguish organic vs paid

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Historical Access | Notes |
|-----------|-------------------|-------|
| Pin analytics | 90 days | Maximum analytics window |
| User analytics | 90 days | Maximum analytics window |
| Board data | Current snapshot | No historical board metrics |
| Pin metadata | No limit | As long as pin exists |

### Analytics Data Lag
| Restriction | Detail |
|-------------|--------|
| Last 2 days excluded | Analytics incomplete for recent days |
| PROCESSING status | Data still being computed, skipped |
| BEFORE_PIN_CREATED | Pin didn't exist on date, skipped |
| BEFORE_BUSINESS_CREATED | Business account didn't exist, skipped |

### What You Can't Get
| Data | Status |
|------|--------|
| Pin impression source | Not available |
| Audience demographics | Not available |
| Comment content | Not available |
| Follower growth history | Not available (snapshot only) |
| Board analytics over time | Not available |
| Shopping/product analytics | Not available |
| Ad performance | Separate ads API required |

## Content Type Limitations

### Supported Content
| Content Type | Fetch | Analytics |
|--------------|-------|-----------|
| Standard pins | ✓ | ✓ Daily metrics |
| Video pins | ✓ | ✓ Video-specific metrics |
| Idea pins | Partial | Limited metrics |
| Product pins | ✓ | Standard metrics only |
| Promoted pins | ✓ | has_been_promoted flag only |
| Private boards | ✗ | Skipped for analytics |
| Secret pins | ✗ | Not accessible |

### Pin Metadata Limitations
- No hashtag performance data
- No pin scheduling information
- Product tags available but limited detail
- Media dimensions may be missing for some pin types

## Authentication Limitations

### OAuth2 Token Lifecycle
| Token Type | Lifespan | Notes |
|------------|----------|-------|
| Access token | 30 days | Must handle refresh |
| Refresh token | 365 days | Long-lived but can expire |

### Token Issues
| Issue | Impact |
|-------|--------|
| 401 on expiration | Immediate failure, no retry |
| No auto-refresh in pipeline | Must be refreshed externally |
| Encrypted storage | Requires decryption key at runtime |

## Pagination Limitations

### Bookmark-Based Pagination
```go
type PinterestPinsResponse struct {
    Items    []PinterestPin `json:"items"`
    Bookmark string         `json:"bookmark,omitempty"`
}
```

### Pagination Issues
| Issue | Impact |
|-------|--------|
| Max 250 per page | Many requests for large boards |
| Bookmark expiration | May expire between requests |
| No total count | Unknown total pins upfront |
| No date filtering for pins | Must fetch all, filter locally |

## Multi-Pin Analytics Limitations

### Batch Analytics
| Constraint | Value |
|------------|-------|
| Max pins per batch | 25 |
| Fallback on failure | Individual pin requests |
| Response format inconsistency | Multiple response shapes handled |

### Response Format Issues
The multi-pin analytics endpoint returns inconsistent response formats:
- Sometimes keyed by pin ID at top level
- Sometimes as array with `items`, `data`, or `results` key
- Requires normalization logic to handle all variants

## Board-Specific Limitations

### Board Account Type
| Limitation | Detail |
|------------|--------|
| Single board only | Board accounts fetch one board |
| No cross-board analytics | Cannot aggregate across boards |
| No board follower details | Count only, no follower list |

### Profile Account Type
| Limitation | Detail |
|------------|--------|
| Public boards only | Private boards skipped |
| All boards fetched | No selective board filtering |
| No board-level analytics API | Only pin-level analytics per board |

## Comparison with Other Platforms

### Feature Comparison
| Feature | Pinterest | Facebook | Instagram | YouTube | TikTok |
|---------|-----------|----------|-----------|---------|--------|
| Post analytics | ✓ | ✓ | ✓ | ✓ | ✗ |
| Historical data | 90 days | 2 years | 2 years | Unlimited | ✗ |
| Audience demographics | ✗ | ✓ | ✓ | ✓ | ✗ |
| Engagement trends | ✓ | ✓ | ✓ | ✓ | ✗ |
| Date filtering | ✓ | ✓ | ✓ | ✓ | ✗ |
| Comment data | ✗ | ✓ | ✓ | ✓ | ✗ |
| Video metrics | ✓ | ✓ | ✓ | ✓ | Partial |
| Batch analytics | ✓ (25) | ✗ | ✗ | ✗ | ✗ |
| Board/playlist support | ✓ | ✗ | ✗ | ✓ | ✗ |

### Impact on Analytics
- 90-day window limits long-term trend analysis
- No demographics prevents audience insights
- No comments prevents sentiment analysis
- Board-level analytics must be derived from pin data

## Known Issues & Quirks

### API Behavior Issues
| Issue | Description |
|-------|-------------|
| Inconsistent timestamps | created_at may be RFC3339 or bare ISO 8601 |
| Analytics lag | Last 2 days data unreliable |
| Multi-pin response variance | Response format varies between calls |
| Board pin count mismatch | API count may differ from actual pins returned |
| Monthly views staleness | May not update in real-time |

### Data Quality Issues
- Pin analytics may show ESTIMATE status for recent dates
- Engagement rate can be zero when impressions are missing
- Board collaborator_count may not reflect pending invitations
- Pin thumbnail URLs may expire or return 404

## Recommendations

### Best Practices
1. **Handle multiple timestamp formats** — Pinterest API is inconsistent
2. **Skip PROCESSING data status** — Avoid storing incomplete metrics
3. **Use multi-pin analytics** — 25x reduction in API calls
4. **Exclude last 2 days** — Analytics data is incomplete
5. **Process only PUBLIC boards** — Private boards return errors
6. **Implement bookmark pagination** — No offset-based alternative

### Workarounds
| Limitation | Workaround |
|------------|------------|
| No demographics | Not possible to workaround |
| No comments | Not possible to workaround |
| 90-day analytics limit | Store daily snapshots for long-term trends |
| No board analytics | Aggregate pin insights per board |
| Analytics lag | Exclude last 2 days from fetch window |
| Token expiration | External refresh mechanism |

### Monitoring
- Track API 429 (rate limit) frequency
- Monitor 401 errors for token expiration
- Alert on multi-pin analytics batch failures
- Log data status distribution (READY vs ESTIMATE vs PROCESSING)

## Summary of Key Limitations

1. **90-day analytics window** — Cannot access data older than 90 days
2. **No audience demographics** — Age, gender, location not available
3. **No comment data** — Cannot analyze engagement quality
4. **Analytics lag** — Last 2 days data is incomplete
5. **Private board exclusion** — Analytics only for public boards
6. **Inconsistent API formats** — Timestamp and response format variability
7. **No follower details** — Count only, no profiles
8. **25-pin batch limit** — Multi-pin analytics capped at 25
