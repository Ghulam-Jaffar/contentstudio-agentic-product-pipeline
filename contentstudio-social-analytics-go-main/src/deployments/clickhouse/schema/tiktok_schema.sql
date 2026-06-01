-- TikTok Tables Schema

-- TikTok Insights Table
CREATE TABLE contentstudiobackend.tiktok_insights
(
    `record_id` String CODEC(ZSTD(1)),
    `tiktok_id` String CODEC(ZSTD(1)),
    `display_name` String CODEC(ZSTD(1)),
    `profile_image` String CODEC(ZSTD(1)),
    `total_follower_count` Int32 CODEC(T64, ZSTD(1)),
    `total_following_count` Int32 CODEC(T64, ZSTD(1)),
    `total_like_count` Int32 CODEC(T64, ZSTD(1)),
    `total_video_count` Int32 CODEC(T64, ZSTD(1)),
    `total_video_views` Int32 CODEC(T64, ZSTD(1)),
    `total_video_likes` Int32 CODEC(T64, ZSTD(1)),
    `total_video_comments` Int32 CODEC(T64, ZSTD(1)),
    `total_video_shares` Int32 CODEC(T64, ZSTD(1)),
    `is_verified` Bool,
    `bio` String CODEC(ZSTD(1)),
    `profile_link` String CODEC(ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY (tiktok_id, record_id)
ORDER BY (tiktok_id, record_id)
SETTINGS index_granularity = 8192;

-- TikTok Posts Table
CREATE TABLE contentstudiobackend.tiktok_posts
(
    `tiktok_id` String CODEC(ZSTD(1)),
    `display_name` String CODEC(ZSTD(1)),
    `profile_link` String CODEC(ZSTD(1)),
    `post_id` String CODEC(ZSTD(1)),
    `cover_image_url` String CODEC(ZSTD(1)),
    `share_url` String CODEC(ZSTD(1)),
    `post_description` String CODEC(ZSTD(1)),
    `hashtags` Array(String),
    `duration` Int32 CODEC(T64, ZSTD(1)),
    `height` Int32 CODEC(T64, ZSTD(1)),
    `width` Int32 CODEC(T64, ZSTD(1)),
    `title` String CODEC(ZSTD(1)),
    `embed_html` String CODEC(ZSTD(1)),
    `embed_link` String CODEC(ZSTD(1)),
    `like_count` Int32 CODEC(T64, ZSTD(1)),
    `comments_count` Int32 CODEC(T64, ZSTD(1)),
    `share_count` Int32 CODEC(T64, ZSTD(1)),
    `view_count` Int32 CODEC(T64, ZSTD(1)),
    `engagement_count` Int32 CODEC(T64, ZSTD(1)),
    `engagement_rate` Float64,
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY post_id
ORDER BY (post_id, created_at)
SETTINGS index_granularity = 8192;