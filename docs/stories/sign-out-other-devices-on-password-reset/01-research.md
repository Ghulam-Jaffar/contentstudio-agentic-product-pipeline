# Research — Sign Out Other Devices on Password Reset

## Background

Security reviewer (Umair) flagged item **#27** from what appears to be an OWASP ASVS V3.3 audit checklist:

> Verify that the application gives the option to terminate all other active sessions after a successful password change (including change via password reset/recovery), and that this is effective across the application, federated login (if present), and any related parties.

**His observation:** change-password terminates other sessions today; reset-password does not.

**Actual finding after codebase verification:** the gap is larger than his message implied, and includes a latent bug.

## Backend findings

### Change password — `POST /changePassword`
[ProfileController.php:117-164](contentstudio-backend/app/Http/Controllers/Accounts/ProfileController.php#L117-L164)

Does the right thing:
1. Verifies old password hash; saves new hash.
2. Calls `UserSessionRepository::terminateAllSessions($userId, $currentToken)` — kills every JWT session in the `user_sessions` MongoDB collection (plus Redis cache eviction and `session_destroyed` broadcast), keeping the current device alive.
3. Calls `TrustedDeviceService::removeUserTrustedDevices($userId)` — wipes the `trusted_devices` collection for the user so previously-trusted devices must re-pass 2FA.

### Password reset — `POST /recoverPassword`
[ResetPassword.php:228-291](contentstudio-backend/app/Http/Controllers/Auth/ResetPassword.php#L228-L291)

Two bugs:
1. **For 2FA users:** the flow returns a `TwoFactorChallengeData` response at line 258 and exits. The `terminateAllSessions` call at line 272 is **unreachable** — it never runs. Every other session of a 2FA-enabled user stays alive after a password reset.
2. **For everyone (2FA or not):** `TrustedDeviceService::removeUserTrustedDevices` is never called. A device trusted before the reset continues to bypass 2FA after the reset (because trusted-device fingerprint matching does not depend on the password).

### Set password — `POST /setPassword`
[AuthController.php:976-1014](contentstudio-backend/app/Http/Controllers/Auth/AuthController.php#L976-L1014)

Used by:
- Newly invited approvers / team members setting their first password
- SSO users adding a password to their existing account

Calls **neither** `terminateAllSessions` **nor** `removeUserTrustedDevices`. No session hygiene at all. If an attacker has hijacked a session before a first-time password is set, the password set doesn't lock them out.

### No shared helper today
`terminateAllSessions` and `removeUserTrustedDevices` are called inline, separately, from each controller. There is no central `invalidateAllSessions($userId, $keepToken)` helper. Every password-write flow has to remember to call both — which is exactly why `recoverPassword` and `setPassword` diverge from `changePassword` today.

### Out-of-scope surfaces (flagged, not in this epic)
These surfaces are NOT touched by any of the three flows today. Not required by ASVS V3.3 strictly, but worth knowing:
- `api_keys` collection — developer API keys survive every password operation
- `mobile_devices` collection (push notification tokens) — survives every password operation; only removed on explicit logout

Decision: leave these out of this epic. Separate follow-up discussion needed because killing them breaks third-party integrations and mobile push.

## Frontend findings

### Change password UX (already good)
[Profile.vue](contentstudio-frontend/src/modules/setting/components/Profile.vue) via [useProfile.ts:251-258](contentstudio-frontend/src/composables/useProfile.ts#L251-L258):

- **Pre-submit blocking dialog** — copy: *"This will log you out of all devices except this one and revoke trusted devices. Are you sure you want to change your password?"*
- **Success toast** — copy: *"Password has been successfully updated and trusted devices have been revoked."*

Both keys already in [common.json](contentstudio-frontend/src/locales/en/common.json) under `common.profile_mixin.*`.

### Reset password UX (silent)
[ResetPassword.vue](contentstudio-frontend/src/components/authentication/ResetPassword.vue):

- On success, shows whatever string the server returns (`alertMessage(response.data.message, 'success')`) — no frontend-controlled copy.
- Silently auto-logs-in and redirects into the app.
- Zero mention of "other sessions have been signed out" or "trusted devices revoked."

### Forgot-password (email step) UX (silent)
[ForgotPassword.vue](contentstudio-frontend/src/modules/account/views/ForgotPassword.vue) — same pattern. Server-driven toast only.

### Set password UX (new invitees + SSO) (silent)
[SetPassword.vue](contentstudio-frontend/src/components/authentication/SetPassword.vue) — success toast key `common.profile_mixin.success.password_set` = *"Password has been set successfully for your account."* Nothing about sessions.

### Existing plumbing already in place
The [LinkedDevices.vue](contentstudio-frontend/src/modules/setting/components/LinkedDevices.vue) settings tab already has a "Logout of all devices" button with a "Revoke trusted devices" checkbox that calls `terminateAllSessionsApi({ revoke_trusted_devices })`. This proves the backend plumbing for "kill sessions + optionally revoke trusted devices" already exists and is endpoint-addressable — we just need a new internal helper that calls both unconditionally.

## Files involved

### Backend
- `contentstudio-backend/app/Http/Controllers/Auth/ResetPassword.php` (method `recoverPassword`)
- `contentstudio-backend/app/Http/Controllers/Auth/AuthController.php` (method `setPassword`)
- `contentstudio-backend/app/Http/Controllers/Accounts/ProfileController.php` (method `changePassword` — refactor to use shared helper)
- `contentstudio-backend/app/Repository/UserSessionRepository.php` (existing `terminateAllSessions`)
- `contentstudio-backend/app/Services/TrustedDeviceService.php` (existing `removeUserTrustedDevices`)
- New shared helper — location to be decided during implementation (likely `app/Services/Auth/` or as a trait / static helper)
- Tests under `tests/Feature/Auth/` or equivalent

### Frontend
- `contentstudio-frontend/src/components/authentication/ResetPassword.vue`
- `contentstudio-frontend/src/components/authentication/SetPassword.vue`
- `contentstudio-frontend/src/locales/*/common.json` (or `auth.json`) — new i18n keys for reset and set-password success copy
- No new components, no new routes, no API changes

## Story split

Two stories under a dedicated epic **"Sign out other devices on password reset"**:

1. `[BE]` Unify session invalidation across all password-write flows — shared helper, fix the 2FA early-return bug, wire into `recoverPassword` and `setPassword`, refactor `changePassword` to use the helper (no regression).
2. `[FE]` Add session-termination messaging to reset-password and set-password success flows — match the existing change-password copy pattern.

Explicitly out of scope: revoking `api_keys` and mobile device push tokens. Parked for a separate conversation with Umair / security.
