---
title: Kafka Topics
description: Documentation for Kafka topics
---

This document lists all Kafka topics required for the Facebook, Instagram, and LinkedIn analytics pipelines. Topics are grouped by logical stage (work-order → raw → parsed) to help DevOps create them in the appropriate clusters.

Conventions
- All topic names are lowercase and use underscores (`_`).
- Partitions/retention are environment-specific; below lists names only.

Stage Reference
- Immediate Work-Order: Triggers real-time processing for a single account; produced by the scheduler/UI, consumed by immediate-processor.
- Raw: Unaltered payloads fetched by the fetcher; consumed by the parser.
- Parsed: Normalised/enriched objects produced by the parser; consumed by the ClickHouse sink.

1) Facebook
- Work-order: `immediate-work-order-facebook`
- Raw: `raw-facebook-posts`, `raw-facebook-videos`, `raw-facebook-insights`
- Parsed: `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-reels-insights`, `parsed-facebook-insights`

2) Instagram
- Work-order: `immediate-work-order-instagram`
- Raw: `raw-instagram-media`, `raw-instagram-insights`
- Parsed: `parsed-instagram-posts`, `parsed-instagram-insights`

3) LinkedIn
- Work-order: `immediate-work-order-linkedin`
- Raw: `raw-linkedin-posts`, `raw-linkedin-images`, `raw-linkedin-videos`, `raw-linkedin-stats`
- Parsed: `parsed-linkedin-posts`, `parsed-linkedin-media-assets`, `parsed-linkedin-stats`

Topic Creation Example
```bash
# Assuming Kafka CLI tools and $BROKER points to your cluster
kafka-topics \
  --create --topic raw-facebook-posts \
  --partitions 6 --replication-factor 3 \
  --bootstrap-server "$BROKER"
```

Automate with a shell script for all topics and ensure retention matches data volume (raw topics often shorter than parsed).
