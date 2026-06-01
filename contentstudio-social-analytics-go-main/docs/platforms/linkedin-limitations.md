# LinkedIn API Limitations & Constraints

## API Rate Limits

### Marketing API Rate Limits
| Limit Type | Value | Notes |
|------------|-------|-------|
| Daily limit | 100,000 calls/day | Per application |
| Per-member | 500 calls/day | Per authenticated user |
| Burst limit | 100 calls/minute | Short-term limit |

### Our Implementation
| Configuration | Value |
|---------------|-------|
| HTTP Timeout | 30 seconds |
| Pagination delay | 100ms between pages |
| Batch size | 200 accounts max |

## Account Type Limitations

### Organization vs Member Profiles
| Feature | Organization Page | Member Profile |
|---------|-------------------|----------------|
| Post fetching | Full access | Limited |
| Follower analytics | Full | Not available |
| Post statistics | Full | Limited |
| Demographic data | Full | Not available |
| Historical data | 12 months | Limited |

### Minimum Requirements
- Organization must be verified for full analytics
- Admin access required for organization insights
- Member profile requires r_organization_social permission

## Data Access Limitations

### Time-Based Restrictions
| Data Type | Historical Access | Notes |
|-----------|-------------------|-------|
| Posts | 12 months | Older posts may be inaccessible |
| Follower data | 12 months | Daily snapshots |
| Post statistics | 12 months | Engagement metrics |
| Page statistics | 12 months | Views and visitors |
| Demographics | Current snapshot | No historical |

### Data Availability Delays
| Data Type | Delay |
|-----------|-------|
| Post engagement | 15-30 minutes |
| Follower counts | Up to 24 hours |
| Post impressions | 24-48 hours |
| Demographic updates | Weekly |

## API Versioning Limitations

### API Version Requirements
```
LinkedIn-Version: 202509  // Required header
```

### Deprecated Endpoints
| Old Endpoint | New Endpoint | Status |
|--------------|--------------|--------|
| `/v2/shares` | `/rest/posts` | v2 deprecated 2024 |
| `/v2/ugcPosts` | `/rest/posts` | v2 deprecated 2024 |
| `organizationalEntityFollowerStatistics` | New endpoints | Changed |

### Breaking Changes
- URN format changes between versions
- Field name changes in responses
- Authentication scope changes

## Content Type Limitations

### Supported Content Types
| Content Type | Fetch | Statistics | Notes |
|--------------|-------|------------|-------|
| Text posts | ✓ | ✓ | Full support |
| Image posts | ✓ | ✓ | Single/multi-image |
| Video posts | ✓ | ✓ | Native video only |
| Articles | ✓ | ✓ | Link previews |
| Documents | ✓ | ✓ | PDFs, slides |
| Polls | ✓ | Partial | Vote counts only |
| Events | Limited | Limited | Basic info only |
| Newsletters | Not supported | Not supported | No API access |

### Video Limitations
- Only native LinkedIn videos (not YouTube embeds)
- Video insights limited to views, no watch time
- Large video metadata may be incomplete

### Poll Limitations
- Vote counts available
- Individual voter data not available
- Poll options text accessible
- No demographic breakdown of votes

## Insights Limitations

### Follower Analytics
| Metric | Availability | Granularity |
|--------|--------------|-------------|
| Total followers | Available | Daily |
| Follower growth | Available | Daily |
| Organic vs Paid | Available | Daily |
| Demographics | Available | Snapshot only |

### Demographic Breakdowns Available
| Breakdown | Max Values | Notes |
|-----------|------------|-------|
| Seniority | All levels | Entry to CXO |
| Industry | Top values | Not comprehensive |
| Company size | All ranges | 1-10 to 10001+ |
| Function | Top values | Job functions |
| Country | Top values | Geo-resolved |

### Demographic Limitations
- Only aggregated data, no individual followers
- Demographics may not sum to 100%
- "Unknown" category for unspecified data
- Updated weekly, not real-time

## Geo ID Resolution

### Geo URN Format
```
// Raw: "urn:li:geo:103644278"
// Must resolve to: "United States"
```

### Resolution Limitations
| Issue | Impact |
|-------|--------|
| API calls required | One call per unique geo ID |
| Rate limiting | Subject to same limits |
| Missing mappings | Some geo IDs may not resolve |
| Caching required | To avoid repeated lookups |

## Authentication Limitations

### Token Expiration
| Token Type | Lifespan | Refresh |
|------------|----------|---------|
| Access token | 60 days | Via refresh token |
| Refresh token | 365 days | Must re-authenticate after |

### Required Scopes
```
r_organization_social     # Read organization posts
r_organization_admin      # Read org admin data
w_organization_social     # Write posts (not needed for analytics)
rw_organization_admin     # Full admin access
r_liteprofile            # Basic profile info
r_member_social          # Member social data
```

### Scope Limitations
- Some scopes require LinkedIn partnership
- Organization scopes require admin access
- Member scopes limited for privacy

## API Response Limitations

### Pagination
| Limit | Value |
|-------|-------|
| Max items per page | 100 |
| Max pages | No hard limit |
| Cursor expiration | ~24 hours |

### Response Size Issues
- Large organizations may timeout
- Pagination required for > 100 items
- Some fields truncated for length

## URN Format Handling

### URN Types
```go
"urn:li:organization:12345678"     // Organization
"urn:li:person:abc123"             // Member
"urn:li:share:7123456789"          // Share (old format)
"urn:li:ugcPost:7123456789"        // UGC Post
"urn:li:activity:7123456789"       // Activity
"urn:li:image:C5..."               // Image asset
"urn:li:video:C5..."               // Video asset
"urn:li:document:C5..."            // Document asset
```

### URN Parsing Issues
- Format may vary between endpoints
- Must extract ID from URN for some calls
- Version differences in URN formats

## Error Handling

### Common Error Codes
| Code | Meaning | Action |
|------|---------|--------|
| 401 | Unauthorized | Re-authenticate |
| 403 | Forbidden | Check permissions |
| 404 | Not found | Resource deleted/private |
| 429 | Rate limited | Backoff and retry |
| 500 | Server error | Retry with backoff |

### LinkedIn-Specific Errors
| Error | Cause | Solution |
|-------|-------|----------|
| "RESOURCE_NOT_FOUND" | Deleted content | Skip resource |
| "ACCESS_DENIED" | Lost admin access | Re-verify permissions |
| "INVALID_ARGUMENT" | URN format issue | Check URN format |

## Competitor Analysis Limitations

### Public Data Only
| Available | Not Available |
|-----------|---------------|
| Public posts | Private posts |
| Reaction counts | Detailed statistics |
| Comment counts | Follower analytics |
| Company info | Page insights |
| Post content | Demographic data |

### Limitations
- No API for competitor insights
- Public posts only via search
- Rate limited same as own data
- No historical competitor data

## Known Issues & Quirks

### Documented Issues
1. **Engagement discrepancies**: API vs UI can differ by 5-15%
2. **Delayed statistics**: Up to 48 hours for final counts
3. **Missing reactions**: Some reaction types not returned
4. **Timezone issues**: All times in UTC, no timezone info

### Data Consistency Issues
| Issue | Impact |
|-------|--------|
| Follower count lag | May be hours behind |
| Post statistics delay | 24-48 hours for accuracy |
| Demographic accuracy | ~95% due to "unknown" |

## Recommendations

### Best Practices
1. **Use batch work orders** - Process multiple accounts efficiently
2. **Cache geo resolutions** - Avoid repeated lookups
3. **Handle URN formats** - Parse and validate URNs
4. **Implement retry logic** - LinkedIn has transient errors
5. **Monitor deprecations** - API changes frequently

### Optimization
- Batch API calls where possible
- Use pagination efficiently
- Cache frequently accessed data
- Implement exponential backoff

### Monitoring
- Track rate limit headers
- Log API version mismatches
- Alert on authentication failures
- Monitor data freshness
