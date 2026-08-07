# Research — Analytics UI Fixes (X duplicate button + Instagram Top Posts time)

Two independent, frontend-only bugs in the Analytics module. Both live in the active platform analytics views under `contentstudio-frontend/src/modules/analytics/views/` (the `analytics_v3` module only holds shared wrappers + competitor reports; the per-platform X/Instagram analytics screens are in `modules/analytics/`).

---

## Issue 1 — X (Twitter): duplicate "Auto sync schedule" control in top-right

### Current State
On the X (Twitter) analytics screen, the top-right toolbar renders the **"Auto sync schedule"** control **twice**.

The toolbar is built in `TabsComponent.vue`. Inside the absolutely-positioned top-right container (`div.absolute.right-4.top-0`) it renders, for X:
1. A **dedicated X block** (`v-if="type === 'twitter' && !isSharedAnalytics"`) — a bordered box containing `<PlatformTooltip>` (which shows the "Auto sync schedule" label) **plus** a Settings gear that opens the sync-config modal. — [TabsComponent.vue:40-63](contentstudio-frontend/src/modules/analytics/views/common/TabsComponent.vue#L40-L63)
2. A **manual "Sync now" button** (X-specific). — [TabsComponent.vue:64-85](contentstudio-frontend/src/modules/analytics/views/common/TabsComponent.vue#L64-L85)
3. A **generic `<PlatformTooltip>`** rendered for *every* platform (`v-if="!isSharedAnalytics"`). — [TabsComponent.vue:143-148](contentstudio-frontend/src/modules/analytics/views/common/TabsComponent.vue#L143-L148)

`PlatformTooltip` renders the literal **"Auto sync schedule"** text only when `type === 'twitter'`. — [PlatformTooltip.vue:170-172](contentstudio-frontend/src/modules/analytics/components/common/PlatformTooltip.vue#L170-L172)

**Result:** For X, both the dedicated block (#1) and the generic tooltip (#3) render "Auto sync schedule", so the label appears twice. For all other platforms only the generic one (#3) renders (as a plain Info icon), which is correct.

### What Needs to Change
- Stop the **generic** `PlatformTooltip` (#3) from rendering for X, since the **dedicated** X block (#1, which also carries the Settings gear for configuring the sync) already provides it. Keep exactly one "Auto sync schedule" control on X, and leave every other platform untouched.

### Files Involved
- [TabsComponent.vue](contentstudio-frontend/src/modules/analytics/views/common/TabsComponent.vue) — the duplicate render (line 143 block should be excluded for X)
- [PlatformTooltip.vue](contentstudio-frontend/src/modules/analytics/components/common/PlatformTooltip.vue) — where the "Auto sync schedule" label comes from (no change expected)

---

## Issue 2 — Instagram Analytics: Top Posts shows the current (live) time instead of the posted time

### Current State
On the Instagram Analytics **overview → Top Posts** card, each post shows a date/time under the account name. It currently displays **the current time (live)** instead of when the post was actually published.

Root cause chain:
- The card formats the date with `createDate(post.post_created_at, false, true).inWorkspaceTimezone().formatDateTime()`. — [TopPostCard.vue:78](contentstudio-frontend/src/modules/analytics/views/instagram_v2/components/TopPostCard.vue#L78)
- `createDate` **falls back to the current time** whenever its input is falsy: `return input ? (utcTime ? dayjs.utc(input) : dayjs(input)) : dayjs()`. — [useDateTime.ts:73-82](contentstudio-frontend/src/composables/useDateTime.ts#L73-L82)
- The Top Posts card is fed **raw, untransformed** overview data: `overviewTopPostsData.value = data?.top_posts` (the untransformed `OVERVIEW_TOP_POSTS` branch), unlike the detailed `TOP_POSTS` branch which runs `transformObject`. — [useInstagramAnalytics.js:217-219](contentstudio-frontend/src/modules/analytics/views/instagram_v2/composables/useInstagramAnalytics.js#L217-L219)
- `TopPosts.vue` passes each raw post straight into `TopPostCard`. — [TopPosts.vue:187-189](contentstudio-frontend/src/modules/analytics/views/instagram_v2/components/TopPosts.vue#L187-L189)

So `post.post_created_at` is empty/absent in the overview Top Posts payload, `createDate` silently returns `dayjs()` (= now), and the card renders the current time.

`post_created_at` is the canonical Instagram field used elsewhere (`getPostDate`, `AnalyticPreview`, `transformObject`), and the raw IG post object also carries `timestamp` and `stored_event_at` (both mapped in `transformObject`). Engineering should confirm which timestamp field the **overview top_posts** endpoint actually returns and bind to it.

### What Needs to Change
- Bind the Top Posts card date to the post's **actual publish/creation timestamp** returned for overview top posts.
- **Guard the empty case** so the card never falls back to the current time — when no timestamp is present, show nothing (or a neutral placeholder) rather than "now".

### Files Involved
- [TopPostCard.vue](contentstudio-frontend/src/modules/analytics/views/instagram_v2/components/TopPostCard.vue) — the date binding (line 78)
- [TopPosts.vue](contentstudio-frontend/src/modules/analytics/views/instagram_v2/components/TopPosts.vue) — feeds raw overview data to the card
- [useInstagramAnalytics.js](contentstudio-frontend/src/modules/analytics/views/instagram_v2/composables/useInstagramAnalytics.js) — `OVERVIEW_TOP_POSTS` branch (raw, untransformed data)
- [useDateTime.ts](contentstudio-frontend/src/composables/useDateTime.ts) — `createDate` fallback-to-now behavior (the trap to guard against)

---

## Summary
- Both are **frontend-only**, **web-only** (no mobile — Instagram/X analytics screens are web).
- Two separate, unrelated bugs → **two separate `[FE]` stories**.
- No backend, API, or schema change expected. Issue 2 assumes the correct timestamp is already present in the overview payload under some key; if it turns out the overview endpoint returns no timestamp at all, that would need a backend follow-up (flagged in the story).
- Product area: **Analytics**. No new UI copy beyond an optional empty-time placeholder.
