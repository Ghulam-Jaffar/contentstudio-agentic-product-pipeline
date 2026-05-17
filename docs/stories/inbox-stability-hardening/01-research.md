# Research — Inbox Stability & Quality Hardening

Service: `social-inbox-manager/` (Python, FastAPI + confluent-kafka + MongoDB + Redis + Pusher).
All 6 stories in this batch are backend technical-debt / observability work. No user-facing UI changes. No mobile impact.

---

## 1. Consumer group code-level improvements

**Current state — `app/kafka/consumer.py` + `app/workers/*.py`**

- `SocialDataConsumer.__init__` derives `group_id` from topic prefix: `facebook-consumer-group`, `instagram-consumer-group`, `linkedin-consumer-group`, `youtube-consumer-group`, `gmb-consumer-group`, `social-inbox-shared-consumer-group`. **Every worker on the same channel shares one group**, regardless of which topic the worker subscribes to (`conversations`, `posts`, `messages`, `comments`, `webhook_message`, `webhook_comment`, `save_*`).
- Result: a single rebalance triggered by one slow consumer (e.g. a long-running `save_*` worker exceeding `max.poll.interval.ms=600000`) blocks the whole channel.
- `_process_messages` loop in every worker is identical (Facebook/Instagram/LinkedIn/YouTube/GMB/pusher):
  - `poll(0.5)` → process one message → `consumer.commit()` (sync, one-at-a-time, no offset batching).
  - No `produce/poll(0)` flushing pattern for any embedded producer activity.
  - No retry/backoff on transient processing errors — the error decorator sends to DLQ and re-raises, but the loop just continues with the next message.
  - `_check_fork()` does not exist on the consumer side (producer has it; consumer doesn't); under gunicorn prefork this is not currently used because workers are stand-alone CLIs, but worth keeping in mind.
- Config (`app/config/kafka_config.py` `KAFKA_CONSUMER_CONFIG`):
  - `enable.auto.commit=False`, `max.poll.interval.ms=600000` (10 min), `session.timeout.ms=45000`, `heartbeat.interval.ms=15000`, `fetch.min.bytes=1`, `fetch.wait.max.ms=100`.
  - No `max.poll.records` analogue (librdkafka uses `queued.max.messages.kbytes` / `fetch.max.bytes`) — defaults apply.
  - The single `group.id=scram-social-media-event-user` in the config base is unused (overwritten in `SocialDataConsumer.__init__`) but it's misleading dead code.

**What needs to change**
- Per-worker-type group IDs (e.g. `facebook-conversations`, `facebook-save-comments`) so a slow worker for one topic doesn't trigger a channel-wide rebalance.
- Tunable commit batching (commit every N messages or every M seconds, whichever first) instead of per-message sync commit.
- A small retry-with-backoff wrapper around transient errors before DLQ (currently every exception goes straight to DLQ).
- Remove the dead `group.id` default in `KAFKA_CONSUMER_CONFIG`.
- Document the new group naming convention in `docs/RUNBOOK.md` / `docs/ARCHITECTURE.md`.

**Files involved**
- `app/kafka/consumer.py`
- `app/config/kafka_config.py`
- `app/workers/facebook_inbox_worker.py` (and the parallel Instagram/LinkedIn/YouTube/GMB/pusher workers — all share the same `_process_messages` shape, so the change is one helper + 6 wrappers)
- `docs/RUNBOOK.md`, `docs/ARCHITECTURE.md` (update)

---

## 2. Logging in inbox for tracing

**Current state**
- 208 logger calls across `app/workers/` and `app/social_sync/`; ~34 `print(...)` calls still in `workers/` and `social_sync/` (e.g. `facebook_inbox_worker.process_facebook_conversations` uses `print` with dashes for "job started"/"job finished").
- Sentry init (`app/sentry/__init__.py`) attaches a `LoggingIntegration` (INFO breadcrumbs, ERROR events), but nothing propagates a per-message trace identifier.
- Worker entrypoint sets `logging.basicConfig(level=logging.ERROR, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")` — plain text, level **ERROR**, meaning most `logger.info(...)` calls in normal operation are filtered out in production. (See `facebook_inbox_worker.main:524`.)
- A handful of files in `app/utils/linkedin_helper.py` use `extra={...}` for structured context, but the rest of the codebase only uses f-string interpolation.
- The error decorator in `facebook_inbox_worker.error_handler` does build a rich `error_context` (kafka topic/partition/offset, account_id, workspace_id, element_id) but only on the error path.

**What needs to change**
- Adopt structured/JSON logging (stdlib `logging` with a JSON formatter — no need to swap to `structlog` unless desired). Default formatter exposes `kafka_topic`, `kafka_partition`, `kafka_offset`, `account_id`, `workspace_id`, `element_id`, `trace_id`, `worker_type`, `worker_id`.
- Generate a `trace_id` per Kafka message in `_process_messages` (UUID4 or carry from a header if present) and bind it on a `logging.LoggerAdapter` / `contextvars.ContextVar` for the duration of message processing so every log line inside `process_facebook_*` carries it.
- Bump default level to INFO in worker entrypoints (the current ERROR default makes ops blind).
- Replace `print(...)` job-start/job-finish lines with `logger.info(...)` carrying the same context (account_id, workspace_id, duration_seconds).
- Optionally propagate `trace_id` into outbound Kafka headers when this worker re-emits to a `save_*` topic, so downstream workers inherit it.

**Files involved**
- New: `app/utils/logging_setup.py` (formatter + adapter helper)
- Touched: all 7 worker files (`app/workers/*.py`), the 5 social_sync strategies (`app/social_sync/*_strategy.py`), `app/sentry/__init__.py`, all 7 `run_*_consumer_workers.py` runners + `cli.py`.

---

## 3. Sorting consistency across inbox

**Current state — inconsistent in three dimensions**

1. **Param name** — repos accept different filter keys for the same concept:
   - `inbox_messages_repository.py` and `inbox_comments_repository.py` use `filters.get("sort_order")`.
   - `inbox_details_repository.py` uses `filters.get("order_by")`.
2. **Sort key** — list views sort by different fields:
   - `inbox_details_repository.get_inbox_details_list` sorts by `updated_at` (line 1404).
   - `inbox_details_repository` in-memory paginated branch sorts by a `get_sort_key` derived from `updated_at` (line 959).
   - Messages and comments sort by `created_time`.
   - Account repos sort by `_id` (`facebook_account_repository.py`, same for GMB/IG/YT/LinkedIn).
3. **Tie-breaker** — some queries have a secondary sort, others don't:
   - `inbox_messages_repository.py:352, 387` → `[("created_time", -1), ("_id", -1)]` (good — stable).
   - `inbox_messages_repository.py:123, 132, 229, 261` → `("created_time", -1)` only (unstable when many messages share a second).
   - `inbox_details_repository.py:1404` → `("updated_at", -1)` only.

   Without a stable tie-breaker, paginated results (skip/limit) can repeat or skip rows when many docs share the same timestamp.

**What needs to change**
- One shared helper (`app/database/mongo/utils/sorting.py` or similar) that takes a `filters` dict and returns `(sort_field, direction, tie_breaker)` so every repo derives the sort the same way.
- Canonicalise the filter key to one name (`order_by` or `sort_order`, pick one) and update all call sites + API request shapes; accept the other as a deprecated alias for one release.
- Add `_id` (or `message_id` / `comment_id`) as the tie-breaker on every paginated list query that currently sorts by `created_time` or `updated_at` only.
- Document the canonical sort contract in `docs/ARCHITECTURE.md`.

**Files involved**
- `app/database/mongo/repository/inbox_messages_repository.py`
- `app/database/mongo/repository/inbox_comments_repository.py`
- `app/database/mongo/repository/inbox_details_repository.py`
- `app/database/mongo/repository/facebook_account_repository.py` (and the parallel GMB/Instagram/LinkedIn/YouTube account repos)
- `app/services/auto_reply/reply_executor.py:497` (uses `sort=[("_id", -1)]` — check whether it should adopt the standard)
- API request schema definitions in `app/api/` that forward `sort_order` / `order_by`

---

## 4. Partition key research & remedy across all channels

**Current state**
- `docs/kafka-partition-skew-fix.md` and `KAFKA_FIXES.md` show a prior fix for Instagram webhook events: switched the partition key from `entry['id']` (account ID) to per-message `mid` / per-comment `id`, eliminating skew where partition 0 carried 45,991 messages of lag vs other partitions at zero.
- Producer fallback in `app/kafka/producer.py` still does `kafka_key = key if key else data.get("id")` — anywhere a caller omits the `key=` argument, partitioning silently falls back to `data.get("id")` (typically account_id), reintroducing skew.
- Confirmed correct partition keys today (per-message / per-comment / per-conversation / per-post):
  - `facebook_strategy.py` — uses `conversation_id` for messages, `post_id` for comments (lines 338, 382, 408, 442, 458, 498, 551, 578, 681, 714).
  - `instagram_strategy.py:492` — uses `post_id` for comments.
  - `facebook_webhook_actions.py` — uses `message_id` / `comment_id` with UUID4 fallback.
  - `instagram_webhook_actions.py` — same.
- **Gaps to verify** during the work:
  - LinkedIn webhooks/strategy — confirm per-comment key.
  - YouTube webhooks/strategy — confirm per-comment key.
  - GMB strategy — confirm per-review key (high-volume locations could skew).
  - Pusher notification topic — key currently?
  - Conversation-id partitioning still concentrates a single hot conversation onto one partition; document whether this is intentional (preserves message order within a conversation, the usual trade-off) and what the mitigation is when one conversation dominates.
- Worker partition assignment is `partition=None` (subscribe), so the coordinator handles assignment — no change needed there.

**What needs to change**
- An audit pass over every `get_producer(...).send_social_data(entry, key=...)` (and the equivalent inside strategies/helpers) to verify the `key` is the highest-cardinality identifier appropriate for the topic.
- Add a small lint/test guard: a unit test that constructs each strategy's producer call paths and asserts the `key` is non-None and is not the account-level identifier.
- Update `app/kafka/producer.py` fallback: instead of silently defaulting to `data.get("id")`, log a WARNING when `key` is omitted so future regressions surface early.
- Document the per-topic partition-key contract in `docs/TOPICS.md`.

**Files involved**
- `app/kafka/producer.py`
- `app/social_sync/linkedin_strategy.py`, `youtube_strategy.py`, `gmb_strategy.py`
- `app/api/webhooks/linkedin_webhook_actions.py`, `youtube_webhook_actions.py`, `gmb_webhook_actions.py` (if present)
- `app/workers/pusher_notification_worker.py` (producer side, if it re-emits)
- `docs/TOPICS.md`
- New: `tests/unit/test_partition_keys.py`

---

## 5. Test coverage for all inbox modules

**Current state**
- 14 unit tests + 3 integration tests under `tests/`.
- Modules **without any matching test file**:
  - **Workers** — every one: `facebook_inbox_worker.py`, `instagram_inbox_worker.py`, `linkedin_inbox_worker.py`, `youtube_inbox_worker.py`, `gmb_inbox_worker.py`, `pusher_notification_worker.py`, `auto_reply_rule_change_worker.py`.
  - **Strategies** — every one: `base_strategy.py`, `facebook_strategy.py`, `instagram_strategy.py`, `linkedin_strategy.py`, `youtube_strategy.py`, `gmb_strategy.py`, `twitter_strategy.py`, `sync_context.py`. (Note: `tests/unit/test_social_sync.py` exists but covers shared sync utilities, not the platform strategies.)
- CI runs SonarQube + CodiumAI PR-Agent (`.github/workflows/ci.yml`) — no `pytest` step today; no coverage gate. Sonar will compute coverage if a report is produced, but nothing publishes one.

**What needs to change**
- Add unit tests for every worker's `_process_messages` and per-topic handlers (mock `SocialDataConsumer`, feed synthetic Kafka messages, assert the right strategy method gets called with the right payload and that errors hit the DLQ producer).
- Add unit tests for each strategy's core methods (account hydration, message/post processing, partition-key extraction).
- Add a `pytest` job to `.github/workflows/ci.yml` that runs unit + integration suites and uploads `coverage.xml` for Sonar to consume.
- Set a baseline coverage threshold (e.g. start at current % — Sonar will block PRs that lower it).

**Files involved**
- New tests under `tests/unit/test_workers_*.py`, `tests/unit/test_strategies_*.py`.
- Updated `.github/workflows/ci.yml` with a `pytest` job.
- `pyproject.toml` — add `pytest-cov` to dev deps; configure `[tool.coverage.run]` `[tool.coverage.report]`.

---

## 6. Ruff integration for code quality

**Current state**
- `pyproject.toml` already lists `ruff = "^0.14.2"` under `[tool.poetry.group.dev.dependencies]` — installed but never invoked.
- **No `[tool.ruff]` configuration block** in `pyproject.toml`.
- No `ruff check` step in `.github/workflows/ci.yml`. SonarQube does some Python rule coverage but is no substitute for ruff's lint + format speed.
- No pre-commit hooks configured at repo root (no `.pre-commit-config.yaml`).

**What needs to change**
- Add a `[tool.ruff]` section to `pyproject.toml` choosing rule families: `E`, `F`, `W`, `I` (imports), `B` (bugbear), `UP` (pyupgrade), `SIM` (simplify); target Python 3.9 (per `pyproject.toml`); exclude `alembic/versions/`, `data_migrations/`, `.serena`.
- Run `ruff check --fix` and `ruff format` once on the whole repo and land the auto-fix diff as the first commit on the story.
- Add a `ruff` job to `.github/workflows/ci.yml` that runs on every PR and blocks on failure.
- Optionally add a `.pre-commit-config.yaml` with the `ruff-pre-commit` hook so contributors get the same feedback locally.
- Document the lint contract in `CONTRIBUTING.md`.

**Files involved**
- `pyproject.toml`
- `.github/workflows/ci.yml`
- New: `.pre-commit-config.yaml` (optional)
- `CONTRIBUTING.md`
- One repo-wide cleanup commit from `ruff check --fix && ruff format`.

---

## Cross-cutting notes

- **Mobile** — none of the 6 items touches iOS or Android. No mobile stories needed.
- **User-visible impact** — none of the 6 changes the inbox UI. Sorting (item 3) is the only candidate, but the change is internal: callers continue to receive sorted lists; what changes is consistency and pagination stability. No FE story needed; the frontend will keep using whichever filter key we standardise on (back-compat alias for one release).
- **Epic** — all 6 go under the current quarterly miscellaneous epic. Per `.claude/shortcut-config.json`, that is `miscellaneous_epics.q2_2026 = 115078`.
- **Product area** — `inbox`. **Skill set** — `backend`. **Project** — `web_app` (2554). **Priority** — recommend `medium` for items 1/3/5/6 and `high` for items 2 (logging) and 4 (partitioning) given prior production incidents; user can override.
