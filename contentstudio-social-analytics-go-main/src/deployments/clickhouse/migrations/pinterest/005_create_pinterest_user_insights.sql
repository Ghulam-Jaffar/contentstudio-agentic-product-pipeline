-- Pinterest User Insights Table
-- Stores daily analytics data for each user

CREATE TABLE IF NOT EXISTS pinterest_user_insights
(
    record_id String CODEC(ZSTD(1)),
    user_id String CODEC(ZSTD(1)),
    date DateTime CODEC(Delta(4), ZSTD(1)),
    data_status String CODEC(ZSTD(1)),
    impression Int64 CODEC(T64, ZSTD(1)),
    pin_clicks Int64 CODEC(T64, ZSTD(1)),
    pin_click_rate Float64 CODEC(ZSTD(1)),
    outbound_click Int64 CODEC(T64, ZSTD(1)),
    saves Int64 CODEC(T64, ZSTD(1)),
    save_rate Float64 CODEC(ZSTD(1)),
    clickthrough Int64 CODEC(T64, ZSTD(1)),
    clickthrough_rate Float64 CODEC(ZSTD(1)),
    engagement Int64 CODEC(T64, ZSTD(1)),
    engagement_rate Float64 CODEC(ZSTD(1)),
    video_mrc_view Int64 CODEC(T64, ZSTD(1)),
    video_start Int64 CODEC(T64, ZSTD(1)),
    video_10s_view Int64 CODEC(T64, ZSTD(1)),
    video_avg_watch_time Int64 CODEC(T64, ZSTD(1)),
    video_v50_watch_time Int64 CODEC(T64, ZSTD(1)),
    full_screen_play Int64 CODEC(T64, ZSTD(1)),
    full_screen_playtime Int64 CODEC(T64, ZSTD(1)),
    profile_visit Int64 CODEC(T64, ZSTD(1)),
    closeup Int64 CODEC(T64, ZSTD(1)),
    quartile_95s_percent_view Int64 CODEC(T64, ZSTD(1)),
    inserted_at DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PRIMARY KEY (record_id)
ORDER BY (record_id, inserted_at)
PARTITION BY toYYYYMM(inserted_at);
