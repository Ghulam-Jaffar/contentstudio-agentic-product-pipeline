# Research: Social Accounts Management — Android (iOS Parity)

## iOS Feature Branch — What's Built (`feature/social-channel`)

**`Views/SocialChannels/SocialChannelsView.swift`** — List screen (SwiftUI)
- Fetches accounts via `POST /fetchSocialAccounts`
- Account cards: avatar + platform badge, name, type, status chip (Valid / Expired)
- Platform filter bottom sheet (multi-select: FB, IG, LinkedIn, TikTok, YouTube, etc.)
- Status filter bottom sheet (Valid / Expired)
- Search by account name
- Per-account 3-dot button → options bottom sheet
- Empty state: "No Social Accounts Connected" + "Connect Social Account" CTA button
- FAB (+) → opens `ConnectSocialAccountsView` (reuses onboarding OAuth flow)
- Delete from options → `RemoveAccountConfirmationView` → `POST /removeIntegration`
- Loading state, error state, pull-to-refresh

**`Views/SocialChannels/BottomSheets/AccountOptionsViewController.swift`** — Options bottom sheet
Options (in order):
1. Account Details (→ navigates to `AccountDetailsView`)
2. ~~Queue Schedule~~ (commented out / removed)
3. Shuffle Posts
4. Default Location (→ opens `DefaultLocationPopupView`)
5. Delete Account (red, trash icon)

**`Views/SocialChannels/AccountDetailsView.swift`** — Account details screen (SwiftUI)
- Account card: avatar, platform badge, name, type
- Reconnect button → `ReconnectManager` → `POST /fetchSocialAuthorizationLinks` → in-app browser OAuth
- Delete button → `RemoveAccountConfirmationView` → `POST /removeIntegration`
- Location row (FB + IG only) → `DefaultLocationPopupView` → `POST account/defaultLocation`
- Connected Via row (name of team member who added it)
- Token Status row: "Token Valid" (green) / "Token Expired" (red)
- Devices row (FB groups/profiles + Instagram only): count of enabled/total devices

**Data model** (`Views/SocialChannels/Models/SocialAccount.swift`):
Platforms: Facebook, Instagram, LinkedIn, TikTok, YouTube, Twitter, Bluesky, Pinterest, Threads, GMB, Tumblr
Status: `Valid` / `Expired`
Fields: id, name, platformName, platformIdentifier, type, platform, status, profileImageURL, isVerified, location (LocationInfo), addedBy, validity, disabledDevices

**Supporting files:**
- `ConnectSocialAccountsView.swift` — connect new accounts platform picker
- `DefaultLocationPopupView.swift` — location search + save popup
- `RemoveAccountConfirmationView.swift` — delete confirmation dialog
- `PlatformFilterViewController.swift` — platform filter bottom sheet
- `StatusFilterViewController.swift` — Valid/Expired filter bottom sheet
- `FetchPlatformsListEndpoint.swift` / `FetchPlatformsListResponse.swift` — available platforms API

---

## Backend APIs (all exist)

| Endpoint | Purpose | Used by |
|---|---|---|
| `POST /fetchSocialAccounts` | List all connected accounts | Already in Android ServiceManager |
| `POST /fetchSocialAuthorizationLinks` | OAuth URLs for connect/reconnect | Need to add to Android |
| `POST /removeIntegration` | Delete/disconnect account | Need to add to Android |
| `POST /account/defaultLocation` | Save default posting location | Need to add to Android |
| `POST /fetchPlatformsList` | Available platforms for connect | Already in Android ServiceManager |

Source: `contentstudio-backend/routes/web/integrations.php`

---

## Android Current State

- `ServiceManager.java:57` — `getAllUserAccounts()` via `POST /fetchSocialAccounts` ✅
- `ServiceManager.java:62` — `fetchPlatformsList()` ✅
- No social accounts management activity, no account details, no connect-from-settings flow

---

## Story Split

1. **[Android] Social accounts list screen** — list + options sheet + connect/reconnect/delete from list
2. **[Android] Account details screen for social accounts** — details view + reconnect + delete + location

---

## Android Files Involved

**New:**
- `Workspace/SocialAccounts/SocialAccountsActivity.java` + layout
- `Workspace/SocialAccounts/SocialAccountAdapter.java`
- `Workspace/SocialAccounts/SocialAccountModel.java`
- `Workspace/SocialAccounts/AccountOptionsBottomSheet.java`
- `Workspace/SocialAccounts/AccountDetailsActivity.java` + layout
- `Workspace/SocialAccounts/DefaultLocationBottomSheet.java`
- `Workspace/SocialAccounts/RemoveAccountConfirmationDialog.java`
- `Workspace/SocialAccounts/PlatformFilterBottomSheet.java`
- `Workspace/SocialAccounts/StatusFilterBottomSheet.java`

**Modified:**
- `Network/ServiceManager.java` — add `fetchSocialAuthorizationLinks`, `removeIntegration`, `saveDefaultLocation`
- `Setting/SettingActivity.java` — add "Social Accounts" entry in settings list
