-- YouTube Tables Schema

-- YouTube Channels Table
CREATE TABLE contentstudiobackend.youtube_channels
(
    `record_id` String CODEC(ZSTD(1)),
    `channel_id` String CODEC(ZSTD(1)),
    `title` String CODEC(ZSTD(1)),
    `description` String CODEC(ZSTD(1)),
    `custom_url` String CODEC(ZSTD(1)),
    `thumbnail_url` String CODEC(ZSTD(1)),
    `external_banner_url` String CODEC(ZSTD(1)),
    `country` String CODEC(ZSTD(1)),
    `subscriber_count` Int64 CODEC(T64, ZSTD(1)),
    `video_count` Int64 CODEC(T64, ZSTD(1)),
    `view_count` Int64 CODEC(T64, ZSTD(1)),
    `published_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY record_id
ORDER BY (record_id, channel_id)
SETTINGS index_granularity = 8192;

-- YouTube Videos Table
CREATE TABLE contentstudiobackend.youtube_videos
(
    `video_id` String CODEC(ZSTD(1)),
    `channel_id` String CODEC(ZSTD(1)),
    `title` String CODEC(ZSTD(1)),
    `description` String CODEC(ZSTD(1)),
    `duration` String CODEC(ZSTD(1)),
    `thumbnail_url` String CODEC(ZSTD(1)),
    `iframe_embed_html` String CODEC(ZSTD(1)),
    `likes` Int64 CODEC(T64, ZSTD(1)),
    `dislikes` Int64 CODEC(T64, ZSTD(1)),
    `views` Int64 CODEC(T64, ZSTD(1)),
    `comments` Int64 CODEC(T64, ZSTD(1)),
    `shares` Int64 CODEC(T64, ZSTD(1)),
    `favorites` Int64 CODEC(T64, ZSTD(1)),
    `saved` Int64 CODEC(T64, ZSTD(1)),
    `subscribers_gained` Int64 CODEC(T64, ZSTD(1)),
    `red_views` Int64 CODEC(T64, ZSTD(1)),
    `minutes_watched` Int64 CODEC(T64, ZSTD(1)),
    `red_minutes_watched` Int64 CODEC(T64, ZSTD(1)),
    `average_view_duration` Int64 CODEC(T64, ZSTD(1)),
    `average_view_percentage` Float64 CODEC(ZSTD(1)),
    `impressions` Int64 CODEC(T64, ZSTD(1)),
    `impressions_click_through_rate` Float64 CODEC(ZSTD(1)),
    `published_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `media_type` String CODEC(ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(published_at)
PRIMARY KEY video_id
ORDER BY (video_id, channel_id)
SETTINGS index_granularity = 8192;

-- YouTube Activity Insights Table
CREATE TABLE contentstudiobackend.youtube_activity_insights
(
    `record_id` String CODEC(ZSTD(1)),
    `channel_id` String CODEC(ZSTD(1)),
    `red_views` Int64 CODEC(T64, ZSTD(1)),
    `views` Int64 CODEC(T64, ZSTD(1)),
    `likes` Int64 CODEC(T64, ZSTD(1)),
    `dislikes` Int64 CODEC(T64, ZSTD(1)),
    `comments` Int64 CODEC(T64, ZSTD(1)),
    `shares` Int64 CODEC(T64, ZSTD(1)),
    `subscribers_gained` Int64 CODEC(T64, ZSTD(1)),
    `estimated_minutes_watched` Int64 CODEC(T64, ZSTD(1)),
    `estimated_red_minutes_watched` Int64 CODEC(T64, ZSTD(1)),
    `average_view_duration` Int64 CODEC(T64, ZSTD(1)),
    `average_view_percentage` Float64 CODEC(ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT now64(6) CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY record_id
ORDER BY (record_id, created_at)
SETTINGS index_granularity = 8192;

-- YouTube Traffic Insights Table
CREATE TABLE contentstudiobackend.youtube_traffic_insights
(
    `record_id` String CODEC(ZSTD(1)),
    `channel_id` String CODEC(ZSTD(1)),
    `paid_views` Int64 CODEC(T64, ZSTD(1)),
    `annotation_views` Int64 CODEC(T64, ZSTD(1)),
    `end_screen_views` Int64 CODEC(T64, ZSTD(1)),
    `campaign_card_view` Int64 CODEC(T64, ZSTD(1)),
    `subscriber_views` Int64 CODEC(T64, ZSTD(1)),
    `no_link_other_views` Int64 CODEC(T64, ZSTD(1)),
    `yt_channel_views` Int64 CODEC(T64, ZSTD(1)),
    `yt_search_views` Int64 CODEC(T64, ZSTD(1)),
    `related_video_views` Int64 CODEC(T64, ZSTD(1)),
    `yt_other_page_views` Int64 CODEC(T64, ZSTD(1)),
    `ext_url_views` Int64 CODEC(T64, ZSTD(1)),
    `playlist_views` Int64 CODEC(T64, ZSTD(1)),
    `notification_views` Int64 CODEC(T64, ZSTD(1)),
    `subscriber_watch_time` Int64 CODEC(T64, ZSTD(1)),
    `non_subsciber_watch_time` Int64 CODEC(T64, ZSTD(1)),
    `shorts_views` Int64 CODEC(T64, ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY record_id
ORDER BY (record_id, channel_id, created_at)
SETTINGS index_granularity = 8192;

-- YouTube Shared Insights Table
CREATE TABLE contentstudiobackend.youtube_shared_insights
(
    `record_id` String CODEC(ZSTD(1)),
    `channel_id` String CODEC(ZSTD(1)),
    `ameba` Int64 CODEC(T64, ZSTD(1)),
    `blogger` Int64 CODEC(T64, ZSTD(1)),
    `copy_paste` Int64 CODEC(T64, ZSTD(1)),
    `cyworld` Int64 CODEC(T64, ZSTD(1)),
    `digg` Int64 CODEC(T64, ZSTD(1)),
    `dropbox` Int64 CODEC(T64, ZSTD(1)),
    `embed` Int64 CODEC(T64, ZSTD(1)),
    `mail` Int64 CODEC(T64, ZSTD(1)),
    `whats_app` Int64 CODEC(T64, ZSTD(1)),
    `other` Int64 CODEC(T64, ZSTD(1)),
    `facebook_messenger` Int64 CODEC(T64, ZSTD(1)),
    `facebook_pages` Int64 CODEC(T64, ZSTD(1)),
    `facebook` Int64 CODEC(T64, ZSTD(1)),
    `fotka` Int64 CODEC(T64, ZSTD(1)),
    `vkontakte` Int64 CODEC(T64, ZSTD(1)),
    `discord` Int64 CODEC(T64, ZSTD(1)),
    `google_plus` Int64 CODEC(T64, ZSTD(1)),
    `goo` Int64 CODEC(T64, ZSTD(1)),
    `hangouts` Int64 CODEC(T64, ZSTD(1)),
    `linkedin` Int64 CODEC(T64, ZSTD(1)),
    `pinterest` Int64 CODEC(T64, ZSTD(1)),
    `myspace` Int64 CODEC(T64, ZSTD(1)),
    `reddit` Int64 CODEC(T64, ZSTD(1)),
    `skype` Int64 CODEC(T64, ZSTD(1)),
    `telegram` Int64 CODEC(T64, ZSTD(1)),
    `twitter` Int64 CODEC(T64, ZSTD(1)),
    `tumblr` Int64 CODEC(T64, ZSTD(1)),
    `viber` Int64 CODEC(T64, ZSTD(1)),
    `weibo` Int64 CODEC(T64, ZSTD(1)),
    `wechat` Int64 CODEC(T64, ZSTD(1)),
    `youtube` Int64 CODEC(T64, ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY record_id
ORDER BY (record_id, channel_id)
SETTINGS index_granularity = 8192;
