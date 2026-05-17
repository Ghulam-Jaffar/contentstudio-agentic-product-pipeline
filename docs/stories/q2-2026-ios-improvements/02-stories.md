# Stories — Q2 2026: iOS Improvements

**Epic (notional, not in Shortcut):** Q2 2026: iOS improvements
**Platform:** iOS
**Type:** UI consistency + UIKit → SwiftUI migration
**Pipeline:** local `/story` (no Shortcut push)

**Story order / dependency:**
Story 4 (UI Component Library) is the foundation — its primitives are referenced by Stories 1–3. Engineering should plan Story 4 to land first or in parallel with the first slice of the others. Each story body declares the dependency explicitly.

| # | Story |
|---|---|
| 1 | [iOS] Rebuild the Inbox surface (list, filters, conversation, post comments) in SwiftUI |
| 2 | [iOS] Rebuild the Settings surface (profile, password, about, help, knowledge base, side menu) in SwiftUI |
| 3 | [iOS] Rebuild the Workspace screen in SwiftUI |
| 4 | [iOS] Expand the shared SwiftUI component library (tokens, primitives, preview gallery) |

---

## Story 1 — [iOS] Rebuild the Inbox surface (list, filters, conversation, post comments) in SwiftUI

### Description
As an iOS user managing my social inbox, I want the Inbox list, filters, conversation thread, and post-comments screens to look and behave consistently with the rest of the app, so that the experience feels modern, predictable, and matches what I see on the web app and in the parts of the iOS app that already use the new UI.

This is a UI rebuild — same data, same endpoints, same features. Every screen under the Inbox section is migrated from UIKit + `.xib` to SwiftUI, sharing the component library from Story 4.

### Workflow
1. The user taps the **Inbox** tab in the side menu.
2. The user lands on the **Inbox list**: a scrollable list of conversations. Each row shows the platform icon, account avatar, contact name, message snippet, timestamp, unread badge, status pill (assigned, archived, etc.), and any applied tags. The user can pull to refresh, swipe a row for quick actions (archive, mark read), or tap the row to open the conversation.
3. The user taps the **filter** icon in the top bar. A unified filter sheet appears with sections for: status, social accounts, tags, assignees, date range, and saved filters. The user toggles options; the sheet shows a live count of matches; tapping **Apply** closes the sheet and refreshes the list. The filter chips persist as a pill row above the list.
4. The user taps a conversation. The **conversation screen** opens, showing the message thread with the contact, the input bar at the bottom, the conversation header (avatar, name, platform, assignment status), and an action menu (archive, assign, tag, mark unread, save reply). Messages render with reactions, attachments, images, and saved-reply chips inline. The user can send a text reply, attach media, pick a saved reply, mention a team member with an internal note, or react to a message.
5. From the conversation header (or from the Inbox tab for "Post comments"), the user can open the **post comments** screen for a published post. This shows the post header (platform, account, post media + caption), the threaded comment list with reactions, and a comment input bar at the bottom. The user can reply, like, hide, delete, or mark a comment as resolved.
6. From any screen, tapping back returns the user to the previous level with state preserved (scroll position, applied filters, draft reply text).

### Acceptance criteria
- [ ] The Inbox tab opens to a SwiftUI inbox list — no UIKit/xib for the rendered list row, header, or empty state.
- [ ] Each conversation row shows: platform icon, account avatar (fallback initials when no image), contact name, latest-message snippet (1 line, truncated), relative timestamp, unread dot when applicable, status pill, and applied tag chips.
- [ ] Swipe-left on a row reveals quick actions: **Archive**, **Mark unread/read**. Swipe-right reveals **Assign**. Actions complete optimistically and roll back on API failure with a toast.
- [ ] Pull-to-refresh on the list re-fetches conversations and shows a SwiftUI refresh spinner.
- [ ] Tapping the filter icon opens a SwiftUI bottom sheet with sections: **Status**, **Social accounts**, **Tags**, **Assignees**, **Date range**, **Saved filters**. The sheet uses `CSBottomSheet` from the component library.
- [ ] Filter sheet shows a live "X conversations match" counter as the user toggles options.
- [ ] "Apply" closes the sheet and updates the list; "Clear all" resets every filter to default; "Cancel" or drag-to-dismiss discards changes.
- [ ] Active filters render as a horizontally scrollable chip row above the list using `CSChip`. Tapping a chip's `×` removes that single filter.
- [ ] Saved-filter selection updates the chips and the list together.
- [ ] Tapping a row opens the conversation screen in SwiftUI; no `UIViewController.present` of the legacy `ConversationViewController` remains in the inbox flow.
- [ ] Conversation screen shows: header (avatar, contact name, platform icon, assignment badge), scrollable message thread, input bar pinned to the bottom, and an action menu (`⋯`) with **Archive**, **Assign**, **Tag**, **Mark unread**, **Save reply**, **Open profile**.
- [ ] Messages render with their text, images (tappable to full-screen), attachments (file icon + filename + size), reactions row, and timestamp. Internal notes render visually distinct (background tint, "Internal note" label) from public replies.
- [ ] The input bar supports: text entry with mentions (`@`), media attach, saved-reply picker, emoji picker, and a "Send as internal note" toggle. Submit is disabled while text is empty or a send is in flight.
- [ ] Tapping a saved-reply chip inserts the saved-reply text into the input.
- [ ] The post-comments surface opens when the user taps a comment notification or a published post's "View comments" action. It shows: post header (account avatar, name, platform icon, timestamp), post media + caption, threaded comment list with avatars and reactions, and a comment input bar at the bottom.
- [ ] Comment actions (`⋯` per comment): **Reply**, **Like**, **Hide**, **Delete**, **Mark resolved**. Actions hit the same endpoints used today.
- [ ] All copy is loaded from the existing `Localization/Inbox/*.json` files (en, de, el, es, fr, it, pl, zh) — no new hardcoded strings.
- [ ] Empty states (no conversations, no filter matches, no comments) use `CSEmptyState` from the component library with the same illustrations + copy used today.
- [ ] Error states (network failure on list load, send failure, action failure) use `CSErrorState` for full-screen failures and `CSInfoBanner` for non-blocking failures.
- [ ] Loading states use `CSLoadingOverlay` (full screen) or skeleton rows in the list — never a blank screen.
- [ ] All four inbox screens use tokens from `Theme.swift` for color, typography, spacing, and radius — no inlined hex values, no inlined point values, no magic numbers.
- [ ] All four inbox screens use `NavigationStack` and SwiftUI sheet/fullScreenCover presentations — no `UINavigationController` push/pop in the new code paths.
- [ ] The legacy controllers (`InboxViewController`, `FiltersViewController`, `ConversationViewController`, `PlatformPostViewController`) and their `.xib` files are removed from the project once the SwiftUI versions ship. No dead code, no orphan xibs.
- [ ] Existing analytics events that fire today on inbox actions (reply sent, conversation archived, comment posted, etc.) continue to fire with the same names and payloads — verify the rebuild doesn't drop any tracking.
- [ ] No regression in supported behavior: the user can still complete every action available in the current Inbox (reply, attach, react, archive, assign, tag, mark read/unread, mark resolved, save reply, view profile, view post, hide/delete comment, navigate threads).

### Mock-ups
N/A — visual parity is judged against the existing web app inbox + ContentStudio's iOS SwiftUI design language (Theme.swift tokens). Design will provide reference screens if needed during planning.

### Impact on existing data
None. No schema changes, no migration, no new fields. Pure UI rebuild on top of the existing Inbox APIs and local models in `Modals/Inbox/`.

### Impact on other products
- **Backend / API:** no changes — same endpoints, same payloads.
- **Web app:** no changes.
- **Android:** no changes — Android Inbox is out of scope for this epic.
- **Chrome extension:** no changes.

### Dependencies
- **[iOS] Expand the shared SwiftUI component library** must provide `CSListRow`, `CSChip`, `CSBottomSheet`, `CSEmptyState`, `CSErrorState`, `CSInfoBanner`, `CSLoadingOverlay`, `CSAvatar`, `CSBadge`, `CSToolbar`, and `CSSearchField` before this story can be fully completed.
- Existing Inbox API client (`InboxBuilder.swift`, models in `Modals/Inbox/`) is reused as-is.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories) — must look correct on iPhone SE (smallest supported) through iPhone Pro Max and iPad in compact/regular size classes.
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — verify all 8 locales in `Localization/Inbox/`.
- [ ] UI theming support (default + white-label, design library components are being used) — N/A for white-label on mobile (iOS does not support white-label theming); component library usage is in-scope.
- [ ] White-label domains impact review — N/A (iOS does not ship white-label).
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — verified none in **Impact on other products**.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points (today, to be replaced):**
- `Controllers/Nav Menu VCs/Inbox/InboxViewController.swift`
- `Controllers/Nav Menu VCs/Inbox/Filters/FiltersViewController.swift` + `FilterXibs/`, `Social Accounts/`, `Tags/`, `Team Members/`
- `Controllers/Nav Menu VCs/Inbox/Conversation/ConversationViewController.swift` + `ConversationXibs/`
- `Controllers/Nav Menu VCs/Inbox/Post/PlatformPostViewController.swift` + `PostCommentSegmentTblCell`
- `Views/Social Inbox/InboxDataTableViewCell.{swift,xib}`, `Views/Social Inbox/Header/InboxHeader.{swift,xib}`
- `Views/Social InboxAccount/` (sub-screens for account filters)
- `Builder/Inbox/InboxBuilder.swift` (API/dependency wiring — keep, reuse)
- `Modals/Inbox/Inbox.swift`, `Modals/Inbox/InboxDetails.swift` (models — keep, reuse)

**Existing SwiftUI to reuse on this surface:**
- `Views/Comments/PostPreviewCommentsView.swift`, `PostPreviewCommentBubbleView.swift`, `PostPreviewCommentInputBar.swift`, `CommentActionsBottomSheetView.swift`, `CommentImageGridView.swift`, `EmojiPickerView.swift`, `MentionHighlightedText.swift`, `CommentReactionChipView.swift`, `ReactionUsersBottomSheetView.swift` — already SwiftUI, designed for the comments use case, should plug into the post-comments rebuild.

**Suggested module layout:**
- `Views/Inbox/SwiftUI/InboxView.swift`
- `Views/Inbox/SwiftUI/InboxFilterSheet.swift`
- `Views/Inbox/SwiftUI/ConversationView.swift`
- `Views/Inbox/SwiftUI/PostCommentsView.swift`
- `ViewModels/Inbox/*ViewModel.swift` (one per screen, ObservableObject)

**Gotchas:**
- `ConversationViewController.swift` is 1857 lines and `PlatformPostViewController.swift` is 2844 lines — both have significant business logic (message pagination, optimistic send, websocket updates, attachment handling, reaction state) that must be extracted into view models, not lost.
- Keyboard handling: the conversation/comment input bars currently use UIKit's `keyboardWillShow` notifications. SwiftUI's `@FocusState` + `.keyboardToolbar` is the right replacement — don't reintroduce manual frame math.
- Existing analytics calls are scattered across the old VCs — audit before deleting.
- `InboxHeader.xib` may be used in other places (e.g. user-listing screens) — verify before deletion.
- Bridging: if any other UIKit screen pushes onto the Inbox flow, wrap the new SwiftUI views with the generalized `UIKitSwiftUIBridge` from Story 4.

---

## Story 2 — [iOS] Rebuild the Settings surface (profile, password, about, help, knowledge base, side menu) in SwiftUI

### Description
As an iOS user managing my account, I want the Settings, Profile, Change Password, About, Help, Help Desk, Knowledge Base, and side-menu screens to look consistent with the rest of the app, so that navigation between them feels seamless and the visual language matches the screens that have already moved to SwiftUI.

This is a UI rebuild — same fields, same endpoints, same destinations. Every screen reachable from the side menu's "Settings" / "About" / "Help" sections is migrated from UIKit + `.xib` to SwiftUI. The side menu itself is rebuilt to match.

### Workflow
1. The user taps the hamburger / side-menu icon. A SwiftUI side menu slides in showing: user avatar + name + email + plan badge at the top, workspace switcher row, and menu sections (Account, Subscription, Help, Other) with rows for Profile, Change Password, Notifications, Subscription, Knowledge Base, Help Desk, About Us, Privacy Policy, Terms of Use, Log out. Version info anchors the bottom.
2. The user taps **Profile**. The Profile screen opens — a SwiftUI form with avatar (tap to change), name, email (read-only with explanation), timezone picker, language picker, and a **Save** CTA. Validation runs as the user edits; Save is disabled until edits are valid; on save, the user sees a success toast and returns to the side menu.
3. The user taps **Change Password**. A SwiftUI form opens with current password, new password, confirm new password, and a **Update password** CTA. Inline validation shows password-strength rules. On success, the user sees a success toast and is returned to the side menu.
4. The user taps **About Us**, **Help**, or **Help Desk**. Each opens a SwiftUI list screen with the same items shown today (links, contact rows, version info, social-media row). Tapping a row either opens a `CSWebView`, opens mail/phone, or pushes to a child SwiftUI detail screen.
5. The user taps **Knowledge Base**, **Privacy Policy**, or **Terms of Use**. The destination opens a SwiftUI screen with a `CSWebView` body, a SwiftUI title bar with back button and share action, and SwiftUI loading + error states.
6. Tapping back returns the user up one level with state preserved.

### Acceptance criteria
- [ ] The side menu opens as a SwiftUI view — no `MenuVC.xib`, no `MenuTblCell.xib`, no `SwitchWorkspaceTblCell.xib`, no `LogoutMenuTblCell.xib`, no `MenuSpaceTblCell.xib`, no `VersionInfoTblCell.xib` in the new code path.
- [ ] Side menu header shows: user avatar (with initials fallback), name, email, current plan badge, and a tap target that opens Profile.
- [ ] Side menu shows the workspace switcher row with: current workspace name, current workspace avatar, role badge, and a chevron that opens the Workspace screen.
- [ ] Side menu sections render in order: **Account** (Profile, Change Password), **Subscription** (Subscription Management — links to existing SwiftUI `SubscriptionManagementView`), **Help** (Knowledge Base, Help Desk, About Us), **Legal** (Privacy Policy, Terms of Use), and the **Log out** row at the bottom above version info.
- [ ] Log out row triggers the existing logout confirmation alert + flow — same behavior as today.
- [ ] Profile screen is a SwiftUI form with: avatar (tap → image picker → upload to existing endpoint), full name (editable, required), email (read-only with a "Contact support to change" subtext), timezone (picker), language (picker — same 8 locales), and a **Save** button.
- [ ] Profile validation: name cannot be empty; Save is disabled when there are no edits or validation fails.
- [ ] Profile Save calls the existing endpoint (no new API), shows a success toast, and pops back to the side menu.
- [ ] Change Password screen is a SwiftUI form with: current password (secure entry, show/hide toggle), new password (secure entry, show/hide toggle, with password-strength meter + rule list — same rules as today), confirm new password (must match), and an **Update password** CTA.
- [ ] Change Password validation: new and confirm must match; new must meet the existing strength rules; CTA is disabled otherwise.
- [ ] Change Password submit calls the existing endpoint, shows success toast, pops back.
- [ ] About Us, Help, and Help Desk screens are all SwiftUI list screens using `CSListRow` and `CSSectionHeader` from the component library. The legacy `SimilarMenuViewController.xib` and `AboutUsTableViewCell.xib` are removed.
- [ ] Each list row supports the same destination types as today: external URL (opens in `CSWebView`), email (opens Mail app), phone (opens Phone app), in-app push (e.g. to a detail screen), and copy-to-clipboard.
- [ ] Knowledge Base, Privacy Policy, and Terms of Use screens each use `CSWebView` for the body, with: SwiftUI navigation bar (back button, page title, share action), SwiftUI loading overlay during initial page load, and SwiftUI error state on failure to load (with a "Try again" CTA).
- [ ] All copy is loaded from the existing `Settings/settings_*.json` and other existing locale files — no new hardcoded user-facing strings.
- [ ] Empty / error / loading states use the component library (`CSEmptyState`, `CSErrorState`, `CSLoadingOverlay`).
- [ ] All Settings screens use `Theme.swift` tokens — no inlined hex, no magic numbers.
- [ ] The legacy `MenuVC.swift`, `ProfileViewController.swift`, `ChangePasswordViewController.swift`, `SimilarMenuViewController.swift`, and all related xibs are removed once the SwiftUI versions ship.
- [ ] Existing analytics events that fire on Settings actions (profile saved, password changed, logout, etc.) continue to fire with the same names + payloads.
- [ ] No regression in supported behavior: every Settings/About/Help destination available today is available in the rebuild.

### Mock-ups
N/A — visual parity is judged against the existing SwiftUI surfaces (e.g. `SubscriptionManagementView`, `DeleteAccountFlowView`) and Theme.swift tokens.

### Impact on existing data
None. No schema changes. Same profile/password endpoints, same Settings/Help/About content sources.

### Impact on other products
- **Backend / API:** no changes.
- **Web app, Android, Chrome extension:** no changes.

### Dependencies
- **[iOS] Expand the shared SwiftUI component library** must provide `CSListRow`, `CSSectionHeader`, `CSWebView`, `CSAvatar`, `CSBadge`, `CSEmptyState`, `CSErrorState`, `CSLoadingOverlay`, and `CSToolbar` before this story can be fully completed.
- **[iOS] Rebuild the Workspace screen** — the workspace switcher row in the side menu navigates into Workspace; the Workspace story owns the destination screen.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories) — verify on iPhone SE through Pro Max and iPad sizes.
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — verify all 8 locales in `Settings/` and elsewhere.
- [ ] UI theming support (default + white-label, design library components are being used) — N/A for white-label on iOS; component library usage in-scope.
- [ ] White-label domains impact review — N/A (iOS does not ship white-label).
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — verified none.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points (today, to be replaced):**
- `Controllers/Menu/MenuVC.swift` + `cell/` xibs + `Model/MenuTitleModel.swift` + `ViewModel/MenuViewModel.swift`
- `Controllers/Nav Menu VCs/Setting/ProfileViewController.swift`
- `Controllers/Nav Menu VCs/Setting/ChangePasswordViewController.swift`
- `Controllers/Nav Menu VCs/Setting - About - Help/SimilarMenuViewController.swift`
- `Controllers/Nav Menu VCs/KB - Privacy Policy - ToU/Web View/`
- `Views/About Us/AboutUsTableViewCell.{swift,xib}`

**Existing SwiftUI to reuse:**
- `Views/Paywall/SubscriptionManagementView.swift` — link from the Subscription row in the side menu.
- `Controllers/Account/DeleteAccountFlowView.swift` and step views — link from Profile or a dedicated row.

**Suggested module layout:**
- `Views/SideMenu/SwiftUI/SideMenuView.swift`
- `Views/Settings/SwiftUI/ProfileView.swift`
- `Views/Settings/SwiftUI/ChangePasswordView.swift`
- `Views/Settings/SwiftUI/StaticListView.swift` (single SwiftUI replacement for `SimilarMenuViewController`, driven by the same config model used today)
- `Views/Settings/SwiftUI/WebContentView.swift` (wraps `CSWebView` + nav + states for KB/Privacy/ToU)

**Gotchas:**
- `SimilarMenuViewController` is reused for several distinct screens via configuration — the SwiftUI replacement must accept the same config so all consumers swap at once.
- `MenuVC.swift` (856 lines) handles routing/deep-linking to many destinations — extract the routing logic into a coordinator type, not view-model.
- Avatar upload uses the same endpoint as profile save — preserve the upload-progress UI.
- `WKWebView` SSO cookies: the KB/Help-Desk web view passes auth cookies — the new `CSWebView` wrapper must preserve cookie + user-agent setup, otherwise users land on a login wall.

---

## Story 3 — [iOS] Rebuild the Workspace screen in SwiftUI

### Description
As an iOS user who works across multiple ContentStudio workspaces, I want the Workspace list/picker and the locked-workspace screen to look consistent with the rest of the app, so that switching workspaces feels native and matches the visual language of the screens I already use.

This is a UI rebuild — same data, same endpoints, same actions. The Workspace screen and Locked Workspace screen are migrated from UIKit + `.xib` to SwiftUI, sharing the component library from Story 4.

### Workflow
1. The user opens the Workspace screen from the side menu (or from the workspace switcher header).
2. The user sees a SwiftUI list of workspaces. Each row shows: workspace avatar, workspace name, the user's role in that workspace (Owner / Admin / Member / etc.), members count, current-workspace indicator (checkmark), and a chevron / action menu.
3. The user can tap a row to switch to that workspace. A confirmation appears ("Switch to workspace X?"); on confirm the app re-bootstraps with the new workspace context and lands on the user's last-used screen there.
4. The user can long-press a row (or tap the action menu) to leave a workspace, request to leave, or copy the workspace ID — same actions as today.
5. If the user has access to a locked workspace (plan / payment / suspension), tapping the row routes to the Locked Workspace screen, which shows: workspace name, lock reason, illustration, and an upgrade/contact-owner CTA.
6. A **Create workspace** button at the bottom (or top — match current placement) routes to the existing create-workspace flow.

### Acceptance criteria
- [ ] The Workspace screen is a SwiftUI view — no `WorkspaceViewController.xib`, no `WorkspaceTableViewCell.xib` in the new code path.
- [ ] Each workspace row uses `CSListRow` and shows: avatar (`CSAvatar` with initials fallback), workspace name, role badge (`CSBadge`), members count subtext, current-workspace checkmark, and an action menu (`⋯`).
- [ ] The current workspace is visually distinct (highlighted background, checkmark) and not tappable for "switch" — its action menu shows **Settings** instead of **Switch**.
- [ ] Tapping a non-current workspace row shows the existing switch-confirmation alert ("Switch to {name}?"). Confirming triggers the existing switch-workspace flow.
- [ ] Action menu (`⋯`) for a non-current workspace shows: **Switch to this workspace**, **Leave workspace** (Owners see **Settings** instead of **Leave**), **Copy workspace ID**.
- [ ] Locked workspaces render with a lock badge on the row. Tapping a locked workspace opens the Locked Workspace screen.
- [ ] The Locked Workspace screen is SwiftUI: workspace name, illustration (existing asset), lock-reason copy (from existing strings), and a primary CTA whose action matches today's behavior (Upgrade / Contact owner / Re-activate).
- [ ] A **Create workspace** entry (button or row) routes to the existing create-workspace flow with no UI change to that flow.
- [ ] Search/filter bar at the top (if present today) is included in the SwiftUI rebuild — uses `CSSearchField`.
- [ ] Empty / loading / error states use `CSEmptyState`, `CSLoadingOverlay`, `CSErrorState` from the component library.
- [ ] All copy is loaded from existing locale files — no new hardcoded strings.
- [ ] All screens use `Theme.swift` tokens — no inlined hex, no magic numbers.
- [ ] The legacy `WorkspaceViewController.swift` and `LockedWorkspaceViewController.swift` and their xibs/cells are removed.
- [ ] Existing analytics events on workspace switch / leave / locked-view continue to fire with the same names + payloads.
- [ ] No regression in supported behavior — every action available on the current Workspace screen is available in the rebuild.

### Mock-ups
N/A — visual parity is judged against existing SwiftUI surfaces and Theme.swift.

### Impact on existing data
None. No schema changes, no migrations.

### Impact on other products
- **Backend / API, web, Android, Chrome extension:** no changes.

### Dependencies
- **[iOS] Expand the shared SwiftUI component library** must provide `CSListRow`, `CSAvatar`, `CSBadge`, `CSSearchField`, `CSEmptyState`, `CSErrorState`, `CSLoadingOverlay`, and `CSToolbar`.
- **[iOS] Rebuild the Settings surface** owns the side-menu entry into Workspace.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — verify on iPhone SE through Pro Max and iPad.
- [ ] Multilingual support — verify all 8 locales.
- [ ] UI theming support — N/A for white-label on iOS; component library usage in-scope.
- [ ] White-label domains impact review — N/A.
- [ ] Cross-product impact assessment — verified none.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points (today, to be replaced):**
- `Controllers/Workspace/WorkspaceViewController.swift`
- `Controllers/Workspace/LockedWorkspaceViewController.swift`
- `Views/Workspace/Cell/WorkspaceTableViewCell.{swift,xib}`
- `Modals/Workspace/` (models — keep, reuse)

**Suggested module layout:**
- `Views/Workspace/SwiftUI/WorkspaceListView.swift`
- `Views/Workspace/SwiftUI/LockedWorkspaceView.swift`
- `Views/Workspace/SwiftUI/WorkspaceRow.swift`
- `ViewModels/Workspace/WorkspaceListViewModel.swift`

**Gotchas:**
- Workspace switching involves re-bootstrapping the app (auth context, feature flags, push-notification registration). Preserve the existing bootstrap call chain — only the trigger UI changes.
- The locked-workspace state has multiple sub-states (trial expired, payment failed, suspended, owner-only-action-required) — the existing controller routes copy by state. Preserve the same state branching in the SwiftUI rebuild.

---

## Story 4 — [iOS] Expand the shared SwiftUI component library (tokens, primitives, preview gallery)

### Description
As an iOS engineer building screens in SwiftUI, I want a documented set of shared components and design tokens so that I can compose new screens quickly and every screen looks consistent across the app. As a designer, I want a single place to see every component the app supports.

This story is the foundation for the other three stories in this epic. It extends the existing `SwiftUI/` directory into a real component library: enforced tokens, a set of primitive components used by Inbox / Settings / Workspace, a debug preview gallery, and a README that explains the rules.

### Workflow
1. An engineer opens the new `SwiftUI/README.md` and reads the rules: build new screens in SwiftUI, use `Theme.swift` tokens for colors / spacing / typography / radii, use the primitives in `SwiftUI/Components/`, bridge to UIKit only via `UIKitSwiftUIBridge`.
2. The engineer opens the new debug-only `ComponentGalleryView` (gated behind a debug flag — visible only in internal builds) and sees every primitive in a scrollable list with `#Preview` blocks demonstrating each state (default, hover/pressed, disabled, loading, error, etc.).
3. The engineer imports a primitive (e.g. `CSListRow`, `CSBottomSheet`) into their new screen, configures via SwiftUI initializers / modifiers, and ships.
4. The engineer adds a new primitive: writes the component in `SwiftUI/Components/`, adds a `#Preview` block, registers it in `ComponentGalleryView`, updates `README.md`.

### Acceptance criteria
- [ ] `SwiftUI/Theme.swift` is the single source of truth for: colors (semantic — `primary`, `surface`, `background`, `text-primary`, `text-secondary`, etc.), typography (named styles — `title1`, `body`, `caption`, etc.), spacing (`s1`–`s8` or named scale), corner radii, elevations, and animation durations.
- [ ] `Theme.swift` exposes SwiftUI `ViewModifier` / `Font` / `Color` extensions so callers write `.csTitleStyle()`, `Color.csPrimary`, `Font.csBody` — not `.font(.system(size: 14, weight: .semibold))` or `Color(hex: "#...")`.
- [ ] All primitives below ship in `SwiftUI/Components/`, each in its own file:
  - `CSListRow.swift` — leading icon/avatar, title, subtitle, trailing chevron / value / accessory, optional badge, tap action, swipe-action support.
  - `CSSectionHeader.swift` — section title, optional supporting text, optional trailing action.
  - `CSSearchField.swift` — bound text, placeholder, clear button, focus state.
  - `CSSegmentedControl.swift` — array of options, bound selection.
  - `CSEmptyState.swift` — illustration, headline, body, optional CTA.
  - `CSErrorState.swift` — illustration, headline, body, retry CTA.
  - `CSBottomSheet.swift` — presentation modifier with grabber, dismiss-on-drag, half-height + full-height detents.
  - `CSAvatar.swift` — image URL, initials fallback, size variants, badge overlay.
  - `CSBadge.swift` — color + label variants for status / role / count.
  - `CSChip.swift` — label, optional leading icon, optional trailing remove (`×`) action, selected state.
  - `CSToolbar.swift` — title, leading back button, trailing actions slot.
  - `CSWebView.swift` — `WKWebView` wrapper exposing url, cookie/header config, loading + error callbacks; preserves auth-cookie passing used today.
- [ ] Each primitive has at least one `#Preview` block showing default + one variant state.
- [ ] A `ComponentGalleryView` SwiftUI screen lists every primitive in a scrollable navigation list. Tapping into a primitive shows its preview variants. The gallery is reachable from a hidden debug entry (e.g. long-press version info in the side menu) and is **stripped from release builds** via `#if DEBUG`.
- [ ] `Views/Planner/SwiftUI/UIKitSwiftUIBridge.swift` is moved to `SwiftUI/Helpers/UIKitSwiftUIBridge.swift` and made generic enough that Inbox, Settings, and Workspace can adopt it without modification.
- [ ] `SwiftUI/README.md` is created at `contentstudio-ios-v2/ContentStudio/SwiftUI/README.md`, covering: token system, every primitive (name + when to use), the bridge pattern, where to add new components, how to add a `#Preview`.
- [ ] An ADR-style note (or section in `README.md`) documents the rule: **new screens are SwiftUI by default; UIKit is only acceptable for screens this epic does not migrate**.
- [ ] No primitive references the inbox/settings/workspace domain — primitives are domain-agnostic.
- [ ] Every primitive uses `Theme.swift` tokens — no inlined colors, fonts, or magic numbers inside primitives.

### Mock-ups
N/A — components are evaluated visually against the Theme.swift tokens and the existing SwiftUI surfaces (e.g. `SubscriptionManagementView`, Comments views, Planner SwiftUI views).

### Impact on existing data
None.

### Impact on other products
- **Backend / API, web, Android, Chrome extension:** no changes.

### Dependencies
- None — this is the foundation story.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — every primitive renders correctly on iPhone SE through Pro Max and iPad in compact/regular size classes.
- [ ] Multilingual support — primitives use SwiftUI text rendering that supports localization out of the box. No hardcoded English copy in any primitive.
- [ ] UI theming support — N/A for white-label on iOS; tokens defined via Theme.swift.
- [ ] White-label domains impact review — N/A.
- [ ] Cross-product impact assessment — verified none.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Existing SwiftUI primitives to extend or absorb:**
- `SwiftUI/Theme.swift`
- `SwiftUI/Buttons.swift`, `SwiftUI/CommonButtons.swift`, `SwiftUI/ButtonAndLabelView.swift` — consolidate into `SwiftUI/Components/CSButton.swift` (consider naming, keep the existing variants)
- `SwiftUI/Textfields.swift`, `SwiftUI/Texts.swift` — consolidate / split into `CSTextField`, `CSText` (or keep modifier-only)
- `SwiftUI/UIComponents.swift` — audit, move each piece into a typed `Components/` file or delete
- `SwiftUI/InfoBannerView.swift` → `Components/CSInfoBanner.swift`
- `SwiftUI/LoadingOverlay.swift` → `Components/CSLoadingOverlay.swift`
- `SwiftUI/SwiftUIAlertExtensions.swift` → `Helpers/SwiftUIAlertExtensions.swift`
- `SwiftUI/Helpers/SwiftUINavigationHelper.swift` — keep, extend

**Suggested module layout:**
- `SwiftUI/Theme.swift` (tokens)
- `SwiftUI/Components/` (one file per primitive, named `CS<Name>.swift`)
- `SwiftUI/Helpers/` (UIKitSwiftUIBridge, navigation helpers, alert extensions)
- `SwiftUI/Gallery/ComponentGalleryView.swift` (debug-only)
- `SwiftUI/README.md`

**Gotchas:**
- `UIKitSwiftUIBridge.swift` currently lives under `Views/Planner/SwiftUI/` — moving it is fine but check all imports.
- The Planner SwiftUI views already use ad-hoc theming (look at `PlannerPostCardView.swift`, `PostPreviewView.swift`) — they should eventually adopt the new tokens but that's a follow-up, not in this story.
- Naming: web app uses `@contentstudio/ui` with `Cst*` prefix — iOS adopting `CS*` (e.g. `CSListRow`) keeps the family without colliding.
- Avoid over-engineering: do not introduce a separate package / Swift module — keep components in the app target unless the team explicitly wants a Swift package.
