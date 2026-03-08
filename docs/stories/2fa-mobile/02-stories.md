# Stories: 2FA + Trust This Device — iOS & Android

---

## Story 1: [iOS] Two-factor authentication setup and trust this device

### Description:

As a ContentStudio user on iOS, I want to enable or disable two-factor authentication from the app's settings, and when signing in with 2FA enabled I want the option to trust my device so I can skip 2FA on future logins from the same device.

This story covers two related gaps in the current iOS app:
1. **Trust this device** — `TwoFactoreView.swift` already handles 2FA verification at login but does not pass `trust_device` or `trust_device_name` to the API.
2. **2FA Settings** — there is no screen to enable or disable 2FA from the iOS app at all.

Web reference: `contentstudio-frontend/src/modules/account/views/TwoFactorVerification.vue` and `contentstudio-frontend/src/modules/setting/components/TwoFactorAuth.vue`.

---

### Workflow:

#### Part A — Trust This Device (on the existing 2FA verification screen)

1. User signs in with email and password on the iOS app.
2. If 2FA is enabled on their account, the app navigates to the existing `TwoFactoreView.swift` screen.
3. Below the code input field, the user sees a toggle row:
   - **Toggle label:** "Trust this device"
   - **Info icon (ℹ)** next to the label — tapping it shows a tooltip/popup:
     > _"If you trust this device, you won't need to enter a 2FA code the next time you log in from this device. If you log in from a different device, you'll need to verify again."_
4. When the user turns on the "Trust this device" toggle, a text field appears below:
   - **Label:** "Device name (optional)"
   - **Placeholder:** "e.g. My iPhone, Work Phone"
5. The "Trust this device" toggle is hidden when the user is in backup code mode.
6. User enters their 6-digit code and taps **Submit**.
7. The app sends `trust_device: true` (and `trust_device_name` if entered) along with the code and `user_info` to `POST /2fa/validator/google`.
8. On success: user is logged in and navigates to the main app. If trust was enabled, future logins from this device skip the 2FA screen.

#### Part B — 2FA Setup in Settings

1. User opens the iOS app and goes to **Settings** → **Profile** (or the security/account settings section).
2. User sees a new row: **"Two-Factor Authentication"** — with a subtitle showing the current status: "Enabled" or "Disabled".
3. User taps the row → navigates to `TwoFactorSetupView`.

**If 2FA is currently disabled (enable flow):**

4. **Step 1 — Password confirmation:**
   - Screen title: "Set Up Two-Factor Authentication"
   - Info banner: _"Two-factor authentication adds an extra layer of security to your account. Each time you sign in, you'll need your password plus a code from your authenticator app."_
   - Text field: **"Current password"** (secure, password input)
   - Primary button: **"Continue"**
   - User enters their current password and taps Continue.
   - App calls `POST /2fa/generator/google` with `{ password }` → receives a QR code image.

5. **Step 2 — Scan QR code and verify:**
   - Screen title: "Scan the QR Code"
   - Instructions:
     1. _"Install Google Authenticator from the App Store or Google Play."_
     2. _"Open the app and scan the QR code below."_
     3. _"Enter the 6-digit code from the app to confirm."_
   - QR code image displayed (150×150).
   - Text field: **"Enter 6-digit code"** (numeric, max 6 characters)
   - **"Trust this device"** toggle (same as login screen, with same info icon and optional device name field)
   - Primary button: **"Enable 2FA"** (disabled until 6 digits entered)
   - Cancel button
   - App calls `POST /2fa/enable/google` with `{ code, trust_device, trust_device_name }`.
   - On success: moves to Step 3.

6. **Step 3 — Save backup codes:**
   - Screen title: "Save Your Backup Codes"
   - Info banner: _"Save these backup codes somewhere safe. Each code can only be used once. If you lose access to your authenticator app, you can use one of these codes to sign in."_
   - 8 backup codes displayed in a grid (2 columns).
   - Action buttons:
     - **"Copy all codes"** — copies all 8 codes to clipboard. Success toast: "Codes copied to clipboard."
     - **"Download as .txt"** — downloads a text file with the codes.
   - Primary button: **"Done"** — returns to Settings. Shows success snackbar: "Two-factor authentication enabled."

**If 2FA is currently enabled (disable flow):**

4. Screen title: "Disable Two-Factor Authentication"
   - Warning banner: _"Disabling two-factor authentication will make your account less secure. You can re-enable it at any time from Settings."_
   - Text field: **"Current password"**
   - Primary button (red/destructive): **"Disable 2FA"**
   - Cancel button
   - App calls `POST /2fa/disable` with `{ password }`.
   - On success: navigates back to Settings. Success snackbar: "Two-factor authentication disabled."

**View existing backup codes (when 2FA is enabled):**
- On the Settings row for Two-Factor Authentication, there is a secondary action: **"View backup codes"**.
- Tapping it calls `POST /2fa/fetch/backup_codes` and navigates to a screen showing the existing backup codes with a "Regenerate codes" option.
- Regenerate calls `POST /2fa/generator/backup_codes`. Success snackbar: "New backup codes generated."

---

### Acceptance criteria:

- [ ] "Two-Factor Authentication" row is visible in iOS Settings with status subtitle "Enabled" or "Disabled"
- [ ] Tapping the row navigates to `TwoFactorSetupView`
- [ ] "Trust this device" toggle is visible on the existing `TwoFactoreView.swift` login screen below the code input
- [ ] Info icon next to "Trust this device" shows the tooltip: _"If you trust this device, you won't need to enter a 2FA code the next time you log in from this device. If you log in from a different device, you'll need to verify again."_
- [ ] When "Trust this device" is toggled ON, a "Device name (optional)" text field appears
- [ ] "Trust this device" toggle is hidden when user is in backup code mode
- [ ] Tapping Submit on the 2FA screen sends `trust_device` and `trust_device_name` in the `POST /2fa/validator/google` request
- [ ] Enable flow Step 1: entering password and tapping Continue calls `POST /2fa/generator/google`; QR code is displayed on Step 2
- [ ] Enable flow Step 2: valid 6-digit code + tapping "Enable 2FA" calls `POST /2fa/enable/google`; on success navigates to backup codes step
- [ ] Enable flow Step 2: "Enable 2FA" button is disabled until exactly 6 digits are entered
- [ ] Enable flow Step 2: "Trust this device" toggle and optional device name field work the same as on the login screen
- [ ] Enable flow Step 3: 8 backup codes displayed; "Copy all codes" copies them to clipboard; "Download as .txt" downloads them; "Done" returns to Settings with success snackbar "Two-factor authentication enabled."
- [ ] Disable flow: entering correct password and tapping "Disable 2FA" calls `POST /2fa/disable`; on success returns to Settings with snackbar "Two-factor authentication disabled."
- [ ] "View backup codes" option visible when 2FA is enabled; calls `POST /2fa/fetch/backup_codes` and shows existing codes
- [ ] "Regenerate codes" calls `POST /2fa/generator/backup_codes`; success snackbar: "New backup codes generated."
- [ ] Error states: wrong password → snackbar "Incorrect password. Please try again." | invalid/expired 2FA code → snackbar "Invalid code. Please check your authenticator app and try again." | network failure → snackbar "Something went wrong. Please try again."
- [ ] Loading overlay shown during all API calls; buttons disabled while loading
- [ ] New iOS `EndPoints.swift` entries: `POST /2fa/generator/google`, `POST /2fa/enable/google`, `POST /2fa/disable`, `POST /2fa/generator/backup_codes`, `POST /2fa/fetch/backup_codes`
- [ ] All UI strings use the app's existing localization system (no hardcoded English strings)

---

### Mock-ups:

N/A — reference web app: `contentstudio-frontend/src/modules/setting/components/TwoFactorAuth.vue` and `contentstudio-frontend/src/modules/account/views/TwoFactorVerification.vue` for design parity.

---

### Impact on existing data:

No schema changes. All backend endpoints already exist. `TwoFactoreView.swift` is modified to include trust device params. New `TwoFactorSetupView.swift` added for settings flow.

---

### Impact on other products:

- **Web:** No impact. 2FA backend APIs are shared and unchanged.
- **Android:** Separate story — **[Android] Two-factor authentication setup and trust this device**.
- **Chrome extension:** No authentication management in Chrome extension.

---

### Dependencies:

None. All backend APIs are live.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, native iOS screen
- [ ] Multilingual support — all UI strings (screen titles, button labels, info banner copy, tooltip text, snackbar messages, error messages) must use the existing iOS localization system
- [ ] UI theming support — N/A, iOS app does not use web theming system
- [ ] White-label domains impact review — N/A for native iOS
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Story 2: [Android] Two-factor authentication setup and trust this device

### Description:

As a ContentStudio user on Android, I want to enable or disable two-factor authentication from the app's settings, and when signing in with 2FA enabled I want the option to trust my device so I can skip 2FA on future logins from the same device.

This story covers two related gaps in the current Android app:
1. **Trust this device** — `TwoFactorAuthenticationActivity.java` already handles 2FA verification at login but does not pass `trust_device` or `trust_device_name` to the API.
2. **2FA Settings** — there is no screen to enable or disable 2FA from the Android app at all.

Web reference: `contentstudio-frontend/src/modules/account/views/TwoFactorVerification.vue` and `contentstudio-frontend/src/modules/setting/components/TwoFactorAuth.vue`.

---

### Workflow:

#### Part A — Trust This Device (on the existing 2FA verification screen)

1. User signs in with email and password on the Android app.
2. If 2FA is enabled on their account, the app navigates to the existing `TwoFactorAuthenticationActivity`.
3. Below the authentication code input field, the user sees a checkbox row:
   - **Checkbox label:** "Trust this device"
   - **Info icon (ℹ)** next to the label — tapping it shows a bottom sheet or info dialog:
     > _"If you trust this device, you won't need to enter a 2FA code the next time you log in from this device. If you log in from a different device, you'll need to verify again."_
4. When the user checks "Trust this device", a text field appears below:
   - **Hint:** "Device name (optional)"
   - **Placeholder:** "e.g. My Android, Work Phone"
5. The "Trust this device" checkbox is hidden when the user is in backup code mode.
6. User enters their 6-digit code and taps **Submit**.
7. The app sends `trust_device: true` (and `trust_device_name` if entered) along with the code and `user_info` to `POST /2fa/validator/google`.
8. On success: user is logged in. If trust was enabled, future logins from this device skip the 2FA screen.

#### Part B — 2FA Setup in Settings

1. User opens the Android app and goes to **Settings** (`SettingActivity`).
2. User sees a new entry: **"Two-Factor Authentication"** — with a secondary label showing the current status: "Enabled" or "Disabled".
3. User taps the entry → navigates to `TwoFactorSetupActivity`.

**If 2FA is currently disabled (enable flow):**

4. **Step 1 — Password confirmation:**
   - Screen title: "Set Up Two-Factor Authentication"
   - Info banner: _"Two-factor authentication adds an extra layer of security to your account. Each time you sign in, you'll need your password plus a code from your authenticator app."_
   - Input field: **"Current password"** (password input type)
   - Primary button: **"Continue"**
   - User enters their current password and taps Continue.
   - App calls `POST /2fa/generator/google` with `{ password }` → receives a QR code image.

5. **Step 2 — Scan QR code and verify:**
   - Screen title: "Scan the QR Code"
   - Instructions:
     1. _"Install Google Authenticator from the App Store or Google Play."_
     2. _"Open the app and scan the QR code below."_
     3. _"Enter the 6-digit code from the app to confirm."_
   - QR code image displayed (150×150 dp).
   - Input field: **"Enter 6-digit code"** (numeric keyboard, max 6 characters)
   - **"Trust this device"** checkbox (same as login screen, with same info icon and optional device name field)
   - Primary button: **"Enable 2FA"** (disabled until 6 digits entered)
   - Cancel/back button
   - App calls `POST /2fa/enable/google` with `{ code, trust_device, trust_device_name }`.
   - On success: navigates to Step 3.

6. **Step 3 — Save backup codes:**
   - Screen title: "Save Your Backup Codes"
   - Info banner: _"Save these backup codes somewhere safe. Each code can only be used once. If you lose access to your authenticator app, you can use one of these codes to sign in."_
   - 8 backup codes displayed in a grid (2 columns).
   - Action buttons:
     - **"Copy all codes"** — copies all 8 codes to clipboard. Success snackbar: "Codes copied to clipboard."
     - **"Download as .txt"** — saves a text file with the codes to device Downloads.
   - Primary button: **"Done"** — returns to Settings. Success snackbar: "Two-factor authentication enabled."

**If 2FA is currently enabled (disable flow):**

4. Screen title: "Disable Two-Factor Authentication"
   - Warning banner: _"Disabling two-factor authentication will make your account less secure. You can re-enable it at any time from Settings."_
   - Input field: **"Current password"**
   - Primary button (red/destructive): **"Disable 2FA"**
   - Cancel/back button
   - App calls `POST /2fa/disable` with `{ password }`.
   - On success: navigates back to Settings. Success snackbar: "Two-factor authentication disabled."

**View existing backup codes (when 2FA is enabled):**
- On the Settings entry for Two-Factor Authentication, there is a secondary action or the setup screen includes: **"View backup codes"**.
- Tapping it calls `POST /2fa/fetch/backup_codes` and shows the existing backup codes with a "Regenerate codes" option.
- Regenerate calls `POST /2fa/generator/backup_codes`. Success snackbar: "New backup codes generated."

---

### Acceptance criteria:

- [ ] "Two-Factor Authentication" entry is visible in `SettingActivity` with secondary label "Enabled" or "Disabled"
- [ ] Tapping the entry navigates to `TwoFactorSetupActivity`
- [ ] "Trust this device" checkbox is visible in `TwoFactorAuthenticationActivity` below the code input field
- [ ] Info icon next to "Trust this device" shows the info text: _"If you trust this device, you won't need to enter a 2FA code the next time you log in from this device. If you log in from a different device, you'll need to verify again."_
- [ ] When "Trust this device" is checked, a "Device name (optional)" input field appears below
- [ ] "Trust this device" checkbox is hidden when user is in backup code mode
- [ ] Tapping Submit sends `trust_device` and `trust_device_name` in the `POST /2fa/validator/google` request body
- [ ] Enable flow Step 1: entering password and tapping Continue calls `POST /2fa/generator/google`; QR code is displayed on Step 2
- [ ] Enable flow Step 2: valid 6-digit code + tapping "Enable 2FA" calls `POST /2fa/enable/google`; on success navigates to backup codes step
- [ ] Enable flow Step 2: "Enable 2FA" button is disabled until exactly 6 digits are entered
- [ ] Enable flow Step 2: "Trust this device" and optional device name work the same as on the login screen
- [ ] Enable flow Step 3: 8 backup codes displayed; "Copy all codes" copies to clipboard with snackbar "Codes copied to clipboard."; "Download as .txt" saves to Downloads; "Done" returns to Settings with snackbar "Two-factor authentication enabled."
- [ ] Disable flow: entering correct password and tapping "Disable 2FA" calls `POST /2fa/disable`; on success returns to Settings with snackbar "Two-factor authentication disabled."
- [ ] "View backup codes" visible when 2FA is enabled; calls `POST /2fa/fetch/backup_codes` and shows existing codes
- [ ] "Regenerate codes" calls `POST /2fa/generator/backup_codes`; success snackbar: "New backup codes generated."
- [ ] Error states: wrong password → snackbar "Incorrect password. Please try again." | invalid/expired 2FA code → snackbar "Invalid code. Please check your authenticator app and try again." | network failure → snackbar "Something went wrong. Please try again."
- [ ] Progress indicator shown during all API calls; buttons disabled while loading
- [ ] `ServiceManager.java` includes new Retrofit endpoints: `POST /2fa/generator/google`, `POST /2fa/enable/google`, `POST /2fa/disable`, `POST /2fa/generator/backup_codes`, `POST /2fa/fetch/backup_codes`
- [ ] All UI strings use the existing `L.menu()` / string resource localization system

---

### Mock-ups:

N/A — reference web app: `contentstudio-frontend/src/modules/setting/components/TwoFactorAuth.vue` and `contentstudio-frontend/src/modules/account/views/TwoFactorVerification.vue` for design parity.

---

### Impact on existing data:

No schema changes. All backend endpoints already exist. `TwoFactorAuthenticationActivity.java` is modified to include trust device params. New `TwoFactorSetupActivity.java` + layout added for settings flow. `ServiceManager.java` gets 5 new authenticated Retrofit endpoint definitions.

---

### Impact on other products:

- **Web:** No impact. 2FA backend APIs are shared and unchanged.
- **iOS:** Separate story — **[iOS] Two-factor authentication setup and trust this device**.
- **Chrome extension:** No authentication management in Chrome extension.

---

### Dependencies:

None. All backend APIs are live.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, native Android screen
- [ ] Multilingual support — all UI strings (screen titles, button labels, info banner copy, info tooltip text, snackbar messages, error messages) must use the existing `L.menu()` / string resource localization system
- [ ] UI theming support — N/A, Android app does not use web theming system
- [ ] White-label domains impact review — N/A for native Android
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
