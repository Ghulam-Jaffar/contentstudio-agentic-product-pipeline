---
title: Instagram Fetcher
description: Documentation for Instagram fetcher
---

Purpose
- Consumes `work-order-instagram`; fetches media and insights from Instagram APIs; publishes to raw topics.

Topics
- In: `work-order-instagram`
- Out: `raw-instagram-media`, `raw-instagram-insights`

Run
```bash
cd src && make instagram_fetcher
../bin/instagram_fetcher
```

Notes
- Follows the same pattern as Facebook fetcher; confirm endpoints and fields match Python.
