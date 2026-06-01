---
title: TikTok Fetcher
description: Documentation for TikTok fetcher
---

Purpose
- Consumes `work-order-tiktok`; fetches media/insights via TikTok APIs; publishes to raw topics.

Topics
- In: `work-order-tiktok` (if used) or scheduler integration
- Out: `raw-tiktok-posts`, `raw-tiktok-insights` (align with actual topics once finalized)

Run
```bash
cd src && make tiktok_fetcher
../bin/tiktok_fetcher
```

Notes
- OAuth/token refresh and rate limiting are critical; follow Python exactly.
