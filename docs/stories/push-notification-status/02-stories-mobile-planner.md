# Push Notification Status — Mobile Planner Stories

Epic: #93708 — Push notification status & handling improvements - web/mobile

These stories mirror **[FE] Allow manual resolution of push notification posts from web app** (sc-109141) for the native mobile apps.

---

## Story 1: [iOS] Display notification statuses and manual resolution in iOS planner

### Description:

The web app (sc-109141) adds support for displaying **Notification Sent** and **Notification Declined** post statuses in the Planner, with manual resolution buttons so users can mark posts as published or not published. The iOS app needs the same capabilities so users managing posts from their iPhone or iPad have a consistent experience.

This story adds:
- Two new status entries to the iOS planner status system
- Status badges with appropriate colors in the post list
- Status filter support so users can filter by Notification Sent or Notification Declined
- Manual resolution buttons ("I Published This" / "I Didn't Post This") on Notification Sent posts in the post preview
- Edit/Reschedule/Delete actions on Notification Declined posts

**Status badge colors (matching web):**
- **Notification Sent:** Amber — border `#F0BB52`, background `#FFF8EA` (same family as Scheduled but distinct)
- **Notification Declined:** Gray — border `#76797C`, background `#F3F3F3` (same family as Draft but distinct label)

**Files to modify:**
- `ContentStudio/HelperClasses/Constant.swift` — add indices 9 and 10 to `plannerStateData` with new `PlannerStateInfo` entries for `notification_sent` and `notification_declined`
- `ContentStudio/HelperClasses/ServiceManager.swift` — add cases 9 → `"notification_sent"` and 10 → `"notification_declined"` to the status-to-API mapping
- `ContentStudio/Views/Planner/SwiftUI/PlannerPostCardView.swift` — add `"notification_sent"` and `"notification_declined"` cases to `statusColors` and `statusText` computed properties
- `ContentStudio/Views/Planner/SwiftUI/PostPreviewView.swift` — add resolution buttons and confirmation flow for Notification Sent posts; add Edit/Reschedule/Delete for Notification Declined posts
- `ContentStudio/Controllers/Nav Menu VCs/Planner/PlannerStatusBottomSheet.swift` — new statuses appear in the filter list
- `ContentStudio/Controllers/Nav Menu VCs/Planner/PlannerStatusFilterViewController.swift` — support new status indices in multi-select
- Localization files — add keys for "Notification Sent", "Notification Declined", and all UI copy below
- Asset catalog — add status icons `notification_sent_post` and `notification_declined_post`

---

### Workflow:

**Viewing Notification Sent posts:**
1. User opens the Planner on iOS
2. User sees a post with an amber **Notification Sent** badge
3. User taps the post to open the post preview
4. User sees the post details plus two action buttons: **I Published This** and **I Didn't Post This**

**Marking as published:**
5. User taps **I Published This**
6. A confirmation alert appears:
   - Title: "Mark as published?"
   - Message: "This will mark the post for [Account Name] as published. This can't be undone."
   - Primary button: "Yes, Mark as Published"
   - Secondary button: "Cancel"
7. User taps "Yes, Mark as Published"
8. Post status updates to **Published** immediately
9. Success feedback: "Post marked as published."

**Marking as not published:**
5. User taps **I Didn't Post This**
6. A confirmation alert appears:
   - Title: "Mark as not published?"
   - Message: "This will mark the post for [Account Name] as not published. You'll still be able to edit and reschedule it from the Planner."
   - Primary button: "Confirm"
   - Secondary button: "Cancel"
7. User taps "Confirm"
8. Post status updates to **Notification Declined** (gray badge)
9. Feedback: "Post marked as not published. You can edit or reschedule it anytime."
10. Resolution buttons are replaced by Edit / Reschedule / Delete actions

**Filtering by notification statuses:**
1. User opens the Planner status filter (bottom sheet)
2. User sees **Notification Sent** and **Notification Declined** as selectable filter options
3. User selects **Notification Sent**
4. Planner list shows only posts with Notification Sent status

---

### Acceptance criteria:

- [ ] `plannerStateData` includes entries for Notification Sent (index 9) and Notification Declined (index 10) with correct colors, icons, and localized labels
- [ ] Posts with `notification_sent` status show an amber badge (border `#F0BB52`, background `#FFF8EA`) with label "Notification Sent" in the planner list
- [ ] Posts with `notification_declined` status show a gray badge (border `#76797C`, background `#F3F3F3`) with label "Notification Declined" in the planner list
- [ ] Status filter bottom sheet includes Notification Sent and Notification Declined options
- [ ] Filtering by Notification Sent shows only posts with that status
- [ ] Filtering by Notification Declined shows only posts with that status
- [ ] "I Published This" and "I Didn't Post This" buttons appear in post preview **only** for posts in Notification Sent status
- [ ] Tapping "I Published This" shows a confirmation alert with the copy specified in Workflow
- [ ] Confirming "I Published This" sends a PATCH request to update post status to `published`
- [ ] On success: post status updates to Published immediately; success feedback shown: "Post marked as published."
- [ ] Tapping "I Didn't Post This" shows a confirmation alert with the copy specified in Workflow
- [ ] Confirming "I Didn't Post This" sends a PATCH request to update post status to `notification_declined`
- [ ] On success: post status updates to Notification Declined; feedback shown: "Post marked as not published. You can edit or reschedule it anytime."
- [ ] After status changes to Notification Declined: resolution buttons are replaced with Edit / Reschedule / Delete actions
- [ ] Notification Declined posts retain all original post data (caption, media, account) and remain editable
- [ ] If API call fails: status remains unchanged; error alert shown: "Something went wrong. Please try again."
- [ ] On Cancel in either confirmation alert: no action taken, user stays on post preview
- [ ] All new user-facing strings are localized (added to all locale files)
- [ ] Empty state message for Notification Sent filter: "No posts with notification sent status"
- [ ] Empty state message for Notification Declined filter: "No posts with notification declined status"

---

### Mock-ups:

- Planner list view showing Notification Sent amber badge and Notification Declined gray badge
- Post preview with "I Published This" and "I Didn't Post This" action buttons
- Confirmation alerts for both actions
- Status filter bottom sheet with new options
  (Figma link to be added)

---

### Impact on existing data:

No impact on existing posts. New status badges and filter options are additive. Only posts that the backend assigns `notification_sent` or `notification_declined` status will display the new badges.

---

### Impact on other products:

- Status changes made on iOS sync to web planner and Android app
- Matches the behavior of **[FE] Allow manual resolution of push notification posts from web app** (sc-109141)

---

### Dependencies:

- **[BE] Add "Notification Sent" and "Notification Declined" post statuses** (sc-109136) — backend must support the new statuses and status transitions
- **[FE] Allow manual resolution of push notification posts from web app** (sc-109141) — for UI/UX parity reference

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, native iOS
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Story 2: [Android] Display notification statuses and manual resolution in Android planner

### Description:

Same scope as **[iOS] Display notification statuses and manual resolution in iOS planner**, implemented for Android.

The Android app needs to display **Notification Sent** and **Notification Declined** post statuses in the Planner, support filtering by them, and provide manual resolution buttons so users can mark posts as published or not published — matching the web app experience (sc-109141).

This story adds:
- Two new status entries to the Android planner status system
- Status badge colors (left border) for both new statuses
- New tabs in the Planner ViewPager for Notification Sent and Notification Declined
- Manual resolution buttons ("I Published This" / "I Didn't Post This") on Notification Sent posts in the post list
- Edit/Reschedule/Delete actions on Notification Declined posts

**Status border colors (matching web):**
- **Notification Sent:** Amber — `#F0BB52`
- **Notification Declined:** Gray — `#76797C`

**Files to modify:**
- `Common/Constants.java` — add indices 9 and 10 to `plannerStateData[]` with new `PlannerStateData` entries
- `Planner/PlannerActivity.java` — ViewPager/TabLayout supports the two new tabs
- `Planner/Fragments/PlannerListViewAdapter.java` — add color cases to `getPlannerBackgroundColor()` for indices 9 and 10; add "I Published This" and "I Didn't Post This" buttons for Notification Sent posts
- `Planner/Fragments/FilterItems/PlannerItemHolder.java` — add button views for resolution actions
- `res/layout/item_planner.xml` — add resolution button layout (hidden by default, shown for Notification Sent posts)
- `res/values/colors.xml` — add `notification_sent` (#F0BB52) and `notification_declined` (#76797C) colors
- `res/values/strings.xml` — add labels, empty state messages, button text, confirmation dialog copy
- Drawable resources — add `notification_sent_icon` and `notification_declined_icon` drawables

---

### Workflow:

**Viewing Notification Sent posts:**
1. User opens the Planner on Android
2. User sees the **Notification Sent** tab in the tab bar, or sees a post with an amber left border in the All tab
3. User taps the **Notification Sent** tab to filter
4. User sees only posts with Notification Sent status
5. Each post shows two action buttons: **I Published This** and **I Didn't Post This**

**Marking as published:**
6. User taps **I Published This**
7. A confirmation dialog appears:
   - Title: "Mark as published?"
   - Message: "This will mark the post for [Account Name] as published. This can't be undone."
   - Positive button: "Yes, Mark as Published"
   - Negative button: "Cancel"
8. User taps "Yes, Mark as Published"
9. Post status updates to **Published** immediately
10. Success snackbar: "Post marked as published."

**Marking as not published:**
6. User taps **I Didn't Post This**
7. A confirmation dialog appears:
   - Title: "Mark as not published?"
   - Message: "This will mark the post for [Account Name] as not published. You'll still be able to edit and reschedule it from the Planner."
   - Positive button: "Confirm"
   - Negative button: "Cancel"
8. User taps "Confirm"
9. Post status updates to **Notification Declined** (gray left border)
10. Snackbar: "Post marked as not published. You can edit or reschedule it anytime."
11. Resolution buttons are replaced by Edit / Reschedule / Delete actions

**Viewing Notification Declined posts:**
1. User taps the **Notification Declined** tab
2. User sees posts with gray left border and Notification Declined label
3. Each post shows Edit / Reschedule / Delete action buttons (same as Draft posts)

---

### Acceptance criteria:

- [ ] `plannerStateData[]` includes entries for Notification Sent (index 9) and Notification Declined (index 10) with correct icons, colors, titles, and empty state messages
- [ ] Posts with `notification_sent` status show an amber left border (`#F0BB52`) in the planner list
- [ ] Posts with `notification_declined` status show a gray left border (`#76797C`) in the planner list
- [ ] Planner TabLayout includes **Notification Sent** and **Notification Declined** tabs
- [ ] Tapping the Notification Sent tab shows only posts with that status
- [ ] Tapping the Notification Declined tab shows only posts with that status
- [ ] "I Published This" and "I Didn't Post This" buttons appear in the post item **only** for posts in Notification Sent status
- [ ] Tapping "I Published This" shows a confirmation dialog with the copy specified in Workflow
- [ ] Confirming "I Published This" sends a PATCH request to update post status to `published`
- [ ] On success: post status updates to Published immediately; snackbar shown: "Post marked as published."
- [ ] Tapping "I Didn't Post This" shows a confirmation dialog with the copy specified in Workflow
- [ ] Confirming "I Didn't Post This" sends a PATCH request to update post status to `notification_declined`
- [ ] On success: post status updates to Notification Declined; snackbar shown: "Post marked as not published. You can edit or reschedule it anytime."
- [ ] After status changes to Notification Declined: resolution buttons are replaced with Edit / Reschedule / Delete actions
- [ ] Notification Declined posts retain all original post data and remain editable
- [ ] If API call fails: status remains unchanged; error snackbar shown: "Something went wrong. Please try again."
- [ ] On Cancel in either confirmation dialog: no action taken
- [ ] All new user-facing strings are added to `strings.xml` (and localized variants)
- [ ] Empty state for Notification Sent tab: "No posts with notification sent status" with appropriate icon
- [ ] Empty state for Notification Declined tab: "No posts with notification declined status" with appropriate icon
- [ ] `getPlannerBackgroundColor()` returns correct color resource for indices 9 and 10

---

### Mock-ups:

- Planner tab bar showing new Notification Sent and Notification Declined tabs
- Post list item with amber left border and resolution buttons
- Post list item with gray left border and Edit/Reschedule/Delete buttons
- Confirmation dialogs for both actions
  (Figma link to be added)

---

### Impact on existing data:

No impact on existing posts. New status tabs and badges are additive. Only posts with `notification_sent` or `notification_declined` status from the API will show the new UI.

---

### Impact on other products:

- Status changes made on Android sync to web planner and iOS app
- Matches the behavior of **[FE] Allow manual resolution of push notification posts from web app** (sc-109141)

---

### Dependencies:

- **[BE] Add "Notification Sent" and "Notification Declined" post statuses** (sc-109136) — backend must support the new statuses and status transitions
- **[FE] Allow manual resolution of push notification posts from web app** (sc-109141) — for UI/UX parity reference

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, native Android
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
