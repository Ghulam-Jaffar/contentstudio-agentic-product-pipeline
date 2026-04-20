# Epics & Stories — Mobile Planner Calendar View

Two parallel epics, one per mobile platform. Backend work deferred — mobile uses existing APIs; gaps logged in `01-research.md` for later if needed.

## iOS Epic

**Name:** [iOS] Planner Calendar View
**Shortcut:** [#116788](https://app.shortcut.com/contentstudio-team/epic/116788)
**Objective:** 2026 - Q2
**Timeline:** 2026-04-20 → 2026-05-01
**State:** To Do

### iOS stories

All iOS stories: product area = iOS Mobile, skill set = Frontend, priority = high, project = Mobile, iteration = `20 April - 01 May - 2026`.

1. **[iOS] Add planner view toggle between List and Calendar with local persistence**
   - View-toggle button at top of planner; bottom sheet with List / Calendar options; selection persists in UserDefaults per device; does not sync with web.
2. **[iOS] Build calendar month view with post-count cells, empty-cell quick-create, and floating Composer FAB**
   - New `CalendarPlannerViewController`, month grid, post-count badges, empty-cell two-tap → Composer, floating Composer FAB, filters reused from `Constant.sharedSelectedStatusIds`.
3. **[iOS] Add calendar Month / Week sub-view toggle and build the Week view**
   - Sub-view toggle bottom sheet; Week view renders week containing focused date; sub-view not persisted (resets to Month on launch).
4. **[iOS] Build day-tap bottom sheet flow with posts-of-day list and full list actions**
   - Day-tap → bottom sheet with View posts / Create post → posts-of-day bottom sheet with `PlannerPostCardView` and full `MorePostOptionsView` actions (Preview, Edit, Duplicate, Download, Delete, Approve, Reject).
5. **[iOS] Open post preview from day posts sheet and restore sheet state on return**
   - Preview opens via existing `PostPreviewViewController` (nav push); returning to calendar restores the posts-of-day sheet at its previous scroll position with any post-action updates reflected.

## Android Epic

**Name:** [Android] Planner Calendar View
**Shortcut:** [#116819](https://app.shortcut.com/contentstudio-team/epic/116819)
**Objective:** 2026 - Q2
**Timeline:** 2026-04-20 → 2026-05-01
**State:** To Do

### Android stories

All Android stories: product area = Android Mobile, skill set = Frontend, priority = high, project = Mobile, iteration = `20 April - 01 May - 2026`.

1. **[Android] Extract post card from PlannerListViewAdapter into a reusable component for calendar day sheet**
   - Pure refactor. Card and bindings extracted out of `PlannerListViewAdapter` / `item_planner.xml` into a reusable holder. List view unchanged.
2. **[Android] Add planner view toggle between List and Calendar with local persistence**
   - View-toggle button; `BottomSheetDialogFragment` with List / Calendar; persisted via `PrefManager`; device-local only.
3. **[Android] Build calendar month view using Material CalendarView with post-count cells, empty-cell quick-create, and floating Composer FAB**
   - New `CalendarPlannerFragment` using existing `material-calendarview` dependency; custom decorators for count badges; empty-cell two-tap → `ComposerActivity`; floating FAB with same permission gating as list view.
4. **[Android] Add calendar Month / Week sub-view toggle and build the Week view**
   - Sub-view toggle `BottomSheetDialogFragment`; Week grid (custom if `material-calendarview` doesn't cleanly support it); not persisted.
5. **[Android] Build day-tap bottom sheet flow with posts-of-day list and full list actions**
   - Day-tap → bottom sheet with View posts / Create post → posts-of-day bottom sheet using the extracted card + full action set (Edit, Duplicate, Delete, Approve, Reject, Replace where applicable), hooked into existing broadcast-intent pattern. **No post preview step — Android does not have one today.**

## Mockups

PO attaches Figma / UI screenshots to each story directly. The attached mockups are the single source of truth for exact copy, spacing, colors, iconography, and micro-interactions. Engineers confirm with PO on anything not in the mockups before implementing.

## Key cross-platform guarantees

- View-mode preference (List / Calendar) is device-local only on both platforms — never sent to backend — so the web app's default view is not affected.
- All existing planner filters continue to apply in calendar view; no filter-specific new code.
- Sub-view (Month / Week) is not persisted per user, matches web behavior.
- Post-count on cells is a badge only — no mini previews. No mixed behavior between platforms.
- Post preview navigation is iOS-only (the Android app has no preview screen today; introducing one is out of scope for this epic).

## Deferred (flagged, not blocking)

- Extend `POST /content-calendar` to accept all planner filters (members, labels, campaigns, approval) — mobile currently assumes all filters work; if the backend silently drops some in the calendar endpoint, engineers raise a follow-up `[BE]` story.
- Add `client_type` separator to `PlannerSavedViews` — only needed if we later want to sync mobile view preference to the backend (not in scope for this epic).
