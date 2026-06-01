-- Pinterest Boards Table
-- Stores board metadata

CREATE TABLE IF NOT EXISTS pinterest_boards
(
    record_id String CODEC(ZSTD(1)),
    user_id String CODEC(ZSTD(1)),
    board_id String CODEC(ZSTD(1)),
    name String CODEC(ZSTD(1)),
    owner String CODEC(ZSTD(1)),
    description String CODEC(ZSTD(1)),
    privacy String CODEC(ZSTD(1)),
    image_cover_url String CODEC(ZSTD(1)),
    pin_thumbnail_urls Array(String) CODEC(ZSTD(1)),
    collaborator_count Int64 CODEC(ZSTD(1)),
    pin_count Int64 CODEC(ZSTD(1)),
    follower_count Int64 CODEC(ZSTD(1)),
    created_at DateTime CODEC(Delta(4), ZSTD(1)),
    inserted_at DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplacingMergeTree(inserted_at)
PRIMARY KEY (record_id)
ORDER BY (record_id, inserted_at)
PARTITION BY toYYYYMM(created_at);
