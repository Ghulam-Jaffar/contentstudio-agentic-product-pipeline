# Epic + stories — Approval workflows on the public API

**Priority: P1.** 3 stories. Nothing is pushed to any tracker.

---

## Epic: Approval workflows on the public API

Multi-level approval workflows are one of ContentStudio's stronger differentiators, particularly for agencies. A workflow defines who approves what, in what order, and what happens when a level approves or rejects. The product has the full feature: create a workflow, edit it, duplicate it, set a default, approve, reject, revoke an approval, re-notify an approver who has not acted, read the approval history, and see the notifications the workflow generated.

The public API has one endpoint: list the workflows. It also has post-level approve and reject, which is genuinely useful, and it accepts a workflow id when creating a post. Everything else is browser-only.

That leaves an odd shape. An integration can send a post into an existing workflow and can approve or reject that post, but cannot create the workflow, cannot change it, cannot revoke an approval that was given in error, cannot chase an approver who has gone quiet, and cannot read the approval history to find out what happened. For an agency automating client sign-off, the history and the revoke are not nice-to-haves; they are the audit trail.

This epic gives the API the whole workflow feature: definitions, actions, and history.

### Out of scope

- Changing how approval workflows behave, including the cascade and notification rules.
- The notification delivery itself. Approval notifications are covered by the separate notifications architecture epic, and this epic must not build a second notification path.

### Stories

1. `[BE] Add approval workflow definition management to the public API`
2. `[BE] Add approval actions and approval history to the public API`
3. `[BE] Expose approval workflows on every developer surface`

---

## [BE] Add approval workflow definition management to the public API

### Description

As a developer provisioning a workspace through the API, I want to create and manage approval workflows, so that I can set up a customer's sign-off process programmatically rather than handing them a browser task in the middle of an automated onboarding.

### Workflow

1. The developer lists a workspace's approval workflows.
2. The developer reads a single workflow, including its levels and the approvers at each level.
3. The developer creates a workflow, defining its levels, the approvers at each level, and the rule for how each level is satisfied.
4. The developer updates a workflow, or deletes it.
5. The developer duplicates an existing workflow as a starting point for another.
6. The developer sets a workflow as the workspace default, or removes the default.
7. The developer creates a post against a workflow, as they can today.

### Acceptance criteria

- [ ] A caller can list workflows, read a single workflow by id, and create, update and delete a workflow.
- [ ] Reading a single workflow returns its levels, the approvers at each level, and each level's satisfaction rule, so a caller can understand the workflow without further calls.
- [ ] A caller can duplicate a workflow, and the copy is independent of the original.
- [ ] A caller can set and remove the workspace default workflow.
- [ ] Validation matches the product: the same limits on levels and approvers, the same rules on who may be an approver, with clear errors naming the field that failed.
- [ ] Authorization uses the existing permission model. A caller who cannot manage workflows in the product cannot manage them through the API.
- [ ] Plan gating is respected. A workspace on a plan without multi-level approvals receives a clear refusal.
- [ ] Editing or deleting a workflow that has posts currently in flight behaves exactly as the product does, and the response makes clear what happened to those posts.
- [ ] The existing list endpoint keeps its response shape, and creating a post against a workflow keeps working as it does today, including the existing workflow action options.
- [ ] Every endpoint appears in the OpenAPI document with request and response examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Workflows are created and modified by callers as they would be in the product. Behavior for posts already in flight is preserved rather than changed.

### Impact on other products

Approval workflows are visible and actionable in the web app and the mobile app. A workflow created through the API must be indistinguishable from one created in the browser in both, and the story must verify that.

### Dependencies

- None for the endpoints.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (validation and error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add approval actions and approval history to the public API

### Description

As a developer automating a client sign-off process, I want the full set of approval actions and the approval history through the API, so that I can correct a mistaken approval, chase an approver who has not responded, and produce an audit trail without anyone opening the browser.

### Workflow

1. The developer approves or rejects a post, as they can today.
2. The developer revokes an approval that was given in error, and the post returns to the state the product would return it to.
3. The developer re-notifies an approver who has not yet acted.
4. The developer reads a post's approval history: who acted, what they did, when, and any note they left.
5. The developer reads the state of an in-flight approval, including which level it is at and who it is waiting on.

### Acceptance criteria

- [ ] A caller can approve and reject a post, unchanged from today.
- [ ] A caller can revoke an approval, and the post returns to exactly the state the product returns it to.
- [ ] A caller can re-notify a named approver who has not yet acted, and the notification sent is the same one the product sends, through the existing notification path rather than a new one.
- [ ] A caller can read a post's approval history: each action, the person who took it, the time, and any note.
- [ ] A caller can read the current state of an in-flight approval, including the level it is at and the approvers it is waiting on.
- [ ] Authorization matches the product exactly. A caller who is not an approver on a post cannot approve it, and a caller who cannot revoke in the product cannot revoke through the API.
- [ ] Attempting an action that is not valid for the post's current state returns a clear error explaining why, not a generic failure.
- [ ] Re-notifying is rate-limited or otherwise protected so it cannot be used to spam an approver, and the limit is documented.
- [ ] Actions taken through the API are attributed in the history as API actions, so the audit trail distinguishes them from actions taken in the browser.
- [ ] Every endpoint appears in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None beyond the approval history entries that actions naturally create. Existing history is not migrated or altered.

### Impact on other products

Approval actions taken through the API must appear correctly in the web and mobile apps, and must trigger the same notifications. The story must verify both rather than assume.

### Dependencies

- Depends on `[BE] Add approval workflow definition management to the public API` for the workflow context its actions operate in.
- Uses the existing notification path. Coordinates with the notifications architecture epic so re-notify does not become a second notification implementation.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages and any notification copy must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (approval notifications and links carry white-label branding)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose approval workflows on every developer surface

### Description

As a developer or an agency using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want approval workflows available there too, so that sign-off can be part of the automation I have already built rather than a manual step in the middle of it.

### Workflow

1. The developer manages workflows and takes approval actions from the CLI.
2. An AI assistant, through the MCP server, can read what is awaiting approval, and approve or reject on the user's behalf.
3. The developer builds a no-code automation that reacts to a post entering approval, or approves a post from another system.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Workflow management and approval actions are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools, including reading what is awaiting the user's approval.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Approving or rejecting through an AI assistant requires the same explicit confirmation the MCP epic defines for actions that change state, so an assistant cannot approve a client's content on a loose instruction.
- [ ] Zapier, Make and n8n expose at minimum a trigger for a post entering or leaving approval, and an action to approve or reject.
- [ ] Where a surface deliberately omits a capability, that omission is recorded with its reason.
- [ ] The public documentation covers the capability on every surface, with a worked example of an automated sign-off flow.
- [ ] The parity check passes for approval workflows across every surface.
- [ ] Every surface enforces the same authorization as the REST endpoints. No surface lets a non-approver approve.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on both endpoint stories in this epic.
- Depends on the developer surface parity contract epic.
- Depends on the MCP epic for how an assistant confirms a state-changing action.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
