# Epic + stories — Publishing gaps on the public API

**Priority: P1.** 4 stories. Nothing is pushed to any tracker.

---

## Epic: Publishing gaps on the public API

Publishing is the strongest part of the public API. A caller can create a post across twelve platforms with per-platform overrides, X threads, Meta Threads threading, Facebook carousels and collaborators, Instagram collaborators and trial reels, LinkedIn polls, first comments, labels, campaigns, approvals and approval workflows, and can schedule, draft, queue or publish into a content category.

Four things are missing, and they are the kind of gap that costs an integration on every single run.

**There is no way to read one post.** A caller can list posts and cannot fetch one by id. So an integration that creates a post and later wants to know what happened to it has to page the list endpoint and filter client-side, every time, for every post. This is the most-hit hole in the publishing surface and the smallest to close.

**Posts cannot be filtered by account or platform.** The list endpoint filters by status, date, approval, labels, campaigns, category, creator and comment status, and not by the social account or the network. So "show me what is scheduled for this Instagram account" is not a question the API can answer, even though it is the first question anyone asks the planner.

**Repeat posts are not supported.** A recurring post is a real product feature with its own repeat type, interval, count and custom schedule. None of it is accepted when creating a post through the API, so a customer who recycles content on a schedule cannot set that up programmatically.

**Post share links do not exist on the API.** A share link is how a customer puts scheduled content in front of a client for review without giving them a login: the client sees the posts, comments, and approves or rejects. It is one of the most agency-relevant features in the product and it is entirely browser-only.

### A note on share links

There are two share link systems and this epic covers one of them. **Post share links**, for client review of scheduled content, are this epic. **Analytics share links**, for sharing a report, are already scoped in the analytics reports epic. They are separate features with separate storage, and the two stories must not be merged or built twice.

### Out of scope

Named explicitly because they were considered and not scoped: bulk post operations, post version history, queue shuffling, planner saved views, and calendar notes. Still open, awaiting a decision.

### Stories

1. `[BE] Add single post read and account and platform filters to the public API`
2. `[BE] Add repeat posts to the public API`
3. `[BE] Add post share links to the public API`
4. `[BE] Expose the publishing additions on every developer surface`

---

## [BE] Add single post read and account and platform filters to the public API

### Description

As a developer building a publishing integration, I want to fetch one post by its id and filter the post list by social account and platform, so that I can check what happened to a post I created without paging the whole list, and answer what is scheduled for a given account.

### Workflow

1. The developer creates a post and keeps its id.
2. The developer later fetches that post by id and sees its current state: whether it published, when, to which accounts, and what failed if anything did.
3. The developer lists posts filtered by one or more social accounts.
4. The developer lists posts filtered by platform.
5. The developer combines those filters with the ones that already exist.

### Acceptance criteria

- [ ] A caller can fetch a single post by id, and the response contains everything the list endpoint returns for that post plus its per-account publishing outcome.
- [ ] The single-post response states, per account, whether the post published, when, the link to the published item where one exists, and the failure reason where it failed.
- [ ] Fetching a post that does not exist, or exists in another workspace, returns a not-found response that does not reveal whether the id exists elsewhere.
- [ ] The list endpoint accepts a filter by one or more social accounts.
- [ ] The list endpoint accepts a filter by one or more platforms.
- [ ] The new filters combine correctly with every existing filter, verified in combination and not only individually.
- [ ] The list endpoint's existing behavior, response shape, default sort and pagination are unchanged for callers that pass no new filters.
- [ ] Filtering by an account the caller cannot access returns no results rather than an error that confirms the account exists, and never returns another workspace's posts.
- [ ] Authorization uses the existing workspace permission model, and posts hidden from the caller by existing rules stay hidden.
- [ ] Both additions appear in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Read-only.

### Impact on other products

None. Read-only additions that do not change existing responses.

### Dependencies

- None.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add repeat posts to the public API

### Description

As a developer automating a customer's recurring content, I want to create repeating posts through the API, so that a post that should go out every week can be set up once programmatically instead of by hand in the browser.

### Workflow

1. The developer creates a post and specifies that it repeats: how it repeats, how often, and how many times.
2. The developer optionally gives it a custom schedule instead of a simple interval.
3. The developer reads the repeating post and sees its schedule and the occurrences it has produced.
4. The developer updates the repeat settings, or stops the repetition.
5. The developer deletes a single occurrence, or the whole repeating series, and the response makes clear which happened.

### Acceptance criteria

- [ ] A caller can create a post that repeats, specifying the repeat type, the interval, and the number of repetitions, matching the options the product offers.
- [ ] A caller can specify a custom schedule instead of a simple interval, where the product supports it.
- [ ] Reading a repeating post returns its repeat settings and its occurrences, so a caller can see what is scheduled without reconstructing it.
- [ ] A caller can update a repeating post's settings, and the response states plainly which occurrences the change affects.
- [ ] A caller can stop repetition without deleting the occurrences already scheduled.
- [ ] A caller can delete a single occurrence, or the whole series, and the two are distinct operations with an unambiguous response.
- [ ] Repeat settings interact correctly with every existing scheduling option, and any combination the product forbids, such as repeating a post into a content category, is refused with a clear reason rather than accepted and silently ignored.
- [ ] Validation matches the product on repeat limits, and exceeding them returns a clear error.
- [ ] A repeating post created through the API is indistinguishable from one created in the browser, verified in the planner.
- [ ] Repeat occurrences count against the customer's post allowance exactly as they do in the product, and a repeat that would exceed the allowance is refused clearly rather than partially created.
- [ ] The existing post create and update endpoints keep their current behavior for callers that do not use repeat settings.
- [ ] The additions appear in the OpenAPI document with examples covering a simple interval and a custom schedule.

### Mock-ups

None. Backend-only story.

### Impact on existing data

A repeating post creates multiple scheduled occurrences, so a single API call can create many posts. The allowance and limit criteria above exist so this cannot be used to fill a customer's calendar unintentionally. No existing posts are migrated.

### Impact on other products

Repeating posts appear in the planner in the web app and the mobile app, and publish through the same pipeline. The story must verify an API-created repeating post behaves identically in both, including how an occurrence is edited or deleted from the app.

### Dependencies

- Depends on `[BE] Add single post read and account and platform filters to the public API` for reading a repeating post and its occurrences.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (validation and error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add post share links to the public API

### Description

As a developer building an agency's client review flow, I want to create and manage post share links through the API, so that a client can review and approve scheduled content without a ContentStudio login and without anyone setting the link up by hand.

### Workflow

1. The developer creates a share link over a set of posts or a filtered view of the planner.
2. The developer optionally protects it with a password, sets an expiry, and chooses whether the client may comment and approve.
3. The developer lists the workspace's share links, reads one, updates it, disables it, or deletes it.
4. A client opens the link, sees the posts, comments, and approves or rejects, exactly as they would from a browser-created link.
5. The developer reads what the client did: the comments left and the approvals or rejections given.

### Acceptance criteria

- [ ] A caller can create a post share link over a chosen set of posts or a filtered selection, matching what the product allows.
- [ ] A caller can set a password, an expiry, and whether the client may comment and take approval actions.
- [ ] A caller can list share links, read one, update it, disable it, and delete it.
- [ ] A link created through the API works exactly as a browser-created one, verified by opening it as an anonymous visitor and completing a review.
- [ ] Disabling or deleting a link stops it working immediately, verified.
- [ ] A caller can read the comments and approval actions a client left through a link.
- [ ] A link grants access only to the posts it was created for, read-only apart from the commenting and approval it was explicitly given. It cannot be used to reach any other post, workspace or capability, verified with a negative test.
- [ ] Client-hidden posts stay hidden through an API-created link exactly as they do today.
- [ ] Authorization uses the existing permission model. A caller who cannot create share links in the product cannot create them through the API.
- [ ] Every endpoint appears in the OpenAPI document with examples, and the documentation states plainly what a share link grants to whoever holds it.
- [ ] This story covers post share links only. Analytics share links are a separate feature covered in the analytics reports epic, and neither implementation is duplicated in the other.

### Mock-ups

None. Backend-only story. The client-facing shared view is existing product output and is not redesigned.

### Impact on existing data

Links created through the API are stored the same way as browser-created ones. Comments and approval actions a client takes through them are recorded the same way. Existing links are unaffected.

### Impact on other products

The shared view is opened outside the app entirely, often on a phone, and is served under white-label domains. The story must verify an API-created link renders and functions correctly under a white-label domain and on a mobile browser. Approvals given through a link must appear in the web app and the mobile app.

### Dependencies

- Coordinates with the notifications architecture epic, since share link invitations and client actions generate notifications through the existing path. This story must not create a second notification path.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (clients open shared links on phones, unchanged from today)
- [ ] Multilingual support (the shared view, password prompt and any invitation must respect locale as they do today)
- [ ] UI theming support (default + white-label)
- [ ] White-label domains impact review (shared links are among the most white-label-visible surfaces in the product)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose the publishing additions on every developer surface

### Description

As a developer using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want single-post reads, account filters, repeat posts and share links available there too, so that the publishing workflow I have already automated is complete rather than mostly complete.

### Workflow

1. The developer reads a post by id and filters posts by account from the CLI, and creates a repeating post from it.
2. An AI assistant, through the MCP server, can answer what is scheduled for a given account, report what happened to a specific post, and create a share link for a client to review.
3. The developer builds a no-code automation that creates a repeating post, or generates a review link when content is ready for a client.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Single-post read, account and platform filters, repeat post creation and management, and share link management are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools. Reading a single post and filtering by account are particularly valuable to an assistant, which currently cannot answer either question directly.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Creating a repeating post or a share link through an AI assistant requires the explicit confirmation the MCP epic defines for state-changing actions. A repeating post publishes many times on a customer's accounts, and a share link exposes unpublished content to whoever holds it, so neither should happen on a loose instruction.
- [ ] Every surface makes clear, at the point of creation, that a share link is accessible to anyone holding it, so a developer does not distribute one unknowingly.
- [ ] Zapier, Make and n8n expose at minimum reading a post by id, creating a repeating post, and creating a share link.
- [ ] Where a surface deliberately omits a capability, that omission is recorded with its reason.
- [ ] The public documentation covers the capability on every surface, with a worked example of a client review flow.
- [ ] The parity check passes for these additions across every surface.
- [ ] Every surface enforces the same authorization as the REST endpoints, including that a share link cannot be created over posts the caller cannot access.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on the three endpoint stories in this epic.
- Depends on the developer surface parity contract epic.
- Depends on the MCP epic for how an assistant confirms a state-changing action.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
