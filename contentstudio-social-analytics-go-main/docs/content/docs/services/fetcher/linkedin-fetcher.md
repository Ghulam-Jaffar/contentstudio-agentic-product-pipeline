---
title: LinkedIn Fetcher
description: Documentation for LinkedIn fetcher
---

Purpose
- Consumes `work-order-linkedin`; fetches posts/images/videos/stats; publishes to raw topics.

Topics
- In: `work-order-linkedin`
- Out: `raw-linkedin-posts`, `raw-linkedin-images`, `raw-linkedin-videos`, `raw-linkedin-stats`

Run
```bash
cd src && make linkedin_fetcher
../bin/linkedin_fetcher
```

Notes
- Validate scopes and rate limits; preserve Python endpoints and params.
