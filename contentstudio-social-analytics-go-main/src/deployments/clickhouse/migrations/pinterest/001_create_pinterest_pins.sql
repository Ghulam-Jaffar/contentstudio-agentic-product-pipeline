-- Pinterest Pins Table
-- Stores pin metadata and media information

CREATE TABLE IF NOT EXISTS pinterest_pins
(
    record_id String CODEC(ZSTD(1)),
    pin_id String CODEC(ZSTD(1)),
    user_id String CODEC(ZSTD(1)),
    board_id String CODEC(ZSTD(1)),
    board_section_id String CODEC(ZSTD(1)),
    parent_pin_id String CODEC(ZSTD(1)),
    title String CODEC(ZSTD(1)),
    note String CODEC(ZSTD(1)),
    description String CODEC(ZSTD(1)),
    link String CODEC(ZSTD(1)),
    dominant_color String CODEC(ZSTD(1)),
    creative_type String CODEC(ZSTD(1)),
    media_type String CODEC(ZSTD(1)),
    cover_image_url String CODEC(ZSTD(1)),
    video_url String CODEC(ZSTD(1)),
    duration String CODEC(ZSTD(1)),
    height String CODEC(ZSTD(1)),
    width String CODEC(ZSTD(1)),
    is_standard Int64 CODEC(T64, ZSTD(1)),
    is_owner Int64 CODEC(T64, ZSTD(1)),
    has_been_promoted Int64 CODEC(T64, ZSTD(1)),
    board_owner String CODEC(ZSTD(1)),
    product_tags Array(String) CODEC(ZSTD(1)),
    created_at DateTime CODEC(Delta(4), ZSTD(1)),
    day_of_week String CODEC(ZSTD(1)),
    hour_of_day Int64 CODEC(T64, ZSTD(1)),
    inserted_at DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PRIMARY KEY (pin_id, created_at)
ORDER BY (pin_id, created_at)
PARTITION BY toYYYYMM(created_at);
