# Listening Module — Entity Relationship Diagram

## Overview

The listening module uses two databases:

- **MongoDB** — source of truth for topic configuration, views, and workspace quotas
- **ClickHouse** — append-optimised analytics store for mention events

---

## MongoDB

```
listening_topics
────────────────────────────────────────────────
_id                   ObjectID       PK
topic_id              String         UK
workspace_id          String
super_admin_id        String
name                  String
status                String         ("active" | "paused" | "deleted")
include_keywords      []String
exclude_keywords      []String
include_any           []String
include_all           []String
exact_match           Bool
case_sensitive        Bool
include_authors       []String
exclude_authors       []String
languages             []String
regions               []String
enabled_platforms     []String
is_initial_sync_done  Bool
mentions_limit        Int
mentions_limit_reached Bool
usage:
  mentions_count      Int
  mentions_limit      Int
last_fetched_at       DateTime
last_fetched_cursors  Map<String,String>
created_at            DateTime
updated_at            DateTime
```

```
listening_workspace_usage
────────────────────────────────────────────────
_id                   ObjectID       PK
workspace_id          String         UK
super_admin_id        String
mention_limit_monthly Int
topic_limit           Int
mentions_this_month   Int
mention_limit_reached Bool
updated_at            DateTime
```

```
listening_views
────────────────────────────────────────────────
_id            ObjectID       PK
workspace_id   String
name           String
icon           String
type           String         ("system" | "user")
filter_preset:
  topic_ids    []String
  platforms    []String
  sentiments   []String
  ai_tags      []String
  min_followers Int
  language     []String
created_at     DateTime
updated_at     DateTime
```

### MongoDB Relationships

```
listening_topics ──────────── listening_workspace_usage
  workspace_id    (logical)     workspace_id
  super_admin_id  (logical)     super_admin_id

listening_views ──────────── listening_topics
  workspace_id    (logical)   workspace_id
  filter_preset.topic_ids     topic_id  (optional filter)
```

> Relationships are logical only — MongoDB has no enforced foreign key constraints.

---

## ClickHouse

```
listening_mentions
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(posted_at)
ORDER BY (topic_id, platform, mention_id)
────────────────────────────────────────────────
mention_id         String                  ← "{platform}:{native_id}"
topic_id           String                  ← FK → listening_topics.topic_id
platform           LowCardinality(String)
native_id          String
content_hash       String
author_id          String
author_name        String
author_handle      String
author_image_url   String
author_url         String
author_followers   Int64         DEFAULT 0
post_text          String
language           LowCardinality(String)  DEFAULT ''
posted_at          DateTime
matched_keywords   Array(String)
total_engagement   Int64
likes_count        Int64         DEFAULT 0
comments_count     Int64         DEFAULT 0
shares_count       Int64         DEFAULT 0
content_type       LowCardinality(String)
media_type         LowCardinality(String)
url                String
media_urls         Array(String)
ai_tags            Array(String)
sentiment_label    LowCardinality(String)
sentiment_score    Float64
sentiment_override String        DEFAULT ''
post_read          Bool          DEFAULT false  ─┐ user interaction flags
post_irrelevant    Bool          DEFAULT false   │ updated via a new insert
bookmark           Bool          DEFAULT false  ─┘ with newer updated_at
created_at         DateTime      DEFAULT now()
updated_at         DateTime      DEFAULT now()   ← ReplacingMergeTree version key
```

```
listening_daily_stats  (Materialized View over listening_mentions)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (topic_id, platform, day)
────────────────────────────────────────────────
topic_id          String
platform          LowCardinality(String)
day               Date
mention_count     UInt64
positive_count    UInt64
negative_count    UInt64
neutral_count     UInt64
total_engagement  UInt64
```

### ClickHouse Relationships

```
listening_mentions ──────── listening_topics  (MongoDB)
  topic_id          (logical)  topic_id
```

> `listening_daily_stats` is automatically populated by ClickHouse from `listening_mentions` inserts.

---

## Cross-Database Data Flow

```
                   MongoDB                          ClickHouse
                   ───────                          ──────────

listening_topics ──► Kafka work order ──► Fetcher ──► Parser ──► Sentiment ──► listening_mentions
     topic_id                                                                        topic_id
     workspace_id                                                                    platform
     keywords                                                                        mention_id
     platforms                                                                       ...
         │
         └──► listening_workspace_usage  (quota check before dispatch)
                   workspace_id

listening_views ──► filter_preset applied at query time to listening_mentions
                         (topic_ids, platforms, sentiments, ai_tags, min_followers, language)
```

---

## Migrations

Sequential listening schema docs are in `deployment/clickhouse/migrations/listening/`:

| File | Change |
|------|--------|
| `000_create_listening_mentions.md` | Document the `listening_mentions` table and `listening_daily_stats` materialized view schema |
| `001_add_interaction_columns.sql` | Add `post_read`, `post_irrelevant`, `bookmark` |
| `002_add_sentiment_override.sql` | Add `sentiment_override` |
| `003_add_author_and_engagement_fields.sql` | Add `author_handle`, `author_image_url`, `likes_count`, `comments_count`, `shares_count`, `media_urls` |
| `004_add_author_url.sql` | Add `author_url` |
| `005_add_listening_filter_fields.sql` | Add `author_followers`, `language`, `ai_tags` |
| `006_seed_random_tags_sentiment.sql` | Seed test data with deterministic sentiment + AI tags |
