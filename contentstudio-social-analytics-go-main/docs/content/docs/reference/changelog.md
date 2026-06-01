---
title: Changelog
description: Recent changes and updates to the analytics pipeline
---

# Changelog

## URL Refresher Job Split

- Added platform-specific URL refresher runners for Facebook, Instagram, LinkedIn, Facebook competitors, and Instagram competitors.
- Added `-platform` and `-accountType` flags to the URL refresher entrypoint.
- The refresher now reads stale media rows from ClickHouse and writes refreshed URLs back after API lookups.

## 2024

### YouTube Integration Enhancements

#### YouTube Consent Time Filter
Added validation for YouTube user consent to ensure compliance with YouTube's data access requirements.

**Changes:**
- Added `consentDays` parameter (30 days) to scheduler queries
- Added consent validation for immediate sync API
- Accounts with expired consent (>30 days) are excluded from processing
- Repository methods updated: `GetYouTubeAccountsNeedingUpdatePaginated`, `CountYouTubeAccountsNeedingUpdate`

**Affected Files:**
- `cmd/jobs/fetcher/youtube.go`
- `db/mongodb/repository.go`
- `api/immediate_work_apis.go`

#### Embedded Document Token Support
Added support for YouTube's embedded document `access_token` format while maintaining backward compatibility.

**Changes:**
- MongoDB BSON decoder now handles both string and embedded document token formats
- Automatic format detection for `access_token` field
- Extracts `access_token` and `refresh_token` from nested document structure

**Token Formats Supported:**

String format (legacy):
```json
{
  "access_token": "ya29.xxx...",
  "refresh_token": "1//xxx..."
}
```

Embedded document format (current):
```json
{
  "access_token": {
    "access_token": "ya29.xxx...",
    "refresh_token": "1//xxx...",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

**Affected Files:**
- `models/db/mongo/social_integration.go`

### API Error Response Standardization

#### JSON Error Responses
Standardized API error responses to return structured JSON instead of plain text.

**Changes:**
- All error responses now return JSON with `code` and `message` fields
- Added error codes: `INVALID_REQUEST`, `MISSING_FIELD`, `CONSENT_EXPIRED`, `INTERNAL_ERROR`
- Consistent error format across all API endpoints

**Response Format:**
```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error description"
}
```

**Error Codes:**

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Invalid HTTP method or request format |
| `MISSING_FIELD` | Required field missing from request body |
| `CONSENT_EXPIRED` | YouTube consent has expired (>30 days) |
| `INTERNAL_ERROR` | Server-side processing error |

**Affected Files:**
- `api/immediate_work_apis.go`
- `api/handle_competitor_work.go`

### Test Coverage Updates

Added comprehensive test coverage for all recent changes:

- `api/immediate_work_apis_test.go` - JSON error response tests, YouTube consent validation tests
- `models/db/mongo/social_integration_test.go` - Embedded document token parsing tests
- `api/handle_competitor_work_test.go` - Updated error response expectations
- Mock implementations updated in `db/mongodb/test_db.go`, `cmd/jobs/fetcher/fetcher_test.go`, `services/tiktok/tiktok-fetcher/main_test.go`

---

## Platform Data Fetch Periods Summary

| Platform | Incremental | Immediate | Full Sync | Special Notes |
|----------|-------------|-----------|-----------|---------------|
| **Facebook** | 14 days | - | 90 days | - |
| **Instagram** | 14 days | 30 days | 89 days | - |
| **LinkedIn** | 10 days | - | 365 days | - |
| **TikTok** | 14 days | - | Unlimited | Max 999 videos for incremental |
| **YouTube** | 14 days | 90 days | 365 days | 30-day consent window required |

### Scheduler Configuration (All Platforms)

| Parameter | Value |
|-----------|-------|
| Update Interval | 6 hours |
| Batch Size | 200 accounts |
