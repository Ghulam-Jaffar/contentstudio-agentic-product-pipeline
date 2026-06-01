---
title: Facebook Fetcher
description: Documentation for Facebook fetcher
---

Purpose
- Consumes `work-order-facebook`; fetches posts, videos, and insights from Graph API; publishes to raw topics.

Topics
- In: `work-order-facebook`
- Out: `raw-facebook-posts`, `raw-facebook-videos`, `raw-facebook-insights`

Run
```bash
cd src && make facebook_fetcher
../bin/facebook_fetcher
```

Config
- Requires `APP_FACEBOOK_APP_SECRET`, `APP_KAFKA_BROKERS`, `APP_DECRYPTION_KEY`, Mongo/CH envs.

Notes
- Uses appsecret_proof; respects pagination; logs fbtrace_id on errors.
