---
title: Scheduler Account Fetcher
description: Documentation for scheduler account fetcher
---

Purpose
- Reads platform accounts from Mongo; publishes work orders to Kafka.
- Handles account scheduling only; the separate URL refresher job updates stale media URLs in ClickHouse.

Run (Facebook example)
```bash
cd src && make account_fetcher
../bin/account_fetcher -socialNetwork facebook -accountType page -syncType incremental
```

Notes
- Uses `APP_MONGO_*` and `APP_KAFKA_*` envs; adjust `accountType` mapping as needed.
- For URL maintenance jobs, see `/docs/services/url-refresher`.
