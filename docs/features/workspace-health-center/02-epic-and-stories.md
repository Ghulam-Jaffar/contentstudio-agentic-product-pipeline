# Epic: Workspace Health Center

## Epic description

When something goes wrong in a ContentStudio workspace, the user finds out piece by piece, if at all. An expired Facebook token shows up as a banner in one place and a warning icon in another. A post that Facebook rejected sits in the Planner under a "failed" filter. A partly failed post or a page Facebook has limited only shows if the user goes looking. Nothing puts it all in one place, and nothing tells the user plainly what will fail next and how to fix it.

The **Health Center** is a new area that shows everything that could stop a post from going out, in one place. It has four tabs:

- **Overview**: the delivery rate, posts published and failed, healthy accounts, and a **Needs your attention** list sorted by which posts will fail soonest, each issue with a plain explanation and a fix button.
- **Accounts**: every connected account, whether it can post right now, when its access runs out, and why.
- **Post delivery**: every post that didn't go out, why, what to do about it, and the technical details for support.
- **Alert settings**: who hears about problems, where (in-app, email, mobile push) and how often. One alert per issue, never one per post.

It also adds a **Health** item to the desktop rail with a **heart icon that fills with the workspace's health**, tabs to the notification panel, and a health banner on Home.

**Prototype (by the CTO):** https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

**Status:** partly skeleton. Scope will be refined after the PO's discussion with the technical team lead.

### Heart icon

| Heart | When |
|---|---|
| Full | No critical issues and a 7-day delivery rate of 95% or more |
| Half | Any warning, or a delivery rate from 80% to under 95% |
| Empty | Any critical issue (for example an account that needs reconnecting), or a delivery rate under 80% |

### Scope

In:

- Account health and post delivery data, and the metrics behind the Overview
- The four Health Center tabs
- Health alerts in-app, by email and by mobile push, with recipients, email digests, quiet hours and a weekly summary
- Notification panel tabs and a Home health banner
- The Health item and heart icon in the desktop rail
- Health alerts as push notifications in the mobile app
- A research story on automatic protection (pausing pages that hit posting limits and similar), before anything is built

Out:

- **Slack.** Not in scope
- Building a second reconnect flow. The Health Center reuses the one from **Reconnect and connect accounts from wherever they're used**
- Inbox, analytics and automation health (later)

### Stories

1. `[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype`
2. `[BE] Track account health: status, access expiry and cause, last successful post and scheduled posts`
3. `[BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center`
4. `[Research] Decide how ContentStudio should protect accounts automatically when platforms limit or break posting`
5. `[BE] Send health alerts by the workspace's alert settings`
6. `[FE] Build the Health Center Overview tab`
7. `[FE] Build the Health Center Accounts tab`
8. `[FE] Build the Health Center Post delivery tab with issue details and CSV export`
9. `[FE] Build Health Center alert settings`
10. `[FE] Add category tabs to the notification panel and a health banner on Home`
11. `[FE] Add the Health item to the rail with a heart that fills with workspace health`
12. `[Flutter] Receive Health Center alerts as push notifications`

---

# [Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype

### Description

As a designer, I want to turn the CTO's Health Center prototype into final designs, so that devs build the four tabs, the notification panel, the Home banner and the heart rail icon from one agreed reference.

The prototype covers the structure and copy. This story finalises it against the design system, removes Slack (out of scope), and adds the heart icon states the PO asked for.

---

### Workflow

1. Designer reviews the CTO prototype and today's related screens: expired-token banners, the social accounts table, the dashboard's failed-posts card, the Planner's failed filter and the notification dropdown.
2. Designer finalises the Overview, Accounts, Post delivery and Alert settings tabs.
3. Designer finalises the notification panel with category tabs and the Home health banner.
4. Designer designs the Health rail item with a heart icon in three states: full, half and empty.
5. Designer covers empty, loading and error states for every tab.

---

### Acceptance criteria

- [ ] Final designs for the Overview, Accounts, Post delivery (list and issue detail) and Alert settings tabs
- [ ] Alert settings show only in-app, email and mobile push. No Slack column, channel or test button
- [ ] The notification panel with All, Health, Publishing, Approvals and Mentions tabs, and the Home health banner
- [ ] The Health rail item with the heart icon in full, half and empty states, plus its tooltip, and how it looks in every rail theme
- [ ] Severity styles for Critical, Warning and Info that stay readable without colour alone
- [ ] Empty ("Everything's running smoothly"), loading and error states for every tab
- [ ] Every element is mapped to an existing `@contentstudio/ui` component, or flagged as a gap
- [ ] Designs are handed off to every `[FE]` story in this epic, and the push notification copy to the `[Flutter]` story

---

### Mock-ups:

CTO prototype: https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

---

### Impact on existing data:

None.

---

### Impact on other products:

Web app. The mobile app gets push notification copy only.

---

### Dependencies:

None. Blocks every `[FE]` story in this epic.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, design only, nothing API-facing changes

---

# [BE] Track account health: status, access expiry and cause, last successful post and scheduled posts

### Description

As a ContentStudio user, I want the app to know, for every connected account, whether it can post right now, when its access runs out and why, so that the Health Center can warn me before posts fail instead of after.

Account validity is tracked today (valid, expired, expiring soon, invalid), but not the cause, the exact expiry, the last successful post, or how many scheduled posts depend on the account. This story adds that and returns it for the Health Center.

**Skeleton.** Details to follow the tech lead discussion.

---

### Workflow

1. User opens the Health Center Accounts tab.
2. ContentStudio returns every account with its status, access expiry and cause, last successful post and number of scheduled posts.
3. Accounts that need action come first.

---

### Acceptance criteria

- [ ] Each account has a health status: **Reconnect required**, **Paused by platform**, **Not found**, **Expiring soon** (within 14 days) or **Healthy**
- [ ] Each account has its access expiry date (or "no expiry") and, where known, the cause in plain words (for example "Password changed on Facebook", "LinkedIn access lasts 60 days", "Access removed in Google account settings")
- [ ] Each account has its last successful post time and its number of scheduled posts, or posts on hold
- [ ] For X, the remaining posts allowed today is returned where available
- [ ] The list can be filtered by status, with counts per status, and is sorted with accounts that need action first
- [ ] Counts for the Overview: healthy accounts out of total, and how many are to reconnect, paused or not found
- [ ] Only accounts the requesting user can access are returned
- [ ] Reconnecting an account moves it back to Healthy straight away, and its scheduled posts go out as planned

---

### Mock-ups:

None. Backend. Reference: the Accounts tab in the CTO prototype: https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

---

### Impact on existing data:

Adds health fields to connected accounts (status cause, expiry, last successful post). Existing validity values are kept.

---

### Impact on other products:

- **AI chat:** can use the same data to answer "what's wrong with my workspace?"
- **Public API and MCP server:** a workspace health check would help agents and automations. Open question for the PO. If yes, a story per surface.

---

### Dependencies:

**[BE] Check connect permission and the return address before sending a user to a platform** (the reconnect action the Health Center reuses).

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness: N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support: N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (any new or changed API is reflected in the public API, CLI, MCP server and automation apps, N/A when nothing API-facing changes)

---

# [BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center

### Description

As a ContentStudio user, I want every post that didn't go out to be recorded with the reason, what we tried and what to do next, and to see my delivery rate over time, so that I can fix what's broken and see whether things are getting better.

Failed posts and their raw platform errors exist today, but there's no grouping by reason, no record of retries, no "preventable" count and no delivery-rate metric. This story adds them and returns them for the Overview and Post delivery tabs.

**Skeleton.** Details to follow the tech lead discussion.

---

### Workflow

1. User opens the Health Center.
2. The Overview shows the 7-day delivery rate and trend, posts published and failed, failures by reason and by platform, and failures per day.
3. The Post delivery tab lists every post that didn't go out, and each one explains what happened.

---

### Acceptance criteria

**Per post**

- [ ] Every post that didn't go out on time has a status: **Failed**, **On hold** or **Retrying**
- [ ] Each has a reason category: expired or revoked access, paused by the platform (posting limit), media rejected, account removed from the workspace, location or account not found, or platform outage after retries
- [ ] Each has a plain "What happened" and "How to fix it", and the two actions to offer
- [ ] Each has a timeline (scheduled time reached, platform response, retry decision, who was notified), the original platform error code, attempts (for example "1 of 1 (not retryable)" or "2 of 5") and a support reference
- [ ] The list is filterable by status and exportable as CSV for a chosen date range

**Metrics**

- [ ] Delivery rate for the last 7 days and the change from the previous 7 days
- [ ] Posts published, posts failed, and how many failures were preventable (expired access or removed accounts)
- [ ] Delivery rate and failed count per platform
- [ ] Failures by reason
- [ ] Failed posts and delivery rate per day

**Issues list**

- [ ] Open issues are grouped (one issue per cause, not one per post) with severity **Critical**, **Warning** or **Info**, the affected accounts, the number of posts affected and when the next one is due
- [ ] Issues are sorted by which posts will fail soonest
- [ ] An issue clears on its own when its cause is fixed

**Heart score**

- [ ] The workspace health state for the rail heart is returned: **full** (no critical issues and 7-day delivery rate of 95% or more), **half** (any warning, or delivery rate from 80% to under 95%) or **empty** (any critical issue, or delivery rate under 80%)

---

### Mock-ups:

None. Backend. Reference: the Overview and Post delivery tabs in the CTO prototype: https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

---

### Impact on existing data:

Adds a record of delivery attempts and issue history. Existing post and error data is kept.

---

### Impact on other products:

- **Dashboard:** the existing failed-posts card should use the same numbers.
- **Public API and MCP server:** open question, as for account health.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness: N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support: N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (any new or changed API is reflected in the public API, CLI, MCP server and automation apps, N/A when nothing API-facing changes)

---

# [Research] Decide how ContentStudio should protect accounts automatically when platforms limit or break posting

### Description

As the ContentStudio team, we want to decide whether and how the system should act on its own when a platform limits or breaks posting, so that users lose fewer posts and accounts don't get restricted further, without ContentStudio doing anything surprising.

The CTO prototype assumes several automatic behaviours that don't exist today. Each needs a decision before it's built:

- **Pause a page that hit a platform posting limit** (for example Facebook limiting a page after the same post went to many pages at once) until the limit lifts, and hold its posts instead of failing them.
- **Stop posting to a Google Business location that isn't found**, so it doesn't fail every day.
- **Flag scheduled posts that still target a removed account**, and let the user pick another account or remove it from the post.
- **Retry temporary platform errors** up to 5 times with increasing waits, and never retry errors that can't succeed.
- **Detect a teammate's phone that stopped accepting reminders** for Facebook group posts.

This is an open question. The outcome is a written recommendation, not code.

---

### Workflow

1. Researcher reviews each behaviour against what the platforms allow and how ContentStudio posts today.
2. Researcher checks the real error codes (for example Facebook 368, Google 404 NOT_FOUND, Pinterest 503) and how often each happens.
3. Researcher recommends, for each behaviour, whether to do it, how, and what the user sees.
4. The PO and tech lead decide, and build stories follow.

---

### Acceptance criteria

- [ ] A written recommendation for each of the five behaviours: do it or not, how it would work, what the user sees, and the risks
- [ ] Data on how often each case happens today, from existing publishing errors
- [ ] Which platform error codes map to "retry", "hold" and "stop"
- [ ] A proposed retry policy (how many tries, how long between them)
- [ ] Reviewed with the PO and tech lead, with the agreed build stories listed

---

### Mock-ups:

None. Reference: the "Needs your attention" examples in the CTO prototype: https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

---

### Impact on existing data:

None. Research only.

---

### Impact on other products:

Decisions here change how publishing behaves for everyone, on web and mobile.

---

### Dependencies:

None. Informs **[BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness: N/A, research only
- [ ] Multilingual support: N/A, research only
- [ ] UI theming support: N/A, research only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, research only

---

# [BE] Send health alerts by the workspace's alert settings

### Description

As a workspace owner, I want to be told about problems once, through the channels I chose, and by the right people, so that issues get fixed quickly without my team drowning in one email per failed post.

This story sends health alerts in-app, by email and by mobile push, following the workspace's alert settings. It always alerts once per issue, never once per post.

**Skeleton.** Details to follow the tech lead discussion. It should follow the **Notifications architecture** epic.

---

### Workflow

1. An account's access expires. ContentStudio sends one "Account needs reconnecting" alert, listing the posts it affects, to the people and channels the settings say.
2. Several posts fail for the same reason. The user gets one alert for the issue, or one hourly or daily digest email if they chose that.
3. During quiet hours, email and push wait. Critical alerts still show in the app.
4. Every Monday at 09:00 the weekly health summary goes out.

---

### Acceptance criteria

- [ ] Alert types: Account needs reconnecting (critical), Platform paused posting, Post failed (only failures that need the user; temporary errors are retried first), Access expiring soon (7 days and 1 day before), Mobile reminder not delivered, Post published (grouped into one per hour), Weekly health summary (Mondays 09:00, workspace time zone)
- [ ] Channels: in-app, email and mobile push only. **No Slack**
- [ ] In-app alerts for accounts that need reconnecting can't be turned off
- [ ] One alert per issue, never one per post
- [ ] Recipients follow the settings: workspace owner (always gets critical alerts), admins, whoever connected the account, whoever scheduled an affected post (only for their own posts), clients with approval access
- [ ] Email for failed posts follows the chosen frequency: as they happen, hourly digest or daily digest at 09:00. Reconnect alerts are always sent straight away
- [ ] Quiet hours delay email and push until they end. Critical alerts still appear in the app
- [ ] Alerts stop once the issue is fixed

---

### Mock-ups:

None. Backend. Reference: the Alert settings tab in the CTO prototype, with Slack removed: https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

---

### Impact on existing data:

Adds alert settings per workspace and a record of alerts sent.

---

### Impact on other products:

- **Existing expired-account emails** are replaced by these alerts, so users don't get both.
- **Mobile app:** receives push alerts (see **[Flutter] Receive Health Center alerts as push notifications**).

---

### Dependencies:

- **[BE] Track account health: status, access expiry and cause, last successful post and scheduled posts**
- **[BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center**
- The **Notifications architecture** epic's research

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness: N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support: N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (any new or changed API is reflected in the public API, CLI, MCP server and automation apps, N/A when nothing API-facing changes)

---

# [FE] Build the Health Center Overview tab

### Description

As a ContentStudio user, I want one screen that tells me how my publishing is doing and what will fail next, so that I can fix the most urgent problem first.

**Skeleton.** Final layout and copy from the design story.

---

### Workflow

1. User opens Health from the rail and lands on Overview.
2. They see the header "Health Center / Everything that could stop a post from going out, in one place." with when it was last updated, and tabs for Overview, Accounts, Post delivery and Alert settings.
3. They see four numbers: delivery rate, posts published, posts failed and healthy accounts.
4. Under **Needs your attention** they see the issues, most urgent first, and click **Reconnect accounts** on the critical one.
5. Lower down they see delivery rate by platform, why posts failed, and failures per day.

---

### Acceptance criteria

- [ ] Header, last-updated time, **Alert settings** and **Compose Post** buttons, and the four tabs with counts ("6 need action", "47 failed")
- [ ] Four cards: Post delivery rate (last 7 days, with change from the previous 7 days), Posts published, Posts failed (with the preventable count), Healthy accounts (with to reconnect, paused and not found)
- [ ] **Needs your attention** lists open issues sorted by which posts fail soonest, with Critical, Warning and Info counts. Each issue shows severity, explanation, affected accounts, impact and two actions
- [ ] Reconnect actions use the reconnect flow from **[FE] Show "Reconnect required" with a Reconnect now button on every expired account**
- [ ] Delivery rate by platform, Why posts failed (with **See failed posts**) and Failed posts per day
- [ ] All-clear state when there are no issues: "Everything's running smoothly"
- [ ] Loading and error states as designed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**. CTO prototype: https://claude.ai/artifact/TYrrHhHuzeNtujZtnADRSQ

---

### Impact on existing data:

None.

---

### Impact on other products:

White-label domains must never show ContentStudio by name here.

---

### Dependencies:

- **[BE] Track account health: status, access expiry and cause, last successful post and scheduled posts**
- **[BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center**
- **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Build the Health Center Accounts tab

### Description

As a ContentStudio user, I want to see every connected account, whether it can post right now and when its access runs out, so that I can reconnect or renew before posts fail.

**Skeleton.** Final layout and copy from the design story.

---

### Workflow

1. User opens the Accounts tab.
2. They see "Every connected account, whether it can post right now, and when its access runs out." with **Connect account** and **Reconnect all**.
3. They filter to **Reconnect required** and reconnect an account. It moves to Healthy and its posts are kept.

---

### Acceptance criteria

- [ ] Filters with counts: All, Reconnect required, Paused by platform, Not found, Expiring in 14 days, Healthy
- [ ] The note "Reconnecting keeps your scheduled posts. They go out on time as soon as access is restored."
- [ ] Columns: Account (name, platform, type), Status, Access (expiry and cause), Last successful post, Scheduled, Actions
- [ ] Actions per status: Reconnect, View schedule, Check location, Renew access, Details
- [ ] **Reconnect all** reconnects every account that needs it, one after another
- [ ] Accounts that need action are listed first. The list is paged
- [ ] Empty, loading and error states as designed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Settings > Social Accounts stays as it is.

---

### Dependencies:

- **[BE] Track account health: status, access expiry and cause, last successful post and scheduled posts**
- **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Build the Health Center Post delivery tab with issue details and CSV export

### Description

As a ContentStudio user, I want to see every post that didn't go out, why, and what to do about it, so that I can fix each one without guessing and give support the details they need.

**Skeleton.** Final layout and copy from the design story.

---

### Workflow

1. User opens Post delivery and sees "Every post that didn't go out, why, and what to do about it."
2. They filter to **Failed** and click a post.
3. The detail panel shows the status, What happened, How to fix it, two actions, the timeline and technical details.
4. They click **Reconnect account**, and the post goes out after reconnecting.
5. They click **Export CSV** to share the list.

---

### Acceptance criteria

- [ ] Date range (default last 7 days) and **Export CSV**
- [ ] Filters with counts: All issues, Failed, On hold, Retrying, Published
- [ ] Rows: post text, account and platform, scheduled time, reason
- [ ] Detail panel: status, account, time, post text, What happened, How to fix it, primary and secondary actions, timeline, and "Technical details for support" (platform error, attempts, reference)
- [ ] Actions work: reconnect, retry, edit and retry, move to a later slot, reschedule, check location, remove location, cancel retries
- [ ] Paged list ("Load more")
- [ ] Empty, loading and error states as designed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**.

---

### Impact on existing data:

None.

---

### Impact on other products:

The Planner's failed filter stays. It can link here.

---

### Dependencies:

- **[BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center**
- **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Build Health Center alert settings

### Description

As a workspace owner, I want to choose who hears about problems, where and how often, so that the right people fix issues fast without alert noise.

**Skeleton.** Final layout and copy from the design story. No Slack.

---

### Workflow

1. Owner opens Alert settings: "Choose who hears about problems, where, and how often. We alert once per issue, never once per post."
2. They turn email on for "Platform paused posting" and set failed-post emails to an hourly digest.
3. They add "Whoever connected the account" as a recipient and set quiet hours from 22:00 to 07:00.
4. They click **Save changes**.

---

### Acceptance criteria

- [ ] A table of alert types against In-app, Email and Mobile push, with a short description of each alert. No Slack column
- [ ] In-app for "Account needs reconnecting" is always on and can't be turned off, with the note explaining why
- [ ] Who gets health alerts: Workspace owner (always gets critical alerts), Admins, Whoever connected the account, Whoever scheduled an affected post, Clients with approval access
- [ ] Email frequency for failed posts: As they happen, Hourly digest (recommended), Daily digest at 09:00 (shows the workspace time zone)
- [ ] Quiet hours with From and To, and the note that critical alerts still appear in the app
- [ ] **Save changes** and **Cancel**, with a success toast and an unsaved-changes warning
- [ ] Only users allowed to manage workspace settings can change these
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**.

---

### Impact on existing data:

None beyond saving the settings.

---

### Impact on other products:

Existing notification preferences stay. The Notifications architecture epic decides how the two join up.

---

### Dependencies:

- **[BE] Send health alerts by the workspace's alert settings**
- **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Add category tabs to the notification panel and a health banner on Home

### Description

As a ContentStudio user, I want my notifications sorted into health, publishing, approvals and mentions, and a clear banner on Home when accounts need fixing, so that urgent problems don't get lost among routine updates.

**Skeleton.** Final layout and copy from the design story.

---

### Workflow

1. User opens the notification bell. They see tabs: All, Health, Publishing, Approvals, Mentions, each with a count.
2. They open Health and see "3 accounts need reconnecting" with **Reconnect accounts**.
3. On Home, a banner reads "3 accounts need reconnecting / 11 scheduled posts will fail until you do. The next one is in 2h 14m." with **Open Health Center** and **Reconnect**.

---

### Acceptance criteria

- [ ] The notification panel has tabs All, Health, Publishing, Approvals and Mentions, with counts
- [ ] Notifications are grouped Today and Earlier. Each shows an icon, title, body, time, category and one action
- [ ] **Mark all as read** and **Open Health Center** in the panel
- [ ] The Home banner shows when any critical health issue is open, with the number of affected posts and when the next one is due, and disappears when fixed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Existing expired-token banners should give way to the Home health banner, so users don't see two. Coordinate with **[FE] Rewrite expired-token banners and alerts to match, and reconnect the right account**.

---

### Dependencies:

- **[BE] Send health alerts by the workspace's alert settings**
- **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Add the Health item to the rail with a heart that fills with workspace health

### Description

As a ContentStudio user, I want a Health item in the rail whose heart icon shows at a glance how healthy my workspace is, so that I notice problems without opening anything.

---

### Workflow

1. User sees **Health** in the desktop rail with a full heart: everything is fine.
2. An account's access expires. The heart turns empty.
3. User hovers it and reads the tooltip, then clicks and lands on the Health Center.
4. After reconnecting, the heart fills again.

---

### Acceptance criteria

- [ ] A **Health** item in the desktop rail opens the Health Center
- [ ] The heart icon is **full** when there are no critical issues and the 7-day delivery rate is 95% or more
- [ ] It is **half** when there is any warning, or the delivery rate is from 80% to under 95%
- [ ] It is **empty** when there is any critical issue (for example an account that needs reconnecting), or the delivery rate is under 80%
- [ ] The icon updates without a page reload when the state changes
- [ ] Tooltip by state: full "Everything's running smoothly", half "A few things need a look", empty "Something is stopping posts from going out"
- [ ] The state is also exposed to screen readers, not shown by the icon alone
- [ ] Works with Customize sidebar (can be moved or hidden) and every rail theme
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Should be placed with the **App shell refresh** epic's rail layout in mind.

---

### Dependencies:

- **[BE] Track post delivery issues, retries and delivery-rate metrics for the Health Center** (heart state)
- **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [Flutter] Receive Health Center alerts as push notifications

### Description

As a ContentStudio mobile user, I want health alerts on my phone, like an account that needs reconnecting, so that I can fix problems even when I'm away from my desk.

**Skeleton.** Follows the **Notifications architecture** epic.

---

### Workflow

1. An account's access expires. The user's phone shows "3 accounts need reconnecting / 11 scheduled posts will fail until you do."
2. User taps it. The app opens the account list, and the user reconnects from the phone.

---

### Acceptance criteria

- [ ] Health alerts the user turned on for mobile push arrive on iOS and Android
- [ ] Tapping an alert opens the right place in the app (the accounts list for reconnect alerts, the post for a failed post), switching workspace if needed
- [ ] Quiet hours are respected, except critical alerts where the settings say so
- [ ] Alerts that are already fixed don't arrive late
- [ ] Copy is translated in every supported language

---

### Mock-ups:

Push copy from **[Design] Finalise the Health Center, notification panel and heart rail icon from the CTO prototype**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Adds a new notification type to the app, alongside the manual-publish reminder and AI chat notifications.

---

### Dependencies:

- **[BE] Send health alerts by the workspace's alert settings**
- The **Notifications architecture** epic

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes
