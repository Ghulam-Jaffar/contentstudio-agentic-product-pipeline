# YouTube Consent Reminder Emails — Fix · Story

**Platform:** Backend (email). No FE, no mobile.

| # | Story | Priority |
|---|---|---|
| S-1 | [BE] Fix YouTube consent reminder emails — schedule, email opt-out, and template compliance | Medium |

---

## S-1 · [BE] Fix YouTube consent reminder emails — schedule, email opt-out, and template compliance
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Integrations · **Priority:** Medium · **Type:** Feature

### Description
As a ContentStudio user who needs to re-grant YouTube Analytics data consent each month, I want the reminder emails to come less often, to respect my email-notification setting, and to look like every other ContentStudio email (correctly branded, with my workspace and account named), so that the reminders are helpful and on-brand rather than noisy and generic.

Today the reminder fires too many times, ignores the user's email preference, and doesn't follow the standard email template.

### Workflow
1. A user's YouTube Analytics consent is approaching its monthly expiry.
2. The user receives a reminder **3 days before** expiry and on the **last day** (existing final reminder) — no longer a week out.
3. If the user has turned **email notifications off**, they receive no reminder.
4. The reminder email follows ContentStudio's standard template — Subject, Header, Body, Footer — with the workspace name in the subject, the correct (white-label-aware) app name throughout, and the affected YouTube account named in the body.

### Acceptance criteria
- [ ] The **7-days-before** reminder is removed; reminders are sent only at **3 days before** expiry and on the **last day** (the existing final/`last_day` reminder is unchanged).
- [ ] A member whose **email notifications are off** does **not** receive the reminder; members who have them on still do.
- [ ] Each tier (3-day, last-day) is sent at most once per account per consent cycle (no duplicate reminders).
- [ ] The email follows ContentStudio's **standard template structure: Subject, Header, Body, Footer** (uses the shared email layout, not an ad-hoc body).
- [ ] The **subject line contains the workspace name**.
- [ ] The email uses the **dynamic app name** (white-label aware) everywhere — no hardcoded "ContentStudio".
- [ ] The **YouTube account name** is used in the body (so the user knows which account needs re-consent).
- [ ] Header and footer match the standard template (branding, links) and respect white-label settings.
- [ ] Removing the 7-day tier does not break the "already sent" de-duplication tracking (no account gets stuck or re-spammed during the transition).
- [ ] No reminders are sent after consent has already expired.

### Mock-ups
N/A — email (follows the existing standard email template).

### Impact on existing data
Uses the existing per-account reminder-sent tracking; the 7-day tier is removed (3-day and last-day tiers unchanged). Ensure in-flight accounts mid-cycle transition cleanly.

### Impact on other products
Email only. No web UI, mobile, or Chrome impact. Honors white-label branding (dynamic app name, header/footer).

### Dependencies
None.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend/email story
- [ ] Multilingual support (email should honor the recipient's locale per the standard email/localization handling)
- [ ] UI theming support — N/A (white-label branding handled via dynamic app name + standard template)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-backend/app/Jobs/Integrations/YouTubeConsentReminderJob.php` — `determineTier()` currently returns `7_day` / `3_day` / `last_day` (day 0). Drop `7_day`; **keep `3_day` and `last_day` as-is**. It pushes the email payload (with `account_name`, `workspace_id`, `days_remaining`, etc.) onto the `email_notification_redis` queue. Add the email-preference check here before enqueuing per member.
- `contentstudio-backend/app/Console/Commands/Integrations/YoutubeConsentReminderCommand.php` — schedules the window (`EARLIEST_REMINDER_DAYS = 7`, `CONSENT_VALIDITY_DAYS = 30`); update the earliest-reminder window now that 7-day is gone, and the description ("at 7, 3, and 0 days").
- `contentstudio-backend/app/Notifications/Account/YouTubeConsentNotification.php` — pulls title/description from `notifications.youtube_consent_reminder.*` config (`description_3_day`, `description_last_day`); align with the new tiers and the template/copy (subject with workspace name, account name in body, dynamic app name).
- `contentstudio-backend/app/Models/Notification/NotificationsSetting.php` — `email_notifications` flag; use it to skip members who've turned email off.
- Standard email template: `contentstudio-backend/resources/views/emails/layout/head.blade.php` and the existing `resources/views/emails/*.blade.php` (e.g. `notifications.blade.php`) — model the Subject/Header/Body/Footer on these; pull the app name from the white-label/app-name config rather than hardcoding.
- Per `AGENTS.md` §9.4, outbound email should honor the recipient locale (`App::setLocale(...)`).
