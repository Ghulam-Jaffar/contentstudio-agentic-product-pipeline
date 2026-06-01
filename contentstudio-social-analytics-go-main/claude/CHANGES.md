# Changes Log

All significant work done on this codebase, organized by session with files changed.

---

## [2026-05-05] Issues 1, 2, 5 — Kafka Retry, Pagination Caps, At-Least-Once Delivery

### Issue 1 — Kafka Producer Retry
**Problem:** `franzKafkaProducer.Produce` called `ProduceSync` once with no retry. Any transient broker error silently dropped the message.

**Fix:** Added 3-attempt retry loop with `attempt * 500ms` exponential backoff inside `Produce`. All platforms benefit automatically.

**Files:**
- `src/kafka/producer.go`

---

### Issue 2 — Pagination Max-Page Caps + Partial Results on Error
**Problem:** `FetchPostsSince` and `FetchVideosSince` (Facebook) had no max-page limit — could loop forever on an account with huge history. On any mid-pagination error all previously fetched pages were discarded (`return nil, err`). Same max-page gap in LinkedIn `FetchPostsPaginated` and TikTok `FetchVideoPaginated`.

**Fix:**
- Facebook: added `maxPagesToFetch` (20) cap to both `Since` functions; changed all mid-pagination errors from `return nil, err` to `return allPosts/allVideos, err` (partial results preserved)
- LinkedIn: added `maxLinkedInPages = 50` cap; returns partial results with error message when hit
- TikTok: added `maxPages = 200` safety cap (on top of existing `maxVideos` limit)

**Files:**
- `src/clients/social/facebook.go`
- `src/clients/social/linkedin.go`
- `src/clients/social/tiktok.go`

---

### Issue 5 — At-Least-Once Delivery via ConsumeWithAck
**Problem:** Consumer committed Kafka offsets immediately after enqueuing jobs to worker channels. If the service restarted before a worker picked up the job, the offset was already committed → work order permanently lost.

**Fix:** Added `ConsumeWithAck` to the `Consumer` interface. Unlike `Consume` (which calls `CommitUncommittedOffsets` after every batch), `ConsumeWithAck`:
- Uses `DisableAutoCommit()` + `MarkCommitRecords` + `CommitMarkedOffsets`
- Passes an `ack func()` to each handler call
- `ack()` marks that record's offset; a background goroutine commits marked offsets every 5 seconds
- Offsets are never committed until the worker calls `ack()` after full processing
- On restart, uncommitted records are redelivered → true at-least-once semantics

TikTok fetcher wired up as the first adopter:
- `WorkOrderMessage` gained `Ack func()` field
- Consumer switched from `Consume` → `ConsumeWithAck`
- Batch-level ack uses `sync.WaitGroup` so it fires only after ALL accounts in the batch complete
- Both `Processor` and `processorWithTracking` call `msg.Ack()` after `HandleWorkOrder`
- All local mock consumers in test files updated to implement the new interface method

**Files:**
- `src/kafka/consumer.go`
- `src/kafka/mock.go`
- `src/services/tiktok/tiktok-fetcher/main.go`
- `src/services/tiktok/tiktok-analytics-sink/mocks_test.go`
- `src/services/tiktok/tiktok-immediate-processor/run_test.go`
- `src/services/twitter/twitter-analytics-sink/mocks_test.go`
- `src/services/facebook/facebook-analytics-sink/mocks_test.go`
- `src/services/gmb/gmb-analytics-sink/mocks_test.go`

**Note:** ~~Instagram, LinkedIn, Twitter fetchers still use `Consume` (the old method). They should be migrated to `ConsumeWithAck` the same way TikTok was.~~ **Completed** — all three migrated (see session below).

---

## [2026-05-05] Cross-Platform Fetcher Reliability Audit — LinkedIn, TikTok, Twitter

### LinkedIn Fetcher — Silent Semaphore Skip
**Problem:** `pageWorkerLoop` and `profileWorkerLoop` both had `sem.Acquire(...) { continue }` with no log — skipped jobs were completely invisible.

**Fix:** Added `log.Error` with `linkedin_id` before `continue` in both loops.

**Files:**
- `src/services/linkedin/linkedin-fetcher/run.go`

---

### TikTok Fetcher — 4× Ignored json.Marshal Errors
**Problem:** `json.Marshal` called with `_` in four places — account work order dispatch, parsed post, raw post wrapper, insights. A marshal failure silently sent `nil` bytes to Kafka or dropped data.

**Fix:** Each marshal now checks the error and logs + skips (or uses `else if` for the insights publish block). Account work order marshal moved outside the `select` so `continue` is possible.

**Files:**
- `src/services/tiktok/tiktok-fetcher/main.go`

---

### Twitter Fetcher — 4× Ignored json.Marshal Errors
**Problem:** Same pattern as TikTok — `json.Marshal` with `_` for account work order dispatch, parsed tweet, raw post wrapper, insights.

**Fix:** Same approach — error checked, logged, and skipped at each site. Account work order marshal moved outside `select`.

**Files:**
- `src/services/twitter/twitter-fetcher/run.go`

---

## [2026-05-05] Cross-Platform HTTP Retry & Instagram Reliability

### Instagram — FetchMediaInsights Parallelization
**Problem:** `FetchMediaInsights` was called sequentially for every media item, making the fetcher slow for accounts with many posts.

**Fix:** Replaced sequential loops with parallel goroutines bounded by a semaphore (`mediaInsightsConc`). Pre-allocated slices allow index-safe goroutine writes.

**Files:**
- `src/services/instagram/instagram-fetcher/run.go`

---

### Instagram Client — FetchMediaSince Retry
**Problem:** `FetchMediaSince` used `httpClient.Do` directly with no retry.

**Fix:** Switched to `doWithRetry` (already present in the client).

**Files:**
- `src/clients/social/instagram.go`

---

### Facebook Client — 5 Functions Switched to doWithRetry
**Problem:** `FetchVideosWithLimit`, `FetchVideosSince`, `FetchPostsWithLimit`, `FetchPostsSince`, `FetchInsights` all called `httpClient.Do` directly with no retry on transient failures.

**Fix:** Each function now uses `doWithRetry`. Removed explicit `waitRate` calls (it's called internally by `doWithRetry`).

**Files:**
- `src/clients/social/facebook.go`

---

### LinkedIn Client — Retry Loop in makeRequest + FetchShares Refactor
**Problem:** `makeRequest` had no retry; `FetchShares` bypassed `makeRequest` entirely with a direct `HTTPClient.Do`.

**Fix:** Added 3-attempt retry loop to `makeRequest`. Refactored `FetchShares` to use `makeRequest`.

**Files:**
- `src/clients/social/linkedin.go`

---

### TikTok Client — doWithRetry with Factory Pattern
**Problem:** No retry on any TikTok API calls. POST bodies (JSON + form-encoded) would silently fail on the second attempt because `io.Reader` is exhausted.

**Fix:** Added `doWithRetry` helper using factory pattern (`makeReq func() (*http.Request, error)`). All 5 functions updated: `FetchUserVideos`, `FetchUserInfo`, `FetchVideoList`, `FetchVideoDetails`, `RefreshToken`.

**Files:**
- `src/clients/social/tiktok.go`

---

### Twitter Client — doWithRetry with OAuth Re-signing
**Problem:** No retry on Twitter API calls. OAuth 1.0a signatures embed timestamp+nonce, so a request from attempt 1 is invalid for attempt 2+.

**Fix:** Added `doWithRetry` with factory pattern. Factory calls `c.signRequest` on every attempt, producing a fresh valid signature each time. Never retries 401 or 429.

**Files:**
- `src/clients/social/twitter.go`

---

### Instagram Fetcher — Reliability Audit Fixes
**Problem (4 issues found):**
1. Semaphore acquire failure in `mediaWorkerLoop`/`insightsWorkerLoop` silently skipped the job with no log
2. `json.Marshal(payload)` error silently ignored (result discarded with `_`)
3. `Producer.Produce` return value completely ignored — Kafka failures invisible
4. Timestamp channel `default:` case silently dropped the update with no log
5. `resolveAccessToken` returned the raw encrypted ciphertext to the API on decryption failure instead of `""`

**Fix:**
1. Added `log.Error` with `instagram_id` when semaphore acquire fails
2. `json.Marshal` error now checked and logged
3. `Producer.Produce` return value now checked and logged
4. Timestamp channel `default:` now logs `Warn`
5. `resolveAccessToken` returns `""` on decryption failure (triggers proper MongoDB error recording and account skip)

**Files:**
- `src/services/instagram/instagram-fetcher/run.go`
- `src/services/instagram/instagram-fetcher/main.go`

---

## [2026-05-05] ConsumeWithAck — Instagram, LinkedIn, Twitter Fetchers

### Instagram Fetcher — ConsumeWithAck Migration
**Problem:** Consumer committed Kafka offsets immediately after enqueuing jobs. If the service restarted before a worker finished, the offset was already committed → work order lost.

**Fix:**
- Added `Ack func()` to `MediaJob` and `InsightsJob` structs
- Updated `processAccount` signature with `mediaAck func()` and `insightsAck func()` params
- Switched `ConsumeWithAck` in `run.go`; each batch uses `sync.WaitGroup` with `wg.Add(2)` per account (one for media, one for insights)
- Workers call `job.Ack()` after processing (and on semaphore failure) to release WaitGroup
- `processAccount` calls both acks immediately on token failure or context cancel (prevents WaitGroup deadlock)

**Files:**
- `src/services/instagram/instagram-fetcher/run.go`
- `src/services/instagram/instagram-fetcher/main.go`

---

### LinkedIn Fetcher — ConsumeWithAck Migration
**Problem:** Same as Instagram — offsets committed before workers finished.

**Fix:**
- Added `Ack func()` to `WorkOrderMessage` in `types.go`
- Switched `consumePageBatches` and `consumeProfileBatches` to `ConsumeWithAck` with WaitGroup
- `pageWorkerLoop` and `profileWorkerLoop` call `wo.Ack()` after processing and on semaphore failure

**Files:**
- `src/services/linkedin/linkedin-fetcher/types.go`
- `src/services/linkedin/linkedin-fetcher/run.go`

---

### Twitter Fetcher — ConsumeWithAck Migration
**Problem:** Same offset commit issue.

**Fix:**
- Added `Ack func()` to `WorkOrderMessage`
- Switched `processBatches` to `ConsumeWithAck` with WaitGroup pattern (`wg.Add(1)` per account, `Ack: wg.Done`)
- `worker` calls `msg.Ack()` after `handleWorkOrder` returns (success or failure)

**Files:**
- `src/services/twitter/twitter-fetcher/run.go`

---

### MockConsumer — ConsumeWithAck Forwarding
**Problem:** `MockConsumer.ConsumeWithAck` was a no-op, causing LinkedIn fetcher tests to report 0 metrics because the real consumer path (ConsumeWithAck) was never exercised.

**Fix:** `MockConsumer.ConsumeWithAck` now wraps `ConsumeFunc` if set, adapting `AcknowledgingMessageHandler` → `MessageHandler` with a no-op ack. Existing test code unchanged.

**Files:**
- `src/kafka/mock.go`

---

### Test Updates
- `instagram-fetcher/main_test.go`: Updated `processAccount` calls to pass `mongoRepo=nil, mediaAck=func(){}, insightsAck=func(){}`. Updated `TestResolveAccessToken_DecryptionFails_ReturnsOriginal` → `TestResolveAccessToken_DecryptionFails_ReturnsEmpty` to reflect the correct behavior (return `""` on decryption failure, not the raw ciphertext).

---

## [Earlier] Analytics API Server — Go Migration

### LinkedIn Analytics API (9 endpoints)
Migrated all LinkedIn analytics endpoints from Laravel/PHP to Go.

**Files:**
- `src/api/` (handler layer)
- `src/services/` (service layer)
- `src/repositories/` (repository + ClickHouse queries)

---

### Facebook Analytics API
Migrated Facebook analytics endpoints.

---

### Instagram Analytics API (14 endpoints)
Migrated all Instagram analytics endpoints including competitor analytics.

---

### YouTube Analytics API (15 endpoints)
Migrated YouTube analytics endpoints. 5 ClickHouse tables: `youtube_videos`, `youtube_insights`, `youtube_channel_demographics`, `youtube_traffic_sources`, `youtube_sharing_services`.

---

### Pinterest Analytics API
Migrated Pinterest analytics endpoints.

---

### GMB (Google My Business) Analytics API
Migrated GMB analytics endpoints.

---

### Overview V2 — Cross-Platform Aggregation (6 endpoints)
Migrated Overview V2 endpoints using `mv_social_daily_metrics` materialized view for cross-platform aggregation.

---

## Template for Future Entries

```
## [YYYY-MM-DD] Short Title

### Feature/Fix Name
**Problem:** What was wrong or missing.

**Fix:** What was done and why.

**Files:**
- path/to/changed/file.go
```
