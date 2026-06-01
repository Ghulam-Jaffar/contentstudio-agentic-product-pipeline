# GMB (Google My Business) API Limitations & Constraints

## API Rate Limits

### Google My Business API Rate Limits
| Limit Type | Value | Notes |
|------------|-------|-------|
| Queries per day | 10,000 | Per project |
| Queries per minute | 600 | Per project |
| Per-user rate limit | 60 req/min | Per OAuth token |

### Our Implementation Limits
| Configuration | Value |
|---------------|-------|
| HTTP Timeout | 30 seconds |
| Max Retries | 1 (in client) |
| Retry Delay | 2 seconds |
| Batch Size | 200 accounts |

## API Access Limitations

### Available vs Not Available
| Feature | Available | Notes |
|---------|-----------|-------|
| Voice of Merchant status | ✓ | Required check before metrics |
| Performance metrics | ✓ | Only if VoM verified |
| Search keywords | ✓ | Only if VoM verified, monthly |
| Local posts | ✓ | All post types |
| Reviews | ✓ | With star ratings and replies |
| Media assets | ✓ | Photos and videos |
| Call history | ✗ | Not available via API |
| Question & Answer | ✗ | Not available via analytics API |
| Booking data | ✗ | Not available (only count metric) |
| Menu data | ✗ | Not available (only click count) |
| Messaging content | ✗ | Not available via analytics |

### Voice of Merchant Gating
| VoM Status | Available Data |
|------------|---------------|
| `hasVoiceOfMerchant = true` | All 5 data types |
| `hasVoiceOfMerchant = false` | Local posts, reviews, media assets only |
| VoM check fails | Local posts, reviews, media assets only (warning logged) |

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Historical Access | Notes |
|-----------|-------------------|-------|
| Performance metrics | ~18 months | Google's rolling window |
| Search keywords | ~18 months | Monthly aggregation only |
| Local posts | All time | As long as post exists |
| Reviews | All time | As long as review exists |
| Media assets | All time | Current snapshot |

### Performance Metrics Constraints
| Restriction | Detail |
|-------------|--------|
| VoM required | Must be verified merchant to access |
| Granularity | Daily only (no hourly) |
| Impressions split | Desktop/mobile × Maps/Search (4 combinations) |
| No demographic data | Cannot break down by user demographics |
| No competitor comparison | No relative performance data |

### Search Keywords Constraints
| Restriction | Detail |
|-------------|--------|
| VoM required | Must be verified merchant to access |
| Time granularity | Monthly only (not daily/weekly) |
| Threshold | Keywords below threshold may be excluded |
| No ranking data | Position in search results not available |
| No click-through data | Only impression counts per keyword |

### Reviews Constraints
| Restriction | Detail |
|-------------|--------|
| No sentiment analysis | Star rating only, no automated sentiment |
| Reply-only | Can see replies but cannot create via analytics |
| No reviewer details | Limited info on reviewer beyond name/photo |

## Content Type Limitations

### Supported Content
| Content Type | Fetch | Analytics |
|--------------|-------|-----------|
| Performance metrics | ✓ (VoM) | Impressions, clicks, actions |
| Search keywords | ✓ (VoM) | Monthly impressions |
| Standard posts | ✓ | State, topic type |
| Event posts | ✓ | Same as standard |
| Offer posts | ✓ | Same as standard |
| Reviews | ✓ | Star rating, comments |
| Photos | ✓ | URL, dimensions, category |
| Videos | ✓ | URL, format |

### What You Can't Get
| Data | Status |
|------|--------|
| Google Ads performance | Separate Ads API |
| Detailed call analytics | Not available |
| Messaging conversations | Not available |
| Booking details | Only aggregate count |
| Food order details | Only aggregate count |
| Website conversion tracking | Not available |
| Audience demographics | Not available |

## Authentication Limitations

### OAuth2 Token Lifecycle
| Token Type | Lifespan | Notes |
|------------|----------|-------|
| Access token | 1 hour | Auto-refreshed by client |
| Refresh token | No expiry | But can be revoked by user |

### Token Refresh
- POST to Google OAuth2 token endpoint
- Requires client_id, client_secret, refresh_token
- 1 retry on failure (in HTTP client)
- If both attempts fail, account is skipped entirely

## Platform Identifier Format

```
accounts/{accountID}/locations/{locationID}
```

- Both IDs are large numeric strings
- Extracted by splitting on `/` and taking parts[1] and parts[3]
- Invalid format causes the work order to be skipped with a warning

## Error Handling Strategy

| Error Type | Behavior |
|------------|----------|
| Token refresh failure | Skip account (error returned) |
| VoM check failure | Log warning, skip perf metrics + keywords, continue others |
| Performance metrics failure | Log warning, continue to next data type |
| Search keywords failure | Log warning, continue to next data type |
| Local posts failure | Log warning, continue to next data type |
| Reviews failure | Log warning, continue to next data type |
| Media assets failure | Log warning, continue to next data type |
| ClickHouse insert failure | Log warning, continue to next data type |
