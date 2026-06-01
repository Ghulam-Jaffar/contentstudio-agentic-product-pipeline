# Twitter API Limitations & Constraints

## API and Auth Constraints

| Constraint | Detail |
|-----------|--------|
| Auth model | OAuth 1.0a user context (consumer key/secret + access token/secret) |
| Signature | HMAC-SHA1 |
| Endpoint family | Twitter API v2 |
| Per-request tweet max | 100 (`max_results`) |

## Pagination Constraints

| Constraint | Detail |
|-----------|--------|
| Pagination token | `meta.next_token` must be used for subsequent pages |
| Large post_count behavior | For `post_count > 100`, implementation uses `max_results=50` per page |
| Early stop conditions | No `next_token`, empty page, context cancel, or target count reached |

## Rate Limit and Error Handling

| HTTP Code | Meaning | Current handling |
|----------|---------|------------------|
| `429` | Rate limit | Request returns error and processing stops current fetch loop |
| `401` | Unauthorized token/app | Request returns error and account processing fails/skips depending on flow |
| `!=200` | API error | Error includes response body for diagnostics |

Notes:

- Inter-page delay is applied during paginated loops to reduce burst pressure.
- API limits vary by app/account tier and are not fully discoverable from code-level config.

## Data Availability Constraints

| Area | Constraint |
|------|------------|
| Historical scope | Twitter flow currently uses `post_count` only; no date-window filter in fetch loop |
| Metrics availability | Some public metrics can be missing or delayed for specific tweets/accounts |
| Protected/deleted content | Tweets or users can disappear between requests |

## Configuration Dependencies

Twitter account processing depends on cross-collection config integrity:

1. `social_integrations.developer_app_id` must exist
2. `developer_apps` record must exist with `analytics_enabled=true`
3. `twitter_jobs_setting` must exist by `platform_id=platform_identifier`

If any of these are missing, the account is intentionally skipped.

## Scheduling Constraints

`twitter_jobs_setting.job_type` governs inclusion:

- `daily`
- `weekly` (weekday match with `trigger_day`)
- `monthly` (day-of-month match with `trigger_day`)
- `never` (always excluded)

Incorrect `trigger_day` values can silently prevent scheduling.

## Storage and Schema Constraints

Production-aligned Twitter table constraints include:

- `twitter_insights.verified` as `String`
- `twitter_posts.author_id_created` as `DateTime64(6)`

If payload conversion types diverge from schema, batch inserts fail.

## Operational Constraints

| Area | Constraint |
|------|------------|
| Immediate queueing | In-memory queue in unified immediate processor can drop in-flight work on process crash |
| Delivery semantics | At-most-once behavior when offsets commit before full downstream completion |
| External dependencies | MongoDB, Kafka, Twitter API, ClickHouse must all be healthy for full success path |

