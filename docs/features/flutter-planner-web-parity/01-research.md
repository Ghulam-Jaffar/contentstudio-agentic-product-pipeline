# Research: Flutter Planner Web Parity

**Date:** 2026-09-24
**Source brief:** "Planner — Web vs Mobile Gap Analysis" (CONTENTSTUDIO MOBILE.pdf, September 2026), verified against code in this research.
**Scope:** Flutter app only (`contentstudio-flutter/`, ships iOS + Android). The web planner is the reference, not a work target.
**Related past work:** `docs/features/mobile-planner-calendar-view/` (native iOS/Android calendar, now superseded by Flutter). Not a duplicate. This epic is the Flutter follow-on for full parity.

---

## 1. Headline

The gap doc's summary holds after verification. Mobile covers the **consumption and single-post half** of the web planner well: list, calendar, core filters, preview, comments, approvals and basic per-post actions. It has almost none of the **collaboration, organisation and export half**: bulk operations, share links, notes, holidays, saved filters, custom views, PDF, post analytics, search, date range, realtime, and three of the web's view modes.

**About 31 web capabilities are missing or partial on mobile.** Two exist only on mobile: download media to the gallery, and the Hourly/Daily week variants plus the long-press day peek. These must be **kept**, not flattened to match the web.

**Most important finding for scoping: the backend already has every endpoint these gaps need.** This is almost entirely `[Flutter]` work. Backend work is limited to a live check of realtime auth with a mobile token, and possibly the analytics base URL.

---

## 2. Gap inventory (verified)

Legend: ✅ verified · ⚠️ partly right / corrected · ❌ gap doc wrong.

### 2.1 View modes
| View | Web | Mobile | Verdict |
|---|---|---|---|
| List | Yes (13 columns) | Yes (card list) | ✅ |
| Compact list | Yes | Missing | ✅ |
| Calendar month / week | Yes | Yes (+ Hourly/Daily week variants, mobile-only) | ✅ |
| Feed | Yes (desktop-only on web) | Missing | ✅ |
| Instagram / TikTok grid | Yes, gated on `grid_view_planning` | Missing; no entitlement key on mobile | ✅ |

### 2.2 Filtering, search, sort
- **Present on mobile:** status, sort, accounts (sheet with search and fixed grouping by platform; no "Group by" option), content type, labels (chips, **no search**), assigned to, created by, approval requested by / assigned to.
- **Missing:** campaigns, content categories, first-comment status, post-comment status (resolved/unresolved), free-text caption search, date-range picker (list mode), automations, CSV batch, "No social account" toggle, per-status counts, saved filters, custom views, "Only" shortcut.
- ⚠️ **Correction:** the doc says mobile "already sends the missing axes as empty values". That is true for `campaigns`, `content_categories`, `comment_status`, `automations`, `social_selection`, `no_social_account` and `csv_id`. **It is not true for `first_comment_statuses` or `search`, which mobile does not send at all.** Mobile also still sends `blog_selection`, which the web has dropped (blog publishing is sunset; remove it).
- **Latent bug:** `PlannerFilter.hasActiveFilters` leaves out `createdByIds`, while `activeFilterCount` includes it. A "Created by" filter alone may not show the "filters active" state.

### 2.3 Post card surface
- **Missing:** colour-coded labels on the card, campaign/content-category badge, content-type badge (Evergreen/RSS/Repeat/Carousel/Image/Video; Blog excluded, sunset), per-account publish sub-status + Live Link, plan issues ("{n} errors"), partial-failure detail, thread/first-comment indicators, locked indicator, hidden-from-clients state.
- **Partial:** the media carousel shows the first item plus a count; hashtags/mentions are tinted but not tappable.
- ⚠️ **Correction:** labels **are** shown on the **preview screen** (`PlannerLabelChips`). They are missing only from the card.
- Channel name / type pill / team / author on the card were removed on purpose (N/A).

### 2.4 Per-post actions
- **Mobile card sheet:** edit, duplicate, approve, reject, delete. **Mobile preview menu:** edit, duplicate, replace, download media, delete. Mark published / not published sits only in the notification-sent footer.
- **Missing:** recycle (to Evergreen), send for approval, change approval, reopen (rejected/failed/missed), hide/unhide from clients, share via link, retry posting (+ live progress), publishing status panel, download/view PDF, resend mobile notification, refresh automation post, shuffle queue, delete from social platforms (per account), cross-workspace delete for global content categories.
- **Partial:** delete is a simple confirm (the web has a wizard with per-network delete); reschedule uses the wrong contract (below).
- ⚠️ **Correction:** mobile "Replace" just opens the composer in edit mode. The web's Replace calls `/replacePlan` to **refresh an automation post**. So mobile's "Replace" is a mislabelled Edit, and "refresh automation post" really is missing.

### 2.5 Reschedule contract: standing data-loss risk (P0)
Mobile drag-reschedule **re-fetches the plan, re-hydrates it into a composer draft, and re-submits the whole plan** to `/processSocialShare` (`contentstudio-flutter/lib/features/planner/data/plan_reschedule_service.dart:40-76`). It is correct today only because hydration is complete. **Any hydration gap silently wipes caption, media, first comment or settings on a drag.** The web calls `reschedulePlan` with only `{ id, execution_time, workspace_id }`, and the backend changes only `execution_time`.

Details for the fix:
- `execution_time` must be sent in **UTC** `YYYY-MM-DD HH:mm:ss` (web `useCalendarView.ts:681-688`). The current mobile path works in workspace-local time. See also memory: posts are stored in UTC, and timezone is a workspace setting.
- The backend does no state validation, so the client must apply the web's guards (`useCalendarView.ts:622-693`): block for the Approver role, posts with a content category, `post_state` other than draft/scheduled/reviewed, published / partially-failed posts, and past times. Today mobile checks only `canPerform.edit` and past time.

### 2.6 Bulk operations
Nothing on mobile: no row selection, select all on page / all matching filter, bulk edit, bulk approve/reject, bulk send-for-approval / change-approval, bulk delete, bulk recycle or bulk share. Long-press today opens the actions sheet; it would have to move to "enter selection mode". The inbox already has a multi-select pattern (`features/inbox/application/inbox_list_state.dart:17-45`, `inbox_bulk_action_bar.dart`) and so does the media library (`media_selection_bar.dart`).

### 2.7 Post preview
- **Missing:** prev/next navigation with prefetch, analytics tab (per-post metrics, per-platform "not supported" messaging, link to analytics), download PDF, realtime updates while open.
- **Partial:** the more-options menu has 5 of the web's ~20 actions; deleted-plan handling.

### 2.8 Comments (closest to parity)
Missing: the All / Resolved / Unresolved filter, and realtime comment sync. The system emoji keyboard is partial (curated picker). Resolve-all / unresolve-all already exist.

### 2.9 Collaboration and sharing
None of it is on mobile: create share link (title, password, notes toggle, 3 sharing modes, external emails, send-for-approval toggle, approval rule), manage shared links, Share Calendar / Share Posts entry points, realtime collaboration.
- ⚠️ **Correction:** web realtime runs on **Centrifugo** (`centrifuge` 5.2.2, `contentstudio-frontend/src/modules/common/services/realtime/RealtimeService.ts`), **not Pusher**. Channels: `plan-collaboration:{workspace_id}` (store_comment, delete_comment, update_comment, resolve_all_comments, store_label, store_status, store_post, store_folder) and `plan-job:{workspace_id}` (retry_posting). The gap doc's channel names are right.
- The share-link product risk (from the gap doc) still stands: links are client-facing; password rules, approval rules and external-email invites have product implications, and **adding an external approver overwrites the internal approver**. This needs product sign-off.

### 2.10 Calendar
- **Missing:** calendar notes (create/edit/delete, private vs public, drag to another date, popover, compose from a note), recurring notes (daily/weekly/monthly/yearly, interval, weekday/day-of-month, until-date or count), holidays (country picker, save to all workspaces, remove, AI content), display toggles (notes, holidays, media thumbnails, compact view, fade published), guard for an unfinished composer draft.
- **Partial:** create from an empty day (two taps, social post only; the web offers "Social Post / Use Template"); "+N more" (day sheet, no drag-out); default calendar view (local only).
- ❌ **Likely misread:** "External calendar events". The web `ExternalCalendarEvent.vue` renders events in the **public share-link calendar**. It is not an external-calendar integration, so it drops out of scope.
- **Missed by the doc:** mobile **hard-codes the week start to Sunday** (`planner_calendar_controller.dart:38`, `firstWeekday = DateTime.sunday`). The web uses the user/workspace first-day setting.
- Notes already arrive in the calendar response: the backend returns `notes` only for `route_name: 'calender_plans'` (`contentstudio-backend/app/Http/Controllers/Planner/PlanController.php:132`), mobile already sends that route, and **mobile's response DTO ignores the key**. Reading notes is DTO + rendering work.

### 2.11 Organisation, export, preferences
Missing: PDF export (white-label aware, permission-gated), post analytics, server-side default view / sort / calendar-view preferences (mobile persists mode to SharedPreferences only: `cs_planner_view_mode`, `cs_planner_week_variant`; it never reads `planner_default_*` from `/me`). Feature gating on mobile covers approval workflows only. Manage columns is N/A (no table).

### 2.12 Mobile-only capabilities to preserve
Download media to the gallery (`gal`, permission-gated); Hourly / Daily week variants; long-press day peek card; five-source caption fallback chain.

**Note on bulk vs peek:** long-press on a *day* opens the peek; long-press on a *post card* opens the actions sheet. Bulk selection takes over the post-card long-press only, so the peek survives.

### 2.13 Summary counts (from the gap doc, after corrections)
| Area | Web | Mobile has | Missing / partial |
|---|---|---|---|
| View modes | 7 | 3 | 3 missing |
| Filtering and sort | 24 | 8 | 13 missing, 1 partial + saved filters, custom views, counts |
| Card surface | 16 | 5 | 9 missing, 2 partial |
| Per-post actions | 23 | 8 | 15 missing, 2 partial |
| Bulk | 8 | 0 | 8 missing |
| Preview | 11 | 6 | 4 missing, 2 partial |
| Comments | 10 | 8 | 2 missing, 1 partial |
| Sharing and realtime | 4 | 0 | 4 missing |
| Calendar extras | 11 → 10 (external events dropped) | 4 | 5 missing, 4 partial + week-start bug |
| Organisation / export | 7 | 1 | 4 missing, 3 partial |

---

## 3. Competitor research (mobile apps specifically)

Vendors document their web planners in depth and their mobile apps thinly. Several help centers (Later, Planable, SocialBee) blocked fetches, so some cells come from App Store listings and search snippets. "Unknown" means no evidence either way.

### 3.1 Competitor table
| Competitor | Mobile planner depth | Key capabilities on mobile | Pricing tier | UX approach | Unique differentiator |
|---|---|---|---|---|---|
| Buffer | Basic | Week + day view, All Channels day view (iOS), create from a date, edit via ⋮, Drafts/Approvals tabs | Free+; approvals on Team | Calendar is thin; work happens in Publish tabs | Approvals as a separate tab |
| Hootsuite | Partial | Planner overview, edit drafts, approve, "Enhanced Post Context" (Mar 2026), mobile AI agent (Jun 2026) | Professional+ | Planner plus push toward Inbox | Richer post detail on mobile |
| Publer | Partial–Full | Day/week/month calendar, create from calendar, status filter chips, edit/delete/duplicate/reuse | Free; Pro/Business | Near-full composer | Feed/grid + holidays on web; not confirmed on mobile |
| Later | Partial | Schedule tab with **List / Preview (IG grid) / Calendar**, **drag to reschedule**, grid rearrange | Starter+ | Visual-first | Grid is a main mobile view; can't create from the mobile calendar |
| Sprout Social | Partial | View/edit/add in calendar, approvals as Tasks, push for approvals and internal comments | Standard+ (approvals Pro+) | Notification/task driven | Approvals arrive as tasks |
| Loomly | Partial–Full | All calendars, preview, change status, assign, comment, duplicate to other calendars, push on every action | Base+ | Workflow-state-centric | Status change as a one-tap action |
| Sendible | Basic–Partial | Weekly calendar, approve with feedback, failed-post alerts | Creator+ | Split companion apps | Failed-post push + approve with note |
| SocialBee | Partial | "Next Posts": list, calendar, IG grid; approval tab by profile/category; comments with @mentions | Bootstrap+ | Category-centric | Filter approvals by category |
| Agorapulse | Partial | Status list, **filters: profile, status, content type, date, labels, timezone**, edit, **group delete**, approve/reject, send for approval (one assignee) | Standard+ | List-first, colour-coded status | Richest mobile filter set |
| Metricool | Partial | Plan/edit/duplicate, send to review by email, IG feed preview, **download reports** | Free+ | Companion; publishes a mobile-vs-web feature table | Honest "web only" docs |
| Planable | Full (for review) | **Calendar, Grid and Feed**, drag in grid, approve, real-time comments, bulk schedule | Basic/Pro | Approval-first | Feed looks like the real network |

### 3.2 Capability matrix (mobile apps)
| Area | Buffer | Hootsuite | Publer | Later | Sprout | Loomly | Sendible | SocialBee | Agorapulse | Metricool | Planable |
|---|---|---|---|---|---|---|---|---|---|---|---|
| View modes | Partial | Partial | Partial | **Yes** | Partial | Partial | Partial | **Yes** | Partial | Partial | **Yes** |
| Filters | Partial | Unknown | Partial | Unknown | Unknown | Unknown | Unknown | Partial | **Yes** | Unknown | Unknown |
| Bulk ops | No | Unknown | No evidence | No evidence | Unknown | No evidence | Unknown | Unknown | Partial | Unknown | Partial |
| Per-post actions | Partial | Partial | Partial | Partial | Partial | Partial | Partial | Partial | Partial | Partial | Partial |
| Client share links | No | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Partial | Unknown |
| Notes/holidays | No | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown |
| Per-post analytics | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | No evidence | Partial | No evidence |
| Realtime/collab | Partial | Partial | Partial | Unknown | **Yes** | **Yes** | Partial | Partial | Partial | Partial | **Yes** |
| PDF/reports | No | Unknown | No | Unknown | Unknown | Unknown | Partial | Unknown | No evidence | **Yes** | No evidence |

### 3.3 Common patterns
- **The market splits mobile vs desktop:** mobile is for **triage, review, approve, quick fixes**; desktop is for planning, bulk work and reporting. Metricool says so outright.
- A status-grouped **list** is often the main mobile view; calendars are shallow (week/day more than month).
- **Approvals, usually reached from a push notification, are the most consistently supported mobile workflow.**
- Failed-post push alerts are common; a retry path is expected.
- IG grid preview is common among visual-first tools (Later, SocialBee, Planable, Metricool).
- **Left to desktop almost everywhere:** bulk ops (only Agorapulse group delete and Planable bulk schedule), notes/holidays (no evidence on mobile anywhere), PDF/report building (Metricool only), per-post analytics in the planner (no one), client share links created on mobile (no one clearly), multi-step approval routing setup.

### 3.4 Differentiators worth borrowing
1. **Planable's Feed view.** Posts look like the network, with one-tap approve and real-time comments. Best-in-class mobile client review.
2. **Agorapulse's filter depth as a bottom sheet.** Proves the full filter set works on a phone.
3. **Later's List / Preview / Calendar switcher** with drag in the calendar.
4. **Metricool's published mobile-vs-web table.** Worth copying for a help article so deliberate web-only gaps aren't reported as bugs.
5. **Creating a client share link from mobile.** Nobody documents it. If ContentStudio ships it, it's a real differentiator (the link already exists on web; mobile needs the create sheet + native share sheet).

### 3.5 User expectations
| Table stakes | Delighters |
|---|---|
| List grouped by day + week/month calendar | Feed preview that looks like the network |
| Filter by channel + status | IG/TikTok grid preview |
| Approve/reject with comment, from a push | Labels / campaigns / categories filters, saved views |
| Edit, delete, duplicate | Drag to reschedule |
| Failed-post alert with retry | Bulk approve / delete |
| Internal comments + @mentions | Share a review link to a client from the phone |
| Data refreshes without manual pull | Per-post performance for published posts |

### 3.6 Sources
Buffer ([calendar on mobile](https://support.buffer.com/article/625-using-the-buffer-calendar-feature-on-mobile), [approvals on mobile](https://support.buffer.com/article/633-creating-managing-and-approving-draft-posts-on-the-mobile-app)); Hootsuite ([App Store notes](https://apps.apple.com/us/app/hootsuite-social-media-tools/id341249709)); Publer ([mobile help](https://publer.com/help/en/article/the-publer-mobile-app-17a4cm3/)); Later ([mobile guide](https://help.later.com/hc/en-us/articles/360043244813-A-Guide-to-the-Later-Mobile-App), [visual planner on mobile](https://help.later.com/hc/en-us/articles/360042742094-Visual-Instagram-Planner-on-Mobile)); Sprout ([mobile apps](https://sproutsocial.com/features/mobile-apps/), [publishing notifications](https://support.sproutsocial.com/hc/en-us/articles/360016087072-Publishing-notifications-on-mobile)); Loomly ([mobile apps](https://loomly.zendesk.com/hc/en-us/articles/38818394484635-How-to-use-Loomly-s-iOS-and-Android-mobile-apps)); Sendible ([mobile](https://www.sendible.com/features/social-media-mobile)); SocialBee ([mobile app](https://help.socialbee.com/hc/en-us/articles/29979248854935-Using-the-SocialBee-Mobile-App)); Agorapulse ([manage posts on mobile](https://support.agorapulse.com/en/articles/13455632-how-to-view-and-manage-posts-from-the-mobile-app)); Metricool ([features in mobile app](https://help.metricool.com/en/article/features-available-in-the-mobile-app-ut3sji/)); Planable ([mobile help](https://help.planable.io/hc/en-us/articles/21715225504796-Mobile-App)).

---

## 4. Codebase analysis

### 4.1 Flutter architecture (planner)
- Riverpod 2 without codegen (`Notifier` / `AsyncNotifier` / `FutureProvider.family`), go_router, one Dio client. Endpoints in `ApiPaths` (`contentstudio-flutter/lib/core/env/app_config.dart:225+`). Results as `ApiResult` (`ApiSuccess` / `ApiFailure` / `ApiNetworkError`).
- Module: `contentstudio-flutter/lib/features/planner/{application,data,domain,presentation}`.
- Providers: `application/planner_providers.dart` (`plannerListProvider`, `plannerCalendarProvider`, `plannerFilterProvider`, `plannerPostDetailProvider`, `plannerLabelsProvider`, `planRescheduleServiceProvider`).
- Controllers: `planner_list_controller.dart` (refresh, silentReload, loadNextPage, deletePost, approve/reject); `planner_calendar_controller.dart` (optimistic drag with generation token, rollback at `:169-240`).
- Routes: `/planner`, `/planner/post/:id`, top-level `/post/:id` for deep links (`lib/app/router.dart:423-451`).
- Refresh on app resume already exists (`presentation/planner_screen.dart:83`).
- `contentstudio-flutter/docs/features/planner.md` is **stale**: "Remaining Work" (`:942-962`) still calls preview a stub and the calendar unbuilt. Update it as part of this epic.

### 4.2 Reusable Flutter pieces
| Need | Reuse |
|---|---|
| Multi-select | Inbox `inbox_list_state.dart:17-45` (`selectedIds`, `isSelecting`, `isBulkRunning`) + `inbox_bulk_action_bar.dart`; media library `media_selection_bar.dart` (select-all) |
| Sheets / dialogs / search | `shared/widgets/cs_bottom_sheet.dart` (`CSSheetActionTile`), `cs_dialog.dart` (`showCSConfirmDialog`), `cs_search_field.dart`, `cs_filter_pill.dart` |
| Campaign / category lookups | `features/composer/data/composer_service.dart:434,508` (`/getPublicationFolders`, `/categories/show`) |
| Members | `plannerWorkspaceMembersProvider` |
| Labels | `PlannerLabelChips`, `LabelResolver` (preview screen) |
| PDF viewing | `pdfx` + `shared/widgets/media/document_preview.dart` |
| File saving | `path_provider`, `permission_handler`; `gal` covers images/videos only, **not PDFs** |
| Links | `url_launcher`, `core/launcher/open_external_url.dart`, `open_in_app_web.dart` |
| Share sheet | **`share_plus` not installed.** New dependency, or clipboard only |
| Charts | `fl_chart` (used in AI Studio) |
| Deep links | `app_links` + `application/approval_deep_link_coordinator.dart` |
| Entitlements | `entitlements/domain/feature_access.dart:9-35` (`AppFeature`), `featureEnabledProvider`; **fail-open** (`:45-50`): a missing flag shows the feature |
| Realtime | **None.** Only an AI SSE decoder exists (`features/ai_assistant/data/ai_sse_decoder.dart`), which can't be used for Centrifugo. The inbox doc marks sockets out of scope |
| i18n | All new strings go into all 8 `assets/i18n/*.json` (parity test enforces it) |

### 4.3 Integration points
- **Filters:** extend `PlannerFilter` (fields, `requestSignature`, `copyWith`, counts; fix `hasActiveFilters`). Fill the empty keys in `PlannerService._buildPayload` (`data/planner_service.dart:651-704`); add `first_comment_statuses` and `search`; remove `blog_selection`. New sections in `presentation/widgets/planner_more_filters_sheet.dart`. List-mode date range goes into `date_range`.
- **Status counts:** new `fetchPlansCount` method (payload minus page/limit/sort_column/order/social_selection, as web `stripCountsExclusions` does). Show counts in `planner_status_filter_sheet.dart`.
- **Card:** extend `PlannerPostDto` / `PlannerPost` (`domain/planner_post.dart:19-42`) with campaign, category, labels, posting sub-status, issues, first comment, thread, `is_locked`, hide-client. Render in `planner_post_card.dart` `_Header` / `_ContentRow`.
- **Actions:** new callbacks in `PlannerCardActions` (`widgets/planner_card_actions_sheet.dart:13-19`), new `PreviewMoreAction` values (`domain/preview_more_action.dart:6`) + `PreviewActionsResolver`, new `PlannerService` methods. Gate each on `can_perform` + `workspaceAccessProvider`.
- **Reschedule:** replace `PlanRescheduleService` with a single `/reschedulePlan` call; keep the optimistic update and rollback.
- **Bulk:** `selectedIds` on `PlannerListState` (inbox pattern); bulk bar in `planner_screen.dart`; `onLongPress` (`planner_screen.dart:309`) switches to entering selection.
- **Notes / holidays:** parse `notes` in `data/dtos/planner_list_response_dto.dart` for calendar responses → `PlannerCalendarState` → render in `planner_month_grid.dart` / `planner_week_*`. Holidays via a separate fetch.
- **Views:** extend `PlannerMode` (`domain/planner_mode.dart:3`), `_fromString` in `application/planner_mode_controller.dart`, and the toggle in `planner_screen.dart`.
- **Grid gate:** add `gridViewPlanning('grid_view_planning')` to `AppFeature`.
- **Week start:** replace `firstWeekday = DateTime.sunday` (`planner_calendar_controller.dart:38`) with the user/workspace setting.
- **Preferences:** parse `planner_default_*` from `/me` (backend `UserData.php:68`, `Account.php:151`).
- **List route name:** mobile list sends `route_name: 'social-modal'` (`planner_service.dart:66`); web sends `list_plans`. Both are V2 paths (`PlanController.php:95`), but per-view response shaping (`getPlannerSpecificResponse`) may differ. Align with web.

### 4.4 Endpoints mobile will need (all exist; POST with `workspace_id` unless noted)
Backend routes: `contentstudio-backend/routes/web/planner.php` (PlanController `:37-60`, `retrySocialShare :76`, `planner/generatePostPDF :78`, `sendPostNotification :83`, PlannerFiltersController `:122-126`, shareLink `:130-137` + public `:22-26`, notes `:147-152`), `routes/api.php:99-106` (saved views) and `:118-125` (countries/holidays), `routes/web/settings.php:71-74` (preferences), `routes/web/analytics.php:290-295` (planner analytics), `routes/web/automation.php:50` (recycle), `routes/web.php:161` (`/realtime/auth`).

| Capability | Endpoint | Payload notes |
|---|---|---|
| Reschedule | `reschedulePlan` | `{id, execution_time (UTC), workspace_id}`; client-side guards required |
| Status counts | `fetchPlansCount` | fetchPlans payload minus paging/sort/social_selection → `plans_count{status:n}` |
| Saved filters | `createPlannerFilter`, `fetchPlannerFilter`, `deletePlannerFilters`, `setDefaultPlannerFilters` | `{name, members, labels, statuses, campaigns, content_category, created_by, platformSelection, no_social_account, approval_*, default, id?}` |
| Custom views | `api/planner/saved-views` (GET/POST/GET id/DELETE), `/toggle-default-view`, `/reorder` | web `api/custom-views.ts` |
| Notes | `notes/save`, `notes/edit`, `notes/delete` (reads via calendar `fetchPlans` `notes` key) | `{note_title, start_date, end_date?, note_color, is_private, description?, is_recurring?, recurring_frequency (Daily/Weekly/Monthly/Yearly), interval_value, weekdays[], end_condition (until/count), recurring_until_date / recurring_count}` |
| Holidays | GET `api/holidays`, GET `api/countries`, GET `api/workspace/selected-countries`, POST `api/workspace/countries`, POST `api/holiday/remove` | `api/holiday/contentWithAI` is AI generation → **web-only**, out of scope |
| Share links | `shareLink/fetch|create|update|delete` | `{name, password?, is_password_protected, show_notes, share_future_content, mode, plans[], calendar_date, view_type, is_single_post, filters, approval_flow, approval_emails, approval_option, …}`; public URL `{webBase}share/planner/{link}` |
| Bulk | `processPlanBulkOperation` (`{operation: scheduled/rejected/deleted, selected_plans[]}` or filter payload for check-all), `processPlannerBulkEdit`, `sendPlansForBulkApprovers` | |
| Retry posting | `retrySocialShare` | `{posting_id}`; progress via `plan-job` `retry_posting` |
| Publishing status panel | GET `fetchPlanPosting?id=` | per-account postings |
| Delete from networks | `removePlanPosting` | `{id, posting_ids:[{id}]}` |
| Cross-workspace category delete | `removePlan` | `remove_from_all_workspaces: true`; mobile posts form-encoded today |
| Refresh automation post | `replacePlan` | `{id, route_name, type, automation_id}` |
| Recycle | `prepareRecyclePosts` | `{plan_ids}` → opens composer |
| Hide from clients | `hideClientAction` | `{id, status:bool}` |
| Resend mobile notification | `sendPostNotification` | `{platform_identifier, planner_id}` |
| Shuffle queue | `shuffleQueuePosts` | `{account_id}` |
| PDF | `planner/generatePostPDF` | `{post_id, …}` → blob; data from GET `plan/preview?type=pdf` |
| Post analytics | analytics-Go base + `analytics/campaignLabelAnalytics/getPlannerAnalytics` | `{id, all_post_ids[], platforms}`; **mobile has no analytics-Go base URL** (Laravel mirror at `routes/web/analytics.php:295`) |
| Preferences | `preferences/setPlannerDefaultView`, `setPlannerDefaultSort`, `setPlannerDefaultCalendarView` | read via `planner_default_*` on `/me` |
| Realtime auth | POST `realtime/auth` | `{}` → connection token; `{channel_name}` → subscription token. Uses `auth()->user()` on the `api` guard, so a mobile bearer token *should* work. **Needs a live test** |

### 4.5 Technical considerations
- **Realtime** needs a new Dart `centrifuge` dependency, a connection lifecycle tied to app foreground (disconnect in background, reconnect on resume), teardown on workspace switch, and event handlers that patch `PlannerListState` / `PlannerCalendarState` / the preview by `plan_id` (no-op if the plan isn't loaded, as on web). It changes cache-invalidation assumptions across list, calendar and preview at once. If deferred, the fallback is the existing resume-refresh + pull-to-refresh.
- **Bulk** is the single largest structural change: a selection model across list, calendar day sheet and controllers.
- **Share links** are client-facing artefacts (see 2.9); **product sign-off needed before build**.
- **Grid views** are billing-gated and need remote platform media (IG/TikTok published posts) that mobile has no client for today.
- **PDF** is a blob download: save with `path_provider`, preview with `pdfx`, hand off via the share sheet. Android scoped storage applies.
- **Post analytics** needs a new base URL (analytics-Go) or the Laravel mirror.
- **Timezone:** reschedule `execution_time` must be UTC; queue slots are wall-clock. Timezone is a workspace setting, never per user.

### 4.6 Platform-specific concerns
- **Share sheet:** add `share_plus` (UIActivityViewController on iOS, ACTION_SEND on Android), or copy to clipboard only. On iPad the share popover needs an anchor rect.
- **Share-link URLs** point to the web `share/planner/:id`. Decide whether app links should intercept them (probably not: the page is for external clients).
- **Touch ergonomics:** long-press on a post card switches to "enter selection"; the actions sheet moves to the card's ⋯ button. The day peek (long-press on a day) is unchanged.
- **Push → retry:** failed-post pushes already deep-link to a post; the retry action should be reachable from that screen.
- **Localization:** 8 locale files per new string.

---

## 5. Recommended approach (input to Step 2)

Combining the gap list with the market split (mobile = triage/review/fix, desktop = plan/bulk/report), a phased epic:

**Phase 1 (P0): safety + table stakes**
1. Switch reschedule to `reschedulePlan` (UTC, web guards). Removes the data-loss risk.
2. Card surface: labels, campaign/category, content-type badges, per-account status + live link, issues / partial failure, first-comment / thread / locked / hidden indicators, carousel paging, tappable hashtags.
3. Filters: search, campaigns, content categories, comment status, first-comment status, date range (list), label search, "No social account", per-status counts; fix `hasActiveFilters`; drop `blog_selection`.
4. Per-post actions on the card sheet + preview menu: send for approval, change approval, reopen, retry posting, publishing status panel, delete from social platforms (delete wizard), hide from clients, refresh automation post (rename mobile "Replace" to Edit), recycle, resend mobile notification.
5. Calendar week start from the user setting (bug).

**Phase 2 (P1): collaboration + organisation**
6. Bulk select on list (approve, reject, delete, send for approval; bulk edit of labels/campaign as a stretch).
7. Calendar notes (view + create/edit/delete, recurring).
8. Holidays (view + country picker; AI holiday content stays web-only).
9. Share via link (create + manage; native share sheet). Needs product sign-off on approval-flow implications.
10. Saved filters (apply + save + set default); server-side default view / sort / calendar preferences.
11. Preview: prev/next navigation, comments filter.
12. Realtime (Centrifugo) for comments, status and retry progress. Could be P2 if the auth test fails.

**Phase 3 (P2): visual + reporting**
13. Instagram / TikTok grid view (gated `grid_view_planning`).
14. Feed view, compact list.
15. Post analytics tab in preview.
16. PDF download / view.
17. Custom views (web-managed; mobile applies them); calendar display toggles.

**Keep as-is:** download media, Hourly/Daily week variants, day peek, caption fallback chain.
**Out of scope:** manage columns (no table), holiday AI content (AI generation is web-only), external calendar events (misread), Blog content type (sunset).
**Documentation:** a help article listing what's on mobile vs web, Metricool-style.

### Story-type expectations
- Almost entirely `[Flutter]` stories, plus **one `[Design]` story** covering the new mobile surfaces (card badges, bulk bar, notes/holiday rendering, share-link sheet, grid/feed).
- Possible `[BE]` items: (a) confirm/fix `realtime/auth` for mobile bearer tokens; (b) expose planner post analytics on a base mobile already uses, if the Laravel mirror isn't enough.
- No developer-surfaces work: no new or changed API.

---

## 6. PO scope decision (2026-09-24)

The PO reviewed the gap list and set the scope directly; the workflow and PRD steps were skipped at their request. Stories stay short and user-POV, with no Mermaid diagrams.

- **In:** Centrifugo realtime (planner, preview, comments; plus a BE check on mobile token auth); Instagram + TikTok grid; filters (campaigns, content categories, first-comment status, comment status, caption search, date range, automations, "No social account", label search, account grouping, "Only" shortcut); per-status counts; saved filters; custom views; all card elements except the Blog badge; per-post actions (safe reschedule, send for approval, change approval, reopen, hide/unhide, share via link, retry + publishing status, download PDF, delete from social platforms, global-category cross-workspace delete, recycle); bulk approve / reject / send for approval / share; preview prev/next + full menu + PDF + realtime; comment filter + realtime; share links create + manage (full, product sign-off noted); calendar notes (incl. recurring), holidays, display toggles, create-from-day menu, "+N more" drag-out, draft guard, week-start fix; server-side default view / sort / calendar view.
- **Out:** Compact List and Feed views, CSV filter, bulk edit / delete / recycle, resend mobile notification, shuffle queue, refresh automation post, post analytics, planner-wide PDF export (per-post PDF stays in), holiday AI content, Blog content type.
- **Not addressed:** mobile's mislabelled "Replace" action (it opens Edit). Left as-is; raise separately if wanted.
- **New Usermaven events proposed:** `planner_filter_saved`, `planner_custom_view_saved`, `post_sent_for_approval`, `post_retried`, `share_link_created`, `planner_note_created`. **Reused:** `post_rescheduled`, `post_approved`, `post_rejected`, `planner_view_changed` (from the Flutter analytics plan). The web planner fires none of these.
