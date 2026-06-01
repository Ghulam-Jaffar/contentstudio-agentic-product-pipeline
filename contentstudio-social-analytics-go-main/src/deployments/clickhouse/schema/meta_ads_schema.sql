-- Meta Ads Tables Schema
-- All tables use ReplacingMergeTree(inserted_at) for deduplication.
-- List tables (account_info, campaigns, adsets, ads) partition by created_time.
-- Insights tables partition by insights_date (date_start = date_stop when time_increment=1).

-- ─────────────────────────────────────────────
-- 1. Ad Account Info
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_account_info
(
    `account_id`      String,
    `name`            String,
    `currency`        LowCardinality(String),
    `account_status`  Int32 DEFAULT 0,
    `timezone_name`   LowCardinality(String),
    `business_id`     String DEFAULT '',
    `business_name`   String DEFAULT '',
    `amount_spent`    String DEFAULT '',
    `balance`         String DEFAULT '',
    `spend_cap`       String DEFAULT '',
    `created_time`    DateTime64(6) DEFAULT '0000000000.000000',
    `inserted_at`     DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_time)
PRIMARY KEY account_id
ORDER BY account_id
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 2. Campaigns
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_campaigns
(
    `account_id`        String,
    `campaign_id`       String,
    `name`              String,
    `status`            LowCardinality(String),
    `effective_status`  LowCardinality(String),
    `objective`         LowCardinality(String) DEFAULT '',
    `daily_budget`      String DEFAULT '',
    `lifetime_budget`   String DEFAULT '',
    `budget_remaining`  String DEFAULT '',
    `start_time`        DateTime64(6) DEFAULT '0000000000.000000',
    `stop_time`         DateTime64(6) DEFAULT '0000000000.000000',
    `created_time`      DateTime64(6) DEFAULT '0000000000.000000',
    `updated_time`      DateTime64(6) DEFAULT '0000000000.000000',
    `inserted_at`       DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_time)
PRIMARY KEY (account_id, campaign_id)
ORDER BY (account_id, campaign_id)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 3. Ad Sets
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_adsets
(
    `account_id`           String,
    `adset_id`             String,
    `name`                 String,
    `campaign_id`          String,
    `status`               LowCardinality(String),
    `effective_status`     LowCardinality(String),
    `daily_budget`         String DEFAULT '',
    `lifetime_budget`      String DEFAULT '',
    `budget_remaining`     String DEFAULT '',
    `billing_event`        LowCardinality(String) DEFAULT '',
    `optimization_goal`    LowCardinality(String) DEFAULT '',
    `bid_strategy`         LowCardinality(String) DEFAULT '',
    `age_min`              Int32 DEFAULT 0,
    `age_max`              Int32 DEFAULT 0,
    `targeting_countries`  Array(String),
    `targeting_json`       String DEFAULT '',
    `start_time`           DateTime64(6) DEFAULT '0000000000.000000',
    `stop_time`            DateTime64(6) DEFAULT '0000000000.000000',
    `end_time`             DateTime64(6) DEFAULT '0000000000.000000',
    `created_time`         DateTime64(6) DEFAULT '0000000000.000000',
    `inserted_at`          DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_time)
PRIMARY KEY (account_id, adset_id)
ORDER BY (account_id, adset_id)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 4. Ads
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_ads
(
    `account_id`              String,
    `ad_id`                   String,
    `name`                    String,
    `adset_id`                String,
    `adset_name`              String DEFAULT '',
    `campaign_id`             String,
    `campaign_name`           String DEFAULT '',
    `status`                  LowCardinality(String),
    `effective_status`        LowCardinality(String),
    `objective`               LowCardinality(String) DEFAULT '',
    `creative_id`             String DEFAULT '',
    `creative_name`           String DEFAULT '',
    `creative_title`          String DEFAULT '',
    `creative_body`           String DEFAULT '',
    `creative_image_url`      String DEFAULT '',
    `creative_thumbnail_url`  String DEFAULT '',
    `creative_object_type`    LowCardinality(String) DEFAULT '',
    `creative_effective_object_story_id` String DEFAULT '',
    `daily_budget`            String DEFAULT '',
    `lifetime_budget`         String DEFAULT '',
    `budget_remaining`        String DEFAULT '',
    `created_time`            DateTime64(6) DEFAULT '0000000000.000000',
    `updated_time`            DateTime64(6) DEFAULT '0000000000.000000',
    `inserted_at`             DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(created_time)
PRIMARY KEY (account_id, ad_id)
ORDER BY (account_id, ad_id)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 5. Campaign Insights  (time_increment=1, insights_date = date_start = date_stop)
--    Actions are flattened from the actions[] array.
--    Outbound / video metrics come from additional response fields.
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_campaign_insights
(
    `account_id`                                     String,
    `campaign_id`                                    String,
    `campaign_name`                                  String,
    `objective`                                      LowCardinality(String) DEFAULT '',
    `insights_date`                                  Date,
    `spend`                                          Float64 DEFAULT 0,
    `impressions`                                    Int64 DEFAULT 0,
    `reach`                                          Int64 DEFAULT 0,
    `clicks`                                         Int64 DEFAULT 0,
    `unique_clicks`                                  Int64 DEFAULT 0,
    `ctr`                                            Float64 DEFAULT 0,
    `unique_ctr`                                     Float64 DEFAULT 0,
    `cpc`                                            Float64 DEFAULT 0,
    `cpm`                                            Float64 DEFAULT 0,
    `cpp`                                            Float64 DEFAULT 0,
    `frequency`                                      Float64 DEFAULT 0,
    `actions_purchase`                               Int64 DEFAULT 0,
    `actions_post_engagement`                        Int64 DEFAULT 0,
    `actions_offsite_conversion_fb_pixel_purchase`   Int64 DEFAULT 0,
    `actions_link_click`                             Int64 DEFAULT 0,
    `actions_lead`                                   Int64 DEFAULT 0,
    `actions_offsite_conversion_fb_pixel_lead`       Int64 DEFAULT 0,
    `actions_mobile_app_install`                     Int64 DEFAULT 0,
    `inserted_at`                                    DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(insights_date)
PRIMARY KEY (account_id, campaign_id, insights_date)
ORDER BY (account_id, campaign_id, insights_date)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 6. Adset Insights  (time_increment=1)
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_adset_insights
(
    `account_id`                                     String,
    `adset_id`                                       String,
    `adset_name`                                     String,
    `campaign_id`                                    String,
    `campaign_name`                                  String,
    `insights_date`                                  Date,
    `spend`                                          Float64 DEFAULT 0,
    `impressions`                                    Int64 DEFAULT 0,
    `reach`                                          Int64 DEFAULT 0,
    `clicks`                                         Int64 DEFAULT 0,
    `unique_clicks`                                  Int64 DEFAULT 0,
    `ctr`                                            Float64 DEFAULT 0,
    `unique_ctr`                                     Float64 DEFAULT 0,
    `cpc`                                            Float64 DEFAULT 0,
    `cpm`                                            Float64 DEFAULT 0,
    `cpp`                                            Float64 DEFAULT 0,
    `frequency`                                      Float64 DEFAULT 0,
    `actions_purchase`                               Int64 DEFAULT 0,
    `actions_post_engagement`                        Int64 DEFAULT 0,
    `actions_offsite_conversion_fb_pixel_purchase`   Int64 DEFAULT 0,
    `actions_link_click`                             Int64 DEFAULT 0,
    `actions_lead`                                   Int64 DEFAULT 0,
    `actions_offsite_conversion_fb_pixel_lead`       Int64 DEFAULT 0,
    `actions_mobile_app_install`                     Int64 DEFAULT 0,
    `inserted_at`                                    DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(insights_date)
PRIMARY KEY (account_id, adset_id, insights_date)
ORDER BY (account_id, adset_id, insights_date)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 7. Ad Insights  (time_increment=1)
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_ad_insights
(
    `account_id`                                     String,
    `ad_id`                                          String,
    `ad_name`                                        String,
    `adset_id`                                       String,
    `campaign_id`                                    String,
    `campaign_name`                                  String,
    `insights_date`                                  Date,
    `spend`                                          Float64 DEFAULT 0,
    `impressions`                                    Int64 DEFAULT 0,
    `reach`                                          Int64 DEFAULT 0,
    `clicks`                                         Int64 DEFAULT 0,
    `unique_clicks`                                  Int64 DEFAULT 0,
    `ctr`                                            Float64 DEFAULT 0,
    `unique_ctr`                                     Float64 DEFAULT 0,
    `cpc`                                            Float64 DEFAULT 0,
    `cpm`                                            Float64 DEFAULT 0,
    `cpp`                                            Float64 DEFAULT 0,
    `frequency`                                      Float64 DEFAULT 0,
    `actions_purchase`                               Int64 DEFAULT 0,
    `actions_post_engagement`                        Int64 DEFAULT 0,
    `actions_offsite_conversion_fb_pixel_purchase`   Int64 DEFAULT 0,
    `actions_link_click`                             Int64 DEFAULT 0,
    `actions_lead`                                   Int64 DEFAULT 0,
    `actions_offsite_conversion_fb_pixel_lead`       Int64 DEFAULT 0,
    `actions_mobile_app_install`                     Int64 DEFAULT 0,
    `inserted_at`                                    DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(insights_date)
PRIMARY KEY (account_id, ad_id, insights_date)
ORDER BY (account_id, ad_id, insights_date)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 8. Demographics: Age & Gender  (breakdowns=age,gender, time_increment=1)
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_demographics_age_gender
(
    `account_id`    String,
    `insights_date` Date,
    `age`           LowCardinality(String),
    `gender`        LowCardinality(String),
    `impressions`   Int64 DEFAULT 0,
    `reach`         Int64 DEFAULT 0,
    `clicks`        Int64 DEFAULT 0,
    `spend`         Float64 DEFAULT 0,
    `ctr`           Float64 DEFAULT 0,
    `cpm`           Float64 DEFAULT 0,
    `cpc`           Float64 DEFAULT 0,
    `cpp`           Float64 DEFAULT 0,
    `frequency`     Float64 DEFAULT 0,
    `inserted_at`   DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(insights_date)
PRIMARY KEY (account_id, insights_date, age, gender)
ORDER BY (account_id, insights_date, age, gender)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 9. Demographics: Device & Platform  (breakdowns=impression_device,publisher_platform,platform_position, time_increment=1)
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_demographics_device_platform
(
    `account_id`         String,
    `insights_date`      Date,
    `impression_device`  LowCardinality(String),
    `publisher_platform` LowCardinality(String),
    `platform_position`  LowCardinality(String),
    `impressions`        Int64 DEFAULT 0,
    `reach`              Int64 DEFAULT 0,
    `clicks`             Int64 DEFAULT 0,
    `spend`              Float64 DEFAULT 0,
    `ctr`                Float64 DEFAULT 0,
    `cpm`                Float64 DEFAULT 0,
    `cpc`                Float64 DEFAULT 0,
    `cpp`                Float64 DEFAULT 0,
    `frequency`          Float64 DEFAULT 0,
    `inserted_at`        DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(insights_date)
PRIMARY KEY (account_id, insights_date, impression_device, publisher_platform, platform_position)
ORDER BY (account_id, insights_date, impression_device, publisher_platform, platform_position)
SETTINGS index_granularity = 8192;

-- ─────────────────────────────────────────────
-- 10. Demographics: Region & Country  (breakdowns=region,country, time_increment=1)
-- ─────────────────────────────────────────────
CREATE TABLE contentstudiobackend.meta_ads_demographics_region_country
(
    `account_id`    String,
    `insights_date` Date,
    `country`       LowCardinality(String),
    `region`        String,
    `impressions`   Int64 DEFAULT 0,
    `reach`         Int64 DEFAULT 0,
    `clicks`        Int64 DEFAULT 0,
    `spend`         Float64 DEFAULT 0,
    `ctr`           Float64 DEFAULT 0,
    `cpm`           Float64 DEFAULT 0,
    `cpc`           Float64 DEFAULT 0,
    `cpp`           Float64 DEFAULT 0,
    `frequency`     Float64 DEFAULT 0,
    `inserted_at`   DateTime64(6) DEFAULT toStartOfHour(now64(6))
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(insights_date)
PRIMARY KEY (account_id, insights_date, country, region)
ORDER BY (account_id, insights_date, country, region)
SETTINGS index_granularity = 8192;
