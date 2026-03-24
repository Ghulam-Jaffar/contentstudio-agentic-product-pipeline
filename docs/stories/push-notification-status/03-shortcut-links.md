# Push Notification Status & Handling — Shortcut Links

Epic: https://app.shortcut.com/contentstudio-team/epic/93708

---

## Stories

| # | Title | Link |
|---|---|---|
| sc-109136 | [BE] Add "Notification Sent" and "Notification Declined" post statuses | https://app.shortcut.com/contentstudio-team/story/109136 |
| sc-109138 | [BE] Create per-account posts for push notification publishing | https://app.shortcut.com/contentstudio-team/story/109138 |
| sc-109139 | [iOS] Allow users to decline push notification posts from iOS app | https://app.shortcut.com/contentstudio-team/story/109139 |
| sc-112172 | [Android] Allow users to decline push notification posts from Android app | https://app.shortcut.com/contentstudio-team/story/112172 |
| sc-109140 | [iOS] Confirm push notification publishing outcome on iOS | https://app.shortcut.com/contentstudio-team/story/109140 |
| sc-112183 | [Android] Confirm push notification publishing outcome on Android | https://app.shortcut.com/contentstudio-team/story/112183 |
| sc-109141 | [FE] Allow manual resolution of push notification posts from web app | https://app.shortcut.com/contentstudio-team/story/109141 |
| sc-114526 | [iOS] Display notification statuses and manual resolution in iOS planner | https://app.shortcut.com/contentstudio-team/story/114526 |
| sc-114527 | [Android] Display notification statuses and manual resolution in Android planner | https://app.shortcut.com/contentstudio-team/story/114527 |

## Changes made (2026-03-24)

- **114526** — NEW [iOS] story: notification status badges, filters, and manual resolution buttons in iOS planner (mirrors sc-109141 for iOS)
- **114527** — NEW [Android] story: notification status badges, filters, and manual resolution buttons in Android planner (mirrors sc-109141 for Android)

## Changes made (2026-02-23)

- **109136** — Renamed, added [BE] prefix, expanded to cover both new statuses (`notification_sent` + `notification_declined`), clarified status transitions, fixed workflows to user POV
- **109138** — Renamed, added [BE] prefix, cleaned up AC and workflow
- **109139** — Scoped to iOS only, renamed to [iOS], fixed "Rejected" → "Notification Declined" throughout
- **112172** — NEW [Android] story (split from 109139)
- **109140** — Scoped to iOS only, renamed to [iOS], fixed "Rejected" → "Notification Declined"
- **112183** — NEW [Android] story (split from 109140)
- **109141** — Added [FE] prefix, fixed "Rejected" → "Notification Declined", added full UI copy (status badge copy/colors, button labels, tooltips, confirmation modals, toasts, error states, filter labels)
