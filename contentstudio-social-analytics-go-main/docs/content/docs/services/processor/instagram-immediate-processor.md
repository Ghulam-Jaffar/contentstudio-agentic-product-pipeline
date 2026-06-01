---
title: Instagram Immediate Processor
description: Documentation for Instagram immediate processor
---

Purpose
- Consumes parsed Instagram topics; applies real-time enrichment; publishes processed topics for sinks.

Topics
- In: `parsed-instagram-posts`, `parsed-instagram-insights`
- Out: `processed_instagram_posts`, `processed_instagram_insights`

Run
```bash
cd src && make instagram_immediate_processor
../bin/instagram_immediate_processor
```

Notes
- Mirror the Facebook immediate processor pattern; ensure idempotency and metrics.
