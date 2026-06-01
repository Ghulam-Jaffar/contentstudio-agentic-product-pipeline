---
title: Instagram ClickHouse Sink
description: Documentation for Instagram ClickHouse sink
---

Purpose
- Consumes parsed Instagram topics; batches inserts into ClickHouse tables.

Topics
- In: `parsed-instagram-posts`, `parsed-instagram-insights`

Run
```bash
cd src && make instagram_clickhouse_sink
../bin/instagram_clickhouse_sink
```

Notes
- Keep batch sizes ~1000; monitor insert latencies; align models with CH schemas.
