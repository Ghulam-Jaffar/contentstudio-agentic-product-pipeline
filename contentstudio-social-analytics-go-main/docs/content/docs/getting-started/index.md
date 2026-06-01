---
title: Documentation Overview
description: Quick navigation and onboarding guide for the Social Analytics project
---

This folder is organized for quick navigation and onboarding.

- Guides
  - `guides/engineer-onboarding.md`: Start here to set up, run, and contribute.
- Planning
  - `planning/migration-plan.md`: End-to-end migration plan (Python → Go) with milestones.
- Architecture
  - `architecture/project-overview-and-code-review.md`: System overview and deep code review.
  - `architecture/social-media-integration-workflow.md`: How to add a new platform workflow (migration principles and patterns).
- Platforms
  - `platforms/facebook/workflow.md`: Facebook pipeline (full breakdown and diagrams).
  - `platforms/facebook/analytics-pipeline-review.md`: Facebook analytics pipeline review notes.
- Services
  - Fetchers: `services/facebook-fetcher.md`, `services/instagram-fetcher.md`, `services/linkedin-fetcher.md`, `services/tiktok-fetcher.md`
  - Parsers: `services/facebook-posts-parser.md`, `services/instagram-posts-parser.md`, `services/linkedin-parser.md`, `services/tiktok-parser.md`
  - Sinks: `services/facebook-clickhouse-sink.md`, `services/instagram-clickhouse-sink.md`, `services/linkedin-clickhouse-sink.md`, `services/tiktok-clickhouse-sink.md`
  - Immediate: `services/facebook-immediate-processor.md`, `services/instagram-immediate-processor.md`, `services/linkedin-immediate-processor.md`, `services/immediate-work-api.md`
- Environments
    - Local Environment: `environment/local-setup.md`
- Reference
  - Kafka topics map: `reference/kafka-topics.md`
  - Improvements/tech debt: `reference/improvements.md`

Notes:
- Some reference files (Kafka topics, improvements) live at the repo root for historical reasons; links above point to them directly.

