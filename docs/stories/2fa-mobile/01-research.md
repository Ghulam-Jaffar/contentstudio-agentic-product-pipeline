# Research: 2FA + Trust This Device — iOS & Android

## Existing Shortcut Epic
- Epic 94625 — currently named "Trust this device - 2FA"
- Stories to UPDATE (not create new):
  - **92605** — "Trust this device - 2FA - Android App"
  - **92606** — "Trust this device - 2FA - iOS"
  - 98324 — doc story, ignore

---

## Web App — What Exists (reference implementation)

### 2FA Verification at Login (`TwoFactorVerification.vue`)
- User lands here after email/password login if 2FA is enabled on their account
- Route: `/2fa/:token` — `token` is a `user_info` value passed from login response
- Input: 6-digit code (authenticator or backup code mode)
- Toggle link: "Verify with backup code" ↔ "Authenticate with Google Authenticator"
- **Trust this device** checkbox (hidden in backup code mode):
  - Label: "Trust this device"
  - Info tooltip: _"If you choose to trust this device, you can temporarily skip 2FA. If you log in with a different device or browser, you will need to authenticate again."_
  - When checked: optional text field appears — "Enter Device Name (Optional)"
- Submit calls:
  - `POST /2fa/validator/google` with `{ user_info, code, trust_device, trust_device_name }`
  - `POST /2fa/validator/backup_codes` with `{ user_info, code }` (trust device N/A for backup code)
- On success: sets JWT token + logged_user, navigates to app

### 2FA Setup in Settings (`TwoFactorAuth.vue` — `src/modules/setting/components/`)
3-step flow (enable) / 1-step flow (disable):

**Step 1 — Password confirmation**
- If 2FA disabled: prompt "Enter your password to enable two-factor authentication"
- If 2FA enabled: prompt "Enter your password to disable two-factor authentication" → confirm dialog
- Calls `POST /2fa/generator/google` with `{ password }` → returns `qrcode_image` (base64) → moves to Step 2

**Step 2 — Scan QR + verify**
- Instructions: install Google Authenticator (App Store / Play Store links), scan QR code, enter code
- QR code image displayed (150x150)
- Input: 6-digit code
- **Trust this device** checkbox + optional device name (same as login screen)
- Calls `POST /2fa/enable/google` with `{ code, trust_device, trust_device_name }` → on success goes to Step 3

**Step 3 — Backup codes**
- Shows 8 backup codes grid
- Options: Download (.txt), Print (.pdf via `POST /2fa/print/backup_codes`), Copy to clipboard
- "View backup codes" link available if 2FA already enabled (calls `POST /2fa/fetch/backup_codes`)
- Regenerate codes button (calls `POST /2fa/generator/backup_codes`)

---

## Backend APIs (all exist, no new endpoints needed)

| Endpoint | Auth Required | Purpose |
|---|---|---|
| `POST /2fa/validator/google` | No | Verify TOTP code at login; accepts `user_info`, `code`, `trust_device`, `trust_device_name` |
| `POST /2fa/validator/backup_codes` | No | Verify backup code at login; accepts `user_info`, `code` |
| `POST /2fa/generator/google` | Yes | Start enable flow — verify password, get QR code |
| `POST /2fa/enable/google` | Yes | Confirm enable with TOTP code |
| `POST /2fa/disable` | Yes | Disable 2FA with password |
| `POST /2fa/generator/backup_codes` | Yes | Generate new backup codes |
| `POST /2fa/fetch/backup_codes` | Yes | Fetch existing backup codes |

Source: `contentstudio-backend/routes/web/auth.php:29,60`

---

## iOS — Current State

### Login 2FA (`TwoFactoreView.swift`)
- File: `contentstudio-ios-v2/ContentStudio/Controllers/Authentication/SwiftUIView/TwoFactoreView.swift`
- SwiftUI view shown after login when 2FA is enabled (triggered in `LoginView.swift:1058`)
- Has: authenticator code input + backup code mode toggle
- **Missing: "Trust this device" checkbox + optional device name + passing `trust_device`/`trust_device_name` in API call**
- Endpoints: `TwoFactorAuthEndpoint` (`/2fa/validator/google`), `BackupCodeAuthEndpoint` (`/2fa/validator/backup_codes`) — defined in `EndPoints.swift`

### 2FA Settings
- **Completely missing** — no 2FA enable/disable screen anywhere in Settings

### Android ServiceManager
- `contentstudio-android-v2/.../Network/ServiceManager.java:31,34`
  ```
  @POST("2fa/validator/google")  → google()
  @POST("2fa/validator/backup_codes")  → backupCodes()
  ```
- No authenticated 2FA endpoints defined (generator, enable, disable, backup codes)

---

## Android — Current State

### Login 2FA (`TwoFactorAuthenticationActivity.java`)
- File: `contentstudio-android-v2/.../Authentication/TwoFactorAuthenticationActivity.java`
- Activity shown after login when 2FA is required
- Has: authenticator code input + backup code mode toggle
- **Missing: "Trust this device" checkbox + optional device name + passing `trust_device`/`trust_device_name` in API call**
- API call sends only `{ code, user_info }` — no trust device params

### 2FA Settings
- **Completely missing** — no 2FA enable/disable screen anywhere in Settings

---

## What Needs to Change

### iOS
1. **`TwoFactoreView.swift`**: Add "Trust this device" toggle (hidden in backup mode) + optional device name field; update `TwoFactorAuthEndpoint` to pass `trust_device` + `trust_device_name`
2. **New `TwoFactorSetupView.swift`**: 3-step enable/disable 2FA flow in Settings (matches web)
3. **Settings entry**: Add "Two-Factor Authentication" row in the Settings/Profile section
4. **New iOS endpoints in `EndPoints.swift`**: `POST /2fa/generator/google`, `POST /2fa/enable/google`, `POST /2fa/disable`, `POST /2fa/generator/backup_codes`, `POST /2fa/fetch/backup_codes`

### Android
1. **`TwoFactorAuthenticationActivity.java`**: Add "Trust this device" checkbox + optional device name field; update API call to pass `trust_device` + `trust_device_name`
2. **New `TwoFactorSetupActivity.java`**: 3-step enable/disable 2FA flow in Settings (matches web)
3. **Settings entry**: Add "Two-Factor Authentication" entry in SettingActivity
4. **New Android ServiceManager endpoints**: all authenticated 2FA endpoints above

---

## Files Involved

**iOS (modified):**
- `TwoFactoreView.swift` — add trust device UI + params
- `EndPoints.swift` — add authenticated 2FA endpoints
- Settings view file (to add 2FA entry)

**iOS (new):**
- `Views/Settings/TwoFactorSetupView.swift` (or similar)

**Android (modified):**
- `Authentication/TwoFactorAuthenticationActivity.java` — add trust device UI + params
- `Network/ServiceManager.java` — add authenticated 2FA Retrofit endpoints
- `Setting/SettingActivity.java` — add 2FA entry

**Android (new):**
- `Authentication/TwoFactorSetupActivity.java` + layout
