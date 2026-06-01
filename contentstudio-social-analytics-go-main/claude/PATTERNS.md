# Coding Patterns & Conventions

Validated patterns established during development. Follow these exactly when adding new code.

---

## HTTP Retry — `doWithRetry`

Every social API client wraps HTTP calls in a retry helper. Rules:

- **3 attempts**, backoff: `attempt * 500ms` between retries
- **Never retry** 401, 403 (auth errors — retrying won't fix them), 429 (rate limit — platform logic handles this)
- **Retry** 5xx responses and network errors

### POST Body — Factory Pattern

`http.Request` body is an `io.Reader`. Once read on attempt 1, it is exhausted — attempt 2 sends an empty body. Fix: pass a factory function so the body reader is recreated fresh each time.

```go
body, status, err := c.doWithRetry(ctx, "MethodName", func() (*http.Request, error) {
    return http.NewRequestWithContext(ctx, http.MethodPost, url,
        bytes.NewReader(bodyBytes))  // fresh reader on every attempt
})
```

Always store the POST body as `[]byte` before the loop. For form-encoded: `bytes.NewReader([]byte(values.Encode()))`.

### GET Requests — Direct Request

GET requests have no body, so no factory needed — pass the request directly.

---

## Twitter — OAuth Re-Signing on Retry

OAuth 1.0a signatures embed a timestamp and nonce. A signature from attempt 1 is rejected as expired/replayed on attempt 2+.

**Fix:** Call `c.signRequest(req, ...)` inside the factory, not outside the loop.

```go
c.doWithRetry(ctx, "FetchUserTweets", func() (*http.Request, error) {
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    c.signRequest(req, oauthToken, oauthTokenSecret)  // fresh signature each attempt
    return req, nil
})
```

---

## Facebook — `waitRate` Inside `doWithRetry`

Facebook's `doWithRetry` calls `waitRate` internally. Never call `waitRate` explicitly before `doWithRetry` — it will double-count rate limit slots.

**When migrating** a Facebook function from `httpClient.Do` to `doWithRetry`: delete the preceding `c.waitRate(...)` call.

---

## Parallel Goroutines — Pre-Allocated Slices

When running goroutines that write to a shared slice by index, pre-allocate the slice so each goroutine writes to its own slot (no mutex needed):

```go
results := make([]Result, len(items))
var wg sync.WaitGroup
for i := range items {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        results[i] = process(items[i])  // safe: each goroutine owns its index
    }(i)
}
wg.Wait()
```

Use a semaphore to bound concurrency:
```go
sem := semaphore.NewWeighted(int64(maxConcurrent))
// inside goroutine:
sem.Acquire(ctx, 1)
defer sem.Release(1)
```

---

## Token Resolution

```go
func resolveAccessToken(token, key, id string, log *logger.Logger) string {
    if token == "" { return "" }
    if strings.HasPrefix(token, "IGAA") || strings.HasPrefix(token, "EAA") { return token }
    if dec, err := crypto.DecryptToken(token, key); err == nil && dec != "" { return dec }
    log.Error().Str("id", id).Msg("Token decryption failed; account will be skipped")
    return ""  // never return the raw ciphertext — it will cause a confusing 401
}
```

On `""` return, the caller records a MongoDB processing error and skips the account cleanly.

---

## Logging Levels

| Level   | When to use                                                        |
|---------|--------------------------------------------------------------------|
| `Error` | Recoverable failure that causes a job/account to be skipped        |
| `Warn`  | Degraded but continuing (e.g. channel full, non-critical drop)     |
| `Info`  | Normal completion of a unit of work                                |
| `Debug` | Verbose detail useful only during active development               |

Always include `Str("instagram_id", ...)` / `Str("account_id", ...)` on every log line inside a job handler so failures are traceable.

---

## Kafka Produce — Always Check the Error

```go
data, err := json.Marshal(payload)
if err != nil {
    log.Error().Err(err).Str("account_id", id).Msg("Marshal failed; skipping Kafka publish")
    return
}
if err := producer.Produce(ctx, topic, key, data); err != nil {
    log.Error().Err(err).Str("account_id", id).Msg("Kafka produce failed")
}
```

Never use `data, _ := json.Marshal(...)` and never ignore `producer.Produce` return value.

---

## ClickHouse — Always Use Batch

```go
batch, err := conn.PrepareBatch(ctx, `INSERT INTO table_name`)
for _, item := range items {
    batch.Append(item.Field1, item.Field2, ...)
}
batch.Send()
```

Individual inserts cause "too many parts" errors in ClickHouse. Always batch (1000 items per batch is the target).
