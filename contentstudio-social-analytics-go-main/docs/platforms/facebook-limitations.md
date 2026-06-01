# Facebook API Limitations & Constraints

## API Rate Limits

### Graph API Rate Limiting
| Limit Type | Value | Notes |
|------------|-------|-------|
| App-level | 200 calls/user/hour | Shared across all users |
| Page-level | 4800 calls/page/day | Per page token |
| Burst limit | 50 calls/second | Short burst allowance |

### Our Implementation Limits
| Limit Type | Value |
|------------|-------|
| Per-token | 4 RPS |
| Global | 12 RPS |
| Burst | 4 requests |

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Historical Access | Notes |
|-----------|-------------------|-------|
| Posts | 2 years | Older posts may be inaccessible |
| Videos | 2 years | Same as posts |
| Insights | 2 years | Daily insights only |
| Page-level insights | 93 days max per request | Must paginate for longer periods |

### Deprecated/Removed Metrics
| Metric | Status | Alternative |
|--------|--------|-------------|
| `post_impressions` | Deprecated (2024) | Use `post_media_view` |
| `post_impressions_unique` | Deprecated | Use reach metrics |
| Lifetime insights for old posts | Limited | May return 0 for posts > 2 years |

## Permission Requirements

### Required Permissions for Analytics
```
pages_read_engagement      # Read page engagement data
pages_read_user_content    # Read user-generated content
read_insights              # Read page and post insights
pages_show_list           # List pages user manages
```

### Permission Limitations
- Business verification required for some insights
- App review required for production use
- Some metrics only available to page admins

## Data Accuracy Issues

### Known Discrepancies
1. **Engagement counts** may differ between Graph API and Facebook UI
2. **Reach metrics** can have ~5% variance
3. **Video views** counted differently (3-second vs completion)
4. **Reactions** may have delayed updates (up to 24 hours)

### Timing Issues
| Issue | Impact |
|-------|--------|
| Insights delay | 15-30 minutes for recent data |
| Engagement sync | Up to 24 hours for final counts |
| Video metrics | 48-72 hours for complete data |

## Content Type Limitations

### Unsupported Content Types
| Content | Status | Notes |
|---------|--------|-------|
| Stories | Limited API | Only business accounts, 24-hour lifespan |
| Reels insights | Partial | Limited metrics compared to UI |
| Live videos | During-live only | Post-live becomes regular video |
| Scheduled posts | Not in insights | Until published |

### Video-Specific Limitations
- Video insights require video to be published > 24 hours
- Some video metrics only for videos > 1 minute
- Live video insights different from uploaded videos

## API Response Limitations

### Field Limits
| Limit | Value |
|-------|-------|
| Max fields per request | ~100 fields |
| Max posts per page | 100 |
| Max pages (pagination) | No hard limit, but rate limited |

### Response Size
- Large responses may be truncated
- Pagination required for > 100 items
- Some fields return empty for privacy reasons

## Authentication Limitations

### Token Expiration
| Token Type | Lifespan | Refresh |
|------------|----------|---------|
| Short-lived | 1-2 hours | Must re-authenticate |
| Long-lived | 60 days | Can extend once |
| Page token | Never expires | Tied to user token validity |

### Token Issues
- Long-lived token can only be extended once
- Page token invalidated if user loses page access
- Token may be invalidated by password change

## Competitor Analysis Limitations

### Public Data Only
| Available | Not Available |
|-----------|---------------|
| Public posts | Private posts |
| Reaction counts | Detailed demographics |
| Comment counts | Insights data |
| Share counts | Follower details |

### Rate Limits for Competitor Data
- Same rate limits apply
- No special access for competitor pages
- May be blocked if detected as scraping

## Platform-Specific Quirks

### Timestamp Format
```
// Facebook returns: "2024-01-15T10:30:00+0000"
// Go requires:      "2024-01-15T10:30:00+00:00"
// Must convert timezone format
```

### Reaction Types
| Reaction | API Name | Notes |
|----------|----------|-------|
| Like | LIKE | Standard |
| Love | LOVE | Standard |
| Haha | HAHA | Standard |
| Wow | WOW | Standard |
| Sad | SAD | Standard |
| Angry | ANGRY | Standard |
| Thankful | THANKFUL | Limited availability (deprecated in some regions) |
| Care | CARE | Added 2020, may not be in older data |

### Attachment Structures
- Attachments can be nested 3+ levels deep
- Some attachments have no media_type field
- Link previews may not include images

## Error Handling

### Common Error Codes
| Code | Meaning | Action |
|------|---------|--------|
| 190 | Invalid/expired token | Re-authenticate |
| 4 | Rate limit exceeded | Backoff and retry |
| 10 | Permission denied | Check permissions |
| 100 | Invalid parameter | Check request format |
| 200 | Requires permission | Request additional permissions |

### Transient Errors
- 500 errors: Retry with backoff
- Timeout errors: Retry once
- "Unknown error": Usually rate limiting

## Recommendations

### Best Practices
1. **Cache tokens** - Minimize token refresh calls
2. **Batch requests** - Use batch API when possible
3. **Incremental sync** - Don't fetch full history daily
4. **Handle deprecations** - Monitor API changelog
5. **Use webhooks** - For real-time updates when possible

### Monitoring
- Track rate limit headers in responses
- Log API errors with context
- Alert on token expiration
- Monitor for API deprecation notices
