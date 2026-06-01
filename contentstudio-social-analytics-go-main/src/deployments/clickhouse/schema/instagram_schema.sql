-- Instagram Tables Schema

-- Instagram Insights Table
CREATE TABLE contentstudiobackend.instagram_insights
(
    `instagram_id` String,
    `record_id` String,
    `username` String,
    `name` String,
    `profile_picture_url` String,
    `follows_count` Int64,
    `followers_count` Int64,
    `media_count` Int64,
    `tags` Int64,
    `impressions` Int64,
    `profile_views` Int64,
    `online_followers` Array(String),
    `audience_city` Array(String),
    `audience_country` Array(String),
    `audience_age` Array(String),
    `audience_gender` Array(String),
    `audience_gender_age` Array(String),
    `audience_locale` Array(String),
    `audience_datetime` DateTime64(6) DEFAULT '0000000000.000000',
    `day_of_week` String,
    `hour_of_day` Int64,
    `year` Int64,
    `month` Int64,
    `online_users_datetime` DateTime64(6) DEFAULT '0000000000.000000',
    `created_time` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_time` DateTime64(6) DEFAULT '0000000000.000000',
    `metadata` Map(String, String),
    `stored_event_at` DateTime64(6) DEFAULT '0000000000.000000',
    `audience_city_by_engagement` Array(String),
    `audience_city_by_reach` Array(String),
    `audience_country_by_engagement` Array(String),
    `audience_country_by_reach` Array(String),
    `audience_age_by_engagement` Array(String),
    `audience_age_by_reach` Array(String),
    `audience_gender_by_engagement` Array(String),
    `audience_gender_by_reach` Array(String),
    `audience_gender_age_by_engagement` Array(String),
    `audience_gender_age_by_reach` Array(String),
    `shares` Int64,
    `reach` Int64,
    `accounts_engaged` Int64,
    `likes` Int64,
    `comments` Int64,
    `saves` Int64,
    `engagement` Int64,
    `views` Int64 CODEC(Delta(8), ZSTD(1))
)
ENGINE = ReplacingMergeTree(updated_time)
PARTITION BY if(position(record_id, '_') > 0, toYYYYMM(toDate(substring(record_id, position(record_id, '_') + 1))), toYYYYMM(stored_event_at))
PRIMARY KEY (instagram_id, record_id)
ORDER BY (instagram_id, record_id)
SETTINGS index_granularity = 8192;

-- Instagram Posts Table
CREATE TABLE contentstudiobackend.instagram_posts
(
    `instagram_id` String,
    `media_id` String,
    `username` String,
    `name` String,
    `profile_picture_url` String,
    `permalink` String,
    `like_count` Int64,
    `comments_count` Int64,
    `engagement` Int64,
    `impressions` Int64,
    `reach` Int64,
    `saved` Int64,
    `video_views` Int64,
    `exits` Int64,
    `replies` Int64,
    `taps_forward` Int64,
    `taps_back` Int64,
    `child_assets_type` Array(String),
    `caption` String,
    `media_type` String,
    `entity_type` String,
    `media_url` Array(String),
    `video_url` Array(String),
    `hashtags` Array(String),
    `day_of_week` String,
    `hour_of_day` Int64,
    `year` Int64,
    `month` Int64,
    `timestamp` Int64,
    `stored_event_at` DateTime64(6) DEFAULT '0000000000.000000',
    `post_created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `shares` Int64,
    `reels_avg_watch_time` Int64,
    `reels_total_watch_time` Int64,
    `views` Int64 CODEC(Delta(8), ZSTD(1)),
    `updated_time` DateTime64(6) DEFAULT '0000000000.000000'
)
ENGINE = ReplacingMergeTree(updated_time)
PARTITION BY toYYYYMM(post_created_at)
PRIMARY KEY (instagram_id, media_id)
ORDER BY (instagram_id, media_id, post_created_at)
SETTINGS index_granularity = 8192;

-- Instagram Competitor Insights Table
CREATE TABLE contentstudiobackend.instagram_competitor_insights
(
    `record_id` String,
    `instagram_account_id` String,
    `total_followed_by_count` Int64,
    `total_following_count` Int64,
    `profile_picture_url` String,
    `page_name` String,
    `metadata` Map(String, String) DEFAULT map(),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000'
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY record_id
ORDER BY record_id
SETTINGS index_granularity = 8192;

-- Instagram Competitor Posts Table
CREATE TABLE contentstudiobackend.instagram_competitor_posts
(
    `instagram_id` Int64,
    `post_id` String,
    `business_account_id` String,
    `total_followed_by_count` Int64,
    `total_following_count` Int64,
    `username` String,
    `name` String,
    `page_category` String,
    `profile_picture_url` String,
    `biography` String,
    `engagement` Int64,
    `like_count` Int64,
    `comments_count` Int64,
    `media_count` Int64,
    `caption` String,
    `media_type` String,
    `media_product_type` String,
    `media_url` String,
    `permalink` String,
    `hashtags` Array(String),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000'
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY post_id
ORDER BY (post_id, created_at, business_account_id, instagram_id)
SETTINGS index_granularity = 8192;