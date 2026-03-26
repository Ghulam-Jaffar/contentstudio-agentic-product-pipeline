# Workflow Design: Multi-Tiered Approval Workflow

## 1. Feature Placement

### Navigation & Entry Points

| Location | Element |
|---|---|
| **Workspace Settings** | New "Approval Workflow" item in left sidebar menu |
| **Composer** | "Send for Approval" button opens sidebar with 2 tabs: Users / Approval Workflows |
| **Planner** | Approval status badges on post cards (all views: calendar, feed, list). Post preview modal shows full detail. |
| **Planner — Bulk Actions** | Select multiple posts → "Send for Approval" → same sidebar panel |
| **Automations** | Existing "Send for Approval" action updated to use new sidebar component |
| **Top Bar** | New dedicated approval notification icon (next to existing notifications & settings icons) |
| **Mobile Apps (iOS & Android)** | New "Pending Approvals" bottom nav item. Updated "Send for Approval" in composer with workflow tab. |
| **Cross-Workspace** | Notification badges on workspace switcher for approval activity in other workspaces |

---

## 2. Module Breakdown

### A. Settings — Approval Workflow Management

#### Role-Based Permissions

| Action | Super Admin | Admin | Collaborator | Approver |
|---|---|---|---|---|
| Manage workflows (create/edit/delete/default) | ✅ Always | ✅ Always | ✅ If permission granted | ✅ If permission granted |
| Send post for approval | ✅ | ✅ | ✅ | ✅ |
| Be added as approver in a level | ✅ | ✅ | ✅ | ✅ |
| Approve / Reject | ✅ | ✅ | ✅ | ✅ |
| Re-notify | ✅ | ✅ | ✅ | ✅ |
| Revoke own approval | ✅ | ✅ | ✅ | ✅ |
| View approval status on posts | ✅ | ✅ | ✅ | ✅ |
| See approval notification icon | ✅ | ✅ | ✅ | ✅ |

**Workflow management permission:** Super Admins and Admins can grant Collaborators and Approvers the ability to manage workflows via a toggle in team member permissions settings ("Allow Approval Workflow Management"). Without this permission, they can participate in approvals but cannot create, edit, delete, or set default workflows.

#### Landing Page (Workflow List)

- Shows all created workflows as **tiles/cards**
- Each tile shows: workflow name, number of levels, number of total members, default badge (if marked)
- **"Add New Workflow"** tile to create a new workflow (visible only to users with workflow management permission)
- Actions per tile (for users with permission): Edit, Duplicate, Delete, Mark as Default
- Only **one workflow** can be marked as Default at a time

#### Create / Edit Workflow

**Workflow properties:**
- **Name** (required)
- **Levels** (1 to 5 max) — each level has:
  - **Title** — editable, with placeholder text ("Level 01:", "Level 02:", etc.) — consistent "Level N" naming throughout the product
  - **Approval Rule** — "Everyone must approve" or "Anyone can approve" (one pre-selected by default)
  - **Team Members** — at least 1 member required per level

**Level management:**
- Drag & drop to reorder levels
- Duplicate a level
- Delete a level (confirmation if members assigned)
- "Add Level" button at the bottom (disabled if at max 5)

**Team member assignment:**
- Fixed right panel showing all eligible team members (anyone with planner access: admins, super admins, collaborators, approvers)
- Search bar to filter members
- Drag & drop members from right panel into levels
- Same member CAN appear in multiple levels
- Levels show member avatars; handle overflow for 10+ members (show "+N more" with tooltip or scrollable area within the level card)

**Workflow save validation:**
- Each level must have at least 1 member
- Approval rule must be selected per level (one pre-selected by default so always valid)
- Workflow can be saved as a **draft** if incomplete — UI clearly marks it as draft and it will NOT appear in the composer's workflow list until complete and published

**Workflow actions:**
- Duplicate entire workflow (creates "[Name] (Copy)")
- Delete workflow — see deletion rules below

#### Workflow Deletion Rules

If a workflow has **posts currently in approval**:
1. Show confirmation dialog listing the number of pending posts
2. Inform: "These posts will be converted to single-user approval (existing approvers at the current level will remain assigned). You will need to manage them manually."
3. On confirm: convert in-flight posts to single-user ad-hoc approval, delete workflow
4. On cancel: no action

#### Default Workflow

- One workflow can be marked as "Default" from settings
- When user opens "Send for Approval" in composer, the Approval Workflows tab auto-selects the default workflow
- User can still switch to a different workflow or to the Users tab
- Default does NOT auto-submit — user always clicks "Send" manually

---

### B. Composer — Send for Approval

#### Approval Sidebar Panel

When user clicks "Send for Approval", a **right sidebar** opens (replacing the current center modal) with two tabs:

**Tab 1: Users (existing ad-hoc mode, updated UI)**
- No hotlist/quick-select row
- Search bar at top
- List of all eligible team members (alphabetical)
- Checkbox to select/deselect users
- Approval rule toggle: "Everyone must approve" / "Anyone can approve"
- Notes/comment box at bottom
- "Send" button

**Tab 2: Approval Workflows**
- List of all workflows (cards)
- **Default workflow is pre-selected** when this tab opens (if one is set) — highlighted/selected state
- User can click a different workflow to switch selection
- **Collapsed view:** Workflow name, level count, member avatars grouped per level with connecting dashes (visual level indicator)
- **Expanded view:** Click chevron to expand — shows "Level 1: [Title]", "Level 2: [Title]" etc. vertically with member avatars and approval rule label ("Everyone" / "Anyone") per level. Consistent "Level N" naming (not "Step N").
- Handling 10+ members in a level: show first few avatars + "+N" badge; expanded view has scrollable member list per level
- Notes/comment box at bottom (always visible, not inside a level)
- "Send" button

**Editing a workflow from composer:**
- If user wants to modify the selected workflow, show a link/button "Edit workflow in Settings"
- Confirmation: "This will minimize the composer and redirect you to Settings. Continue?"
- On return, the composer restores and reflects the updated workflow

**What happens on "Send":**
- Post status remains as selected (Draft/Scheduled) — approval is an overlay, not a status replacement
- Post enters approval process
- Level 1 approvers are notified (in-app + email + mobile push)
- Post shows "Pending Approval" badge in planner

#### Same component reused in:
- Planner bulk actions (select multiple posts → Send for Approval)
- Automations (existing approval action)
- Mobile composer (both iOS & Android — add the Approval Workflows tab alongside existing Users tab)

---

### C. Approval Flow — How It Progresses

#### Happy Path (Workflow Mode)

1. Creator sends post for approval using Workflow X (3 levels)
2. **Level 1** approvers receive notifications
3. Based on Level 1 rule:
   - "Anyone can approve" → one approval advances to Level 2
   - "Everyone must approve" → all must approve before advancing to Level 2
4. **Level 2** approvers are notified that Level 1 is complete and their approval is pending
5. Same logic for Level 2
6. **Level 3** completes → post is **fully approved**
7. Post automatically transitions to its intended status (Scheduled → publishes at scheduled time, Draft → stays as approved draft)

#### "Anyone Can Approve" — Level Completion Notification

When a level uses "Anyone can approve" and User A approves (advancing the workflow):
- Users B and C at the same level are **notified that the level is complete and their review is no longer needed** at this level
- Their pending status is cleared — no action required from them
- They do NOT receive a notification to approve/reject

#### Rejection

1. Any approver at any level can **reject** the post (with optional comment)
2. Post moves to **"Rejected"** status
3. **All other pending approvers in the same level are notified** that the post was rejected (so they don't waste time reviewing)
4. Creator is notified of rejection
5. Post stays in "Rejected" status until creator takes action

#### Missed Review (Post Passes Scheduled Time While in Approval)

1. Post scheduled time passes while still in approval → post moves to **"Missed Review"** status
2. The approval process does **not** fail or auto-cancel — it remains in whatever approval state it was in
3. **Only the approvers at the current active level** are notified of Missed Review — past level approvers and future level approvers are NOT notified (they have no actions available anyway)
4. Only current level approvers retain the Approve / Reject options on a Missed Review post
5. When an approver approves/rejects a post in Missed Review, they are prompted: **"This post has missed its scheduled time. Would you like to reschedule it?"**
6. If yes → reschedule picker opens; if no → post stays in Missed Review but approval progresses
7. Check existing frontend planner implementation for exact reschedule flow — replicate same behavior for workflow approval posts

#### Post Deleted While in Approval

1. Creator deletes a post that is mid-approval (any state: Pending Approval, Missed Review)
2. **Only the current active level approvers are notified:** "[Post title] has been deleted and is no longer pending your approval"
3. Past level approvers and future level approvers are NOT notified
4. Approval process terminates — no further action required from anyone

#### Schedule Date Changed While Approval is In Progress

- Creator changes the scheduled date/time of a post that is still in approval (not fully approved yet)
- **No re-approval required** — approval is about content, not timing
- **No notification sent** to approvers — the content they are reviewing is unchanged

#### Approval History (Post-Publish)

- Approval history is **permanently linked to the post**
- After a post is published, the approval status panel remains accessible and shows the full history: each level, each approver's decision, timestamps
- This is read-only — no actions available once published

#### Approval Status Indicators

| Status | Icon | Color |
|---|---|---|
| Approved | Checkmark | Green |
| Pending | Clock/Circle | Orange |
| Rejected | X/Cross | Red |

**Tooltips on hover** show: user name, action taken, timestamp

---

### D. Approval Status Tracking

#### In Planner (All Views: Calendar, Feed, List)

- Posts in approval show a **status badge** on the card
- **Hover/click** the badge → popup showing:
  - Workflow name (or "Ad-hoc Approval")
  - Each level with status (Completed / In Progress / Not Started)
  - Per-user status within each level (approved/pending/rejected)
  - **"Re-notify"** button for pending approvers whose feedback is delayed

#### Approver Experience

When an approver receives a notification and clicks it:
- Taken directly to **Planner in feed view** showing only the post(s) assigned to them (filtered, no other posts visible)
- For multiple posts notified at once: planner shows all of them filtered in feed view
- The approval status panel is open/accessible from the post
- Approver can see:
  - The creator's note (sent at time of approval request — separate from comments thread)
  - Past levels' decisions (who approved/rejected at completed levels, with timestamps)
  - Other approvers' decisions at the current level (open review — not blind)
  - Comments from all participants
- Approver can: Approve, Reject (with optional comment), Add comment, Re-notify (if they are the creator), Revoke own approval (see rules below)

#### In Planner (All Views: Calendar, Feed, List)

- Posts in approval show a **status badge** on the card
- **Hover/click** the badge → popup showing:
  - Workflow name (or "Ad-hoc Approval")
  - Each level with status (Completed / In Progress / Not Started)
  - Per-user status within each level (approved/pending/rejected)
  - **"Re-notify"** button for pending approvers whose feedback is delayed

#### In Post Preview Modal (Detailed View)

- Full approval status panel
- Each level shown with all members and their statuses
- **Re-notify** option for individual pending approvers (label: "Re-notify")
- **Re-notify cooldown:** Maximum once every **2 hours** per approver. If re-notify was sent less than 2 hours ago, the button is disabled with tooltip showing when next re-notify is available.
- **"Revoke approval"** — an approver can revoke their own approval/rejection, but only if:
  - Their level is still the **currently active level** (not a completed past level)
  - The post has **not been fully approved** yet
  - For "Anyone can approve" rule: if their approval already closed the level and advanced the workflow, revoke is **not possible** (level is locked)
  - On revoke: their status resets to pending, creator is notified, other approvers on the same level are notified
- **No requester revoke:** The post creator cannot revoke a specific approver's decision — they edit the post to trigger the confirmation flow
- **No cancel from here:** To cancel the approval process, the user edits the post (see Edit Scenarios below)
- Applies to **both single-level and workflow approval** — same logic, same component

---

### E. Edit Scenarios & Confirmation Dialogs

**Core principle:** Editing a post never automatically changes approval status. The creator always gets a choice.

#### Scenario 1: Edit Post in Active Single-User Approval

User edits content → clicks Save Draft / Save Schedule → **Confirmation dialog:**

> **"This post is pending approval"**
> - [User A] — Approved
> - [User B] — Pending
>
> You've updated the content. Would you like to:
> - **Re-notify approvers** — Approved/rejected statuses reset to pending, approvers are notified of the update
> - **Keep current approval status** — Save changes without affecting the approval process
> - **Remove approval** — Post exits approval process, reverts to draft/scheduled
> - **Cancel**

#### Scenario 2: Edit Post in Active Workflow Approval

User edits content → clicks Save Draft / Save Schedule → **Confirmation dialog:**

> **"This post is in approval workflow: [Workflow Name]"**
> - Level 1 (Completed) — 2/2 approved
> - Level 2 (In Progress) — 1/3 approved, 1 pending
> - Level 3 (Not Started)
>
> You've updated the content. Would you like to:
> - **Restart from Level 1** — All approvals reset, workflow starts over
> - **Re-notify current level** — Only Level 2 resets, Level 1 stays approved
> - **Keep current approval status** — Save changes without affecting the workflow
> - **Remove approval** — Post exits approval process, reverts to draft/scheduled
> - **Cancel**

#### Scenario 3: Edit Rejected Post (Single-User) — Bug Fix

Currently broken: saving silently moves post to draft/scheduled.

User edits content → clicks Save Draft / Save Schedule → **Confirmation dialog:**

> **"This post was rejected"**
> - [User A] — Rejected
> - [User B] — Approved
>
> Would you like to:
> - **Resend for approval** — Statuses reset to pending, approvers re-notified
> - **Remove approval** — Post moves to draft/scheduled without approval
> - **Cancel**

#### Scenario 4: Edit Rejected Post (Workflow)

User edits content → clicks Save Draft / Save Schedule → **Confirmation dialog:**

> **"This post was rejected at Level 2 of [Workflow Name]"**
> - Level 1 (Completed) — 2/2 approved
> - Level 2 (Rejected) — Rejected by [User C]
> - Level 3 (Not Started)
>
> Would you like to:
> - **Restart from Level 1** — All approvals reset, workflow starts over
> - **Resume from Level 2** — Only Level 2 resets, re-notifies Level 2 approvers
> - **Remove approval** — Post moves to draft/scheduled without approval
> - **Cancel**

#### Scenario 5: Edit Post in "Missed Review" State

Behaves identically to active approval scenarios (Scenario 1 for single-user, Scenario 2 for workflow). The same confirmation dialog is shown — the missed schedule time does not change the edit flow.

---

### F. Workflow Modification While Posts Are In-Flight

When a workflow is edited from Settings while posts are using it:

| Change | Level Not Started | Level In Progress | Level Completed |
|---|---|---|---|
| **New member added** | Added, will be notified when level starts | Added, notified immediately | NOT added retroactively. Applies to future posts only. |
| **Member removed (no post assigned)** | Removed cleanly | Removed cleanly | Removed cleanly |
| **Member removed (post pending)** | Their assignment cancelled, notified | Their assignment cancelled, notified | NOT removed. Greyed out with tooltip: "Completed approval before removal" |
| **Level added** | Applies to all in-flight posts that haven't reached this point | Applies normally | Applies normally |
| **Level removed** | Applies to all in-flight posts | Posts at this level **auto-advance to the next level** (next level's approvers notified); if this was the **last level**, posts are **auto-completed (fully approved)** and creator notified | Does not affect completed — applies to future posts |
| **Approval rule changed "Everyone"→"Anyone"** | Applies to new posts | If 1+ approvals already exist → level **auto-completes and advances** | No effect |
| **Approval rule changed "Anyone"→"Everyone"** | Applies to new posts | Requires all members to approve going forward | Stays completed — no retroactive invalidation |
| **Workflow name or level title renamed** | Reflects live on all in-flight posts immediately | Reflects live | Reflects live |

**All changes apply to new posts immediately.** For in-flight posts, the rules above apply.

#### Save Confirmation When In-Flight Posts Exist

Whenever the user clicks "Save Workflow" or "Save Changes" on a workflow that has in-flight posts, a **confirmation dialog** is shown before saving — not just the warning banner. The banner alone is not sufficient.

**Confirmation dialog:**
- Title: `"Save changes to "[Workflow Name]"?"`
- Body: `"[N] post(s) are currently in approval using this workflow. Saving these changes will immediately affect their approval process."`
- If any level was **deleted**: append `"[N] post(s) at removed levels will automatically advance to the next level. If a removed level was the last, those posts will be fully approved."`
- Primary button: `"Save & Apply Changes"`
- Secondary button: `"Cancel"`

On "Save & Apply Changes": changes are saved and all mid-flight rules are applied immediately.
On "Cancel": no changes saved; user returns to the workflow editor.

The warning banner (`⚠ [N] post(s) are currently using this workflow. Some changes may affect those posts.`) remains visible while editing, but the final confirmation dialog is the gate before saving.

#### Team Member Role / Permission Change

- A team member's role changes (e.g., Admin → Collaborator) or their "Allow Approval Workflow Management" permission is revoked
- **In-flight approvals:** Unaffected — role change does not remove them from posts they are already assigned to
- **Workflow management access:** Immediately revoked — they can no longer create, edit, delete, or set default workflows
- **Existing workflows they created:** Remain active and visible but they cannot edit them until permission is restored

#### Approver Removed from Workspace

When an admin removes a team member who is currently assigned as an approver on in-flight posts:
1. **Confirmation dialog on removal:** "This member is assigned as an approver on [N] pending post(s) and/or is part of [N] approval workflow(s). Removing them will affect these approvals."
2. On confirm:
   - Their assignment is removed from all in-flight approvals
   - For multi-level workflows: the level **auto-continues** without them (does not block the approval process)
   - If the removed approver was the only member of a level using "Everyone must approve" → that level is considered passed (no approvers left to block it)
   - **Creator is notified** for each affected post: "[Member name] was removed from the workspace and has been removed from this post's approval workflow."
3. Their membership in saved workflow definitions is also removed

---

### G. Notifications

#### Dedicated Approval Notification Panel

- **New icon** in the top bar — right side, positioned to the **left of the existing notifications bell icon** (i.e., between the bell and any other right-side icons)
- Opens a panel showing **approval-specific activity feed**
- Completely separate from general notifications

**Notification triggers:**

| Event | Who is Notified | Channels |
|---|---|---|
| Post submitted for approval | Level 1 approvers | In-app, email, mobile push |
| Level completed, next level pending | Next level approvers | In-app, email, mobile push |
| Post approved at a level | Creator | In-app, email |
| Post fully approved (all levels) | Creator | In-app, email, mobile push |
| Post rejected | Creator + all other pending approvers in same level | In-app, email, mobile push |
| Comment added on post in approval | Creator + all assigned approvers | In-app, email |
| Post content updated during approval | All assigned approvers (depending on creator's choice) | In-app, email |
| Approver re-notified | The specific approver | In-app, email, mobile push |
| Approver revokes their approval | Creator + other approvers on same active level | In-app, email |

**Additional trigger — level removed from workflow (auto-advance):**

| Event | Who is Notified | Channels |
|---|---|---|
| Level removed, post auto-advances to next level | Creator (per affected post) + next level's approvers | In-app, email |
| Level removed and it was the last — post auto-approved | Creator (per affected post) | In-app, email, mobile push |

**Quick category filters in the panel:**
- Pending Approvals (assigned to me)
- Rejected (my posts or posts I'm assigned to)
- Updated (posts modified during approval)
- New Comments

Clicking a category → navigates to Planner with appropriate filters applied.

#### Cross-Workspace Notification Badges

- Workspace switcher dropdown shows notification count per workspace
- Workspace page tiles show individual notification counts
- Covers all approval events happening in other workspaces

---

### H. Planner Filter Updates

**Existing approval-related filters already exist** in planner with custom views:
- Pending Approval (covers all in-progress approval, regardless of how many levels are done)
- Custom views for "My pending approvals" and "My requested approvals"

**Rename:** "Under Review" → "Pending Approval" (if not already done)

**Post approval statuses (no "Partially Approved" — not needed):**
- **Pending Approval** — in approval at any level (Level 1, 2, 3, etc.)
- **Missed Review** — scheduled time passed, still in approval
- **Rejected** — rejected at any level
- Approved posts revert to their underlying status (Scheduled / Draft)

**New data in existing filters:**
- Approval workflow name and current level information available in filter results

---

### I. Mobile App

#### Existing (already implemented):
- Approve / Reject / Comment on posts
- Send for approval (single-user mode)
- Push notifications for approval events

#### New additions:
- **"Pending Approvals"** item in bottom navigation
  - Opens a view filtered to posts needing the user's approval
  - Can switch between: "Needs My Approval" / "My Requested Approvals"
  - Can filter by team member
- **Approval Workflows tab** in mobile composer's "Send for Approval" flow
  - Same two tabs as web: Users / Approval Workflows
  - Default workflow pre-selected
  - Collapsed/expanded view adapted for mobile
- **Approval status in post detail** — shows level breakdown with member statuses

---

### J. Bulk Actions (Approval)

**Bulk Send for Approval (Creator):**
- User selects multiple posts in Planner → clicks "Send for Approval" from bulk action menu
- **Pre-flight override check:** Before the sidebar opens, the system checks if any selected posts already have an active approval process (internal or external). If any do, a confirmation dialog is shown — see Section K for override rules.
- Same approval sidebar opens (Users / Approval Workflows tabs)
- All selected posts are sent through the chosen approval path
- Each post gets independent approval tracking
- Same component used in Automations

**Bulk Approve / Reject (Approver):**
- Approver can select multiple posts in their Pending Approvals view and approve or reject them all at once
- Existing bulk action mechanism in planner — check current implementation and extend for workflow approval
- Bulk reject prompts for a single comment that applies to all rejected posts

**Rescheduling a Fully Approved Post:**
- A fully approved post can have its scheduled time changed by the creator
- **No re-approval required** — time change does not affect content
- Approvers are not notified of a schedule time change

---

### K. Single Post Send for Approval from Planner

Posts sitting in the Planner (Draft or Scheduled) can be sent for approval directly — without opening the Composer.

**Entry points:**
- Post card 3-dot (···) menu → "Send for Approval"
- Post detail / preview modal → "Send for Approval" button (shown when post is NOT currently in any approval)

**Behaviour:**
- Opens the same `SendForApprovalSidebar` component used in the Composer (Users tab + Approval Workflows tab)
- If the post already has an active approval process (internal or external), the override warning is shown — see Section L
- Notes field, Send button, Cancel button — all identical to Composer flow
- On Send: post enters approval; planner badge updates in real-time

---

### L. Approval Override Rules (All Entry Points)

**Core principle:** A post can only have one active approval at a time — internal workflow, internal ad-hoc, or external share-link approval. A new approval always replaces the previous one. The user is always warned before this happens.

**This applies to every entry point that initiates an approval:**
- Composer → Send for Approval
- Planner post card / detail → Send for Approval (single post)
- Planner bulk action → Send for Approval (multiple posts)
- Planner share link modal → Step 2: add emails and send for external approval

#### Single Post Override Warning (Composer or Planner)

When the Send for Approval sidebar opens for a post that already has an active approval process:

- An `Alert` (warning variant) is shown at the top of the sidebar:
  - For internal ad-hoc: `"This post is already in an approval process. Starting a new one will cancel it — currently assigned approvers will be notified."`
  - For internal workflow: `"This post is in [Workflow Name] (Level [N]). Starting a new approval will cancel it — Level [N] approvers will be notified."`
  - For external approval: `"This post has an active external approval via share link. Starting an internal approval will cancel it."`
- User can still proceed: fill in the sidebar and click Send — the warning is informational, not a blocker
- On Send: old approval is cancelled (notifications dispatched), new approval begins

#### Bulk Send Override Confirmation (Planner)

When the user selects multiple posts and clicks "Send for Approval" — before the sidebar opens:

- System checks each selected post for an existing approval
- If **any** selected posts have an active approval:
  - Confirmation `Dialog`:
    - Title: `"Some posts are already in approval"`
    - Body: `"[N] of [M] selected posts already have an active approval process. Starting a new one will cancel their current approvals — assigned approvers will be notified."`
    - Primary button: `"Continue"`
    - Secondary button: `"Cancel"`
  - On "Continue": sidebar opens normally; all selected posts (including those with existing approvals) will be processed
  - On "Cancel": no action; selection remains

#### External Share Link Override Warning (SharePlanModal)

When a user is in the Share via Link modal (Step 2) and toggles on "Send for Approval" (adds emails to request approval from an external client):

- If the selected post has an active **internal** approval process:
  - `Alert` (warning variant) shown inside the share modal: `"This post is currently in an internal approval process. Sending it for external approval will cancel it — currently assigned internal approvers will be notified."`
- User can proceed or dismiss
- On confirm / send: internal approval state is fully cleared (all in-flight approver records cleaned up), external approval replaces it, cancellation notifications sent to previously assigned internal approvers

#### Cancellation Notifications on Override

When any approval is overridden, the system sends cancellation notifications to the approvers who were previously assigned at the current active level (same scope as post-deletion notification — current active level only):

| Notification | Copy |
|---|---|
| In-app title | `"[Post title] has been sent for a new approval"` |
| In-app description | `"The approval process you were part of for [post title] in [workspace name] has been replaced. No further action is required from you."` |
| Email subject | `"No action needed — approval replaced for [post title]"` |
| Email body | `"[Creator name] has started a new approval process for [post title] in [workspace name]. The previous approval process has been cancelled. No further action is required from you."` |
| Channels | In-app, email |
| Who is notified | Current active level approvers only (same rule as post-deletion) |

**Backend fix required:** The API currently does not clean up internal approval state when external approval overrides it (the reverse direction works correctly). This must be fixed — when external approval replaces an internal approval, all previous internal approval state must be fully cleared before the new external approval is written. Backend team to handle in their technical spec.

---

## 3. Scope Summary

### In Scope (v1)

| Module | Items |
|---|---|
| **Settings** | Workflow CRUD, role permissions, workflow management permission toggle, draft workflows, default workflow, duplicate level/workflow, drag & drop levels and members, member search, save validation |
| **Composer** | Sidebar with 2 tabs (Users / Workflows), notes (separate from comments), default pre-selection, edit-to-settings redirect, override warning when post already in approval |
| **Approval Flow** | Multi-level progression, everyone/anyone rules, level completion notification ("no longer needed"), rejection with notification to other assignees, missed review behavior, approval history post-publish |
| **Edit Handling** | 4 confirmation dialog scenarios (active single/workflow, rejected single/workflow), remove approval option, bug fix for rejected post |
| **Workflow Mid-Flight** | Add/remove member rules, deletion with conversion to single-user, approver removed from workspace handling |
| **Status Tracking** | Badges on all planner views, hover popup, detailed post preview panel, re-notify (2hr cooldown), revoke approval (active level only), approver experience (open review, past levels visible) |
| **Notifications** | Dedicated approval icon (left of bell), all event triggers, in-app + email + mobile push, cross-workspace badges, override cancellation notifications |
| **Planner** | Filter updates, rename to "Pending Approval", clear status definitions (no "Partially Approved"), rescheduling approved posts, single-post "Send for Approval" from post card/detail |
| **Bulk Actions** | Creator: multi-post send for approval with override confirmation; Approver: bulk approve/reject from pending view |
| **External Share Link** | Override warning when post has active internal approval; backend cleanup fix for internal→external override |
| **Mobile** | Pending Approvals nav item, workflow tab in composer, approval status in post detail |

### Deferred (Future)

| Item | Reason |
|---|---|
| Auto-reminders for delayed approvers | Complexity — v2 enhancement |
| Image annotations from approver | Complexity — v2 enhancement |
| Activity vs Notifications split (platform-wide) | Broader initiative beyond approval workflow |

### Design Decisions (Finalized)

1. **Approval sidebar:** Right sidebar. Final layout to be informed by Figma designs and Gain app reference screenshots.
2. **10+ members in a level:** Use "+N more" badge in collapsed view (consistent with avatar stacking patterns used elsewhere in CS). Expanded view shows a scrollable list of all members within the level card.
3. **Max levels:** Fixed at 5. Not user-configurable. This keeps the UI manageable and covers virtually all real-world approval chains. Can be revisited in v2 if customer feedback demands more.
