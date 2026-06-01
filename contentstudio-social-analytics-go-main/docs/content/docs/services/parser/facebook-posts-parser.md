---
title: Facebook Posts Parser
description: Documentation for Facebook posts parser
---

Purpose
- Consumes raw Facebook topics; parses into structured models; publishes parsed topics.

Topics
- In: `raw-facebook-posts`, `raw-facebook-videos`, `raw-facebook-insights`
- Out: `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-insights`, `parsed-facebook-reels-insights`

Run
```bash
cd src && make facebook_posts_videos_parser
../bin/facebook_posts_videos_parser
```

Notes
- Routes by topic suffix; leverages `internal/parsing/facebook_parser.go`.
