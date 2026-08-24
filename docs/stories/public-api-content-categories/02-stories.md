# Epic + stories — Content categories on the public API

**Priority: P1.** 3 stories. Nothing is pushed to any tracker.

---

## Epic: Content categories on the public API

Content categories are how a customer sets up evergreen, slot-based scheduling: a category holds a theme, the category holds time slots, and a post dropped into the category publishes at the next free slot. It is one of the most automation-friendly features in the product, and it is exactly what an API-first customer would want to drive programmatically.

The public API can list categories and can post into one. It cannot create a category, rename it, delete it, manage its slots, or find its next available slot. So an integration can use a category that a human already made in the browser, and nothing more. Setting up the automation, the part a developer would most want to script, is the part that has to be done by hand.

The slot side matters as much as the category itself. Without slot management an integration cannot build a schedule, and without a next-slot lookup it cannot tell the customer when a post it just queued will actually go out. Both exist in the product and neither is reachable.

This epic gives the API the whole feature: categories, their slots, and who on the team may use them.

### Out of scope

- Changing how categories or slots behave. This is exposure, not redesign.
- Global categories that span workspaces. The product code carries a note questioning whether they are still needed, so they are excluded pending that answer rather than exposed by default. Confirm before treating this as final.

### Stories

1. `[BE] Add content category CRUD and slot management to the public API`
2. `[BE] Add content category team access management to the public API`
3. `[BE] Expose content categories on every developer surface`

---

## [BE] Add content category CRUD and slot management to the public API

### Description

As a developer building on ContentStudio, I want to create and manage content categories and their posting slots through the API, so that I can set up slot-based scheduling programmatically instead of asking my customer to configure it in the browser first.

### Workflow

1. The developer lists a workspace's content categories.
2. The developer reads a single category, including its slots.
3. The developer creates a category, giving it a name and its settings.
4. The developer updates a category, or deletes it.
5. The developer adds, updates and removes posting slots on a category, and can clear all of a category's slots in one call.
6. The developer asks for a category's next available slot, and gets back when a post queued into it would publish.
7. The developer shuffles the queued posts in a category to reorder them.
8. The developer creates a post into a category, as they can today, and the post takes the next free slot.

### Acceptance criteria

- [ ] A caller can list a workspace's content categories, read a single category by id, create, update and delete a category.
- [ ] Reading a single category returns its slots, so a caller does not need a second call to understand the schedule.
- [ ] A caller can create, update and delete individual slots on a category, and clear all slots on a category in one call.
- [ ] A caller can request a category's next available slot and receives the time a post queued now would publish, in a stated timezone.
- [ ] A caller can shuffle the queued posts within a category.
- [ ] Validation matches the product: the same rules on category names, slot times and slot limits, with clear errors naming the field that failed.
- [ ] Authorization uses the existing workspace permission model. A caller who cannot manage categories in the product cannot manage them through the API.
- [ ] Plan gating is respected. A workspace on a plan without content categories receives a clear refusal, not an empty list.
- [ ] Deleting a category with posts queued into it behaves exactly as the product does, and the response makes clear what happened to those posts.
- [ ] Existing behavior is unchanged: the current list endpoint keeps its response shape, and posting into a category by id keeps working as it does today.
- [ ] Every endpoint appears in the OpenAPI document with request and response examples.
- [ ] Errors follow the existing public API error conventions rather than introducing a new shape.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Categories and slots are created and modified by callers as they would be in the product. No migration, and no change to how existing categories are stored.

### Impact on other products

The web app and the mobile app manage categories through their own endpoints and are unaffected. The story must verify that a category created through the API is indistinguishable from one created in the browser, in both clients.

### Dependencies

- None for the endpoints. The surface story in this epic depends on this one.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (validation and error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add content category team access management to the public API

### Description

As a developer managing a customer's workspace through the API, I want to control which team members may use which content categories, so that I can provision a workspace completely without a human finishing the job in the browser.

### Workflow

1. The developer lists the team members who have access to a given category.
2. The developer lists the categories a given team member has access to.
3. The developer grants or revokes a member's access to categories.
4. A member without access to a category cannot post into it, exactly as in the product.

### Acceptance criteria

- [ ] A caller can list the team members with access to a given category.
- [ ] A caller can list the categories a given team member has access to.
- [ ] A caller can set a team member's category access, granting and revoking in one call.
- [ ] Authorization matches the product: only a caller permitted to manage members can change category access.
- [ ] A member denied access to a category cannot post into it through any surface, verified through the API as well as the product.
- [ ] Changing access does not affect posts already queued into a category by that member, and the response or documentation states what happens to them.
- [ ] Endpoints appear in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Access records are created and removed as they would be in the product.

### Impact on other products

Category access is enforced in the web and mobile apps too. The story must verify that access set through the API is honoured in both.

### Dependencies

- Depends on `[BE] Add content category CRUD and slot management to the public API` for the category identifiers it operates on.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose content categories on every developer surface

### Description

As a developer using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want content category management available there too, so that I am not forced onto the REST API for one part of a workflow I otherwise drive from one place.

### Workflow

1. The developer manages content categories from the CLI.
2. An AI assistant, through the MCP server, can list, create and manage categories and read a category's next available slot on the user's behalf.
3. The developer builds a no-code automation in Zapier, Make or n8n that creates a category or queues a post into one.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Content category management is available in the CLI, with commands covering list, read, create, update, delete, slot management and next-slot lookup.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes the same capabilities as tools, so an assistant can manage categories on a user's behalf.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Zapier, Make and n8n expose the capabilities appropriate to each: at minimum creating a category, queueing a post into one, and reading the next available slot as a trigger or lookup.
- [ ] Where a surface deliberately omits a capability, for example destructive deletes in a no-code tool, that omission is recorded with its reason rather than left as a silent gap.
- [ ] The public documentation covers the capability on every surface, with a worked example of setting up slot-based scheduling end to end.
- [ ] The parity check passes for content categories across every surface.
- [ ] Every surface enforces the same authorization and plan gating as the REST endpoints. No surface is a weaker path to the same action.
- [ ] Customer-facing text on each surface, including command help, tool descriptions and error messages, is written for a developer reader and translated where the surface supports it.

### Mock-ups

None. No graphical UI in this story. Command help and tool descriptions are specified in the acceptance criteria.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on both endpoint stories in this epic.
- Depends on the developer surface parity contract epic. Without it this story is six separate hand-copied changes rather than one.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review (white-label customers use the same surfaces)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
