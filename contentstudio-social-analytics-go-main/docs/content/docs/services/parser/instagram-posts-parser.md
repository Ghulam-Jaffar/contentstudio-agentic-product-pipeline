---
title: Instagram Posts Parser
description: Documentation for Instagram posts parser
---

Purpose
- Consumes Instagram raw topics; parses media/insights into structured models; publishes parsed topics.

Topics
- In: `raw-instagram-media`, `raw-instagram-insights`
- Out: `parsed-instagram-posts`, `parsed-instagram-insights`

Run
```bash
cd src && make instagram_posts_parser
../bin/instagram_posts_parser
```

Notes
- Ensure field mapping parity against Python outputs; add golden tests for key payloads.
