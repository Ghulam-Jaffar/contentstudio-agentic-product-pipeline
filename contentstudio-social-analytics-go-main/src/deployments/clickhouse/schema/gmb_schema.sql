CREATE TABLE contentstudiobackend.gmb_daily_metrics
(
    `gmb_id` String,
    `account_id` String,
    `location_id` String,
    `account_name` String,
    `location_name` String,
    `platform_name` String,
    `inserted_at` DateTime64(6) DEFAULT toStartOfHour(now64(6)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `business_impressions_desktop_maps` Int64,
    `business_impressions_desktop_search` Int64,
    `business_impressions_mobile_maps` Int64,
    `business_impressions_mobile_search` Int64,
    `call_clicks` Int64,
    `website_clicks` Int64,
    `business_direction_requests` Int64,
    `business_conversations` Int64,
    `business_bookings` Int64,
    `business_food_orders` Int64,
    `business_food_menu_clicks` Int64
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (gmb_id, created_at)
ORDER BY (gmb_id, created_at)
SETTINGS index_granularity = 8192;

CREATE TABLE contentstudiobackend.gmb_media_assets
(
    `gmb_id` String,
    `account_id` String,
    `location_id` String,
    `account_name` String,
    `location_name` String,
    `platform_name` String,
    `language_code` String,
    `inserted_at` DateTime64(6) DEFAULT toStartOfHour(now64(6)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `media_name` String,
    `source_url` String,
    `media_format` String,
    `location_association_category` String,
    `google_url` String,
    `thumbnail_url` String,
    `width_pixels` Int64,
    `height_pixels` Int64
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (gmb_id, media_name)
ORDER BY (gmb_id, media_name)
SETTINGS index_granularity = 8192;

CREATE TABLE contentstudiobackend.gmb_search_keywords_monthly
(
    `gmb_id` String,
    `account_id` String,
    `location_id` String,
    `account_name` String,
    `location_name` String,
    `platform_name` String,
    `inserted_at` DateTime64(6) DEFAULT toStartOfHour(now64(6)),
    `keyword_month` DateTime64(6) DEFAULT '0000000000.000000',
    `keyword` String,
    `impressions_value` Int64,
    `impressions_threshold` Int64
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(keyword_month)
PRIMARY KEY (gmb_id, keyword_month, keyword)
ORDER BY (gmb_id, keyword_month, keyword)
SETTINGS index_granularity = 8192;

CREATE TABLE contentstudiobackend.gmb_local_posts
(
    `gmb_id` String,
    `account_id` String,
    `location_id` String,
    `account_name` String,
    `location_name` String,
    `platform_name` String,
    `language_code` String,
    `inserted_at` DateTime64(6) DEFAULT toStartOfHour(now64(6)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_at` DateTime64(6) DEFAULT '0000000000.000000',
    `post_name` String,
    `summary` String DEFAULT '',
    `state` String,
    `topic_type` String,
    `search_url` String,
    `media_names` Array(String),
    `media_formats` Array(String),
    `media_google_urls` Array(String)
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (gmb_id, post_name)
ORDER BY (gmb_id, post_name)
SETTINGS index_granularity = 8192;

CREATE TABLE contentstudiobackend.gmb_reviews
(
    `gmb_id` String,
    `account_id` String,
    `location_id` String,
    `account_name` String,
    `location_name` String,
    `platform_name` String,
    `inserted_at` DateTime64(6) DEFAULT toStartOfHour(now64(6)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_at` DateTime64(6) DEFAULT '0000000000.000000',
    `review_id` String,
    `review_name` String,
    `reviewer_display_name` String,
    `reviewer_profile_photo_url` String,
    `star_rating` Int32,
    `comment` String,
    `reply_comment` String,
    `reply_update_time` DateTime64(6) DEFAULT '0000000000.000000'
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (gmb_id, review_id)
ORDER BY (gmb_id, review_id)
SETTINGS index_granularity = 8192;
