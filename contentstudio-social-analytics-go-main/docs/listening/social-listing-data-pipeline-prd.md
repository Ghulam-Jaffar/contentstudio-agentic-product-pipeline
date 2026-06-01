# Story: [BE] Data ingestion pipeline for 18-platform social listening

> Roadmap note: This document is future-state planning for the 18-platform expansion. Current v1 implementation scope is the 6-platform Data365 pipeline in `social-listening-implementation.md`, which remains the execution source of truth for this repo today.

## Details
	
- ID: sc-113241
- Type: feature
- Status: In Progress
- Requester: Ghulam Jaffar
- Owners: Zaid Bin Tariq
- Team: Back End
- Project: Web App
- Epic: Social Listening
- Iteration: 23 March - 03 April - 2026

## Description

**Description:**
As a backend system, I need a future expansion path from the current 6-platform Data365 implementation to an 18-platform ingestion architecture so that users eventually receive a broader stream of mentions matching their keywords across supported APIs, public sources, and crawl-based channels.

---

**Workflow:**

1. A topic is created or reactivated → the system registers a monitoring job for that topic's keywords + platform scope.
2. Per-platform crawlers/integrations fetch new content from their respective APIs (or RSS/scraping where no API exists):
   - **Streaming/Polling APIs:** Twitter/X (streaming), Reddit (PRAW), Bluesky (AT Protocol firehose), YouTube (search API).
   - **Business APIs:** Instagram, Facebook, LinkedIn, TikTok, Pinterest, Threads (Graph/official APIs where available).
   - **Public APIs:** Hacker News, GitHub, DEV.to, Stack Overflow.
   - **RSS + crawling:** Podcasts, Newsletters, News, Blogs.
3. Each fetched item is matched against the topic's keyword rules (keywords, include ANY, include ALL, negative terms, negative authors, exact match, case sensitive).
4. Global filters are applied (workspace-level negative terms, blocked authors, excluded subreddits, language filter).
5. Matching items are stored as `Mention` records in the database with: platform, author, content, URL, publish time, engagement counts, matched topic ID(s), workspace ID, and raw metadata.
6. Mention is queued for AI processing (smart tagging + sentiment — handled by Story 4).
7. Paused topics are excluded from crawling. Topics belonging to workspaces that have hit the monthly mention limit are excluded.
8. A background scheduler re-runs crawls on per-platform intervals (e.g., Twitter/X streaming is continuous; blogs are every 30 minutes).

---

**Acceptance criteria:**

- [ ] The future ingestion architecture is defined for all 18 roadmap platforms when a topic is in active state
- [ ] Keywords are matched according to: basic keyword match, include ANY (OR), include ALL (AND), negative terms (excluded), negative authors (excluded), exact match, case sensitive toggles
- [ ] Global workspace filters (negative terms, blocked authors, excluded subreddits, language) are applied before storing a mention
- [ ] Mentions are stored with: id, workspace_id, topic_id, platform, author_name, author_handle, author_follower_count, content, url, published_at, ingested_at, engagement (likes/comments/shares/reactions), is_read (default false), is_bookmarked (default false), is_irrelevant (default false)
- [ ] A mention matching multiple topics produces duplicate rows per matching topic
- [ ] Paused topics do not trigger new ingestion but their historical mentions remain queryable
- [ ] When the active mention quota materialized from Laravel billing is exhausted, no new mentions are ingested; existing mentions are unaffected
- [ ] Platform fetch failures (API down, rate limited) are retried with exponential backoff (3 retries); after all retries fail, the error is logged and the platform is marked as "degraded" for that crawl cycle
- [ ] Crawl intervals are configurable per platform; default: Twitter/X near-real-time, Reddit 15min, news/blogs 30min
- [ ] Unit tests cover keyword matching logic for all filter combinations

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `mentions` collection/table. New `listening_topics` and `listening_workspace_config` records. No impact on existing data.

**Impact on other products:** None in V1. Future: Inbox integration for "Open in Inbox" from a mention.

**Dependencies:** None (foundational story).

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Language filtering is part of the ingestion logic; must correctly handle non-English content and apply the workspace language filter
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — N/A, data layer has no white-label concern
- [ ] Cross-product impact assessment — Ingestion pipeline is isolated; no cross-product impact in V1
