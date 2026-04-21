# Epic & Stories — Sign Out Other Devices on Password Reset

## Epic

**Name:** Sign out other devices on password reset
**Objective:** 2026 - Q2 (id `114402`)
**Timeline:** 2026-04-20 → 2026-05-01 (current sprint)
**State:** To Do
**Group:** Backend (primary)

### Epic description (will be pushed to Shortcut)

#### Why this epic exists

Security review flagged OWASP ASVS V3.3 item #27: after a password change via reset/recovery, all other active sessions must be terminable. ContentStudio's **change-password** flow (authenticated user, Settings → Security) already does this correctly — it signs the user out of all other devices and revokes trusted devices. The **password reset** flow (forgot-password email link → set new password) and the **set-password** flow (new team invitees setting their first password, SSO users adding a password) **do not** consistently do this. This epic closes that gap.

#### Verified gaps (from codebase review)

1. `ResetPassword::recoverPassword` has a latent bug: for users with 2FA enabled, the flow returns a `TwoFactorChallengeData` response early and the existing `terminateAllSessions` call is **unreachable code** — 2FA users get zero session termination on password reset.
2. `ResetPassword::recoverPassword` never calls `TrustedDeviceService::removeUserTrustedDevices` — trusted devices survive a password reset for every user, 2FA or not.
3. `AuthController::setPassword` calls neither — new-invitee and SSO-add-password flows have no session hygiene at all.
4. The two cleanup calls (`terminateAllSessions` + `removeUserTrustedDevices`) are inlined separately in each controller today. No shared helper. Every password-write flow has to remember to call both, which is why `recoverPassword` and `setPassword` diverge from `changePassword`.

#### What's in scope

- Extract a shared helper that invalidates both sessions and trusted devices.
- Wire it into `recoverPassword` (fixing the 2FA early-return bug in the process) and `setPassword`.
- Refactor `changePassword` to use the same helper — no behavior change, just deduplication.
- Surface the outcome in the UI on reset-password and set-password success, matching the messaging change-password already has.

#### Explicitly out of scope

- Revoking developer API keys (`api_keys` collection) on password operations.
- Revoking mobile device push-notification tokens (`mobile_devices` collection) on password operations.

These surfaces are currently untouched on every flow including change-password. Widening scope to cover them has real user impact (third-party integrations break, push notifications stop) and requires a separate product/security conversation. Parked as a follow-up research item.

---

## Story 1

**Title:** `[BE] Unify session invalidation across password reset and first-time set-password flows`

**Custom fields:**
- Priority: High (security fix — closes OWASP ASVS V3.3 gap)
- Product area: Settings
- Skill set: Backend

**Description:**

As a ContentStudio user, when I reset my password via the forgot-password email link or set a password for the first time (as a newly invited team member or an SSO user adding a password), I want all my other active sessions signed out and my trusted devices revoked — the same way they are when I change my password from Settings → Security today — so that if an attacker has an active session on my account, my password action actually locks them out.

**Workflow (end-to-end behavior the user experiences):**

1. A user's account is compromised — an attacker has an active session or a trusted device.
2. The user does one of:
   a. Resets their password via the forgot-password email link (`POST /recoverPassword`)
   b. Sets their password for the first time after being invited as an approver / team member, or as an SSO user adding a password (`POST /setPassword`)
3. On success, the attacker's session is immediately invalidated — the attacker's JWT is removed from `user_sessions`, their Redis session cache is evicted, and a `session_destroyed` event is broadcast to the attacker's client.
4. The attacker's trusted device is removed from `trusted_devices` — they can no longer bypass 2FA on future login attempts from that device.
5. The rightful user is signed in on the device they just used for the reset / set operation.

**Acceptance criteria:**

- [ ] A shared helper method (suggested name: `invalidateAllSessions($userId, ?string $keepToken = null)`) is introduced that wraps `UserSessionRepository::terminateAllSessions($userId, $keepToken)` + `TrustedDeviceService::removeUserTrustedDevices($userId)`. Location and signature determined during implementation; placement should follow existing `Services/` or repository conventions.
- [ ] `ProfileController::changePassword` is refactored to call the new helper instead of the two inline calls — no behavior change, no regression.
- [ ] `ResetPassword::recoverPassword` is updated so the helper is called for **every** user, including users with 2FA enabled. The helper must be invoked **before** any `TwoFactorChallengeData` early-return so the session-termination logic is not unreachable code.
- [ ] For 2FA-enabled users resetting their password, verified that other JWT sessions are terminated and trusted devices are cleared at the moment the new password is persisted (not only after 2FA challenge completion).
- [ ] `AuthController::setPassword` is updated to call the helper after persisting the new password.
- [ ] Feature tests added / updated for all three paths: change-password, recover-password (with and without 2FA), set-password. Each test asserts: (a) new password persists, (b) all other `user_sessions` rows are deleted, (c) all `trusted_devices` rows for the user are deleted, (d) the current device's session survives in change-password, (e) for recover-password and set-password, the response includes a fresh valid session for the device that performed the action.
- [ ] Existing change-password tests still pass (no regression on a working flow).
- [ ] API response shape of `recoverPassword` and `setPassword` is unchanged (FE story can rely on existing responses; this is a pure backend behavior change).
- [ ] If the shared helper fails (e.g., MongoDB outage), the password write still succeeds — failure to clean up sessions does not mask a successful password change from the user. Failures are logged via the existing log pipeline.

**Mock-ups:** N/A — backend-only story.

**Impact on existing data:**

None at rest. Going forward, `user_sessions` and `trusted_devices` rows will be cleared more aggressively on reset / set-password than they are today. Users currently holding stale sessions / trusted devices on accounts that subsequently go through a reset or first-time-set will be signed out once those flows run — this is the intended behavior.

**Impact on other products:**

- **Web app:** direct dependency; the FE companion story relies on this change being in place.
- **Mobile apps (iOS / Android):** mobile apps authenticate via the same JWT sessions. An active mobile session that was logged in before a password reset will be signed out on the next authenticated request. This matches the behavior a user already expects on change-password today.
- **Chrome extension:** same JWT session model; same behavior as mobile.
- **White-label:** no impact.

**Dependencies:**

None. This story is the foundation and can ship independently of the FE story. If BE ships first and FE lags, users will see the existing generic success toast but sessions will still be correctly invalidated behind the scenes.

**Global quality & compliance (wherever applicable):**

- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, backend-only; no user-facing strings introduced in this story
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — ensure white-label workspaces behave the same (they use the same auth stack)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 2

**Title:** `[FE] Surface session-termination messaging on password reset and first-time set-password success`

**Custom fields:**
- Priority: Medium (completes the UX for the security fix; not security-critical on its own)
- Product area: Settings
- Skill set: Frontend

**Description:**

As a user who just reset my password via the forgot-password email link, or set my password for the first time as a new team invitee / SSO user, I want the success message to tell me that all my other active sessions and trusted devices were signed out — the same way the change-password flow already confirms it from Settings → Security — so I understand what just happened to my account and feel confident that it is now secure.

**Workflow:**

1. User clicks the forgot-password email link and lands on the Reset Password screen (or, for a new invitee / SSO user, lands on the Set Password screen).
2. User enters a new password + confirm password, clicks **Reset** / **Save Password**.
3. On successful save, a toast appears with copy that explicitly confirms: *"Password updated. All other active sessions have been signed out and trusted devices have been revoked."* (exact wording aligned with the existing change-password success copy, tweaked for the post-reset / post-first-set context).
4. Existing post-success navigation behavior is preserved — the user continues into the app as they do today on reset / set-password. (No change to routing; scope is copy-only.)
5. The new copy is i18n-keyed and mirrored across all supported locale directories under `src/locales/`.

**Acceptance criteria:**

- [ ] Add new i18n key for the reset-password success toast, e.g. `common.profile_mixin.success.password_reset` with English copy: *"Password has been successfully reset. All other active sessions have been signed out and trusted devices have been revoked."*
- [ ] Add new i18n key for the set-password success toast, e.g. `common.profile_mixin.success.password_set_with_session_termination` with English copy: *"Password has been successfully set. All other active sessions have been signed out and trusted devices have been revoked."* (Consider whether this key replaces or supplements the existing `common.profile_mixin.success.password_set` — confirm during implementation.)
- [ ] `ResetPassword.vue` (`src/components/authentication/ResetPassword.vue`) success handler replaces the server-driven toast with the new i18n-keyed string. Server error paths continue to use the server message.
- [ ] `SetPassword.vue` (`src/components/authentication/SetPassword.vue`) success handler replaces the server-driven toast with the new i18n-keyed string.
- [ ] Both new keys are added to **every** locale directory under `src/locales/` (per frontend CLAUDE.md rule on locale parity — English is source of truth; all other locales get the key, initially mirroring the English string if no translation is ready yet).
- [ ] The forgot-password (email-step) success toast on `ForgotPassword.vue` is left alone — no session termination happens at that step.
- [ ] The existing change-password success toast copy (`common.profile_mixin.success.password_updated`) is left alone — no regression, no duplication of the fix.
- [ ] Toast uses the existing `useAlertStore` / `alertMessage` pattern already in these files — no new toast infrastructure.
- [ ] Verified manually: reset flow shows the new copy; set-password flow shows the new copy; navigation after success is unchanged.

**Mock-ups:** N/A — copy-only change, no visual redesign.

**Impact on existing data:**

None.

**Impact on other products:**

- **Mobile apps (iOS / Android):** mobile apps have their own login UIs. Mobile will continue to show whatever copy they render today; a separate mobile story can mirror this change if the PO wants parity. Flagged, not included in this story.
- **Chrome extension / white-label:** no impact. Same web app codebase.

**Dependencies:**

Pairs with — but does not hard-depend on — **[BE] Unify session invalidation across password reset and first-time set-password flows**. If the FE ships first, the copy becomes accurate once BE ships. If BE ships first (likely), the copy just lags behind until FE ships.

**Global quality & compliance (wherever applicable):**

- [ ] Mobile responsiveness — verified (toast component is already responsive across supported viewports)
- [ ] Multilingual support — new keys added to every locale under `src/locales/` per frontend CLAUDE.md
- [ ] UI theming support — N/A, reuses existing toast component which already handles theming / white-label
- [ ] White-label domains impact review — toast copy is generic; no white-label branding decisions required
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
