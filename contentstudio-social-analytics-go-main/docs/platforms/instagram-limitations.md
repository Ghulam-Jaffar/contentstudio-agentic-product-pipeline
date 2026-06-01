# Instagram API Limitations & Constraints

## API Rate Limits

### Graph API Rate Limiting (via Facebook)
| Limit Type | Value | Notes |
|------------|-------|-------|
| App-level | 200 calls/user/hour | Shared with Facebook |
| Per-user | 200 calls/hour | Per Instagram account |
| Burst limit | 50 calls/second | Short burst allowance |

### Our Implementation Limits
| Limit Type | Value |
|------------|-------|
| Per-token | 4 RPS |
| Global | 12 RPS |
| Burst | 4 requests |

## Account Type Limitations

### Business vs Creator vs Personal
| Feature | Business | Creator | Personal |
|---------|----------|---------|----------|
| Media fetching | ✓ | ✓ | ✗ |
| Insights | ✓ | Limited | ✗ |
| Demographics | ✓ | Limited | ✗ |
| Story insights | ✓ | ✓ | ✗ |
| Hashtag search | ✓ | ✗ | ✗ |

### Minimum Requirements
- Business/Creator account required for analytics
- Must be connected to Facebook Page (Business)
- Account must have 100+ followers for some insights

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Historical Access | Notes |
|-----------|-------------------|-------|
| Media | 2 years | Older media may be inaccessible |
| Insights (daily) | 30 days | Rolling window |
| Insights (lifetime) | All time | But may be incomplete for old posts |
| Stories | 24 hours | Stories auto-delete |
| Story insights | 14 days | After story expires |

### Insights Availability
| Metric Type | Availability |
|-------------|--------------|
| Impressions | 2 years of data |
| Reach | 2 years of data |
| Engagement | 2 years of data |
| Demographics | Current snapshot only |
| Follower count history | Not available via API |

## Content Type Limitations

### Media Type Support
| Content Type | API Support | Insights Available |
|--------------|-------------|-------------------|
| Image | Full | Full |
| Video | Full | Full |
| Carousel | Full | Aggregate only |
| Reel | Full | Limited metrics |
| Story | Limited (24h) | Limited (14 days) |
| IGTV | Deprecated | Legacy support only |
| Live | During only | No post-live insights |
| Guide | Not supported | Not available |

### Carousel Limitations
- Individual carousel item insights not available
- Only aggregate metrics for entire carousel
- Can't determine which slide performed best
- Child media URLs may expire

### Reels Limitations
- Some metrics not available (saves, shares vary)
- `plays` metric may differ from UI
- Watch time metrics inconsistent
- Audio attribution not available via API

### Story Limitations
- Only accessible for 24 hours after posting
- Insights available for 14 days after expiry
- No historical story data
- Mentions/tags not fully exposed
- Poll/quiz results partially available

## Insights Limitations

### Demographic Breakdowns
| Breakdown | Availability | Limitation |
|-----------|--------------|------------|
| Age | Available | Grouped ranges only (18-24, 25-34, etc.) |
| Gender | Available | M/F/U only |
| City | Top 45 only | Not comprehensive |
| Country | Top 45 only | Not comprehensive |

### Metric Discrepancies
| Issue | Description |
|-------|-------------|
| Reach vs Impressions | Can show reach > impressions (known bug) |
| Engagement counts | May differ from UI by 5-10% |
| Follower count | Cached, may be hours behind |
| Video views | Counted at 3 seconds, differs from UI |

## API Response Limitations

### Pagination
| Limit | Value |
|-------|-------|
| Max media per request | 100 |
| Max insights metrics | 30 per request |
| Cursor expiration | ~24 hours |

### Field Availability
- Some fields return null for old media
- Insights may return empty for low-engagement posts
- Caption may be truncated at 2200 characters

## Authentication Limitations

### Token Requirements
| Token Type | Lifespan | Notes |
|------------|----------|-------|
| Short-lived | 1 hour | From OAuth flow |
| Long-lived | 60 days | Exchanged from short-lived |
| Page token | Never expires | But tied to user access |

### Permission Scopes
```
instagram_basic           # Basic profile info
instagram_content_publish # Publish content (not needed for analytics)
instagram_manage_insights # Read insights data
pages_read_engagement     # Required for some Instagram features
pages_show_list          # List connected pages
```

### Connected via Facebook Limitations
- Instagram must be connected to Facebook Page
- Token may be Facebook Page token, not Instagram token
- Disconnecting Facebook breaks Instagram access

## Competitor Analysis Limitations

### Public Data Access
| Available | Not Available |
|-----------|---------------|
| Public media | Private accounts |
| Like counts | Detailed insights |
| Comment counts | Follower demographics |
| Follower count (public) | Engagement rates |
| Bio information | Story data |

### Hashtag Limitations
- Hashtag search limited to Business accounts
- Max 30 unique hashtags per 7-day rolling period
- Results limited to top/recent media
- No historical hashtag data

## Platform-Specific Quirks

### Media URL Expiration
```
// Instagram CDN URLs expire after ~24-48 hours
// Must refresh URLs or cache images locally
"media_url": "https://scontent.cdninstagram.com/..." // Expires!
```

### Insights Request Format
```go
// Must specify metric names explicitly
// Different metrics for different media types
imageMetrics := "impressions,reach,likes,comments,saved,shares"
videoMetrics := "impressions,reach,likes,comments,saved,shares,video_views"
reelMetrics  := "impressions,reach,likes,comments,saved,shares,plays"
```

### Engagement Calculation Issues
- `engagement` metric deprecated, must calculate manually
- `saved` count not always available
- `shares` only for Reels

## Error Handling

### Common Error Codes
| Code | Meaning | Action |
|------|---------|--------|
| 190 | Invalid token | Re-authenticate |
| 4 | Rate limit | Backoff and retry |
| 10 | Permission denied | Check scopes |
| 100 | Invalid parameter | Check media type/metrics |
| 24 | Too many requests | Wait and retry |

### Instagram-Specific Errors
| Error | Cause | Solution |
|-------|-------|----------|
| "Media not found" | Deleted or private | Skip media |
| "Insights not available" | < 100 followers | Cannot retrieve |
| "Invalid metric" | Wrong media type | Check media type first |

## Known Bugs & Issues

### Documented Issues
1. **Reach > Impressions**: Can occur, acknowledged by Meta
2. **Zero insights**: Posts < 24 hours may show 0
3. **Delayed updates**: Insights may lag 2-24 hours
4. **Carousel inconsistency**: Child count may differ from UI

### Workarounds
- Cache data and retry for zero insights
- Use longer time windows for accuracy
- Cross-validate with UI periodically
- Handle null/missing fields gracefully

## Recommendations

### Best Practices
1. **Check media type** before requesting insights
2. **Cache media URLs** - they expire quickly
3. **Handle missing data** - not all media has insights
4. **Use incremental sync** - avoid re-fetching all media
5. **Validate account type** - personal accounts fail silently

### Monitoring
- Track API error rates by type
- Monitor for token expiration
- Alert on sudden data drops
- Log media types for debugging
