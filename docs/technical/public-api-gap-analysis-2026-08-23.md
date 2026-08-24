# Public API gap analysis

Date: 2026-08-23
Question: what can the ContentStudio product do that the public API cannot?

**Method.** Enumerated the full public API v1 surface from `contentstudio-backend/routes/api/v1.php` (222 route declarations), the published OpenAPI document (`storage/api-docs/api-docs.json`, 189 documented paths), the inbox service's own `/v1` surface (`social-inbox-manager/app/api/v1/routes/`), and the analytics service's API surface. Compared against the product surface: `routes/web/*.php`, `routes/api.php`, `routes/api/approval-workflows.php`, and the inbox service's legacy routes. Then cross-checked every gap against the API work already planned or shipped in `docs/features/` and `docs/stories/`, so this doc separates **unplanned gaps** from **gaps that already have stories**.

A caveat on the numbers: route counts are not capability counts. A single product route like `processPlanBulkOperation` covers many operations, and one API route can serve several. Treat counts as a sense of proportion, not a score.

---

## 1. Summary

The public API is in better shape than the framing "what's missing" implies. Publishing is close to product parity, workspace and team management shipped recently, the inbox proxy is near-complete against the inbox service's own `/v1`, and analytics read coverage is genuinely strong — Bluesky analytics exists in the API and the twelve analytics surfaces include Meta Ads and Google Ads.

The gaps cluster in four places, and they are not evenly serious:

| Area | State | Headline gap |
|---|---|---|
| **Publishing / composer** | Strong | No single-post read, no account or platform filter on listing, no bulk operations, no repeat posts, no automations at all |
| **Workspace management** | Strong for the workspace itself, thin around it | Content categories and approval workflows are read-only, media is upload-and-list only |
| **Inbox** | Near-complete on conversations | No auto-reply rules, no saved replies, no brand help docs |
| **Analytics** | Strong on reads, absent on outputs | No reports, no scheduled reports, no competitor analytics, no social listening |
| **Whole modules** | Absent | Automations, social listening, discovery, billing and usage, white-label, hashtags and UTMs |
| **Cross-cutting** | Two real problems | Webhooks cannot be managed via the API, and the OpenAPI document has drifted from the code |

The single most consequential finding is not an endpoint. It is that **an API-first customer cannot complete an API-first workflow without logging into the web app** — to manage their webhooks, to create a content category, to set up an approval workflow, or to see what they have consumed. That undercuts the API-centric plan more than any individual missing route.

---

## 2. Publishing and composer

### What the API already has

Genuinely broad. `POST /workspaces/{id}/posts` accepts:

- Publish types: `now`, `scheduled`, `draft`, `queued`, `content_category`
- Per-platform overrides for all twelve platforms, with per-platform text, media and `post_type`
- X threads (`twitter_options.threaded_tweets`), Meta Threads threading (`threads_options.multi_threads`)
- Facebook carousel with cards, CTA and end card; Facebook collaborators; Facebook text backgrounds
- Instagram collaborators and trial reels; LinkedIn polls and title; GMB, YouTube, TikTok and Pinterest option blocks
- First comment with its own account list
- Approval by approver list, plus approval workflows by `workflow_id` with a `workflow_action`
- Labels, `campaign_id`, `content_category_id`, `hide_client`
- Media by URL or by `media_ids`

Plus `PUT` to edit, `DELETE`, `POST .../approval` for approve and reject, post comments read and write, `GET /platforms`, OAuth connect and reconnect, Bluesky by credentials, Facebook Group manual add, account disconnect, and best-time-to-post via `POST .../scheduling/optimal-times`.

### Gaps: unplanned

1. **No single-post read.** There is `GET /posts` (list) but no `GET /posts/{post_id}`. Every integration that creates a post and later wants its current state has to page the list endpoint and filter client-side. This is the most obvious hole in the publishing surface and the cheapest to close.
2. **No account or platform filter on `GET /posts`.** `PostIndexRequest` supports `status`, `date_from`, `date_to`, `approval_assigned_to`, `approval_requested_by`, `labels`, `campaigns`, `content_category`, `created_by`, `comment_status`, `page`, `per_page`. There is no way to ask "posts for this Instagram account" or "posts on LinkedIn", which the planner does routinely.
3. **No bulk operations.** The product has `processPlanBulkOperation`, `processPlannerBulkEdit`, `sendPlansForBulkApprovers` and `plans/swapExecutionTime`. The API is one post per call.
4. **No repeat posts.** `repeat_post`, `repeat_parent` and `publish_time_options` are on the `Plans` model and drive a real product feature. `PostStoreRequest` accepts none of them.
5. **No queue management.** `shuffleQueuePosts` and slot manipulation have no API equivalent, even though `queued` is an accepted publish type.
6. **No version history.** `fetchVersionHistory` is product-only.
7. **No tasks.** Post comments are exposed; the parallel tasks feature (`saveTask`, `getTasks`, `editTaskStatus`) is not. Comment reactions and resolve/unresolve are also product-only, though basic comment create and list are exposed.
8. **No post share links.** `ShareLinkController` under the planner covers create, update, delete and list for management, plus the client-facing surface where a reviewer verifies a password, views the scheduled posts, comments, and approves or rejects. None of it is on the public API. **Correction to an earlier version of this doc,** which pointed at the `share-link` and `shared` route groups in `routes/api.php`: those are `AnalyticsShareLinkController`, a separate feature for sharing analytics reports. The two share-link systems have separate controllers and separate storage and must not be conflated.
9. **No planner saved views or filters.** `planner/saved-views`, `createPlannerFilter`, `setDefaultPlannerFilters` are product-only.
10. **No calendar notes.** Product-only.
11. **`instagram_posting_option`** is not exposed, so an integration cannot choose between direct publishing and the mobile push-notification route for accounts where both apply.

### Gaps: already planned or shipped

The publishing API has been worked hard and most of what an audit would have flagged a year ago is done: `api-account-connection` (v1.13), `publishing-api-v15-v16` (team members, labels and campaigns), `publishing-api-v17` (workspace management), `publishing-api-v19-v22` (post filters and statuses, X threads, Threads threading, Facebook carousel), `api-delete-posts-from-native-platforms`, `api-post-editing-approvals-account-disconnect`, `api-platform-overrides-parity`, `publishing-api-labels-campaigns`. Check each of these for shipped-versus-open before treating anything above as new.

---

## 3. Workspace management

### What the API already has

`GET /me`, `GET /workspaces`, `POST /workspaces`, `PUT /workspaces/{id}`, `DELETE /workspaces/{id}`, full team member CRUD with per-action permission gating (`save_member`, `update_member`, `remove_member`), `GET .../accounts`, account disconnect, and full CRUD on labels and campaigns.

That is a real workspace management surface, and it is recent.

### Gaps: unplanned

1. **Content categories are read-only.** `GET .../content-categories` exists. The product has create, update, delete, `next_slot`, `shuffle_posts`, slot management, `delete_all`, plus per-member category access (`member/{member_id}/category-access`, `categories-for-member`, `category/{category_id}/members`). An API caller can post *into* a category but cannot create one, which makes category-based scheduling only half-usable programmatically.
2. **Approval workflows are read-only.** `GET .../approval-workflows` exists. The product has 18 routes: create, update, delete, duplicate, set and remove default, approve, reject, revoke approval, re-notify a member, save-with-approval-action, approval history, cascade jobs, and approval notifications with mark-read. The API has post-level approve and reject but no workflow definition management.
3. **Media library is upload-and-list only.** `MediaController` has exactly two methods, `index` and `upload`. The product storage surface has 47 routes: folders (`fetch/folders`, `create`, `update`, `delete`), `remove`, `archive`, `move`, `updateNote`, `flag-brand-asset`, `uploadByLink`, `uploadByBytes`, `uploadDocument`, `fetch/stats`, `limits`, `export-media-csv`. An integration can put media in and never organise, tag, move or delete it.
4. **No team member social-account access.** `team/social-account-access` controls which accounts a member may use. Not exposed, so an API caller can invite a member but not scope them to accounts.
5. **No hashtag groups.** `hashtag/save`, `fetch`, `remove` are product-only, despite hashtags being a publishing concern.
6. **No UTM presets.** `utm/save`, `fetch`, `remove`, `editStatus`, `changeAttachUtmStatus` are product-only. For an automation-driven customer, campaign tracking presets are exactly the sort of thing they would want to set programmatically.
7. **No user or workspace preferences.** Timezone update (`updateWorkspaceTimezone`), pause and resume posting (`pausePosting`, `resumePosting`), workspace logo upload, default workspace, notification preferences. Pause posting in particular is an operational control an integration might legitimately need.
8. **No white-label, custom domains or SSO.** `routes/web/whitelabel.php`, `workspace/{id}/domains`, `workspace/{id}/saml/idps` are product-only. Defensible as an admin-console concern, but worth a deliberate decision rather than an omission.

---

## 4. Inbox

### What the API already has

The most complete of the four areas relative to its own service. The v1 proxy covers `elements/search`, `elements/summary`, bulk triage via `PATCH elements` (which consolidates the four legacy state endpoints for done, pending, archive and assignee), mark read, tag add and remove, contact read and update, full tag CRUD plus merge and delete, conversation messages and notes, bookmarks, post comments read and reply, comment hide, unhide, like, unlike and delete, GMB review reply and reply delete, message bookmark and delete, and message send with multipart media support.

Cross-checked against `social-inbox-manager/app/api/v1/routes/`, the proxy exposes essentially the whole `/v1` surface. The only `/v1` routes not proxied are the `sync` family (`conversations/{id}/sync`, `messages/{id}/sync`, `comments/{id}/sync`, `posts/{id}/sync`) and `sync-jobs`, which are plausibly internal.

### Gaps: unplanned

1. **No auto-reply rules.** `routes/web/inbox.php` has full rule CRUD plus toggle, `validate-coherence`, `suggest-keywords`, `auto-reply/limits`, `auto-reply/quota`, and `auto-reply/documents`. None of it is on the public API. This is the largest inbox gap: AI auto-replies are a metered, differentiating feature and are entirely unmanageable programmatically.
2. **No saved replies.** `App\Repository\Inbox\InboxSavedReplyRepository` backs a product feature with no API surface.
3. **No brand help docs.** `BrandHelpDocController` and the document indexing that feeds auto-reply context are product-only.
4. **`sidebar_details`** has no proxied equivalent, so the per-conversation context panel data is not retrievable in one call.

**Correction to an earlier version of this doc.** It previously listed "no inbox views or filters" as a gap, on the claim that the product persists inbox filter configurations. It does not. The inbox filter drawer keeps a value locally as a browser convenience; there is no persisted, server-side saved view, no collection and no endpoint, and nothing in the inbox service either. Planner saved views exist and are a separate feature. So this was never an API gap.

**Scope decision (2026-08-24): the inbox is out of scope.** No inbox work is planned from this analysis, neither the API items above nor any product capability behind them. The findings stay here as a record. Note that the inbox is already the best-covered area of the public API, so nothing above blocks an integration from doing useful work today.

---

## 5. Analytics

### What the API already has

Twelve surfaces, all proxying the analytics pipeline: Facebook, Instagram, YouTube, LinkedIn, TikTok, Pinterest, Twitter/X, Bluesky, GMB, Meta Ads, Google Ads, and campaigns-and-labels. Depth is real: YouTube alone exposes 20 endpoints including daily-granularity trends, watch time, performance schedule and demographics. Google Ads exposes keywords, search terms, Shopping, four conversion endpoints and a conversion funnel. Most surfaces include `ai-insights`. Bluesky exposes a `capabilities` endpoint that reports each metric the AT Protocol cannot provide, with the reason, instead of returning zeroes.

On several points the API is *ahead* of the product web routes: Bluesky and the ads surfaces are in the API, and `routes/web/analytics.php` has no Bluesky prefix.

### Gaps: unplanned

1. **No reports.** The whole `AnalyticsReports` controller is product-only: `reports/save`, `show`, `remove`, `retry`, `list` (download). An integration cannot generate or fetch a PDF report.
2. **No scheduled reports.** `ScheduleReports` (`reports/schedule/save`, `show`, `remove`, `send`) is product-only. Scheduled white-labelled reporting is a common agency ask and a natural API feature.
3. **No cross-workspace reports.** `WorkspaceReportsController` (`workspace-reports/generate`, `show`, `list`, `remove`, `options`) produces the org-level PDF spanning workspaces. Product-only.
4. **No competitor analytics.** `FacebookCompetitorController` and `InstagramCompetitorController` cover competitor report CRUD, comparison and search. Product-only, and this is a paid, differentiating surface.
5. **No shareable report links.** The `X-Shareable-ID` mechanism the analytics service honours has no public API counterpart for creating or managing those links.
6. **No analytics job triggers.** `analytics/triggerJob` and `analytics/triggerCompetitorJob` are product-only, so an integration cannot force a refresh before reading.
7. **No account preferences.** `setAccountPreferences` is product-only.
8. **Platform coverage asymmetry.** Threads, Tumblr and Telegram can be published to through the API but have no analytics surface anywhere, in the product or the API. That is a product gap rather than an API gap, and `analytics-network-research-bluesky-threads` already covers the research.

### Gaps: already planned

`public-analytics-api-rollout`, `public-api-ads-analytics` and `public-api-best-time-to-post` cover the read surfaces, which is why they look complete. Reports and competitor analytics are not covered by any existing deliverable found.

---

## 6. Whole modules with no public API at all

| Module | Product surface | Notes |
|---|---|---|
| **Automations** | `routes/web/automation.php`, 50 routes: evergreen automations with variations and recycling, RSS automations, CSV bulk automations, bulk caption generation, per-post CSV actions | Nothing. The single biggest absent module, and the one closest in spirit to what an automation-first customer buys the API for |
| **Social listening** | `routes/api.php` `listening` prefix, gated on `listening_access` | Nothing. Mentions, topics, credit usage all product-only |
| **Discovery / content curation** | `routes/web/discovery.php`, plus Feedly and OPML import | Nothing. Arguably fine, since discovery is an exploratory UI feature |
| **Billing, subscription and usage** | `routes/billing.php`, 72 routes including the X wallet usage list and CSV export, plan limits, addons | Nothing. See the cross-cutting section: no way to read your own limits or consumption |
| **AI text and captions** | `customizeCaptions`, `post/generatePosts`, `chat`, `replyWithAi`, `transcribeAudio`, brand voice and brand knowledge, reels workflow | AI **images** and **videos** are on the API with a proper async job contract. AI **text** is not, which is the odd asymmetry: the cheapest AI capability is the one an integration cannot call |
| **White-label, domains, SSO** | `routes/web/whitelabel.php`, domains, SAML IdPs | Nothing. Probably correct, but should be a stated decision |

---

## 7. Cross-cutting problems

These matter more than most individual endpoints.

### 7.1 Webhooks cannot be managed through the API

The public webhooks feature shipped: `WebhookEndpointController`, `WebhookDeliveryLogController` and `WebhookEventTypeController` exist, with create, update, delete, rotate secret, test, delivery logs and an event-type catalog. But they are registered in `routes/api.php` inside `Route::middleware(['auth', 'set.locale'])`, which is **session authentication**. They are not on the `api.key` stack in `routes/api/v1.php`.

So the flagship API feature for event-driven integrations can only be configured by a human logging into the web app. An integration cannot provision its own webhook, rotate its own secret, or read its own delivery failures. For a customer on the API-centric plan this is a workflow break, not a missing nice-to-have.

The same applies to API key management and request logs, which is more defensible: creating the first key has to happen somewhere trusted. Webhooks have no such excuse.

### 7.2 The OpenAPI document has drifted from the code

`storage/api-docs/api-docs.json` documents 189 paths. `routes/api/v1.php` declares 222 routes. Not directly comparable, but one concrete drift is confirmed: **the ten Bluesky analytics endpoints are entirely undocumented.** The only Bluesky path in the document is `add/bluesky`. A capability we built, including the thoughtful `capabilities` endpoint, is invisible to every consumer reading the docs.

Worth an explicit audit of code-versus-document, and worth a check in CI so it cannot drift again.

### 7.3 No way for a caller to read its own limits or consumption

There is no endpoint answering "how many posts can I still publish", "how many AI credits do I have left", "what is my rate limit", or "what have I consumed this period". The product computes all of it (`PlanHelper`, `SubscriptionLimits`, the X wallet usage endpoint at `billing/.../usage` with CSV export). An integration discovers its limits by hitting them and reading an error.

This connects directly to the usage visibility epic. Whatever per-event usage record that produces should have a public read surface as part of the same work, not as an afterthought.

### 7.4 Read-only asymmetries are the recurring pattern

Content categories, approval workflows, media, inbox tags versus inbox rules — the recurring shape is that the API can *list* a thing and *reference* it when creating a post, but not *create* the thing itself. That forces every non-trivial integration into a hybrid where setup happens in the browser and only the repetitive part is automated. It is worth deciding whether that is the intended product boundary or just where the work stopped.

---

## 8. Competitive framing

`docs/features/api-centric-plan/01-research.md` already benchmarks Late/Zernio and Ayrshare. Against that benchmark, ContentStudio's API is **stronger on analytics depth and on approval workflows** and **weaker on developer experience**: Late ships SDKs in eight languages, and this analysis found no ContentStudio SDKs beyond the CLI, the MCP servers and the no-code connectors.

Two of the gaps above are things Late explicitly markets: unified media management, and programmatic account connection (which ContentStudio now has). Automations and listening are ContentStudio product strengths that the API does not expose at all, so they are differentiation currently left on the table.

---

## 9. Suggested priority

Ordered by value against cost, not by size.

**Tier 1, small and high value**

1. `GET /posts/{post_id}` — single-post read. Trivial, and closes the most-hit hole.
2. Account and platform filters on `GET /posts`.
3. Move webhook management onto the `api.key` stack so integrations are self-service.
4. Audit and regenerate the OpenAPI document, starting with Bluesky analytics, and add a CI check.

**Tier 2, unblocks API-first workflows**

5. Content category CRUD, so category-based scheduling is fully programmatic.
6. Media library management: folders, delete, move, update.
7. A limits-and-usage read endpoint, built alongside the usage visibility epic.
8. Approval workflow definition CRUD.

**Tier 3, exposes differentiating product**

9. Inbox auto-reply rules, saved replies and brand help docs.
10. Analytics reports: generate, list, download, plus scheduled reports.
11. AI text and caption generation, closing the asymmetry with images and videos.
12. Bulk post operations and repeat posts.

**Tier 4, deliberate decisions rather than obvious gaps**

13. Automations: evergreen, RSS, CSV.
14. Competitor analytics and social listening.
15. White-label, domains and SSO — decide and document, even if the answer is no.

---

## 10. What to verify before acting on this

1. **Shipped-versus-planned per item.** Several `docs/features/publishing-api-*` and `docs/stories/api-*` deliverables were written before the current code state. Confirm what is already live before writing a story for it.
2. **Webhook middleware placement** may be deliberate for a security reason not visible in the routes. Confirm with whoever built it before moving it.
3. **The `sync` inbox endpoints** are treated here as internal. Confirm.
4. **The white-label and SSO omission** may be a policy decision already taken. Confirm before proposing work.
5. **Route counts** in this doc are declaration counts, not capability counts. Do not use them in a business case without normalising them.
