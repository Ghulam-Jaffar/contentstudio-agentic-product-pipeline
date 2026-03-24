# Push Notification Status — Mobile Planner Research

## Context

Existing stories (sc-109139, sc-112172, sc-109140, sc-112183) cover the **push notification action screen** — decline and confirm flows triggered from push notifications. The **[FE] story (sc-109141)** covers the web planner side: status filters, badges, and manual resolution buttons.

This research covers the **missing mobile planner** side: displaying notification statuses in the iOS/Android planner, filtering by them, and adding manual resolution buttons in post preview — mirroring the web [FE] story.

---

## iOS — Current State

### Status System
- **File:** `contentstudio-ios-v2/ContentStudio/HelperClasses/Constant.swift` (lines 337-420)
- 9 statuses defined in `plannerStateData` array (indices 0-8): All, Scheduled, Published, Partially Failed, Failed, Rejected, Under Review, Missed Review, Draft
- Each status has: tagId, stateImg (asset name), stateColor (UIColor), stateMsg (localized)
- **No `notification_sent` or `notification_declined` entries exist**

### Status Filter
- **Filter VC:** `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/Planner/PlannerStatusFilterViewController.swift`
- **Bottom Sheet:** `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/Planner/PlannerStatusBottomSheet.swift`
- Supports multi-select status filtering via `selectedStatusIds: [Int]`
- Status-to-API mapping in `ServiceManager.swift` (lines 159-180)

### Post Card / Badge
- **File:** `contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/PlannerPostCardView.swift` (lines 75-137)
- `statusColors` computed property switches on `post.post_state` string
- Badge: border color + semi-transparent background

### Post Preview
- **File:** `contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/PostPreviewView.swift`
- `PostPreviewDelegate` has: approvePost, rejectPost, editPost, replacePost, deletePost
- **No "I Published This" / "I Didn't Post This" actions exist**

### What Needs to Change
- Add indices 9 (`notification_sent`) and 10 (`notification_declined`) to `plannerStateData`
- Add amber color for Notification Sent, gray for Notification Declined
- Add status image assets for both
- Add localized strings for both statuses
- Add cases to `PlannerPostCardView.statusColors` and `statusText`
- Add cases to `ServiceManager` status-to-API mapping
- Add "I Published This" and "I Didn't Post This" buttons to `PostPreviewView` for Notification Sent posts
- After decline: show Edit/Reschedule/Delete actions for Notification Declined posts

---

## Android — Current State

### Status System
- **File:** `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Common/Constants.java` (lines 231-241)
- 9 statuses in `plannerStateData[]` array (indices 0-8): All, Scheduled, Published, Partially Failed, Failed, Rejected, Under Review, Missed Review, Draft
- Each has: index, icon drawable, color resource, title string, empty state message

### Status Filter / Tabs
- **Activity:** `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/PlannerActivity.java`
- Uses ViewPager + TabLayout with 9 tabs (one per status)
- Each tab → `PostsFragment` with status index

### Post Card / Badge
- **Adapter:** `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/PlannerListViewAdapter.java`
- Status shown as colored left border on `panelType` View
- Color from `getPlannerBackgroundColor(plannerType)`
- **Layout:** `item_planner.xml`

### Post Preview
- No dedicated post detail screen — details shown inline in list
- Action buttons: Approve, Reject, Edit, Delete, Duplicate, Replace
- **No "I Published This" / "I Didn't Post This" actions**

### What Needs to Change
- Add indices 9 (`notification_sent`) and 10 (`notification_declined`) to `plannerStateData[]`
- Add amber color for Notification Sent, gray for Notification Declined in `colors.xml`
- Add icon drawables for both statuses
- Add string resources for both statuses
- Add new tabs in `PlannerActivity` ViewPager
- Add color cases to `getPlannerBackgroundColor()`
- Add "I Published This" and "I Didn't Post This" buttons to post item layout + adapter
- After decline: show Edit/Reschedule/Delete actions for Notification Declined posts

---

## Files Involved

### iOS
- `ContentStudio/HelperClasses/Constant.swift` — status enum + plannerStateData
- `ContentStudio/HelperClasses/ServiceManager.swift` — API status mapping
- `ContentStudio/Views/Planner/SwiftUI/PlannerPostCardView.swift` — badge colors/text
- `ContentStudio/Views/Planner/SwiftUI/PostPreviewView.swift` — post detail actions
- `ContentStudio/Controllers/Nav Menu VCs/Planner/PlannerStatusFilterViewController.swift` — filter
- `ContentStudio/Controllers/Nav Menu VCs/Planner/PlannerStatusBottomSheet.swift` — filter UI
- Localization files (new keys for status names)
- Asset catalog (new status icons)

### Android
- `Common/Constants.java` — plannerStateData array
- `Planner/PlannerStateData.java` — status data class
- `Planner/PlannerActivity.java` — tab setup
- `Planner/Fragments/PlannerListViewAdapter.java` — list adapter
- `Planner/Fragments/FilterItems/PlannerItemHolder.java` — view holder
- `res/layout/item_planner.xml` — post item layout
- `res/values/colors.xml` — status colors
- `res/values/strings.xml` — status labels + empty state messages
