---
title: Facebook Immediate Processor
description: Documentation for Facebook immediate processor
---

Purpose
- Consumes `parsed-facebook_*` topics; performs real-time enrichment/validation; publishes `processed_facebook_*` topics for sinks.

Topics
- In: `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-insights`, `parsed-facebook-reels-insights`
- Out: `processed-facebook-posts`, `processed-facebook-media-assets`, `processed-facebook-video-insights`, `processed-facebook-insights`, `processed-facebook-reels-insights`

Run
```bash
cd src && make facebook_immediate_processor
../bin/facebook_immediate_processor
```

Notes
- Keep transformations stateless and idempotent; add metrics for per-topic throughput and error rates.
