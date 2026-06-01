# TikTok API Limitations & Constraints

## API Rate Limits

### Content Posting API Limits
| Limit Type | Value | Notes |
|------------|-------|-------|
| Per-user | 100 requests/day | Per authenticated user |
| Video list | 20 videos/request | Max per page |
| Burst limit | 10 requests/minute | Short-term limit |

### Our Implementation
| Configuration | Value |
|---------------|-------|
| HTTP Timeout | 30 seconds |
| Inter-page delay | 100ms |
| Max videos per request | 20 |

## API Access Limitations

### Available vs Not Available
| Feature | Available | Notes |
|---------|-----------|-------|
| Video list | ✓ | User's own videos |
| Video details | ✓ | Basic metadata |
| User info | ✓ | Profile information |
| Video analytics | ✗ | Not available via API |
| Historical data | ✗ | No date-based queries |
| Comment data | ✗ | Not available |
| Follower list | ✗ | Not available |

### Critical Missing Features
1. **No per-video analytics API** - Cannot get historical view counts
2. **No date-based filtering** - Must fetch all and filter locally
3. **No engagement history** - Only current snapshot
4. **No audience demographics** - Not exposed via API

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Access | Notes |
|-----------|--------|-------|
| Videos | Current only | No historical analytics |
| User info | Current snapshot | No history |
| Engagement | Current counts | No time series |

### What You Can't Get
| Data | Status |
|------|--------|
| View count history | Not available |
| Follower growth | Not available |
| Engagement over time | Not available |
| Video performance trends | Not available |
| Audience insights | Not available |
| Comment content | Not available |
| Share destinations | Not available |

## Content Type Limitations

### Supported Content
| Content Type | Fetch | Analytics |
|--------------|-------|-----------|
| Regular videos | ✓ | Current stats only |
| TikTok Stories | ✗ | Not available |
| LIVE videos | ✗ | Not available |
| Duets | Partial | As regular video |
| Stitches | Partial | As regular video |
| Drafts | ✗ | Not available |

### Video Metadata Limitations
- Duration in seconds (integer)
- Dimensions available
- No audio/music information
- No hashtag performance data
- No effect/filter information

## Authentication Limitations

### OAuth2 Token Lifecycle
| Token Type | Lifespan | Notes |
|------------|----------|-------|
| Access token | 24 hours | Must refresh frequently |
| Refresh token | Variable | Check expires_in |

### Token Refresh Issues
| Issue | Impact |
|-------|--------|
| Frequent expiration | Must implement refresh logic |
| Refresh token expiry | User must re-authenticate |
| Scope changes | May invalidate tokens |

### Required Scopes
```
user.info.basic     # Basic profile information
video.list          # Access to video list
```

### Scope Limitations
- Limited scopes available for analytics
- Some scopes require partnership
- No scope for detailed analytics

## Pagination Limitations

### Cursor-Based Pagination
```go
type VideoListResponse struct {
    Data struct {
        Videos  []Video `json:"videos"`
        Cursor  int64   `json:"cursor"`   // For next page
        HasMore bool    `json:"has_more"` // More pages exist
    } `json:"data"`
}
```

### Pagination Issues
| Issue | Impact |
|-------|--------|
| Max 20 per page | Many requests for active accounts |
| No date filtering | Must fetch all, filter locally |
| Cursor expiration | May expire between requests |
| No total count | Unknown total videos upfront |

## Insights Generation Workaround

Since TikTok doesn't provide account insights API, we generate them from video data:

### Generated Insights
```go
type GeneratedInsights struct {
    TikTokID       string // Account ID
    FollowerCount  int64  // From user info
    FollowingCount int64  // From user info
    LikeCount      int64  // From user info (total likes received)
    VideoCount     int64  // From user info

    // Aggregated from videos
    TotalViews     int64  // Sum of all video views
    TotalLikes     int64  // Sum of all video likes
    TotalComments  int64  // Sum of all video comments
    TotalShares    int64  // Sum of all video shares
}
```

### Limitations of Generated Insights
- Only reflects fetched videos, not all videos
- No historical comparison
- View counts may be stale
- Incremental sync may miss deleted videos

## Error Handling

### Common Error Codes
| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 6 | Rate limited | Backoff and retry |
| 10001 | Invalid parameters | Check request |
| 10002 | Invalid access token | Refresh token |
| 10003 | Scope error | Check permissions |

### TikTok-Specific Errors
| Error | Cause | Solution |
|-------|-------|----------|
| "access_token_invalid" | Expired token | Refresh token |
| "rate_limit_exceeded" | Too many requests | Wait and retry |
| "scope_not_authorized" | Missing scope | Re-authenticate |

## Comparison with Other Platforms

### Feature Comparison
| Feature | TikTok | Facebook | Instagram | YouTube |
|---------|--------|----------|-----------|---------|
| Video analytics | ✗ | ✓ | ✓ | ✓ |
| Historical data | ✗ | ✓ | ✓ | ✓ |
| Audience demographics | ✗ | ✓ | ✓ | ✓ |
| Engagement trends | ✗ | ✓ | ✓ | ✓ |
| Date filtering | ✗ | ✓ | ✓ | ✓ |
| Comment data | ✗ | ✓ | ✓ | ✓ |

### Impact on Analytics
- Cannot provide trend analysis
- Cannot calculate growth metrics
- Cannot compare time periods
- Limited competitive analysis

## Known Issues & Quirks

### API Behavior Issues
| Issue | Description |
|-------|-------------|
| Inconsistent counts | View counts may differ from app |
| Delayed updates | Stats may lag by hours |
| Missing videos | Recently deleted videos still in list |
| Cursor instability | May skip or duplicate videos |

### Data Quality Issues
- Engagement counts are point-in-time snapshots
- No way to verify data accuracy
- Private videos may appear in list but fail to fetch details

## Recommendations

### Best Practices
1. **Implement token refresh** - Tokens expire frequently
2. **Use cutoff time** - Filter locally by creation time
3. **Cache video data** - Reduce API calls
4. **Handle missing analytics** - Generate from available data
5. **Track video IDs** - Detect deleted videos

### Workarounds
| Limitation | Workaround |
|------------|------------|
| No date filtering | Store last sync, filter locally |
| No analytics | Track snapshots over time in DB |
| No demographics | Not possible to workaround |
| Token expiration | Aggressive refresh before expiry |

### Monitoring
- Track token refresh success rate
- Monitor API error rates
- Alert on authentication failures
- Log video count changes

## Future Considerations

### Potential API Improvements
TikTok's API is relatively new and may improve:
- Analytics API may be added
- Historical data access may come
- More scopes may be available

### Partnership Options
- TikTok for Business may have more access
- Research API has different capabilities
- Marketing API (if available) may help

## Summary of Key Limitations

1. **No analytics API** - Most critical limitation
2. **No historical data** - Cannot trend over time
3. **No date filtering** - Must fetch all videos
4. **Short token lifespan** - Frequent refreshes needed
5. **Limited video metadata** - No audio/hashtag analytics
6. **No audience data** - Demographics unavailable
7. **No comment access** - Cannot analyze sentiment
