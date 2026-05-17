# Research — Q2 2026: Android Improvements

**Epic theme:** Bring the Android Planner up to parity with iOS (UI + options) and add a Post Preview surface that today only exists on iOS.

**Pipeline:** `/story` (local docs only — not pushed to Shortcut)
**Platforms:** Android only (no iOS, web, or backend changes)
**Repo:** `contentstudio-android-v2/`

---

## Current State

The Android app (`contentstudio-android-v2/`) is a Java/XML codebase using the legacy Activity + Fragment + RecyclerView pattern (no Compose, no MVVM enforcement). Package root is `com.muneeb.lumotive`.

### Planner (Android — today)

| Surface | File | Size | Stack |
|---|---|---|---|
| Activity | [PlannerActivity.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/PlannerActivity.java) | 787 lines | Java + XML |
| State | [PlannerStateData.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/PlannerStateData.java) | — | Java |
| Filter fragment | [PlannerFilterFragment.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/PlannerFilterFragment.java) | — | Java |
| List adapter | [PlannerListViewAdapter.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/PlannerListViewAdapter.java), [PlannerListItem.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/PlannerListItem.java), [SectionHeaderPlannerList.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/SectionHeaderPlannerList.java) | — | Java |
| Posts fragment | [PostsFragment.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/PostsFragment.java) + [BasePostsFragment.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/BasePostsFragment.java) | — | Java |
| Approve/Reject | `Fragments/ApproveRejectFragment/` | — | Java |
| Notification status | [PlannerNotificationStatusConfirmDialogFragment.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/Fragments/PlannerNotificationStatusConfirmDialogFragment.java) | — | Java |
| Filter sub-screens | `PlannerOtherFilter/ByLabel`, `PlannerOtherFilter/ByMember`, `PlannerOtherFilter/ByOrder`, `PlannerOtherFilter/ByType` | — | Java |
| Social-account picker | [PlannerSocialAccountSelectionFragment.java](contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/PlannerSocialAccountSelection/PlannerSocialAccountSelectionFragment.java) | — | Java |
| Layouts | [activity_planner.xml](contentstudio-android-v2/app/src/main/res/layout/activity_planner.xml), [content_planner.xml](contentstudio-android-v2/app/src/main/res/layout/content_planner.xml), [item_planner.xml](contentstudio-android-v2/app/src/main/res/layout/item_planner.xml), [planner_filters.xml](contentstudio-android-v2/app/src/main/res/layout/planner_filters.xml), [section_header_planner_view.xml](contentstudio-android-v2/app/src/main/res/layout/section_header_planner_view.xml), [fragment_planner_add_comment_view.xml](contentstudio-android-v2/app/src/main/res/layout/fragment_planner_add_comment_view.xml), [fragment_planner_approve_schedule_view.xml](contentstudio-android-v2/app/src/main/res/layout/fragment_planner_approve_schedule_view.xml), [fragment_planner_social_account_selection.xml](contentstudio-android-v2/app/src/main/res/layout/fragment_planner_social_account_selection.xml) | — | XML |
| Strings | [values/planner_strings.xml](contentstudio-android-v2/app/src/main/res/values/planner_strings.xml) (de/el/es/fr/it/pl/zh) | — | — |

### Planner (iOS — reference)

| Surface | File | Size | Stack |
|---|---|---|---|
| Base VC | [PlannerBaseViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Planner/PlannerBaseViewController.swift) | — | UIKit + SwiftUI bridge |
| Filter options | [PlannerFilterOptionsView.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Planner/PlannerFilterOptionsView.swift) | — | SwiftUI |
| Status filter | [PlannerStatusFilterViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Planner/PlannerStatusFilterViewController.swift), [PlannerStatusBottomSheet.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Planner/PlannerStatusBottomSheet.swift) | — | SwiftUI |
| Sort bottom sheet | [PlannerSortBottomSheetPreview.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Planner/PlannerSortBottomSheetPreview.swift) | — | SwiftUI |
| Post card | [PlannerPostCardView.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/PlannerPostCardView.swift), [PlannerDataTableViewCellSwiftUI.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/PlannerDataTableViewCellSwiftUI.swift) | — | SwiftUI |
| Media preview | [MediaPreviewView.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/MediaPreviewView.swift) | — | SwiftUI |
| **Post preview (the canonical UI)** | [PostPreviewView.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/PostPreviewView.swift) | **3334 lines** | SwiftUI |
| Older UIKit post preview (still present) | [PostPreviewViewController.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/PostPreviewViewController.swift) | 708 lines | UIKit |
| Other Filters | [Views/Planner/Other Filter/](contentstudio-ios-v2/ContentStudio/Views/Planner/Other%20Filter/) | — | UIKit |
| Approve/reject (SwiftUI) | [ApproveRejectPostSwiftUI.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/Approve%20Reject/SwiftUI/ApproveRejectPostSwiftUI.swift) | — | SwiftUI |

### Post Preview on Android — today

A grep of `contentstudio-android-v2/app/src/main/` for `PostPreview`, `post_preview`, `previewPost` returns **zero matches**. The Android app has **no post-preview surface at all** in the Planner — tapping a post jumps straight to the approve/reject/edit flow without a per-platform rendered preview of how the post will appear on each network.

---

## What Needs to Change

### Planner revamp (Android — parity with iOS)

Bring Android Planner UI options and visual layout to match iOS:

- **List card layout** — match `PlannerPostCardView` from iOS: post thumbnail, platform icons, scheduled-for timestamp, status pill (scheduled / published / rejected / failed / draft / evergreen / repeat), social-account row, action menu (replace / edit / delete / approve), repeat indicator.
- **Status filter row / bottom sheet** — match iOS `PlannerStatusBottomSheet`: All, Scheduled, Published, Failed, Rejected, Draft, In Review, Missed Review, Content Category, External Action.
- **Sort bottom sheet** — match iOS `PlannerSortBottomSheetPreview`: by date asc/desc, by social account, by type, by label, by member.
- **Other Filter screen** — consolidate Android's `PlannerOtherFilter/ByLabel`, `ByMember`, `ByOrder`, `ByType` into a single filter sheet matching the iOS `PlannerFilterOptionsView` layout: section headers, multi-select chips, "Clear all" / "Apply" actions, sticky CTA.
- **Section headers** — date-grouped sections styled the same way as iOS (day name + date, sticky on scroll, count badge).
- **Empty state** — illustration + headline + sub-copy matching iOS.
- **Approve / Reject flow UI** — match the iOS `ApproveRejectPostSwiftUI` layout (status, post preview, comment box, approve/reject CTAs).
- **Social account selector** — match `PlannerSocialAccountSelection` styling on iOS.
- **No data-model or API changes** — same endpoints, same status transitions, same notification triggers.

### Post Preview in Planner (Android — new)

A net-new screen on Android, mirroring iOS [PostPreviewView.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/PostPreviewView.swift):

- Tapping a post in the Planner list opens a full post-preview screen.
- Per-platform rendered preview (Facebook, Instagram feed/reel/story, LinkedIn, Twitter/X, TikTok, YouTube, Pinterest, Threads, Bluesky, GMB) — switchable via the platform tabs at the top.
- Sections: post header (profile avatar + name + scheduled time), media (image grid / video / carousel), caption, hashtags, first-comment, link-preview / OG card, like/reaction counts (mocked), location, mentions.
- Action row: edit, duplicate, replace media, delete, approve, reject, share preview link.
- Comments / approval-discussion thread inline (same data the iOS `PostPreviewCommentsView` already shows).
- Reuses existing data — no new API. Same `Submission`/`PlannerListItem` model that `PostsFragment` already loads.

---

## UX Reference

iOS Planner + iOS Post Preview are the canonical reference. No external benchmarks needed — every option on iOS must exist on Android, no extras.

---

## Mobile Context

This epic is **Android-only**. The iOS Planner + Post Preview are already shipped (see iOS reference files above) and are the source of truth for visual + interaction parity. iOS-side scope lives in [Q2 2026: iOS improvements](docs/stories/q2-2026-ios-improvements/01-research.md) and explicitly excludes Planner.

Existing Android app:
- minSdk per `app/build.gradle`
- Light theme only (no dark mode — project rule)
- 8 locales (en, de, el, es, fr, it, pl, zh)
- No Compose adoption yet — stories should use the existing Java + XML stack to stay consistent with the rest of the app, unless engineering decides otherwise during planning

---

## Files Involved (high-level)

**Planner revamp:**
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/` (entire subtree)
- `contentstudio-android-v2/app/src/main/res/layout/activity_planner.xml`
- `contentstudio-android-v2/app/src/main/res/layout/content_planner.xml`
- `contentstudio-android-v2/app/src/main/res/layout/item_planner.xml`
- `contentstudio-android-v2/app/src/main/res/layout/planner_filters.xml`
- `contentstudio-android-v2/app/src/main/res/layout/section_header_planner_view.xml`
- `contentstudio-android-v2/app/src/main/res/layout/fragment_planner_*.xml`
- `contentstudio-android-v2/app/src/main/res/values*/planner_strings.xml`
- `contentstudio-android-v2/app/src/main/res/drawable*/planner_*.png` (likely needs new vectors)

**Post Preview (new):**
- New: `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Planner/PostPreview/`
  - `PostPreviewActivity.java` (or Fragment, depending on engineering preference)
  - `PostPreviewPlatformTabAdapter.java`
  - Per-platform preview fragments: `FacebookPreviewFragment`, `InstagramPreviewFragment`, `LinkedInPreviewFragment`, `TwitterPreviewFragment`, `TikTokPreviewFragment`, `YouTubePreviewFragment`, `PinterestPreviewFragment`, `ThreadsPreviewFragment`, `BlueskyPreviewFragment`, `GMBPreviewFragment`
- New layout files: `activity_post_preview.xml`, `fragment_post_preview_<platform>.xml`, `item_post_preview_media.xml`, `item_post_preview_comment.xml`
- New strings: `values*/post_preview_strings.xml`
- New drawables for platform-chrome glyphs (likes, comments, share icons per platform)

---

## Out of Scope

- Dark mode (project-wide rule)
- RTL layout (project-wide rule)
- Backend / API changes — same endpoints, same payloads
- iOS (covered by the parallel Q2 2026 iOS epic)
- Migrating Android to Jetpack Compose (separate, much larger decision — flag for engineering but do not assume)
- Composer / publishing flow on Android (out of this epic)
- Inbox / Settings / Workspace on Android (out of this epic)
