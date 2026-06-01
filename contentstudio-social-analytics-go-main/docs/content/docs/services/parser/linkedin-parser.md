---
title: LinkedIn Parser
description: Documentation for LinkedIn parser
---

Purpose
- Consumes LinkedIn raw topics; parses items into structured models; publishes parsed topics.

Topics
- In: `raw-linkedin-posts`, `raw-linkedin-images`, `raw-linkedin-videos`, `raw-linkedin-stats`
- Out: `parsed-linkedin-posts`, `parsed-linkedin-media-assets`, `parsed-linkedin-stats`

Run
```bash
cd src && make linkedin_parser
../bin/linkedin_parser
```

Notes
- Align field mappings with Python; verify media assets handling; add golden tests.
