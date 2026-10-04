# Research: Workspace Health Center

**Source:** Team meeting, 2026-09-30. A new page listing every workspace-level problem in one place: accounts that need connecting or reconnecting, anything with problems such as posting failures (for example Facebook returning errors on publish), and logs. Split from the app shell epic at the PO's call. **Skeleton epic.** The PO will discuss it further and add detail.

## Related work

- Helpin epic **Reconnect and connect accounts from wherever they're used** (`0a8daf97-...`, CONT-4156 to 4161, sprint Oct 05 - Oct 18). It adds a shared "Reconnect required" message and a **Reconnect now** action that reconnects the exact account and returns the user to where they were. The Health Center must reuse that action and copy, not build a second one. Local: `docs/features/account-reconnect-and-connect/`.
- The app shell refresh epic may give the Health Center a slot in the rail or top bar (`docs/features/app-shell-refresh/`).

## What exists today

**Backend (`contentstudio-backend/`)**
- No single workspace-health endpoint.
- Account state: `validity` (valid / expired / expiring_soon / invalid), `validity_error`, `validity_status`, `invalid_tries`, `sent_invalid_email`, `state` on `app/Models/Integrations/Platforms/SocialIntegrations.php` (82-89). Set by `app/Jobs/Integrations/Validate/SocialAccountsJob.php` (`setAccountValidity`). Filterable via `filters['validity']` in `app/Repository/Integrations/Platforms/SocialRepo.php:244` (Pinterest repo too). Email: `POST /notifications/accounts/expired` (`routes/web/notifications.php:31`).
- Failed posts: `PlansRepository::fetchPlansCounts` (`app/Repository/Publish/Planner/PlansRepository.php` ~870-980) counts `failed`, `partial`, `first_comment_failed`. Used by `POST fetchPlansCount` (`routes/web/planner.php:41`) and `POST getContentPublishingStats` (`routes/web/analytics.php:228`).
- Per-post errors: Mongo `posting` collection (`app/Models/Publish/Planner/Posting.php`, `Plans::posting()`), `error` / `error_message` written in `app/Libraries/Publish/Posting/Posting.php`.
- No per-workspace aggregated error log.

**AI agents (`contentstudio-ai-agents/`)**
- `src/integrations/contentstudio/models.py:104` maps invalid/expired to `needs_reconnect`; `bootstrap.py unhealthy_accounts`; toolkit `workspace_list_accounts(needs_reconnect=True)`; publishing toolkit surfaces posts with `errors` (`toolkits/publishing.py:466`). An agent could answer "what's wrong with my workspace?" from the same endpoint.

**Frontend (`contentstudio-frontend/src/`)**
- Expired tokens: `modules/common/components/header-notifications/HeadNotificationSlider.vue` (reads `getImportantNotifications.expired`, `stores/setting/useWorkspaceNotificationStore.ts:92`; CTAs route to social accounts with `filter: 'expired'`, line 281).
- `SocialAccountsDatatable.vue` (validity handling 700, 807, 1210), `components/dashboard/SocialAccountsCard.vue` (token_expired), `SocialAccountListItem.vue`, `CstAccountCheckBox.vue`.
- Failed posts: `components/dashboard/ContentPublishingCard.vue:125-147` (failed / partial counts, `redirectToPlanner('failed')`), Planner status filter `modules/planner_v2/components/FilterSidebar.vue:486`.
- No aggregated issues page.

## Candidate issue types for v1

| Issue | Source | Fix action |
|---|---|---|
| Account needs reconnecting (expired, invalid) | `validity` | Reconnect now (from the reconnect epic) |
| Account expiring soon | `validity = expiring_soon` | Refresh now |
| Account missing a permission (for example LinkedIn or Threads analytics scope) | platform-specific | Grant access |
| Failed posts | `posting.error` | Open post, retry |
| Partially failed posts | plan status `partial` | Open post, retry failed accounts |
| First comment failed | `first_comment_failed` | Open post |

Later candidates: inbox sync stopped, analytics sync failing, feeds or automations failing (RSS, evergreen), webhook delivery failures, plan limits reached, payment problems.

## Open questions for the PO

- Which issue types are in v1, and which are "logs" (history) vs "open issues" (need action)?
- Who sees it: all members, or admins only? Do collaborators and approvers see a filtered version?
- Is it per workspace only, or also an all-workspaces view for agencies?
- Is a problem resolved automatically when its cause goes away (account reconnected, post retried), or does the user dismiss it?
- Where it lives in navigation (rail item, top bar icon with a count, settings page)
- Should it notify (email, in-app) or only collect?
- Should it be exposed on the public API and MCP server so agents and automations can check workspace health? That would need a story per surface.

## CTO prototype (2026-09-30)

Canvas "Measurement & Health Center": https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ (5 artboards, created 2026-09-29). Superseded as the working reference by the updated design canvas below. Summary of what it specifies:

**Navigation.** A "Health" item in the desktop rail, between Library and API. PO addition: the rail icon is a heart that fills with workspace health (empty, half, full).

**Overview tab.** Header "Health Center / Everything that could stop a post from going out, in one place. Updated 2 minutes ago." with Alert settings and Compose Post. Tabs: Overview, Accounts ("6 need action"), Post delivery ("47 failed"), Alert settings.
- KPI cards: Post delivery rate, last 7 days (96.4%, "Up 0.8 points on the previous 7 days"); Posts published (1,284, "Across 42 connected accounts"); Posts failed (47, "25 were preventable: expired access or removed accounts"); Healthy accounts (36 of 42, "3 to reconnect · 2 paused · 1 not found").
- "Needs your attention": issues sorted by which posts will fail soonest, with Critical, Warning and Info counts. Each card has severity, plain explanation, affected accounts, impact ("11 scheduled posts affected, next in 2h 14m") and two actions. Examples: Reconnect 3 accounts (Critical). Facebook paused posting to 2 pages (Warning, "We've paused them until Oct 1, 09:40 so they don't get restricted further. Posts due in that window are on hold, not lost."). 2 scheduled posts still use a removed account (Warning). A teammate isn't getting mobile reminders (Info). A Google Business location wasn't found (Info, "We've stopped posting to it so it doesn't fail every day").
- Delivery rate by platform (7 days, bars from 80%, failed count per platform).
- Why posts failed: Expired or revoked access, Paused by the platform (posting limit), Media the platform rejected, Account removed from the workspace, Platform outage (still failing after 5 tries). "25 of these could have been avoided by reconnecting sooner or cleaning up removed accounts." See failed posts.
- Failed posts per day, with delivery rate under each day.

**Accounts tab.** "Every connected account, whether it can post right now, and when its access runs out." Connect account, Reconnect all. Filters: All, Reconnect required, Paused by platform, Not found, Expiring in 14 days, Healthy. Note: "Reconnecting keeps your scheduled posts. They go out on time as soon as access is restored." Columns: Account, Status, Access (expiry date and cause, e.g. "Password changed on Facebook", "LinkedIn access lasts 60 days", "Access removed in Google account settings", "12 of 17 posts left today" for X), Last successful post, Scheduled, Actions (Reconnect, View schedule, Check location, Renew access, Details). Accounts needing action listed first.

**Post delivery tab.** "Every post that didn't go out, why, and what to do about it." Last 7 days, Export CSV. Filters: All issues, Failed, On hold, Retrying, Published. Row: post text, account, scheduled time, reason pill. Detail panel: status, "What happened", "How to fix it", primary and secondary action, Timeline (scheduled time reached, platform response, retry decision, who was notified), "Technical details for support" (platform error code, attempts e.g. "1 of 1 (not retryable)" / "2 of 5", reference id). Reason examples with codes: Facebook 190/460 access expired, Google 404 location not found, Facebook 368 posting limit (on hold, will resume), Instagram 9004 media rejected (GIF), Google invalid_grant access revoked, LinkedIn 401 expired token, Pinterest 503 retrying.

**Notification panel (Home).** Bell panel with tabs All, Health, Publishing, Approvals, Mentions; grouped Today / Earlier; each item has icon, title, body, time, action and category; Mark all as read; Open Health Center. Home also shows a banner: "3 accounts need reconnecting / 11 scheduled posts will fail until you do. The next one is in 2h 14m." with Open Health Center and Reconnect.

**Alert settings tab.** "Choose who hears about problems, where, and how often. We alert once per issue, never once per post."
- Matrix of alerts by channel (In-app, Email, Slack, Mobile push): Account needs reconnecting (Critical; in-app can't be turned off), Platform paused posting, Post failed (only failures that need you; temporary errors retried first), Access expiring soon (7 days and 1 day before), Mobile reminder not delivered, Post published (grouped hourly), Weekly health summary (Mondays 09:00).
- Who gets alerts: Workspace owner (always gets critical), Admins, Whoever connected the account, Whoever scheduled an affected post (own posts only), Clients with approval access.
- Email frequency for failed posts: As they happen, Hourly digest (recommended), Daily digest at 09:00 (workspace time zone). Reconnect alerts always immediate.
- Slack: connected channel, Send test, Change channel.
- Quiet hours: email, Slack and push wait; critical alerts still show in the app.

**New system behaviour the prototype assumes** (needs tech lead confirmation):
- Pausing a page that hit a platform posting limit until the limit lifts, holding its posts instead of failing them.
- Stopping posting to a Google Business location that isn't found.
- Flagging scheduled posts that still target a removed account.
- Retrying temporary platform errors up to 5 times with increasing waits, and not retrying errors that can't succeed.
- Tracking mobile reminder device health (push token no longer accepted).
- Alerting once per issue, not once per post, and digests.
- ~~A Slack integration for alerts.~~ **Out of scope (PO, 2026-09-30): no Slack.** Alert channels are in-app, email and mobile push only.

Related: Helpin epic **Notifications architecture** (`ea0a4fe5-...`) and **Reconnect and connect accounts from wherever they're used** (CONT-4156 to 4161).

## Team decisions (PO, 2026-10-01)

After the PO's discussion with the team, the scope was narrowed. Where this and the CTO prototype differ, this wins.

- **Accounts tab:** one row per account. Accounts needing reconnection first, then accounts whose last post failed, then healthy. Repeated failures on one account update the same row ("Last 2 posts failed"), never a new row. Details shows the account's error log and failed posts. Filters: search, platform, status (Reconnect required, Not found, Last post failed, Expiring soon, Healthy).
- **Post delivery tab:** one entry per account per failed post (a post failing on 5 of 10 accounts = 5 entries). First entry selected by default, details in a side panel. Show only the raw error log and the timeline. Access expired or revoked: Reconnect plus the help doc link. Any other failure: Retry. No "what happened / how to fix" text and no reason categories.
- **Overview tab:** keep the four top tiles. Needs your attention: reconnect, removed account, Google Business location not found, mobile reminders. Keep delivery rate by platform. Drop "Why posts failed". Failed posts per day and published posts per day in one widget, switched with a `SegmentedControl`.
- **Dropped:** "Paused by platform" everywhere (platforms send no signal), the Alert settings tab, and all new health alerts (in-app and email notifications already exist and are managed in notification settings). That removes the alerts BE story, the alert settings FE story, the notification panel and Home banner story, and the Flutter push story.
- **Kept:** the heart rail icon and its thresholds, the research story for Bilal Tariq (now also covering whether posting limits can be detected and failure-reason classification).
- **Follow-up (PO, 2026-10-01):** no new Home banner, because the existing bottom reconnect banner (`contentstudio-frontend/src/modules/common/components/header-notifications/HeadNotificationSlider.vue`, mounted in `Home.vue`) already covers it. No CSV export for now. One `[Flutter]` story brings the whole Health Center to the app (reuses CONT-4229).

## Design canvas (2026-10-02)

Working reference for every story: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26 ("Workspace Health Center"). Built from the CTO prototype, which sits outside our organization and can't be edited, and updated to the team decisions above. Five boards:

- **Overview:** no Alert settings, no paused issue, no "Why posts failed". Delivery rate by platform next to one posts-per-day widget with a working Failed / Published toggle. Healthy accounts tile: reconnect, last post failed, not found, expiring soon.
- **Accounts:** status filters (Reconnect required, Not found, Last post failed, Expiring soon, Healthy), platform dropdown and search. One row per account in the agreed order, with "Last 2 posts failed" on a single row.
- **Accounts · Details:** drawer with the account's failed posts (Retry each) and its raw error log.
- **Post delivery:** one entry per account per failed post, first selected, side panel with raw error log and timeline. Reconnect plus "Why did this happen?" for expired access, Retry for the rest (works on the canvas). No CSV export, no reason pills.
- **Rail · Heart states:** full, half and empty, unselected and selected, with rules and tooltips.

The Google Business "not found" item no longer says we stop posting to it. That behaviour is still open in the research story.

