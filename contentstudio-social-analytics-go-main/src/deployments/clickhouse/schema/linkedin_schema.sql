-- LinkedIn Tables Schema

-- LinkedIn Insights Table (daily buckets - one record per day per linkedin_id)
CREATE TABLE contentstudiobackend.linkedin_insights
(
    `linkedin_id` String,
    `record_id` String,
    `impressionCount` Int64,
    `organicFollowerCount` Int64,
    `totalFollowerCount` Int64,
    `paidFollowerCount` Int64,
    `daily_follower_count` Int64 DEFAULT 0, -- Daily follower count for that specific day
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000',
    `created_at` DateTime64(6) DEFAULT '0000000000.000000', -- Date bucket from API timeRange
    `reach` Int64,
    `repost` Int64,
    `comments` Int64,
    `post_clicks` Int64,
    `reactions` Int64 DEFAULT 0,
    `engagement` Float64 DEFAULT 0,
    `followers_by_seniority` String CODEC(ZSTD(1)),
    `followers_by_industry` String CODEC(ZSTD(1)),
    `followers_by_country` String CODEC(ZSTD(1)),
    `followers_by_city` String CODEC(ZSTD(1)),
    `organization_name` String CODEC(ZSTD(1)),
    -- Page view statistics
    `page_views` Int64 DEFAULT 0,
    `unique_visitors` Int64 DEFAULT 0,
    `desktop_page_views` Int64 DEFAULT 0,
    `mobile_page_views` Int64 DEFAULT 0,
    `overview_page_views` Int64 DEFAULT 0,
    `about_page_views` Int64 DEFAULT 0,
    `jobs_page_views` Int64 DEFAULT 0,
    `people_page_views` Int64 DEFAULT 0,
    `careers_page_views` Int64 DEFAULT 0,
    `life_at_page_views` Int64 DEFAULT 0,
    `insights_page_views` Int64 DEFAULT 0,
    `products_page_views` Int64 DEFAULT 0,
    `page_views_by_country` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_region` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_industry` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_seniority` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_function` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_staff_count` String DEFAULT '' CODEC(ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(toDate(substring(record_id, position(record_id, '_') + 1)))
PRIMARY KEY (linkedin_id, record_id)
ORDER BY (linkedin_id, record_id)
SETTINGS index_granularity = 8192;

-- ALTER statement to add created_at column to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `created_at` DateTime64(6) DEFAULT '0000000000.000000';

-- ALTER statement to add reactions column to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `reactions` Int64 DEFAULT 0;

-- ALTER statement to add engagement column to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `engagement` Float64 DEFAULT 0;

-- ALTER statement to add daily_follower_count column to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `daily_follower_count` Int64 DEFAULT 0;

-- ALTER TABLE statements to add page view columns to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `unique_visitors` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `desktop_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `mobile_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `overview_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `about_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `jobs_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `people_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `careers_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `life_at_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `insights_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `products_page_views` Int64 DEFAULT 0;
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views_by_country` String DEFAULT '' CODEC(ZSTD(1));
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views_by_region` String DEFAULT '' CODEC(ZSTD(1));
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views_by_industry` String DEFAULT '' CODEC(ZSTD(1));
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views_by_seniority` String DEFAULT '' CODEC(ZSTD(1));
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views_by_function` String DEFAULT '' CODEC(ZSTD(1));
-- ALTER TABLE contentstudiobackend.linkedin_insights ADD COLUMN IF NOT EXISTS `page_views_by_staff_count` String DEFAULT '' CODEC(ZSTD(1));

-- LinkedIn Posts Table
CREATE TABLE contentstudiobackend.linkedin_posts
(
    `linkedin_id` String,
    `post_id` String,
    `activity` String,
    `media_type` String,
    `article_url` String,
    `article_title` String,
    `post_data` String,
    `image` String,
    `media` Array(String),
    `type` String,
    `hashtags` Array(String),
    `comments` Int64,
    `total_engagement` Float64 DEFAULT 0,
    `favorites` Int64,
    `title` String,
    `day_of_week` String,
    `hour_of_day` Int64,
    `created_at` DateTime64(6) DEFAULT '0000000000.000000',
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000',
    `updated_time` DateTime64(6) DEFAULT '0000000000.000000',
    `poll_data` String,
    `reach` Int64,
    `repost` Int64,
    `post_clicks` Int64,
    `impressions` Int64,
    `published_at` DateTime64(6) DEFAULT '0000000000.000000',
    `last_modified_at` DateTime64(6) DEFAULT '0000000000.000000',
    `lifecycle_state` String DEFAULT '',
    `visibility` String DEFAULT '',
    `is_reshare_disabled` Bool DEFAULT false,
    `feed_distribution` String DEFAULT '',
    `third_party_channels` Array(String) DEFAULT []
)
ENGINE = ReplacingMergeTree(saving_time)
PARTITION BY toYYYYMM(published_at)
PRIMARY KEY (linkedin_id, post_id)
ORDER BY (linkedin_id, post_id)
SETTINGS index_granularity = 8192;

-- ALTER statements to add new columns to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `published_at` DateTime64(6) DEFAULT '0000000000.000000';
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `last_modified_at` DateTime64(6) DEFAULT '0000000000.000000';
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `updated_time` DateTime64(6) DEFAULT '0000000000.000000';
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `lifecycle_state` String DEFAULT '';
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `visibility` String DEFAULT '';
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `is_reshare_disabled` Bool DEFAULT false;
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `feed_distribution` String DEFAULT '';
-- ALTER TABLE contentstudiobackend.linkedin_posts ADD COLUMN IF NOT EXISTS `third_party_channels` Array(String) DEFAULT [];

-- ALTER statement to change total_engagement from Int64 to Float64 in linkedin_posts:
-- ALTER TABLE contentstudiobackend.linkedin_posts MODIFY COLUMN `total_engagement` Float64 DEFAULT 0;

-- LinkedIn Geo Mapping Table (cache for geo ID to name resolution)
-- Stores mappings from LinkedIn geo IDs to human-readable location names.
-- Used to avoid repeated calls to LinkedIn Geo API for the same IDs.
CREATE TABLE contentstudiobackend.linkedin_geo_mapping
(
    `geo_id` String,
    `geo_name` String,
    `geo_type` String DEFAULT '',
    `created_at` DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree()
PRIMARY KEY geo_id
ORDER BY geo_id;

-- ALTER statement to add geo_type column to existing table:
-- ALTER TABLE contentstudiobackend.linkedin_geo_mapping ADD COLUMN IF NOT EXISTS `geo_type` String DEFAULT '';
