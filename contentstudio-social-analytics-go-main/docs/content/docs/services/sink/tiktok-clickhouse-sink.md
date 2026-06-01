---
title: TikTok ClickHouse Sink
description: Documentation for TikTok ClickHouse sink
---

Purpose
- Consumes parsed TikTok topics; batches inserts into ClickHouse tables.

Run
```bash
cd src && make tiktok_clickhouse_sink
../bin/tiktok_clickhouse_sink
```

Notes
- Confirm ClickHouse schemas; tune batch sizes; add metrics.
