# ContentStudio AI Chat — Use Case Catalog (v1.1)

**Purpose:** every job a social media manager will bring to the chat, and the exact tool sequence that serves it. This is the source for stories and for backend tool-description writing.

> **v1.1:** corrected against the codebase (2026-07-28). Domain D (Inbox) is **not blocked** — the API is live. UC-A11 (reviews) is now end-to-end, not read-only. See [As-Built-Baseline doc](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d).

**Format per use case:** what the user says → which tools run → what can go wrong.

**Priority:** P0 = launch blocker · P1 = launch nice-to-have · P2 = fast follow

---

## Domain A — Publishing & Scheduling

### UC-P01 · Create and schedule a single post — **P0**, Phase 2
> "Post this to Instagram and LinkedIn tomorrow at 10am: [text]"
> "Schedule a post about our Friday sale on all accounts"

`context_get` → resolve accounts by name → `posts_validate` → **preview card** → `posts_create`

**Edge cases**
- Account named ambiguously ("our Facebook" when three FB pages exist) → account picker component, do not guess
- No time given → default to next available slot, state the assumption out loud, offer to change
- Time already in the past → reject before calling, ask
- Timezone → always confirm in workspace timezone, never UTC
- Text over a platform limit → offer to shorten for that platform only, keep the rest
- Account token expired → error envelope suggests `accounts_connect` with `process=reconnect`

### UC-P02 · One idea, customised per platform — **P0**, Phase 2
> "Write a post about our new feature and adapt it for each platform"

`context_get` → AI drafts per-platform variants → per-platform preview → `posts_validate` → `posts_create`

The tone difference is the whole value here: LinkedIn professional and longer, Instagram visual and hashtag-led, X short and punchy, TikTok casual with a hook. A single identical caption blasted everywhere is what users already hate about scheduling tools.

**Edge cases:** user edits one variant and expects the others untouched · hashtags on IG/LI but not LinkedIn-heavy · character limits differ per platform · first comment for hashtags on Instagram

### UC-P03 · Post from the media library — **P0**, Phase 2
> "Use the beach photo from my library and make an Instagram post"
> "I just uploaded new product shots — create posts from them"

`media_list` (search / recent) → **thumbnail grid** → user picks → AI writes caption from context → `posts_create` with `media.media_ids[]`

**Edge cases:** vague reference ("the blue one") → show grid, let them click · media wrong aspect ratio for the post type · video too long for the platform · multiple images → carousel vs separate posts, ask

### UC-P04 · Post from AI-generated media — **P0**, Phase 2
> "Make an image of a coffee cup on a wooden table and post it to Instagram"

Existing image/video gen tool → `media_upload` (auto-save to library) → `posts_create`

**Non-negotiable:** every AI-generated asset saves to the media library, and `tiktok_options.is_aigc` is forced `true` by the backend for TikTok. This is a platform compliance requirement, not a preference.

### UC-P05 · Bulk schedule a set — **P1**, Phase 2
> "Schedule these 5 posts across next week at the best times"

AI drafts all → **one combined preview listing all 5 with times** → N × `posts_create` (batched server-side)

**Edge cases:** partial failure (3 succeed, 2 fail) → report exactly which, offer retry on the failures only, never silently swallow · rate limits · one confirmation for the batch, not five

### UC-P06 · Create as draft — **P0**, Phase 2
> "Draft a post about X, don't schedule it yet"

`posts_create` with `publish_type: "draft"`. Downgrades the tier to near-green — drafts publish nothing.

### UC-P07 · Send for approval — **P1**, Phase 2
> "Create this post and send it to Sarah for approval"

`context_get` (team) → resolve name → `posts_create` with `approval.approvers[]`, `approve_option`, `notes`

**Edge cases:** approver not a workspace member · creator cannot be their own approver (API rule) · anyone vs everyone → default `anyone`, mention it

### UC-P08 · Approve / reject from chat — **P1**, Phase 2
> "What's waiting for my approval?" → "Approve the first two, reject the third — the image is wrong"

`posts_list` (`approval_assigned_to[]` = me) → `posts_approve` per post

**Edge cases:** tool only offered if the user is actually an approver · rejection should carry a comment — prompt for one if missing

### UC-P09 · Edit an existing post — **P0**, Phase 2 — **BLOCKED**
> "Change the caption on tomorrow's post"
> "Move Friday's LinkedIn post to Monday"

`posts_list` → identify → `posts_update` → confirm

**⚠️ No update endpoint exists.** This is the second thing every user will try. Delete-and-recreate loses the ID, approval state and comment thread. **[API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #1 — P0.**

### UC-P10 · Delete a post — **P1**, Phase 2
> "Delete the post scheduled for Friday"

`posts_list` → disambiguate → 🔴 confirm naming the post → `posts_delete`

**Edge cases:** multiple matches → never guess, list them · already published → deletion may not remove it from the platform; say so before confirming

### UC-P11 · Schedule into a content category — **P1**, Phase 2
> "Add this to my Tips & Tricks category"

`content_categories_list` → `posts_create` with `content_category_id` + `publish_type: "content_category"`. Category supplies accounts and timing.

### UC-P12 · Queue a post — **P2**, Phase 2
> "Add this to the queue"

`posts_create` with `publish_type: "queued"`. **Gap:** nothing exposes the queue or the next free slot, so the AI cannot tell the user when it will actually go out. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #7.

### UC-P13 · Threads and multi-part posts — **P1**, Phase 2
> "Turn this article into a Twitter thread"

AI splits into parts → `twitter_options.threaded_tweets[]` or `threads_options.multi_threads[]` → preview each part → `posts_create`

**Edge cases:** each part must respect the character limit independently · media per part · thread ordering must be visible in the preview

### UC-P14 · Facebook carousel — **P2**, Phase 2
> "Make a carousel of our 5 products with links"

`media_list` → build cards → `post_type: "carousel"` **plus** `facebook_options.carousel.is_carousel_post = true` with 2–10 cards.

Missing that second flag is a 400. Put it in the tool description in capitals.

### UC-P15 · First comment — **P1**, Phase 2
> "Post this and put the hashtags in the first comment"

`posts_create` with `first_comment.message` + `accounts[]` (subset of main accounts).

---

## Domain B — Calendar & Content Operations

### UC-C01 · What's coming up — **P0**, Phase 1
> "What's scheduled this week?" · "Show me tomorrow's posts" · "What's going out on Instagram Friday?"

`posts_list` (`status[]=scheduled`, date range) → **calendar or list component**

**Edge cases:** empty → say so and offer to fill it, don't just return nothing · relative dates resolved in workspace timezone

### UC-C02 · Find gaps in the calendar — **P1**, Phase 1
> "Which days next week have nothing scheduled?"

`posts_list` (next 7 days) → group by day → highlight empty → offer to fill

Strong agent candidate (see [Agents & Skills](https://app.helpin.ai/share/f4cb43a256cbad3ee529632028d38f30), Content Gap Watcher).

### UC-C03 · Filter by campaign / label / category / author — **P1**, Phase 1
> "Show me everything in the Black Friday campaign"
> "What has Ali scheduled?"

`context_get` name→id → `posts_list` with `campaigns[]` / `labels[]` / `content_category[]` / `created_by[]`

### UC-C04 · Approval queue status — **P1**, Phase 1
> "What's waiting for approval?" · "What have I submitted that's still pending?"

`posts_list` with `approval_assigned_to[]` or `approval_requested_by[]`

### UC-C05 · Unresolved feedback — **P2**, Phase 1
> "Which posts have unresolved comments?"

`posts_list` with `comment_status=unresolved`

### UC-C06 · Reschedule in bulk — **P2**, Phase 2 — **BLOCKED**
> "Push everything next week forward by one day"

Needs `posts_update` plus bulk. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gaps #1 and #10.

### UC-C07 · Create a campaign and fill it — **P1**, Phase 2
> "Create a Summer Sale campaign and put these 4 posts in it"

`campaigns_manage` (create) → `posts_create` × 4 with `campaign_id`

`color` is required — default it server-side.

### UC-C08 · Organise with labels — **P2**, Phase 2
> "Label all of this week's video posts as 'Video'"

`labels_manage` → apply. Blocked for existing posts by gap #1.

---

## Domain C — Analytics & Reporting

### UC-A01 · Overall performance — **P0**, Phase 1
> "How did we do last week?" · "Give me a snapshot across all accounts"

`analytics_workspace_overview` → **KPI grid with period-over-period deltas**

**Depends on the composite tool ([Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §6.7).** Without it this is 12+ calls and 30 seconds. This is the single most common analytics question in the product.

### UC-A02 · Single platform deep dive — **P0**, Phase 1
> "How is Instagram doing this month?"

`analytics_summary` → `analytics_trend(audience_growth)` → `analytics_top_posts` → narrative + chart

### UC-A03 · Which account is winning — **P0**, Phase 1
> "Which account performed best last quarter?" · "Which is underperforming?"

`analytics_compare_accounts` → ranked table

**Must normalise to rates.** Raw engagement always crowns the biggest account, which is not the question.

### UC-A04 · Top and worst posts — **P0**, Phase 1
> "What were my top 5 posts this month?" · "Which posts flopped?"

`analytics_top_posts` (`direction: top|least`)

**Edge case:** `least` only exists for X, YouTube, TikTok and Pinterest. For FB/IG/LI, fetch top with ascending `order_by` where supported, otherwise say plainly that we can't rank worst on that platform. Do not fabricate.

### UC-A05 · Why did this post do well — **P1**, Phase 1
> "Why did the Tuesday reel get so much reach?"

`analytics_get_post` → compare against `analytics_summary` averages → explanation

### UC-A06 · Audience demographics — **P1**, Phase 1
> "Who follows us on Instagram?" · "Where is my audience?"

`analytics_breakdown(demographics | location)` — FB, IG, LI only

### UC-A07 · Best time to post — **P0**, Phase 1 — **PARTIALLY BLOCKED**
> "When should I post?"

`analytics_breakdown(active_hours)` — **Facebook and Instagram only.**

For X, LinkedIn, TikTok, YouTube, Pinterest we have no data. Interim: derive from our own published-post performance via `analytics_top_posts` + publish timestamps and label it clearly as an estimate. Proper fix in [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #5. **This is a flagship workflow with a hole in it — worth prioritising.**

### UC-A08 · Hashtag performance — **P2**, Phase 1
> "Which hashtags work best for us?"

`analytics_breakdown(hashtags)` — IG, LI

### UC-A09 · Period comparison — **P1**, Phase 1
> "Compare this month to last month"

`analytics_summary` already returns current/previous/difference/percentage. One call, not two.

### UC-A10 · Client-ready report — **P0**, Phase 1
> "Build me a report for the client covering last month"

`analytics_workspace_overview` → `analytics_top_posts` per account → `analytics_breakdown` → formatted narrative report

Highest-value output for agencies. Should be exportable (PDF / shareable link) and is the obvious flagship default **Skill** (`/client-report`).

### UC-A11 · Google Business reviews — **P2**, Phase 1
> "How are our reviews looking?" · "Reply to that 2-star review"

`reviews_list` to read → `inbox_reply` (`POST /inbox/reviews/{review_id}/reply`) to answer.

**Corrected in v1.1:** replies are supported. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) gap #9 is resolved, so this use case is end-to-end, not read-only. Reply always gets a 🟡 preview — a public review response cannot be unsent.

### UC-A12 · YouTube specifics — **P2**, Phase 1
> "How are people finding my videos?" · "What's my watch time trend?"

`analytics_breakdown(traffic_sources | sharing_platforms)` · `analytics_trend(watch_time | views)`

---

## Domain D — Inbox & Engagement (Phase 3, **not blocked**)

> **v1.1:** the Inbox API is live — 28 endpoints. Every use case below is buildable now. Tool names here follow the v1.0 six-tool sketch; map them onto the re-cut nine-tool set in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §8 (`inbox_list_items`, `inbox_summary`, `inbox_get_thread`, `inbox_reply`, `inbox_triage`, `inbox_assign`, `inbox_tags`, `inbox_notes`, `inbox_moderate_comment`). Note UC-I04 is unblocked specifically: comments are filterable by source post via `GET /inbox/posts/{post_id}/comments`.

### UC-I01 · What's new — **P0**
> "Any new messages?" · "What came in overnight?"

`inbox_list_conversations` (status=new) → grouped by account and type

### UC-I02 · Read a thread — **P0**
> "Show me the conversation with @username"

`inbox_get_conversation`

### UC-I03 · Reply — **P0**
> "Reply and tell them we ship on Tuesday"

AI drafts in brand voice → 🟡 preview → `inbox_reply`

**Always preview.** A message sent to a real customer cannot be unsent.

### UC-I04 · Comments on a post — **P0**
> "Show me the comments on yesterday's Instagram post and reply to them"

`posts_list` → `inbox_list_conversations` (type=comment, filtered to that post) → draft replies → **batch preview** → `inbox_reply` ×N

**This is the one users will phrase as "post comments."** The model must not reach for `posts_internal_notes`. Tool descriptions must make the distinction explicit.

### UC-I05 · Triage and prioritise — **P1**
> "What needs my attention?" · "Which messages are actually worth replying to?"

`inbox_list_conversations` → AI scores by sentiment, intent, effort, recency → ranked list

Vista's `/triage-inbox` skill. Direct competitive parity target.

### UC-I06 · Bulk clear — **P1**
> "Mark all the spam as done"

`inbox_bulk_action`. Needs a real bulk endpoint.

### UC-I07 · Assign — **P1**
> "Assign all the support questions to Ali"

`inbox_assign`

### UC-I08 · Sentiment watch — **P1**
> "Is anyone angry with us right now?"

`inbox_list_conversations` + sentiment field → flag negatives

Requires sentiment on the API. Feeds the Sentiment Spike Watcher agent.

---

## Domain E — Media Library

### UC-M01 · Find media — **P1**, Phase 1
> "Show me my recent uploads" · "Find the product photos"

`media_list` (`search`, `type`, `sort`) → thumbnail grid

### UC-M02 · Upload from chat — **P1**, Phase 2
> [drags a file] "Add this to my library and post it to Instagram"

`media_upload` → `posts_create`

### UC-M03 · Generate and save — **P0**, Phase 2
> "Create an image for a post about sustainability"

Existing gen tool → `media_upload` → available for reuse

### UC-M04 · Clean up — **P2** — **BLOCKED**
> "Delete the old media from last year"

No delete endpoint. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #11.

---

## Domain F — Accounts & Connections

### UC-N01 · What's connected — **P0**, Phase 1
> "What accounts do I have connected?"

`accounts_list` → grouped by platform with status

### UC-N02 · Connect an account — **P1**, Phase 2
> "I want to connect my TikTok"

`platforms_list` → `accounts_connect` → **render a Connect button** → user completes OAuth → confirm in-thread → refresh `context_get`

**Edge cases:** OAuth happens outside the chat — needs completion detection so the AI can confirm rather than leaving the thread hanging · `return_url` back to the chat session · plan account limit reached → explain and offer upgrade

### UC-N03 · Fix a broken account — **P0**, Phase 2
> "Why did my Instagram post fail?" · "Which accounts need reconnecting?"

`accounts_list` (status) → `accounts_connect` with `process=reconnect`

**Expired tokens are a top support-ticket driver.** The AI catching this proactively is real, measurable value — and a strong agent ([Agents & Skills](https://app.helpin.ai/share/f4cb43a256cbad3ee529632028d38f30), Publishing Health Monitor).

### UC-N04 · Credential connections — **P2**, Phase 2
> "Connect my Bluesky"

`accounts_add_credentials`. **Credentials must be entered through a secure input component, never typed into the chat transcript.**

### UC-N05 · Platform capabilities — **P2**, Phase 1
> "Can I schedule stories on Instagram?" · "What can I post to Pinterest?"

`platforms_list` + static capability knowledge. Mostly answerable without a tool call, but the model must not invent capabilities we don't support.

---

## Domain G — Team, Workspace & Governance

### UC-G01 · Who's on the team — **P1**, Phase 1
> "Who has access to this workspace?"

`team_manage(list)`

### UC-G02 · Invite someone — **P2**, Phase 2
> "Invite sarah@company.com as an approver"

`team_manage(invite)` · 🟡 confirm · admin-only

### UC-G03 · Switch or list workspaces — **P1**, Phase 1
> "Switch to the Acme workspace" · "How many workspaces do I have?"

`workspaces_manage(list)` → switch context → refresh `context_get`

**Edge case:** switching mid-conversation invalidates every resolved ID in the thread. State the switch clearly and re-bootstrap.

### UC-G04 · New client workspace — **P2**, Phase 2
> "Set up a workspace for our new client Acme"

`workspaces_manage(create)` — `logo` and `timezone` are required. Prompt for logo upload or generate a placeholder.

### UC-G05 · Cross-workspace question — **P2**, Phase 2 — **BLOCKED**
> "Which of my clients had the best month?"

No cross-workspace analytics. Major agency ask. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #14.

---

## Domain H — Flagship Multi-Step Workflows

These are the demo moments. Each chains many tools and shows why chat beats clicking.

### W1 · Niche → trends → campaign → scheduled posts — **P0 aspiration, Phase 2+**
> "Look at my niche, find what's trending, and build me a week of content"

```
context_get (accounts, niche)
  → [Discover / trending — ⚠️ NO API]
  → AI generates 5–7 content ideas
  → user picks
  → campaigns_manage(create)
  → AI writes per-platform copy
  → media generation where needed → media_upload
  → analytics_breakdown(active_hours) for timing  ⚠️ FB/IG only
  → combined preview of the full week
  → posts_create × N with campaign_id
```

**This is the flagship, and it has two holes: no trending API and no best-time data on 6 of 8 platforms.** Both are in [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) (gaps #8 and #5). Discover is our biggest genuine differentiator over Vista Social and it is currently unreachable from the API. Worth pulling forward.

### W2 · Media dump → posts — **P0**, Phase 2
> "I uploaded 10 photos, turn them into a week of Instagram posts"

`media_list` → AI writes a caption per image → best-time spread → batch preview → `posts_create` × 10

### W3 · Weekly review → next week's plan — **P0**, Phase 2
> "How did last week go, and what should I post next week?"

`analytics_workspace_overview` → `analytics_top_posts` → AI finds the pattern in the winners → proposes next week's mix → creates drafts

**The most defensible workflow we can build.** Analysis feeding creation, in one conversation, is genuinely hard for a human to do quickly and is exactly what a chat interface is good at.

### W4 · Repurpose a winner — **P1**, Phase 2
> "My best post last month — turn it into content for the other platforms"

`analytics_top_posts` → `analytics_get_post` → adapt per platform → `posts_create`

### W5 · New user onboarding — **P1**, Phase 2
> "Help me get set up"

`accounts_list` (empty) → guided `accounts_connect` × N → ask about brand and niche → `campaigns_manage(create)` → generate first week → schedule

Turns an empty state into a scheduled calendar in one conversation. Direct activation-rate impact.

### W6 · Morning briefing — **P0**, Phase 3 → becomes the flagship Agent
> "What do I need to know this morning?"

`posts_list` (today) → `inbox_list_conversations` (new) → `analytics_workspace_overview` (yesterday) → `accounts_list` (health) → one digest

Vista's "Theo." This should be our **default agent**, pre-installed and enabled for every workspace on Phase 4 launch. It is the single best demonstration that the automation is working.

### W7 · Client report — **P0**, Phase 2 → becomes the flagship Skill
> "/client-report Acme, last month"

`analytics_workspace_overview` → per-account `analytics_summary` → `analytics_top_posts` → `analytics_breakdown` → formatted, branded, exportable report

### W8 · Inbox zero — **P1**, Phase 3
> "Help me clear my inbox"

`inbox_list_conversations` → triage and rank → draft replies for the top set → batch preview → `inbox_reply` × N → `inbox_bulk_action` for the rest

---

## Cross-cutting behaviours

These apply to every use case above and should each be a story in the Phase 0/1 epic.

| # | Behaviour | Rule |
|---|---|---|
| 1 | **Ambiguous account** | Never guess. Render an account picker. |
| 2 | **Relative dates** | Always resolve in workspace timezone, always state the resolved date back ("Friday 1 Aug"). |
| 3 | **Missing required field** | Ask one focused question, not a form. Never fail a call to discover it. |
| 4 | **Partial batch failure** | Report exactly which items failed and why. Offer retry on failures only. |
| 5 | **Permission denied** | Explain what the user lacks and who can grant it. Never expose the tool as unavailable-for-mystery-reasons. |
| 6 | **Expired token** | Detect, explain, offer reconnect inline. |
| 7 | **Plan limit hit** | Explain the limit, show the upgrade path. |
| 8 | **Rate limit** | Back off, tell the user, retry — do not silently drop. |
| 9 | **Empty result** | Say so plainly and offer the next useful action. Never an empty table. |
| 10 | **Long-running work** | Stream progress ("Fetching Instagram… 3 of 12 accounts"). Silence past 3 seconds reads as broken. |
| 11 | **User changes their mind mid-flow** | Preview state must be editable without restarting the whole conversation. |
| 12 | **Workspace switch mid-thread** | Invalidate all resolved IDs, re-bootstrap, say so. |

---

## Priority rollup for the Phase 1 epic

**Read-only, ship first, cannot break anything:**

UC-C01, UC-C02, UC-C03, UC-C04 · UC-A01, UC-A02, UC-A03, UC-A04, UC-A05, UC-A07, UC-A09, UC-A10 · UC-M01 · UC-N01, UC-N03 (detection only) · UC-G01, UC-G03

**17 use cases. 14 tools. Zero write risk.** This is the beta.

> **v1.1 note on where the work actually is.** The calendar-side reads (UC-C01–C04, UC-G01) are largely already served by the live `mcp_fetch_*` tools. **Every UC-A analytics use case is unserved** — 99 endpoints with no chat exposure. So of the 17, roughly 8 are new build and they are almost all analytics. Scope the Phase 1 epic accordingly: it is an analytics epic with a few reads attached, not an even spread. See [As-Built Baseline](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d) §6.
