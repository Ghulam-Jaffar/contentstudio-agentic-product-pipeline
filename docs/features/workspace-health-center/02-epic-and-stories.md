# Epic: Workspace Health Center

## Epic description

When something goes wrong in a ContentStudio workspace, the user finds out piece by piece, if at all. An expired Facebook token shows up as a banner in one place and a warning icon in another. A post that Facebook rejected sits in the Planner under a "failed" filter. An account whose posts keep failing only shows if the user goes looking. Nothing puts it all in one place, and nothing tells the user plainly what needs fixing.

The **Health Center** is a new area that shows everything that could stop a post from going out, in one place. It has three tabs:

- **Overview**: four headline numbers (delivery rate, posts published, posts failed, healthy accounts), a **Needs your attention** list, delivery rate by platform, and failed and published posts per day.
- **Accounts**: one row per account that needs reconnecting, whose last post failed, or that is healthy, with filters and a details view of its errors and failed posts.
- **Post delivery**: one entry per account per failed post, with the error log, the timeline and the right action: reconnect or retry.

It also adds a **Health** item to the desktop rail with a **heart icon that fills with the workspace's health**.

**Design canvas:** https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26. Updated on 2026-10-02 to the scope the team agreed on 2026-10-01, starting from the CTO's original prototype.

### Heart icon

| Heart | When |
|---|---|
| Full | No critical issues and a 7-day delivery rate of 95% or more |
| Half | Any warning, or a delivery rate from 80% to under 95% |
| Empty | Any critical issue (for example an account that needs reconnecting), or a delivery rate under 80% |

### Scope

In:

- Account health and post delivery data, and the numbers behind the Overview
- The Overview, Accounts and Post delivery tabs
- The Health item and heart icon in the desktop rail
- The same Health Center in the mobile app
- A research story on how ContentStudio should handle broken accounts and failures automatically

Out:

- **Alert settings and health notifications.** In-app and email notifications already exist and are managed in notification settings. No alert settings tab, no new in-app, email or push alerts, no notification panel changes
- **A new Home banner.** The existing reconnect banner at the bottom of the app already tells users when accounts need reconnecting
- **CSV export** of failed posts (not for now)
- **"Paused by platform".** Platforms don't tell us when they limit an account, so it isn't shown anywhere
- **"Why posts failed" by reason.** We don't classify failure reasons yet. Post delivery shows the raw error log
- **Slack**
- Building a second reconnect flow. The Health Center reuses the one from **Reconnect and connect accounts from wherever they're used**
- Inbox, analytics and automation health (later)

### Stories

1. `[Design] Finalise the Health Center tabs and heart rail icon from the design canvas`
2. `[BE] Track account health: reconnect status, consecutive failed posts, access expiry and error log`
3. `[BE] Track per-account post failures and delivery-rate metrics for the Health Center`
4. `[Research] Decide how ContentStudio should protect accounts automatically when platforms limit or break posting`
5. `[FE] Build the Health Center Overview tab`
6. `[FE] Build the Health Center Accounts tab`
7. `[FE] Build the Health Center Post delivery tab with per-account error logs`
8. `[FE] Add the Health item to the rail with a heart that fills with workspace health`
9. `[Flutter] Bring the Health Center to the mobile app`

---

# [Design] Finalise the Health Center tabs and heart rail icon from the design canvas

### Description

As a designer, I want to turn the Health Center design canvas into final designs for the scope the team agreed, so that devs build the three tabs and the heart rail icon from one agreed reference.

The design canvas already reflects the scope the team agreed: no alert settings, no notification panel changes, no "Paused by platform" state and no "Why posts failed" widget. It shows per-account rows that count consecutive failures, per-account post delivery entries with the raw error log, a failed or published posts per day widget with a toggle, and the heart states. This story turns it into final, handed-off designs.

---

### Workflow

1. Designer reviews the design canvas and the scope in this epic.
2. Designer finalises the Overview tab: four headline numbers, Needs your attention, delivery rate by platform, and posts per day with a Failed and Published toggle.
3. Designer finalises the Accounts tab: one row per account, the "Last 2 posts failed" pattern, filters, search and the details view.
4. Designer finalises the Post delivery tab: the list with the first entry selected, and the side panel with the error log, timeline and action.
5. Designer designs the Health rail item with a heart icon in three states: full, half and empty.
6. Designer covers empty, loading and error states for every tab.

---

### Acceptance criteria

- [ ] Final designs for the Overview, Accounts and Post delivery tabs. No Alert settings tab
- [ ] No "Paused by platform" status, filter or issue anywhere, and no "Why posts failed" widget
- [ ] The posts-per-day widget switches between **Failed** and **Published** with a `SegmentedControl`
- [ ] Accounts rows show repeated failures on one account as a single row ("Last 2 posts failed"), never as extra rows
- [ ] Post delivery shows one entry per account per failed post, and the side panel with error log, timeline and either **Reconnect** plus a help link, or **Retry**
- [ ] The Health rail item with the heart icon in full, half and empty states, its tooltip, and how it looks in every rail theme
- [ ] Severity styles for Critical, Warning and Info that stay readable without colour alone
- [ ] Empty ("Everything's running smoothly"), loading and error states for every tab
- [ ] Every element is mapped to an existing `@contentstudio/ui` component, or flagged as a gap
- [ ] Phone layouts of the three tabs and the health indicator for the mobile app
- [ ] Designs are handed off to every `[FE]` and `[Flutter]` story in this epic

---

### Mock-ups:

Design canvas, updated to the agreed scope: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

None.

---

### Impact on other products:

Web app, plus the mobile screens for **[Flutter] Bring the Health Center to the mobile app**.

---

### Dependencies:

None. Blocks every `[FE]` and `[Flutter]` story in this epic.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, design only, nothing API-facing changes

---

# [BE] Track account health: reconnect status, consecutive failed posts, access expiry and error log

### Description

As a ContentStudio user, I want the app to know, for every connected account, whether it needs reconnecting, whether its recent posts are failing, and when its access runs out, so that the Health Center shows me which accounts to fix first.

Account validity is tracked today (valid, expired, expiring soon, invalid), but not how many posts in a row have failed on an account, its error log, or the exact expiry. This story adds that and returns it for the Accounts tab and the Overview.

---

### Workflow

1. User opens the Health Center Accounts tab.
2. Accounts that need reconnecting come first.
3. Next come accounts whose last post failed. An account whose last 2 posts failed shows once, as "Last 2 posts failed", not as two entries.
4. Then the healthy accounts.
5. User opens an account's details and sees its error log and the posts that failed on it.

---

### Acceptance criteria

**Statuses**

- [ ] Each account has one status: **Reconnect required**, **Not found** (for example a Google Business location that no longer exists), **Last post failed**, **Expiring soon** (within 14 days) or **Healthy**
- [ ] There is no "Paused by platform" status
- [ ] Order: Reconnect required first, then Not found, then Last post failed, then Expiring soon, then Healthy

**Consecutive failures**

- [ ] For an account whose most recent post failed, the number of failures in a row since its last successful post is returned (for example 2 for "Last 2 posts failed")
- [ ] A new failure on the same account increases that number on the same account. It never creates a second entry
- [ ] A successful post on that account resets it to Healthy

**Details**

- [ ] Each account returns its access expiry date (or "no expiry") and, where known, the cause in plain words (for example "Password changed on Facebook", "LinkedIn access lasts 60 days")
- [ ] Each account returns its last successful post time and number of scheduled posts
- [ ] An account's details return its error log (each failed attempt with time, post and the platform's error message) and the posts that failed on it, newest first, paged

**Filters**

- [ ] The list can be searched by account name and filtered by platform and by status, with counts per status
- [ ] Counts for the Overview: healthy accounts out of total, and how many need reconnecting or are not found
- [ ] Only accounts the requesting user can access are returned
- [ ] Reconnecting an account moves it out of Reconnect required straight away, and its scheduled posts go out as planned

---

### Mock-ups:

None. Backend. Reference: the Accounts tab on the design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

Adds health fields to connected accounts (consecutive failure count, expiry, last successful post). Existing validity values are kept.

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

# [BE] Track per-account post failures and delivery-rate metrics for the Health Center

### Description

As a ContentStudio user, I want every failed post to be recorded per account, with the platform's error and what to do next, and to see my delivery rate over time, so that I can fix each failure and see whether things are getting better.

A post going to 10 accounts can fail on 5 of them for different reasons, and each of those needs its own retry or reconnect. So this story records failures **per account per post**, not per post. It also adds the numbers behind the Overview.

We don't classify failure reasons yet. Each failure carries the platform's raw error and one of two actions: **reconnect** when access expired or was revoked, or **retry** for any other posting failure.

---

### Workflow

1. A post goes to 10 accounts and fails on 5.
2. The Post delivery tab shows 5 entries, one per account, each with its own error and action.
3. On one entry, access had expired, so it offers Reconnect. On another, the platform rejected the post, so it offers Retry.
4. The Overview shows the delivery rate, published and failed counts, delivery rate by platform, and failed and published posts per day.

---

### Acceptance criteria

**Per account per post**

- [ ] Each account a post failed on is its own entry: post, account, platform, scheduled time and status Failed
- [ ] Each entry has the platform's error message and code as returned, the number of attempts, and a timeline (scheduled time reached, platform response, retries)
- [ ] Each entry has one action type: **reconnect** when the failure was expired or revoked access, otherwise **retry**
- [ ] Retrying an entry retries only that account. A successful retry removes the entry from the failed list and counts as published
- [ ] Entries can be filtered by platform and account, searched by post text, and are returned newest first, paged

**Metrics**

- [ ] Delivery rate for the last 7 days and the change from the previous 7 days, counted per account per post
- [ ] Posts published and posts failed in the last 7 days, and how many failures were on accounts that needed reconnecting or were removed (preventable)
- [ ] Delivery rate and failed count per platform
- [ ] Failed posts per day and published posts per day for the last 7 days

**Needs your attention**

- [ ] Open issues, one per cause, each with severity, the affected accounts, the number of posts affected and when the next one is due:
    - Accounts that need reconnecting (Critical)
    - Scheduled posts that still use a removed account (Warning)
    - A Google Business location that wasn't found (Info)
    - A teammate isn't getting mobile reminders (Info)
- [ ] No "paused by platform" issue
- [ ] Issues are sorted by which posts will fail soonest, and clear on their own when the cause is fixed

**Heart score**

- [ ] The workspace health state for the rail heart is returned: **full** (no critical issues and 7-day delivery rate of 95% or more), **half** (any warning, or delivery rate from 80% to under 95%) or **empty** (any critical issue, or delivery rate under 80%)

---

### Mock-ups:

None. Backend. Reference: the Overview and Post delivery tabs on the design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

Adds a per-account record of failed attempts. Existing post and error data is kept.

---

### Impact on other products:

- **Dashboard:** the existing failed-posts card should use the same numbers.
- **Planner:** a post that failed on some accounts still shows as partly failed there.
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

As the ContentStudio team, we want to decide whether and how the system should act on its own when posting breaks, so that users lose fewer posts and accounts don't get restricted further, without ContentStudio doing anything surprising.

The CTO's original prototype assumed behaviours that don't exist today. Each needs a decision before it's built:

- **Stop posting to a Google Business location that isn't found**, so it doesn't fail every day.
- **Flag scheduled posts that still target a removed account**, and let the user pick another account or remove it from the post.
- **Retry temporary platform errors** a set number of times with increasing waits, and never retry errors that can't succeed.
- **Detect a teammate's phone that stopped accepting reminders** for Facebook group posts.
- **Pause a page that hit a platform posting limit.** Platforms don't notify us of this today, so it's out of the Health Center for now. The research should say whether there is any reliable way to detect it.
- **Classify failure reasons** (expired access, media rejected, platform outage and so on) so a "Why posts failed" view could come later.

This is an open question. The outcome is a written recommendation, not code.

---

### Workflow

1. Researcher reviews each behaviour against what the platforms allow and how ContentStudio posts today.
2. Researcher checks the real error codes (for example Facebook 368, Google 404 NOT_FOUND, Pinterest 503) and how often each happens.
3. Researcher recommends, for each behaviour, whether to do it, how, and what the user sees.
4. The PO and tech lead decide, and build stories follow.

---

### Acceptance criteria

- [ ] A written recommendation for each behaviour above: do it or not, how it would work, what the user sees, and the risks
- [ ] Data on how often each case happens today, from existing publishing errors
- [ ] Which platform error codes mean "reconnect", "retry" and "stop"
- [ ] A proposed retry policy (how many tries, how long between them)
- [ ] Whether posting limits can be detected reliably on any platform
- [ ] Reviewed with the PO and tech lead, with the agreed build stories listed

---

### Mock-ups:

None. Reference: the "Needs your attention" examples on the design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

None. Research only.

---

### Impact on other products:

Decisions here change how publishing behaves for everyone, on web and mobile.

---

### Dependencies:

None. Informs **[BE] Track per-account post failures and delivery-rate metrics for the Health Center**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness: N/A, research only
- [ ] Multilingual support: N/A, research only
- [ ] UI theming support: N/A, research only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, research only

---

# [FE] Build the Health Center Overview tab

### Description

As a ContentStudio user, I want one screen that tells me how my publishing is doing and what needs fixing, so that I can deal with the most urgent problem first.

---

### Workflow

1. User opens Health from the rail and lands on Overview.
2. They see the header "Health Center / Everything that could stop a post from going out, in one place." with when it was last updated, and tabs for Overview, Accounts and Post delivery.
3. They see four numbers: delivery rate, posts published, posts failed and healthy accounts.
4. Under **Needs your attention** they see the open issues, most urgent first, and click **Reconnect accounts** on the critical one.
5. Lower down they see delivery rate by platform, and a posts-per-day chart. They switch it from **Failed** to **Published** to compare.

---

### Acceptance criteria

**Header and tabs**

- [ ] Header with the last-updated time and a **Compose Post** button
- [ ] Tabs: Overview, Accounts (with a count of accounts that need action) and Post delivery (with a count of failed posts). No Alert settings tab

**Top cards**

- [ ] Post delivery rate, last 7 days, with the change from the previous 7 days ("Up 0.8 points on the previous 7 days")
- [ ] Posts published, with "Across {n} connected accounts"
- [ ] Posts failed, with "{n} were preventable: expired access or removed accounts"
- [ ] Healthy accounts, "{healthy} of {total}", with how many need reconnecting and how many are not found

**Needs your attention**

- [ ] Lists open issues sorted by which posts fail soonest, with Critical, Warning and Info counts. Each shows severity, explanation, affected accounts, impact and two actions
- [ ] Shows only these issue types: accounts that need reconnecting, scheduled posts that still use a removed account, a Google Business location that wasn't found, and a teammate not getting mobile reminders
- [ ] No "paused by platform" issue
- [ ] Reconnect actions use the reconnect flow from **[FE] Show "Reconnect required" with a Reconnect now button on every expired account**

**Charts**

- [ ] Delivery rate by platform for the last 7 days, with the failed count per platform
- [ ] One posts-per-day chart for the last 7 days with a `SegmentedControl` to switch between **Failed** and **Published**. Default: Failed. Today is marked "Today is still in progress"
- [ ] No "Why posts failed" widget

**States**

- [ ] All-clear state when there are no issues: "Everything's running smoothly"
- [ ] Loading and error states as designed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**. Design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

None.

---

### Impact on other products:

White-label domains must never show ContentStudio by name here.

---

### Dependencies:

- **[BE] Track account health: reconnect status, consecutive failed posts, access expiry and error log**
- **[BE] Track per-account post failures and delivery-rate metrics for the Health Center**
- **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**

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

As a ContentStudio user, I want to see which accounts need reconnecting, which ones' posts are failing, and which are healthy, so that I can fix the right account before more posts fail.

---

### Workflow

1. User opens the Accounts tab and sees "Every connected account, whether it can post right now, and when its access runs out." with **Connect account** and **Reconnect all**.
2. At the top are accounts that need reconnecting. Next are accounts whose last post failed. One shows "Last 2 posts failed".
3. Then the healthy accounts.
4. User filters to Instagram and searches "Bakery".
5. User clicks **Details** on an account and sees its error log and the posts that failed on it.
6. User reconnects an account and it moves out of Reconnect required.

---

### Acceptance criteria

**List**

- [ ] One row per account, never more than one. Order: Reconnect required, Not found, Last post failed, Expiring soon, Healthy
- [ ] An account with repeated failures shows the count on its single row: "Last post failed", "Last 2 posts failed", "Last 3 posts failed" and so on
- [ ] Columns: Account (name, platform, type), Status, Access (expiry and cause), Last successful post, Scheduled, Actions
- [ ] The note "Reconnecting keeps your scheduled posts. They go out on time as soon as access is restored."
- [ ] No "Paused by platform" status or filter

**Filters**

- [ ] Search by account name, with placeholder "Search accounts"
- [ ] Platform filter (`Dropdown`)
- [ ] Status filter with counts: All, Reconnect required, Not found, Last post failed, Expiring soon, Healthy

**Actions**

- [ ] Reconnect required: **Reconnect**. Not found: **Check location**. Last post failed: **Details**. Expiring soon: **Renew access**. Healthy: **Details**
- [ ] **Reconnect all** reconnects every account that needs it, one after another
- [ ] **Details** opens the account's error log (each failed attempt with time, post and the platform's error message) and the posts that failed on it, newest first. From a failed post, the user can retry it or open it in Post delivery

**States**

- [ ] Paged list
- [ ] Empty (no accounts match the filters: "No accounts match these filters"), loading and error states as designed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**. Design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

None.

---

### Impact on other products:

Settings > Social Accounts stays as it is.

---

### Dependencies:

- **[BE] Track account health: reconnect status, consecutive failed posts, access expiry and error log**
- **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Build the Health Center Post delivery tab with per-account error logs

### Description

As a ContentStudio user, I want to see every post that failed, account by account, with the platform's error and one clear action, so that I can reconnect or retry each one without guessing.

A post going to 10 accounts can fail on 5 of them. Each of those 5 is its own entry here, because each needs its own reconnect or retry.

---

### Workflow

1. User opens Post delivery and sees "Every post that didn't go out, and what to do about it."
2. The first entry is selected, and its details show in the side panel.
3. Its error says access expired. The panel offers **Reconnect** and a "Why did this happen?" help link.
4. User clicks another entry. The platform rejected that post, so the panel offers **Retry**.
5. User retries it. It goes out, and the entry leaves the list.

---

### Acceptance criteria

**List**

- [ ] One entry per account per failed post. A post that failed on 5 of its 10 accounts shows 5 entries
- [ ] Each entry: post text, account and platform, scheduled time
- [ ] Newest first, with "Load more"
- [ ] Search by post text, and filters by platform and account

**Side panel**

- [ ] The first entry is selected by default and its details show in the side panel. Clicking another entry shows its details
- [ ] The panel shows the post text, account, scheduled time, the platform's error log as returned, and a timeline (scheduled time reached, platform response, each retry)
- [ ] No "What happened" or "How to fix it" explanations, and no failure-reason categories
- [ ] Access expired or revoked: **Reconnect** (primary) and a "Why did this happen?" link to the help article
- [ ] Any other failure: **Retry** (primary). Retrying shows "Retrying..." and then either removes the entry with "Posted to {account}" or shows the new error

**States**

- [ ] Empty: "No failed posts in the last 7 days"
- [ ] Loading and error states as designed
- [ ] All copy comes from translation keys

---

### Mock-ups:

From **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**. Design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

None.

---

### Impact on other products:

The Planner's failed filter stays. It can link here.

---

### Dependencies:

- **[BE] Track per-account post failures and delivery-rate metrics for the Health Center**
- **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**

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

From **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**. Heart states on the design canvas: https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26

---

### Impact on existing data:

None.

---

### Impact on other products:

Should be placed with the **App shell refresh** epic's rail layout in mind.

---

### Dependencies:

- **[BE] Track per-account post failures and delivery-rate metrics for the Health Center** (heart state)
- **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [Flutter] Bring the Health Center to the mobile app

### Description

As a ContentStudio mobile user, I want the same Health Center in the app as on the web, so that I can see which accounts need fixing and which posts failed, and reconnect or retry them from my phone.

This story brings the web Health Center to the mobile app with the same data and the same rules: the Overview, Accounts and Post delivery tabs, and the heart that shows workspace health. It uses the same backend as web, so nothing new is needed on the server. No push alerts: notifications already exist.

---

### Workflow

1. User opens the app menu and sees **Health** with a heart icon. The heart is empty because an account needs reconnecting.
2. User taps it. The Overview shows the four headline numbers and Needs your attention.
3. User taps **Reconnect accounts** and reconnects from the phone.
4. User opens Accounts and sees "Last 2 posts failed" on one account. They tap it and see its error log and failed posts.
5. User opens Post delivery, taps a failed entry, reads the error and taps **Retry**. The post goes out.

---

### Acceptance criteria

**Entry**

- [ ] A **Health** item in the app menu opens the Health Center, with the heart in full, half or empty state using the same rules as web
- [ ] The heart state updates when the app comes back to the foreground

**Overview**

- [ ] The four headline numbers: delivery rate (7 days, with change), posts published, posts failed (with preventable count), healthy accounts
- [ ] Needs your attention with the same issue types as web: accounts that need reconnecting, posts on a removed account, a Google Business location not found, a teammate not getting mobile reminders. Each with its actions
- [ ] Delivery rate by platform, and posts per day with a Failed and Published toggle

**Accounts**

- [ ] One row per account in the same order as web, with "Last {n} posts failed" on a single row for repeated failures
- [ ] Search, platform filter and status filter
- [ ] Tapping an account opens its error log and failed posts. Reconnect works from the app

**Post delivery**

- [ ] One entry per account per failed post, newest first
- [ ] Tapping an entry opens its error log and timeline, with **Reconnect** and the help link for expired access, or **Retry** for other failures

**General**

- [ ] No "Paused by platform", no "Why posts failed", no alert settings, no CSV export, matching web
- [ ] Empty ("Everything's running smoothly"), loading, offline and error states
- [ ] Works on iOS and Android, on small and large phones
- [ ] All copy is translated in every supported language

---

### Mock-ups:

Phone layouts from **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**.

---

### Impact on existing data:

None. Uses the same data as web.

---

### Impact on other products:

Matches the web Health Center. Any later change to the web tabs should be mirrored here.

---

### Dependencies:

- **[BE] Track account health: reconnect status, consecutive failed posts, access expiry and error log**
- **[BE] Track per-account post failures and delivery-rate metrics for the Health Center**
- **[Design] Finalise the Health Center tabs and heart rail icon from the design canvas**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes
