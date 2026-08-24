# Epic + stories — Automations on the public API

**Priority: P3.** 4 stories. The largest surface in this programme, which is why it sequences after the parity contract is proven on smaller epics. Nothing is pushed to any tracker.

---

## Epic: Automations on the public API

Automations are the part of ContentStudio closest in spirit to what an API-first customer buys the API for, and they have no public API at all.

There are three kinds. **Evergreen** automations recycle a library of posts on a repeating schedule. The product can also generate AI variations of those posts, which is deliberately not part of the API scope here. **RSS** automations watch a feed and publish new items. **CSV bulk** automations take a spreadsheet of posts, process it, and let the customer review and act on each row before it goes out. Roughly fifty routes in the product, and none of it reachable programmatically.

The mismatch is the point. A customer who wants ContentStudio to publish from their own content pipeline is describing an RSS or bulk automation. A customer who wants their best content resurfaced on a schedule is describing evergreen. Both are automation-shaped, both are what the API exists for, and both currently require a human in the browser to set up and to maintain.

There is a second reason to do this: it is differentiation currently left on the table. The API-first competitors we benchmark against do not have automations. We do, and we do not expose them.

This epic gives the API all three automation types.

### The wizard question

Automations are built through a multi-step wizard in the UI, which raises a fair question: how does that work as an API?

It works because **the wizard is a UI affordance, not the way an automation is actually created.** The backend's save takes the entire automation configuration in one request and knows nothing about steps. For an evergreen automation that whole configuration is about a dozen fields: a name, the accounts, the schedule options, the posts, the cycle gap and whether it is enabled, the replan type, and optionally a content category. That is an ordinary request body.

The step machinery exists in exactly one place, the separate *draft* endpoints, and it exists for one reason: so a human who is halfway through the form can close the tab and come back. Those endpoints carry a step number and validate only the fields belonging to that step. The product's own code says as much, describing the draft save as what happens when the user clicks Next.

And the wizards are shorter than they look. All three are **two steps**: broadly, pick the accounts and settings, then supply the content.

So the design follows directly:

1. **Creation is one call**, hitting the same terminal save the wizard's last step hits. No step parameter, no partial state, no ordering requirement.
2. **The draft and step endpoints are deliberately not exposed.** They are resume-a-half-finished-form state for a browser session. An API caller has the whole configuration before it makes the call, so there is nothing to resume, and exposing a step counter would be modelling our UI in someone else's integration.
3. **A wizard gives a human incremental feedback, and an API caller needs an equivalent.** They get it two ways: per-field validation errors on the create call that name what failed rather than rejecting the payload wholesale, and an optional validate-without-creating call so a caller can check a configuration before committing to it. Both are acceptance criteria below.

Two things in this epic genuinely are multi-phase, and it is worth being clear that this is for data reasons rather than wizard reasons:

**Bulk CSV is the one genuine exception**, and it is staged for data reasons rather than wizard reasons: a file has to be uploaded before anything can be parsed, parsing is asynchronous because it can produce thousands of posts, and the posts have to exist before a caller can review them. Its story is written as an asynchronous batch with per-post actions, and section below explains the shape.

### How bulk CSV works as an API

The UI shows upload, then settings, then review. The API collapses the first two and keeps the third, because the first two are a form and the third is genuinely stateful.

1. **One call carries the file and the settings together.** This is already how the product works: the process call takes the uploaded file in the same request as the accounts, the name and the rest of the configuration. A caller does not upload, then configure, then trigger. It posts the file and its settings once.
2. **That call returns immediately** with a batch reference and the number of posts the file will produce, and the parsing runs asynchronously on its own queue. This is existing behavior, not new.
3. **The caller polls the batch** for progress and state.
4. **Parsing creates the posts in a pending state.** There is no separate confirm-the-batch step, because the posts already exist by the time review begins.
5. **The caller reviews and acts per post**: approve it, delete it, or enrich it first by attaching a label or a campaign. Approving is what schedules it. Actions work on one post or on many at once.
6. **Approval respects the caller's own permission.** A caller without scheduling permission approving a post sends it into the approval flow rather than straight to scheduled, exactly as in the product.

So the API shape is: one multipart call in, a batch to poll, a paginated list of pending posts to read, and per-post or bulk actions to approve or discard. That is an ordinary asynchronous batch resource. Nothing about it needs a wizard, and nothing about it needs the draft endpoints.

### Out of scope

- The draft and step endpoints, per the reasoning above. Creation is one call.

- Changing how any automation behaves, including the AI variation generation.
- Discovery and content curation, which feed automations but are an exploratory UI feature and are not in scope.

### Stories

1. `[BE] Add evergreen automation management to the public API`
2. `[BE] Add RSS automation management to the public API`
3. `[BE] Add bulk CSV automation to the public API`
4. `[BE] Expose automations on every developer surface`

---

## [BE] Add evergreen automation management to the public API

### Description

As a developer automating a customer's content recycling, I want to create and manage evergreen automations through the API, so that a library of recurring content can be provisioned and maintained from my own system instead of by hand.

### Workflow

1. The developer lists a workspace's evergreen automations.
2. The developer reads one, including its posts, its schedule and its settings.
3. The developer creates an automation: which accounts, what schedule, what settings.
4. The developer adds posts to the automation, and removes them.
5. The developer updates the automation, pauses or resumes it, and deletes it.
6. The developer reads when the automation next publishes.

### Acceptance criteria

- [ ] Creating an automation is a **single call** that accepts the complete configuration: name, accounts, schedule options, posts, cycle gap settings, replan type, and optionally a content category. There is no step parameter and no required call ordering.
- [ ] A caller can optionally validate a configuration without creating anything, receiving the same errors a create would return. This is the API's substitute for a wizard's per-step feedback.
- [ ] Validation errors name every field that failed and why, rather than rejecting the payload with one generic message. A caller must be able to fix a configuration from the response alone.
- [ ] The draft and step endpoints used by the browser wizard are not exposed, and the browser wizard continues to use them unchanged.
- [ ] A caller can list evergreen automations, read one by id, create, update and delete one.
- [ ] Reading an automation returns its posts, its schedule and its settings, so the caller can understand it without further calls.
- [ ] A caller can add and remove posts on an automation as a sub-resource, so a large content library does not have to be sent in the create call. The product already supports adding a post to an existing automation, and this exposes that.
- [ ] Posts can be supplied inline in the create call for a small set, or added afterwards through the sub-resource for a large one. Both paths produce an identical automation.
- [ ] Reading an automation's posts is paginated, so a large library does not produce an unbounded response.
- [ ] A caller can pause and resume an automation without deleting it.
- [ ] A caller can read when an automation next publishes.
- [ ] AI variation generation is not exposed. A caller supplies its own post text, and the product's variation feature continues to work unchanged in the browser.
- [ ] Validation matches the product on schedules, account selection and post limits, with clear errors.
- [ ] Authorization uses the existing workspace permission model, and a caller can only build automations over accounts they have access to.
- [ ] Plan gating is respected.
- [ ] Deleting or pausing an automation behaves exactly as the product does for posts it has already queued, and the response states what happened.
- [ ] Every endpoint appears in the OpenAPI document with examples, including the async variation flow.

### Mock-ups

None. Backend-only story.

### Impact on existing data

Automations created through the API are stored the same way as product-created ones. No AI credits are consumed, since variation generation is out of scope.

### Impact on other products

Automations created through the API appear in the product's automation views and publish through the same pipeline. The story must verify they are indistinguishable there.

### Dependencies

- None.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (validation and error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add RSS automation management to the public API

### Description

As a developer connecting a customer's content pipeline to their social accounts, I want to create and manage RSS automations through the API, so that publishing from a feed can be set up and adjusted programmatically.

### Workflow

1. The developer lists a workspace's RSS automations.
2. The developer reads one, including its feed, accounts and settings.
3. The developer creates an automation from a feed URL, choosing accounts and settings.
4. The developer updates it, pauses or resumes it, and deletes it.
5. The developer pulls the feed's history to backfill, as the product allows.
6. The developer reads what the automation has published and what it will publish next.

### Acceptance criteria

- [ ] Creating an automation is a **single call** that accepts the complete configuration. There is no step parameter and no required call ordering.
- [ ] A caller can optionally validate a configuration without creating anything, including checking that the feed is reachable and parseable, so a caller can verify a feed before committing to an automation.
- [ ] Validation errors name every field that failed and why, rather than rejecting the payload with one generic message.
- [ ] The draft and step endpoints used by the browser wizard are not exposed, and the browser wizard continues to use them unchanged.
- [ ] A caller can list RSS automations, read one by id, create, update and delete one.
- [ ] A caller can pause and resume an automation.
- [ ] A caller can trigger a feed history pull, matching the product's behavior.
- [ ] A caller can read what the automation has published and what is queued next.
- [ ] Creating an automation validates the feed URL and returns a clear, specific error when a feed is unreachable, empty, or not a valid feed, rather than accepting it and failing silently later.
- [ ] Validation otherwise matches the product on schedules, account selection and settings, with clear errors.
- [ ] Authorization uses the existing workspace permission model, and plan gating is respected.
- [ ] Deleting or pausing an automation behaves exactly as the product does for items already queued, and the response states what happened.
- [ ] A history pull is rate-limited so it cannot be used to flood a customer's queue, and the limit is documented.
- [ ] Every endpoint appears in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

Automations created through the API are stored the same way as product-created ones. A history pull can create a large number of queued posts, so the story must state the limits that apply.

### Impact on other products

RSS automations created through the API appear in the product's automation views. The story must verify this.

### Dependencies

- None beyond the shared conventions in this epic.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (validation and error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add bulk CSV automation to the public API

### Description

As a developer moving a batch of posts into ContentStudio, I want to submit a bulk file with its settings in one call and then review and approve the posts it produces, so that importing a customer's content calendar does not require a person uploading a spreadsheet.

### Workflow

1. The developer submits the bulk file together with its settings, in one call: the accounts to post to, a name for the batch, and the rest of the configuration.
2. The call returns immediately with a batch reference and how many posts the file will produce. Parsing runs in the background.
3. The developer polls the batch and sees its progress and state.
4. When parsing finishes, the posts exist in a pending state. The developer reads them, paginated, with any per-post problem attached to the post it affects.
5. The developer optionally enriches posts before approving: attaching a label, attaching a campaign.
6. The developer approves posts, one at a time or many at once. Approving is what schedules them.
7. The developer deletes posts it does not want, one at a time or many at once.
8. The developer abandons the whole batch, and nothing that has not already been approved is scheduled.

### Acceptance criteria

**Submission**

- [ ] A caller submits the file and its settings in a **single call**. There is no separate upload step, no separate configure step, and no trigger step.
- [ ] The call returns immediately with a batch reference and the number of posts the file is expected to produce, and parsing continues in the background. This matches how the product already behaves.
- [ ] The accepted file format and its column definitions are documented precisely enough for a developer to build a valid file without trial and error, with a downloadable sample.
- [ ] Submitting a malformed file, or one whose columns do not match, returns a clear error identifying what is wrong rather than a generic failure.
- [ ] Limits on file size and row count match the product, are documented, and exceeding them is refused before any processing starts.
- [ ] The number of posts a file will produce accounts for the number of accounts selected, as the product does, and that figure is returned before the caller commits to it.

**Batch state and review**

- [ ] A caller can poll the batch and read its state, distinguishing queued, parsing, ready and failed, with progress while parsing and a reason on failure.
- [ ] Once parsing completes, the posts it produced exist in a pending state and can be read, paginated, so a large batch does not produce an unbounded response.
- [ ] A problem with an individual row is attached to the post it affects, not returned as one aggregate error for the batch.
- [ ] A caller can list its batches with their state and creation time.

**Acting on posts**

- [ ] A caller can approve a single pending post, and approving is what schedules it.
- [ ] A caller can approve many pending posts in one call.
- [ ] A caller can delete a single pending post, and delete many in one call.
- [ ] A caller can attach a label and a campaign to pending posts before approving them.
- [ ] Approval respects the caller's own permission: a caller without scheduling permission approving a post sends it into the approval flow rather than straight to scheduled, exactly as in the product.
- [ ] A caller can abandon a batch, and nothing still pending is scheduled. Posts already approved are unaffected, and the response makes that distinction explicit.
- [ ] There is deliberately **no confirm-the-batch operation**, because the posts already exist by the time review begins. The documentation says so plainly, so a caller does not go looking for one.

**General**

- [ ] Posts produced by a batch are identical to posts created through the existing post endpoint, and publish through the same pipeline.
- [ ] Authorization uses the existing workspace permission model, and a caller can only target accounts it has access to.
- [ ] Plan gating is respected, and the batch counts against the customer's post allowance as it does in the product. A batch that would exceed the allowance is refused with a clear reason before parsing, not partway through.
- [ ] Bulk AI caption generation for rows that need it is available, and consumes credits exactly as in the product, with the cost surfaced or the request clearly refused when the allowance is exhausted. **This is the one AI-credit-bearing part of the epic. If AI is to be kept out of the automations API entirely, as it is for evergreen variations, drop this criterion and the capability with it.**
- [ ] Every endpoint appears in the OpenAPI document with examples covering the full flow from submission to approval, plus the sample file.

### Mock-ups

None. Backend-only story.

### Impact on existing data

A batch creates pending posts, and approving them schedules them. Because one call can create a large number, the limit and allowance criteria above exist so a customer's calendar cannot be filled unintentionally. A parsing failure partway through must leave a comprehensible state: the batch reports what it produced and what it did not, rather than leaving half a calendar with no explanation.

### Impact on other products

Posts produced by a batch appear in the planner in the web app and the mobile app, and pending posts appear there for review as they do today. The story must verify a batch submitted through the API is indistinguishable from one uploaded in the browser, including that a post can be approved from the app.

### Dependencies

- Creates posts through the same path as the existing post endpoint rather than a parallel one.
- If bulk caption generation is kept, its credit consumption should be recorded through whatever the usage visibility epic establishes rather than a parallel path.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (validation and per-post error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose automations on every developer surface

### Description

As a developer or agency using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want automations available there too, so that the most automation-shaped feature in the product can be driven from the tools I automate with.

### Workflow

1. The developer manages evergreen and RSS automations from the CLI, and submits bulk files from it.
2. An AI assistant, through the MCP server, can set up an evergreen automation or an RSS feed on the user's behalf.
3. The developer builds a no-code automation that creates an RSS automation when a customer adds a feed in their own system, or pushes a batch of posts as a bulk import.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Evergreen, RSS and bulk automation management are available in the CLI, including submitting a bulk file and acting on its rows.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Creating or deleting an automation through an AI assistant requires the explicit confirmation the MCP epic defines for state-changing actions, since an automation publishes on a customer's accounts repeatedly and without further review.
- [ ] Zapier, Make and n8n expose at minimum creating an RSS automation, adding posts to an evergreen automation, and pausing or resuming an automation.
- [ ] The asynchronous nature of bulk processing is handled correctly on every surface: no surface blocks indefinitely, and each has a documented way to learn a batch is ready for review.
- [ ] Where a surface deliberately omits a capability, that omission is recorded with its reason. Bulk file submission in a no-code tool in particular may not be practical, and whichever way that is decided is documented.
- [ ] The public documentation covers the capability on every surface, with a worked example for each automation type.
- [ ] The parity check passes for automations across every surface.
- [ ] Every surface enforces the same authorization, plan gating and credit consumption as the REST endpoints.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None beyond what a caller deliberately creates.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on the three endpoint stories in this epic.
- Depends on the developer surface parity contract epic. This epic is the largest in the programme, so shipping it before the contract exists means the most hand-copying of any epic here.
- Depends on the MCP epic for how an assistant confirms a state-changing action.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
