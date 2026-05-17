# Research — Q2 2026: iOS Improvements

**Epic theme:** UI consistency improvements across the iOS app + incremental migration from UIKit/Swift to SwiftUI. Four focus areas: Inbox, Settings (profile/password/about/KB/help desk), Workspace, and the shared SwiftUI component library.

**Pipeline:** `/story` (local docs only — not pushed to Shortcut)
**Platforms:** iOS only (no Android, web, or backend changes)
**Repo:** `contentstudio-ios-v2/`

---

## Current State

The iOS app (`contentstudio-ios-v2/`) is a long-lived UIKit codebase. Most navigation, list views, and forms are implemented as `UIViewController` subclasses backed by `.xib` / `.storyboard` files. A SwiftUI surface exists but is partial — concentrated in the Planner, Comments, and Paywall flows, plus a small set of shared SwiftUI primitives (`Theme.swift`, `Buttons.swift`, `Textfields.swift`, `Texts.swift`, `UIComponents.swift`, `InfoBannerView.swift`, `LoadingOverlay.swift`, `CommonButtons.swift`).

**Prevalence today:**
- `import SwiftUI` files: ~126
- `class … : UIViewController` files: ~87
- The four focus areas (Inbox, Settings, Workspace, shared library) are still predominantly UIKit/xib.

### Inbox

| Surface | File | Size | Stack |
|---|---|---|---|
| Main inbox list | [InboxViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Inbox/InboxViewController.swift) | 361 lines | UIKit + xib |
| Filters | [FiltersViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Inbox/Filters/FiltersViewController.swift) + `FilterXibs/`, `Social Accounts/`, `Tags/`, `Team Members/` | 447 lines + xibs | UIKit + xib |
| Conversation | [ConversationViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Inbox/Conversation/ConversationViewController.swift) + `ConversationXibs/` | 1857 lines | UIKit + xib |
| Post comments | [PlatformPostViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Inbox/Post/PlatformPostViewController.swift) + `PostCommentSegmentTblCell` | 2844 lines | UIKit + xib |
| Cells / data | [InboxDataTableViewCell.swift](contentstudio-ios-v2/ContentStudio/Views/Social%20Inbox/InboxDataTableViewCell.swift), [InboxHeader.swift](contentstudio-ios-v2/ContentStudio/Views/Social%20Inbox/Header/InboxHeader.swift), [InboxBuilder.swift](contentstudio-ios-v2/ContentStudio/Builder/Inbox/InboxBuilder.swift) | — | UIKit + xib |
| Localization | [Localization/Inbox/inbox_en.json](contentstudio-ios-v2/ContentStudio/Localization/Inbox/inbox_en.json) (de/el/es/fr/it/pl/zh) | — | — |
| Existing partial SwiftUI on this surface | [CommentActionsBottomSheetView.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/CommentActionsBottomSheetView.swift), [CommentImageGridView.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/CommentImageGridView.swift), [PostPreviewCommentsView.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/PostPreviewCommentsView.swift), [PostPreviewCommentBubbleView.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/PostPreviewCommentBubbleView.swift), [PostPreviewCommentInputBar.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/PostPreviewCommentInputBar.swift), [EmojiPickerView.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/EmojiPickerView.swift), [MentionHighlightedText.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/MentionHighlightedText.swift), [CommentReactionChipView.swift](contentstudio-ios-v2/ContentStudio/Views/Comments/CommentReactionChipView.swift) | — | SwiftUI |

### Settings (Profile, Password, About, KB, Help desk)

| Surface | File | Size | Stack |
|---|---|---|---|
| Side menu (entry point) | [MenuVC.swift](contentstudio-ios-v2/ContentStudio/Controllers/Menu/MenuVC.swift) + cells (`MenuTblCell`, `SwitchWorkspaceTblCell`, `LogoutMenuTblCell`, `MenuSpaceTblCell`, `VersionInfoTblCell`) | 856 lines | UIKit + xib |
| Profile | [ProfileViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Setting/ProfileViewController.swift) | 259 lines | UIKit + xib |
| Change password | [ChangePasswordViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Setting/ChangePasswordViewController.swift) | 222 lines | UIKit + xib |
| About / Help / Help Desk (shared shell) | [SimilarMenuViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Nav%20Menu%20VCs/Setting%20-%20About%20-%20Help/SimilarMenuViewController.swift) + [AboutUsTableViewCell.swift](contentstudio-ios-v2/ContentStudio/Views/About%20Us/AboutUsTableViewCell.swift) | 298 lines | UIKit + xib |
| Knowledge Base / Privacy / ToU | `Controllers/Nav Menu VCs/KB - Privacy Policy - ToU/Web View/` | — | UIKit `WKWebView` |
| Settings localization | [Settings/settings_en.json](contentstudio-ios-v2/ContentStudio/Settings/settings_en.json) (de/el/es/fr/it/pl/zh) | — | — |
| Account deletion (already SwiftUI) | [DeleteAccountFlowView.swift](contentstudio-ios-v2/ContentStudio/Controllers/Account/DeleteAccountFlowView.swift), step 1/2/3 views | — | SwiftUI |

`SimilarMenuViewController` is a generic table-view shell reused for several static pages (About Us, Help, Help Desk, etc.) — driven by configuration. Rebuilding it in SwiftUI replaces multiple endpoints at once.

### Workspace

| Surface | File | Size | Stack |
|---|---|---|---|
| Workspace list / picker | [WorkspaceViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Workspace/WorkspaceViewController.swift) | 442 lines | UIKit + xib |
| Locked workspace state | [LockedWorkspaceViewController.swift](contentstudio-ios-v2/ContentStudio/Controllers/Workspace/LockedWorkspaceViewController.swift) | 115 lines | UIKit + xib |
| Cell | [WorkspaceTableViewCell.swift](contentstudio-ios-v2/ContentStudio/Views/Workspace/Cell/WorkspaceTableViewCell.swift) + xib | — | UIKit |
| Models | `Modals/Workspace/` | — | — |

### Shared SwiftUI component library

| File | Purpose |
|---|---|
| [Theme.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/Theme.swift) | Colors, typography, spacing tokens |
| [Buttons.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/Buttons.swift), [CommonButtons.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/CommonButtons.swift), [ButtonAndLabelView.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/ButtonAndLabelView.swift) | Button variants |
| [Textfields.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/Textfields.swift), [Texts.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/Texts.swift) | Text input + label primitives |
| [UIComponents.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/UIComponents.swift) | Misc reusable components |
| [InfoBannerView.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/InfoBannerView.swift), [LoadingOverlay.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/LoadingOverlay.swift) | Feedback / state components |
| [SocialLoginButton.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/SocialLoginButton.swift), [SignUpComponents.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/SignUpComponents.swift), [TermsPrivacyLinksView.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/TermsPrivacyLinksView.swift) | Auth-domain components |
| [SwiftUIAlertExtensions.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/SwiftUIAlertExtensions.swift), [SwiftUINavigationHelper.swift](contentstudio-ios-v2/ContentStudio/SwiftUI/Helpers/SwiftUINavigationHelper.swift) | Helpers & extensions |
| Bridge | [UIKitSwiftUIBridge.swift](contentstudio-ios-v2/ContentStudio/Views/Planner/SwiftUI/UIKitSwiftUIBridge.swift) (currently scoped to Planner) |

**Gaps observed (component library):**
- No central catalog/preview gallery (no Storybook-equivalent or SwiftUI `#Preview` index)
- Tokens (Theme.swift) cover basics but are not enforced — colors/spacing are still inlined in many SwiftUI files
- No shared list/row, search bar, segmented control, empty state, error state, bottom sheet, avatar, badge, chip, toolbar primitives — Inbox / Settings / Workspace each ship their own xibs
- No reusable `WebView` wrapper for KB/Privacy/ToU
- No formal design-system docs file for iOS (the web side has [docs/ui-components.md](docs/ui-components.md), iOS does not)

---

## What Needs to Change

### Inbox (rebuild on SwiftUI, consistent IA)
- **Main inbox screen:** rebuild list cell + header + bottom action sheet in SwiftUI; preserve filter pill row, unread badge, platform icon, conversation snippet, timestamp. Keep current behavior (read/unread, archive, assign, tag).
- **Filters:** consolidate Social Accounts / Tags / Team Members / saved-filters sub-screens under a single SwiftUI filter sheet — same options, more consistent layout, persistent state on close.
- **Conversation:** rebuild the message thread, input bar, attachments preview, save-reply menu, assignment header in SwiftUI. Keep all message types (text, image, emoji reactions, saved replies, internal notes, attachments).
- **Post comments:** rebuild the PlatformPostViewController surface (post header, threaded comments, reaction bar, comment input, YouTube preview) in SwiftUI, reusing the existing SwiftUI primitives in `Views/Comments/`.
- **No behavior change** — purely UI rebuild + consistency. Same API endpoints, same data flow.

### Settings (UI consistency + SwiftUI migration)
- **Profile screen:** SwiftUI form — avatar upload, name, email (read-only), timezone, language, save button. Same fields as today.
- **Change password:** SwiftUI form — current password, new password, confirm. Same validations, same submit endpoint.
- **About Us / Help / Help Desk:** replace the `SimilarMenuViewController` xib-driven shell with a SwiftUI list pattern (links + version info + section headers). Same content, same destination URLs.
- **Knowledge Base / Privacy / ToU:** keep `WKWebView` for the body, but wrap navigation chrome (title bar, loading state, error state) in SwiftUI for visual consistency.
- **Side menu (MenuVC):** rebuild the side-menu list cells + section headers + workspace switcher in SwiftUI. Same destinations, same logout / version cells.

### Workspace
- **Workspace list/picker:** rebuild in SwiftUI — workspace name, role badge, members count, current-workspace indicator, locked state. Keep the existing tap-to-switch + long-press-to-leave behavior.
- **Locked workspace screen:** SwiftUI illustration + upgrade CTA.
- **Cell consistency:** replace `WorkspaceTableViewCell.xib` with a SwiftUI row used by Profile/Settings/Side-menu too.

### Shared SwiftUI component library (foundation)
- **Token enforcement:** extend `Theme.swift` to be the single source of truth — colors, typography, spacing, radii, elevations. Add SwiftUI `ViewModifier`s so callers can't pass raw hex/pixel values.
- **Component set additions:** add the missing primitives the other three stories will lean on — `CSListRow`, `CSSectionHeader`, `CSSearchField`, `CSSegmentedControl`, `CSEmptyState`, `CSErrorState`, `CSBottomSheet`, `CSAvatar`, `CSBadge`, `CSChip`, `CSToolbar`, `CSWebView`. Naming matches the web app's `@contentstudio/ui` convention.
- **Preview gallery:** add a debug-only `ComponentGalleryView` listing every primitive with `#Preview` blocks so designers/engineers can scan them.
- **Docs:** add `contentstudio-ios-v2/ContentStudio/SwiftUI/README.md` describing tokens, components, and the "build in SwiftUI, bridge through `UIKitSwiftUIBridge` only when wrapping into UIKit nav" rule.
- **Bridge:** generalize `UIKitSwiftUIBridge.swift` out of `Views/Planner/SwiftUI/` so all four focus areas can use it.

---

## UX Reference

Existing ContentStudio web app patterns (Inbox, Workspace, Settings) are the canonical UI reference — iOS should match those affordances where they make sense on mobile. SwiftUI gives access to native iOS patterns (Form, List, NavigationStack, sheet presentations, ContextMenu) that aren't worth re-inventing — use them.

---

## Mobile Context

This epic is iOS-only. Android has its own separate epic ([Q2 2026: Android improvements](docs/stories/q2-2026-android-improvements/01-research.md)) focused on Planner parity + post preview — explicitly **out of scope** here.

Existing iOS app supports:
- iOS 15+ (per Podfile)
- Light theme only (no dark mode — per project rule)
- 8 locales: en, de, el, es, fr, it, pl, zh
- No white-label theming on mobile

---

## Files Involved (high-level)

**Inbox:**
- `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/Inbox/` (entire subtree)
- `contentstudio-ios-v2/ContentStudio/Views/Social Inbox/`
- `contentstudio-ios-v2/ContentStudio/Views/Social InboxAccount/`
- `contentstudio-ios-v2/ContentStudio/Modals/Inbox/`
- `contentstudio-ios-v2/ContentStudio/Builder/Inbox/InboxBuilder.swift`
- `contentstudio-ios-v2/ContentStudio/Localization/Inbox/*.json`

**Settings:**
- `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/Setting/`
- `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/Setting - About - Help/`
- `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/KB - Privacy Policy - ToU/`
- `contentstudio-ios-v2/ContentStudio/Controllers/Menu/`
- `contentstudio-ios-v2/ContentStudio/Views/About Us/`
- `contentstudio-ios-v2/ContentStudio/Settings/*.json`

**Workspace:**
- `contentstudio-ios-v2/ContentStudio/Controllers/Workspace/`
- `contentstudio-ios-v2/ContentStudio/Views/Workspace/`
- `contentstudio-ios-v2/ContentStudio/Modals/Workspace/`

**UI Component library:**
- `contentstudio-ios-v2/ContentStudio/SwiftUI/` (entire subtree)
- New: `contentstudio-ios-v2/ContentStudio/SwiftUI/Components/` (expanded primitives)
- New: `contentstudio-ios-v2/ContentStudio/SwiftUI/README.md`
- Move/promote: `Views/Planner/SwiftUI/UIKitSwiftUIBridge.swift` → `SwiftUI/Helpers/`

---

## Out of Scope

- Dark mode (project-wide rule — iOS does not support dark mode)
- RTL layout (project-wide rule)
- Backend / API changes — every story in this epic is UI-only
- Android (covered by the parallel Q2 2026 Android epic)
- Functional changes to Inbox/Settings/Workspace behaviors — UI rebuild only
- New SwiftUI screens outside the four areas (e.g. Planner is already SwiftUI; Composer is its own track)
