# Epic + stories — Workspace settings and team account access on the public API

**Priority: P2.** 3 stories. Nothing is pushed to any tracker.

---

## Epic: Workspace settings and team account access on the public API

Workspace management on the public API shipped recently and is genuinely good: create, update and delete a workspace, and full team member CRUD with per-action permission gating. What is missing is the layer just underneath, which is what makes a workspace actually usable.

Two things in particular. **Team members can be invited but not scoped to accounts.** The product lets an owner say "this member may only use these three social accounts", and that control has no API surface, so an integration can add a member to a workspace but not limit what they can reach. For an agency onboarding a client's staff programmatically, that is the difference between a complete provisioning flow and one that ends with a manual security step.

**And the workspace's own settings are not settable.** Timezone in particular: a workspace created through the API takes a default timezone, and every scheduled post is interpreted against it, so getting it wrong quietly shifts a customer's whole publishing schedule. Pause and resume posting is the other one worth having, because it is the operational lever a customer wants when something is wrong and it currently requires a browser.

This epic completes workspace provisioning so an integration can set a workspace up end to end.

### Out of scope

- White-label configuration, custom domains and SSO. Deliberately excluded pending a decision on whether they should ever be API-manageable.
- Billing and subscription changes. Not in scope for this epic.
- Onboarding flow state, which is a UI concern.

### Stories

1. `[BE] Add team member social account access to the public API`
2. `[BE] Add workspace settings management to the public API`
3. `[BE] Expose workspace settings and team access on every developer surface`

---

## [BE] Add team member social account access to the public API

### Description

As a developer provisioning a workspace through the API, I want to control which social accounts each team member may use, so that I can finish onboarding a client's team without leaving a manual permissions step for a human.

### Workflow

1. The developer lists a team member's social account access.
2. The developer grants a member access to specific accounts.
3. The developer updates a member's access, adding and removing accounts.
4. The developer confirms the member can only see and post to the accounts they were granted.

### Acceptance criteria

- [ ] A caller can read which social accounts a given team member has access to.
- [ ] A caller can grant a member access to specific social accounts.
- [ ] A caller can update a member's access, adding and removing accounts in one call.
- [ ] A member's access set through the API is enforced everywhere: the member cannot list, select or post to an account they were not granted, verified through the API and in the product.
- [ ] Authorization matches the product: only a caller permitted to update members can change account access.
- [ ] Removing a member's access to an account behaves exactly as the product does for posts that member already scheduled to it, and the response or documentation states what happens.
- [ ] Granting access to an account the caller themselves cannot access is refused.
- [ ] Every endpoint appears in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Access records are created and removed as they would be in the product.

### Impact on other products

Account access is enforced in the web app, the mobile app and the Chrome extension. The story must verify access set through the API is honoured in all of them, since this is a permission boundary and a gap in any client is a real exposure.

### Dependencies

- Overlaps the service authorization epic, which is assessing per-account access enforcement across the browser-reachable services. This story must not create a second enforcement model, and should reuse whatever that epic concludes.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add workspace settings management to the public API

### Description

As a developer creating and managing workspaces through the API, I want to set a workspace's timezone and its operational settings, so that scheduled posts publish when my customer expects and I can pause publishing without opening the browser.

### Workflow

1. The developer reads a workspace's current settings.
2. The developer sets the workspace timezone, either at creation or afterwards.
3. The developer pauses posting for the workspace, and reads how many posts are affected.
4. The developer resumes posting.
5. The developer sets the workspace logo and the remaining workspace-level settings the product exposes.

### Acceptance criteria

- [ ] A caller can read a workspace's settings, including its timezone, its posting state, and its logo.
- [ ] A caller can set a workspace's timezone, both when creating a workspace and afterwards.
- [ ] Changing the timezone behaves exactly as the product does for already-scheduled posts, and the response or documentation states plainly what happens to them. A caller must never be able to silently move a customer's schedule without knowing.
- [ ] A caller can pause posting for a workspace, and read how many scheduled posts the pause affects.
- [ ] A caller can resume posting, and the affected posts return to the state the product returns them to.
- [ ] A caller can set the workspace logo.
- [ ] The remaining workspace-level settings the product exposes are readable and settable, and any deliberately excluded setting is listed with its reason.
- [ ] Validation matches the product, including the accepted timezone values, with clear errors.
- [ ] Authorization uses the existing workspace permission model.
- [ ] The existing workspace create and update endpoints keep their current behavior and response shapes for callers that do not use the new fields.
- [ ] Every endpoint appears in the OpenAPI document with examples, and the timezone documentation is explicit about the effect on scheduled posts.

### Mock-ups

None. Backend-only story.

### Impact on existing data

Changing a workspace timezone affects how already-scheduled posts are interpreted. That behavior is inherited from the product unchanged, but because it is now reachable programmatically the story must document it explicitly and make it visible in the response. Pausing posting changes the state of scheduled posts exactly as the product does.

### Impact on other products

Workspace timezone drives scheduling in the web app, the mobile app and the Chrome extension. A pause set through the API must be visible and honoured in all of them. The story must verify this.

### Dependencies

- Extends the workspace management endpoints already shipped rather than replacing them.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (the workspace logo is a white-label-visible asset)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose workspace settings and team access on every developer surface

### Description

As a developer or agency provisioning workspaces through the CLI, an AI assistant, or a no-code tool, I want workspace settings and team account access available there too, so that onboarding a client is one automated flow rather than an automated flow with a manual permissions step in the middle.

### Workflow

1. The developer provisions a workspace end to end from the CLI: create it, set its timezone, invite members, scope their account access.
2. An AI assistant, through the MCP server, can read workspace settings and pause posting when the user asks.
3. The developer builds a no-code automation that creates a workspace and invites a team when a customer signs up in their own system.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Workspace settings read and write, pause and resume posting, and team account access management are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Pausing posting and changing account access through an AI assistant require the explicit confirmation the MCP epic defines for state-changing actions, since both affect a customer's live publishing.
- [ ] Zapier, Make and n8n expose at minimum creating a workspace with a timezone, inviting a member, and setting that member's account access.
- [ ] Where a surface deliberately omits a capability, that omission is recorded with its reason.
- [ ] The public documentation covers the capability on every surface, with a worked example of provisioning a client workspace end to end.
- [ ] The parity check passes for workspace settings and team access across every surface.
- [ ] Every surface enforces the same authorization as the REST endpoints. No surface allows a caller to widen a member's account access beyond their own.

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
