# PRD: Multi-Tiered Approval Workflow

**Author:** Product
**Last Updated:** 2026-03-25
**Status:** In Review
**Target Release:** Q2 2026
**Shortcut Epic:** https://app.shortcut.com/contentstudio-team/epic/47607
**Design:** https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034
**Implementation Docs:** `06-frontend-data-contract.md`, `07-final-qa-checklist.md`, `08-frontend-implementation-map.md`

---

## 1. Overview

Multi-tiered approval workflow allows ContentStudio workspaces to define named, reusable approval workflows with up to 5 sequential levels. Each level contains one or more approvers and an approval rule (everyone must approve, or anyone can approve). Posts can be submitted via a predefined workflow or custom user selection. The workflow progresses level by level, notifying each tier only when the previous one is complete. This replaces the current single-level approval modal with a right-sidebar panel, adds a dedicated approval notification icon in the top bar, extends approval tracking across all planner views, and brings parity to mobile apps.

The backend will implement this as a new parallel system alongside the existing single-level approval — not modifying existing approval logic. The frontend interacts entirely through new API endpoints. Backend architecture and implementation details are owned by the backend team in a separate technical spec.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio's current approval system is single-level: a post creator selects one or more users, sets one rule (anyone/everyone), and all approvers are notified simultaneously. There is no concept of sequential stages — a legal team cannot review after an internal team has approved. For agencies and larger teams, this forces workarounds (manual follow-up, re-sending posts) or drives them to external tools.

Additionally:
- The approval UI is a center modal that obscures the composer, making it hard to reference content while selecting approvers.
- There is no saved approval template — users re-select the same people every time.
- Approval notifications are mixed into general notifications, making it easy to miss time-sensitive requests.
- The mobile apps do not support workflow-based approvals.
- Editing a rejected post silently resets post status (confirmed bug in `ApprovalBuilder`).

**Who has this problem?**

- **Agencies** managing client content: need client approval after internal review
- **Enterprise teams**: legal/compliance review after marketing approval
- **Mid-size teams**: manager sign-off after writer prepares content
- Any workspace with 3+ roles involved in content creation

**What happens if we don't solve it?**

- Agencies continue churning to Planable, Gain, or Sprout Social — all of which have multi-level approval
- Support burden from the rejection/edit bug and approval confusion
- Missed upsell opportunity — approval workflows are a high-value agency/enterprise feature

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Increase approval feature adoption | % of workspaces using approval per month | +40% vs baseline | Product analytics |
| Reduce approval-related support tickets | Support ticket volume tagged "approval" | -30% in 60 days post-launch | Intercom |
| Drive workflow adoption | % of approval submissions using a saved workflow (vs custom approval) | 50% within 90 days | Product analytics |
| Fix rejection edit bug | Zero reports of rejected posts auto-moving to draft/scheduled | 0 recurrences | Bug tracker / support |
| Mobile parity | % of mobile approval actions (approve/reject) using workflow tab | 30% within 60 days | Mobile analytics |

---

## 4. Target Users

**Primary Persona: Agency Account Manager**
Manages 5–20 client workspaces. Creates content, routes it through an internal creative review, then to the client for sign-off. Currently manually tracks approvals. Needs reliable, repeatable multi-step workflows per client.

**Secondary Persona: Enterprise Social Media Manager**
Manages brand accounts with compliance requirements. Content must pass through a manager, then legal, before scheduling. Needs sequential level enforcement so compliance is never skipped.

**Secondary Persona: Approver (Client or Team Member)**
Receives content for review. Needs clear notifications, simple approve/reject actions, and visibility into where the post stands in the pipeline.

**Non-Users (out of scope):**
- Solo users with no team — approval workflows are irrelevant
- External approvers via share link — existing external approval system is unchanged in v1

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Workspace Admin | create a named approval workflow with multiple levels and assign team members to each | I don't have to re-select approvers every time I send a post | P0 |
| US-2 | Content Creator | send a post through a saved workflow with one click | I don't manually coordinate who reviews in what order | P0 |
| US-3 | Content Creator | send a post to custom approvers (without a workflow) as before | I retain flexibility for one-off approvals | P0 |
| US-4 | Approver | get notified only when the previous level has approved | I'm not interrupted before it's my turn | P0 |
| US-5 | Content Creator | see exactly which level a post is at and who has/hasn't approved | I can follow up or re-notify specific people | P0 |
| US-6 | Content Creator | edit a post mid-approval and choose whether to restart, re-notify the current level, or keep approvals | I control the approval state after making changes | P0 |
| US-7 | Approver | revoke my own approval if I change my mind before the workflow completes | I can correct a mistake | P1 |
| US-8 | Workspace Admin | mark a workflow as default | The most-used workflow is pre-selected and saves clicks | P1 |
| US-9 | Approver | bulk approve/reject multiple posts from my Pending Approvals view | I can process a queue efficiently | P1 |
| US-10 | Content Creator | bulk send multiple posts through the same approval workflow | I can route a campaign's content in one action | P0 |
| US-11 | Content Creator | see approval activity in a dedicated notification area | Approval updates don't get lost in general notifications | P1 |
| US-12 | Mobile User | send posts through a saved workflow and approve/reject from mobile | I can manage approvals on the go | P1 |
| US-13 | Content Creator | edit a rejected post and choose to resend or remove the approval | The current broken behavior (silent status reset) is fixed | P0 |
| US-14 | Workspace Admin | control which non-admin roles can manage workflows | Workflow setup is admin-controlled by default | P1 |
| US-15 | Content Creator | send a single post for approval directly from the Planner (not just from the Composer) | I don't have to open the editor just to route a post for review | P1 |
| US-16 | Content Creator | see a clear warning before starting a new approval that would override an existing one | I'm never surprised that a previous approval was silently cancelled | P0 |

---

## 6. Requirements

### 6.1 Must Have (P0)

**Settings — Approval Workflow Management**
- Create, edit, duplicate, delete approval workflows from Workspace Settings → Approval Workflows
- Workflow has: name, up to 5 levels, each level has a title, approval rule (everyone/anyone), and 1+ members
- Drag & drop to reorder levels within a workflow
- Duplicate a level (creates copy with same members and rule)
- Delete a level (confirmation dialog if members assigned)
- Fixed right panel with all eligible workspace members; search + drag & drop to assign to levels
- Save workflow as draft (incomplete, not visible in composer until published/completed)
- Workflow deletion: if in-flight posts exist, show confirmation and convert those posts to single-user custom approval

**Send for Approval — Sidebar Panel (replacing current center modal)**
- Right sidebar replaces existing `ApprovalModal.vue` center modal in: Composer, Planner bulk actions, Automations
- Two tabs: "Users" (existing custom approval mode, updated UI) and "Approval Workflows" (new)
- Users tab: search, alphabetical list, checkbox select, Everyone/Anyone rule, notes field, Send button
- Approval Workflows tab: workflow cards (collapsed/expanded), default pre-selected, notes field, Send button
- Collapsed card: workflow name, level count, member avatars grouped by level with connecting visual indicator
- Expanded card: "Level N: [Title]" per level, member avatars, rule label ("Everyone" / "Anyone")
- "Edit in Settings" link in tab for workflow modification (with redirect confirmation dialog)

**Approval Flow Progression**
- Level-by-level progression: Level 1 completes → Level 2 notified → etc.
- "Everyone must approve" rule: all members must approve before level advances
- "Anyone can approve" rule: one approval advances to next level; remaining approvers at that level notified it's no longer needed
- Rejection: post moves to Rejected; all other pending approvers at same level notified
- Missed Review: approval process continues; only current active level approvers notified; reschedule prompt on action
- Post deletion during approval: current active level approvers notified; process terminates
- Approval history permanently linked to post (read-only after publish)

**Edit/Rejection Bug Fix and Confirmation Dialogs**
- Editing a post in active approval (single-user) → confirmation: Re-notify / Keep status / Remove approval
- Editing a post in active approval (workflow) → confirmation: Restart from Level 1 / Re-notify current level / Keep status / Remove approval
- Editing a rejected post (single-user) → confirmation: Resend for approval / Remove approval *(bug fix: currently silently resets status)*
- Editing a rejected post (workflow) → confirmation: Restart from Level 1 / Resume from Level 2 / Remove approval
- Editing a Missed Review post → same dialogs as active approval scenarios

**Workflow Modification Mid-Flight Rules**
- New member added to not-started level: added, notified when level starts
- New member added to in-progress level: added, notified immediately
- New member added to completed level: NOT added retroactively
- Member removed (no post): removed cleanly
- Member removed (post pending): assignment cancelled, member notified
- Member removed (post completed): greyed out, tooltip shown, not removed
- Level removed while in progress: posts at that level auto-advance to the next level; if it was the last level, those posts become fully approved
- Approval rule "Everyone"→"Anyone" on active level with existing approvals: auto-completes and advances
- Approval rule "Anyone"→"Everyone" on completed level: no retroactive invalidation
- Workflow/level name rename: reflects live on all in-flight posts
- Approver removed from workspace: auto-continue level, creator notified

**Notifications — All Channels (in-app, email, mobile push)**
- All new notification types defined in Section 8
- Dedicated approval notification icon in top bar (left of bell)
- All existing notification types updated/extended for multi-level context

**Bug Fix**
- Editing a rejected post no longer silently moves it to draft/scheduled

**Mobile Updates**
- Approval Workflows tab added to Send for Approval in composer (iOS + Android)
- Default workflow pre-selected on Approval Workflows tab

---

### 6.2 Should Have (P1)

- Mark one workflow as Default; auto-selected when Approval Workflows tab opens
- "Allow Approval Workflow Management" permission toggle for Collaborators/Approvers in team member settings
- Re-notify with 2-hour cooldown (button disabled with countdown tooltip)
- Revoke own approval: active level only, not possible if level already closed via "Anyone" rule
- Dedicated approval notification panel with category filters (Pending, Rejected, Updated, Comments)
- Cross-workspace notification badges on workspace switcher
- Duplicate entire workflow (creates "[Name] (Copy)")
- Bulk approve/reject from approver's Pending Approvals view
- "Pending Approvals" bottom nav item on mobile (iOS + Android)
- Approval status badges on all planner views with hover popup
- Reschedule a fully approved post: no re-approval required
- Schedule date change during approval: no re-approval, no notification

---

### 6.3 Nice to Have (P2)

- Visual indicator for re-notify cooldown (timer tooltip showing exact time until re-notify enabled)
- Draft workflow indicator in Settings list (distinct visual treatment)
- Approval workflow analytics (avg approval time per level, % approved vs rejected) — future

---

### 6.4 Explicitly Out of Scope

- Auto-reminders for delayed approvers (v2)
- Image/annotation on post from approver (v2)
- Activity vs Notifications split (platform-wide initiative, separate effort)
- External approval UI redesign (the share link modal itself is unchanged; only the override warning and backend cleanup are added)
- AI-powered approval routing
- Approval for blog posts or other non-social content types
- White-label theming of approval emails (uses existing white-label system)

---

## 7. User Flow (High Level)

### Happy Path — Create Workflow and Use It

1. Admin navigates to Settings → Approval Workflows
2. Clicks "Create a Workflow" tile
3. Enters workflow name, drags team members to Level 1, sets rule to "Everyone"
4. Adds Level 2, drags client to Level 2, sets rule to "Anyone"
5. Saves workflow (marks as Default)
6. Opens Composer, creates a post, clicks "Send for Approval"
7. Right sidebar opens — "Approval Workflows" tab shown, default workflow pre-selected
8. Adds an optional note, clicks "Send"
9. Level 1 approvers receive in-app notification, email, and push notification
10. Level 1 approver opens planner, sees post in Pending Approval state
11. Approver clicks "Approve" with optional comment
12. If "Everyone" rule: all Level 1 approvers approve → Level 2 notified
13. Level 2 (client) approves → post fully approved → moves to Scheduled status
14. Creator receives "fully approved" notification
15. Post publishes at scheduled time

### Rejection Path
1. Level 2 client rejects → post moves to Rejected
2. All other pending Level 2 approvers notified "no longer needed"
3. Creator notified with rejection reason
4. Creator opens post, edits content, save triggers confirmation dialog
5. Creator selects "Resume from Level 2" → Level 2 re-notified
6. Level 2 approves → post fully approved

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-01 | A workflow must have 1–5 levels | Keeps UI manageable; covers all real-world chains |
| BR-02 | Each level must have at least 1 member to be published (saved as draft otherwise) | Prevents empty levels blocking approvals |
| BR-03 | Approval rule defaults to "Everyone must approve" on new levels | Safer default; avoids accidental easy approvals |
| BR-04 | Same team member can appear in multiple levels | Agencies often have one PM reviewing at every stage |
| BR-05 | Only one workflow can be marked as Default per workspace | Clear default selection |
| BR-06 | Default workflow is pre-selected in composer but not auto-submitted | User always confirms before sending |
| BR-07 | Workflow tabs (Users / Approval Workflows) are mutually exclusive per post submission | A post is either custom approval or workflow-based, never both |
| BR-08 | Approval and scheduling are independent: editing schedule date during approval requires no re-approval | Approval is about content, not timing |
| BR-09 | Editing content during active approval triggers a confirmation dialog; no silent status changes | User always controls approval state |
| BR-10 | Editing a rejected post: save triggers confirmation dialog (bug fix — currently silently changes status) | Critical bug fix |
| BR-11 | "Anyone can approve": first approval closes the level; remaining approvers notified "no longer needed" | Prevents unnecessary review work |
| BR-12 | "Anyone can approve": if one person's approval already advanced the level, revoke is not possible (level locked) | Prevents workflow regression |
| BR-13 | Revoke approval is only possible while the approver's level is the currently active level AND the post is not fully approved | Prevents retroactive rework on closed levels |
| BR-14 | On revoke: status resets to pending, creator notified, other active-level approvers notified | Full transparency |
| BR-15 | Re-notify: maximum once every 2 hours per approver per post | Prevents notification spam |
| BR-16 | Rejection notifies ALL other pending approvers at the same level | Prevents wasted review effort |
| BR-17 | Post deletion during approval notifies ONLY current active level approvers | Past/future levels have no action to take |
| BR-18 | Missed Review: only current active level approvers are notified of missed status | Same reasoning as BR-17 |
| BR-19 | Fully approved post rescheduled: no re-approval required, no notifications | Time change is not a content change |
| BR-20 | Workflow deleted with in-flight posts: posts converted to single-user custom approval using current level's approvers | Preserves pending reviews |
| BR-21 | New member added to completed level: NOT added retroactively to in-flight posts | Completed levels are sealed |
| BR-22 | Member removed from workflow while post assigned and pending: their assignment cancelled, they are notified | Immediate effect on in-flight posts |
| BR-23 | Member removed while already approved: greyed out with tooltip, not removed | Preserves approval record |
| BR-24 | Approval rule changed "Everyone"→"Anyone" on active level with ≥1 existing approval: level auto-completes and advances | Rule change applies immediately |
| BR-25 | Approval rule changed "Anyone"→"Everyone" on a completed level: stays completed (no retroactive invalidation) | Protects completed stages |
| BR-26 | Workflow name / level title renamed: reflects live on all in-flight posts | Live reference, not snapshot |
| BR-27 | Approver removed from workspace: level auto-continues without them; creator notified | Removal never blocks a workflow |
| BR-28 | If removed approver was the ONLY member on a level using "Everyone" rule: that level is considered passed | No single-person level should become an infinite blocker |
| BR-29 | Role/permission change: in-flight approvals unaffected | Role change does not remove active assignments |
| BR-30 | "Allow Approval Workflow Management" permission revoked: existing workflows remain active but the user loses edit access immediately | Least-surprise behavior |
| BR-31 | Approval level titles use "Level 01:", "Level 02:" naming format throughout the product consistently | Visual and naming consistency |
| BR-32 | Creator's note (sent with approval request) is separate from the post comments thread | The note is a one-time context message, not part of the ongoing comment thread |
| BR-33 | Approval history is permanently linked to the post; read-only after publish | Audit trail |
| BR-34 | Removing an in-progress level from a workflow causes all posts currently at that level to **auto-advance to the next level** (next level's approvers notified). If the removed level was the **last level**, those posts are **auto-completed (fully approved)** and the creator is notified. | Blocking removal was overly restrictive — auto-advance keeps the workflow moving without orphaning posts |
| BR-35 | Bulk approve/reject: a single comment (optional for approve, prompted for reject) applies to all posts in batch | Consistent UX with existing single-post behavior |
| BR-36 | Draft workflows (incomplete) are NOT shown in the composer's Approval Workflows tab | Prevents users from sending to broken/empty workflows |
| BR-37 | Super Admins and Admins always have workflow management access regardless of permission settings | Admins own workspace configuration |
| BR-38 | Approval notification icon counts ONLY approval-related events; general notifications remain separate | Dedicated channel for time-sensitive approval activity |
| BR-39 | A post can only have ONE active approval at a time — internal workflow, internal custom approval, or external share-link approval. These are mutually exclusive. | Prevents conflicting approval states and notification loops |
| BR-40 | Starting any new approval (any type, any entry point) on a post with an existing active approval replaces the previous approval entirely | Verified current behavior in `ApprovalBuilder.php` and `ShareLinkController::setupPlansForExternalApproval` — new approval always overwrites |
| BR-41 | Before any override takes effect, the user must see a warning. For single post: an Alert banner in the sidebar. For bulk: a confirmation Dialog before the sidebar opens. For external share link: an Alert in the share modal. | No silent cancellations — user is always informed |
| BR-42 | On override: cancellation notifications are sent to current active level approvers only (same scope as post-deletion) — in-app + email | Mirrors the post-deletion notification rule (BR-17); past and future levels are not notified |
| BR-43 | "Send for Approval" is available directly on single posts in the Planner (post card 3-dot menu and post detail view), not only from the Composer | Codebase already has bulk Planner send (`DataTable.vue`); single-post is a new addition |
| BR-44 | Backend fix required: `ShareLinkController::setupPlansForExternalApproval` must fully clean up internal approval state before writing external approval — currently this cleanup is missing (one-way bug: internal→external has no `removeInternalApprovalState` equivalent) | Data consistency; mirrors `PlanController::sendPlansForBulkApprovers` which already calls `removeExternalApproverActions` when internal overrides external |
| BR-45 | When saving edits to a workflow that has in-flight posts, a confirmation dialog is shown before saving (not just a banner warning). Title: "Save changes to [Workflow Name]?" — user must explicitly confirm "Save & Apply Changes" or cancel. | The banner alone is insufficient — changes that affect live posts require explicit user intent |
| BR-46 | Auto-advance on level removal: two notifications dispatched per affected post — (1) creator notified that the level was removed and post advanced; (2) next level's approvers receive `workflow_level_advanced` notification. If last level removed → creator receives `post_fully_approved` notification. | Mirrors the normal level-completion notification flow |

---

## 9. Open Questions

| Question | Options | Owner | Decision |
|---|---|---|---|
| Should draft workflows appear in Settings list with a visual indicator, or be hidden? | Show with "Draft" badge vs hide entirely | Product | Show with "Draft" badge |
| When a workflow is duplicated, should the copy also inherit the "Default" status? | Yes / No | Product | No — only one default; duplicate starts without default |
| Should the "Pending Approvals" mobile nav item show a badge count? | Yes (unread count) / No | Product / Mobile | Yes — same logic as notification icon |
| Should Automations support the new Approval Workflows tab (workflow selection)? | Yes / No / Phase 2 | Product | Yes — same sidebar component; should be updated in same release |

---

## 10. Notifications — Complete Specification

This section defines ALL notification events, copy, email subjects, and channels. Existing notification infrastructure uses `ApprovalObserver` → Redis queue → `ApprovalNotification` → email + broadcast.

New notification types extend the existing `notifications.php` localization file and `request_approval.blade.php` email template.

---

### 10.1 Existing Notification Types (Updated)

#### `pending_approval` — Post Sent for Approval (Single-User or Workflow Level 1)

**Updated CTA text:** Approval emails reference approve/reject actions only.

| Field | Value | Notes |
|---|---|---|
| In-app title | `:requested_by_name has sent you content to review` | Unchanged |
| In-app description | `<strong>:requested_by_name</strong> has sent you :post_text for review in <strong>:workspace_name</strong>` | For workflow: append `(Level :level_number: :level_title)` |
| Email subject | `Approval required for :post_text – :workspace_name` | Unchanged |
| Email CTA text | `Please review the post and approve or reject it.` | Applies to single-user and workflow approvals |
| Email button | `Review Post` | Updated label |
| Channels | In-app, email, mobile push | Same |
| Who is notified | All selected approvers (single-user) or Level 1 approvers (workflow) | Same |

---

#### `approve_approval` — Approver Approved a Post (to Creator)

| Field | Value |
|---|---|
| In-app title | `:approver_name has approved :post_text` |
| In-app description | `Your post in <strong>:workspace_name</strong> has been approved by :approver_name.` + for workflow: `Level :level_number (:level_title) is complete.` |
| Email subject | `Your post has been approved – :workspace_name` |
| Email headline | `:approver_name approved your post` |
| Email body | `:approver_name has approved :post_text in :workspace_name.` + for workflow: `Level :level_number (:level_title) has been completed. The post will move to the next approval level.` |
| Email CTA text | For workflow (levels remain): `The post will now move to the next approval level.` / For single-level complete: `No further action is required at this time.` |
| Email button | `View Post` |
| Channels | In-app, email |
| Who is notified | Post creator |

---

#### `reject_approval` — Approver Rejected a Post (to Creator)

| Field | Value |
|---|---|
| In-app title | `:approver_name has rejected :post_text` |
| In-app description | `Your post in <strong>:workspace_name</strong> was rejected by :approver_name. Please review the feedback and update the post.` |
| Email subject | `Your post has been rejected – :workspace_name` |
| Email headline | `:approver_name rejected your post` |
| Email body | `:approver_name has rejected :post_text in :workspace_name.` + if comment: `Their feedback: ":last_action_note"` |
| Email CTA text | `Please review the feedback and update the post when ready.` |
| Email button | `Review Post` |
| Channels | In-app, email, mobile push |
| Who is notified | Post creator |

---

### 10.2 New Notification Types

#### `workflow_level_advanced` — Level Complete, Next Level Now Pending

Sent to approvers at the NEXT level when the previous level is fully completed.

| Field | Value |
|---|---|
| In-app title | `Your approval is needed for :post_text` |
| In-app description | `Level :previous_level_number (:previous_level_title) of <strong>:workflow_name</strong> is complete. Your review is now needed at Level :level_number (:level_title) in <strong>:workspace_name</strong>.` |
| Email subject | `Your approval is needed for :post_text – :workspace_name` |
| Email headline | `It's your turn to review` |
| Email body | `Level :previous_level_number (:previous_level_title) has been completed for :post_text in :workspace_name. You are Level :level_number (:level_title) in the :workflow_name workflow. Please review and approve or reject the post.` |
| Email CTA text | `Please review the post and approve or reject it.` |
| Email button | `Review Post` |
| Channels | In-app, email, mobile push |
| Who is notified | All members of the newly active level |

---

#### `post_fully_approved` — Post Approved at All Levels (to Creator)

Sent when the final level of a workflow completes.

| Field | Value |
|---|---|
| In-app title | `:post_text has been fully approved` |
| In-app description | `All :total_levels levels of <strong>:workflow_name</strong> have been completed for your post in <strong>:workspace_name</strong>. The post will proceed as scheduled.` |
| Email subject | `Your post has been fully approved – :workspace_name` |
| Email headline | `All approvals complete — your post is ready` |
| Email body | `Great news! :post_text has received approval from all :total_levels levels of the :workflow_name workflow in :workspace_name. The post will proceed to its scheduled publication time.` |
| Email CTA text | `No further action is required.` |
| Email button | `View Post` |
| Channels | In-app, email, mobile push |
| Who is notified | Post creator |

---

#### `level_no_action_needed` — "Anyone" Rule: Level Complete, Others No Longer Needed

Sent to remaining approvers at a level when "Anyone can approve" rule causes the level to close after one approval.

| Field | Value |
|---|---|
| In-app title | `No action needed — :post_text has moved forward` |
| In-app description | `Level :level_number (:level_title) of <strong>:workflow_name</strong> has been completed by another reviewer in <strong>:workspace_name</strong>. Your review is no longer needed at this level.` |
| Email subject | `No action needed — :post_text in :workspace_name` |
| Email headline | `No action needed` |
| Email body | `The approval at Level :level_number (:level_title) for :post_text in :workspace_name has been completed by another reviewer. Your review is no longer needed at this level.` |
| Email CTA text | `No action is required from you.` |
| Email button | `View Post` (secondary, no primary CTA) |
| Channels | In-app, email |
| Who is notified | All other pending approvers at the same level |

---

#### `rejection_no_action_needed` — Post Rejected, Others No Longer Needed

Sent to other pending approvers at the same level when any one approver rejects.

| Field | Value |
|---|---|
| In-app title | `:rejecter_name has rejected :post_text` |
| In-app description | `:rejecter_name has rejected :post_text in <strong>:workspace_name</strong>. Your review is no longer needed for this post.` |
| Email subject | `No action needed — :post_text has been rejected – :workspace_name` |
| Email headline | `No action needed` |
| Email body | `:rejecter_name has rejected :post_text in :workspace_name. Your approval is no longer needed.` |
| Email CTA text | `No action is required from you at this time.` |
| Email button | `View Post` (secondary) |
| Channels | In-app, email |
| Who is notified | All other pending approvers in the same level (not the rejector, not the creator) |

---

#### `post_deleted_approval` — Post Deleted While in Approval

| Field | Value |
|---|---|
| In-app title | `:creator_name has deleted a post` |
| In-app description | `A post you were assigned to review in <strong>:workspace_name</strong> has been deleted by :creator_name. No further action is required.` |
| Email subject | `Post deleted — no action needed – :workspace_name` |
| Email headline | `Post deleted` |
| Email body | `:creator_name has deleted a post that was pending your approval in :workspace_name. No further action is required.` |
| Email CTA text | `No action is required from you.` |
| Email button | None |
| Channels | In-app, email |
| Who is notified | Current active level approvers ONLY (not past levels, not future levels) |

---

#### `approval_revoked_creator` — Approver Revoked Their Approval (to Creator)

| Field | Value |
|---|---|
| In-app title | `:approver_name has revoked their approval for :post_text` |
| In-app description | `:approver_name has revoked their approval for :post_text in <strong>:workspace_name</strong>. The post is back to pending at Level :level_number (:level_title).` |
| Email subject | `:approver_name revoked their approval – :workspace_name` |
| Email headline | `Approval revoked` |
| Email body | `:approver_name has revoked their approval for :post_text in :workspace_name. The post is pending again at Level :level_number (:level_title).` |
| Email CTA text | `No action is required from you at this time.` |
| Email button | `View Post` |
| Channels | In-app, email |
| Who is notified | Post creator |

---

#### `approval_revoked_approvers` — Approver Revoked, Other Same-Level Approvers Notified

| Field | Value |
|---|---|
| In-app title | `:approver_name has revoked their approval for :post_text` |
| In-app description | `:approver_name has revoked their approval for :post_text at Level :level_number (:level_title) in <strong>:workspace_name</strong>. Your review may now be required.` |
| Email subject | `Action may be needed — :approver_name revoked approval – :workspace_name` |
| Email headline | `Approval revoked by :approver_name` |
| Email body | `:approver_name has revoked their approval for :post_text at Level :level_number (:level_title) in :workspace_name. Depending on the approval rule for this level, your action may now be required.` |
| Email CTA text | `Please review the post if action is needed.` |
| Email button | `Review Post` |
| Channels | In-app, email |
| Who is notified | Other approvers at the same active level (not the person who revoked) |

---

#### `content_updated_approval` — Post Content Updated While in Approval

Sent when creator edits post content and selects "Re-notify approvers", "Re-notify current level", or "Restart from Level 1".

| Field | Value |
|---|---|
| In-app title | `:creator_name has updated :post_text` |
| In-app description | `:creator_name has updated the content of :post_text that is pending your approval in <strong>:workspace_name</strong>. Please review the updated content.` |
| Email subject | `:creator_name updated a post pending your review – :workspace_name` |
| Email headline | `Post content updated` |
| Email body | `:creator_name has made changes to :post_text in :workspace_name. The post content has been updated and requires your review again.` + if note attached: `Note from :creator_name: ":note"` |
| Email CTA text | `Please review the updated post and approve or reject it.` |
| Email button | `Review Post` |
| Channels | In-app, email, mobile push |
| Who is notified | Approvers based on creator's dialog choice: all approvers (Re-notify all) / current level only (Re-notify current) / Level 1 and above (Restart) |

---

#### `re_notify` — Manual Re-notify by Creator

| Field | Value |
|---|---|
| In-app title | `Reminder: your review is needed for :post_text` |
| In-app description | `:creator_name is waiting for your approval on :post_text in <strong>:workspace_name</strong>.` |
| Email subject | `Reminder: your approval is needed for :post_text – :workspace_name` |
| Email headline | `Friendly reminder — your review is needed` |
| Email body | `:creator_name is waiting for your approval on :post_text in :workspace_name. The post is currently at Level :level_number (:level_title).` |
| Email CTA text | `Please review the post and approve or reject it.` |
| Email button | `Review Post` |
| Channels | In-app, email, mobile push |
| Who is notified | The specific approver re-notified |
| Cooldown | Maximum once per approver per post per 2 hours. UI button disabled with tooltip: `"Re-notified less than 2 hours ago. Available again at [HH:MM AM/PM]."` |

---

#### `level_removed_auto_advanced` — Level Removed from Workflow, Post Auto-Advanced (to Creator)

Sent to the post creator when an admin removes a level that a post was currently at, causing the post to automatically advance.

| Field | Value |
|---|---|
| In-app title | `"Your post has been moved forward in [Workflow Name]"` |
| In-app description | `"Level :removed_level_number (:removed_level_title) was removed from the :workflow_name workflow. Your post :post_text has automatically advanced to Level :next_level_number (:next_level_title) in :workspace_name."` |
| Email subject | `"Your post advanced in the approval workflow — :workspace_name"` |
| Email headline | `"A level was removed — your post moved forward"` |
| Email body | `"Level :removed_level_number (:removed_level_title) was removed from the :workflow_name workflow. Your post :post_text in :workspace_name has automatically advanced to Level :next_level_number (:next_level_title). The next reviewers have been notified."` |
| Email CTA text | `"No action is required from you at this time."` |
| Email button | `"View Post"` |
| Channels | In-app, email |
| Who is notified | Post creator |

Next level's approvers are notified via the existing `workflow_level_advanced` notification type.

---

#### `level_removed_auto_approved` — Last Level Removed, Post Auto-Approved (to Creator)

Sent to the post creator when the removed level was the final level, causing the post to be fully approved.

| Field | Value |
|---|---|
| In-app title | `":post_text has been fully approved"` |
| In-app description | `"The final level of :workflow_name was removed, and your post :post_text in :workspace_name has been automatically fully approved. It will proceed as scheduled."` |
| Email subject | `"Your post has been fully approved — :workspace_name"` |
| Email headline | `"Your post is fully approved"` |
| Email body | `"The final approval level in the :workflow_name workflow was removed. As a result, your post :post_text in :workspace_name has been automatically approved and will proceed to its scheduled publication time."` |
| Email CTA text | `"No further action is required."` |
| Email button | `"View Post"` |
| Channels | In-app, email, mobile push |
| Who is notified | Post creator |

---

#### `approver_removed_post` — Approver Removed from Workspace, Affecting Creator's Post

| Field | Value |
|---|---|
| In-app title | `An approver was removed from :post_text` |
| In-app description | `:member_name was removed from the workspace and has been removed from the approval of :post_text in <strong>:workspace_name</strong>. Level :level_number (:level_title) will continue with the remaining approvers.` |
| Email subject | `Approver removed — :post_text in :workspace_name` |
| Email headline | `:member_name was removed from this approval` |
| Email body | `:member_name was removed from your workspace and has been automatically removed from the approval workflow for :post_text. The approval process at Level :level_number (:level_title) will continue with the remaining approvers.` |
| Email CTA text | `No action is required unless you'd like to reassign this review.` |
| Email button | `View Post` |
| Channels | In-app, email |
| Who is notified | Post creator (one notification per affected post) |

---

### 10.3 Email Template Notes

All emails use the existing `resources/views/emails/notifications/post/request_approval.blade.php` template. The following variables must be passed for new notification types:

| New Variable | Used In | Value |
|---|---|---|
| `$workflow_name` | All workflow notifications | Name of the approval workflow |
| `$level_number` | Level-specific notifications | Current level number (1, 2, 3...) |
| `$level_title` | Level-specific notifications | Custom title of the level |
| `$total_levels` | `post_fully_approved` | Total number of levels in workflow |
| `$previous_level_number` | `workflow_level_advanced` | Previous (completed) level number |
| `$previous_level_title` | `workflow_level_advanced` | Previous level title |
| `$rejecter_name` | `rejection_no_action_needed` | Name of the approver who rejected |
| `$approver_name` | `approval_revoked_*`, `approve_approval` | Name of the approver |
| `$creator_name` | `content_updated_approval`, `post_deleted_approval`, `re_notify` | Name of post creator |
| `$member_name` | `approver_removed_post` | Name of removed workspace member |

The conditional CTA button behavior in `request_approval.blade.php` (lines 49-52):
- `pending_approval`, `reject_approval`, all new "action needed" types → primary button: "Review Post"
- `approve_approval`, `post_fully_approved`, all "no action needed" types → button: "View Post"
- `post_deleted_approval` → no CTA button

---

## 11. UI Copy — Complete Specification

### Settings — Approval Workflows Landing Page

| Element | Copy |
|---|---|
| Page title | `Approval Workflows` |
| Left nav item | `Approval Workflows` |
| Empty state headline | `No Approval Workflows Yet` |
| Empty state subtext | `Create approval workflows to streamline your team's content review process. Set up multi-level approvals with custom rules for each stage.` |
| Empty state CTA button | `Create Workflow` |
| "Create a Workflow" tile headline | `Create a Workflow` |
| "Create a Workflow" tile example text | `e.g. Campaign Approvals` |
| Default badge on tile | `Default` |
| 3-dot menu: set default | `Set as Default` |
| 3-dot menu: remove default | `Remove as Default` |
| 3-dot menu: edit | `Edit` |
| 3-dot menu: duplicate | `Duplicate` |
| 3-dot menu: delete | `Delete` |
| Workflow card: level count | `[N] Levels` |
| No permission state (workflow list is read-only) | `You don't have permission to manage approval workflows. Contact your workspace admin to make changes.` |

---

### Settings — Create / Edit Workflow

| Element | Copy |
|---|---|
| Page title (create) | `Create New Workflow` |
| Page title (edit) | `Edit Workflow` |
| Workflow name input placeholder | `Title of Workflow` |
| Workflow name input label | `Workflow Name` |
| Draft badge | `Draft` |
| Draft state hint | `This workflow is incomplete and won't appear in the composer until all levels have at least one member.` |
| 3-dot menu in header | (Options: `Duplicate Workflow`, `Delete Workflow`) |
| Right panel header | `Members` |
| Right panel hint | `Drag & drop members to workflow levels` |
| Right panel filter tabs | `All` / `Team` / `Client` |
| Right panel search placeholder | `Search members...` |
| Right panel empty search | `No members found` |
| Level title placeholder | `Level 01:` / `Level 02:` / etc. |
| Level 3-dot menu: duplicate | `Duplicate Level` |
| Level 3-dot menu: delete | `Delete Level` |
| Delete level confirmation title | `Delete this level?` |
| Delete level confirmation body | `This will remove Level [N] and all its assigned members. If any in-flight posts are currently at this level, they'll automatically advance to the next level. If this was the last level, those posts will be fully approved.` |
| Delete level confirmation buttons | `Delete Level` / `Cancel` |
| "Post needs approval from:" label | `Post needs approval from:` |
| "Everyone" radio label | `Everyone` |
| "Anyone" radio label | `Anyone` |
| "Everyone" tooltip (standard) | `All assigned team members must approve before this level is complete and the post moves to the next step.` |
| "Anyone" tooltip (standard) | `The level is complete as soon as any one team member approves. Other assigned members are notified that their review is no longer needed.` |
| "Everyone" tooltip (< 2 members) | `Add at least 2 team members to enable this option.` |
| Add level button | `+ Add New Level` |
| Add level button (at max 5) | `+ Add New Level` (disabled, tooltip: `Maximum 5 levels per workflow.`) |
| Save button | `Save Workflow` (create) / `Save Changes` (edit) |
| Save as draft button | `Save as Draft` |
| Cancel button | `Cancel` |
| Member drag placeholder in level | `Drag and drop people from Members` |
| Member overflow badge | `+[N] more` |
| Duplicate workflow creates | `[Workflow Name] (Copy)` |

**Workflow Deletion Confirmations:**

No in-flight posts:
- Title: `Delete "[Workflow Name]"?`
- Body: `This approval workflow will be permanently deleted. This action cannot be undone.`
- Buttons: `Delete Workflow` (destructive) / `Cancel`

With in-flight posts:
- Title: `Delete "[Workflow Name]"?`
- Body: `[N] post(s) are currently in this approval workflow. Deleting it will convert those posts to single-user approval — the currently assigned approvers at each post's active level will remain assigned. You'll need to manage them manually.`
- Buttons: `Delete Workflow` / `Cancel`

**Workflow Modification Mid-Flight Warning Banner** (shows at top of edit page when workflow has in-flight posts):
- `⚠ [N] post(s) are currently using this workflow. Some changes may affect those posts.`

**Save Confirmation Dialog** (shown on "Save Workflow" / "Save Changes" when in-flight posts exist):

Base (no level deletions):
- Title: `Save changes to "[Workflow Name]"?`
- Body: `[N] post(s) are currently in approval using this workflow. Saving these changes will immediately affect their approval process.`
- Primary button: `Save & Apply Changes`
- Secondary button: `Cancel`

With level deletion(s):
- Title: `Save changes to "[Workflow Name]"?`
- Body: `[N] post(s) are currently in approval using this workflow. Saving these changes will immediately affect their approval process.`
- Additional line: `[N] post(s) at the removed level(s) will automatically advance to the next level. If a removed level was the last, those posts will be fully approved.`
- Primary button: `Save & Apply Changes`
- Secondary button: `Cancel`

---

### Send for Approval — Sidebar

| Element | Copy |
|---|---|
| Panel title | `Send for Approval` |
| Panel info tooltip (ℹ) | `Send this post to your team for review before publishing. The post won't be published until the approval process is complete.` |
| Close button | `✕` |
| Tab 1 | `Users` |
| Tab 2 | `Approval Workflows` |
| Search placeholder (Users tab) | `Search by name...` |
| No results (Users tab) | `No team members found` |
| "Post needs approval from:" label | `Post needs approval from:` |
| "Everyone" radio | `Everyone` |
| "Anyone" radio | `Anyone` |
| "Everyone" tooltip (< 2 selected) | `Select at least 2 team members to use this option.` |
| "Everyone" tooltip (≥ 2 selected) | `All selected team members must approve before the post is scheduled.` |
| "Anyone" tooltip | `The post will be scheduled as soon as any one selected team member approves.` |
| Notes placeholder (Users tab) | `Add a note for the approver(s)...` |
| Notes placeholder (Workflows tab) | `Add a note for this approval request...` |
| "Send" button | `Send` |
| "Cancel" button | `Cancel` |
| Restricted account warning on member card (tooltip) | `This team member only has access to certain social accounts. Make sure the post's accounts align with their permissions.` |
| Permission warning modal title | `Restricted Access` |
| Permission warning modal body | `Some selected team members only have access to specific social accounts. Sending this post for approval may limit their ability to review all content. Do you want to proceed?` |
| Permission warning confirm button | `Send Anyway` |
| Minimum approver error (no one selected) | `Please select at least one approver.` |

**Approval Workflows Tab:**

| Element | Copy |
|---|---|
| Empty state (no workflows) | `No approval workflows yet.` |
| Empty state subtext | `Create one in Settings → Approval Workflows.` |
| Default badge on workflow card | `Default` |
| Collapsed card: level count | `[N] Levels` |
| Expanded level label | `Level [N]: [Title]` |
| Rule label (expanded) | `Everyone` / `Anyone` |
| "Edit in Settings" link | `Edit in Settings →` |
| Edit redirect confirmation title | `Leave the composer?` |
| Edit redirect confirmation body | `You'll be redirected to Approval Workflow settings. Your composer content will be minimized but not lost — it will be here when you return.` |
| Edit redirect confirm button | `Go to Settings` |
| Edit redirect cancel button | `Stay Here` |

---

### Approval Status Panel — Planner Hover Popup

| Element | Copy |
|---|---|
| Popup header (workflow) | `[Workflow Name]` |
| Popup header (custom approval) | `Approval Status` |
| Level row header | `Level [N]: [Title]` |
| Level status badge: completed | `Completed` |
| Level status badge: in progress | `In Progress` |
| Level status badge: not started | `Not Started` |
| Per-user: approved | `Approved on [date]` |
| Per-user: rejected | `Rejected on [date]` |
| Per-user: pending | `Awaiting approval` |
| Re-notify button | `Re-notify` |
| Re-notify disabled tooltip | `Re-notified less than 2 hours ago. Available again at [HH:MM AM/PM].` |
| Counter badge on post card | `[approved]/[total]` (e.g. `1/3`) |

---

### Post Preview Modal — Approval Status Panel

| Element | Copy |
|---|---|
| Panel title | `Approval Status` |
| Workflow name subheading | `[Workflow Name]` |
| Custom approval subheading | `Custom Approval` |
| Creator's note label | `Request note from [Creator name]:` |
| Approve button | `Approve` |
| Reject button | `Reject` |
| Approve confirmation title | `Approve Post` |
| Approve confirmation body | `You're about to approve this post. You can add an optional comment.` |
| Bulk approve confirmation title | `Approve [N] Posts` |
| Bulk approve confirmation body | `You're about to approve [N] posts. Add an optional comment.` |
| Approve comment placeholder | `Add your comment here (optional)` |
| Approve confirm button | `Yes, Approve` |
| Reject confirmation title | `Reject Post` |
| Reject confirmation body | `You're about to reject this post. Please add a comment explaining what needs to change.` |
| Reject comment placeholder | `Explain what needs to change...` |
| Reject confirm button | `Reject Post` |
| Revoke button (visible to approver only) | `Revoke approval` |
| Revoke confirmation title | `Revoke your approval?` |
| Revoke confirmation body | `Your approval will be reset to pending. [Creator name] and other reviewers at this level will be notified.` |
| Revoke confirm button | `Revoke` |
| Revoke cancel button | `Cancel` |
| Revoke not available tooltip (level locked) | `This level has already advanced — your approval can no longer be revoked.` |
| Re-notify button | `Re-notify` |
| Re-notify disabled tooltip | `Re-notified less than 2 hours ago. Available again at [HH:MM AM/PM].` |
| Fully approved post: panel footer | `All levels completed — this post is approved.` |
| Post-publish: panel footer | `This post has been published. Approval history is preserved for reference.` |

---

### Edit Confirmation Dialogs

**Scenario 1: Edit post in active single-user approval**

- Title: `This post is pending approval`
- Body: Shows list of approvers with their current status
- Question: `You've updated the content. What would you like to do?`
- Option 1: `Re-notify approvers` — sub-label: `All pending and approved statuses reset. Approvers will be re-notified.`
- Option 2: `Keep current approval status` — sub-label: `Save changes without affecting the approval process.`
- Option 3: `Remove approval` — sub-label: `The post exits the approval process and reverts to draft/scheduled.`
- Cancel link: `Cancel`

**Scenario 2: Edit post in active workflow approval**

- Title: `This post is in an approval workflow`
- Body: Shows workflow name + level breakdown with statuses
- Question: `You've updated the content. What would you like to do?`
- Option 1: `Restart from Level 1` — sub-label: `All approvals reset. The workflow starts over from the beginning.`
- Option 2: `Re-notify current level` — sub-label: `Only Level [N] resets. Previous levels stay approved.`
- Option 3: `Keep current approval status` — sub-label: `Save changes without affecting the workflow.`
- Option 4: `Remove approval` — sub-label: `The post exits the approval workflow and reverts to draft/scheduled.`
- Cancel link: `Cancel`

**Scenario 3: Edit rejected post (single-user)**

- Title: `This post was rejected`
- Body: Shows list with rejected/approved statuses
- Question: `You've updated the content. What would you like to do?`
- Option 1: `Resend for approval` — sub-label: `All statuses reset to pending. Approvers will be re-notified.`
- Option 2: `Remove approval` — sub-label: `The post moves to draft/scheduled without requiring approval.`
- Cancel link: `Cancel`

**Scenario 4: Edit rejected post (workflow)**

- Title: `This post was rejected at Level [N] of [Workflow Name]`
- Body: Shows workflow level breakdown with rejection detail
- Question: `You've updated the content. What would you like to do?`
- Option 1: `Restart from Level 1` — sub-label: `All approvals reset. The workflow starts over from the beginning.`
- Option 2: `Resume from Level [N]` — sub-label: `Only Level [N] resets. Previous levels' approvals are preserved.`
- Option 3: `Remove approval` — sub-label: `The post moves to draft/scheduled without requiring approval.`
- Cancel link: `Cancel`

**Scenario 5: Edit post in Missed Review state** → Same copy as Scenarios 1–2 (depending on single-user or workflow)

---

### Approval Override — Single Post Warning (in sidebar)

Shown as an `Alert` (warning variant) at the top of the Send for Approval sidebar when the post already has an active approval:

| Scenario | Alert copy |
|---|---|
| Post in internal custom approval | `"This post is already in an approval process. Starting a new one will cancel it — currently assigned approvers will be notified."` |
| Post in internal workflow approval | `"This post is in [Workflow Name] (Level [N]). Starting a new approval will cancel it — Level [N] approvers will be notified."` |
| Post in external share-link approval | `"This post has an active external approval via share link. Starting an internal approval will cancel it."` |

No blocking confirmation needed — user can proceed by filling in the sidebar and clicking Send.

---

### Approval Override — Bulk Send Confirmation Dialog

Shown before the sidebar opens when one or more selected posts already have an active approval:

- Title: `"Some posts are already in approval"`
- Body: `"[N] of [M] selected posts already have an active approval process. Starting a new one will cancel their current approvals — assigned approvers will be notified."`
- Primary `Button`: `"Continue"`
- Secondary `Button`: `"Cancel"`

---

### Approval Override — External Share Link Warning (in share modal)

Shown as an `Alert` (warning variant) inside the share modal when the user toggles on "Send for Approval" and the post has an active internal approval:

- `"This post is currently in an internal approval process. Sending it for external approval will cancel it — currently assigned approvers will be notified."`

---

### Team Member Removal Confirmation (when member has in-flight approvals)

- Title: `Remove [Member Name]?`
- Body: `This member is assigned as an approver on [N] pending post(s) and is part of [N] approval workflow(s). Removing them will automatically cancel their pending approvals. The affected workflows will continue with the remaining approvers, and post creators will be notified.`
- Confirm button: `Remove Member`
- Cancel button: `Cancel`

---

### Approval Notification Panel (Dedicated Icon)

| Element | Copy |
|---|---|
| Panel title | `Approvals` |
| Empty state headline | `No approval activity yet` |
| Empty state subtext | `When posts are sent for approval or reviewed, activity will appear here.` |
| Category tab: all | `All` |
| Category tab: pending | `Pending` |
| Category tab: rejected | `Rejected` |
| Category tab: updated | `Updated` |
| Category tab: comments | `Comments` |
| Notification item link | `View post →` |
| Mark all read button | `Mark all as read` |
| Notification icon tooltip | `Approval notifications` |
| Badge (unread count) | Numeric badge (matches existing bell pattern) |

---

### Team Member Permissions Settings

| Element | Copy |
|---|---|
| Permission toggle label | `Manage Approval Workflows` |
| Permission toggle description | `Allow this member to create, edit, duplicate, delete, and set default approval workflows.` |
| Permission off state | `Only admins can manage workflows` |

---

### Mobile — "Pending Approvals" Nav Item

| Element | Copy |
|---|---|
| Bottom nav label | `Pending Approvals` |
| Screen title | `Pending Approvals` |
| Segment: needs my approval | `Needs My Approval` |
| Segment: my requests | `My Requests` |
| Empty state (needs my approval) | `You're all caught up!` / `No posts are waiting for your approval right now.` |
| Empty state (my requests) | `No pending requests` / `Posts you've sent for approval will appear here.` |

---

### Planner Status Labels

| Status | Display Label |
|---|---|
| `review` / `pending_approval` | `Pending Approval` |
| `missed_review` | `Missed Review` |
| `rejected_approval` / `rejected` | `Rejected` |

**Note:** "Under Review" should be confirmed as renamed to "Pending Approval" — check existing `planner.json` i18n key `planner.filter_sidebar.plan_approval_status.*` and update accordingly.

---

## 12. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Workflow modification mid-flight creates unexpected state | Medium | High | Comprehensive business rules (Section 8); thorough testing matrix for each edit scenario |
| Notification volume overwhelms approvers | Medium | Medium | Re-notify cooldown (2hr); "no longer needed" notifications prevent stale tasks |
| Mobile push notification infrastructure doesn't support all new event types | Medium | Medium | Audit existing FCM/APNs setup in iOS and Android before BE implementation; new event types map to existing push dispatch system |
| New approval system introduces parallel backend logic alongside existing single-level approval | High | High | Backend team to implement as a fully separate system; existing single-level path must remain unchanged for backwards compatibility |
| New email notifications have different layout needs (level hierarchy, workflow names) | Low | Medium | Backend team to create a separate email template for workflow notifications; not mixing with existing approval email |
| Bulk approve/reject doesn't scale for large post volumes | Low | Medium | Backend team to handle pagination/chunking as needed in their implementation |
| Composer modal → sidebar migration breaks existing integrations (automations, bulk actions) | Medium | High | Shared sidebar component used in all contexts; automations team involved in testing |
| Draft workflows accidentally shown in composer | Low | High | API must filter draft workflows from the composer endpoint response |
| iOS/Android lack workflow-specific approval infrastructure | High | Medium | New Approval Workflows tab in mobile composer; API payload is identical to web — only UI addition needed |

---

## 13. Dependencies

**API contract (backend must provide):**
- `GET /workspaces/{id}/approval-workflows` — list all published workflows for a workspace
- `POST /workspaces/{id}/approval-workflows` — create a workflow
- `GET /approval-workflows/{id}` — single workflow with levels, members, and `in_flight_posts_count`
- `PUT /approval-workflows/{id}` — update workflow; returns `requires_save_confirmation` when in-flight posts exist; accepts `?confirmed=true` to apply
- `DELETE /approval-workflows/{id}` — delete; returns `in_flight_count` for confirmation; accepts `?force=true`
- `POST /approval-workflows/{id}/duplicate` — duplicate workflow
- `PUT /approval-workflows/{id}/set-default` / `remove-default` — manage default flag
- `POST /plans/{id}/approve` — approve at current level; returns `is_missed_review` flag when applicable
- `POST /plans/{id}/reject` — reject with comment
- `POST /plans/{id}/revoke-approval` — revoke own approval; returns 422 if level already advanced
- `POST /plans/{id}/re-notify/{member_id}` — re-notify a specific approver; enforces 2hr cooldown; returns `next_available_at` on 429
- `POST /plans/{id}/save-with-approval-action` — save post content with an explicit approval action choice (re-notify / keep / restart / remove etc.)
- `GET /plans/{id}/approval-history` — full audit trail
- `DELETE /workspace-members/{id}` — returns `in_flight_posts_count` + `workflow_definitions_count` before removal; accepts `?confirmed=true`

**Plan API response must include** (when plan has active approval):
- `approval.workflow_id`, `approval.workflow_name`, `approval.current_level`, `approval.total_levels`
- `approval.workflow_levels[]` — per-level: title, rule, members with status + timestamp + comment
- `approval.level_status[]` — per-level aggregate: not_started / in_progress / completed / rejected
- `approval_confirmation_required` object when saving content on a post in active/rejected approval
- `is_missed_review: true` when approver acts on a missed review post

**Real-time (Pusher):**
- Existing approval socket events must include workflow fields in payload for real-time UI updates

**Mobile push (FCM / APNs):**
- All new notification event types (Section 10) must be dispatched via existing push infrastructure

**Frontend dependencies:**
- New `SendForApprovalSidebar.vue` component — must be complete before Composer, Planner, and Automations can consume it
- Existing `useApproval.js` composable and `usePublishApprovalStore.ts` extended for workflow state
- Planner bulk action menu and Automations "Send for Approval" action updated to use new sidebar

**Blockers:**
- API contract (endpoint shapes + response schemas) must be agreed between FE and BE before parallel development begins
- Shared sidebar component architecture agreed before FE work splits across Composer, Planner, Automations
- iOS/Android: confirm whether new bottom nav item requires App Store submission timeline alignment

---

## 14. Appendix

**Frontend i18n keys to update/extend:**
- `composer.approval_modal.*` — update for sidebar rename and new workflow tab
- `planner.filter_sidebar.plan_approval_status.*` — "Under Review" → "Pending Approval"; add level-specific labels
- `planner.approval_confirmation.*` — extend for workflow approve/reject dialogs

**Backend technical spec:** To be created separately by the backend team, covering architecture, file structure, data models, and implementation approach.

**Figma:** https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034
**Workflow Design:** `docs/features/multi-tiered-approval-workflow/02-workflow.md`

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-03-25 | Product | Initial PRD based on approved workflow design |
