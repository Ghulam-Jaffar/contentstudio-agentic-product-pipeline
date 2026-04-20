# Research — Mobile Planner Calendar View (iOS + Android)

## Feature summary

Add a calendar view to the Planner on iOS and Android. Today mobile only has list view. Calendar view has two sub-views (Month, Week). Each day cell shows a post-count badge if populated; empty cells offer "+ create"; tapping a populated cell opens a bottom sheet with View posts / Create post. Full list-view actions available in the posts-of-day sheet. Per-user view preference (List vs Calendar) persists locally on the device only — does not mix with web's saved view.

## iOS codebase (contentstudio-ios-v2)

### Architecture
- **Hybrid UIKit + SwiftUI**, MVVM-ish.
- **Root controller:** `PlannerBaseViewController.swift` — manages filter options and compose button
- **List view:** `ScheduledTableViewController.swift` — grouped UITableView, posts grouped into date-headed sections
- **Data model:** `PlannerResponse.swift` — holds all post properties (post_state, multimedia, canPerform, approval, etc.)

### Reusable components
- **Post card:** `PlannerPostCardView.swift` (SwiftUI, ~400 lines) — already reusable; takes `PlannerResponse` and delegate
- **Action menu:** `MorePostOptionsView.swift` — Preview / Edit / Duplicate / Download media / Delete, plus Approve / Reject for approval posts, gated by `canPerform` flags
- **Post preview:** `PostPreviewViewController.swift` — nav push, hides nav bar + tab bar, returns via `NotificationCenter` `WillReturnToPlannerScreen`
- **Bottom sheet:** custom `BottomSheetViewController.swift` + `BottomSheet` library import via `Constant.bottomSheetConfiguration`
- **Composer entry:** single helper `Router.MoveToComposer(from:editMode:postId:isDuplicate:)` — works for create and edit
- **Prefs:** `UserDefaults`; filters live in `Constant.sharedSelectedStatusIds` (in-memory, not persisted)

### Calendar library
- **None integrated today.** Pick during implementation — FSCalendar (UIKit) or a custom SwiftUI grid are both viable given the hybrid architecture.

### Backend API
- Planner list hits `POST /fetchPlans` with pagination, filters, empty `date_range`. Calendar can reuse this with a filled `date_range`, or use the existing web-calendar endpoint `POST /content-calendar`.

## Android codebase (contentstudio-android-v2)

### Architecture
- **Java + XML Views + Fragments**, MVP pattern (no MVVM).
- **Root activity:** `PlannerActivity` — TabLayout + ViewPager, 5 tabs
- **List fragment:** `PostsFragment` — hosts RecyclerView, handles API calls, filtering, pagination, actions
- **Base:** `BasePostsFragment` — abstract common setup
- **Adapter:** `PlannerListViewAdapter` + `PlannerItemHolder` (inner class) — tightly coupled; not yet reusable
- **Data model:** `PlannerListItem` Java POJO

### Reusable components
- **Post card:** `item_planner.xml` — card layout. **Tightly coupled to the adapter** — needs extraction to be reusable in the day-posts bottom sheet. This is a prep story.
- **Post actions in list:** Approve, Reject, Replace (currently hidden), Edit (Intent → `EditComposerActivity`), Duplicate, Delete (broadcast → `removePlanServiceCall`)
- **Post preview:** **does not exist on Android.** (iOS-only feature.)
- **Bottom sheet:** `BottomSheetDialogFragment` pattern in place (e.g. `LanguageBottomSheet`) with a reusable `BottomSheetDialogTheme`
- **Composer entry:** `PlannerActivity.showComposer()` → `ComposerActivity`. Edit via `EditComposerActivity` with `planId` extra.
- **Prefs:** `PrefManager` + `SharedPreferences` (file: `"content_pref"`, `commit()` style)
- **Filters:** static fields on `Constants` (e.g. `filterTypeId`, `plannerSocialList`), broadcast-intent pattern on "Done"

### Calendar library
- **Already in `build.gradle`:** `com.github.prolificinteractive:material-calendarview:2.0.0` — imported but unused today. Use this for the month view. Existing calendar styles (`CustomCalendarDayViewStyle`, `CustomCalendarMonth`, `CustomCalendarViewStyle`) can be reused / extended.

### Backend API
- Same `POST /fetchPlans` (Retrofit interface in `ServiceManager.java`). `date_range` field exists but is sent empty today.

## Backend (contentstudio-backend)

### Planner list endpoint
- `POST /fetchPlans` (routes/web/planner.php) → `PlanController::fetchPlans()`
- Accepts: `type`, `members`, `labels`, `statuses`, `date_range` (`"YYYY-MM-DD - YYYY-MM-DD"`), `platformSelection`, `blog_selection`, `campaigns`, `created_by_members`, `approval_*`, `page`, `limit`
- Returns paginated posts. `execution_time.date` is UTC; backend converts workspace-timezone dates to UTC before querying.

### Calendar endpoint
- `POST /content-calendar` (routes/api.php) → `ContentCalenderController::getContentCalendarData()`
- Returns `plans`, `notes`, `holidays`, `workspace` for the given date range.
- **Filters supported: `statuses` only.** Members, labels, campaigns, approval filters are **not** accepted.
- Returns full plan objects (no per-day count aggregation). Mobile must group by `start` date (already formatted as `Y-m-d` in workspace timezone) and count.

### View preference persistence (web)
- `PlannerSavedViews` collection + `PlannerSavedViewsRepository` + `PlannerSavedViewsController`.
- Fields: `view_type` (`List | Calendar | Grid | Feed | ...`), `visibility`, `types`, `statuses`, `members`, `labels`, `calendar_settings`, `default`, etc.
- **No `client_type` field** — web and mobile would collide today.

### Mobile strategy (per PO decision)
- **No backend changes** for this feature. Local persistence only (UserDefaults on iOS, PrefManager on Android). If mobile engineers hit API gaps during implementation (e.g. calendar endpoint missing filters), raise it with the PO as a follow-up `[BE]` story — not a blocker for this epic.

## Known gaps (noted, intentionally deferred)

1. `POST /content-calendar` does not accept all planner filters (members, labels, campaigns, approval). If the PO wants full filter parity in calendar view, a backend story is needed later. Mobile stories assume filters apply end-to-end — if backend doesn't honor a filter, the mobile dev raises it.
2. `PlannerSavedViews` has no `client_type` separator. Web and mobile use different storage mechanisms in this epic (mobile = device-local only), so no collision — but if mobile ever wants to sync the preference to the server, a backend story is needed.

## Timezone handling

Backend converts workspace-local `date_range` to UTC before querying. Mobile passes `date_range` in workspace timezone. Response `start` field is already workspace-local `Y-m-d`, suitable for direct cell assignment.

## Sub-view (Month / Week) persistence

Not persisted at the user level on either platform (matches web). Sub-view choice is in-memory only. On next app launch, calendar opens in Month view by default.
