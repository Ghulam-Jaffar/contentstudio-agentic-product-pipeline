-- Facebook Tables Schema

-- Facebook Insights Table
CREATE TABLE contentstudiobackend.facebook_insights
(
    `hash_id` String,
    `page_id` String,
    `page_category` String,
    `day_of_week` String,
    `year` Int64,
    `month` Int64,
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000',
    `talking_about_count` Int64,
    `page_fans` Int64,
    `page_fans_city` Array(String),
    `page_fans_country` Array(String),
    `page_fans_locale` Array(String),
    `page_fans_age` Array(String),
    `page_fans_gender` Array(String),
    `page_fans_gender_age` Array(String),
    `page_fan_adds_by_paid_non_paid_unique` Array(String),
    `page_fan_adds_unique` Int64,
    `page_fan_removes_unique` Int64,
    `page_fans_by_like_source_unique` Array(String),
    `page_fans_by_unlike_source_unique` Array(String),
    `page_fans_by_like` Int64,
    `page_fans_by_unlike` Int64,
    `page_total_actions` Int64,
    `page_post_engagements` Int64,
    `page_impressions` Int64,
    `page_impressions_unique` Int64,
    `page_impressions_organic` Int64,
    `page_impressions_paid` Int64,
    `page_video_views_paid` Int64,
    `page_video_views` Int64,
    `page_video_views_organic` Int64,
    `page_video_views_autoplayed` Int64,
    `page_video_views_click_to_play` Int64,
    `page_video_repeat_views` Int64,
    `page_negative_feedback` Int64,
    `page_positive_feedback` Int64,
    `page_negative_feedback_by_type` Array(String),
    `page_positive_feedback_by_type` Array(String),
    `page_fans_online` Array(String),
    `active_users` Int64,
    `positive_sentiment` Int64,
    `negative_sentiment` Int64,
    `posts_count` Int64,
    `likes_count` Int64,
    `type_count` Array(String),
    `message_count` Array(String),
    `prime_time` DateTime64(6) DEFAULT '0000000000.000000',
    `page_follows` Int64,
    `page_views` Int64,
    `created_time` DateTime DEFAULT now(),
    `updated_at` DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY (year * 100) + month
PRIMARY KEY (page_id, hash_id)
ORDER BY (page_id, hash_id)
SETTINGS index_granularity = 8192;

-- Facebook Posts Table
CREATE TABLE contentstudiobackend.facebook_posts
(
    `page_name` String,
    `page_id` String,
    `post_id` String,
    `permalink` String,
    `status_type` String,
    `media_type` String,
    `video_id` String,
    `category` String,
    `published_by` String,
    `published_by_url` String,
    `shared_from_name` String,
    `shared_from_id` String,
    `shared_from_link` String,
    `like` Int32,
    `love` Int32,
    `haha` Int32,
    `wow` Int32,
    `sad` Int32,
    `angry` Int32,
    `total` Int64,
    `shares` Int32,
    `comments` Int32,
    `post_clicks` Int64,
    `total_engagement` Int64,
    `post_engaged_users` Int64,
    `day_of_week` String,
    `hour_of_day` Int32,
    `created_time` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_time` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_at` DateTime64(3) DEFAULT now64(3),
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000',
    `message_tags` Array(String),
    `post_metadata` String,
    `caption` String,
    `description` String,
    `full_picture` String,
    `link` String,
    `post_impressions` Int64,
    `post_impressions_unique` Int64,
    `post_impressions_paid` Int64,
    `post_impressions_paid_unique` Int64,
    `post_impressions_organic` Int64,
    `post_impressions_organic_unique` Int64,
    `post_impressions_viral` Int64,
    `post_impressions_viral_unique` Int64,
    `post_video_views` Int64,
    `total_impressions` Int64,
    `page_impressions_unique` Int64,
    `thankful` Int64,
    `metadata` Map(String, String) DEFAULT map()
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_time)
PRIMARY KEY (page_id, post_id)
ORDER BY (page_id, post_id)
SETTINGS index_granularity = 8192;

-- Facebook Media Assets Table
CREATE TABLE contentstudiobackend.facebook_media_assets
(
    `page_id` String,
    `post_id` String,
    `media_id` String,
    `asset_type` String,
    `call_to_action` String,

    `CTA_type` String,
    `link` String,
    `caption` String,
    `description` String,
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `inserted_at` DateTime64(6),
    `updated_at` DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (post_id, media_id)
ORDER BY (page_id, post_id, media_id)
SETTINGS index_granularity = 8192;

-- Facebook Video Insights Table
CREATE TABLE contentstudiobackend.facebook_video_insights
(
    `post_id` String,
    `page_id` String,
    `video_id` String,
    `created_time` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_time` DateTime64(6) DEFAULT '0000000000.000000',
    `total_video_followers` Int64,
    `total_video_views` Int64,
    `total_video_views_unique` Int64,
    `total_video_views_autoplayed` Int64,
    `total_video_views_organic` Int64,
    `total_video_views_organic_unique` Int64,
    `total_video_views_paid` Int64,
    `total_video_views_paid_unique` Int64,
    `total_video_views_sound_on` Int64,
    `total_video_views_by_distribution_type` Array(String),
    `total_video_view_time_by_distribution_type` Array(String),
    `total_video_view_time_by_country_id` Array(String),
    `total_video_view_time_by_region_id` Array(String),
    `total_video_view_time_by_age_bucket_and_gender` Array(String),
    `total_video_play_count` Int64,
    `total_video_consumption_rate` Float64,
    `total_video_complete_views` Int64,
    `total_video_complete_views_unique` Int64,
    `total_video_complete_views_autoplayed` Int64,
    `total_video_complete_views_clicked_to_play` Int64,
    `total_video_complete_views_organic` Int64,
    `total_video_complete_views_organic_unique` Int64,
    `total_video_complete_views_paid` Int64,
    `total_video_complete_views_paid_unique` Int64,
    `video_asset_60s_video_view_total_count_by_is_monetizable` Array(String),
    `total_video_15min_excludes_shorter_views` Int64,
    `total_video_15min_excludes_shorter_views_unique` Int64,
    `total_video_60s_excludes_shorter_views` Int64,
    `total_video_30s_views` Int64,
    `total_video_30s_views_unique` Int64,
    `total_video_30s_views_autoplayed` Int64,
    `total_video_30s_views_clicked_to_play` Int64,
    `total_video_30s_views_organic` Int64,
    `total_video_30s_views_paid` Int64,
    `total_video_30s_views_sound_on` Int64,
    `total_video_10s_views` Int64,
    `total_video_10s_views_unique` Int64,
    `total_video_10s_views_autoplayed` Int64,
    `total_video_10s_views_clicked_to_play` Int64,
    `total_video_10s_views_organic` Int64,
    `total_video_10s_views_paid` Int64,
    `total_video_10s_views_sound_on` Int64,
    `total_video_15s_views` Int64,
    `total_video_avg_time_watched` Int64,
    `total_video_view_total_time` Int64,
    `total_video_view_total_time_organic` Int64,
    `total_video_view_total_time_paid` Int64,
    `total_video_retention_graph_autoplayed` Array(String),
    `total_video_retention_graph_clicked_to_play` Array(String),
    `total_video_retention_graph_gender_male` Array(String),
    `total_video_retention_graph_gender_female` Array(String),
    `total_video_impressions` Int64,
    `total_video_impressions_unique` Int64,
    `total_video_impressions_paid_unique` Int64,
    `total_video_impressions_paid` Int64,
    `total_video_impressions_organic_unique` Int64,
    `total_video_impressions_organic` Int64,
    `total_video_impressions_viral_unique` Int64,
    `total_video_impressions_viral` Int64,
    `total_video_impressions_fan_unique` Int64,
    `total_video_impressions_fan` Int64,
    `total_video_impressions_fan_paid_unique` Int64,
    `total_video_impressions_fan_paid` Int64,
    `total_video_stories_by_action_type` Array(String),
    `total_video_reactions_by_type_total` Array(String),
    `total_engagement` Int64,
    `total_video_ad_break_earnings` Float64,
    `total_video_ad_break_ad_impressions` Int64,
    `total_video_ad_break_ad_cpm` Int64,
    `updated_at` DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_time)
PRIMARY KEY (page_id, post_id, video_id)
ORDER BY (page_id, post_id, video_id)
SETTINGS index_granularity = 8192;

-- Facebook Reels Insights Table
CREATE TABLE contentstudiobackend.facebook_reels_insights
(
    `post_id` String CODEC(ZSTD(1)),
    `page_id` String CODEC(ZSTD(1)),
    `average_time_watched` Int64 CODEC(T64, ZSTD(1)),
    `total_time_watched_in_ms` Int64 CODEC(T64, ZSTD(1)),
    `play_count` Int64 CODEC(T64, ZSTD(1)),
    `impressions_unique` Int64 CODEC(T64, ZSTD(1)),
    `reel_followers` Int64 CODEC(T64, ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `updated_at` DateTime64(3) DEFAULT now64(3) CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (page_id, post_id)
ORDER BY (page_id, post_id)
SETTINGS index_granularity = 8192;

-- Facebook Competitor Insights Table
CREATE TABLE contentstudiobackend.facebook_competitor_insights
(
    `record_id` String,
    `page_id` String,
    `followers_count` Int64,
    `total_fan_count` Int64,
    `talking_about_this` Int64,
    `biography` String,
    `profile_picture_url` String,
    `page_name` String,
    `page_category` String,
    `emails` Array(String),
    `birthday` String,
    `were_here_count` Int64,
    `cover_photo_url` String,
    `permalink` String,
    `metadata` Map(String, String) DEFAULT map(),
    `inserted_at` DateTime DEFAULT '0000000000'
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY (record_id, page_id)
ORDER BY (record_id, page_id)
SETTINGS index_granularity = 8192;

-- Facebook Competitor Posts Table
CREATE TABLE contentstudiobackend.facebook_competitor_posts
(
    `facebook_id` String,
    `post_id` String,
    `followers_count` Int64,
    `fan_count` Int64,
    `page_name` String,
    `page_category` String,
    `biography` String,
    `post_engagement` Int64,
    `like` Int64,
    `haha` Int64,
    `angry` Int64,
    `sad` Int64,
    `thankful` Int64,
    `love` Int64,
    `total_post_reactions` Int64,
    `comments` Int64,
    `shares` Int64,
    `caption` String,
    `media_type` String,
    `status_type` String,
    `shared_from_name` String,
    `shared_from_id` String,
    `shared_from_pic` String,
    `shared_created_at` DateTime DEFAULT '0000000000',
    `permalink` String,
    `hashtags` Array(String),
    `day_of_week` String,
    `hour_of_day` Int64,
    `created_at` DateTime DEFAULT '0000000000',
    `inserted_at` DateTime DEFAULT '0000000000',
    `wow` Int64
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (facebook_id, post_id)
ORDER BY (facebook_id, post_id)
SETTINGS index_granularity = 8192;

-- Facebook Competitor Media Assets Table
CREATE TABLE contentstudiobackend.facebook_competitor_media_assets
(
    `media_id` String,
    `post_id` String,
    `page_id` String,
    `caption` String,
    `description` String,
    `link` String,
    `asset_type` String,
    `call_to_action` String,
    `cta_type` String,
    `created_at` DateTime DEFAULT '0000000000',
    `inserted_at` DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (post_id, media_id)
ORDER BY (post_id, media_id)
SETTINGS index_granularity = 8192;