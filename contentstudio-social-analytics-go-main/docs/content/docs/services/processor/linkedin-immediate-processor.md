---
title: LinkedIn Immediate Processor
description: Documentation for LinkedIn immediate processor
---

Purpose
- Consumes parsed LinkedIn topics; applies real-time enrichment/validation; publishes processed topics for sinks.

Topics
- In: `parsed-linkedin-posts`, `parsed-linkedin-media-assets`, `parsed-linkedin-stats`
- Out: `processed-linkedin-posts`, `processed-linkedin-media-assets`, `processed-linkedin-stats`

Run
```bash
cd src && make linkedin_immediate_processor
../bin/linkedin_immediate_processor
```

Notes
- Ensure parity with Python immediate processing (if any); add error handling and metrics.
