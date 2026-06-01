-- Pinterest Users Table
-- Stores user profile data

CREATE TABLE IF NOT EXISTS pinterest_users
(
    record_id String CODEC(ZSTD(1)),
    user_id String CODEC(ZSTD(1)),
    username String CODEC(ZSTD(1)),
    about String CODEC(ZSTD(1)),
    profile_image String CODEC(ZSTD(1)),
    website_url String CODEC(ZSTD(1)),
    business_name String CODEC(ZSTD(1)),
    board_count Int64 CODEC(T64, ZSTD(1)),
    pin_count Int64 CODEC(T64, ZSTD(1)),
    account_type String CODEC(ZSTD(1)),
    follower_count Int64 CODEC(T64, ZSTD(1)),
    following_count Int64 CODEC(T64, ZSTD(1)),
    monthly_views Int64 CODEC(T64, ZSTD(1)),
    inserted_at DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PRIMARY KEY (user_id, inserted_at)
ORDER BY (user_id, inserted_at)
PARTITION BY toYYYYMM(inserted_at);
