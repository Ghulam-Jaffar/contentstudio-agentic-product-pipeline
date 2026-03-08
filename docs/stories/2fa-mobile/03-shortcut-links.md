# Shortcut Links: 2FA + Trust This Device — iOS & Android

## Epic
- **Two-Factor Authentication — iOS & Android** — https://app.shortcut.com/contentstudio-team/epic/94625

## Stories
- **[iOS] Two-factor authentication setup and trust this device** — https://app.shortcut.com/contentstudio-team/story/92606
- **[Android] Two-factor authentication setup and trust this device** — https://app.shortcut.com/contentstudio-team/story/92605

## Sprint
09 March - 20 March - 2026 (iteration id: 111909)

## Web Reference
- Login verification: `contentstudio-frontend/src/modules/account/views/TwoFactorVerification.vue`
- Settings setup: `contentstudio-frontend/src/modules/setting/components/TwoFactorAuth.vue`

## Key Files Changed
**iOS:** `TwoFactoreView.swift` (trust device), new `TwoFactorSetupView.swift`, `EndPoints.swift` (+5 endpoints)
**Android:** `TwoFactorAuthenticationActivity.java` (trust device), new `TwoFactorSetupActivity.java`, `ServiceManager.java` (+5 endpoints), `SettingActivity.java`
