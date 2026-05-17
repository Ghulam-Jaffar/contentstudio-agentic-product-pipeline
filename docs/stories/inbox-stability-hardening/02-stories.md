# Stories — Inbox Stability & Quality Hardening

All 6 stories are `[BE]` against `social-inbox-manager/`. No FE, iOS, or Android stories. Epic: **Inbox Sync Reliability Phase 2** (assigned manually in Shortcut).

---

## Story 1 — `[BE]` Split inbox consumer groups per worker type and harden the poll loop

### Description

As an inbox operator, I want each social-inbox worker type (conversations, messages, comments, webhooks, saves) to run under its own Kafka consumer group with batched offset commits and bounded retry on transient errors, so that one slow worker on a single topic does not trigger a channel-wide rebalance and we stop losing throughput to redundant per-message sync commits.

### Workflow

1. The inbox runs N worker processes per channel today (Facebook, Instagram, LinkedIn, YouTube, GMB, Pusher, auto-reply-rule-change). Every worker for the same channel currently joins a single consumer group named `<channel>-consumer-group`, regardless of which topic it consumes.
2. After this change, each worker joins a group named `<channel>-<worker_type>` (e.g. `facebook-conversations`, `facebook-save-comments`, `instagram-webhook-comment`, `linkedin-comments`, `pusher-notifications`).
3. The poll loop commits offsets every N processed messages or every M seconds (whichever first), where N and M are configurable via env. Default: N=25, M=5s.
4. When a handler raises a transient error (configurable allow-list — e.g. `ConnectionError`, `TimeoutError`, Mongo `AutoReconnect`), the loop retries the same message up to 3 times with exponential backoff (1s, 4s, 16s) before sending it to the DLQ. Non-transient errors continue to go straight to DLQ as today.
5. The dead `group.id="scram-social-media-event-user"` default in `KAFKA_CONSUMER_CONFIG` is removed. `app/kafka/consumer.py` is the only place that sets `group.id`.
6. The new naming convention and tunables are documented in `docs/RUNBOOK.md` and `docs/ARCHITECTURE.md`.

### Acceptance criteria

- [ ] Every worker subscribes to Kafka with a `group.id` of the form `<channel>-<worker_type>`; no two worker types share a `group.id`.
- [ ] Restarting one `save_comments` worker triggers a rebalance only inside the `<channel>-save-comments` group; consumers on `<channel>-conversations`, `<channel>-messages`, etc. keep processing without being assigned new partitions.
- [ ] Offset commits are batched: with `KAFKA_COMMIT_BATCH_SIZE=25` and `KAFKA_COMMIT_BATCH_INTERVAL_S=5`, an observed `commit()` rate of roughly `messages_processed / 25` (or `seconds_elapsed / 5`, whichever is lower) is recorded in logs.
- [ ] When a handler raises an exception listed in the transient-retry allow-list, the worker retries the same Kafka message up to 3 times with backoff 1s → 4s → 16s before forwarding it to the existing error DLQ topic; a non-transient exception goes to the DLQ on the first failure (current behaviour preserved).
- [ ] On `SIGTERM` / `SIGINT`, the worker flushes any uncommitted offsets before exiting (no offset loss on graceful shutdown).
- [ ] `KAFKA_CONSUMER_CONFIG` in `app/config/kafka_config.py` no longer contains a hard-coded `group.id` key.
- [ ] `docs/RUNBOOK.md` and `docs/ARCHITECTURE.md` describe the per-worker group naming and the commit-batching / retry tunables.
- [ ] All env defaults are safe (matching current behaviour at N=1, M=0, retries=0 if a deployment sets these to disable the new behaviour).

### Mock-ups

N/A — backend-only change.

### Impact on existing data

- New Kafka consumer groups (`<channel>-<worker_type>`) start consuming from `auto.offset.reset=earliest` if not bootstrapped. To avoid replay storms on first deploy, the rollout playbook in `RUNBOOK.md` must include the step *"For each new group ID, pre-create the group with the current committed offset of the parent `<channel>-consumer-group` before starting the new workers."* (e.g. `kafka-consumer-groups.sh --reset-offsets --to-offset <N>`).
- No MongoDB schema changes.

### Impact on other products

- No frontend, mobile, or Chrome-extension impact — change is internal to the inbox manager service.
- White-label domains: no impact.

### Dependencies

None.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — backend-only)
- [ ] Multilingual support (N/A — no user-facing strings)
- [ ] UI theming support (N/A — backend-only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points**
- `app/kafka/consumer.py` — `SocialDataConsumer.__init__` is where the group_id is derived; accept `worker_type` and incorporate it into the default.
- `app/workers/facebook_inbox_worker.py` — `_process_messages` is the canonical poll loop; the same shape is mirrored in `instagram_inbox_worker.py`, `linkedin_inbox_worker.py`, `youtube_inbox_worker.py`, `gmb_inbox_worker.py`, `pusher_notification_worker.py`, `auto_reply_rule_change_worker.py`. Consider extracting a shared `BaseInboxWorker._process_messages` in `app/workers/_base.py` so the change lives in one place.

**Existing patterns**
- The producer side already has fork-safety / singleton management in `app/kafka/producer.py::_KafkaProducerManager`; consider a parallel `_KafkaConsumerLoop` helper for the new commit-batching logic.
- Graceful shutdown is already wired via `setup_signal_handlers` in each worker — extend `cleanup()` to call `consumer.commit(asynchronous=False)` of any deferred offsets.

**Suggested env vars**
- `KAFKA_COMMIT_BATCH_SIZE` (default 25)
- `KAFKA_COMMIT_BATCH_INTERVAL_S` (default 5)
- `KAFKA_TRANSIENT_RETRY_MAX` (default 3)
- `KAFKA_TRANSIENT_RETRY_BASE_S` (default 1)

**Gotchas**
- librdkafka uses `queued.max.messages.kbytes`, not `max.poll.records` — keep that in mind if tuning per-batch fetch sizes.
- `auto.offset.reset=earliest` means new consumer groups must be pre-positioned at the current committed offset of the legacy group; otherwise the inbox will re-emit historical events from the Kafka retention window.

---

## Story 2 — `[BE]` Add structured tracing logs across inbox workers and strategies

### Description

As an on-call engineer, I want every log line a Kafka message produces in the inbox manager to carry a consistent set of structured fields — `trace_id`, `worker_type`, `worker_id`, `kafka_topic`, `kafka_partition`, `kafka_offset`, `account_id`, `workspace_id`, `element_id` — so that I can trace one event end-to-end across workers, strategies, and DLQs without manually correlating timestamps.

### Workflow

1. When a Kafka message arrives in `_process_messages`, the worker generates a `trace_id` (UUID4, or the value of a `trace_id` header if the upstream producer set one) and binds it on a `contextvars.ContextVar` for the duration of that message.
2. Every `logger.info / warning / error` call inside the worker, the strategy it invokes, and any helper module called from that strategy automatically includes the bound `trace_id`, the Kafka coordinates, and the business identifiers (`account_id`, `workspace_id`, `element_id`) in the log record's `extra` payload.
3. Logs are emitted as JSON (one record per line) so they parse cleanly in the existing log aggregator without regex.
4. Existing `print(...)` statements in workers and strategies (job-started / job-finished banners) are replaced with `logger.info(...)` carrying `duration_seconds`, `account_id`, `workspace_id`, `job_id`.
5. The worker entrypoint sets the default log level to **INFO** (not ERROR as today) so operational events are visible in production.
6. When this worker re-emits to a downstream topic (e.g. a sync worker pushing to a `save_*` topic), the `trace_id` is added to the outbound Kafka message headers so the downstream worker inherits it.

### Acceptance criteria

- [ ] Every log line emitted from `app/workers/*.py` and `app/social_sync/*_strategy.py` during message processing is valid JSON with fields `timestamp`, `level`, `logger`, `message`, `trace_id`, `worker_type`, `worker_id`.
- [ ] When the log line is emitted from inside the message-processing scope, the JSON record additionally contains `kafka_topic`, `kafka_partition`, `kafka_offset`, and (when extractable from the payload) `account_id`, `workspace_id`, `element_id`.
- [ ] A single Kafka message produces log lines with the **same** `trace_id` across the worker, the strategy, and the helper layers it calls — verified by a test that processes one synthetic message and asserts all emitted records share one `trace_id`.
- [ ] When the worker re-emits to a `save_*` topic, the outbound Kafka message carries a header `trace_id` set to the inbound `trace_id`; the downstream worker reads that header and binds it for its own processing scope.
- [ ] No `print(` calls remain in `app/workers/` or `app/social_sync/` (verified by a ruff/grep CI check or a unit test).
- [ ] The worker entrypoint (`main()` in each `*_inbox_worker.py`) defaults to `logging.INFO`; the level remains overridable via `LOG_LEVEL` env var.
- [ ] Sentry breadcrumbs and events continue to fire (no regression in the existing `LoggingIntegration` wiring in `app/sentry/__init__.py`).
- [ ] `docs/RUNBOOK.md` documents the JSON log schema and how to grep / filter logs by `trace_id`.

### Mock-ups

N/A — backend-only.

### Impact on existing data

None. No schema or storage changes.

### Impact on other products

- Downstream log consumers (Sentry, log aggregator) will see a new JSON format. Confirm the aggregator's parser tolerates it before rollout; this is a runbook step.
- No frontend, mobile, or Chrome-extension impact.

### Dependencies

- Should land before or alongside **[BE] Split inbox consumer groups per worker type and harden the poll loop** so the new commit-batching and retry events are observable from day one.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — backend-only)
- [ ] Multilingual support (N/A — log messages are operator-facing, English only)
- [ ] UI theming support (N/A — backend-only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points**
- New: `app/utils/logging_setup.py` exposing `configure_json_logging(default_level: str)`, `bind_trace_context(**fields)`, and a `TraceLoggerAdapter` that pulls from a module-level `contextvars.ContextVar`.
- `app/workers/facebook_inbox_worker.py::_process_messages` is the place to allocate the `trace_id` and bind context; replicate in the other 6 workers (or extract into the shared base from Story 1).
- `app/sentry/__init__.py` — keep `LoggingIntegration` but pass the JSON formatter so structured fields survive into Sentry events.

**Existing patterns to reuse**
- The `error_handler` decorator in `facebook_inbox_worker.py` already builds an `error_context` dict with the right keys (`kafka_topic`, `kafka_partition`, `kafka_offset`, `account_id`, `workspace_id`, `element_id`) — pull that into the trace-context binder so happy-path logs also get it.
- `app/utils/linkedin_helper.py` already uses `logger.error(..., extra={...})` in a handful of places — the new adapter should make this the default everywhere.

**Suggested env vars**
- `LOG_FORMAT=json|plain` (default `json` in non-dev environments)
- `LOG_LEVEL=INFO` (default)

**Gotchas**
- Don't put very large objects (full payloads) on `extra={}` — log aggregators will truncate or reject. Cap payload echo at ~500 bytes for any non-error log.
- The `confluent_kafka` Python client doesn't auto-propagate headers — explicit `headers=[("trace_id", trace_id.encode())]` on `producer.produce(...)` is required.

---

## Story 3 — `[BE]` Standardise sort parameter, field, and tie-breaker across inbox repositories

### Description

As a developer of inbox features, I want every inbox repository to accept a single canonical sort parameter and apply a stable tie-breaker on paginated lists, so that pagination does not repeat or skip rows when many documents share the same timestamp, and so I do not have to remember whether a given endpoint wants `sort_order` or `order_by`.

### Workflow

1. A single helper, `apply_inbox_sort(query, filters, default_field)`, is added under `app/database/mongo/utils/sorting.py`.
2. The helper accepts the canonical filter key `sort_order` (with `order_by` accepted as a deprecated alias for one release; a `DeprecationWarning` is logged when the alias is used).
3. The helper returns a MongoDB sort spec of the form `[(default_field, direction), ("_id", direction)]`, guaranteeing a stable tie-breaker even when many documents share the same timestamp.
4. Every repository that today calls `.sort("created_time", ...)`, `.sort("updated_at", ...)`, or `.sort("_id", ...)` is migrated to use the helper.
5. API request schemas in `app/api/` that forward sort parameters accept `sort_order` going forward and pass it through unchanged.
6. The canonical sort contract is documented in `docs/ARCHITECTURE.md` (which field is the primary sort for each list endpoint, and that `_id` is always the tie-breaker).

### Acceptance criteria

- [ ] A new helper `app/database/mongo/utils/sorting.py::apply_inbox_sort` exists and is unit-tested for: default direction (`desc`), explicit `asc`/`desc`, the deprecated `order_by` alias, an unknown value (falls back to `desc`), and the tie-breaker is always `_id` matching the primary direction.
- [ ] The following repository methods are migrated to use the helper and now return a stable tie-broken sort spec:
  - `inbox_messages_repository.get_messages_by_conv_id` (was `("created_time", -1 if sort == "desc" else 1)` — gains `_id` tie-breaker)
  - `inbox_messages_repository` — all other `.sort("created_time", ...)` call sites
  - `inbox_comments_repository` — all `.sort(...)` call sites
  - `inbox_details_repository.get_inbox_details_list` (was `("updated_at", -1)` — gains `_id` tie-breaker)
  - `inbox_details_repository` in-memory `all_elements.sort(...)` branch — secondary key is `_id`
  - `facebook_account_repository`, `instagram_account_repository`, `linkedin_account_repository`, `youtube_account_repository`, `gmb_account_repository` — `.sort("_id", -1)` calls migrated for consistency.
- [ ] Every API endpoint that forwards a sort parameter accepts `sort_order` in the request body / query string. If `order_by` is sent, the request still succeeds and a one-time `DeprecationWarning` log line fires per process.
- [ ] A regression test feeds an inbox conversation list with 50 conversations that all share `updated_at == "2026-05-17T12:00:00Z"`, paginates with `limit=10` across 5 pages, and asserts that every conversation appears exactly once and never on two pages.
- [ ] `docs/ARCHITECTURE.md` lists the canonical primary sort field for each inbox list endpoint and states that `_id` is the tie-breaker on every paginated query.
- [ ] No call site in `app/database/`, `app/services/`, or `app/api/` still uses the bare `("created_time", ...)` or `("updated_at", ...)` form without a tie-breaker (verified by a grep-based unit test).

### Mock-ups

N/A — no UI changes. The inbox UI consumes whichever order the API returns; ordering becomes more predictable, not different.

### Impact on existing data

- None. No schema migrations. Existing indexes covering `(workspace_id, conversation_id, created_time)` and `(workspace_id, updated_at)` already include `_id` implicitly as the last field in any compound index, so the tie-breaker does not require new indexes; if EXPLAIN shows a regression, add the secondary key explicitly.

### Impact on other products

- Web frontend continues to send `sort_order` (which was already the dominant key); requests that still use `order_by` keep working through the deprecation period.
- No mobile, Chrome-extension, or white-label impact.

### Dependencies

None.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — backend-only)
- [ ] Multilingual support (N/A)
- [ ] UI theming support (N/A — backend-only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points**
- `app/database/mongo/repository/inbox_messages_repository.py` — call sites at lines 101, 123, 132, 218, 229, 250, 261, 352, 387.
- `app/database/mongo/repository/inbox_comments_repository.py` — call sites at lines 161, 207, 323, 614.
- `app/database/mongo/repository/inbox_details_repository.py` — call sites at lines 455, 567, 727, 959, 1145, 1399, 1404.
- `app/services/auto_reply/reply_executor.py:497` — `sort=[("_id", -1)]` — review whether it should adopt the helper.
- Account repos: `facebook_account_repository.py` (73, 117, 121), and the parallel GMB / Instagram / LinkedIn / YouTube repos.

**Gotchas**
- MongoDB tie-breaker sort works only when `_id` is included in the compound sort *and* the result set is bounded; ensure `.limit(...)` always follows. The repo methods all already apply `.limit(...)` — no new limits needed.
- The deprecation log line should fire once per process (use a module-level `_warned: bool` flag), not once per request, to avoid log spam.

---

## Story 4 — `[BE]` Audit and standardise Kafka partition keys across all inbox channels

### Description

As an inbox operator, I want every social-platform topic in the inbox manager to use a partition key that distributes load evenly across all partitions, so that one high-volume account (or one chatty conversation, or one busy LinkedIn page) cannot create a single-partition lag spike like the 45,991-message Instagram incident we already fixed.

### Workflow

1. Every call site that publishes to a Kafka topic is reviewed: each `get_producer(...).send_social_data(payload, key=...)` (and any direct `producer.produce(...)`) is checked against a per-topic contract documented in `docs/TOPICS.md`.
2. For each topic the contract specifies: the partition key field, why it was chosen (cardinality vs. ordering trade-off), and any known concentration risk (e.g. "one conversation may dominate one partition by design — accepted because per-conversation order must be preserved").
3. `app/kafka/producer.py` is updated so that when a caller forgets to pass `key=`, the producer logs a `WARNING` (instead of silently falling back to `data.get("id")`) so future regressions surface in monitoring.
4. A new unit test, `tests/unit/test_partition_keys.py`, instantiates each strategy and webhook handler, captures every outgoing Kafka call, and asserts that `key` is non-None and is not equal to the inbound `account_id` (the prior skew vector).
5. `docs/TOPICS.md` is updated with a per-topic partition-key table; `docs/kafka-partition-skew-fix.md` is referenced as background.

### Acceptance criteria

- [ ] Every Kafka producer call in `app/social_sync/*_strategy.py` and `app/api/webhooks/*.py` passes an explicit, non-None `key` argument.
- [ ] The `key` is documented per topic in `docs/TOPICS.md` (e.g. *Facebook conversations sync* → `conversation_id`; *Facebook comments sync* → `post_id`; *Instagram comments webhook* → per-comment `id`; *LinkedIn comments webhook* → per-comment URN; *YouTube comments webhook* → per-comment `id`; *GMB review events* → per-review `id`; *Pusher notifications* → per-notification id).
- [ ] LinkedIn, YouTube, and GMB webhook and strategy code paths use the highest-cardinality per-event id available (not `account_id`).
- [ ] `app/kafka/producer.py::SocialDataProducer.send_social_data` emits `logger.warning("kafka_key_missing", extra={"topic": self.topic})` exactly once per topic per process when called without a `key`.
- [ ] `tests/unit/test_partition_keys.py` passes and covers Facebook / Instagram / LinkedIn / YouTube / GMB strategies and webhook actions; the test asserts that for each call path the captured `key` is non-None, is a string, and is not equal to the account id used in the test fixture.
- [ ] A grep-based test or CI rule fails the build if any new `get_producer(...).send_social_data(payload)` call appears without `key=` (informational — can be lint-enforced via ruff custom rule or a simple `pytest` collection check).
- [ ] `docs/TOPICS.md` lists the per-topic partition-key contract and a one-line rationale for each.

### Mock-ups

N/A — backend-only.

### Impact on existing data

- No MongoDB or schema impact.
- Kafka: changing partition keys reshards traffic across partitions. Existing in-flight messages already keyed by the old scheme stay where they are; once consumers drain, all new messages route by the new keys. The runbook for rollout must include a *"drain partition X before declaring complete"* step.

### Impact on other products

- No frontend, mobile, or Chrome-extension impact.

### Dependencies

- Benefits from **[BE] Add structured tracing logs across inbox workers and strategies** landing first, so the new `kafka_key_missing` warning lands in the structured log stream.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — backend-only)
- [ ] Multilingual support (N/A)
- [ ] UI theming support (N/A — backend-only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points**
- `app/kafka/producer.py::SocialDataProducer.send_social_data` — current fallback `kafka_key = key if key else data.get("id")` is the silent skew vector. Replace with explicit-warning behaviour.
- `app/social_sync/linkedin_strategy.py`, `youtube_strategy.py`, `gmb_strategy.py` — the gaps to audit (Facebook + Instagram were already fixed per `docs/kafka-partition-skew-fix.md`).
- `app/api/webhooks/` — confirm LinkedIn / YouTube / GMB webhook action files mirror the Instagram pattern (per-event id, not account id).

**Existing patterns to copy**
- `app/api/webhooks/instagram_webhook_actions.py::_extract_message_key` and `_extract_comment_key` already implement the right pattern. Mirror their shape for LinkedIn / YouTube / GMB.
- `docs/kafka-partition-skew-fix.md` is the canonical write-up; cite it in the new TOPICS contract.

**Gotchas**
- Per-conversation keys still concentrate one hot conversation onto one partition by design — this is the accepted trade-off because per-conversation message order must be preserved. Document explicitly so it doesn't get "fixed" later.
- The producer's fork-safety (`_check_fork()` in `_KafkaProducerManager`) is unrelated to partition keys and stays as-is.

---

## Story 5 — `[BE]` Add unit-test coverage for all inbox workers and platform strategies, wire pytest into CI

### Description

As an inbox engineer, I want every worker and every platform strategy in `social-inbox-manager` to have a unit-test file that exercises its message-processing paths and partition-key extraction, and I want `pytest` to run on every PR, so that regressions in inbox sync logic are caught before merge instead of in production.

### Workflow

1. A unit-test file is added per worker — `tests/unit/test_facebook_inbox_worker.py`, `test_instagram_inbox_worker.py`, `test_linkedin_inbox_worker.py`, `test_youtube_inbox_worker.py`, `test_gmb_inbox_worker.py`, `test_pusher_notification_worker.py`, `test_auto_reply_rule_change_worker.py`.
2. A unit-test file is added per strategy — `test_facebook_strategy.py`, `test_instagram_strategy.py`, `test_linkedin_strategy.py`, `test_youtube_strategy.py`, `test_gmb_strategy.py`, `test_base_strategy.py`, `test_sync_context.py`. (The existing `test_social_sync.py` covers shared utilities; the new files cover the platform-specific code.)
3. Each worker test mocks `SocialDataConsumer`, feeds a synthetic Kafka message, and asserts: (a) the right strategy method is called with the right arguments, (b) on success, `commit()` is invoked, (c) on the error_handler path, `send_errors_to_kafka` is invoked with the expected `error_type` (`processing_error` / `database_error` / `validation_error`).
4. Each strategy test covers: account hydration (when account exists vs. does not exist), the happy path of the main processing method, and one error path that should be caught and re-raised by the worker decorator.
5. A new `pytest` job is added to `.github/workflows/ci.yml` that installs dev dependencies, runs `pytest --cov=app --cov-report=xml`, and uploads the coverage report so the existing SonarQube job picks it up.
6. `pyproject.toml` gains `pytest-cov` in the dev group and a `[tool.coverage.run]` / `[tool.coverage.report]` section excluding `app/alembic/versions/`, `app/data_migrations/`, and `tests/`.
7. The PR description from the CI run includes the coverage delta versus `develop`.

### Acceptance criteria

- [ ] Every file under `app/workers/` (except `__init__.py`) has a matching `tests/unit/test_<filename>.py` that runs in the new CI job and passes.
- [ ] Every file under `app/social_sync/` (except `__init__.py`) has a matching `tests/unit/test_<filename>.py` that runs in the new CI job and passes.
- [ ] Worker tests cover, at minimum: (a) successful message → strategy method call → commit, (b) handler raises → `send_errors_to_kafka` called with `error_type="processing_error"`, (c) handler raises `pymongo` error → `send_errors_to_kafka` called with `error_type="database_error"`, (d) message payload is missing `account_id` → strategy returns early without committing.
- [ ] Strategy tests cover, at minimum: (a) `*_account_details` is `None` → no further work, (b) happy-path of one processing method per strategy (e.g. `process_conversations_sync` for Facebook), (c) partition-key extraction returns the expected per-event id (cross-checked with Story 4's `test_partition_keys.py`).
- [ ] `.github/workflows/ci.yml` has a new `pytest` job that runs on `pull_request` events, installs deps via `poetry install`, runs `pytest --cov=app --cov-report=xml --cov-report=term`, and uploads `coverage.xml` as an artifact.
- [ ] The SonarQube job reads the uploaded `coverage.xml` (via the existing `sonar-project.properties` — add `sonar.python.coverage.reportPaths=coverage.xml`).
- [ ] PR CI fails when any new commit lowers the per-file coverage of a file already covered by the new suite (Sonar quality-gate or a `pytest-cov --fail-under=` floor).
- [ ] Total test count before vs. after is reported in `tests/README.md` (or similar) so the team has a baseline.

### Mock-ups

N/A — backend-only.

### Impact on existing data

None.

### Impact on other products

- No frontend, mobile, or Chrome-extension impact.

### Dependencies

- Benefits from **[BE] Audit and standardise Kafka partition keys across all inbox channels** because it provides the `test_partition_keys.py` reusable fixtures.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — backend-only)
- [ ] Multilingual support (N/A)
- [ ] UI theming support (N/A — backend-only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points**
- New tests under `tests/unit/test_*_inbox_worker.py` and `tests/unit/test_*_strategy.py`.
- Touched: `.github/workflows/ci.yml`, `pyproject.toml`, `sonar-project.properties`, `tests/conftest.py` (or new — add shared Kafka-message fixture and Mongo mock fixtures).

**Existing patterns to copy**
- `tests/unit/test_kafka_producer_manager.py` and `tests/unit/test_kafka.py` already show how to mock confluent-kafka in this repo. Reuse those patterns for consumer mocks.
- `tests/integration/test_facebook_webhooks.py` demonstrates the FastAPI test-client style; integration coverage will continue to live there.

**Gotchas**
- `confluent_kafka.Message` is a C-extension type — use a small `FakeMessage` namedtuple in `conftest.py` exposing `.value()`, `.key()`, `.topic()`, `.partition()`, `.offset()` rather than trying to construct the real type.
- `MongoConnection` is a module-level singleton — many tests will need to patch `app.database.mongo.MongoConnection` at import time. Add a `mock_mongo` autouse fixture in `conftest.py`.

---

## Story 6 — `[BE]` Integrate Ruff into social-inbox-manager and enforce on CI

### Description

As an inbox engineer, I want `ruff` to lint and format every Python file in `social-inbox-manager`, with the rule set committed to source and enforced on every PR, so that import-order drift, unused imports, basic bug-risk patterns, and formatting churn stop showing up in code review.

### Workflow

1. A `[tool.ruff]` configuration block is added to `pyproject.toml` selecting the rule families `E`, `F`, `W` (pycodestyle / pyflakes / pycodestyle warnings), `I` (import sorting), `B` (flake8-bugbear), `UP` (pyupgrade), and `SIM` (flake8-simplify), targeting Python 3.9 (`target-version = "py39"`).
2. Generated and legacy directories are excluded: `app/alembic/versions/`, `app/data_migrations/`, `.serena`, `tests/dataset`.
3. A repo-wide cleanup commit runs `ruff check --fix` followed by `ruff format` and lands the auto-fix diff as a single commit (separate from any behavioural change) so reviewers can scan it quickly.
4. A `ruff` job is added to `.github/workflows/ci.yml` that runs on every PR and blocks merges when `ruff check` reports any error or `ruff format --check` reports a formatting drift.
5. A `.pre-commit-config.yaml` is added at the repo root with the `ruff-pre-commit` hook so contributors get the same feedback locally before pushing.
6. `CONTRIBUTING.md` documents the new lint contract and how to run `poetry run ruff check --fix && poetry run ruff format` locally.

### Acceptance criteria

- [ ] `pyproject.toml` contains a `[tool.ruff]` section with `target-version = "py39"`, `line-length = 120` (matching the existing de-facto width), and the selected rule families above. Excludes are listed.
- [ ] `poetry run ruff check .` exits 0 on the `develop` branch after the cleanup commit lands.
- [ ] `poetry run ruff format --check .` exits 0 on the `develop` branch after the cleanup commit lands.
- [ ] `.github/workflows/ci.yml` has a new `ruff` job that runs on `pull_request`, installs deps via `poetry install`, runs `ruff check .` and `ruff format --check .`, and the job blocks the merge when either command fails.
- [ ] `.pre-commit-config.yaml` exists at the repo root, references `astral-sh/ruff-pre-commit` pinned to a specific version, and runs both `ruff` (with `--fix`) and `ruff-format` on staged files.
- [ ] `CONTRIBUTING.md` (new or updated) explains: how to install the pre-commit hook, how to run ruff locally, and what to do if a rule should be ignored for a specific line (`# noqa: <rule>` with a comment).
- [ ] The cleanup commit is isolated (does only ruff-driven changes — no logic changes); its diff is reviewable in one sitting.
- [ ] No `# noqa` annotations are introduced in the cleanup commit without an inline comment explaining why.

### Mock-ups

N/A — backend-only.

### Impact on existing data

None.

### Impact on other products

- No frontend, mobile, or Chrome-extension impact.

### Dependencies

None.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — backend-only)
- [ ] Multilingual support (N/A)
- [ ] UI theming support (N/A — backend-only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Current state**
- `pyproject.toml` already declares `ruff = "^0.14.2"` under `[tool.poetry.group.dev.dependencies]`; the binary is installed but never invoked.
- No `[tool.ruff]` block exists.
- No pre-commit config exists.
- `.github/workflows/ci.yml` runs SonarQube and CodiumAI PR-Agent — neither replaces ruff for fast inline feedback.

**Suggested rule selection**
```toml
[tool.ruff]
target-version = "py39"
line-length = 120
exclude = [
  "app/alembic/versions",
  "app/data_migrations",
  ".serena",
  "tests/dataset",
]

[tool.ruff.lint]
select = ["E", "F", "W", "I", "B", "UP", "SIM"]
ignore = [
  "E501",  # line-length handled by formatter
  "B008",  # FastAPI Depends() default uses function calls
]
```

**Gotchas**
- The repo currently uses Python 3.9 syntax. `UP` rules will suggest some Python 3.10+ idioms (`X | Y` unions, `match` statements) — pin `target-version = "py39"` so ruff respects the runtime.
- The `B008` rule fires on FastAPI's idiomatic `Depends(...)` in function signatures — ignore it.
- The first ruff-format pass on a large codebase produces a big diff. Land it as a standalone commit so `git blame` history stays clean (most teams add it to a `.git-blame-ignore-revs` file).
