# Technical Solution Document – Plan Validation via Denormalized Fields in Social Integrations

**Document Version:** 1.0
**Date:** 2026-02-08
**Author:** Engineering Team
**Status:** Draft

---

## 1. Purpose

The purpose of this Technical Solution Document (TSD) is to provide a complete and detailed technical implementation plan for adding plan/subscription validation to the Social Analytics Pipeline by denormalizing plan fields directly into the `social_integrations` collection. This approach eliminates the need for joins or lookups at query time, providing optimal read performance.

---

## 2. Problem Statement

The Social Analytics Pipeline fetches social accounts from MongoDB and processes them through platform-specific fetchers (Instagram, Facebook, LinkedIn, TikTok). Currently, there is no validation to check if the associated user/workspace has an active subscription plan, resulting in wasted resources processing accounts that should not be included.

**Key problems include:**

| Problem | Impact |
|---------|--------|
| No plan validation in pipeline | Processing 13K+ accounts including expired subscriptions |
| Plan data in separate `planners` collection | Requires expensive joins at query time |
| ~900K API calls per full sync | Wasted API quota on inactive accounts |
| Processing time ~7+ hours | Could be reduced by filtering inactive plans |
| PHP manages plans, Go fetches accounts | Data lives in different systems |

---

## 3. Solution Overview

**Selected Approach: Denormalization (Option 3)**

Add plan-related fields directly to the `social_integrations` collection. When PHP updates a user's plan, it simultaneously updates all associated social integrations. The Go pipeline filters accounts using these denormalized fields without any joins.

### 3.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Plan Update Flow                                    │
│                                                                                  │
│   ┌──────────────┐         ┌──────────────┐         ┌────────────────────────┐  │
│   │              │  HTTP   │              │ Update  │       MongoDB          │  │
│   │    User      │────────▶│  PHP Backend │────────▶│                        │  │
│   │              │         │              │         │  ┌──────────────────┐  │  │
│   └──────────────┘         └──────────────┘         │  │    planners      │  │  │
│                                   │                  │  └──────────────────┘  │  │
│                                   │                  │           +            │  │
│                                   │ Sync Plan Fields │  ┌──────────────────┐  │  │
│                                   └─────────────────▶│  │ social_          │  │  │
│                                                      │  │ integrations     │  │  │
│                                                      │  │ + plan_active    │  │  │
│                                                      │  │ + plan_expires_at│  │  │
│                                                      │  │ + plan_tier      │  │  │
│                                                      │  └──────────────────┘  │  │
│                                                      └────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                           Analytics Pipeline Flow                                │
│                                                                                  │
│  ┌────────────────┐    ┌─────────────────┐    ┌─────────────┐    ┌───────────┐  │
│  │    MongoDB     │    │ Unified Account │    │   Kafka     │    │ Platform  │  │
│  │    Query:      │───▶│    Fetcher      │───▶│ Work Orders │───▶│ Fetchers  │  │
│  │ plan_active:   │    │                 │    │             │    │           │  │
│  │   true         │    └─────────────────┘    └─────────────┘    └───────────┘  │
│  │ plan_expires_at│                                                      │      │
│  │   > now()      │                                                      ▼      │
│  └────────────────┘                                              ┌───────────┐  │
│                                                                  │ClickHouse │  │
│         NO JOINS REQUIRED - Direct indexed query                 │   Sink    │  │
│                                                                  └───────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Data Flow Diagram

```
                                    PLAN CHANGE EVENT
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                  PHP BACKEND                                     │
│                                                                                  │
│  1. User upgrades/downgrades/cancels plan                                       │
│  2. Update planners collection                                                   │
│  3. Find all social_integrations for user                                       │
│  4. Bulk update plan fields in social_integrations                              │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                   MONGODB                                        │
│                                                                                  │
│  social_integrations: {                                                         │
│      platform_type: "instagram",                                                │
│      platform_identifier: "17841454628",                                        │
│      plan_active: true,              ◄─── Denormalized                          │
│      plan_expires_at: ISODate(...),  ◄─── Denormalized                          │
│      plan_tier: "professional",      ◄─── Denormalized                          │
│      ...                                                                        │
│  }                                                                              │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            GO PIPELINE (Account Fetcher)                         │
│                                                                                  │
│  filter := bson.M{                                                              │
│      "platform_type": "instagram",                                              │
│      "validity": "valid",                                                       │
│      "plan_active": true,            ◄─── Simple indexed query                  │
│      "plan_expires_at": {"$gt": now},◄─── No joins needed                       │
│  }                                                                              │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Architecture & Components

### 4.1 Component Breakdown

| Component | Responsibility | Technology | Changes Required |
|-----------|---------------|------------|------------------|
| PHP Backend | Update plan fields on subscription changes | Laravel/PHP | New service method |
| MongoDB social_integrations | Store denormalized plan fields | MongoDB | Schema update |
| Unified Account Fetcher | Query with plan filter | Golang | Filter update |
| Reconciliation Job | Periodic sync for consistency | Golang | New job |
| Monitoring | Track plan sync health | Prometheus/Grafana | New metrics |

### 4.2 Schema Design

**social_integrations Collection - Updated Schema:**

```javascript
{
    // Existing fields
    _id: ObjectId("64a1b2c3d4e5f6789012345"),
    platform_type: "instagram",
    platform_identifier: "17841454628238556",
    user_id: ObjectId("62a96faa5f0693c6e3078aa3"),
    workspace_id: ObjectId("62a96faa5f0693c6e3078aa4"),
    validity: "valid",
    state: "added",
    access_token: "encrypted_token_here",
    refresh_token: "encrypted_refresh_token",
    last_analytics_updated_at: ISODate("2026-02-07T09:00:00Z"),
    created_at: ISODate("2025-01-15T10:30:00Z"),
    updated_at: ISODate("2026-02-07T09:00:00Z"),

    // NEW: Denormalized Plan Fields
    plan_active: true,                                // Is the plan currently active?
    plan_expires_at: ISODate("2026-12-31T23:59:59Z"), // When does the plan expire?
    plan_tier: "professional",                        // Plan tier: starter/professional/agency
    plan_synced_at: ISODate("2026-02-07T08:00:00Z")   // Last sync timestamp
}
```

### 4.3 Index Strategy

```javascript
// Primary query index for account fetcher
db.social_integrations.createIndex(
    {
        "platform_type": 1,
        "validity": 1,
        "state": 1,
        "plan_active": 1,
        "plan_expires_at": 1
    },
    { name: "idx_platform_plan_filter" }
)

// Index for finding accounts by user (for PHP bulk updates)
db.social_integrations.createIndex(
    {
        "user_id": 1,
        "platform_type": 1
    },
    { name: "idx_user_platform" }
)

// Index for reconciliation job
db.social_integrations.createIndex(
    {
        "plan_synced_at": 1
    },
    { name: "idx_plan_sync_time" }
)
```

---

## 5. Integration Details

| System | Direction | Protocol | Purpose | Auth |
|--------|-----------|----------|---------|------|
| MongoDB (social_integrations) | Read/Write | MongoDB Driver | Account storage with plan fields | SCRAM-SHA-256 |
| MongoDB (planners) | Read | MongoDB Driver | Source of truth for reconciliation | SCRAM-SHA-256 |
| PHP Backend | Write | MongoDB Driver | Updates plan fields on subscription changes | SCRAM-SHA-256 |
| Kafka | Outbound | Kafka Protocol | Work order production | SASL (optional) |
| Logging | Outbound | STDOUT/Sentry | Structured logs and error tracking | API Key |

---

## 6. Performance & Scalability

### 6.1 Performance Comparison

| Metric | Before (No Plan Filter) | After (Denormalized) | Improvement |
|--------|------------------------|----------------------|-------------|
| Accounts queried | 13,000+ | ~8,000 (active only) | 38% reduction |
| Query execution time | ~800ms (with $lookup) | ~50ms (indexed) | 94% faster |
| API calls per sync | ~900,000 | ~550,000 | 39% reduction |
| Pipeline duration | ~7 hours | ~4 hours | 43% faster |

### 6.2 Write Performance

| Operation | Frequency | Documents Affected | Duration |
|-----------|-----------|-------------------|----------|
| Plan activation | ~100/day | ~5 accounts/user | <50ms |
| Plan expiration | ~50/day | ~5 accounts/user | <50ms |
| Bulk update (tier change) | ~20/day | ~5 accounts/user | <50ms |
| Reconciliation job | 1/day | ~1000 stale accounts | ~5 minutes |

---

## 7. Migration Plan

### Phase 1: Schema Preparation (Day 1)

```javascript
// Step 1: Add new fields with default values
db.social_integrations.updateMany(
    { plan_active: { $exists: false } },
    {
        $set: {
            plan_active: true,           // Default active for existing
            plan_expires_at: null,       // Will be populated
            plan_tier: null,             // Will be populated
            plan_synced_at: null         // Will be populated
        }
    }
)

// Step 2: Create indexes
db.social_integrations.createIndex(
    {
        "platform_type": 1,
        "validity": 1,
        "state": 1,
        "plan_active": 1,
        "plan_expires_at": 1
    },
    { name: "idx_platform_plan_filter", background: true }
)

db.social_integrations.createIndex(
    { "user_id": 1, "platform_type": 1 },
    { name: "idx_user_platform", background: true }
)
```

### Phase 2: Initial Data Population (Day 1-2)

```javascript
// Populate plan data from planners collection using aggregation
db.planners.aggregate([
    {
        $lookup: {
            from: "social_integrations",
            localField: "user_id",
            foreignField: "user_id",
            as: "accounts"
        }
    },
    { $unwind: "$accounts" },
    {
        $project: {
            account_id: "$accounts._id",
            plan_active: { $eq: ["$status", "active"] },
            plan_expires_at: "$expires_at",
            plan_tier: "$plan_tier"
        }
    }
]).forEach(function(doc) {
    db.social_integrations.updateOne(
        { _id: doc.account_id },
        {
            $set: {
                plan_active: doc.plan_active && doc.plan_expires_at > new Date(),
                plan_expires_at: doc.plan_expires_at,
                plan_tier: doc.plan_tier,
                plan_synced_at: new Date()
            }
        }
    )
})
```

### Phase 3: PHP Deployment (Day 2-3)

1. Deploy PHP changes with feature flag disabled
2. Enable feature flag for 10% of plan updates
3. Monitor for errors and data consistency
4. Gradually increase to 100%

### Phase 4: Go Pipeline Update (Day 3-4)

1. Deploy Go changes with feature flag
2. Run account fetcher with logging only (don't filter)
3. Verify filter would reduce count correctly
4. Enable filter for production

### Phase 5: Enable Reconciliation Job (Day 5)

1. Schedule reconciliation job to run daily at off-peak hours
2. Monitor job execution and error rates
3. Tune batch size if needed

---

## 8. Monitoring & Alerting

### 8.1 Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `plan_sync_total` | Total plan sync operations | Error rate > 5% |
| `accounts_filtered_by_plan` | Accounts excluded due to inactive plans | Sudden drop > 20% |
| `plan_reconciliation_duration` | Duration of reconciliation job | > 30 minutes |
| `stale_plan_accounts` | Accounts with plan_synced_at > 48h | > 1000 accounts |

### 8.2 Alerts

| Alert | Condition | Severity | Action |
|-------|-----------|----------|--------|
| High plan sync failures | Error rate > 5% | Warning | Check PHP logs |
| Reconciliation job failed | Job exit code != 0 | Critical | Investigate immediately |
| Stale plan data | Accounts with plan_synced_at > 48h | Warning | Run reconciliation |
| Unexpected account drop | Active accounts < 50% of yesterday | Critical | Verify plan data |

### 8.3 Logging

```go
// Account fetcher logging
log.Info().
    Int64("total_accounts", totalCount).
    Int64("active_plans", activePlanCount).
    Int64("expired_plans", expiredPlanCount).
    Msg("Account fetch summary with plan filtering")
```

---

## 9. Assumptions & Dependencies

| Assumption | Impact if Invalid | Mitigation |
|------------|-------------------|------------|
| PHP team can modify plan update flow | Cannot sync plan fields | Use reconciliation job as primary method |
| Plan changes are infrequent (<1000/day) | High write load on social_integrations | Implement batch updates with queue |
| MongoDB available for writes | Plan sync fails | Retry logic with exponential backoff |
| Planners collection is source of truth | Inconsistent data | Daily reconciliation job |
| Index creation won't impact production | Slow queries during build | Use background index creation |

---

## 10. Rollback Plan

| Scenario | Action | Recovery Time |
|----------|--------|---------------|
| PHP sync causing errors | Disable feature flag, revert PHP code | 5 minutes |
| Go filter too aggressive | Remove plan_active filter from query | 5 minutes |
| Data corruption | Restore from backup, run full reconciliation | 1-2 hours |
| Performance degradation | Drop new index, revert to old query | 10 minutes |

**Rollback Command (Go):**

```go
// Remove plan filter - process all valid accounts
filter := bson.M{
    "platform_type": platformType,
    "validity":      mongo3.ValidityValid,
    "state":         bson.M{"$in": validStates},
    // Plan filter removed for rollback
}
```

---

## 11. Security Considerations

| Concern | Mitigation |
|---------|------------|
| Plan data exposure | Plan fields contain no PII, only status/tier |
| Unauthorized plan modification | Only PHP backend with proper auth can update |
| Data tampering | Reconciliation job detects and fixes inconsistencies |
| Audit trail | plan_synced_at tracks last update time |

---

## 12. Testing Strategy

### 12.1 Unit Tests

```go
func TestBuildNeedingUpdateFilter_WithPlanFilter(t *testing.T) {
    tests := []struct {
        name     string
        platform string
        expected bson.M
    }{
        {
            name:     "Instagram with plan filter",
            platform: "instagram",
            expected: bson.M{
                "platform_type": "instagram",
                "validity":      "valid",
                "plan_active":   true,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            filter := buildNeedingUpdateFilter(tt.platform, nil, 6)
            assert.Equal(t, tt.expected["plan_active"], filter["plan_active"])
        })
    }
}
```

### 12.2 Integration Tests

```go
func TestAccountFetcher_FiltersByPlan(t *testing.T) {
    // Setup: Create accounts with different plan states
    activeAccount := createTestAccount(t, "instagram", true, future)
    expiredAccount := createTestAccount(t, "instagram", true, past)
    inactiveAccount := createTestAccount(t, "instagram", false, future)

    // Execute: Run account fetcher
    accounts, err := fetcher.GetAccountsNeedingUpdate(ctx, "instagram")

    // Verify: Only active plan account returned
    assert.NoError(t, err)
    assert.Len(t, accounts, 1)
    assert.Equal(t, activeAccount.ID, accounts[0].ID)
}
```

---

## 13. Future Enhancements

| Enhancement | Priority | Effort | Description |
|-------------|----------|--------|-------------|
| Plan-based rate limiting | Medium | Medium | Higher tier = more API calls allowed |
| Plan-based feature flags | Low | Low | Enable features per plan tier |
| Real-time sync via Change Streams | Low | High | MongoDB change streams for instant sync |
| Plan usage quotas | Medium | High | Track and enforce API call limits per plan |
| Analytics by plan tier | Low | Medium | ClickHouse reports segmented by plan |

---

## 14. Appendix

### A. Field Definitions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| plan_active | Boolean | Yes | Whether the user's plan is currently active |
| plan_expires_at | DateTime | No | When the plan expires (null = never) |
| plan_tier | String | No | Plan tier: "starter", "professional", "agency" |
| plan_synced_at | DateTime | Yes | Last time plan data was synced |

### B. Plan Tier Mapping

| Tier | API Calls/Day | Accounts Limit | Features |
|------|---------------|----------------|----------|
| starter | 1,000 | 5 | Basic analytics |
| professional | 10,000 | 25 | Full analytics |
| agency | 100,000 | Unlimited | Full + white label |

### C. MongoDB Commands Reference

```javascript
// Check accounts without plan data
db.social_integrations.countDocuments({
    plan_active: { $exists: false }
})

// Check accounts with expired plans
db.social_integrations.countDocuments({
    plan_active: true,
    plan_expires_at: { $lt: new Date() }
})

// Force reconciliation for specific user
db.social_integrations.updateMany(
    { user_id: ObjectId("...") },
    { $set: { plan_synced_at: null } }
)
```

---

## 15. Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Technical Lead | __________________ | __________________ | ________ |
| Product Manager | __________________ | __________________ | ________ |
| PHP Lead | __________________ | __________________ | ________ |
| Go Lead | __________________ | __________________ | ________ |
| DevOps Lead | __________________ | __________________ | ________ |

---

**Document History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-08 | Engineering Team | Initial document |
