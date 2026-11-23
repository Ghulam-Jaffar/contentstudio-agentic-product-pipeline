# CONT-4933 — Feed view on the client share-link page: code verification

Checked on 2026-11-23 against `contentstudio-frontend` `origin/develop` (fetched 2026-11-23 00:16), plus `origin/features` for the Approval board. Paths are relative to `contentstudio-frontend/`.

## Share page today
- Route `/share/planner/:id` → `src/modules/planner_v2/views/SharePlans.vue` (`src/router.ts:309`).
- Views picked by `state.currentView` in `SharePlans.vue:55-144`: `SharePlansCalendarView`, `SharePlansListView` (desktop), inline `DataRowCardMobile` stack (phone).
- Switcher: `SegmentedControl` in `components/share-plans/SharePlansControls.vue:115-131`, options hard-coded at 180-183, icon ternary at 124. Sort menu only for List (line 72).
- `switchView()` in `composables/useSharePlansData.ts:362-390`. Saved view read at 85-92 from the single localStorage key `share_link_view` (not per link), accepts only `list`/`calendar`.
- One infinite query, `composables/useSharePlansQuery.ts` (limit 10 for List, 500 for Calendar, line 67).
- Scroll paging finds `.f-table` / `#mobile-card-view` by selector (`useSharePlansData.ts:400-415`). No end-of-list or load-more error message.
- Filters are client-side over loaded pages only, matched to `localStorage.external_email` (`useSharePlansData.ts:239-277`).
- Phone breakpoint: `window.innerWidth < 900` (`useSharePlansData.ts:233-237`, `675`).

## Link settings
- Password: `useSharePlansData.ts:452-457`, `PublicPasswordProtectedModal.vue`, `onVerify` 474-491.
- Comments / approve-reject flags: read at `useSharePlansData.ts:434-436`; used in `DataRow.vue:179-197`, `DataRowCardMobile.vue:132-160`, `PlanPreviewPanel.vue:341-351`, `SharePlansBulkActionsBar.vue`, `ExternalActionsModal.vue:246-262`.
- Name/email step: `ExternalActionsModal.vue:298-305`, `ak` added at 322.
- External emails + Anyone/Everyone: set in `SharePlanModal.vue:75-136` / `useShareLinkForm.ts:590-595`; share page only sees `?ak=` → `approverEmail` (`useSharePlansData.ts:420-449`); hides actions after approval completes (`DataRow.vue:320-328`, `DataRowCardMobile.vue:592-600`, `PlanPreviewPanel.vue:1035-1046`). Rule enforced in the backend.
- Share future updates: only the banner (`SharePlans.vue:33-53`); posts decided by backend `shareLink/fetchPlans`.

## Planner Feed card
- `src/modules/planner_v2/components/PlannerFeedCard.vue`, used by `views/FeedView.vue`.
- Left rail `:186-243` (Edit/Duplicate/Delete gated by `can_perform`; "…" Download PDF always shown). Right panel `:522-651` (team approve/reject + `CommentsAndNotes`). No `group-hover` actions.
- Layout CSS only under `.planner_component .planner_feed_view` (`src/assets/styles/legacy/modules/composer/planner/_planner.css:1958-2268`): `.feed_center` fixed `width/min-width: 500px`, `.feed_left` 150-180px; right panel `min-w-100 w-120`.
- No comment count on the card.
- Feed icon is the `FeedViewIcon` component (`composables/usePlannerViewSwitcher.ts:95`), not an `@contentstudio/ui` icon name.
- Previews resolve accounts through `useChannel().getFreshAccount`, which falls back to the plan's own account when the workspace store is empty (public page).

## Approval board (CONT-4270)
- Commit `514d41270c` is on `origin/features` (and `origin/feature/approval-board`), not on `develop` or `master`.
- On `features`: switcher gets `board`; localStorage check accepts `board`; bulk bar shows when `currentView !== 'calendar'` (`SharePlans.vue:166`), so Feed would get the bar automatically but cards need a select control.
- Existing copy to reuse: `share_plans.no_posts_found`, `no_posts_found_filtered`, `loading_more`, `board.select`, `board.error`.

## Backend
None needed. Feed reuses `shareLink/fetchPlans`, `shareLink/action`, `shareLink/comment` through `ExternalActionsModal` and `PlanPreviewPanel`.
