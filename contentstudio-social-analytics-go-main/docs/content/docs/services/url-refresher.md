---
title: URL Refresher Job
description: Platform-specific job for refreshing stale media URLs and thumbnails in ClickHouse
---

Purpose
- Refreshes stale media URLs after ingestion so ClickHouse keeps usable thumbnails and post links.
- Runs as a platform-specific batch job instead of a Kafka pipeline stage.

Supported modes
- `facebook`: refreshes Facebook post thumbnails from `facebook_posts`.
- `instagram`: refreshes Instagram media URLs from `instagram_posts`.
- `linkedin`: refreshes LinkedIn post URLs from `linkedin_posts`.
- `facebook-competitor`: refreshes competitor Facebook media URLs and shared post pictures.
- `instagram-competitor`: refreshes competitor Instagram media URLs and profile pictures.
- `competitors`: runs both competitor refreshers.
- `all`: runs every refresher sequentially.

Run
```bash
cd src
make url_refersher
../bin/url_refersher -platform facebook -accountType page
```

Flags
- `-platform`: one of `facebook`, `instagram`, `linkedin`, `facebook-competitor`, `instagram-competitor`, `competitors`, or `all`
- `-accountType`: optional filter for Facebook, Instagram, and LinkedIn account selection

Dependencies
- MongoDB for unified social and competitor account lookup
- ClickHouse for reading stale media rows and writing refreshed URLs
- Redis for competitor token selection
- Facebook/Instagram/LinkedIn API credentials and `APP_DECRYPTION_KEY`

Notes
- Facebook account type `page` is normalized to the stored `Page` value during lookup.
- Instagram and LinkedIn refreshers decrypt long-lived tokens when available.
- The job only updates rows with older media content, so it is safe to run repeatedly.
