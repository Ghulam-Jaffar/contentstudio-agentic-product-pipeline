# Epic + Stories: Listening Add-on

**Pipeline Step:** 4 of 5
**Date:** 2026-03-06
**PRD:** `docs/features/social-listening/03-prd.md`

---

## Epic

**Title:** Listening Add-on

**Description:**

Listening is a $49/month paid add-on that turns ContentStudio into a complete brand monitoring hub. It watches 18 platforms — from Twitter/X and Reddit to blogs, podcasts, and newsletters — for user-defined keywords and surfaces every match in a single, intelligent feed. AI-powered smart tagging (10 categories) and sentiment scoring automatically classify every mention so users can see what matters most at a glance.

Users create topics (brand names, competitor names, industry keywords) with precise keyword matching rules and platform scope. Everything flows into a global feed where saved views act like filtered inboxes — "Crisis Monitor", "Brand Love", "Competitor Intel". Users can reply to mentions inline (with AI compose assistance), save important mentions as bookmarks, set spike alerts via email, and export analytics as PDF/CSV or scheduled reports.

The feature is gated as a paid add-on: trial users see an upsell page; paid CS subscribers can enable it immediately. Default limits are 5 topics and 10,000 mentions/month, with add-ons available for more. Listening is web-only in V1 with full mobile responsiveness.

---

## Story Assignments at a Glance

| # | Title | Team | Priority | Project |
|---|---|---|---|---|
| 1 | [BE] Data ingestion pipeline for 18-platform social listening | Backend | P0 | Web App |
| 2 | [BE] Topic management API with keyword rules and pause/resume | Backend | P0 | Web App |
| 3 | [BE] Mention query and filtering API with views support | Backend | P0 | Web App |
| 4 | [BE] Smart tagging and sentiment scoring pipeline | Backend | P0 | Web App |
| 5 | [BE] Listening alerts engine — volume spike and sentiment spike detection | Backend | P0 | Web App |
| 6 | [BE] Analytics aggregation API for mentions, sentiment, platform, and tag data | Backend | P0 | Web App |
| 7 | [BE] Export service — PDF/CSV, scheduled email reports, and shareable analytics links | Backend | P1 | Web App |
| 8 | [BE] Global settings and topic types API | Backend | P1 | Web App |
| 9 | [BE] Listening add-on subscription management and usage limits | Backend | P0 | Web App |
| 10 | [FE] Listening module shell — navigation, routing, and user state management | Frontend | P0 | Web App |
| 11 | [FE] Listening landing page for trial, locked, and expired states | Frontend | P0 | Web App |
| 12 | [FE] Onboarding setup wizard with AI-powered topic suggestions | Frontend | P1 | Web App |
| 13 | [FE] Mention feed with filter bar, mention cards, and action pill | Frontend | P0 | Web App |
| 14 | [FE] Inline mention reply with account selector and AI compose toolbar | Frontend | P1 | Web App |
| 15 | [FE] Topic management UI — create, edit, pause, and delete topics | Frontend | P0 | Web App |
| 16 | [FE] Views sidebar and custom view creation | Frontend | P1 | Web App |
| 17 | [FE] Bookmarks page with search and filters | Frontend | P1 | Web App |
| 18 | [FE] Analytics dashboard with charts, KPI cards, and AI insights | Frontend | P1 | Web App |
| 19 | [FE] Export modal — download, scheduled reports, and share link | Frontend | P1 | Web App |
| 20 | [FE] Alerts management UI — create, edit, and manage alert rules | Frontend | P1 | Web App |
| 21 | [FE] Settings page — global filters and topic type management | Frontend | P1 | Web App |
| 22 | [FE] Usage bar, limit warnings, and add-on upgrade prompts | Frontend | P0 | Web App |

---

## Stories

---

### Story 1: [BE] Data ingestion pipeline for 18-platform social listening

**Description:**
As a backend system, I need to continuously crawl and ingest mentions from 18 platforms for all active topics so that users receive a real-time stream of mentions matching their keywords.

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

- [ ] Mention ingestion is active for all 18 platforms when a topic is in active state
- [ ] Keywords are matched according to: basic keyword match, include ANY (OR), include ALL (AND), negative terms (excluded), negative authors (excluded), exact match, case sensitive toggles
- [ ] Global workspace filters (negative terms, blocked authors, excluded subreddits, language) are applied before storing a mention
- [ ] Mentions are stored with: id, workspace_id, topic_id, platform, author_name, author_handle, author_follower_count, content, url, published_at, ingested_at, engagement (likes/comments/shares/reactions), is_read (default false), is_bookmarked (default false), is_irrelevant (default false)
- [ ] A mention matching multiple topics produces one record linked to all matching topic IDs (not duplicates)
- [ ] Paused topics do not trigger new ingestion but their historical mentions remain queryable
- [ ] When a workspace's monthly mention count reaches the limit (default 10,000), no new mentions are ingested; existing mentions are unaffected
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

---

### Story 2: [BE] Topic management API with keyword rules and pause/resume

**Description:**
As a backend system, I need to expose CRUD endpoints for Listening topics so that users can create, edit, pause, resume, and delete topics with full keyword configuration and type assignment.

---

**Workflow:**

1. User creates a topic via the UI → `POST /api/listening/topics` is called.
2. API validates:
   - `name` is required
   - `keywords` must have ≥1 entry
   - `platforms` must have ≥1 entry
   - Keyword conflicts detected and returned as structured errors (not just text):
     - A keyword that also appears in `negative_terms`
     - An include term that also appears in `negative_terms`
     - A term appearing in both `include_any` and `include_all`
3. Topic is created with status `active`; workspace topic count is incremented.
4. If workspace topic count would exceed the limit (default 5), creation is rejected with error code `TOPIC_LIMIT_REACHED`.
5. User pauses a topic → `PATCH /api/listening/topics/:id` with `{ "status": "paused" }` → monitoring stops, historical mentions remain.
6. User resumes → same PATCH with `{ "status": "active" }` → monitoring restarts.
7. User deletes → `DELETE /api/listening/topics/:id` → topic and its mention associations are removed (soft-delete with 30-day retention).
8. User edits → `PUT /api/listening/topics/:id` → full replacement of keyword rules; re-queues the topic for ingestion with new rules.

**Topic data model:**
```
{
  id, workspace_id, name, color, type (own_brand|competitor|industry|custom),
  custom_type_id (nullable), keywords[], platforms[],
  ai_context_hint (text, max 200 chars, nullable),
  include_any[], include_all[], negative_terms[], negative_authors[],
  exact_match (bool), case_sensitive (bool),
  status (active|paused), created_at, updated_at
}
```

---

**Acceptance criteria:**

- [ ] `POST /api/listening/topics` creates a topic; returns 201 with full topic object
- [ ] `GET /api/listening/topics` returns all topics for the workspace (active + paused), ordered by created_at desc
- [ ] `PUT /api/listening/topics/:id` updates a topic; returns updated topic object
- [ ] `DELETE /api/listening/topics/:id` soft-deletes the topic; returns 204
- [ ] `PATCH /api/listening/topics/:id` supports `{ status: "paused" | "active" }` for pause/resume
- [ ] Creating a topic when workspace is at topic limit returns 422 with `error_code: "TOPIC_LIMIT_REACHED"`
- [ ] Creating a topic without name returns 422: `{ field: "name", message: "Topic name is required" }`
- [ ] Creating a topic without keywords returns 422: `{ field: "keywords", message: "Add at least one keyword" }`
- [ ] Creating a topic without platforms returns 422: `{ field: "platforms", message: "Select at least one platform" }`
- [ ] Keyword conflict validation returns structured errors per conflict type (not just a generic message)
- [ ] `ai_context_hint` is stored and included in responses; max 200 chars enforced
- [ ] Topic type can be `own_brand`, `competitor`, `industry`, or `custom` (with `custom_type_id`)
- [ ] Paused topics stop ingesting new mentions; existing mentions are still queryable
- [ ] Custom topic types CRUD: `GET/POST /api/listening/topic-types`, `PUT/DELETE /api/listening/topic-types/:id`; built-in types cannot be deleted

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `listening_topics` and `listening_custom_topic_types` records. No existing data affected.

**Impact on other products:** None.

**Dependencies:** **[BE] Data ingestion pipeline for 18-platform social listening** (topics must exist before ingestion begins).

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Topic names and keywords can contain any Unicode; API must handle correctly
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — No cross-product impact in V1

---

### Story 3: [BE] Mention query and filtering API with views support

**Description:**
As a backend system, I need to expose endpoints to query, filter, and paginate mentions — and to manage saved views (filter presets) — so that the feed and bookmarks UIs can render the right mentions for any filter combination.

---

**Workflow:**

1. Feed loads → `GET /api/listening/mentions` is called with query params.
2. Supported filters: `topic_ids[]`, `platforms[]`, `sentiments[]` (positive|neutral|negative), `ai_tags[]`, `min_followers`, `date_from`, `date_to`, `sort` (newest|oldest|most_engaged), `view_id`, `is_bookmarked`, `is_read`.
3. Pagination: cursor-based (not offset) for real-time feeds; default 25 per page.
4. Response includes each mention with: id, platform, author, content, url, published_at, topic, sentiment (AI-assigned + user-override), ai_tags, engagement, is_read, is_bookmarked, keyword_matches[] (for highlighting).
5. User marks mention as read → `PATCH /api/listening/mentions/:id` with `{ is_read: true }`.
6. User marks all as read → `POST /api/listening/mentions/mark-all-read` (with optional filters).
7. User bookmarks → `PATCH /api/listening/mentions/:id` with `{ is_bookmarked: true }`.
8. User marks irrelevant → `PATCH /api/listening/mentions/:id` with `{ is_irrelevant: true }` → excluded from feed queries by default.
9. User overrides sentiment → `PATCH /api/listening/mentions/:id` with `{ sentiment_override: "positive"|"neutral"|"negative" }` → stored and used in analytics.

**Views CRUD:**
- `GET /api/listening/views` — all views for workspace (system + user-created), ordered by type (system first) then name
- `POST /api/listening/views` — create view with name, icon, and filter preset
- `PUT /api/listening/views/:id` — update (user-created only)
- `DELETE /api/listening/views/:id` — delete (user-created only; system views return 403)
- System views are seeded per workspace on first Listening activation: "All Mentions" (no filters) and "High Relevance" (ai_tags: [own_brand_mention, buy_intent])

---

**Acceptance criteria:**

- [ ] `GET /api/listening/mentions` supports all documented filter params; returns paginated cursor-based results
- [ ] Irrelevant mentions (`is_irrelevant: true`) are excluded from feed results by default; can be included with `include_irrelevant=true`
- [ ] `keyword_matches[]` field in response contains the specific keyword(s) that caused the mention to match (for frontend highlighting)
- [ ] `PATCH /api/listening/mentions/:id` supports: `is_read`, `is_bookmarked`, `is_irrelevant`, `sentiment_override`
- [ ] `POST /api/listening/mentions/mark-all-read` marks all unread mentions as read (respecting any filter params passed)
- [ ] Unread count is returned in a separate `GET /api/listening/mentions/unread-count` endpoint
- [ ] Views CRUD endpoints enforce: system views cannot be edited or deleted (403 response)
- [ ] View filter preset stores: topic_ids[], platforms[], sentiments[], ai_tags[], min_followers, language (same shape as the mention filter params)
- [ ] `GET /api/listening/mentions?is_bookmarked=true` returns only bookmarked mentions (powers the Bookmarks page)
- [ ] Sentiment override is persisted to `sentiment_override` field; analytics queries use override when present, fallback to AI-assigned sentiment

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `listening_mentions` and `listening_views` collections. No existing data affected.

**Impact on other products:** None in V1.

**Dependencies:** **[BE] Data ingestion pipeline for 18-platform social listening**, **[BE] Topic management API with keyword rules and pause/resume**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Content field stores original language; language filter is applied at ingestion; API returns content as-is
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — No cross-product impact in V1

---

### Story 4: [BE] Smart tagging and sentiment scoring pipeline

**Description:**
As a backend system, I need to automatically classify every ingested mention with a sentiment score and up to N smart tags so that users can filter, prioritize, and analyze their mentions without reading every one manually.

---

**Workflow:**

1. A new mention is stored → it's queued to the AI processing pipeline (Dramatiq/Redis job).
2. The job sends the mention content + topic context (including the `ai_context_hint` if set) to the AI agents pipeline.
3. Sentiment classification: model assigns one of `positive`, `neutral`, `negative`.
4. Smart tag classification: model assigns 0–N tags from the 10-category taxonomy:
   - Own Brand Mention, Competitor Mention, Industry Insight, Buy Intent, Bug Report, User Feedback, Promotional Post, Product Question, Event, Hiring
5. Tags and sentiment are written back to the mention record.
6. If the AI job fails, the mention is stored without tags/sentiment (nullable fields); retry is attempted up to 3 times.
7. Batch reprocessing endpoint available for re-tagging existing mentions (e.g., if topic's `ai_context_hint` is updated): `POST /api/listening/topics/:id/reprocess-mentions`.

---

**Acceptance criteria:**

- [ ] Every ingested mention is processed for sentiment and smart tags within 60 seconds of ingestion (p95)
- [ ] Sentiment field is one of: `positive`, `neutral`, `negative`, `null` (unprocessed)
- [ ] `ai_tags` field is an array of 0–N values from the 10-tag taxonomy
- [ ] The `ai_context_hint` from the topic is included in the classification prompt to improve accuracy for ambiguous brand names
- [ ] Failed AI processing jobs are retried up to 3 times; after all retries fail, mention is stored with `ai_tags: []` and `sentiment: null`
- [ ] `POST /api/listening/topics/:id/reprocess-mentions` triggers batch reprocessing of all mentions for that topic
- [ ] Sentiment override by user (`sentiment_override` field) is never overwritten by the AI pipeline
- [ ] Unit tests verify the tag taxonomy mapping is complete and case-consistent

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** Adds `sentiment`, `ai_tags` fields to `listening_mentions`. No other data affected.

**Impact on other products:** Uses `contentstudio-ai-agents/` pipeline — coordinate with AI team on prompt design. AI features are web-only; no mobile impact.

**Dependencies:** **[BE] Data ingestion pipeline for 18-platform social listening**, **[BE] Mention query and filtering API with views support**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Classification model must handle English content correctly; non-English content may be tagged with reduced accuracy (acceptable for V1); document language support limitations
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — Uses AI agents pipeline; coordinate on token cost and rate limits

---

### Story 5: [BE] Listening alerts engine — volume spike and sentiment spike detection

**Description:**
As a backend system, I need to detect unusual spikes in mention volume or negative sentiment for each alert rule and send email notifications to configured recipients so that users are informed of potential PR issues in near-real-time.

---

**Workflow:**

1. A background job runs every 15 minutes for each workspace with active alert rules.
2. For each alert rule:
   - Loads the rule's associated view (filter preset) and its trigger configuration.
   - **Volume Spike:** Calculates the 7-day rolling average of mentions matching the view. If current hour's (or current period's) rate exceeds the average by the configured threshold (e.g., 50%), trigger fires.
   - **Sentiment Spike:** Calculates the % of negative mentions among all mentions in the last 1 hour (or configurable window). If it exceeds the configured threshold (e.g., 30%), trigger fires.
3. If trigger fires and the rule's `last_alerted_at` was more than 1 hour ago (cooldown), an alert email is sent to all recipients.
4. `last_alerted_at` is updated to prevent duplicate emails during the same spike.
5. Alert email contains: workspace name, view name, trigger type and threshold, mention count/sentiment %, and a deep link to ContentStudio Listening (pre-filtered to the triggering view).
6. Alert rules CRUD: `GET/POST /api/listening/alerts`, `PUT/DELETE /api/listening/alerts/:id`, `PATCH /api/listening/alerts/:id` (for toggle active/paused).

**Alert rule data model:**
```
{
  id, workspace_id, view_id,
  volume_spike_enabled (bool), volume_spike_threshold_pct (int, 10-500),
  sentiment_spike_enabled (bool), sentiment_spike_threshold_pct (int, 10-100),
  recipients_team_member_ids[], recipients_email_addresses[],
  status (active|paused), last_alerted_at, created_at
}
```

---

**Acceptance criteria:**

- [ ] Alert detection job runs every 15 minutes for each workspace with ≥1 active alert rule
- [ ] Volume spike: fires when mentions in the current window exceed the 7-day rolling average by the configured threshold %
- [ ] Sentiment spike: fires when negative mentions % of total (last 1 hour window) exceeds the configured threshold %
- [ ] Alert cooldown of 1 hour prevents duplicate emails for the same alert rule during a sustained spike
- [ ] Email is sent to all `recipients` (both team member emails and manually entered addresses)
- [ ] Email contains: workspace name, alert name, trigger description, and a deep-link URL to the Listening feed pre-filtered to the relevant view
- [ ] Alert rules with both triggers disabled cannot be saved (enforced at API level)
- [ ] Alert rules with no recipients cannot be saved (enforced at API level)
- [ ] Paused alert rules are skipped by the detection job
- [ ] Paused topics/views still generate historical data for lookback calculations
- [ ] `GET /api/listening/alerts` returns all alert rules with their current status and view details
- [ ] Alert CRUD endpoints return appropriate errors for invalid view IDs or missing fields

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `listening_alerts` table. Reuses existing transactional email infrastructure.

**Impact on other products:** Uses existing email delivery service. Deep-link URL format must be coordinated with frontend routing.

**Dependencies:** **[BE] Mention query and filtering API with views support**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Email templates should use English only for V1; i18n of emails is deferred
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — Deep-link URL in alert email should use the workspace's custom domain if on white-label
- [ ] Cross-product impact assessment — Uses transactional email system; ensure alert emails don't conflict with other system emails in rate limits

---

### Story 6: [BE] Analytics aggregation API for mentions, sentiment, platform, and tag data

**Description:**
As a backend system, I need to expose aggregated analytics endpoints that power the Listening Analytics dashboard — including KPI summaries, time-series charts, sentiment distribution, platform breakdown, and tag breakdown — so that users can track trends and understand their brand's online presence.

---

**Workflow:**

1. User visits Analytics tab → frontend calls `GET /api/listening/analytics/summary` with date range and topic filter.
2. Summary response returns: total mentions, positive sentiment %, average daily mentions, topics tracked count.
3. Frontend calls individual chart endpoints (or one combined endpoint):
   - `GET /api/listening/analytics/volume-over-time` → array of `{ date, topic_id, count }` for line chart (one series per topic)
   - `GET /api/listening/analytics/sentiment-trend` → array of `{ date, positive, neutral, negative }` for stacked area chart
   - `GET /api/listening/analytics/sentiment-distribution` → `{ positive_pct, neutral_pct, negative_pct }` for donut
   - `GET /api/listening/analytics/by-platform` → array of `{ platform, count }` sorted desc for donut + list
   - `GET /api/listening/analytics/by-tag` → array of `{ tag, count }` sorted desc for bar chart
4. All endpoints accept: `date_from`, `date_to`, `topic_ids[]` (empty = all topics).
5. `GET /api/listening/analytics/ai-insights` → given a chart_type and current data, returns 3 AI-generated insight strings (used by each chart's "AI Insights" button).

---

**Acceptance criteria:**

- [ ] All analytics endpoints accept `date_from`, `date_to`, `topic_ids[]` params
- [ ] Summary endpoint returns: `total_mentions`, `positive_sentiment_pct`, `avg_daily_mentions`, `topics_tracked_count`
- [ ] Volume-over-time endpoint returns daily granularity within the date range; one series per topic
- [ ] Sentiment-trend endpoint returns daily positive/neutral/negative counts; uses `sentiment_override` when present, falls back to AI sentiment
- [ ] By-platform returns all 18 platforms (with 0 counts for platforms with no mentions) or only active platforms — sorted by count desc
- [ ] By-tag returns all 10 smart tag categories with their counts
- [ ] AI insights endpoint accepts `chart_type` (volume_over_time|sentiment_trend|sentiment_distribution|by_platform|by_tag) and returns 3 string insights
- [ ] All endpoints return empty/zero data gracefully (no 500 errors on empty date ranges)
- [ ] Sentiment % calculations use user-overridden sentiment where present
- [ ] Response time ≤ 2s for typical date ranges (up to 90 days) with up to 5 topics

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New aggregated analytics queries over `listening_mentions`. No writes to existing data.

**Impact on other products:** AI insights endpoint uses `contentstudio-ai-agents/` pipeline.

**Dependencies:** **[BE] Smart tagging and sentiment scoring pipeline**, **[BE] Mention query and filtering API with views support**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Analytics data is numeric; AI insights are English-only for V1
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — Uses AI agents pipeline for insights; coordinate on token budget

---

### Story 7: [BE] Export service — PDF/CSV, scheduled email reports, and shareable analytics links

**Description:**
As a backend system, I need to support three export modes — immediate download (PDF/CSV), scheduled recurring email reports, and read-only shareable analytics links — so that users can share Listening analytics with clients and stakeholders.

---

**Workflow:**

1. **Download:** `POST /api/listening/exports/download` with `{ format: "pdf"|"csv", date_from, date_to, topic_ids[] }` → generates file and returns download URL (or streams directly). PDF includes charts rendered server-side; CSV contains raw mention data.
2. **Schedule:** `POST /api/listening/exports/schedules` with `{ frequency: "weekly"|"monthly", day_of_week|day_of_month, time, recipients: { team_member_ids[], email_addresses[] } }` → creates a scheduled report job. Reports are sent on the configured schedule to all recipients.
3. **Share link:** `POST /api/listening/exports/share-links` with `{ date_range_type, topic_ids[] }` → creates a read-only link valid for 30 days. Link opens a simplified analytics view with no edit controls. `DELETE /api/listening/exports/share-links/:token` revokes the link.
4. Scheduled report job runs on its configured cadence, generates a PDF of the current analytics for the configured date range and topic scope, and emails it to recipients.

---

**Acceptance criteria:**

- [ ] `POST /api/listening/exports/download` with `format: "pdf"` returns a PDF analytics report with all 5 charts and 4 KPI cards rendered
- [ ] `POST /api/listening/exports/download` with `format: "csv"` returns a CSV with one row per mention: date, platform, author, content (truncated to 500 chars), sentiment, tags, topic, URL
- [ ] Scheduled report jobs fire within 5 minutes of their configured time
- [ ] Scheduled report emails are sent to all recipients (team members + manual addresses)
- [ ] Share links expire exactly 30 days after creation
- [ ] Share link endpoint returns a URL-safe token; accessing the URL with an expired token returns a 410 Gone response with explanatory message
- [ ] Share link analytics view is read-only (no filtering, no export, no edit controls)
- [ ] `DELETE /api/listening/exports/share-links/:token` revokes the link immediately (subsequent requests return 410)
- [ ] Scheduled report CRUD: GET (list), POST (create), PUT (update), DELETE (delete)

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `listening_export_schedules` and `listening_share_links` tables.

**Impact on other products:** Uses existing email infrastructure and PDF generation service.

**Dependencies:** **[BE] Analytics aggregation API for mentions, sentiment, platform, and tag data**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — PDF/email templates in English only for V1
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — Share links should use workspace's custom domain on white-label; PDF report header should use workspace name/logo
- [ ] Cross-product impact assessment — PDF generation may require a headless browser service (Puppeteer); ensure this is provisioned

---

### Story 8: [BE] Global settings and topic types API

**Description:**
As a backend system, I need to persist and expose workspace-level global filter settings and custom topic type management so that users can configure system-wide mention filtering rules that apply across all topics.

---

**Workflow:**

1. `GET /api/listening/settings` returns current workspace global settings: `{ negative_terms[], blocked_authors[], excluded_subreddits[], language_filter[] }`.
2. `PUT /api/listening/settings` replaces the full settings object (all fields optional; omitted fields reset to empty).
3. Settings are applied at ingestion time (Story 1) — before a mention is stored, all global filters are checked.
4. Custom topic types: `GET /api/listening/topic-types` returns built-in (read-only) + custom types. `POST` creates a custom type. `PUT/:id` renames. `DELETE/:id` removes (if no topics are using it; if topics reference it, return 422 with list of affected topics).

---

**Acceptance criteria:**

- [ ] `GET /api/listening/settings` returns current settings; returns empty arrays for unset fields (never null)
- [ ] `PUT /api/listening/settings` persists all four fields; changes take effect on the next crawl cycle
- [ ] `negative_terms`, `blocked_authors`, `excluded_subreddits` are arrays of strings; `language_filter` is an array of ISO 639-1 language codes
- [ ] Built-in topic types (own_brand, competitor, industry_term) are returned from `GET /api/listening/topic-types` with `is_builtin: true` and cannot be deleted or renamed
- [ ] Custom topic type deletion returns 422 if any active topics reference it, with the list of affected topic names in the error response
- [ ] Custom topic type names must be unique within the workspace; duplicate name returns 422
- [ ] Settings changes are logged for auditing (workspace_id, changed_by, changed_at)

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `listening_workspace_settings` record per workspace. New `listening_custom_topic_types` records.

**Impact on other products:** None.

**Dependencies:** **[BE] Topic management API with keyword rules and pause/resume**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — Topic type names can contain any Unicode; language_filter accepts ISO 639-1 codes
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — None

---

### Story 9: [BE] Listening add-on subscription management and usage limits

**Description:**
As a backend system, I need to track each workspace's Listening subscription state and enforce topic/mention limits so that the add-on is correctly gated by plan, and usage is measured and capped accurately.

---

**Workflow:**

1. Listening add-on is activated for a workspace → `listening_subscription` record created with: `status` (active|paused|expired), `activated_at`, `topic_limit` (default 5), `mention_limit_monthly` (default 10,000).
2. Monthly mention counter resets on the subscription renewal date (not the 1st of month).
3. At each ingestion cycle, the running monthly mention count is checked before storing new mentions. If `mentions_this_month >= mention_limit_monthly`, new mentions are discarded and the workspace is flagged.
4. Usage endpoints:
   - `GET /api/listening/usage` returns: `{ topics_count, topics_limit, mentions_this_month, mention_limit_monthly, reset_date, status }`
5. Subscription status affects the user state exposed to the frontend:
   - `no_subscription` → locked (paid plan) or trial
   - `active` (new/first time, not yet set up) → unlocked_new
   - `active` (set up) → unlocked
   - `expired` or `cancelled` → expired
6. Add-on limits add-ons: `PATCH /api/listening/subscription/limits` (admin/billing system only) to update `topic_limit` and `mention_limit_monthly` when a user purchases add-ons.

---

**Acceptance criteria:**

- [ ] `GET /api/listening/usage` returns accurate counts updated in real-time (or within 1 minute of last ingestion)
- [ ] Monthly mention counter resets on the subscription renewal date, not the calendar month start
- [ ] When `mentions_this_month >= mention_limit_monthly`, new mentions are not stored; existing mentions are unaffected
- [ ] Workspace topic count is decremented when a topic is deleted and incremented when one is created
- [ ] `status` values returned are: `no_subscription`, `trial`, `active`, `expired`
- [ ] `PATCH /api/listening/subscription/limits` requires admin/billing auth; updates limits without affecting existing data
- [ ] At 90% mention usage, a flag `approaching_limit: true` is returned in usage response (for frontend warning UI)
- [ ] Subscription expiry is detected via webhook from the billing system; `status` is updated to `expired` within 5 minutes of expiry event

---

**Mock-ups:** N/A — backend only

**Impact on existing data:** New `listening_subscriptions` table. Integration with billing system.

**Impact on other products:** Billing system integration required. Billing team must be notified to add the Listening add-on SKU.

**Dependencies:** None (can be built in parallel with other BE stories; must be ready before FE subscription-gating stories).

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — N/A
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review — White-label workspaces follow the same add-on model
- [ ] Cross-product impact assessment — Billing system integration; coordinate with billing team on SKU setup

---

### Story 10: [FE] Listening module shell — navigation, routing, and user state management

**Description:**
As a ContentStudio user on any plan, I want to click "Listening" in the top navigation and be taken to the right experience for my subscription state so that I can either explore the feature, activate it, or use it if I'm already subscribed.

---

**Workflow:**

1. User clicks "Listening" in the ContentStudio global top nav bar.
2. The Listening module loads. Frontend fetches `GET /api/listening/usage` to determine `userState`.
3. Based on `userState`, the correct view renders:
   - `trial` → LandingPage (trial variant)
   - `no_subscription` (paid plan) → LandingPage (locked variant)
   - `expired` → LandingPage (expired variant)
   - `active` (first setup needed) → SetupWizard
   - `active` (setup complete) → Main feed app (Feed section, last-used view)
4. The module shell renders:
   - **PrimaryNav** (200px left panel): Listening logo + wordmark, section tabs (Feed / Bookmarks / Analytics / Alerts / Settings), Topics list with context menus, usage bar (active subscribers only).
   - **ViewsSidebar** (210px, only when Feed section is active): "For You" system views + user-created views.
   - **Main content area**: renders whichever section is active.
5. On mobile (< 768px), PrimaryNav is hidden by default. A hamburger icon in the content area header opens PrimaryNav as a full-screen Drawer overlay.
6. The ViewsSidebar is collapsible: a toggle button collapses it; when collapsed, the Feed nav item shows an expand hint icon. On mobile, the ViewsSidebar renders as a Drawer.
7. Navigating to a different section (e.g., Analytics) closes the ViewsSidebar on mobile.

---

**UI Copy:**

**PrimaryNav header:**
- Logo icon + wordmark: "Listening"

**Section navigation labels:**
- "Feed" (Rss icon)
- "Bookmarks" (Bookmark icon)
- "Analytics" (BarChart3 icon)
- "Alerts" (Bell icon)
- "Settings" (Settings icon)

**Topics section header:** "TOPICS" (uppercase, muted)
- "+" tooltip: "Add topic"
- No topics state: "No topics yet"
- Topic context menu: "Edit" / "Pause" / "Resume" / "Delete"

**Collapse/expand:**
- Collapse tooltip (ViewsSidebar): "Hide views panel"
- Expand hint tooltip (Feed nav icon): "Show views panel"

**Mobile nav:**
- Hamburger opens full Drawer; "×" closes it

---

**Acceptance criteria:**

- [ ] Clicking "Listening" in the top nav always works regardless of subscription state — no 404 or disabled state
- [ ] `userState` is determined from `GET /api/listening/usage` on every page load; loading state shows skeleton
- [ ] Trial users see LandingPage trial variant; no path into the main app
- [ ] Paid plan, no subscription → LandingPage locked variant
- [ ] Expired → LandingPage expired variant
- [ ] Active subscription, no topics yet → SetupWizard
- [ ] Active subscription, ≥1 topic → Main feed (Feed section, "All Mentions" view default)
- [ ] PrimaryNav is always visible on desktop (≥768px); hidden behind hamburger on mobile
- [ ] ViewsSidebar is only present when "Feed" is the active section
- [ ] ViewsSidebar collapses and expands via toggle button; state persists within the session
- [ ] Active section tab is highlighted with `bg-primary-cs-50` background and `text-primary-cs-500` text
- [ ] Topics list renders a colored dot + name + optional PauseCircle icon per topic
- [ ] "···" context menu on topic shows: Edit, Pause (if active) or Resume (if paused), Delete
- [ ] Pause action opens confirmation modal before executing (see Story 15 for copy)
- [ ] Delete action executes immediately with no confirmation modal
- [ ] PrimaryNav usage bar is visible only when `userState = active` (see Story 22)
- [ ] Mobile Drawer closes when a section or topic is selected

---

**Mock-ups:** See prototype at https://cs-prototypes.vercel.app/features/listening2

**Impact on existing data:** None. New frontend module; no existing data modified.

**Impact on other products:** TopBar "Listening" nav item already exists. No other cross-product impact.

**Dependencies:** **[BE] Listening add-on subscription management and usage limits**, **[BE] Topic management API with keyword rules and pause/resume**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Full mobile support required: hamburger nav, drawer overlay, stacked layouts
- [ ] Multilingual support — All labels must use i18n strings; topic names can be any Unicode
- [ ] UI theming support — Active state uses `bg-primary-cs-50` / `text-primary-cs-500`; never hardcode `#7C3AED` or Tailwind blue classes
- [ ] White-label domains impact review — Module layout must work on white-label domains
- [ ] Cross-product impact assessment — TopBar navigation is shared; ensure "Listening" item styling is consistent with other nav items

---

### Story 11: [FE] Listening landing page for trial, locked, and expired states

**Description:**
As a ContentStudio user who doesn't yet have the Listening add-on, I want to see a clear, compelling page explaining what Listening does, how much it costs, and how to activate it so that I can decide to subscribe.

---

**Workflow:**

1. User without Listening access clicks "Listening" in the top nav → LandingPage renders.
2. Three variants based on `userState`:
   - **Trial:** Badge "Add-on · $49/mo" (violet), headline, trial gate banner (orange), pricing card with disabled CTA.
   - **Locked:** Badge "Add-on · $49/mo" (violet), headline, Enable + Preview Demo CTAs active, pricing card with active CTA.
   - **Expired:** Badge "Add-on Expired" (orange), different headline, Re-enable + Preview Demo CTAs, pricing card.
3. "Enable Listening" / "Re-enable Listening" CTA → navigates to the billing/upgrade flow (existing CS billing page with the Listening add-on pre-selected).
4. "Preview Demo" → loads the main feed with mock data and no subscription required (demo mode).
5. Pricing card "Add to ContentStudio" → same billing flow as the primary CTA.
6. Trial users' CTAs ("Upgrade Plan First", "Upgrade Plan") → navigate to plan upgrade flow (not the Listening add-on flow).

---

**UI Copy:**

**Trial variant:**
- Badge: "Add-on · $49/mo" (violet)
- Headline: "Never miss a mention that matters"
- Subtitle: "Track your brand, competitors, and industry keywords across 18 platforms — all in one intelligent feed, inside ContentStudio."
- Trial gate banner headline: "Not available on trial plans"
- Trial gate banner body: "Listening is a paid add-on. Upgrade to any ContentStudio plan to unlock it."
- Trial gate CTA: "Upgrade Plan" →
- Pricing card CTA: "Upgrade Plan First" (orange, no billing link — upgrade plan first)
- Pricing card footer: "Upgrade to any paid plan to enable add-ons."

**Locked variant:**
- Badge: "Add-on · $49/mo" (violet)
- Headline: "Never miss a mention that matters"
- Subtitle: same as above
- Primary CTA: "Enable Listening"
- Secondary CTA: "Preview Demo" (with PlayCircle icon)
- Pricing card CTA: "Add to ContentStudio" →
- Pricing card footer: "Cancel anytime from your billing settings."

**Expired variant:**
- Badge: "Add-on Expired" (orange)
- Headline: "Re-enable Listening"
- Subtitle: "Your Listening add-on has expired. Re-enable it to resume monitoring your brand, competitors, and industry keywords across 18 platforms."
- Primary CTA: "Re-enable Listening"
- Secondary CTA: "Preview Demo"
- Pricing card CTA: "Add to ContentStudio" →
- Pricing card footer: "Cancel anytime from your billing settings."

**Feature highlights (3 cards, all variants):**
- Card 1: Icon Rss — "Feed-First" — "Every mention from 18 platforms in one clean, prioritized feed."
- Card 2: Icon Sparkles — "AI-Powered" — "Smart-tagged, sentiment-scored, and noise-filtered automatically."
- Card 3: Icon Bell — "Instant Alerts" — "Get notified the moment something important happens — email alerts when volume spikes."

**Pricing card:**
- Title: "Listening Add-on"
- Price: $49 / month
- Subtext: "Billed to your existing ContentStudio subscription"
- Feature list (with checkmarks):
  - 5 topics / keywords included
  - 10,000 mentions / month
  - 18 platforms monitored
  - Smart tagging + sentiment analysis
  - Custom views & filters
  - Email alerts
  - Add-ons: +topics & +mentions available

---

**Acceptance criteria:**

- [ ] Trial variant shows the orange gate banner; "Enable Listening" and "Add to ContentStudio" CTAs are styled orange and link to plan upgrade (not Listening add-on)
- [ ] Locked variant shows violet badge, active "Enable Listening" CTA → billing flow, active "Preview Demo" → demo mode
- [ ] Expired variant shows orange "Add-on Expired" badge, "Re-enable Listening" headline, CTA → billing flow
- [ ] "Preview Demo" button on locked/expired loads the main feed in demo mode with mock data (no billing required)
- [ ] Feature highlight cards render correctly on mobile (stacked, single column)
- [ ] Pricing card renders with all 7 feature list items + checkmark icons
- [ ] Page is fully responsive: single column on mobile, max-width centered on desktop
- [ ] Pricing card CTA is disabled + styled orange for trial; enabled + styled `bg-primary-cs-500` for locked/expired
- [ ] Webhook is NOT mentioned in the feature card copy (email alerts only for V1)
- [ ] "Cancel anytime from your billing settings." text links to billing settings page

---

**Mock-ups:** See prototype at https://cs-prototypes.vercel.app/features/listening2 (switch DebugToggle to trial / locked / expired)

**Impact on existing data:** None.

**Impact on other products:** Links to existing billing/upgrade flow. Ensure billing flow supports Listening add-on SKU.

**Dependencies:** **[FE] Listening module shell — navigation, routing, and user state management**, **[BE] Listening add-on subscription management and usage limits**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Fully responsive; CTA buttons stack vertically on mobile
- [ ] Multilingual support — All copy must be in i18n strings
- [ ] UI theming support — Primary CTA uses `bg-primary-cs-500`; trial CTA uses warning orange (non-themed, intentional)
- [ ] White-label domains impact review — $49/mo pricing is ContentStudio-specific; white-label pricing must be configurable
- [ ] Cross-product impact assessment — Links to billing flow; ensure billing page is ready to handle Listening add-on

---

### Story 12: [FE] Onboarding setup wizard with AI-powered topic suggestions

**Description:**
As a first-time Listening subscriber, I want a guided setup wizard that analyzes my website to suggest topics automatically so that I can start monitoring my brand without needing to manually type every keyword from scratch.

---

**Workflow:**

1. User activates Listening (or has active subscription with no topics yet) → SetupWizard renders.
2. **Step 1 — Website Input:**
   - User sees a card with a globe icon, headline, and URL input field.
   - User types their website URL (e.g., `https://contentstudio.io`) and clicks "Analyze Website".
   - Loading animation runs in two phases: "Detecting brand & competitors..." (0.9s) then "Generating topic suggestions..." (1.1s).
   - On completion, Step 2 loads with AI-generated topic suggestions pre-populated.
   - Alternatively, user clicks "Skip, I'll add topics manually →" → goes to Step 2 with empty list + TopicFormModal opens immediately.
3. **Step 2 — Topics Review:**
   - If topics were AI-generated, a banner reads: "✦ Our AI suggested these based on your website"
   - Each topic is shown as a card with: colored left border, topic name, keyword pills (up to 4, "+N more" if >4), a type selector dropdown (Own Brand / Competitor / Industry Term / custom), Edit ✏ icon, Delete 🗑 icon.
   - Edit → opens TopicFormModal pre-filled (see Story 15).
   - Delete → removes topic card immediately.
   - "Add another topic" (dashed button at bottom) → opens empty TopicFormModal.
   - "Start Monitoring →" button is disabled if no topics remain.
   - Clicking "Start Monitoring →" → calls `POST /api/listening/topics` for any un-saved topics, then transitions to the main feed.

---

**UI Copy:**

**Step 1:**
- Headline: "Set up your monitoring"
- Subtext: "Enter your website and we'll detect your brand, competitors, and suggest topics automatically."
- Input label: "Your website"
- Input placeholder: "https://yourcompany.com"
- Primary CTA: "Analyze Website →"
- Secondary link: "Skip, I'll add topics manually →"
- Loading phase 1 text: "Detecting brand & competitors..."
- Loading phase 2 text: "Generating topic suggestions..."

**Step 2:**
- Headline: "Your topics"
- AI banner: "✦ Our AI suggested these based on your website"
- Topic type selector options: "Own Brand", "Competitor", "Industry Term" + any custom types
- Edit icon tooltip: "Edit topic"
- Delete icon tooltip: "Remove topic"
- Add button: "+ Add another topic"
- Primary CTA: "Start Monitoring →"
- Disabled state tooltip: "Add at least one topic to start monitoring"
- Empty state (all deleted): "No topics yet — add one to start monitoring."

---

**Acceptance criteria:**

- [ ] Step 1 URL input validates that the entered value is a valid URL; shows inline error "Please enter a valid website URL (e.g. https://yourcompany.com)" for invalid inputs
- [ ] "Analyze Website" button triggers the two-phase loading animation; minimum 2 seconds of animation even if AI responds faster
- [ ] AI-suggested topics are displayed in Step 2 with the AI banner visible
- [ ] "Skip" link goes directly to Step 2 with an empty topic list and TopicFormModal open
- [ ] Each topic card shows: colored border (type-colored), name, keyword pills (max 4 visible, "+N more" badge if overflow), type selector
- [ ] Type selector dropdown shows built-in types + any custom types; selection updates the topic's type
- [ ] Edit icon opens TopicFormModal pre-filled with that topic's data
- [ ] Delete icon removes the topic card from the list without confirmation
- [ ] "Add another topic" opens an empty TopicFormModal
- [ ] "Start Monitoring →" is disabled when topic list is empty
- [ ] Clicking "Start Monitoring →" saves all topics via API and navigates to the Feed

---

**Mock-ups:** See prototype at https://cs-prototypes.vercel.app/features/listening2 (DebugToggle → "New Subscriber")

**Impact on existing data:** Creates topics for the workspace. No existing data affected.

**Impact on other products:** Uses AI pipeline for website analysis and topic suggestion.

**Dependencies:** **[FE] Listening module shell — navigation, routing, and user state management**, **[BE] Topic management API with keyword rules and pause/resume**, **[FE] Topic management UI — create, edit, pause, and delete topics**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Wizard cards stack and are readable on mobile; topic cards wrap correctly
- [ ] Multilingual support — All copy in i18n; website URL and topic names can be any Unicode
- [ ] UI theming support — Primary CTAs use `bg-primary-cs-500`; AI banner uses themed light background
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — Uses AI pipeline; AI features are web-only (no mobile equivalent for this story)

---

### Story 13: [FE] Mention feed with filter bar, mention cards, and action pill

**Description:**
As a Listening subscriber, I want to see all my mentions in a scrollable, filterable feed so that I can quickly scan what's happening across platforms and take action on the mentions that matter.

---

**Workflow:**

1. User is in the Feed section → active view's filter preset is applied → mentions load.
2. Filter bar (horizontal, scrollable on mobile) renders chips:
   - **Topic** chip → multi-select popover (all topics with colored dots)
   - **Sentiment** chip → multi-select popover: Positive / Neutral / Negative
   - **Tags** chip → multi-select popover (all 10 smart tags with colored dots)
   - **Min Followers** chip → single-select: Any reach / 500+ / 1k+ / 5k+ / 10k+ / 50k+ / 100k+
   - **Date Range** chip → presets (All time / Last 24h / Last 7d / Last 30d / Last 90d) + custom date range picker
   - **Sort** chip → single-select: Newest first / Oldest first / Most engaged
   - **Mark all read** button → appears only when unread mentions exist; marks all as read after confirmation
3. Mention cards render in chronological order (default: newest first).
4. Each mention card shows: platform icon, author avatar, author name, handle, time (relative: "2h ago"), topic badge (colored), sentiment badge (color-coded), AI tag pills, engagement stats (likes / comments / shares), content preview (3 lines, expanded on click), matched keywords highlighted in content.
5. Unread mentions have a subtle left border indicator.
6. Hover on card → floating action pill appears (top-right of card): Reply · Bookmark · Open ↗ · Mark read · Mark irrelevant · ··· (more)
7. "Open ↗" → opens the original post URL in a new tab.
8. "Mark as read" → removes unread indicator immediately.
9. "Mark as irrelevant" → hides the card from the feed with a brief fade-out animation. A brief undo banner: "Marked as irrelevant. Undo?"
10. "Copy link" (in ··· more menu) → copies the mention URL to clipboard; toast: "Link copied!"
11. Bookmark action → toggles bookmark (filled icon = bookmarked); reflected immediately.
12. Sentiment badge on card is clickable → opens a dropdown: Positive / Neutral / Negative (current selection highlighted). Selecting a different value calls the override API.
13. **Empty states:**
    - No mentions at all: icon + "No mentions yet" + "Your topics are being set up. Mentions will appear here as they're detected, usually within the first 30–60 minutes."
    - Filters active, no results: "No mentions match your filters" + "Try adjusting your filters or clearing the date range." + "Clear filters" button.

---

**UI Copy:**

**Filter chips:**
- "Topic" chip label; popover header: "Filter by Topic"
- "Sentiment" chip label; popover header: "Filter by Sentiment"; options: "Positive", "Neutral", "Negative"
- "Tags" chip label; popover header: "Filter by Tag"
- "Min Followers" chip label; popover header: "Minimum Reach"; options: "Any reach", "500+ followers", "1,000+ followers", "5,000+ followers", "10,000+ followers", "50,000+ followers", "100,000+ followers"
- "Date Range" chip; presets: "All time", "Last 24 hours", "Last 7 days", "Last 30 days", "Last 90 days", "Custom range"
- "Sort" chip; options: "Newest first", "Oldest first", "Most engaged"
- "Mark all read" button; confirmation tooltip: "Mark all visible mentions as read?"

**Action pill:**
- "Reply" (send icon) — only on supported platforms
- "Bookmark" (bookmark icon, fills when active) — tooltip: "Save to bookmarks"
- "Open" (external link icon) — tooltip: "Open original post"
- "Mark as read" — tooltip: "Mark as read"
- "Mark as irrelevant" — tooltip: "Hide from feed"
- "Copy link" — tooltip: "Copy link to this post"

**Sentiment badge options:**
- "Positive", "Neutral", "Negative"
- Override tooltip: "This sentiment was detected automatically. Click to correct it."

**Reply icon on unsupported platform:**
- Tooltip: "Replies aren't available for [Platform Name] posts"

**Undo banner:**
- "Marked as irrelevant. Undo?" — auto-dismisses after 5 seconds

**Empty state (no mentions yet):**
- Headline: "No mentions yet"
- Subtext: "Your topics are being set up. Mentions will appear here as they're detected, usually within the first 30–60 minutes."

**Empty state (filters active, no results):**
- Headline: "No mentions match your filters"
- Subtext: "Try adjusting your filters or clearing the date range."
- CTA: "Clear all filters"

---

**Acceptance criteria:**

- [ ] Filter bar renders with all 6 filter chips + sort + mark-all-read (latter only when unread exist)
- [ ] Each filter chip opens its popover on click and closes on outside click or re-click
- [ ] Multi-select chips (Topic, Sentiment, Tags) show count badge when ≥1 filter is active (e.g., "Tags · 3")
- [ ] Filters are applied cumulatively (AND logic between different filter types)
- [ ] Sort chip changes the order of mentions in the feed without full page reload
- [ ] "Mark all read" marks all currently-visible mentions as read; unread indicator disappears
- [ ] Mention card shows correct: platform icon, relative time, topic badge, sentiment badge (color-coded), AI tag pills, engagement counts
- [ ] Matched keyword(s) are highlighted in the mention content text
- [ ] Action pill appears on card hover (desktop) and is always visible on mobile (tap to toggle)
- [ ] "Open" link opens the original URL in a new browser tab
- [ ] "Mark as irrelevant" hides the card with fade animation + shows 5-second undo banner; "Undo" restores it
- [ ] Bookmark toggle updates instantly; persists via API call
- [ ] Sentiment badge click opens dropdown; selecting a value calls `PATCH /api/listening/mentions/:id` with `sentiment_override`; badge updates immediately
- [ ] Reply icon is at 35% opacity on unsupported platforms and shows the correct tooltip; clicking it does nothing
- [ ] Both empty states render correctly with accurate copy
- [ ] Feed is paginated (infinite scroll or "Load more"); next page loads when user scrolls to bottom
- [ ] Loading state: skeleton cards (3–5) while initial data loads

---

**Mock-ups:** See prototype at https://cs-prototypes.vercel.app/features/listening2 (Feed section)

**Impact on existing data:** None.

**Impact on other products:** None direct; mention card reply action is covered by Story 14.

**Dependencies:** **[BE] Mention query and filtering API with views support**, **[FE] Listening module shell — navigation, routing, and user state management**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Filter bar horizontally scrollable on mobile; action pill always visible on touch devices; cards are full-width on mobile
- [ ] Multilingual support — Mention content renders as-is (any language); all UI labels in i18n
- [ ] UI theming support — Topic badges use `bg-primary-cs-50 / text-primary-cs-500`; active filter chips use themed border/background
- [ ] White-label domains impact review — No specific impact
- [ ] Cross-product impact assessment — No cross-product impact

---

### Story 14: [FE] Inline mention reply with account selector and AI compose toolbar

**Description:**
As a Listening subscriber, I want to reply to social mentions directly from the feed without leaving ContentStudio, with AI assistance to draft or improve my reply, so that I can engage quickly and consistently with my audience.

---

**Workflow:**

1. User clicks "Reply" on a mention card (supported platforms only).
2. An inline reply panel expands below the card's content area (card stays visible above).
3. **Account selector:** Shows all connected ContentStudio accounts for that platform. If the user has multiple accounts, they can select which one replies.
   - Expired account: shown in list with ⚠ icon; tooltip: "This account's connection has expired. Reconnect it in Settings → Connected Accounts to reply."
   - No accounts: orange banner replaces the selector: "No [Platform Name] accounts connected. Connect an account to reply from ContentStudio." with a "Connect Account →" link.
4. **Textarea:** Placeholder "Write a reply..." Below textarea: character count where relevant (e.g., Twitter 280 chars).
5. **AI Compose toolbar** (always visible below textarea):
   - "✦ Write with AI" → sends the mention content + context to AI → generates a reply draft and fills the textarea (1.2s animation).
   - If textarea has text: additional buttons appear: "Rephrase", "Improve", "Shorten", "Lengthen", "Grammar Fix" — each transforms the existing text via AI.
6. **Footer:** "Cancel" (clears + closes panel) | "Reply" (primary, disabled if no text or no valid account selected).
7. User clicks "Reply" → API call to send via connected account → button shows spinner → success: button turns green "Sent! ✓" for 1.2s → panel collapses.
8. Error (expired token on send): red inline error below textarea: "This account's connection has expired. Please reconnect it in Settings → Connected Accounts."

---

**UI Copy:**

**Panel header:**
- "Reply as" label above account selector

**Account selector:**
- Default placeholder: "Select account"
- Expired account tooltip: "This account's connection has expired. Reconnect it in Settings → Connected Accounts to reply."

**No accounts banner:**
- "No [Platform Name] accounts connected. Connect an account to reply from ContentStudio."
- CTA link: "Connect Account →" (links to Settings → Connected Accounts)

**Textarea:**
- Placeholder: "Write a reply..."
- Character counter: shown for Twitter (280), threads (500); hidden for platforms with no limit

**AI toolbar:**
- "✦ Write with AI" — tooltip: "Let AI draft a reply based on this mention. You can edit it before sending."
- "Rephrase" — tooltip: "Rewrite your reply in different words while keeping the same meaning."
- "Improve" — tooltip: "Enhance your reply to sound more professional and engaging."
- "Shorten" — tooltip: "Make your reply more concise without losing the key message."
- "Lengthen" — tooltip: "Expand your reply with more detail or context."
- "Grammar Fix" — tooltip: "Fix any spelling or grammar errors in your reply."

**Buttons:**
- Cancel: "Cancel"
- Send: "Reply" (idle) → spinner → "Sent! ✓" (success, green, 1.2s) → collapses

**Error:**
- Token expired on send: "This account's connection has expired. Please reconnect it in Settings → Connected Accounts."
- Send failed (generic): "Couldn't send your reply. Please try again."

---

**Acceptance criteria:**

- [ ] Reply panel opens inline below the mention card when "Reply" is clicked
- [ ] Account selector shows only accounts connected to the mention's platform
- [ ] Expired accounts appear in the selector with ⚠ icon and the expired tooltip
- [ ] Selecting an expired account does not disable the Reply button immediately, but clicking Reply shows the expired error
- [ ] "No accounts" state shows the orange banner with the connect link; Reply button is disabled
- [ ] "Write with AI" generates a draft in ≤ 2 seconds; draft fills the textarea; spinner shown during generation
- [ ] Rephrase/Improve/Shorten/Lengthen/Grammar Fix buttons only appear when textarea has content
- [ ] Each AI action transforms the existing textarea content in-place
- [ ] Character counter appears for Twitter (280 limit, turns red at 280); hidden for other platforms
- [ ] Reply button is disabled when textarea is empty or no valid account is selected
- [ ] Clicking Reply shows a spinner on the button during the API call
- [ ] Successful send shows green "Sent! ✓" for 1.2 seconds then closes the panel
- [ ] Token-expired error (detected at send time) shows the red inline error message
- [ ] "Cancel" closes the panel and clears the textarea content
- [ ] AI features (Write with AI, Rephrase, etc.) are web-only; no mobile implementation

---

**Mock-ups:** See prototype at https://cs-prototypes.vercel.app/features/listening2 (hover a mention → Reply)

**Impact on existing data:** Reply is sent via existing connected account OAuth tokens.

**Impact on other products:** Uses existing connected accounts system. AI compose uses `contentstudio-ai-agents/` pipeline.

**Dependencies:** **[FE] Mention feed with filter bar, mention cards, and action pill**, **[BE] Mention query and filtering API with views support**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Reply panel expands below card on mobile; AI toolbar wraps to second row if needed
- [ ] Multilingual support — Reply textarea accepts any Unicode; AI generates English replies (V1)
- [ ] UI theming support — Primary Reply button uses `bg-primary-cs-500`; AI buttons use `text-primary-cs-500` border styling
- [ ] White-label domains impact review — No impact
- [ ] Cross-product impact assessment — Uses connected accounts (same as Publisher/Inbox); uses AI agents pipeline (web-only)

---

### Story 15: [FE] Topic management UI — create, edit, pause, and delete topics

**Description:**
As a Listening subscriber, I want to create and manage monitoring topics with precise keyword rules so that I can control exactly what mentions I receive and avoid noise.

---

**Workflow:**

1. User clicks "+" in the Topics section of PrimaryNav → TopicFormModal opens (empty, create mode).
2. User clicks "···" → Edit on an existing topic → TopicFormModal opens (pre-filled, edit mode).
3. **TopicFormModal fields (in order):**
   - Color swatch (click to open 8-color picker) + Name input (inline)
   - Topic Type selector (Own Brand / Competitor / Industry Term / custom types / "+ New type")
   - Keywords tag input (type a keyword + Enter/comma to add)
   - Platforms multi-select (all 18 platforms; "Select all" shortcut at top)
   - AI Context Hint textarea (optional, 200 char max)
   - ▼ Advanced Settings accordion (collapsed by default):
     - Include ANY of these terms (tag input)
     - Include ALL of these terms (tag input)
     - Negative terms (tag input)
     - Negative authors (tag input)
     - Exact match toggle
     - Case sensitive toggle
   - Global filters summary (read-only, shows active global settings; "Edit in Settings →" link)
4. User clicks "Save Topic" → validation runs. On conflict, Advanced accordion auto-opens with red error.
5. On success: modal closes, topic appears in PrimaryNav topics list.
6. **Pause:** "···" → Pause → confirmation modal opens (cannot be bypassed).
7. **Resume:** "···" → Resume → immediate (no modal).
8. **Delete:** "···" → Delete → immediate deletion (no confirmation modal). Topic disappears from list.
9. **Inline custom type creation:** In type selector, user types a new type name → "+ Create '[name]' type" option appears → clicking saves the custom type and applies it to the topic.

---

**UI Copy:**

**Modal header:**
- Create: "Create topic"
- Edit: "Edit topic"

**Fields:**
- Name label: "Topic name" | Placeholder: "e.g. ContentStudio, My Brand" | Error: "Topic name is required"
- Type label: "Topic type" | Helper text: "Helps organize your topics. You can create custom types in Settings."
- Keywords label: "Keywords" | Placeholder: "Type a keyword and press Enter" | Helper text: "We'll find mentions containing any of these keywords." | Error: "Add at least one keyword to start monitoring"
- Platforms label: "Platforms" | Helper text: "Select which platforms to monitor. You can always change this later." | Error: "Select at least one platform to monitor"
- AI Context Hint label: "AI context hint (optional)" | Placeholder: "e.g. ContentStudio is a social media tool, not a photography studio" | Helper text: "Help our AI understand your brand better. Useful if your name has a common alternate meaning. Max 200 characters."
- Character counter for AI Context Hint: "X / 200"

**Advanced section:**
- Toggle label: "Advanced settings"
- Include ANY label: "Include ANY of these terms" | Helper text: "Mentions must contain at least one of these terms in addition to your keywords. Example: if your keyword is 'ContentStudio' and you add 'pricing' here, you'll only get mentions that also include 'pricing'."
- Include ALL label: "Include ALL of these terms" | Helper text: "Mentions must contain every one of these terms. Example: add 'pricing' and 'free plan' to only get mentions that discuss both."
- Negative terms label: "Exclude these terms" | Helper text: "Mentions containing any of these words will be filtered out. Example: add 'spam' or 'advertisement' to remove promotional mentions."
- Negative authors label: "Exclude these authors" | Placeholder: "@handle or u/username" | Helper text: "Mentions from these authors will never appear, even if they match your keywords."
- Exact match toggle: "Exact match only" | Tooltip: "When on, only mentions that contain your keyword as a complete word will match. For example, 'content' won't match 'ContentStudio' — only exact mentions of 'content' will."
- Case sensitive toggle: "Case sensitive" | Tooltip: "When on, 'ContentStudio' won't match 'contentstudio'. Useful if your brand has a unique capitalization that you want to track precisely."

**Global filters summary:**
- Section label: "Global filters also apply"
- Helper text: "The filters below are active for all your topics. Change them in Settings."
- "Edit in Settings →" link

**Validation errors (in Advanced accordion, shown inline):**
- Keyword + Negative Terms conflict: "'[keyword]' is in both Keywords and Exclude Terms — these would cancel each other out and match nothing. Remove it from one list to continue."
- Include term + Negative Terms: "'[term]' appears in both Include Terms and Exclude Terms — they cancel each other out. Remove it from one."
- Include ANY + Include ALL: "'[term]' appears in both Include ANY and Include ALL — Include ALL already implies Include ANY. Remove it from one."

**Pause modal:**
- Title: "Pause this topic?"
- Body: "Pausing '[topic name]' will stop looking for new mentions. Your existing mentions will stay in your feed and you can resume monitoring at any time."
- Cancel: "Cancel"
- Confirm: "Pause topic"

**Buttons:**
- Create mode: "Create Topic"
- Edit mode: "Save Changes"
- Cancel: "Cancel"

---

**Acceptance criteria:**

- [ ] "+" button in Topics section opens empty TopicFormModal in create mode
- [ ] Edit action opens TopicFormModal pre-filled with topic's current data
- [ ] Color picker shows 8 color swatches; selected color is applied to the topic dot in PrimaryNav
- [ ] Keywords tag input: adds tag on Enter or comma; tags are removable via × button
- [ ] Platforms selector shows all 18 platforms with icons; "Select all" checkbox selects/deselects all
- [ ] AI Context Hint textarea enforces 200 char max; character counter shows remaining chars
- [ ] Advanced accordion is collapsed by default; expands on click; auto-opens when there's a conflict error
- [ ] All 3 conflict types are detected simultaneously and each shows its own error message inline
- [ ] Save is blocked when any conflict error is present
- [ ] On successful save (create): topic appears in PrimaryNav topics list immediately
- [ ] On successful save (edit): topic updates in-place in the list
- [ ] Pause action opens confirmation modal with correct topic name; confirming calls pause API; topic dot goes gray + PauseCircle icon appears
- [ ] Resume action (no modal) calls resume API; topic dot returns to color + PauseCircle icon removed
- [ ] Delete action removes topic from list immediately with no confirmation
- [ ] "+ New type" option in type selector creates the custom type and applies it to the topic
- [ ] When workspace is at topic limit (5/5), create modal shows error "You've reached your topic limit (5/5). Remove an existing topic or upgrade to add more." and Save button is disabled

---

**Mock-ups:** See prototype (TopicFormModal)

**Impact on existing data:** Creates/updates topic records.

**Impact on other products:** None.

**Dependencies:** **[BE] Topic management API with keyword rules and pause/resume**, **[FE] Listening module shell — navigation, routing, and user state management**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Modal is full-screen on mobile; tag inputs and platform selector are touch-friendly
- [ ] Multilingual support — All labels in i18n; topic names and keywords support any Unicode
- [ ] UI theming support — Primary button uses `bg-primary-cs-500`; active platform pills use `bg-primary-cs-50 / border-primary-cs-200`
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — None

---

### Story 16: [FE] Views sidebar and custom view creation

**Description:**
As a Listening subscriber, I want to create custom filtered views that I can quickly switch between so that I can organize my mention stream into meaningful segments like "Crisis Monitor", "Brand Love", or "Competitor Intel".

---

**Workflow:**

1. ViewsSidebar (210px) renders to the right of PrimaryNav when Feed is the active section.
2. "FOR YOU" section (system views, read-only): "All Mentions" (no filter) and "High Relevance" (Buy Intent + Own Brand Mention).
3. "VIEWS" section: user-created views sorted by last-used order.
4. Clicking a view → activates it, updating the feed with that view's filter preset.
5. Hovering a user-created view → shows "···" context menu: Duplicate, Manage alerts, Delete.
   - Duplicate → creates a copy with "(copy)" appended to the name; auto-selects it.
   - Manage alerts → opens CreateAlertModal pre-filled with this view selected.
   - Delete → confirmation modal. If the deleted view was active, falls back to "All Mentions".
6. "[+] Add view" button → opens CreateViewModal.
7. **CreateViewModal fields:**
   - View name (required) + Icon picker (7 icons: ⛨ shield, ✕ target, ♡ heart, 🛒 cart, ✦ sparkle, ⚡ bolt, 🔔 bell)
   - Filter by Topics (multi-select, all workspace topics)
   - Filter by Platforms (multi-select, all 18)
   - Language (multi-select, 16 languages)
   - Filter by Sentiment (pill toggle: Positive / Neutral / Negative, multi-select)
   - Filter by Tags (multi-select, all 10 smart tags)
   - Minimum Author Reach (single-select: Any / 500+ / 1k+ / 5k+ / 10k+ / 50k+ / 100k+)
   - "Save View" → creates view; modal closes; new view appears in sidebar and is auto-selected.
8. Sidebar collapse: clicking collapse button (×) hides the ViewsSidebar. When hidden, Feed item in PrimaryNav shows a panel-expand icon. Clicking it re-opens the sidebar.

---

**UI Copy:**

**ViewsSidebar:**
- Section header: "FOR YOU"
- System views: "All Mentions", "High Relevance"
- Section header: "VIEWS"
- Add button tooltip: "Create a new view"
- Context menu: "Duplicate", "Manage alerts", "Delete"

**Delete confirmation modal:**
- Title: "Delete this view?"
- Body: "Deleting '[view name]' cannot be undone. Any alerts linked to this view will also be deleted."
- Cancel: "Cancel"
- Confirm: "Delete view" (destructive, red)

**CreateViewModal:**
- Title: "Create view"
- Name field label: "View name" | Placeholder: "e.g. Crisis Monitor, Brand Love" | Error: "View name is required"
- Icon picker label: "Icon"
- Filter sections: "Topics", "Platforms", "Language", "Sentiment", "Tags", "Minimum reach"
- Minimum reach options: "Any reach", "500+ followers", "1,000+", "5,000+", "10,000+", "50,000+", "100,000+"
- CTA: "Save View"
- Cancel: "Cancel"

**Sidebar toggle:**
- Collapse tooltip: "Hide views panel"
- PrimaryNav expand hint tooltip: "Show views panel"

---

**Acceptance criteria:**

- [ ] ViewsSidebar renders only when Feed section is active
- [ ] System views ("All Mentions", "High Relevance") are listed first and have no context menu
- [ ] Clicking any view updates the feed immediately with that view's filters applied
- [ ] User-created views show "···" context menu on hover
- [ ] Duplicate creates a copy with "(copy)" suffix and auto-selects the copy
- [ ] "Manage alerts" opens CreateAlertModal with the view pre-selected
- [ ] Delete shows confirmation modal; confirms removes view; if active, feed falls back to "All Mentions"
- [ ] Delete confirmation mentions that linked alerts will also be deleted
- [ ] CreateViewModal name field is required; error shown if empty on submit
- [ ] Icon picker shows 7 options; selected icon is highlighted and applied to the view in the sidebar
- [ ] All filter fields in CreateViewModal are optional; saving with no filters creates an "all mentions" view
- [ ] Saved view appears in sidebar, is auto-selected, and the feed updates to show its filtered results
- [ ] Sidebar collapses/expands via toggle; state persists in session
- [ ] On mobile, ViewsSidebar opens as a Drawer overlay (not inline)
- [ ] Loading state: skeleton list items while views load from API

---

**Mock-ups:** See prototype (ViewsSidebar + CreateViewModal)

**Impact on existing data:** Creates view records.

**Impact on other products:** None.

**Dependencies:** **[BE] Mention query and filtering API with views support**, **[FE] Mention feed with filter bar, mention cards, and action pill**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — ViewsSidebar as full-screen Drawer on mobile
- [ ] Multilingual support — View names in any Unicode; all UI labels in i18n
- [ ] UI theming support — Active view uses `bg-primary-cs-50 / text-primary-cs-500`
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — None

---

### Story 17: [FE] Bookmarks page with search and filters

**Description:**
As a Listening subscriber, I want a dedicated page where I can view and search all my bookmarked mentions so that I can reference important posts when creating reports or presentations.

---

**Workflow:**

1. User clicks "Bookmarks" in PrimaryNav → Bookmarks page renders.
2. Filter bar: text search input + Platform chip + Sentiment chip + Tags chip + Date Range chip.
3. Results: same MentionCard component as the Feed (minus the "Bookmark" action which stays filled/blue).
4. Text search matches mention content OR author name (case-insensitive).
5. All filters apply as AND logic.
6. Empty states:
   - No bookmarks: illustration + headline + subtext.
   - No search/filter matches: headline + subtext + "Clear all filters" button.

---

**UI Copy:**

**Page header:** "Bookmarks"

**Search:**
- Placeholder: "Search bookmarks..."

**Filter chips:** "Platform", "Sentiment", "Tags", "Date Range" (same options as Feed filter bar)

**Empty state (no bookmarks):**
- Headline: "No bookmarks yet"
- Subtext: "Bookmark important mentions from your feed to save them here for later. Great for client reports, team reviews, or tracking key conversations."

**Empty state (no matches):**
- Headline: "No bookmarks match your search"
- Subtext: "Try adjusting your search term or clearing your filters."
- CTA: "Clear all filters"

---

**Acceptance criteria:**

- [ ] Bookmarks page shows only mentions where `is_bookmarked = true`
- [ ] Text search filters by mention content and author name in real-time (debounced 300ms)
- [ ] Platform, Sentiment, Tags, Date Range chips filter the results correctly (AND logic)
- [ ] Mention cards on this page are the same component as the Feed (same actions, same layout)
- [ ] The Bookmark action on cards in this page unbookmarks the mention and removes it from the list immediately
- [ ] Both empty states render with correct copy
- [ ] "Clear all filters" resets all active chips and clears the search input
- [ ] Loading state: skeleton cards while data loads

---

**Mock-ups:** See prototype (Bookmarks section)

**Impact on existing data:** None.

**Impact on other products:** None.

**Dependencies:** **[BE] Mention query and filtering API with views support**, **[FE] Mention feed with filter bar, mention cards, and action pill**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Search + filter bar scrollable on mobile; cards full-width
- [ ] Multilingual support — Search works on Unicode content; all labels in i18n
- [ ] UI theming support — Active filter chips use themed styling
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — None

---

### Story 18: [FE] Analytics dashboard with charts, KPI cards, and AI insights

**Description:**
As a Listening subscriber, I want to see visual analytics about my mentions — trends, sentiment, platform distribution, and tag breakdown — so that I can understand my brand's online presence and report on it to stakeholders.

---

**Workflow:**

1. User clicks "Analytics" in PrimaryNav → Analytics page renders.
2. Filter bar: Topic filter chip (multi-select) + Date Range chip + "Export" button (opens ExportModal, Story 19).
3. KPI cards (4, horizontal row, stacked on mobile):
   - Total Mentions, Positive Sentiment %, Avg. Daily Mentions, Topics Tracked.
4. Charts (in order, stacked vertically):
   - **Mentions Over Time** (line chart): one line per topic, color-coded to topic color. X-axis: dates. Y-axis: mention count.
   - **Sentiment Trend** (stacked area chart): positive/neutral/negative stacked by day.
   - **Side-by-side row (stacked on mobile):**
     - Sentiment Distribution (donut): % positive/neutral/negative. Center: % positive in large text.
     - By Platform (donut + ranked list): platform names + counts sorted descending.
   - **By Smart Tag** (horizontal bar chart): all 10 tags, sorted by count descending.
5. Each chart has a "✦ AI Insights" button → clicking opens a popover with 3 AI-generated bullet points analyzing the chart data.
6. Date filter changes all charts + KPI cards simultaneously.
7. Topic filter changes all charts + KPI cards simultaneously.

---

**UI Copy:**

**Page header:** "Analytics"

**Filter bar:**
- Topic chip: "Topics" | popover header: "Filter by Topic"
- Date chip: same presets as Feed (All time / 24h / 7d / 30d / 90d / Custom)

**KPI cards:**
- "Total Mentions" | subtext: "In selected period"
- "Positive Sentiment" | format: "72%" | subtext: "Of all mentions"
- "Avg. Daily Mentions" | subtext: "In selected period"
- "Topics Tracked" | subtext: "Currently active"

**Charts:**
- "Mentions Over Time" | Y-axis label: "Mentions" | X-axis: dates
- "Sentiment Trend" | legend: "Positive", "Neutral", "Negative"
- "Sentiment Distribution" | center label: "Positive"
- "By Platform" | header: "Platform Breakdown"
- "By Smart Tag" | header: "Mentions by Tag"

**AI Insights button:**
- Button label: "✦ AI Insights"
- Popover loading: "Generating insights..."
- Popover header: "AI Insights"

**Empty state (no data for date range):**
- Headline: "No data for this period"
- Subtext: "Try selecting a different date range or add more topics to start collecting mentions."

---

**Acceptance criteria:**

- [ ] All 5 charts render with correct chart types (line, stacked area, donut ×2, horizontal bar)
- [ ] 4 KPI cards render with correct values from the analytics API
- [ ] Topic filter updates all charts and KPI cards simultaneously
- [ ] Date range filter updates all charts and KPI cards simultaneously
- [ ] Each chart's "✦ AI Insights" button opens a popover with 3 bullet points
- [ ] AI Insights popover shows a loading state while insights are being generated
- [ ] Line chart shows one colored line per topic (using topic's assigned color)
- [ ] Sentiment Distribution donut shows the positive % in large text in the center
- [ ] By Platform donut is accompanied by a ranked list of all platforms with counts
- [ ] By Smart Tag horizontal bar chart shows all 10 tags (with 0 for tags with no mentions)
- [ ] Charts and KPI cards are in a single-column stacked layout on mobile
- [ ] "By Platform" and "Sentiment Distribution" are side-by-side on desktop, stacked on mobile
- [ ] "Export" button opens ExportModal (see Story 19)
- [ ] Loading state: skeleton KPI cards and chart placeholder boxes while data loads
- [ ] Empty state renders when API returns no data for the selected filters

---

**Mock-ups:** See prototype (Analytics section)

**Impact on existing data:** None (read-only).

**Impact on other products:** Uses AI pipeline for AI Insights.

**Dependencies:** **[BE] Analytics aggregation API for mentions, sentiment, platform, and tag data**, **[FE] Listening module shell — navigation, routing, and user state management**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Single column on mobile; charts are scrollable/zoomable on touch
- [ ] Multilingual support — KPI values are numeric; chart labels in i18n
- [ ] UI theming support — Topic line colors use assigned topic colors (not primary theme color); AI Insights button uses `text-primary-cs-500`
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — AI Insights uses AI agents pipeline (web-only)

---

### Story 19: [FE] Export modal — download, scheduled reports, and share link

**Description:**
As a Listening subscriber, I want to export my analytics data as a PDF or CSV, schedule recurring report emails, or create a shareable read-only link so that I can share Listening insights with clients and leadership.

---

**Workflow:**

1. User clicks "Export" button in Analytics filter bar → ExportModal opens (3 tabs).
2. **Tab 1: Download**
   - Format selector: "PDF Report" | "CSV Data" (segmented control, PDF default)
   - Date range chip (same presets as Analytics filter bar)
   - "Generate & Download" button → triggers API → file downloads
   - Success: green alert banner "Report ready! Your download will start automatically."
3. **Tab 2: Schedule**
   - Frequency: "Weekly" | "Monthly" (segmented control)
   - Weekly: "Send on" day-of-week dropdown
   - Monthly: "Send on day" picker (1st–28th)
   - Time: hour selector (06:00–18:00 in 1-hour increments)
   - Recipients: team member multi-select + manual email chip input (Enter/comma to add)
   - "Save Schedule" → creates scheduled report; button disabled until ≥1 recipient added
   - Success: green banner "Schedule saved! Reports will be sent automatically."
4. **Tab 3: Share**
   - Info text explaining read-only link.
   - "Generate Share Link" button → generates URL + shows copy button.
   - Copy → clipboard → toast "Link copied to clipboard!"
   - Footer: "Share links expire after 30 days."

---

**UI Copy:**

**Modal title:** "Export Analytics"

**Tab 1 (Download):**
- Format label: "Format"
- Format options: "PDF Report", "CSV Data"
- Date range label: "Date range"
- CTA: "Generate & Download"
- Success: "Report ready! Your download will start automatically."

**Tab 2 (Schedule):**
- Frequency label: "Frequency"
- Frequency options: "Weekly", "Monthly"
- Weekly field label: "Send on" | placeholder: "Select day"
- Monthly field label: "Send on day" | placeholder: "Select day of month"
- Time label: "Send at" | placeholder: "Select time"
- Recipients label: "Recipients"
- Recipients helper: "Add team members or enter any email address. Press Enter or comma to add."
- Recipients placeholder: "Enter email address..."
- CTA: "Save Schedule" (disabled until ≥1 recipient)
- Success: "Schedule saved! Reports will be sent automatically."

**Tab 3 (Share):**
- Info text: "Generate a read-only link you can share with anyone — like a client or your leadership team. They can view your analytics but cannot make any changes."
- CTA: "Generate Share Link"
- After generation: "Copy link" button
- Toast: "Link copied to clipboard!"
- Footer: "Share links expire after 30 days."

---

**Acceptance criteria:**

- [ ] Modal opens with 3 tabs: Download, Schedule, Share
- [ ] Tab 1: PDF and CSV format options; date range chip is functional; "Generate & Download" triggers file download
- [ ] Tab 1: success green banner shown after successful download trigger
- [ ] Tab 2: frequency selector switches between weekly and monthly sub-fields correctly
- [ ] Tab 2: "Save Schedule" is disabled until ≥1 recipient is added
- [ ] Tab 2: email chip input accepts emails on Enter and comma; invalid emails show inline error "Enter a valid email address"
- [ ] Tab 2: success banner shows after schedule is saved
- [ ] Tab 3: "Generate Share Link" calls API and renders the resulting URL in a readonly input with a Copy button
- [ ] Tab 3: copy button copies URL to clipboard; toast confirms
- [ ] Tab 3: "Share links expire after 30 days." footer is visible

---

**Mock-ups:** See prototype (ExportModal in Analytics section)

**Impact on existing data:** Creates export schedule and share link records.

**Impact on other products:** None.

**Dependencies:** **[BE] Export service — PDF/CSV, scheduled email reports, and shareable analytics links**, **[FE] Analytics dashboard with charts, KPI cards, and AI insights**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Modal is full-screen on mobile; tab switcher scrollable
- [ ] Multilingual support — All labels in i18n; email inputs accept any valid email
- [ ] UI theming support — Primary CTAs use `bg-primary-cs-500`
- [ ] White-label domains impact review — Share link uses workspace custom domain on white-label
- [ ] Cross-product impact assessment — Schedule email uses existing email infrastructure

---

### Story 20: [FE] Alerts management UI — create, edit, and manage alert rules

**Description:**
As a Listening subscriber, I want to set up alerts that notify me by email when there's an unusual spike in mentions or a rise in negative sentiment so that I'm alerted to potential PR issues before they escalate.

---

**Workflow:**

1. User clicks "Alerts" in PrimaryNav → Alerts page renders.
2. Page shows: active alerts section + paused alerts section + "New Alert" button.
3. Each alert card shows: view name, trigger pills (Volume Spike / Sentiment Spike with threshold), recipient avatars/email chips, active/paused toggle, "···" menu (Edit / Delete).
4. Toggle → switches alert between active/paused instantly.
5. Delete → confirmation modal → removes alert.
6. "New Alert" / Edit → CreateAlertModal:
   - View selector: dropdown (all views, required)
   - Volume Spike card: toggle ON/OFF + configurable threshold % (slider or number input, 10–500%)
   - Sentiment Spike card: toggle ON/OFF + configurable threshold % (10–100%)
   - At least one trigger must be enabled.
   - Recipients: team member multi-select + email chip input.
   - "Save Alert" → disabled until ≥1 trigger ON and ≥1 recipient added.
7. Empty state (no alerts): illustration + headline + CTA.

---

**UI Copy:**

**Page header:** "Alerts"
**Page subheader:** "Get notified when something unusual happens across your monitoring."

**Alert card:**
- Volume Spike pill: "Volume Spike · [X]% above average" (amber)
- Sentiment Spike pill: "Sentiment Spike · [X]% negative" (red)
- Toggle: active → "Active" | paused → "Paused"
- Context menu: "Edit", "Delete"

**CreateAlertModal:**
- Title (create): "Create alert"
- Title (edit): "Edit alert"
- View selector label: "Monitor this view" | placeholder: "Select a view" | Error: "Select a view to monitor"
- Volume Spike card title: "Volume Spike"
- Volume Spike card body: "Notify me when mentions suddenly increase."
- Volume Spike toggle label: "Enabled"
- Volume Spike threshold label: "Notify when mentions are [X]% above the 7-day average"
- Sentiment Spike card title: "Sentiment Spike"
- Sentiment Spike card body: "Notify me when negative mentions increase."
- Sentiment Spike toggle label: "Enabled"
- Sentiment Spike threshold label: "Notify when negative mentions exceed [X]% of total"
- Recipients label: "Send alerts to"
- Recipients helper: "Add team members or enter any email. Press Enter or comma to add."
- At-least-one-trigger error: "Enable at least one alert type before saving."
- CTA: "Save Alert"
- Cancel: "Cancel"

**Delete confirmation:**
- Title: "Delete this alert?"
- Body: "This alert will be permanently removed and notifications will stop."
- Cancel: "Cancel"
- Confirm: "Delete alert" (destructive, red)

**Empty state:**
- Headline: "No alerts set up yet"
- Subtext: "Alerts let you know when something unusual happens in your feed — like a sudden spike in mentions or a rise in negative sentiment. Set one up so you never miss a critical moment."
- CTA: "Create your first alert"

---

**Acceptance criteria:**

- [ ] Alerts page shows active and paused sections; each section only visible when it has items
- [ ] Each alert card shows the view name, trigger pills with threshold values, and recipient info
- [ ] Toggle on alert card switches status instantly; confirmed with API call
- [ ] "···" menu shows Edit and Delete options
- [ ] Delete shows confirmation modal; confirms removes the alert
- [ ] CreateAlertModal view selector is required; shows error if save attempted without selection
- [ ] Volume Spike and Sentiment Spike each have independent toggles and threshold inputs
- [ ] Threshold inputs accept integers (Volume: 10–500, Sentiment: 10–100); out-of-range values are rejected inline
- [ ] "Save Alert" is disabled if no triggers are enabled OR no recipients added
- [ ] Recipients chip input validates email format; invalid emails show "Enter a valid email address"
- [ ] Successful save closes modal; new alert appears in the active section
- [ ] Empty state renders correctly when no alerts exist
- [ ] "Create your first alert" CTA opens CreateAlertModal

---

**Mock-ups:** See prototype (Alerts section)

**Impact on existing data:** Creates alert rule records.

**Impact on other products:** Alert emails use existing email infrastructure.

**Dependencies:** **[BE] Listening alerts engine — volume spike and sentiment spike detection**, **[FE] Views sidebar and custom view creation**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Alert cards stack; CreateAlertModal full-screen on mobile
- [ ] Multilingual support — All labels in i18n; email inputs accept any valid email
- [ ] UI theming support — Volume Spike pill uses amber; Sentiment Spike uses red (intentional, not themed)
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — None

---

### Story 21: [FE] Settings page — global filters and topic type management

**Description:**
As a Listening subscriber, I want to configure workspace-wide filters that apply across all my topics, and manage my custom topic types, so that I can reduce noise globally without touching each topic individually.

---

**Workflow:**

1. User clicks "Settings" in PrimaryNav → Settings page renders.
2. Two tabs: "Global Filters" | "Topic Types"
3. **Global Filters tab:**
   - Negative Terms: tag input, type + Enter to add. All matching mentions are excluded from all topics.
   - Blocked Authors: tag input (@handle or u/username). Mentions from these authors never appear.
   - Excluded Subreddits: tag input (r/subredditname). Useful for ignoring off-topic subreddits.
   - Language Filter: multi-select dropdown (16 languages; empty = all languages monitored).
   - "Save Global Filters" button (sticky or at bottom; disabled if no changes).
   - Info block (read-only, always visible): "These settings apply to all topics. For topic-level filters, edit the individual topic."
4. **Topic Types tab:**
   - Built-in types (read-only): Own Brand, Competitor, Industry Term — with "built-in" badge and no edit/delete.
   - Custom types: listed below with Edit (inline rename) and Delete (×) per type.
   - "Add Type" button → inline form: "Type name..." input → "Create" / "Cancel".
   - Delete: if the type is in use by any topics, show error. Otherwise remove immediately.

---

**UI Copy:**

**Page header:** "Settings"

**Tab 1: Global Filters**
- Negative Terms label: "Excluded keywords" | Helper text: "Mentions containing any of these words will be hidden from your entire feed, across all topics. Example: add 'spam', 'bot', or 'advertisement' to clean up your feed." | Placeholder: "Type a word and press Enter"
- Blocked Authors label: "Blocked authors" | Helper text: "Mentions from these users will never appear in your feed, regardless of topic. Use @handle for Twitter/Instagram, u/username for Reddit." | Placeholder: "@handle or u/username"
- Excluded Subreddits label: "Excluded subreddits" | Helper text: "Mentions from these subreddits will be hidden. Useful for removing off-topic communities. Example: r/memes" | Placeholder: "r/subredditname"
- Language Filter label: "Monitor only these languages" | Helper text: "Leave empty to monitor mentions in all languages. Select one or more languages to only collect mentions written in those languages." | Placeholder: "All languages"
- Save button: "Save Global Filters"
- Info block: "These filters apply across all topics. To set filters for a specific topic only, edit that topic directly."

**Tab 2: Topic Types**
- Section heading: "Built-in types" — "Built-in types can't be edited or removed."
- Built-in type badge: "built-in"
- Section heading: "Custom types"
- Add button: "Add Type"
- Inline create form: "Type name..." input | "Create" | "Cancel"
- Inline edit: TextInput pre-filled with type name | "Save" | "Cancel"
- Delete in-use error: "This type is used by [N] topic(s). Remove it from those topics before deleting."
- Empty custom types: "No custom types yet. Create one to organize your topics your way."

---

**Acceptance criteria:**

- [ ] Global Filters tab shows 4 field groups: negative terms, blocked authors, excluded subreddits, language filter
- [ ] All four fields use tag inputs (except language which is a multi-select); tags are removable via ×
- [ ] "Save Global Filters" is disabled when no changes have been made since the last save
- [ ] Saving shows a success toast: "Global filters saved"
- [ ] Changes take effect for new mention ingestion (not retroactively applied to existing mentions)
- [ ] Info block is visible at all times as a read-only summary
- [ ] Topic Types tab shows 3 built-in types with "built-in" badges and no edit/delete controls
- [ ] Custom types listed below built-in types with inline edit and delete (×) controls
- [ ] "Add Type" shows inline form; "Create" button disabled until name is entered; Enter key submits
- [ ] Duplicate type name shows inline error: "A type with this name already exists"
- [ ] Delete of a type used by topics shows error with count; delete of unused type removes immediately
- [ ] Inline edit of a custom type: Enter saves, Escape cancels
- [ ] Language filter placeholder shows "All languages" when nothing selected

---

**Mock-ups:** See prototype (Settings section)

**Impact on existing data:** Updates workspace global settings. No mention data retroactively changed.

**Impact on other products:** None.

**Dependencies:** **[BE] Global settings and topic types API**, **[FE] Listening module shell — navigation, routing, and user state management**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Settings tabs stack; tag inputs touch-friendly on mobile
- [ ] Multilingual support — All labels in i18n; field values (blocked terms, authors) accept any Unicode
- [ ] UI theming support — Primary save button uses `bg-primary-cs-500`
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — None

---

### Story 22: [FE] Usage bar, limit warnings, and add-on upgrade prompts

**Description:**
As a Listening subscriber, I want to see how many topics and mentions I've used this month so that I know when I'm approaching my limits and can take action before monitoring stops.

---

**Workflow:**

1. When `userState = active`, the bottom of PrimaryNav shows the Usage bar section.
2. Two progress bars:
   - **Topics:** X / 5 — bar fills `bg-primary-cs-500`; turns red and bar fills red when at limit (5/5).
   - **Mentions this month:** X.Xk / 10k — bar fills `bg-primary-cs-500`; turns amber at ≥90% (9,000+); turns red at 100% (10,000).
3. Below both bars: "Need more? Add-ons available →" link → opens billing/add-ons page.
4. **At 90% mention usage:** inline amber warning appears below the mentions bar: "Approaching limit — you're at [X]% of your monthly mentions."
5. **At 100% mention usage:** red warning + the link changes to "Upgrade now to resume monitoring →".
6. **At topic limit (5/5):** topics bar turns red; clicking "+" to add a new topic shows an error in the modal: "You've reached your topic limit (5/5). Remove a topic or add more with a plan upgrade."

---

**UI Copy:**

**Usage bar labels:**
- Topics row: "Topics" | count: "[X] / 5"
- Mentions row: "Mentions this month" | count: "[X.X]k / 10k"

**At 90% mentions:**
- Warning text (amber): "Approaching your monthly limit — add more mentions to avoid interruptions."

**At 100% mentions:**
- Warning text (red): "Monthly limit reached — new mentions are paused until your limit resets or you upgrade."
- CTA link: "Upgrade now to resume monitoring →"

**Topic limit reached (in TopicFormModal):**
- Error message: "You've reached your topic limit (5/5). Remove an existing topic or add more topics with a plan upgrade."
- CTA in error: "Upgrade plan →"

**Normal state CTA:**
- "Need more? Add-ons available →"

---

**Acceptance criteria:**

- [ ] Usage bar section is visible only when `userState = active`; hidden for trial/locked/expired
- [ ] Topics bar shows correct count vs limit; bar width is proportional (e.g., 3/5 = 60% fill)
- [ ] Topics bar and count text turn red when topics count equals the limit
- [ ] Mentions bar shows correct count vs limit; count formatted as "X.Xk / 10k"
- [ ] Mentions bar turns amber at ≥90% of limit; count text turns amber
- [ ] Mentions bar turns red at 100% of limit; count text turns red
- [ ] Amber warning text appears below mentions bar at ≥90%
- [ ] Red warning text and "Upgrade now" link appear at 100%
- [ ] "Need more? Add-ons available →" is always visible (except replaced by "Upgrade now" at 100%)
- [ ] All CTA links navigate to the correct billing/add-on page
- [ ] Usage data is fetched from `GET /api/listening/usage` and refreshed on navigation
- [ ] Topic limit error in TopicFormModal is shown when workspace is already at limit

---

**Mock-ups:** See prototype (PrimaryNav bottom section)

**Impact on existing data:** None (read-only display).

**Impact on other products:** Links to billing page (add-on purchase flow).

**Dependencies:** **[BE] Listening add-on subscription management and usage limits**, **[FE] Listening module shell — navigation, routing, and user state management**.

**Global quality & compliance:**
- [ ] Mobile responsiveness — Usage bar is visible at the bottom of the mobile drawer nav
- [ ] Multilingual support — Usage counts are numeric; all labels in i18n
- [ ] UI theming support — Normal fill uses `bg-primary-cs-500`; amber and red are intentional semantic colors, not themed
- [ ] White-label domains impact review — Add-on pricing ($49/mo) is CS-specific; white-label pricing must be configurable
- [ ] Cross-product impact assessment — Links to billing; ensure billing team is aware of add-on SKU
