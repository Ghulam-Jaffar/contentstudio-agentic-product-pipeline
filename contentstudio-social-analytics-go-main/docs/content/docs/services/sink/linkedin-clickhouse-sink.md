---
title: LinkedIn ClickHouse Sink
description: Documentation for LinkedIn ClickHouse sink
---

Purpose
- Consumes parsed LinkedIn topics; batches inserts into ClickHouse tables.

Topics
- In: `parsed-linkedin-posts`, `parsed-linkedin-media-assets`, `parsed-linkedin-stats`

Run
```bash
cd src && make linkedin_clickhouse_sink
../bin/linkedin_clickhouse_sink
```

Notes
- Keep batch operations; ensure tables exist; monitor throughput and failures.
