# Stories: Shared Link Feedback Controls (Comments + Approve/Reject)

---

## Story 1: [BE] Add share link feedback permission flags and enforce external action restrictions

### Description:

Shared planner links currently allow external users to comment and/or approve/reject based on status and approval flow, but there is no explicit per-link permission model to disable these actions. This story adds backend support for two explicit permissions and enforces them on action endpoints.

**New/updated share-link permission fields:**

- `allow_external_comments` (boolean)
- `allow_external_approval_actions` (boolean)

**Backend scope:**

1. Persist both fields on share links during create/update flows
2. Return both fields from share-link read APIs used by the frontend (`shareLink/get`, `shareLink/fetch` as applicable)
3. Enforce both permissions on external action endpoints:
- `shareLink/comment` rejects when comments are disabled
- `shareLink/action` rejects when approve/reject is disabled
4. Preserve backward compatibility for existing links by defaulting missing values to `true`

**Expected behavior details:**

- Existing links (created before this feature) should continue current behavior unless explicitly updated
- If permission fields are omitted on update, existing stored values should remain unchanged (no accidental resets)
- Validation should prevent invalid state combinations where approval requests are initiated while `allow_external_approval_actions` is false
- Error responses for blocked actions should be explicit and user-safe so frontend can show clear toasts

---

### Workflow:

1. Workspace owner creates or edits a shareable planner link and saves feedback permissions
2. A client opens the link and reviews posts
3. If comments are disabled for that link, the client cannot submit comments
4. If approve/reject is disabled for that link, the client cannot approve or reject posts
5. If both are disabled, the client can still view shared content but cannot leave feedback

---

### Acceptance criteria:

- [ ] Share link create/update APIs accept and persist `allow_external_comments` and `allow_external_approval_actions`
- [ ] Share link read APIs used by web (`shareLink/get` and `shareLink/fetch` response payloads where link metadata is returned) include both permission fields
- [ ] Existing links without these fields are treated as `allow_external_comments=true` and `allow_external_approval_actions=true`
- [ ] `shareLink/comment` rejects requests when `allow_external_comments=false`
- [ ] `shareLink/action` rejects approve/reject requests when `allow_external_approval_actions=false`
- [ ] Rejection responses include clear messages suitable for frontend to display to end users
- [ ] If update payload omits permission fields, existing stored values are preserved
- [ ] Attempting to initiate approval flow while `allow_external_approval_actions=false` is rejected with a validation/business-rule error
- [ ] `ak`-based approver flow remains functional when `allow_external_approval_actions=true`
- [ ] No regression in password-protected links, plan fetching, and existing approval/comment history rendering

---

### Mock-ups:

N/A - backend-only story.

---

### Impact on existing data:

- Share-link documents gain two boolean fields: `allow_external_comments` and `allow_external_approval_actions`
- Existing records remain valid and should default to permissive behavior (`true/true`) when fields are absent

---

### Impact on other products:

- Mobile apps: Not affected (planner share-link review flow is web-only)
- Chrome extension: Not affected
- Other planner internal authenticated views: Not directly affected

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support - N/A, backend-only story
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 2: [FE] Add share link feedback permission toggles and gate external actions in shared planner UI

### Description:

This story introduces explicit UI controls for shared-link feedback permissions and applies those permissions across all external action entry points in planner shared-link views.

**Product decisions finalized for this story:**

1. Permission toggles are in **Step 1** (the form with title/password/sharing options)
2. `Send for approval` remains in **Step 2**, always visible, but disabled when approvals are not allowed
3. Disabled tooltip copy should reference **"previous step"** (not "step 1")
4. Back button remains visible in disabled states so users can go back, change settings, click **Update**, and return to Step 2
5. Email copy remains simple and static (no dynamic variant)
6. Both permission toggles are **enabled by default** for new links

**Frontend scope:**

1. In `SharePlanModal.vue` Step 1, add two toggles:
- `Allow comments`
- `Allow approve/reject`
2. In Step 2, keep `Send for approval` visible but disabled when needed, with tooltip-priority logic:
- First priority: approvals disabled by permission from previous step
- Second priority: existing future-content restriction
3. Reuse existing back/update flow:
- Back button returns to Step 1
- `Update` applies settings to current link and returns user to Step 2
4. Gate external action UI across shared-link surfaces:
- Bulk actions in `SharePlans.vue`
- Row action buttons in `DataRow.vue` and `DataRowCardMobile.vue`
- Preview approve/reject footer and external comment panel in `PlannerPostPreview_v2.vue` + `CommentsAndNotes.vue`
- Defensive gating in `ExternalActionsModal.vue` (for stale/open edge cases)
5. Ensure edit/update payload preservation path includes the new flags (`ManageLinksModal.vue` helper payload)
6. Add i18n strings in planner locale files and wire English copy

---

### Workflow:

1. User opens Planner and starts creating a shareable link
2. In the first step, user enters title/password and chooses feedback permissions using:
- `Allow comments`
- `Allow approve/reject`
   - Default for new links: both toggles are ON
3. User generates/updates the link and goes to the next step
4. In the next step, user sees `Send for approval`
- If approvals are disabled from the previous step, it is visible but disabled with a tooltip explaining why
- User can click `Back`, change settings, click `Update`, and continue in the same flow
5. User adds email addresses and shares link via email
6. Client opens the shared link:
- If comments are allowed, client can comment
- If approve/reject is allowed, client can approve/reject where eligible
- If either is disabled, those actions are hidden/disabled in all shared-link UI entry points

---

### Acceptance criteria:

- [ ] Step 1 in `SharePlanModal` includes `Allow comments` and `Allow approve/reject` toggles
- [ ] For newly created links, both toggles are ON by default when Step 1 first opens
- [ ] Step 1 toggles are loaded correctly in edit mode from existing link data
- [ ] Step 2 `Send for approval` remains visible
- [ ] If `Allow approve/reject=false`, `Send for approval` is disabled and tooltip reads: `Enable "Allow approve/reject" in the previous step to use this option.`
- [ ] When both disable reasons apply, permission-disabled tooltip has priority over future-content tooltip
- [ ] Back button is shown in this disabled state and returns user to Step 1
- [ ] After changing settings and clicking `Update`, the existing link is updated and user returns to Step 2 (no new link creation)
- [ ] Email field label is `Share link via email (optional)`
- [ ] Email helper text is `Add email addresses to share this link via email.`
- [ ] In shared-link list view, Approve/Reject bulk actions are hidden or disabled when approvals are not allowed
- [ ] In shared-link list/mobile rows, Approve/Reject buttons are hidden or disabled when approvals are not allowed
- [ ] In shared-link preview, approve/reject footer buttons are hidden or disabled when approvals are not allowed
- [ ] In shared-link preview/comments, external comment input and submit are hidden or disabled when comments are not allowed
- [ ] If both permissions are false, shared link becomes view-only (no comment/approve/reject actions exposed)
- [ ] Defensive client-side checks in action modals/components prevent submission attempts when action is disallowed
- [ ] No hardcoded primary colors are introduced; use theme-aware classes/design system patterns
- [ ] Existing password-protection and share-link generation/update flows remain intact

---

### Mock-ups:

N/A - extends existing share-link modal and shared-link planner action surfaces using existing component patterns.

---

### UI Copy

**Step 1 - Feedback permissions section**

- Section title: `Client feedback permissions`
- Toggle 1 label: `Allow comments`
- Toggle 1 tooltip: `Clients can leave comments on shared posts when this is on.`
- Toggle 2 label: `Allow approve/reject`
- Toggle 2 tooltip: `Clients can approve or reject shared posts when this is on.`
- Default state (new link): both toggles ON

**Step 2 - Approval control behavior**

- Control label: `Send for approval`
- Disabled tooltip (permission): `Enable "Allow approve/reject" in the previous step to use this option.`
- Existing future-content tooltip remains for future-content restriction and is used only when permission rule does not apply
- Back button label: `Back`
- Update button label in edit flow: `Update`

**Step 2 - Email copy (static/simple)**

- Field label: `Share link via email (optional)`
- Placeholder: `Enter email addresses`
- Helper text: `Add email addresses to share this link via email.`

**Shared link view-only experience**

- When actions are disabled, action buttons are not shown (or shown disabled where required by layout)
- No new jargon-heavy text should appear in shared page controls

---

### Impact on existing data:

- Consumes and sends new permission fields introduced by backend story
- No direct schema migration in frontend; relies on backend defaults for legacy links

---

### Impact on other products:

- Mobile apps: Not affected (shared planner link interaction is web flow)
- Chrome extension: Not affected
- White-label domains: impacted only via UI rendering; must remain theme-token compliant

---

### Dependencies:

Depends on: **[BE] Add share link feedback permission flags and enforce external action restrictions**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (share modal + shared link action controls should remain usable on small screens)
- [ ] Multilingual support (new strings added to translation files with fallback handling)
- [ ] UI theming support (default + white-label, design library/components and theme tokens)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
