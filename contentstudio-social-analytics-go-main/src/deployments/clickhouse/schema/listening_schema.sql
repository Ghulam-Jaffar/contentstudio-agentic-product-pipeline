-- listening_mentions: stores normalized social listening mentions.
-- Uses ReplacingMergeTree to deduplicate by (topic_id, platform, mention_id) keeping the latest updated_at.
-- post_read / post_irrelevant / bookmark are user-interaction flags updated via a new insert
-- with the same sort key + a newer updated_at value (ReplacingMergeTree upsert pattern).
CREATE TABLE IF NOT EXISTS listening_mentions
(
    mention_id       String,
    topic_id         String,
    platform         LowCardinality(String),
    native_id        String,
    content_hash     String,
    author_id        String,
    author_name      String,
    author_handle    String,
    author_image_url String,
    author_url       String,
    author_followers Int64     DEFAULT 0,
    post_text        String,
    language         LowCardinality(String) DEFAULT '',
    posted_at        DateTime,
    matched_keywords Array(String),
    total_engagement Int64,
    likes_count      Int64     DEFAULT 0,
    comments_count   Int64     DEFAULT 0,
    shares_count     Int64     DEFAULT 0,
    content_type     LowCardinality(String),
    media_type       LowCardinality(String),
    url              String,
    media_urls       Array(String),
    ai_tags          Array(String),
    sentiment_label  LowCardinality(String),
    sentiment_score  Float64,
    created_at       DateTime DEFAULT now(),
    updated_at       DateTime DEFAULT now(),
    post_read        Bool     DEFAULT false,
    post_irrelevant  Bool     DEFAULT false,
    bookmark         Bool     DEFAULT false,
    sentiment_override String DEFAULT ''
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(posted_at)
ORDER BY (topic_id, platform, mention_id)
SETTINGS index_granularity = 8192;

-- listening_daily_stats: materialized view aggregating daily mention counts and engagement per topic/platform.
CREATE MATERIALIZED VIEW IF NOT EXISTS listening_daily_stats
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (topic_id, platform, day)
AS
SELECT
    topic_id,
    platform,
    toDate(posted_at)                              AS day,
    count()                                        AS mention_count,
    sumIf(1, sentiment_label = 'positive')         AS positive_count,
    sumIf(1, sentiment_label = 'negative')         AS negative_count,
    sumIf(1, sentiment_label = 'neutral')          AS neutral_count,
    sum(total_engagement)                          AS total_engagement
FROM listening_mentions
GROUP BY topic_id, platform, day;
