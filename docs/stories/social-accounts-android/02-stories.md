# Stories: Social Accounts Management — Android

---

## Story 1: [Android] Social accounts list screen

### Description:

As a ContentStudio user on Android, I want to see all my connected social accounts for the current workspace and manage them (connect new, reconnect expired, delete) from a dedicated screen in the app, so I can handle my account connections without opening the web app.

This is the Android parity of the iOS `SocialChannelsView.swift` (feature/social-channel branch).

---

### Workflow:

1. User opens the Android app and goes to **Settings**.
2. User taps **"Social Accounts"** (new entry in `SettingActivity`).
3. App calls `POST /fetchSocialAccounts` with the current `workspace_id` and displays the accounts list.
4. Each account card shows:
   - Profile avatar with platform logo badge (bottom-right)
   - Account name (e.g., "TrendVibe Fashion")
   - Account type (e.g., "Page", "Profile", "Group")
   - Status chip: **"Valid"** (green) or **"Expired"** (red/orange)
   - 3-dot menu icon (right side)
5. User can **filter by platform**: taps a filter chip row at the top (or a filter button) to open a Platform filter bottom sheet (multi-select: Facebook, Instagram, LinkedIn, TikTok, YouTube, Twitter, Pinterest, etc.). Selected platforms filter the list.
6. User can **filter by status**: Status filter bottom sheet with two options: Valid, Expired.
7. User can **search**: taps the search icon to expand a search bar; types to filter accounts by name in real time.
8. User taps the **3-dot icon** on any account → bottom sheet slides up with these options:
   - **Account Details** → navigates to Account Details screen
   - **Shuffle Posts** → triggers shuffle confirmation dialog, calls shuffle API
   - **Default Location** → opens location search bottom sheet (FB/IG accounts only; hidden for others)
   - **Delete Account** (red) → opens confirmation dialog: "Remove [Account Name]?" with "Cancel" / "Remove" CTAs → on confirm, calls `POST /removeIntegration` → account removed from list
9. **Empty state** (no accounts connected):
   - Illustration (social platform icons)
   - **Headline:** "No Social Accounts Connected"
   - **Subtext:** "Connect your social media accounts to start scheduling and managing posts from ContentStudio."
   - **CTA button:** "Connect Social Account"
10. **FAB (+)** button always visible (when accounts exist) → opens Connect Social Account screen (platform picker → OAuth in-app browser → account added → user returned to list with updated data).
11. Pull-to-refresh refreshes the accounts list.
12. **Loading state:** progress indicator while fetching accounts.
13. **Error state:** "Unable to load accounts. Pull down to refresh." with retry.

---

### Acceptance criteria:

- [ ] "Social Accounts" entry is visible in `SettingActivity` settings list
- [ ] Tapping "Social Accounts" navigates to `SocialAccountsActivity`
- [ ] App calls `POST /fetchSocialAccounts` on screen load with current `workspace_id`
- [ ] Each account card shows avatar (with platform badge), name, type, and Valid/Expired status chip
- [ ] Platform filter bottom sheet allows multi-select; filtered list updates on selection
- [ ] Status filter bottom sheet shows Valid / Expired options; list filters accordingly
- [ ] Search bar filters accounts by name in real time
- [ ] Tapping 3-dot on any account opens the options bottom sheet with: Account Details, Shuffle Posts, Default Location (FB/IG only), Delete Account
- [ ] Delete Account: shows confirmation dialog "Remove [Account Name]?" → on confirm calls `POST /removeIntegration` → account removed from list; success snackbar: "Account removed successfully."
- [ ] Delete Account: on API failure, shows snackbar: "Failed to remove account. Please try again."
- [ ] Default Location: opens location search bottom sheet (only for Facebook and Instagram accounts)
- [ ] Empty state shows "No Social Accounts Connected" headline with "Connect Social Account" CTA
- [ ] FAB (+) opens connect flow (platform picker → OAuth via in-app browser)
- [ ] Pull-to-refresh re-fetches and updates the list
- [ ] Loading state shown while fetching; error state shown on network failure

---

### Mock-ups:

N/A — reference iOS `feature/social-channel` branch → `SocialChannelsView.swift` for visual design parity.

---

### Impact on existing data:

No data model changes. `POST /fetchSocialAccounts` and `POST /removeIntegration` already exist in the backend. Android `ServiceManager.java` needs 1 new endpoint: `POST /removeIntegration`.

---

### Impact on other products:

- **Web:** No impact.
- **iOS:** Separate feature, already in progress on `feature/social-channel` branch.
- **Chrome extension:** No social account management in Chrome extension.

---

### Dependencies:

None. All backend APIs are live.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, native Android screen
- [ ] Multilingual support — all UI strings (screen title "Social Accounts", status chips "Valid"/"Expired", empty state copy, options labels, confirmation dialog copy, snackbar messages) must use existing `L.menu()` / string resource localization system
- [ ] UI theming support — N/A, Android app does not use web theming system
- [ ] White-label domains impact review — N/A for native Android
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Story 2: [Android] Account details screen for social accounts

### Description:

As a ContentStudio user on Android, I want to view the details of a connected social account — including its connection status, who added it, and linked devices — and be able to reconnect or delete it from the same screen, so I can manage individual accounts without going to the web app.

This is the Android parity of the iOS `AccountDetailsView.swift` (feature/social-channel branch).

---

### Workflow:

1. From the Social Accounts list, user taps **"Account Details"** in the 3-dot options sheet.
2. App navigates to `AccountDetailsActivity` and displays the account's details.
3. Screen layout:
   - **Header:** back arrow + "Account Details" title
   - **Account card:**
     - Profile avatar with platform badge (bottom-right)
     - Account name + account type (e.g., "TrendVibe Fashion · Page")
     - **Reconnect** button (orange, with refresh icon) — always visible
     - **Delete** button (red trash icon) — always visible
   - **Info section (card rows):**
     - **Location** — visible for Facebook and Instagram accounts only. Shows current default location name (tappable) or "Add location" (tappable, blue) → opens location search bottom sheet → `POST account/defaultLocation`
     - **Connected Via** — shows "ContentStudio" + name of team member who connected the account (from `addedBy` field)
     - **Token Status** — "Token Valid" chip (green) or "Token Expired" chip (red)
     - **Devices** — visible for Facebook (Profile/Group) and Instagram only. Shows "X/Y device(s)" count (blue, tappable) or "No devices" (grey). Count comes from `DeviceManager`.

4. **Reconnect flow:**
   - User taps "Reconnect"
   - App calls `POST /fetchSocialAuthorizationLinks` to get the OAuth URL for that platform
   - Opens URL in in-app browser (Chrome Custom Tab or WebView)
   - On OAuth completion, browser closes and app refreshes account status
   - Success snackbar: "Account reconnected successfully."
   - Error snackbar: "Reconnection failed. Please try again."

5. **Delete flow:**
   - User taps the trash icon
   - Confirmation dialog appears: **"Remove account?"** with subtext: "This will disconnect [Account Name] from ContentStudio. You can reconnect it at any time." CTAs: **"Cancel"** / **"Remove"** (red)
   - On confirm: calls `POST /removeIntegration` with `type` and platform identifier
   - On success: navigates back to Social Accounts list; list refreshes; snackbar: "Account removed successfully."
   - On failure: snackbar: "Failed to remove account. Please try again."

6. **Location flow** (Facebook/Instagram only):
   - User taps the Location row
   - Bottom sheet opens: search bar with placeholder "Search by location"
   - User types → results appear → user taps a result
   - App calls `POST account/defaultLocation` to save
   - Bottom sheet closes; Location row updates with new location name
   - Success snackbar: "Default location saved."

---

### Acceptance criteria:

- [ ] Navigating from 3-dot options → Account Details opens `AccountDetailsActivity` with the correct account data
- [ ] Account card shows avatar with platform badge, name, type, Reconnect button, Delete (trash) button
- [ ] Location row is visible only for Facebook and Instagram accounts; hidden for all others
- [ ] Location row shows existing default location name if set; shows "Add location" (blue) if not set
- [ ] Tapping Location row opens location search bottom sheet; selecting a location calls `POST account/defaultLocation`; success snackbar "Default location saved."
- [ ] Connected Via row shows team member name who added the account (from API `addedBy` field)
- [ ] Token Status row shows "Token Valid" (green chip) or "Token Expired" (red chip) based on `validity` field
- [ ] Devices row is visible only for Facebook Profile/Group and Instagram accounts; hidden for Facebook Pages and all other platforms
- [ ] Tapping Reconnect calls `POST /fetchSocialAuthorizationLinks`; opens OAuth URL in Chrome Custom Tab; on completion snackbar: "Account reconnected successfully."
- [ ] Tapping trash icon opens confirmation dialog with account name; tapping "Remove" calls `POST /removeIntegration`; on success navigates back to list and shows snackbar "Account removed successfully."
- [ ] On delete failure, shows snackbar: "Failed to remove account. Please try again."
- [ ] On reconnect failure, shows snackbar: "Reconnection failed. Please try again."
- [ ] Loading overlay shown during reconnect and delete API calls
- [ ] `ServiceManager.java` includes `fetchSocialAuthorizationLinks` and `removeIntegration` Retrofit endpoints

---

### Mock-ups:

N/A — reference iOS `feature/social-channel` branch → `AccountDetailsView.swift` for visual design parity.

---

### Impact on existing data:

No schema changes. Uses existing `POST /fetchSocialAuthorizationLinks` and `POST /removeIntegration` backend routes (`contentstudio-backend/routes/web/integrations.php:134,157`).

Android `ServiceManager.java` needs 2 new Retrofit endpoints:
- `POST /fetchSocialAuthorizationLinks`
- `POST /removeIntegration`

And optionally `POST account/defaultLocation` if not already defined.

---

### Impact on other products:

- **Web:** No impact.
- **iOS:** Separate feature, already in progress on `feature/social-channel` branch.
- **Chrome extension:** No account management in Chrome extension.

---

### Dependencies:

Depends on: **[Android] Social accounts list screen** (provides the entry navigation and the account data object passed to this screen).

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, native Android screen
- [ ] Multilingual support — all UI strings (screen title "Account Details", row labels "Location"/"Connected Via"/"Token Status"/"Devices", button labels "Reconnect"/"Remove", dialog copy, snackbar messages) must use the existing localization system
- [ ] UI theming support — N/A, Android app does not use web theming system
- [ ] White-label domains impact review — N/A for native Android
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
