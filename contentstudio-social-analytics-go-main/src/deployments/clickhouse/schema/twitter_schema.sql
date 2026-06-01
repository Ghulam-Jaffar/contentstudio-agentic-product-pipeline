-- Twitter Tables Schema

-- Twitter Insights Table
CREATE TABLE contentstudiobackend.twitter_insights
(
    `twitter_id` String CODEC(ZSTD(1)),
    `record_id` String CODEC(ZSTD(1)),
    `name` String CODEC(ZSTD(1)),
    `username` String CODEC(ZSTD(1)),
    `profile_image_url` String CODEC(ZSTD(1)),
    `description` String CODEC(ZSTD(1)),
    `verified` String CODEC(ZSTD(1)),
    `account_created_date` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `followers_count` Int64 CODEC(T64, ZSTD(1)),
    `following_count` Int64 CODEC(T64, ZSTD(1)),
    `tweet_count` Int64 CODEC(T64, ZSTD(1)),
    `listed_count` Int64 CODEC(T64, ZSTD(1)),
    `like_count` Int64 CODEC(T64, ZSTD(1)),
    `day_of_week` Int64 CODEC(T64, ZSTD(1)),
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000'
)
ENGINE = ReplacingMergeTree(saving_time)
PARTITION BY toYYYYMM(saving_time)
PRIMARY KEY record_id
ORDER BY record_id
SETTINGS index_granularity = 8192;

-- Twitter Posts Table
CREATE TABLE contentstudiobackend.twitter_posts
(
    `twitter_id` String CODEC(ZSTD(1)),
    `name` String CODEC(ZSTD(1)),
    `username` String CODEC(ZSTD(1)),
    `profile_image_url` String CODEC(ZSTD(1)),
    `followers_count` Int64 CODEC(T64, ZSTD(1)),
    `following_count` Int64 CODEC(T64, ZSTD(1)),
    `tweet_count` Int64 CODEC(T64, ZSTD(1)),
    `listed_count` Int64 CODEC(T64, ZSTD(1)),
    `tweet_id` String CODEC(ZSTD(1)),
    `edit_history_tweet_ids` Array(String) CODEC(ZSTD(1)),
    `author_id` String CODEC(ZSTD(1)),
    `author_username` String CODEC(ZSTD(1)),
    `id_created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `author_id_created` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `tweeted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `hashtags` Array(String) CODEC(ZSTD(1)),
    `permalink` String CODEC(ZSTD(1)),
    `tweet_type` String CODEC(ZSTD(1)),
    `urls` Array(String) CODEC(ZSTD(1)),
    `media_url` Array(String) CODEC(ZSTD(1)),
    `username_mentioned` Array(String) CODEC(ZSTD(1)),
    `userid_mentioned` Array(String) CODEC(ZSTD(1)),
    `lang` String CODEC(ZSTD(1)),
    `tweet_text` String CODEC(ZSTD(1)),
    `impression_count` Int64 CODEC(T64, ZSTD(1)),
    `retweet_count` Int64 CODEC(T64, ZSTD(1)),
    `reply_count` Int64 CODEC(T64, ZSTD(1)),
    `like_count` Int64 CODEC(T64, ZSTD(1)),
    `bookmark_count` Int64 CODEC(T64, ZSTD(1)),
    `quote_count` Int64 CODEC(T64, ZSTD(1)),
    `total_engagement` Int64 CODEC(T64, ZSTD(1)),
    `day_of_week` Int64 CODEC(T64, ZSTD(1)),
    `hour_of_day` Int64
)
ENGINE = ReplacingMergeTree(saving_time)
PARTITION BY toYYYYMM(tweeted_at)
PRIMARY KEY tweet_id
ORDER BY (tweet_id, tweeted_at)
SETTINGS index_granularity = 8192;