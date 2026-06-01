---
title: Facebook ClickHouse Sink
description: Documentation for Facebook ClickHouse sink
---

Purpose
- Consumes parsed topics; batches inserts into ClickHouse tables.

Topics
- In: `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-insights`, `parsed-facebook-reels-insights`

Run
```bash
cd src && make facebook_clickhouse_sink
../bin/facebook_clickhouse_sink
```

Notes
- Uses `clickhouse-go` `PrepareBatch`; configured via `APP_CLICKHOUSE_*` envs.
