-- Materialized Views Schema

-- Base Table for Daily Followers
CREATE TABLE contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` UInt64,
    `updated_at` DateTime
)
ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, platform, account_id)
SETTINGS index_granularity = 8192;

-- Facebook Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_facebook TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int64,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(saving_time) AS date,
    CAST('facebook', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    page_id AS account_id,
    argMax(page_follows, saving_time) AS followers_count,
    max(saving_time) AS updated_at
FROM contentstudiobackend.facebook_insights
WHERE page_follows > 0
GROUP BY
    toDate(saving_time),
    page_id;

-- Instagram Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_instagram TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int64,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(stored_event_at) AS date,
    CAST('instagram', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    instagram_id AS account_id,
    argMax(followers_count, stored_event_at) AS followers_count,
    max(stored_event_at) AS updated_at
FROM contentstudiobackend.instagram_insights
GROUP BY
    toDate(stored_event_at),
    instagram_id
HAVING followers_count > 0;

-- LinkedIn Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_linkedin TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int64,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(inserted_at) AS date,
    CAST('linkedin', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    linkedin_id AS account_id,
    argMax(totalFollowerCount, inserted_at) AS followers_count,
    max(inserted_at) AS updated_at
FROM contentstudiobackend.linkedin_insights
WHERE totalFollowerCount > 0
GROUP BY
    toDate(inserted_at),
    linkedin_id;

-- Pinterest Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_pinterest TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int64,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(inserted_at) AS date,
    CAST('pinterest', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    board_id AS account_id,
    argMax(follower_count, inserted_at) AS followers_count,
    max(inserted_at) AS updated_at
FROM contentstudiobackend.pinterest_boards
WHERE follower_count > 0
GROUP BY
    toDate(inserted_at),
    board_id;

-- TikTok Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_tiktok TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int32,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(inserted_at) AS date,
    CAST('tiktok', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    tiktok_id AS account_id,
    argMax(total_follower_count, inserted_at) AS followers_count,
    max(inserted_at) AS updated_at
FROM contentstudiobackend.tiktok_insights
WHERE total_follower_count > 0
GROUP BY
    toDate(inserted_at),
    tiktok_id;

-- Twitter Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_twitter TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int64,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(saving_time) AS date,
    CAST('twitter', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    twitter_id AS account_id,
    argMax(followers_count, saving_time) AS followers_count,
    max(saving_time) AS updated_at
FROM contentstudiobackend.twitter_insights
GROUP BY
    toDate(saving_time),
    twitter_id
HAVING followers_count > 0;

-- YouTube Daily Followers Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_followers_youtube TO contentstudiobackend.mv_social_daily_followers
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `followers_count` Int32,
    `updated_at` DateTime64(6)
)
AS SELECT
    toDate(inserted_at) AS date,
    CAST('youtube', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    channel_id AS account_id,
    argMax(subscriber_count, inserted_at) AS followers_count,
    max(inserted_at) AS updated_at
FROM contentstudiobackend.youtube_channels
WHERE subscriber_count > 0
GROUP BY
    toDate(inserted_at),
    channel_id;

-- Base Table for Daily Metrics
CREATE TABLE contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, platform, account_id)
SETTINGS index_granularity = 8192;

-- Facebook Daily Metrics Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_metrics_facebook TO contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
AS WITH latest_posts AS
    (
        SELECT
            post_id,
            page_id,
            argMax(total, saving_time) AS total_reactions,
            argMax(comments, saving_time) AS total_comments,
            argMax(shares, saving_time) AS total_shares,
            argMax(post_impressions_unique, saving_time) AS reach,
            argMax(post_impressions, saving_time) AS impressions,
            toDate(created_time) AS post_date
        FROM contentstudiobackend.facebook_posts
        GROUP BY
            post_id,
            page_id,
            toDate(created_time)
    )
SELECT
    post_date AS date,
    CAST('facebook', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    page_id AS account_id,
    sumState(toUInt64((total_reactions + total_comments) + total_shares)) AS engagement_sum,
    sumState(toUInt64(reach)) AS reach_sum,
    sumState(toUInt64(impressions)) AS impressions_sum,
    uniqState(post_id) AS posts_count,
    sumState(toUInt64(total_reactions)) AS reactions_sum,
    sumState(toUInt64(total_comments)) AS comments_sum,
    sumState(toUInt64(total_shares)) AS shares_sum,
    sumState(toUInt64(0)) AS saves_sum,
    sumState(toUInt64(0)) AS video_views_sum
FROM latest_posts
GROUP BY
    post_date,
    page_id;

-- Instagram Daily Metrics Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_metrics_instagram TO contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
AS WITH latest_posts AS
    (
        SELECT
            media_id AS post_id,
            instagram_id,
            argMax(engagement, stored_event_at) AS total_engagement,
            argMax(reach, stored_event_at) AS total_reach,
            argMax(views, stored_event_at) AS total_views,
            argMax(like_count, stored_event_at) AS likes,
            argMax(comments_count, stored_event_at) AS comments,
            argMax(saved, stored_event_at) AS saves,
            argMax(shares, stored_event_at) AS shares,
            toDate(post_created_at) AS post_date
        FROM contentstudiobackend.instagram_posts
        GROUP BY
            media_id,
            instagram_id,
            toDate(post_created_at)
    )
SELECT
    post_date AS date,
    CAST('instagram', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    instagram_id AS account_id,
    sumState(toUInt64(total_engagement)) AS engagement_sum,
    sumState(toUInt64(total_reach)) AS reach_sum,
    sumState(toUInt64(total_views)) AS impressions_sum,
    uniqState(post_id) AS posts_count,
    sumState(toUInt64(likes)) AS reactions_sum,
    sumState(toUInt64(comments)) AS comments_sum,
    sumState(toUInt64(shares)) AS shares_sum,
    sumState(toUInt64(saves)) AS saves_sum,
    sumState(toUInt64(total_views)) AS video_views_sum
FROM latest_posts
GROUP BY
    post_date,
    instagram_id;

-- LinkedIn Daily Metrics Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_metrics_linkedin TO contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
AS WITH latest_posts AS
    (
        SELECT
            post_id,
            linkedin_id,
            argMax(total_engagement, saving_time) AS total_engagement,
            argMax(reach, saving_time) AS reach,
            argMax(impressions, saving_time) AS impressions,
            argMax(favorites, saving_time) AS favorites,
            argMax(comments, saving_time) AS comments,
            argMax(repost, saving_time) AS repost,
            toDate(created_at) AS post_date
        FROM contentstudiobackend.linkedin_posts
        GROUP BY
            post_id,
            linkedin_id,
            toDate(created_at)
    )
SELECT
    post_date AS date,
    CAST('linkedin', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    linkedin_id AS account_id,
    sumState(toUInt64(total_engagement)) AS engagement_sum,
    sumState(toUInt64(reach)) AS reach_sum,
    sumState(toUInt64(impressions)) AS impressions_sum,
    uniqState(post_id) AS posts_count,
    sumState(toUInt64(favorites)) AS reactions_sum,
    sumState(toUInt64(comments)) AS comments_sum,
    sumState(toUInt64(repost)) AS shares_sum,
    sumState(toUInt64(0)) AS saves_sum,
    sumState(toUInt64(0)) AS video_views_sum
FROM latest_posts
GROUP BY
    post_date,
    linkedin_id;

-- Pinterest Daily Metrics Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_metrics_pinterest TO contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
AS WITH pin_metrics AS
    (
        SELECT
            p.pin_id,
            p.board_id,
            p.user_id,
            toDate(p.created_at) AS post_date,
            argMax(coalesce(pi.engagement, 0), pi.created_at) AS engagement,
            argMax(coalesce(pi.impression, 0), pi.created_at) AS impressions,
            argMax(coalesce(pi.saves, 0), pi.created_at) AS saves,
            argMax(coalesce(pi.pin_clicks, 0), pi.created_at) AS clicks,
            argMax(coalesce(pi.outbound_click, 0), pi.created_at) AS outbound_clicks,
            argMax(coalesce(pi.video_start, 0), pi.created_at) AS video_starts
        FROM contentstudiobackend.pinterest_pins AS p
        LEFT JOIN contentstudiobackend.pinterest_pin_insights AS pi ON (p.pin_id = pi.pin_id) AND (p.user_id = pi.user_id)
        GROUP BY
            p.pin_id,
            p.board_id,
            p.user_id,
            toDate(p.created_at)
    )
SELECT
    post_date AS date,
    CAST('pinterest', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    board_id AS account_id,
    sumState(toUInt64(engagement)) AS engagement_sum,
    sumState(toUInt64(impressions)) AS reach_sum,
    sumState(toUInt64(impressions)) AS impressions_sum,
    uniqState(pin_id) AS posts_count,
    sumState(toUInt64(clicks)) AS reactions_sum,
    sumState(toUInt64(0)) AS comments_sum,
    sumState(toUInt64(saves)) AS shares_sum,
    sumState(toUInt64(saves)) AS saves_sum,
    sumState(toUInt64(video_starts)) AS video_views_sum
FROM pin_metrics
GROUP BY
    post_date,
    board_id;

-- TikTok Daily Metrics Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_metrics_tiktok TO contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
AS WITH latest_posts AS
    (
        SELECT
            post_id,
            tiktok_id,
            argMax(engagement_count, inserted_at) AS engagement,
            argMax(view_count, inserted_at) AS views,
            argMax(like_count, inserted_at) AS likes,
            argMax(comments_count, inserted_at) AS comments,
            argMax(share_count, inserted_at) AS shares,
            toDate(created_at) AS post_date
        FROM contentstudiobackend.tiktok_posts
        GROUP BY
            post_id,
            tiktok_id,
            toDate(created_at)
    )
SELECT
    post_date AS date,
    CAST('tiktok', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    tiktok_id AS account_id,
    sumState(toUInt64(engagement)) AS engagement_sum,
    sumState(toUInt64(views)) AS reach_sum,
    sumState(toUInt64(views)) AS impressions_sum,
    uniqState(post_id) AS posts_count,
    sumState(toUInt64(likes)) AS reactions_sum,
    sumState(toUInt64(comments)) AS comments_sum,
    sumState(toUInt64(shares)) AS shares_sum,
    sumState(toUInt64(0)) AS saves_sum,
    sumState(toUInt64(views)) AS video_views_sum
FROM latest_posts
GROUP BY
    post_date,
    tiktok_id;

-- YouTube Daily Metrics Materialized View
CREATE MATERIALIZED VIEW contentstudiobackend.mv_social_daily_metrics_youtube TO contentstudiobackend.mv_social_daily_metrics
(
    `date` Date,
    `platform` Enum8('facebook' = 1, 'instagram' = 2, 'linkedin' = 3, 'tiktok' = 4, 'pinterest' = 5, 'youtube' = 6, 'twitter' = 7, 'threads' = 8, 'gmb' = 9, 'bluesky' = 10),
    `account_id` String,
    `engagement_sum` AggregateFunction(sum, UInt64),
    `reach_sum` AggregateFunction(sum, UInt64),
    `impressions_sum` AggregateFunction(sum, UInt64),
    `posts_count` AggregateFunction(uniq, String),
    `reactions_sum` AggregateFunction(sum, UInt64),
    `comments_sum` AggregateFunction(sum, UInt64),
    `shares_sum` AggregateFunction(sum, UInt64),
    `saves_sum` AggregateFunction(sum, UInt64),
    `video_views_sum` AggregateFunction(sum, UInt64)
)
AS WITH latest_videos AS
    (
        SELECT
            video_id,
            channel_id,
            argMax(likes, inserted_at) AS total_likes,
            argMax(dislikes, inserted_at) AS total_dislikes,
            argMax(comments, inserted_at) AS total_comments,
            argMax(shares, inserted_at) AS total_shares,
            argMax(views, inserted_at) AS total_views,
            toDate(published_at) AS post_date
        FROM contentstudiobackend.youtube_videos
        GROUP BY
            video_id,
            channel_id,
            toDate(published_at)
    )
SELECT
    post_date AS date,
    CAST('youtube', 'Enum8(\'facebook\' = 1, \'instagram\' = 2, \'linkedin\' = 3, \'tiktok\' = 4, \'pinterest\' = 5, \'youtube\' = 6, \'twitter\' = 7, \'threads\' = 8, \'gmb\' = 9, \'bluesky\' = 10)') AS platform,
    channel_id AS account_id,
    sumState(toUInt64(((total_likes + total_dislikes) + total_comments) + total_shares)) AS engagement_sum,
    sumState(toUInt64(total_views)) AS reach_sum,
    sumState(toUInt64(total_views)) AS impressions_sum,
    uniqState(video_id) AS posts_count,
    sumState(toUInt64(total_likes)) AS reactions_sum,
    sumState(toUInt64(total_comments)) AS comments_sum,
    sumState(toUInt64(total_shares)) AS shares_sum,
    sumState(toUInt64(0)) AS saves_sum,
    sumState(toUInt64(total_views)) AS video_views_sum
FROM latest_videos
GROUP BY
    post_date,
    channel_id;