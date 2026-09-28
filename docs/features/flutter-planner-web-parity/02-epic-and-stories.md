# Epic and Stories: Flutter Planner Web Parity

**Date:** 2026-09-24
**Stories in this batch:** 27 (1 design, 1 backend, 25 Flutter)
**Scope source:** PO scope call on the "Planner, Web vs Mobile Gap Analysis" doc (2026-09-24). Workflow and PRD steps skipped at the PO's request: stories are kept short and user-POV.

---

## Epic

### Title

**Bring the mobile app Planner to parity with the web Planner**

### Description

The mobile app's Planner covers the day-to-day half of the web Planner well: the list and calendar, the core filters, the post preview, comments, approvals and the basic post actions. It covers almost none of the collaboration and organisation half. On a phone, a user can't search their posts, filter by campaign or category, see why a post failed and retry it, select several posts to approve at once, send a client a review link, or see calendar notes and holidays. Teams who approve and fix content on the go keep having to find a laptop.

This epic closes those gaps in the Flutter app, so one change ships to iOS and Android. It adds the Instagram and TikTok grid views, the full web filter set with search, date range, per-status counts, saved filters and custom views, richer post cards, the missing post actions, bulk approve / reject / send for approval / share, client share links, calendar notes, holidays and display options, profile-synced preferences, and real-time updates for the Planner, post preview and comments. It also fixes one standing risk: dragging a post to a new date on mobile re-saves the whole post today, when all that should change is the time.

The backend already supports all of this for the web app. The work is almost entirely in the mobile app, plus one design story and one check that real-time updates accept the app's sign-in.

### In scope

- **Views:** Instagram grid and TikTok grid
- **Filters:** campaigns, content categories, first-comment status, comment status, automations, "No social account", label search, account grouping, the "Only" shortcut, caption search, date range, per-status counts, saved filters, custom views
- **Post cards:** labels, campaign / category, content type, per-account status and live link, issues and partial failures, thread and first-comment indicators, locked, hidden from clients, swipeable media, tappable hashtags and mentions
- **Post actions:** safe reschedule, send for approval, change approval, reopen, retry posting, publishing status, delete from social platforms, cross-workspace delete for global category posts, recycle, hide / unhide from clients, share via link, download PDF
- **Bulk:** approve, reject, send for approval, share via link
- **Preview:** full options menu, previous / next, download PDF, real-time updates
- **Comments:** All / Resolved / Unresolved filter, real-time sync
- **Sharing:** create, share and manage share links
- **Calendar:** notes (including recurring), holidays, display options (notes, holidays, media thumbnails, compact view, fade published), create-from-day menu, "+N more" drag-out, unfinished draft guard, correct first day of the week
- **Preferences:** default view, sort and calendar view synced with the user profile
- **Real-time** updates over the same service the web app uses

### Explicitly out of scope

- Compact List and Feed view modes
- CSV batch filter
- Bulk edit, bulk delete, bulk recycle
- Resend mobile notification (the user is already in the app)
- Shuffle queue
- Refresh automation post
- Post analytics in the planner and the preview
- Planner-wide PDF export (the per-post Download PDF is in scope)
- Holiday AI content (AI generation is web-only)
- Blog content type (blog publishing is sunset)
- Manage columns (no table on mobile)

### Keep as-is

These mobile-only strengths stay, and nothing in this epic removes them: download media to the phone's gallery, the Hourly and Daily week views, the long-press day peek, and the caption fallback.

### Sequencing

1. **First:** **[Flutter] Reschedule posts from the calendar without re-saving the whole post**, which removes a data-loss risk, and **[Design] Design the mobile Planner parity screens**.
2. **In parallel:** **[BE] Confirm real-time planner updates accept mobile app sign-in**, which unblocks the real-time story.
3. **Then:** card, filter and action stories in any order. **[Flutter] Create and share planner share links** comes before the bulk story, because bulk "Share via Link" reuses it.
4. Calendar stories last. **[Flutter] Add calendar display options to the planner** depends on the notes and holidays stories.

### Success measure

A user can find any post, see what happened to it, fix it and get it approved from their phone, without switching to the web app. Crashes and support tickets about posts losing their content after a drag drop to zero.

---

## Stories

1. **[Design] Design the mobile Planner parity screens**
2. **[BE] Confirm real-time planner updates accept mobile app sign-in**
3. **[Flutter] Reschedule posts from the calendar without re-saving the whole post**
4. **[Flutter] Show labels, campaign, category and content type on planner post cards**
5. **[Flutter] Show publishing status and post indicators on planner post cards**
6. **[Flutter] Add campaign, category, comment and automation filters to the planner**
7. **[Flutter] Search planner posts by caption and filter by date range**
8. **[Flutter] Show post counts on each status in the planner filter**
9. **[Flutter] Save planner filters and set a default filter**
10. **[Flutter] Create, apply and manage custom views in the planner**
11. **[Flutter] Sync the default planner view, sort and calendar view with the user profile**
12. **[Flutter] Send for approval, change approval and reopen posts from the planner**
13. **[Flutter] Retry failed posts and view per-account publishing status**
14. **[Flutter] Delete posts from social platforms when deleting from the planner**
15. **[Flutter] Recycle posts and hide posts from clients from the planner**
16. **[Flutter] Download and view a planner post as PDF**
17. **[Flutter] Create and share planner share links**
18. **[Flutter] Manage shared planner links**
19. **[Flutter] Select multiple posts to approve, reject, send for approval or share**
20. **[Flutter] Move between posts in the planner post preview**
21. **[Flutter] Filter comments by resolved status in the post preview**
22. **[Flutter] Keep the planner, post preview and comments updated in real time**
23. **[Flutter] Add Instagram and TikTok grid views to the planner**
24. **[Flutter] View, create and edit calendar notes in the planner**
25. **[Flutter] Show holidays in the planner calendar**
26. **[Flutter] Add calendar display options to the planner**
27. **[Flutter] Improve creating posts from the planner calendar**

---

## [Design] Design the mobile Planner parity screens

### Description:

As a mobile app user, I want the new Planner features to feel like they were built for a phone and not squeezed in from the web, so that I can use them one-handed and without zooming.

---

### Workflow:

The designer produces phone and tablet screens for every new surface in this epic, using the app's existing components and patterns.

---

### Acceptance criteria:

- [ ] Post card: layout for labels, campaign / category badge, content-type badge, per-account status row, issues pill, thread / first-comment / locked / hidden icons, and swipeable media, without the card getting taller than today for a typical post
- [ ] Filter sheet: new sections (campaigns, categories, comment status, first-comment status, automations, "No social account", date range), per-status counts, search bar, and saved filter / custom view pickers
- [ ] Post actions: the full card action sheet and preview options menu, ordered by how often each action is used
- [ ] Bulk selection: selection mode, selected count, "Select all", and the bottom action bar
- [ ] Share link: create sheet (steps: what to share, what viewers can do, other options), success state with Copy / Share, and the manage links list
- [ ] Calendar: note and holiday rendering in month, week hourly and week daily views; note create / edit sheet including repeat options; display options sheet
- [ ] Grid views: Instagram and TikTok grid, including the locked state for plans without grid access
- [ ] Publishing status sheet and retry progress
- [ ] Empty, loading and error states for every new list or sheet
- [ ] Works on the smallest supported phone and on tablets in both orientations

---

### Mock-ups:

Output of this story.

---

### Impact on existing data:

None.

---

### Impact on other products:

None. Mobile app only.

---

### Dependencies:

None. Blocks the Flutter stories in this epic that add new surfaces.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, design story)

---

## [BE] Confirm real-time planner updates accept mobile app sign-in

### Description:

As a mobile app user, I want the Planner to update on its own when a teammate changes something, so that I never act on a stale post. The web app already gets these updates. The mobile app signs in differently, so we need to confirm it can connect too.

---

### Workflow:

No user-facing screen. The behaviour the user sees is in **[Flutter] Keep the planner, post preview and comments updated in real time**.

---

### Acceptance criteria:

- [ ] A signed-in mobile app user can get a real-time connection token and channel tokens for the planner update channels of a workspace they belong to
- [ ] A user is refused channel tokens for a workspace they are not a member of
- [ ] An expired or signed-out mobile session is refused
- [ ] If anything needed to be fixed for mobile sign-in, web real-time behaviour is unchanged

---

### Mock-ups:

N/A, backend only.

---

### Impact on existing data:

None.

---

### Impact on other products:

Web app real-time must keep working unchanged.

---

### Dependencies:

None. Blocks **[Flutter] Keep the planner, post preview and comments updated in real time**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (N/A, backend only)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (N/A, backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, internal real-time sign-in only, no public API change)

---

## [Flutter] Reschedule posts from the calendar without re-saving the whole post

### Description:

As a mobile app user, I want dragging a post to a new date or time to change only its time, so that my caption, media, first comment and settings are never touched by a reschedule.

Today a drag on mobile re-saves the entire post. That is correct only while the app reads every setting back perfectly, and any gap silently wipes content. The web app changes only the time, and mobile should do the same.

---

### Workflow:

1. User drags a post to another day in month view, or another time slot in the week view.
2. The post moves straight away.
3. If the move is allowed, it stays there and the user sees "Post rescheduled".
4. If it isn't allowed, the post snaps back and the user sees why.

---

### Acceptance criteria:

- [ ] Rescheduling changes only the post's date and time; caption, media, first comment, labels, accounts and settings are unchanged afterwards
- [ ] The saved time matches what the user dropped on, in the workspace's timezone, including across daylight-saving changes
- [ ] A post cannot be dragged to a time in the past: "You can't schedule a post in the past. Pick a later time."
- [ ] Published and partially failed posts cannot be dragged: "Published posts can't be rescheduled."
- [ ] Posts in a content category cannot be dragged: "This post follows its content category's schedule. Change the category schedule to move it."
- [ ] Posts that aren't a draft, scheduled or in review cannot be dragged: "Only draft, scheduled and in review posts can be rescheduled."
- [ ] Approvers cannot drag posts
- [ ] Users without edit permission cannot drag posts
- [ ] If saving fails, the post returns to its original slot and the user sees "Couldn't reschedule this post. Please try again."
- [ ] When a reschedule succeeds, a `post_rescheduled` Usermaven event fires with `{ source: 'calendar_drag', view: 'month' | 'week_hourly' | 'week_daily' }`

---

### Mock-ups:

No new screens.

---

### Impact on existing data:

None. Only the post's scheduled time changes.

---

### Impact on other products:

None.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses an existing endpoint, no API change)

---

## [Flutter] Show labels, campaign, category and content type on planner post cards

### Description:

As a mobile app user, I want each post card to show its labels, campaign, content category and content type, so that I can tell posts apart at a glance without opening each one.

---

### Workflow:

1. User opens the Planner.
2. Each card shows the post's coloured labels, its campaign or content category, and a small content-type badge.
3. If a post has several images or videos, the user swipes through them on the card.
4. The user taps a #hashtag or @mention in the caption to see all posts with it.

---

### Acceptance criteria:

- [ ] Labels show as coloured chips in each label's colour. If there are more than fit, the card shows the first two and "+N"
- [ ] The campaign name, or the content category name in its colour, shows as a badge
- [ ] Content-type badge shows for Evergreen, RSS, Repeat, Carousel, Image and Video posts (no Blog badge)
- [ ] A card with several media items can be swiped through, with a dot indicator. It no longer shows only the first item and a count
- [ ] Tapping a #hashtag in the caption runs a caption search for that hashtag
- [ ] Tapping an @mention in the caption runs a caption search for that mention
- [ ] Cards with none of these show no empty space for them

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

- **[Design] Design the mobile Planner parity screens**
- Hashtag and mention taps use the search from **[Flutter] Search planner posts by caption and filter by date range**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [Flutter] Show publishing status and post indicators on planner post cards

### Description:

As a mobile app user, I want each card to show how the post did on each account and whether anything is wrong with it, so that I can spot a failed or blocked post without opening it.

---

### Workflow:

1. User opens the Planner.
2. A published post shows a status for each account. The user taps "View live" to open the live post.
3. A post with problems shows a red "2 errors" pill. The user taps it to see what's wrong.
4. Small icons show if the post is a thread, has a first comment, is locked, or is hidden from clients.

---

### Acceptance criteria:

- [ ] Each account on a published or partly published post shows its own status: Published, Failed or Processing
- [ ] A published account shows "View live", which opens the live post on the network
- [ ] A post with issues shows a pill "{n} error" / "{n} errors". Tapping it lists each issue with the account it applies to
- [ ] A partly failed post shows "Published on {x} of {y} accounts"
- [ ] A thread icon shows on X / Threads / Bluesky thread posts, with the hint "This post is a thread"
- [ ] A first-comment icon shows on posts with a first comment, with the hint "This post has a first comment"
- [ ] A lock icon shows on locked posts, with the hint "This post is locked and can't be edited right now"
- [ ] An eye-off icon shows on posts hidden from clients, with the hint "Hidden from clients on share links"

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [Flutter] Add campaign, category, comment and automation filters to the planner

### Description:

As a mobile app user, I want the same filters as on the web Planner, so that I can narrow down to exactly the posts I need to review on my phone.

---

### Workflow:

1. User taps the filter button in the Planner.
2. User picks one or more campaigns, content categories, first-comment statuses, comment statuses or automations, or turns on "No social account".
3. The Planner shows only matching posts and the filter button shows how many filters are on.
4. The user can tap "Only" next to any option to pick just that one.

---

### Acceptance criteria:

- [ ] New filter sections: **Campaigns**, **Content Categories**, **First Comment** (Published / Failed / Scheduled), **Comments** (Resolved / Unresolved), **Automations**, and a **No social account** switch
- [ ] "No social account" tooltip: "Show posts that aren't linked to any social account, for example drafts saved before picking accounts."
- [ ] Campaigns, Content Categories and Labels each have a search box: "Search campaigns", "Search categories", "Search labels"
- [ ] The accounts filter gains a "Group by platform" option
- [ ] Each option has an "Only" shortcut that selects just that option in its section
- [ ] Filters combine with each other and with the existing filters, in list and calendar views
- [ ] The active-filter indicator counts every filter, including "Created by"
- [ ] "Clear all" resets every filter
- [ ] Empty result: headline "No posts match these filters", subtext "Try removing a filter or two.", button "Clear filters"

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Search planner posts by caption and filter by date range

### Description:

As a mobile app user, I want to search my posts by their words and pick a date range, so that I can find a specific post fast without scrolling.

---

### Workflow:

1. User taps the search icon in the Planner and types, for example, "Black Friday".
2. The list updates to posts whose caption contains that text.
3. In list view, the user taps the date range and picks a start and end date, or a preset such as "Next 7 days".
4. The list shows only posts in that range.

---

### Acceptance criteria:

- [ ] Search box placeholder: "Search posts by caption"
- [ ] Results update shortly after the user stops typing, without a button
- [ ] Search works in list and calendar views
- [ ] List view has a date range picker with presets: Today, Next 7 days, Next 30 days, This month, Custom
- [ ] The chosen range shows on the filter bar, for example "Oct 1 - Oct 15", and can be cleared
- [ ] Search with no results: headline "No posts found for "{text}"", subtext "Check the spelling or try a shorter word."

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Show post counts on each status in the planner filter

### Description:

As a mobile app user, I want to see how many posts are in each status, so that I know at a glance how much is waiting for review or has failed.

---

### Workflow:

1. User opens the status filter.
2. Each status shows its count, for example "In Review 12" or "Failed 3".
3. Counts update when the user changes other filters.

---

### Acceptance criteria:

- [ ] Every status in the status filter shows its post count for the current filters
- [ ] Counts refresh when any other filter, the search or the date range changes
- [ ] While counts load, a small placeholder shows in place of the number
- [ ] If counts fail to load, the statuses still work and show no number

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

**[Flutter] Add campaign, category, comment and automation filters to the planner**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Save planner filters and set a default filter

### Description:

As a mobile app user, I want to save a set of filters I use often and pick one as my default, so that I don't have to rebuild it every time. Filters saved on the web appear in the app and the other way round.

---

### Workflow:

1. User sets up filters, for example "In Review" and the "Client A" label.
2. User taps "Save filter", enters "Client A review" and saves.
3. Later the user opens "Saved filters" and taps "Client A review" to apply it.
4. User marks it as default, so the Planner opens with it.

---

### Acceptance criteria:

- [ ] "Save filter" sheet: title "Save filter", field "Filter name" with placeholder "e.g. Client A review", buttons "Save" / "Cancel"
- [ ] Name is required: "Please enter a name for this filter"
- [ ] Saved filters list shows each filter with "Apply", "Set as default", "Update with current filters" and "Delete"
- [ ] Delete asks "Delete this saved filter?" with "Delete" / "Cancel"
- [ ] The default filter is applied when the Planner opens
- [ ] Filters saved on the web appear in the app and the other way round
- [ ] Empty state: "No saved filters yet", subtext "Set up your filters, then tap Save filter to reuse them in one tap."
- [ ] When a user saves a filter, a `planner_filter_saved` Usermaven event fires with `{ is_default, filter_count }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None. Uses the same saved filters as the web app.

---

### Impact on other products:

Filters saved in the app show in the web Planner.

---

### Dependencies:

**[Flutter] Add campaign, category, comment and automation filters to the planner**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Create, apply and manage custom views in the planner

### Description:

As a mobile app user, I want to use the same custom views my team uses on the web, and create my own, so that the Planner opens exactly the way I work.

---

### Workflow:

1. User taps the view picker at the top of the Planner and sees their custom views.
2. User taps one; the Planner switches to its filters, view mode and calendar settings.
3. User taps "New view", names it, picks what to include and who can see it, and saves.
4. User can edit, delete, reorder or set a view as default.

---

### Acceptance criteria:

- [ ] Custom views created on the web show in the app, in the same order
- [ ] Applying a view applies its filters, view mode and calendar settings
- [ ] "New view" sheet: field "View name" (placeholder "e.g. This week's client posts"), visibility "Only me" / "Everyone in this workspace", buttons "Save view" / "Cancel"
- [ ] Visibility tooltip: "Choose Everyone to let your whole team use this view. Only you can edit it."
- [ ] Views can be edited, deleted (with confirmation "Delete this view?"), reordered by drag, and set as default
- [ ] The default view is applied when the Planner opens
- [ ] Empty state: "No custom views yet", subtext "Save your filters and layout as a view to switch to it in one tap."
- [ ] When a user saves a custom view, a `planner_custom_view_saved` Usermaven event fires with `{ visibility, is_default }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None. Uses the same custom views as the web app.

---

### Impact on other products:

Views created in the app show on the web.

---

### Dependencies:

- **[Flutter] Save planner filters and set a default filter**
- **[Flutter] Add calendar display options to the planner** (for a view's calendar settings)

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Sync the default planner view, sort and calendar view with the user profile

### Description:

As a mobile app user, I want the Planner to remember my preferred view, sort order and calendar view across my phone and the web, so that it opens the way I like everywhere.

---

### Workflow:

1. User switches the Planner to calendar month view and sorts by "Last updated".
2. User opens the app on another phone, or the web app, and sees the same view and sort.

---

### Acceptance criteria:

- [ ] Changing the view mode (list / calendar) saves it as the user's default
- [ ] Changing the sort saves it as the user's default
- [ ] Changing the calendar view (month / week) saves it as the user's default
- [ ] The Planner opens with the saved defaults on any device, including the web app
- [ ] The Hourly / Daily week choice stays a per-device setting, since the web has no equivalent
- [ ] When the user changes the view mode, the existing `planner_view_changed` Usermaven event fires with `{ view }`

---

### Mock-ups:

No new screens.

---

### Impact on existing data:

Uses the same saved preferences as the web app. A mobile change changes the web default too.

---

### Impact on other products:

Web Planner defaults follow changes made in the app.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Send for approval, change approval and reopen posts from the planner

### Description:

As a mobile app user, I want to send a post for approval, change who approves it, or reopen a rejected or failed post from my phone, so that the approval flow doesn't stall while I'm away from my desk.

---

### Workflow:

1. User taps "..." on a draft post and picks "Send for Approval".
2. User picks approvers and the rule (anyone or everyone), adds an optional note, and sends.
3. On a post already in review, the user picks "Change Approval" to edit the approvers.
4. On a rejected, failed or missed post, the user picks "Reopen Post" to move it back to draft.

---

### Acceptance criteria:

- [ ] "Send for Approval", "Change Approval" and "Reopen Post" appear in the card actions and the preview options menu, only for posts and users where the web offers them
- [ ] Send for Approval sheet: title "Send for Approval", approver picker, rule "Anyone can approve" / "Everyone must approve", note field (placeholder "Add a note for your approvers (optional)"), buttons "Send" / "Cancel"
- [ ] Rule tooltip: "Anyone: the first approval moves the post forward. Everyone: every approver must approve before it moves on."
- [ ] At least one approver is required: "Pick at least one approver"
- [ ] Success toasts: "Sent for approval", "Approval updated", "Post reopened as a draft"
- [ ] The post's status on the card updates straight away
- [ ] When a post is sent for approval, a `post_sent_for_approval` Usermaven event fires with `{ approver_count, rule, source: 'single' }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Approvers get their usual notifications.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Retry failed posts and view per-account publishing status

### Description:

As a mobile app user, I want to see exactly which accounts a post failed on and retry it, so that when I get a "post failed" notification I can fix it right there.

---

### Workflow:

1. User opens a failed post, or taps "..." and picks "Publishing Status".
2. A sheet lists each account with its status, the error reason for failed ones, and "View live" for published ones.
3. User taps "Retry" next to a failed account.
4. The row shows "Retrying..." and then updates to Published or Failed.

---

### Acceptance criteria:

- [ ] "Publishing Status" is in the card actions and preview options menu for published, partly failed and failed posts
- [ ] Sheet title: "Publishing Status". Each row shows the account, status, time and, for failures, the network's error in plain words
- [ ] "Retry" shows on each failed account row. On a fully failed post, "Retry all" also shows
- [ ] While retrying, the row shows "Retrying..." and can't be tapped again
- [ ] The result updates in the sheet without closing it
- [ ] If the retry request itself fails: "Couldn't retry this post. Please try again."
- [ ] Opening a failed-post push notification lands on the post with "Retry" reachable in one tap
- [ ] When a user taps Retry, a `post_retried` Usermaven event fires with `{ platform, retry_all }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

- **[Design] Design the mobile Planner parity screens**
- Live retry progress uses **[Flutter] Keep the planner, post preview and comments updated in real time**. Until that ships, the sheet refreshes after a retry

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Delete posts from social platforms when deleting from the planner

### Description:

As a mobile app user, I want to choose whether deleting a published post also removes it from the social networks, so that I can take a post down from my phone.

---

### Workflow:

1. User taps "..." on a published post and picks "Delete".
2. The delete sheet asks: "Delete from ContentStudio only" or "Also delete from social platforms".
3. If the user picks the second, they choose which accounts to delete it from.
4. The user confirms. The sheet shows which deletions worked and lets them retry any that failed.

---

### Acceptance criteria:

- [ ] Delete sheet title: "Delete post?", options "Delete from ContentStudio only" / "Also delete from social platforms"
- [ ] Choosing social platforms lists each published account with a checkbox, all selected by default
- [ ] Accounts that can't be deleted by the network are marked and not selectable, with the note: "Due to API limitations, posts from Facebook Groups, Instagram, TikTok, Facebook Stories, and GBP video posts can't be deleted from the platform. Delete them directly on the network."
- [ ] Buttons: "Delete" (destructive) / "Cancel"
- [ ] After deleting, the sheet shows each account's result. Failed ones show "Retry"
- [ ] For a post created from a global content category, the sheet also offers "Delete from all workspaces", with the hint "This post was added to several workspaces from a global category. Delete it everywhere at once."
- [ ] Unpublished posts keep the simple confirm: "Delete this post? This can't be undone."

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

Deleting from platforms removes the live post from the network. This can't be undone.

---

### Impact on other products:

None.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Recycle posts and hide posts from clients from the planner

### Description:

As a mobile app user, I want to recycle a good post and hide posts I don't want clients to see, so that I can manage content and client views from my phone.

---

### Workflow:

1. User taps "..." on a published post and picks "Recycle Post".
2. The composer opens with the post ready to add to an Evergreen campaign.
3. On any post, the user picks "Hide from Clients". The post gets the hidden icon.
4. The user picks "Show to Clients" to undo it.

---

### Acceptance criteria:

- [ ] "Recycle Post" is in the card actions and preview options menu wherever the web offers it. It opens the composer with the post's content
- [ ] "Hide from Clients" / "Show to Clients" toggles the post's visibility on share links
- [ ] Tooltip: "Hidden posts stay in your Planner but don't appear on any share link you send to clients."
- [ ] Success toasts: "Post hidden from clients", "Post visible to clients"
- [ ] The card's hidden icon updates straight away

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Hidden posts stop showing on the web share-link page.

---

### Dependencies:

**[Flutter] Show publishing status and post indicators on planner post cards** (for the hidden icon)

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Download and view a planner post as PDF

### Description:

As a mobile app user, I want to download a post as a PDF, so that I can send it to a client or keep a record from my phone.

---

### Workflow:

1. User opens a post and picks "Download PDF" from the options menu.
2. The app prepares the PDF and opens it in a viewer.
3. User taps share to save it to Files or send it on.

---

### Acceptance criteria:

- [ ] "Download PDF" is in the card actions and preview options menu, only for users allowed to download PDFs on the web
- [ ] While preparing: "Preparing PDF..."
- [ ] The PDF opens in an in-app viewer with a share button that opens the phone's share sheet (save to Files, email, and so on)
- [ ] On white-label domains the PDF carries the white-label branding, as on the web
- [ ] On failure: "Couldn't create the PDF. Please try again."

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Create and share planner share links

### Description:

As a mobile app user, I want to create a share link for my calendar or some posts and send it to a client from my phone, so that they can review and approve content without a ContentStudio account.

Share links are client-facing. Adding external approvers replaces the post's internal approver, as it does on the web. **Product sign-off on this behaviour on mobile is needed before build.**

---

### Workflow:

1. User taps "Share" in the Planner and picks "Share Calendar", or "Share via Link" on a post.
2. **What to share:** user enters a title and picks "Current view only", "Future updates for this view" or "Share full calendar".
3. **What viewers can do:** user picks "Add comments" and optionally "Allow approve/reject".
4. **Other options:** password protection, "Share Notes", and optional client emails with "Send for approval".
5. User taps "Create Link", then "Copy Link" or "Share" to send it with any app on the phone.

---

### Acceptance criteria:

- [ ] Entry points: "Share Calendar" from the Planner header, and "Share via Link" in the card actions and preview options menu
- [ ] Use the web's copy for every field, option, hint and error. Key strings: title "Create a secure shareable link", subtext "Share content with clients and gather feedback.", title placeholder "Enter a title to make your link easily identifiable."
- [ ] Sharing options tooltip: "Choose whether this link is a one-time snapshot or a live view that automatically reflects future changes."
- [ ] Validation: "Link name must be at least 3 characters.", "Link name must be less than 50 characters.", "Password must be at least 6 characters when protection is enabled.", "Please enter a valid email address"
- [ ] Approval rule shows "Anyone" / "Everyone" with the warning "Adding external approver will overwrite existing internal approver"
- [ ] Send for approval is off and explained for future-update links: "Approvals are only supported when sharing existing posts. Future posts cannot be sent for approval."
- [ ] Success screen: "Shareable link created", "Your link is ready, share it and gather feedback.", buttons "Copy Link" and "Share"
- [ ] "Share" opens the native iOS / Android share sheet with the link
- [ ] "Copy Link" shows "Shareable link copied to clipboard"
- [ ] Calendar share with nothing to share: "Cannot share the calendar. Please create a post or share notes first."
- [ ] When a link is created, a `share_link_created` Usermaven event fires with `{ mode, approval_enabled, password_protected, source: 'calendar' | 'post' | 'bulk' }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

Creates share links, the same as the web. External approvers replace the internal approver on those posts.

---

### Impact on other products:

Links open the existing web share page. Clients get the existing invitation emails.

---

### Dependencies:

- **[Design] Design the mobile Planner parity screens**
- Product sign-off on external approval from mobile

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review (share link must use the white-label domain)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Manage shared planner links

### Description:

As a mobile app user, I want to see and manage the share links I've sent, so that I can turn off a link a client should no longer use, or resend one.

---

### Workflow:

1. User taps "Share", then "Manage Shared Links".
2. User sees all links with their title, type and whether they're on.
3. User searches, copies or shares a link, turns it off or on, edits it, or deletes it.

---

### Acceptance criteria:

- [ ] List shows each link's title, what it shares, created date and an on / off switch
- [ ] Search box: "Search links"
- [ ] Actions per link: "Copy Link", "Share", "Edit", "Delete"
- [ ] Edit opens the create sheet with the title "Edit Shareable Link". Saving shows "Link details successfully updated"; saving with no changes shows "Please make some changes to save."
- [ ] Turning a link off stops it opening for clients, who see "The link is invalid or disabled by the user."
- [ ] Delete asks "Delete this link? Anyone with it will lose access." with "Delete" / "Cancel"
- [ ] Empty state: "No shared links yet", subtext "Share your calendar or a few posts with a client to collect feedback and approvals.", button "Share Calendar"

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

Same links as the web. Changes show on both.

---

### Impact on other products:

Disabled or deleted links stop working on the web share page.

---

### Dependencies:

**[Flutter] Create and share planner share links**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Select multiple posts to approve, reject, send for approval or share

### Description:

As a mobile app user, I want to select several posts at once and approve, reject, send for approval or share them together, so that I can clear a review queue in one go.

---

### Workflow:

1. User long-presses a post card in list view. Selection mode starts with that post selected.
2. User taps more posts, or taps "Select all".
3. A bar at the bottom shows "{n} selected" and the actions "Approve", "Reject", "Send for Approval" and "Share via Link".
4. User picks an action, confirms, and sees the result.

---

### Acceptance criteria:

- [ ] Long-press on a post card starts selection mode. The card's "..." button still opens its actions
- [ ] Long-press on a calendar day still opens the day peek, unchanged
- [ ] "Select all" selects every post on screen, then offers "Select all {n} posts matching your filters"
- [ ] Only actions the user is allowed to take are shown
- [ ] Posts the action can't apply to are skipped, with the message "Only {x} of {y} selected posts can be {action}. Published and failed posts are excluded."
- [ ] Reject asks for an optional comment, as the single-post reject does
- [ ] "Send for Approval" opens the same sheet as **[Flutter] Send for approval, change approval and reopen posts from the planner**
- [ ] "Share via Link" opens the share link sheet with the selected posts
- [ ] Tapping "Cancel" or the back button leaves selection mode
- [ ] Bulk approve / reject fire the existing `post_approved` / `post_rejected` Usermaven events once per action with `{ source: 'bulk', count }`
- [ ] Bulk send for approval fires `post_sent_for_approval` with `{ approver_count, rule, source: 'bulk' }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Approvers get their usual notifications.

---

### Dependencies:

- **[Flutter] Send for approval, change approval and reopen posts from the planner**
- **[Flutter] Create and share planner share links**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Move between posts in the planner post preview

### Description:

As a mobile app user, I want to move to the next or previous post from the preview, so that I can review a batch of posts without going back to the list each time.

---

### Workflow:

1. User opens a post from the Planner.
2. User swipes left or right, or taps the next / previous arrows.
3. The next post in the current list opens straight away.

---

### Acceptance criteria:

- [ ] Next / previous follow the order and filters of the list or calendar the user came from
- [ ] The next post opens without a loading wait in normal conditions
- [ ] The arrows are disabled on the first and last post
- [ ] The preview's options menu includes every action added in this epic
- [ ] If a post was deleted by someone else, the preview shows "This post has been deleted" with a "Back to Planner" button

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [Flutter] Filter comments by resolved status in the post preview

### Description:

As a mobile app user, I want to show only resolved or unresolved comments, so that I can focus on the feedback that still needs action.

---

### Workflow:

1. User opens a post's comments.
2. User taps the filter icon and picks "Unresolved Comments".
3. Only unresolved comments show.

---

### Acceptance criteria:

- [ ] Filter options (same as web): "All Comments", "Resolved Comments", "Unresolved Comments"
- [ ] The filter icon tooltip reads "Filter Comments"
- [ ] Resolving a comment while "Unresolved Comments" is on removes it from the list
- [ ] Empty filtered state: "No unresolved comments" / "No resolved comments"

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [Flutter] Keep the planner, post preview and comments updated in real time

### Description:

As a mobile app user, I want changes my teammates make to show up on my phone right away, so that I never approve an old version or miss a new comment.

---

### Workflow:

1. User has the Planner or a post open.
2. A teammate on the web comments, changes a status or label, or edits the post.
3. The change appears on the user's phone within a few seconds, without pulling to refresh.

---

### Acceptance criteria:

- [ ] New, edited and deleted comments, and "resolve all", appear in an open comments thread without refreshing
- [ ] Status, label and post changes update the matching card in the list and calendar, and the open preview
- [ ] Retry progress in **[Flutter] Retry failed posts and view per-account publishing status** updates live
- [ ] Changes for posts not currently loaded are ignored without error
- [ ] Updates stop when the app goes to the background and catch up when it returns
- [ ] Switching workspace stops updates for the old workspace and starts them for the new one
- [ ] If the real-time connection drops, the Planner still refreshes when the app returns to the foreground and on pull-to-refresh

---

### Mock-ups:

No new screens.

---

### Impact on existing data:

None.

---

### Impact on other products:

Uses the same real-time service as the web app.

---

### Dependencies:

**[BE] Confirm real-time planner updates accept mobile app sign-in**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (N/A, no new copy)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [Flutter] Add Instagram and TikTok grid views to the planner

### Description:

As a mobile app user, I want to see how my scheduled Instagram and TikTok posts will look on my profile grid, so that I can plan a feed that looks good together, on the phone where people actually see it.

---

### Workflow:

1. User opens the view switcher in the Planner and picks "Instagram Grid" or "TikTok Grid".
2. User picks an account.
3. The grid shows scheduled posts first, followed by already published posts from that account.
4. User taps a tile to open the post preview.

---

### Acceptance criteria:

- [ ] View switcher gains "Instagram Grid" and "TikTok Grid"
- [ ] Account picker shows only connected Instagram or TikTok accounts
- [ ] Scheduled posts are marked so they're easy to tell apart from published ones
- [ ] Instagram grid uses the profile's portrait tile shape; TikTok uses its portrait video tiles
- [ ] Users on plans without grid view see the options with a lock, and tapping shows "Grid view isn't included in your plan. Upgrade to preview your feed." with "Upgrade" / "Not now"
- [ ] No accounts connected: "Connect an Instagram account to preview your grid" with "Connect account"
- [ ] Nothing scheduled: "No upcoming posts for this account", subtext "Schedule a post to see how it fits your grid."
- [ ] Switching to a grid fires the existing `planner_view_changed` Usermaven event with `{ view: 'instagram_grid' | 'tiktok_grid' }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] View, create and edit calendar notes in the planner

### Description:

As a mobile app user, I want to see and add notes on my calendar, such as "Product launch" or "Campaign brief due", so that my team's plans are visible next to the posts.

---

### Workflow:

1. User opens the calendar and sees notes on their dates.
2. User taps a note to read it, or taps "New Note" on a day.
3. User enters a title, optional description, dates, colour and visibility, and optionally sets it to repeat.
4. User saves. The note appears on the calendar.
5. From a note, the user can edit, duplicate, delete, drag it to another date, or turn it into a post.

---

### Acceptance criteria:

- [ ] Notes show in month, week hourly and week daily views in their colour. Private notes show a lock icon, recurring notes a repeat icon
- [ ] Note sheet uses the web's copy: "Add Note" / "Edit Note", "Title" ("Add your title here"), "Description" ("Add your note body here (optional)"), "Start Date", "End Date", "Color", "Visibility" ("Only me" / "Everyone"), "Repeat"
- [ ] Visibility hints: "Note will be shown to you only" / "Note will be visible to all team members in this workspace"
- [ ] Repeat: every N day / week / month / year, days of the week for weekly, ends "On" a date or "After" N occurrences, with the summary line, for example "Note will be repeated every 2 weeks on Mon and Thu, until Dec 31"
- [ ] Validation: "Note title is required.", "Start date is required."
- [ ] Toasts: "Note saved successfully.", "Note updated successfully.", "Note deleted successfully."
- [ ] Delete asks "Delete Note" / "Are you sure you want to delete this note?"
- [ ] Dragging a note to another date moves it. On failure it snaps back with "Failed to update note date. Please try again."
- [ ] "Social Post" on a note opens the composer with the note's title and description
- [ ] When a note is saved for the first time, a `planner_note_created` Usermaven event fires with `{ is_private, is_recurring }`

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None. Same notes as the web.

---

### Impact on other products:

Notes created in the app show on the web calendar and on share links that include notes.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Show holidays in the planner calendar

### Description:

As a mobile app user, I want to see public holidays for the countries I post to, so that I can plan around them.

---

### Workflow:

1. User opens calendar options and taps "Holiday settings".
2. User picks one or more countries, and optionally applies them to all their workspaces.
3. Holidays appear on the calendar.
4. User taps a holiday to see its details, or removes it.

---

### Acceptance criteria:

- [ ] Holiday settings: searchable country list ("Search countries") and a switch "Use these countries in all my workspaces"
- [ ] Holidays show on their dates in month, week hourly and week daily views, labelled "Public Holiday"
- [ ] Tapping a holiday shows "Holiday details" with its name, date and country, and "Remove holiday"
- [ ] Remove asks "Remove Holiday" / "Are you sure you want to remove this holiday from your calendar?". Success: "Holiday removed successfully"; failure: "Failed to remove holiday"
- [ ] No AI content option for holidays (AI generation is web-only)

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None. Same holiday settings as the web.

---

### Impact on other products:

Country choices made in the app apply on the web.

---

### Dependencies:

**[Design] Design the mobile Planner parity screens**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Add calendar display options to the planner

### Description:

As a mobile app user, I want to choose what my calendar shows, so that I can keep it clean on a small screen or see everything when I need to.

---

### Workflow:

1. User taps the calendar options icon.
2. User switches Notes, Holidays, Media Thumbnails, Compact View or Fade out published posts on or off.
3. The calendar updates straight away.

---

### Acceptance criteria:

- [ ] Options: "Notes", "Holidays", "Media Thumbnails", "Compact View", "Fade out published posts"
- [ ] Tooltips: Media Thumbnails "Show or hide media thumbnails to clean up your calendar and view more posts at once."; Compact View "Show each post as a single line so more posts fit on each day."; Fade out published posts "Dim posts that have already been published so you can focus on what's coming up."
- [ ] Each option applies to month, week hourly and week daily views
- [ ] Choices are saved to the user's profile and match the web
- [ ] When notes are hidden and the user saves a note: "Note saved successfully. Notes are currently hidden, enable them from Calendar settings to view."

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

Choices made in the app apply on the web calendar.

---

### Dependencies:

- **[Flutter] View, create and edit calendar notes in the planner**
- **[Flutter] Show holidays in the planner calendar**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, uses existing endpoints, no API change)

---

## [Flutter] Improve creating posts from the planner calendar

### Description:

As a mobile app user, I want creating a post from a calendar day to work like the web, and the calendar to start the week on my chosen day, so that the calendar fits how I plan.

---

### Workflow:

1. User taps an empty day or time slot.
2. A menu offers "Social Post", "Use Template" and "New Note".
3. If the user has an unfinished composer draft, the app asks what to do with it first.
4. On a busy day, the user taps "+N more" and can drag a post out of the day sheet onto another day.

---

### Acceptance criteria:

- [ ] Tapping an empty day or slot shows "Social Post", "Use Template" ("Start from a saved template") and "New Note"
- [ ] "Social Post" opens the composer with that date and time filled in; "Use Template" opens the template picker first
- [ ] With an unfinished draft open in the composer: "You have an unfinished post. Continue editing it or start a new one?" with "Continue editing" / "Start new"
- [ ] A post can be dragged from the "+N more" day sheet onto another day
- [ ] The calendar's first day of the week follows the user's setting, not always Sunday
- [ ] Past days don't offer "Social Post"

---

### Mock-ups:

See **[Design] Design the mobile Planner parity screens**.

---

### Impact on existing data:

None.

---

### Impact on other products:

None.

---

### Dependencies:

- **[Flutter] View, create and edit calendar notes in the planner** (for "New Note")
- **[Flutter] Reschedule posts from the calendar without re-saving the whole post** (for dragging out of the day sheet)

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (phones and tablets)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)
