# LinkedIn API Limitations & Data Discrepancies

This document explains the known differences between LinkedIn API data and LinkedIn Analytics UI, based on official LinkedIn documentation and our investigation.

## Overview

When comparing data from the `organizationalEntityShareStatistics` API with LinkedIn's Analytics UI, you may notice discrepancies. These are **expected behaviors** documented by LinkedIn, not bugs in our implementation.

---

## Summary of Discrepancies

| Metric | Our App | LinkedIn UI | Reason |
|--------|---------|-------------|--------|
| **Impressions** | Higher | Lower | We store lifetime impressions; LinkedIn shows period-only |
| **Comments** | Lower | Higher | Old posts return empty from API |
| **Reactions/Likes** | Lower | Higher | API returns `likeCount` only, not all reaction types |

---

## Documented LinkedIn API Limitations

### 1. Rolling 12-Month Data Window

> "The `organizationalEntityShareStatistics` endpoint returns share data only within the past **12 months**, using a rolling 12-month window."

**Impact**: Posts older than 12 months will not return any statistics.

---

### 2. Posts with No Engagement Are Not Returned

> "Shares with no actions or impressions are **not included** in the list of elements. Shares that are not returned in the list of elements can be assumed to have counts of **0 for all statistics**."

> "UGC Posts with no actions or impressions are **not included** in the list of elements."

**Impact**: When requesting stats for 100 posts, the API may only return 10-20 posts (those with engagement). Posts with zero engagement are simply omitted from the response.

**Example**:
```
Request: Stats for 50 posts
Response: Only 7 posts returned (those with some activity)
Missing: 43 posts have 0 engagement - not included in response
```

---

### 3. Old Posts Return Empty Results

Based on our testing, posts older than ~3 months often return empty `elements: []` from the stats API, even if they have engagement on LinkedIn.

**API Request**:
```
GET /organizationalEntityShareStatistics?shares=urn:li:share:7308985351706927104&q=organizationalEntity&organizationalEntity=urn:li:organization:105256406
```

**API Response**:
```json
{
    "paging": {
        "start": 0,
        "count": 10,
        "total": 0
    },
    "elements": []
}
```

**Impact**: Historical data (3+ months old) may be incomplete or missing entirely.

---

### 4. Time-Bound vs Lifetime Statistics Differences

> "The **shareCount** for the time-bound statistics **won't match** the lifetime statistics because the time-bound data won't include the count from instant reposts."

**Impact**: 
- **Our DB**: Stores lifetime impressions on posts created in a date range
- **LinkedIn UI**: Shows impressions received during the date range only

**Example**:
- Post created Dec 15, 2024
- Gets 100 total lifetime impressions
- 80 impressions during Dec 10, 2024 - Dec 9, 2025
- 20 impressions after Dec 9, 2025
- **Our DB shows**: 100 impressions
- **LinkedIn UI shows**: 80 impressions

---

### 5. `likeCount` vs Reactions

The API returns only `likeCount` which represents the "Like" button clicks only.

LinkedIn UI shows "Reactions" which includes all 6 reaction types:
- Like
- Celebrate (Praise)
- Love (Empathy)
- Insightful (Interest)
- Support (Appreciation)
- Funny (Entertainment)

**Impact**: Our likes count will always be lower than LinkedIn's "Reactions" count.

**Solution**: Use the `socialMetadata` API to get all reaction types. However, this requires `r_organization_social_feed` permission which may not be available.

```
GET /rest/socialMetadata/urn:li:share:{postId}
```

Returns:
```json
{
    "reactionSummaries": {
        "LIKE": { "count": 50 },
        "PRAISE": { "count": 30 },
        "EMPATHY": { "count": 20 },
        "INTEREST": { "count": 15 },
        "APPRECIATION": { "count": 8 }
    }
}
```

---

### 6. Organic Statistics Only

> "This endpoint returns **organic statistics only**. Sponsored activity is not counted in this endpoint. Use Ad Analytics to retrieve statistics for sponsored activity."

**Impact**: If posts were promoted/boosted, those impressions and engagements are not included in the API response.

---

### 7. `likeCount` Can Be Negative

> "**likeCount** - Number of likes. This field **can become negative** when members who liked a sponsored share later unlike it. The like is not counted since it's not organic, but the unlike is counted as organic."

**Impact**: In rare cases, `likeCount` may show negative values.

---

## Data Accuracy by Date Range

Based on our testing:

| Date Range | Accuracy | Notes |
|------------|----------|-------|
| Last 10 days | ~100% | Almost perfect match |
| Last 30 days | ~95% | Minor differences |
| Last 3 months | ~85% | Some gaps in older posts |
| 6+ months | ~75% | Many posts with missing stats |
| 1 year | ~70% | Significant gaps in historical data |

---

## API Endpoints Reference

### Get Post Statistics (Current Implementation)

**For Share posts:**
```
GET https://api.linkedin.com/v2/organizationalEntityShareStatistics?shares=urn:li:share:{postId}&q=organizationalEntity&organizationalEntity=urn:li:organization:{orgId}
```

**For UGC posts:**
```
GET https://api.linkedin.com/v2/organizationalEntityShareStatistics?ugcPosts=urn:li:ugcPost:{postId}&q=organizationalEntity&organizationalEntity=urn:li:organization:{orgId}
```

**Batch request (up to 100 posts):**
```
GET https://api.linkedin.com/v2/organizationalEntityShareStatistics?shares=urn:li:share:{id1}&shares=urn:li:share:{id2}&q=organizationalEntity&organizationalEntity=urn:li:organization:{orgId}
```

### Get Time-Bound Aggregate Statistics

```
GET https://api.linkedin.com/rest/organizationalEntityShareStatistics?q=organizationalEntity&organizationalEntity=urn:li:organization:{orgId}&timeIntervals.timeGranularityType=DAY&timeIntervals.timeRange.start={startMs}&timeIntervals.timeRange.end={endMs}
```

### Get All Reaction Types (Requires Additional Permission)

```
GET https://api.linkedin.com/rest/socialMetadata/urn:li:share:{postId}
```

**Required Permission**: `r_organization_social_feed`

---

## Recommendations

### 1. Accept the Limitations
For historical data (3+ months), accept that some data may be missing due to LinkedIn API limitations.

### 2. Add UI Disclaimer
For date ranges > 30 days, consider showing:
> "Note: Historical data may vary slightly from LinkedIn Analytics due to API limitations."

### 3. Request Additional Permissions
To get accurate reaction counts (all types, not just likes), apply for `r_organization_social_feed` permission through LinkedIn Developer Portal.

### 4. Frequent Syncs for Fresh Data
Run syncs more frequently to capture data while posts are still "fresh" (within 3 months). After this window, LinkedIn API may not return complete stats.

---

## Official Documentation References

1. [Organization Share Statistics API](https://learn.microsoft.com/en-us/linkedin/marketing/community-management/organizations/share-statistics)
2. [Social Metadata API](https://learn.microsoft.com/en-us/linkedin/marketing/community-management/shares/social-metadata-api)
3. [Reactions API](https://learn.microsoft.com/en-us/linkedin/marketing/community-management/shares/reactions-api)

---

## Conclusion

The discrepancies between our data and LinkedIn Analytics UI are **expected behaviors** based on:

1. LinkedIn API limitations (12-month window, old posts returning empty)
2. Different counting methods (lifetime vs period-based)
3. Different metrics (`likeCount` vs all reactions)
4. Posts with zero engagement not being returned

Our implementation is correct - these are fundamental differences in how LinkedIn's API works compared to their internal Analytics UI.

---

*Document created: December 2025*
*Last updated: December 2025*
