-- Pinterest Tables Schema

-- Pinterest Users Table
CREATE TABLE contentstudiobackend.pinterest_users
(
    `user_id` String CODEC(ZSTD(1)),
    `profile_image` String CODEC(ZSTD(1)),
    `website_url` String CODEC(ZSTD(1)),
    `username` String CODEC(ZSTD(1)),
    `about` String CODEC(ZSTD(1)),
    `business_name` String CODEC(ZSTD(1)),
    `board_count` Int64 CODEC(T64, ZSTD(1)),
    `pin_count` Int64 CODEC(T64, ZSTD(1)),
    `account_type` String CODEC(ZSTD(1)),
    `follower_count` Int64 CODEC(T64, ZSTD(1)),
    `following_count` Int64 CODEC(T64, ZSTD(1)),
    `monthly_views` Int64 CODEC(T64, ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYear(inserted_at)
PRIMARY KEY user_id
ORDER BY user_id
SETTINGS index_granularity = 8192;

-- Pinterest Boards Table
CREATE TABLE contentstudiobackend.pinterest_boards
(
    `record_id` String CODEC(ZSTD(1)),
    `user_id` String CODEC(ZSTD(1)),
    `board_id` String CODEC(ZSTD(1)),
    `name` String CODEC(ZSTD(1)),
    `owner` String CODEC(ZSTD(1)),
    `description` String CODEC(ZSTD(1)),
    `privacy` String CODEC(ZSTD(1)),
    `image_cover_url` String CODEC(ZSTD(1)),
    `pin_thumbnail_urls` Array(String) CODEC(ZSTD(1)),
    `collaborator_count` Int64 CODEC(T64, ZSTD(1)),
    `pin_count` Int64 CODEC(T64, ZSTD(1)),
    `follower_count` Int64 CODEC(T64, ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY record_id
ORDER BY (record_id, created_at)
SETTINGS index_granularity = 8192;

-- Pinterest Pins Table
CREATE TABLE contentstudiobackend.pinterest_pins
(
    `pin_id` String CODEC(ZSTD(1)),
    `board_id` String CODEC(ZSTD(1)),
    `user_id` String CODEC(ZSTD(1)),
    `title` String CODEC(ZSTD(1)),
    `note` String CODEC(ZSTD(1)),
    `parent_pin_id` String CODEC(ZSTD(1)),
    `board_section_id` String CODEC(ZSTD(1)),
    `description` String CODEC(ZSTD(1)),
    `board_owner` String CODEC(ZSTD(1)),
    `media_type` String CODEC(ZSTD(1)),
    `cover_image_url` String CODEC(ZSTD(1)),
    `video_url` String CODEC(ZSTD(1)),
    `duration` String CODEC(ZSTD(1)),
    `height` String CODEC(ZSTD(1)),
    `width` String CODEC(ZSTD(1)),
    `dominant_color` String CODEC(ZSTD(1)),
    `product_tags` Array(String) CODEC(ZSTD(1)),
    `creative_type` String CODEC(ZSTD(1)),
    `is_standard` Int64 CODEC(T64, ZSTD(1)),
    `is_owner` Int64 CODEC(T64, ZSTD(1)),
    `has_been_promoted` Int64 CODEC(T64, ZSTD(1)),
    `hour_of_day` Int64 CODEC(T64, ZSTD(1)),
    `day_of_week` String CODEC(ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (pin_id, created_at)
ORDER BY (pin_id, created_at)
SETTINGS index_granularity = 8192;

-- Pinterest Pin Insights Table
CREATE TABLE contentstudiobackend.pinterest_pin_insights
(
    `record_id` String CODEC(ZSTD(1)),
    `user_id` String CODEC(ZSTD(1)),
    `pin_id` String CODEC(ZSTD(1)),
    `pin_clicks` Int64 CODEC(T64, ZSTD(1)),
    `video_mrc_view` Int64 CODEC(T64, ZSTD(1)),
    `full_screen_play` Int64 CODEC(T64, ZSTD(1)),
    `outbound_click` Int64 CODEC(T64, ZSTD(1)),
    `video_v50_watch_time` Int64 CODEC(T64, ZSTD(1)),
    `clickthrough` Int64 CODEC(T64, ZSTD(1)),
    `clickthrough_rate` Int64 CODEC(T64, ZSTD(1)),
    `engagement` Int64 CODEC(T64, ZSTD(1)),
    `engagement_rate` Int64 CODEC(T64, ZSTD(1)),
    `video_start` Int64 CODEC(T64, ZSTD(1)),
    `profile_visit` Int64 CODEC(T64, ZSTD(1)),
    `closeup` Int64 CODEC(T64, ZSTD(1)),
    `full_screen_playtime` Int64 CODEC(T64, ZSTD(1)),
    `video_avg_watch_time` Int64 CODEC(T64, ZSTD(1)),
    `video_10s_view` Int64 CODEC(T64, ZSTD(1)),
    `quartile_95s_percent_view` Int64 CODEC(T64, ZSTD(1)),
    `user_follow` Int64 CODEC(T64, ZSTD(1)),
    `impression` Int64 CODEC(T64, ZSTD(1)),
    `saves` Int64 CODEC(T64, ZSTD(1)),
    `save_rate` Int64 CODEC(T64, ZSTD(1)),
    `data_status` String CODEC(ZSTD(1)),
    `day_of_week` String CODEC(ZSTD(1)),
    `hour_of_day` Int64 CODEC(T64, ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY record_id
ORDER BY (record_id, pin_id, created_at)
SETTINGS index_granularity = 8192;

-- Pinterest User Insights Table
CREATE TABLE contentstudiobackend.pinterest_user_insights
(
    `record_id` String CODEC(ZSTD(1)),
    `user_id` String CODEC(ZSTD(1)),
    `pin_clicks` Int64 CODEC(T64, ZSTD(1)),
    `pin_click_rate` Int64 CODEC(T64, ZSTD(1)),
    `video_mrc_view` Int64 CODEC(T64, ZSTD(1)),
    `full_screen_play` Int64 CODEC(T64, ZSTD(1)),
    `outbound_click` Int64 CODEC(T64, ZSTD(1)),
    `video_v50_watch_time` Int64 CODEC(T64, ZSTD(1)),
    `clickthrough` Int64 CODEC(T64, ZSTD(1)),
    `clickthrough_rate` Int64 CODEC(T64, ZSTD(1)),
    `engagement` Int64 CODEC(T64, ZSTD(1)),
    `engagement_rate` Int64 CODEC(T64, ZSTD(1)),
    `video_start` Int64 CODEC(T64, ZSTD(1)),
    `profile_visit` Int64 CODEC(T64, ZSTD(1)),
    `closeup` Int64 CODEC(T64, ZSTD(1)),
    `full_screen_playtime` Int64 CODEC(T64, ZSTD(1)),
    `video_avg_watch_time` Int64 CODEC(T64, ZSTD(1)),
    `video_10s_view` Int64 CODEC(T64, ZSTD(1)),
    `quartile_95s_percent_view` Int64 CODEC(T64, ZSTD(1)),
    `impression` Int64 CODEC(T64, ZSTD(1)),
    `saves` Int64 CODEC(T64, ZSTD(1)),
    `save_rate` Int64 CODEC(T64, ZSTD(1)),
    `data_status` String CODEC(ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY record_id
ORDER BY (record_id, created_at)
SETTINGS index_granularity = 8192;