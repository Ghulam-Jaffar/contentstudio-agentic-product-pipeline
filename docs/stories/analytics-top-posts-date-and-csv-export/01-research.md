# Research: Published date column and CSV export for analytics Top Posts

**Date:** 2026-09-11
**Deliverable:** `02-stories.md`

This doc stays local. Codebase paths, entry points, gotchas and open questions live here and never in a story body.

---

## 1. The ask

In analytics, on the **Top Posts table** (the Posts tab surface with the "Top 100" selector, not the five-post cards on the Overview):

1. Add a **published date** column to the table.
2. Show the published date in the **post preview panel**, in the list of stats on the right.
3. Add **CSV export** of the table, so a user can export up to 100 top posts, sort by date or anything else in a spreadsheet, and share it.
4. Apply this to **every analytics surface that has a posts tab**, not just Facebook.
5. Put the export control at the **far right of the table header, next to the "Top 100" selector**.

The stated purpose of the export matters for the design: the user wants to **sort by date** after exporting. That is a constraint on the date format written into the CSV, not just a nice-to-have. See section 5.

---

## 2. Backlog dedupe

Checked `docs/stories/` and `docs/features/` before writing. Nothing covers this. Three deliverables are adjacent and need coordinating rather than duplicating:

| Existing deliverable | Relationship |
|---|---|
| `docs/stories/analytics-top-least-posts-consistency/` | **Closest, and it touches the same header area.** It adds a "Top posts by {metric}" dropdown plus a Least Posts section to Facebook, Instagram, LinkedIn and TikTok, to match Pinterest and YouTube. Important detail: that story is about the **Overview's five-post Top Posts section**, and its backend half is the **Go analytics service**, not Laravel. It does not touch the Posts tab table, the "Top 100" selector, columns, or export. If the two are in flight together, the header row of the table is the one place they could collide. |
| `docs/stories/analytics-report-sections-and-overview-cards/` | Its second story caps Overview top posts at five, which confirms the Overview cards and the Posts tab table are separate surfaces with separate limits. No export or date scope. |
| `docs/stories/analytics-report-date-time-preferences/` | **Directly relevant.** Users have a saved **date format** and **clock format** preference, applied to analytics reports: four date formats (month-first short, month-first long, year-first, day-first) and two clock formats (12 and 24 hour), with a documented fallback when unset. A new date column and an exported date must respect the same preferences or the product will contradict itself. |
| `docs/features/gmb-analytics/` | Useful precedent: its Posts tab already specifies top-performing posts showing "preview text, post type badge, **date**, views, CTA clicks". So a published date is already considered part of a top-posts row on at least one platform, which suggests the data is available rather than missing. |
| `docs/stories/post-previews-inbox-analytics/` | Adopted shared post previews across Analytics. The preview panel this feature adds a date to is likely that shared component, which means one change covers every platform. To be confirmed from the code. |
| `docs/stories/analytics-report-email-copy-and-download/` | Analytics already has a queued export path for PDF reports delivered by email. Relevant only as a pattern to compare against, since a 100-row CSV is small enough not to need it. |

---

## 3. Two different surfaces, and this feature targets one of them

Worth stating plainly because the naming is confusing:

- **Overview Top Posts section** shows five post cards, has a "Top posts by {metric}" dropdown on some platforms, and is capped at five by an existing story. **Not this feature.**
- **Posts tab Top Posts table** shows a full table with columns (Posts, Media Type, Engagements, Impressions, Reach, Post Clicks, Reactions, Comments) and a "Top 100" selector at the top right. **This is the surface.**

The export control goes in the table header, to the right of the "Top 100" selector.

---

## 4. Frontend: current state

### The table is one shared component, but its columns are declared per platform

`contentstudio-frontend/src/modules/analytics/views/common/AnalyticsPostsTable.vue` is the single shared Top Posts table. There are **no per-platform copies**. It is consumed by ten platform sections plus the "show more" modals:

| Platform | Section |
|---|---|
| Facebook, Instagram, LinkedIn, Twitter/X, Pinterest, YouTube, Threads, Bluesky, GMB | `views/<platform>/components/PostsSection.vue` |
| TikTok | `views/tiktok/components/PostTable.vue`, wrapped by its `PostsSection.vue` |
| Show-more modals | `views/{facebook,instagram,linkedin,tiktok,youtube}/components/TopPostsModal.vue` |

**The catch:** columns are not a config array. The shared table takes `validHeadersList: string[]` and three mutator callbacks, and builds its columns from that list. Every non-thumbnail cell falls through to a single numeric-styled span run through `mutateBodyValues`, and **every existing formatter in `bodyValuesMap` is `formatNumber()`**. So a date column is the first non-numeric cell the table has ever rendered.

Adding a column therefore means editing **ten platform composables**, roughly five touch points each:

1. `validPostsTableHeaders` (and `validPostsTableHeadersReport`, which exists separately for Facebook, Instagram and LinkedIn)
2. `headerTitles`
3. `headerTooltips`
4. `bodyValuesMap`, needing its first date formatter
5. `nonSortableItems`, or a real `order_by` API key

The composables are `views/<platform>/composables/use<Platform>Analytics.ts`. Watch three inconsistencies: TikTok uses `coverImage` rather than `thumbnail` as its thumbnail key, GMB exports `postsTableHeaders` rather than `validPostsTableHeaders`, and sorting is a mixed model that both sorts locally and emits `updateSort` mapped to an API `order_by`.

### The published date is already on every row

This is the most useful finding. Every platform's row transform already hoists a published timestamp onto the row object. It is simply never listed in the headers array.

| Platform | Raw API field | Row field |
|---|---|---|
| Facebook | `created_at` | `createdAt` |
| Instagram | `created_at` | **`postCreatedAt`** |
| LinkedIn | `published_at` and `created_at` | `createdAt`, from `created_at` |
| Twitter/X | `created_at` | `createdAt` |
| YouTube | `created_at` | `createdAt` |
| Pinterest | `created_at` | `createdAt` optional |
| TikTok | `created_time` | `createdAt` |
| Threads | `published_at` | `createdAt` |
| Bluesky | `published_at`, plus optional `indexed_at` | `createdAt` |
| GMB/GBP | `created_at` | **`created_at`**, stays snake_case |
| Campaign and Label | `published_at` | `published_at`, and **already rendered** |

So **no backend work is needed to obtain the date.** The work is rendering it and normalizing the name.

Two gotchas:

- **Name drift.** Nine platforms expose `createdAt`, Instagram exposes `postCreatedAt`, GMB exposes `created_at`. The shared table's own `PostRow` type already declares `published_at?: string`, used only for the row key and never populated by any platform. That is the obvious normalized name to adopt, so the shared table can read one field instead of ten.
- **Do not confuse sync timestamps with publish time.** Facebook, Instagram and LinkedIn also carry `saved_at`, and Threads and Bluesky carry `last_synced_at`. Those are when ContentStudio fetched the row, not when the post went out.

### The post preview is a modal, and its stats list is a config array

Clicking the **first cell only** of a row opens a modal, not a side drawer. Shell is `views/common/`-adjacent `components/competitor/PerformancePostPreviewModal.vue`, with the preview on the left and a stats sidebar on the right. Per-platform wrappers are `views/<platform>/components/<Platform>PostModal.vue`.

The sidebar stats list **is** a proper config array, `postModalFields: PostModalField[]`, declared per platform as `{ label, key, iconSrc, iconClass, tooltip }`. Adding a published date row is a one-line addition per platform, with one obstacle: the template renders `{{ selectedPost[key] }}` raw, so a formatted date needs either a pre-formatted field on the row or a `format` hook added to `PostModalField`. That interface is **redeclared in each of eight platform composables** with no shared type, so adding a hook means editing all eight.

Note the preview itself already shows a timestamp. The shared published-post preview work from `docs/stories/post-previews-inbox-analytics/` has shipped: `modules/analytics/utils/publishedPostSource.ts` has a mapper per surface, every mapper already sets `created_time`, and the preview chrome renders it next to the account name. What is missing is the date in the **stats sidebar**, which is what the ask is about.

### Date formatting already respects user preferences

`src/composables/useDateTime.ts` plus a Day.js prototype plugin in `src/utils/dayjs.ts` is the single entry point, and the repo rules mandate `createDate()` over Moment or bare `new Date()`.

- `profile.date_format`, one of `MM/DD/YY`, `MMM DD, YYYY`, `YYYY MMM DD`, `DD MMM YYYY`
- `profile.time_format`, `12h` or `24h`
- Workspace timezone from the active workspace, defaulting to UTC

The correct call is `createDate(value, false, true).inWorkspaceTimezone().formatDate()`. The third argument parses the raw value as UTC before converting, which matters because the stored values are UTC. Using `.formatDate()` without `.inWorkspaceTimezone()` applies the user's format but leaves the date in UTC.

Two things to know: there is a **fallback discrepancy**, where `useDateTime`'s default is `MMM DD, YYYY` but the prototype's bare fallback is `YYYY-MM-DD`, so they disagree before the config is wired. And the companion story `docs/stories/analytics-report-date-time-preferences/` is backend-scoped, because the web app already honours all three preferences. So this feature needs no preference plumbing.

### Precedent worth knowing, and why this ask deliberately differs from it

The team's most recent analytics posts table, the Campaign and Label Posts tab, **already renders a published date and deliberately chose not to give it a column.** It puts it as a secondary line inside the post cell, next to the account name, with an explicit code comment saying it identifies the post "without spending three more columns".

The ask here is explicitly for a column, which is the right call for a table the user intends to export and sort. Worth noting the divergence for the design conversation, and worth considering whether the date should *also* appear under the caption in the post cell, since that costs nothing and helps scanning.

### There is no CSV export anywhere in the frontend, and no CSV library

- `package.json` contains **no** CSV or spreadsheet library. No `papaparse`, no `xlsx`, no `file-saver`.
- There is **zero client-side CSV generation** in `src/`. Nothing matches `text/csv`. `new Blob` appears six times, all for PDF, audio or images.
- The one CSV download in the product is **server-generated**: the API request logs export. The backend returns a blob and the frontend only triggers the save, via an object URL and a synthetic anchor.
- The media library "Export as CSV" is **server-side and delivered by email**. No file ever reaches the browser. It alerts that an email will be sent.
- Social listening's export modal offers a CSV option but dispatches a server-side mutation.
- The analytics export control that exists, `views/common/ExportButton.vue`, is a dropdown offering **Export PDF, Email PDF and Schedule PDF**, gated on the `exports_schedule_reports` plan feature. It is PDF only despite some docs claiming otherwise.

**Consequences for the spec.** A client-side Top Posts CSV would be the first in the codebase. Three routes:

1. **Hand-rolled client-side util, no new dependency.** Recommended. The rows are already in memory, a hundred rows is trivial, and the user gets an instant download instead of an email. Costs about thirty lines of correct escaping.
2. **Add a CSV dependency.** The repo rules flag a new dependency as a stop-and-ask item, and it is not worth it for one table.
3. **Server-side, following the media library.** Heaviest, and delivering a hundred-row table by email is a worse experience than a click.

One real constraint on option 1: **the table only holds the current page in memory.** The limit selector controls a server-side refetch, so at a limit of 10 the browser has 10 rows, not 100. "Export the table" and "export the top 100" are therefore different features. See the open questions.

### Scope, precisely

Both route trees mount the same components, so `/analyze/` and the older `/analytics/` tree do **not** double the work.

Surfaces with a Posts tab that uses the shared table, all ten in scope: Facebook, Instagram, LinkedIn, Twitter/X, Pinterest, YouTube, TikTok, Threads, Bluesky, GMB/GBP.

Surfaces that render posts but **not** through this table, and therefore need a decision rather than assumed inclusion:

- **Campaign and Label performance report Posts tab** uses a newer, better table stack, `views/common/data_table/AnalyticsDataTable.vue`, which does have a real column config array with server-persisted column visibility and a toolbar with slots. It already shows a published date. It would only need the export, and its toolbar has a `#toolbar-trailing` slot ready for it.
- **Overview and group top posts** is a card grid with per-platform tabs, no table.
- **Competitor analytics** top and least performing posts are card grids, no table.
- **Per-platform Overview "Top posts"** is a five-card grid, whose "show more" modal *does* mount the shared table. So the modal inherits both the column and the export for free, which is a bonus rather than a decision.

---

## 5. Backend: current state

### The table is served by the Go analytics service, not Laravel

The most important structural fact. The Vue app calls the **Go analytics service directly**, on its own base URL from `VITE_ANALYTICS_GO_URL`. Every top-posts URL in `contentstudio-frontend/src/api/analytics-store.ts` and `src/api/analytics.ts` is built on that base, for example `${analyticsGoBaseUrl}analytics/overview/facebook/getTopPosts`.

There are three backend surfaces and it matters which one the work lands in:

| Surface | Shape | Who calls it | Data source |
|---|---|---|---|
| **Go analytics service** | `GET /analytics/overview/{platform}/{topPosts, sortedTopPosts, getTopPosts}` | the Vue app and share links | ClickHouse |
| **Laravel public API v1** | `GET /api/v1/workspaces/{id}/analytics/{platform}/top-posts` | external API customers | a verbatim proxy to the Go service |
| **Laravel legacy web routes** | `POST /analytics/overview/{platform}/getTopPosts` | nothing in the current frontend | raw ClickHouse SQL in PHP |

All new work lands in the **Go service**. The public API proxy needs no change to inherit a new field. The legacy Laravel implementation is a parallel dead path and should be left alone.

Routes are registered per platform in `contentstudio-social-analytics-go/src/api/analytics/router.go`, with a handler package per platform. Data lives in **ClickHouse**, database `contentstudiobackend`, not MongoDB.

### The published date is already returned, inconsistently

Every platform already returns a published timestamp. The problem is that no two platforms agree on the name, the format, or whether the workspace timezone was applied.

| Platform | JSON field | Timezone applied | Format |
|---|---|---|---|
| Facebook | `created_at` | yes | RFC3339 |
| Instagram | `created_at` | yes | RFC3339 |
| LinkedIn | `created_at` **and** `published_at`, different columns | yes | RFC3339 |
| YouTube | `created_at`, but the struct field is `PublishedAt` | yes | RFC3339 |
| Pinterest | `created_at` | yes | RFC3339 |
| GMB/GBP | `created_at` | yes | RFC3339 |
| TikTok | `created_time` | yes | RFC3339 |
| Bluesky | `published_at` | **no, always UTC** | RFC3339 |
| Twitter/X | `created_at` | **no** | **raw `YYYY-MM-DD HH:MM:SS`** |
| Campaign and Label | `published_at` | yes | RFC3339 |

The timezone is applied by a helper that is **copy-pasted into eight service packages**, with Bluesky carrying its own variant that ignores the timezone argument entirely and TikTok carrying a string-input variant.

### Four real defects this feature exposes

These are not hypothetical. Shipping a visible date column and a CSV without fixing them would publish the inconsistency to customers.

1. **Twitter/X returns an unconverted, non-RFC3339 date.** The service assigns the raw ClickHouse string straight through. So X posts would display and export in UTC while every other platform shows workspace-local time, with a different string shape on top.
2. **Bluesky always returns UTC**, even when a timezone is supplied.
3. **On YouTube and Bluesky, the ClickHouse `created_at` column is the snapshot day, not the publish date.** The Go comments say so explicitly. Naming a date column off `created_at` blindly would be wrong on those two platforms. `published_at` is the correct column there.
4. **Twitter/X cannot sort by date at all.** Its order-by whitelist has no date field, and an unknown value silently falls back to engagement. Facebook accepts `order_by=created_time`, Instagram `post_created_at`, LinkedIn `created_at`, TikTok `sort_order=created_time`, Bluesky `published_at` or `created_at`. YouTube, Pinterest and GMB have no date token either.

Also worth knowing: `day_of_week` and `hour_of_day` are derived at ingest from the raw platform timestamp, so they are baked to whatever offset the platform returned and will not agree with a workspace-timezone date column at the edges.

### Limits

`limit` is a query parameter, clamped server-side to `maxLimit = 100` on every platform, which is exactly where the "Top 100" selector tops out. Handler defaults vary (15 for Facebook, Instagram, YouTube and GMB, 5 for TikTok, Twitter and Bluesky, 3 for LinkedIn). `offset` is supported on most platforms, so an export beyond 100 would need pagination. The Laravel proxy validates the same cap in its request DTO.

### Threads: confirmed to exist, but not on the branch that was searched

**Resolved. Threads analytics exists and is deployed to the QA environment.**

Worth recording because it will trip up the next person: a sweep of the Go analytics service found **no Threads implementation on the working branch**, while the frontend has a full Threads Posts section using the shared table, Threads types carry `published_at`, and the Laravel proxy references Threads routes. The explanation is that the Go side lives on a branch that has reached QA but not the branch that was searched.

Two practical consequences for whoever picks this up:

1. **Threads is in scope**, with no check needed.
2. **Do not conclude it is missing** if it is still absent from the branch you are on. Find the branch it lives on, and coordinate so the date normalization and the export land there rather than being written against a copy that does not yet exist. If Threads merges after this work starts, it needs the same treatment applied rather than inheriting it.

The documented Threads contract already returns `published_at`, so it fits the normalization without a special case.

### The CSV precedent to copy already exists, in the Go service

Social listening already has exactly this feature: `GET /api/listening/analytics/export?format=csv`. It is synchronous, streamed, and a direct download:

```go
w.Header().Set("Content-Type", "text/csv")
w.Header().Set("Content-Disposition", "attachment; filename=social_mentions_export.csv")
```

and in the service, the details worth copying verbatim:

```go
csvWriter := csv.NewWriter(w)
defer csvWriter.Flush()

// Write BOM for Excel compatibility with UTF-8
w.Write([]byte{0xEF, 0xBB, 0xBF})

header := []string{"Date", "Platform", "Topic", "Author", ...}
...
row.PostedAt.Format("2006-01-02 15:04:05"),
```

Note the three good decisions in there: a **UTF-8 BOM** so Excel does not mangle non-ASCII captions, a `Limit` of 100, and an **ISO-style sortable date**. One caveat visible in that handler: once the headers are written it cannot emit an error response, which the code comments on.

There is also a ready-made frontend precedent for triggering a server CSV download: the API request logs export in settings fetches a blob and saves it with an object URL and a synthetic anchor. So both halves of a server-side export already exist in the codebase.

Other export patterns, for completeness: Laravel has three `StreamedResponse` plus `fputcsv` endpoints (API request logs, webhook deliveries, wallet usage), all capped at 10,000 rows and all labelling their date column "Timestamp (UTC)" to sidestep formatting. The media library CSV is the **anti-pattern**: it writes a temp file into the process working directory, uploads to storage, emails a link, and only unlinks the temp file on email success. Do not copy it.

### Date and time preferences, and the Go package that already applies them

Preferences are **per user**, not per workspace: `user.date_format` holding a Day.js token and `user.time_format` holding `12h` or `24h`, with model defaults of `MMM DD, YYYY` and `12h` and a legacy fallback under `user.preferences`. The canonical reader is in the users repository and prefers the top-level attribute over the legacy location.

Crucially, **the Go service already has a package that maps these to Go layouts**, built for analytics report generation:

```go
const Default = "Jan 02, 2006"

var layouts = map[string]string{
	"MM/DD/YY":     "01/02/06",
	"MMM DD, YYYY": "Jan 02, 2006",
	"YYYY MMM DD":  "2006 Jan 02",
	"DD MMM YYYY":  "02 Jan 2006",
	"YYYY-MM-DD":   ISO,
	...
}
```

It also handles messy input shapes including TikTok's `"2026-08-24 00:00:00"` quirk. So if the export should honour the user's format, the mapping is already written and does not need inventing.

Workspace timezone is a workspace-level field defaulting to `America/New_York`, canonicalised through an alias table. For the live dashboard path it reaches the Go service as a **query parameter**, and there is no server-side workspace-timezone lookup in the analytics read path. **If the client omits the timezone, everything silently renders UTC.**

### Authorization, and one thing worth flagging

- The Laravel public API analytics routes use membership-only permission checking with no per-role gate.
- The Laravel legacy analytics web routes have **no permission middleware** at all on the per-platform blocks.
- The Go service authenticates the caller by share token, bearer JWT or API key, but **performs no workspace-membership or per-role check in the analytics read path**. Any authenticated user who knows a workspace id and platform id can read that data. A CSV export added there inherits that posture. Out of scope to fix, but it should be written down rather than discovered later.
- **There is no analytics-specific role restriction anywhere**, which means an **approver can read analytics and would be able to export**. If the export should be gated, a new action string would need adding to every role branch of the permission helper, and because the approver branch has no default case it would deny approvers unless explicitly allowed.
- **There is no rate limiting on the Go analytics service at all.** The public API v1 has a 100-per-minute throttle keyed by API key, but no analytics-specific limiter exists.
- Because the Go service is on a different origin from the app, a download response needs `Access-Control-Expose-Headers: Content-Disposition` for the browser to read the filename. Laravel's PDF download helpers already do this and are the precedent.

---

## 6. Recommendations

### Build the export in the Go service, not in the browser

Both halves of a server-side export already exist in the codebase, which settles it:

- The **Go social listening CSV export** is the server pattern: streamed, `text/csv`, `Content-Disposition`, UTF-8 BOM, capped at 100, sortable date.
- The **API request logs export** in settings is the frontend pattern: fetch a blob, save it with an object URL and a synthetic anchor.

A client-side CSV was the alternative. It is rejected for three reasons: it would be the first client-side CSV in the repo, the only CSV library route is a new dependency which the repo rules flag as stop-and-ask, and above all **the table only holds the currently loaded page in memory**. The limit selector triggers a server refetch, so at a limit of 10 the browser has ten rows. Since the ask is explicitly "export up to 100", the export has to be able to fetch 100 independently of what is on screen, which a server endpoint does naturally.

### Put a sortable date in the CSV, and the user's preferred format on screen

The stated purpose is to sort by date in a spreadsheet after exporting. Two of the four supported date formats, `MM/DD/YY` and `MMM DD, YYYY`, do not sort lexicographically and are not reliably parsed by spreadsheets. So:

- **On screen**, the column uses the user's saved date format and the workspace timezone, via the existing helper. Consistent with the rest of the product.
- **In the CSV**, the date is written as `YYYY-MM-DD HH:MM` in the workspace timezone, with the timezone named in the column header, for example `Published (Asia/Karachi)`. Spreadsheets parse that natively into a real date, which means it both sorts correctly and then displays in the reader's own locale. This is also what the existing listening export already chose.

The alternative is two columns, one formatted per preference and one ISO. Rejected as clutter for a hundred-row table, but it is a one-line change if the team prefers it.

### Normalize the field name once, in the Go response

Nine platforms return `created_at`, TikTok returns `created_time`, Bluesky and Campaign and Label return `published_at`, LinkedIn returns both, and the frontend then re-derives three different row field names on top. The shared table already declares an unpopulated `published_at` field, which is the obvious target.

Adding `published_at` to every platform's top-posts response, sourced from the correct column per platform, means the shared table reads one field instead of ten and the frontend needs no per-platform date plumbing. Keep the existing fields in place so nothing that reads them breaks.

### Fix the four defects in the same story as the date

Because the date becomes visible and exportable, the Twitter/X non-converted raw string, the Bluesky UTC-only behaviour, the YouTube and Bluesky snapshot-day confusion, and the eight copies of the timezone helper all become customer-visible problems rather than internal untidiness.

---

## 7. Decisions taken, and what is still open

### Decided

1. **Scope is every analytics posts table.** The eleven platform Posts tabs that use the shared table, being Facebook, Instagram, LinkedIn, Twitter/X, Pinterest, YouTube, TikTok, Threads, Bluesky and GMB/GBP, **plus the Campaign and Label performance report Posts tab**. Threads is confirmed and on QA, see section 5.
2. **The export exports the current selection.** Whatever number the selector is set to is exactly how many rows the CSV contains. The selector stays the single control, and the file always matches the screen.
3. **The published date is its own sortable column on every one of those tables**, Campaign and Label included, rather than a line under the caption. A column is what makes it sortable on screen and a real field in the export, which is the point of the feature. Campaign and Label shows the date under the caption today and gains a column like every other table. Whether it keeps the under-caption line as well is a minor tidiness call for design, and keeping it is the default.
4. **The export control is hidden on analytics share links.** A share-link viewer is not a workspace member, so they can read the table but cannot pull the data out of it. The date column itself still shows on share links.
5. **No extra gate on the export. Whoever can access analytics can export.** Not plan-gated, and not role-gated. The export carries exactly the access the top-posts endpoints already carry, so anyone who can read the table can download it. Deliberately different from the existing analytics PDF export, which is gated on a paid plan feature, because a table CSV is far cheaper to produce than a scheduled report.
6. **Twitter/X gains a published-date sort field.** Rather than leaving the column non-sortable on X alone, which users would notice as an inconsistency, the ordering whitelist gets a date field so the column sorts on every platform.
7. **Mobile stays out.** The Flutter app has no analytics posts table and a CSV download is not a mobile-shaped action. Recorded as a decision rather than an omission.

### Nothing is open

Every question raised by this research has been answered. Two consequences of decision 5 are worth writing down plainly rather than discovering later, neither of which changes the decision:

- **Approvers can export.** Analytics has no role gate at all today, so approvers can already read analytics, and tying the export to analytics access means they can export too. That is the intended reading of "whoever can access analytics can do this".
- **The export inherits the Go service's existing authorization posture.** That service authenticates the caller but performs no workspace-membership check in the analytics read path, so anyone authenticated who knows a workspace and platform id can already read that data, and will be able to export it. This is pre-existing and out of scope to fix here, but it should be on someone's list rather than nobody's.

### Not in scope, because there is no table to export

The Overview and group top posts, the competitor top and least performing posts, and the per-platform Overview "Top posts" section are all **card grids, not tables**. There is no column set to add a date to and no table to export. The per-platform Overview "show more" modal is the exception: it mounts the shared table, so it inherits both the column and the export automatically.

The Meta Ads and Google Ads tables use the same newer table stack as Campaign and Label, but they are campaign, keyword and ad tables rather than posts tables, so they are outside "analytics posts tables".

---

## 8. Story shape

Three stories, following the consolidation pattern used for the marketing feedback epics:

- **`[BE]`** one story in the Go analytics service: normalize the published date across platforms, fix the four defects, and add the streamed CSV export endpoint.
- **`[FE]`** one story: the date column on the shared table, the date in the post preview stats sidebar, and the export control in the table header next to the "Top 100" selector.
- **`[Design]`** one story: the column, the export control placement and its states.

No Flutter story. No Laravel story, since the public API proxy inherits the new field automatically and the legacy web routes are dead.
