# Epic: Threads inbox

## Epic description

Threads accounts can be connected and published to, but there is no way to see or answer what people say back. Every reply and every mention is invisible in ContentStudio, so anyone managing a Threads presence has to leave the product and work in the Threads app, and nothing they do there is recorded against the post.

This epic brings Threads into the unified inbox as a first-class network, alongside Facebook, Instagram, LinkedIn, YouTube and Google Business Profile: replies on our posts, mentions of the account, and the ability to answer, hide, unhide, moderate and delete from inside ContentStudio, with the same assignment, tagging, notes, saved replies and automation that every other network already gets.

### Scope for v1

**The scope rule for this epic: if the Threads API supports it, we build it.** Engineering's original recommendation deferred three capabilities on cost grounds. Product has since decided that anything technically possible ships in v1, so all three are in. Only genuine API impossibilities are excluded.

**In scope**

- Replies to posts published by the connected Threads account
- Mentions of the connected account
- Actions: reply, hide, unhide, delete our own reply
- **Receiving media replies:** images, videos, carousels, GIFs and polls that people send us, rendered in the thread
- **Sending media replies:** replying with an image or video from inside the inbox
- **Reply approvals:** the queue of replies Threads holds for the account's review, with approve and reject
- Repost and like counts on Threads posts
- Threads coverage in the public API, MCP server, CLI and the AI assistant's inbox tools

**Out of scope, because the API does not allow it**

- **Direct messages.** Threads exposes no messaging API to anyone. This is not a prioritisation call.
- **Quote posts that do not mention the account.** Threads only delivers quotes that carry an actual mention, so a quote of our post from someone who did not tag us never reaches us. Nothing can be built to catch these.
- **Deleting replies written by other people.** Threads permits deleting only our own. Hiding is the equivalent action and it is in scope.

### What the widened scope costs

Three capabilities came in that the engineering research had deliberately parked. Each brings real work, and all of it lands in v1 rather than a follow-up.

**Sending media replies** cannot complete inside a single request. Media has to be staged, then polled until Threads finishes processing it, then published. That means a background job, a visible sending state on the reply, a failure path and a retry. A text reply is one round trip; a media reply is a small pipeline. This is the largest single addition.

**Reply approvals** are a moderation queue that no other network in the inbox has. There is no existing surface that fits, so this is the one part of the epic that genuinely needs design before it can be built, and it is why this epic now carries a design story it originally did not need.

**Repost and like counts** depend on a third permission scope that has not been submitted to App Review. It was previously the safe thing to drop. It no longer is, so **submitting that scope becomes a dated action with an owner, and it sits on the critical path.** Submit it now, in parallel with the build, not when the build is finished.

### Permission scopes

| Scope | Covers | State |
| --- | --- | --- |
| Read replies | Replies, threads, approvals queue | **Approved** |
| Manage mentions | Mentions of the account | **Approved** |
| Manage insights | Repost and like counts | **Not submitted.** Now on the critical path. Needs an owner and a submission date before this epic starts. |

### Before estimating: five questions to answer with a tester account

Each is answerable in under a day and each changes scope if it goes the wrong way.

1. **Do we get replies on Threads posts that were not published through ContentStudio?** The single most important unknown. If the answer is no, v1 only covers posts we published, and the epic's headline promise does not hold. Answer this first.
2. **How does the reply approvals queue actually behave?** What is held, what the account owner can do with it, whether a decision is reversible, and whether anything notifies us when a reply enters the queue. This directly sizes the design story, so answer it before design starts.
3. **How long does media processing take in practice, and how often does it fail?** This decides whether a sent media reply feels instant with a brief spinner or needs a genuine pending state the user can navigate away from.
4. **What comes back for a reply from a private account?** Mentions from private accounts are documented as never being delivered. Whether the same applies to replies is unknown, and it determines whether we need a placeholder for missing author details.
5. **What is the real webhook latency and duplicate-delivery rate?** This sizes the reconciliation poll with observed data instead of a guess.

### A live bug that blocks this epic and is not part of it

**Threads access tokens are never refreshed.** The long-lived token is obtained once when the account is connected and never renewed, and Threads long-lived tokens expire after 60 days.

This means **Threads publishing is already failing in production** for every account connected more than 60 days ago, and it fails quietly. The inbox does not cause it, but the inbox makes it much worse, because an expired token turns an occasional failed post into an inbox that is permanently and inexplicably empty.

This needs **its own ticket, and it needs to ship before the inbox work lands.** It is deliberately not folded into the backend story below, because it is a publishing bug that deserves to be tracked and fixed on its own timeline.

Related, and worth surfacing rather than treating as an error: private Threads profiles cannot refresh tokens at all and are on a permanent re-authentication cycle. That belongs in account health, not in an error toast.

### Stories

1. `[Design] Reply approvals and media reply states for the Threads inbox`
2. `[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`
3. `[FE] Add Threads as a network in the unified inbox`
4. `[Flutter] Verify and complete Threads inbox support in the mobile app`
5. `[Full Stack] Expose Threads inbox through the public API, MCP, CLI and the AI assistant`

Build order: design starts immediately and runs alongside the backend. Backend first, then the frontend and full-stack stories in parallel, then Flutter.

---

## Story 1

### Title

`[Design] Reply approvals and media reply states for the Threads inbox`

### Description

Two parts of the Threads inbox have no equivalent anywhere in the product today, and neither can be built by following an existing pattern.

**Reply approvals.** Threads holds some replies for the account owner to review before they appear publicly. No other network in our inbox has a moderation queue, so there is no surface to extend. The open questions are where a user encounters pending replies at all, whether that is a filter, a separate view or something on the conversation itself, how obvious we make it that replies are waiting, and what approving or rejecting looks and feels like, including whether a decision can be undone.

**Sending a media reply.** Unlike a text reply, which posts and appears immediately, a media reply is staged and processed before it goes live. That gap is short but not instant and it can fail. The reply needs a visible state while it is in flight, a clear failure state with a retry, and behaviour that makes sense if the user navigates away and comes back. Getting this wrong produces the worst outcome in an inbox, which is a user who cannot tell whether their reply was sent.

Everything else about the Threads inbox reuses the existing layout and needs no design.

### Workflow

**Pending replies**

1. The user opens the inbox and can tell, without hunting, that replies are waiting for review.
2. They reach the pending replies and see who wrote each one and what it says.
3. They approve one and it becomes publicly visible on Threads.
4. They reject one and it does not.
5. They understand what happens to a reply they leave alone.

**Sending a media reply**

1. The user attaches an image or video to a Threads reply and sends it.
2. The reply shows as sending, clearly enough that the user does not send it twice.
3. It completes and becomes an ordinary reply in the thread.
4. Or it fails, says why in words the user can act on, and offers a retry.
5. The user navigates away mid-send and comes back to a state that still makes sense.

### Acceptance criteria

- [ ] Designs cover where pending replies live in the inbox and how a user knows they exist.
- [ ] Designs cover the pending reply item itself, with approve and reject, and state clearly whether a decision is reversible.
- [ ] Designs cover the empty state when nothing is pending.
- [ ] Designs cover the sending, sent, failed and retry states for a media reply.
- [ ] Designs cover what a user sees on returning to a reply that was still sending when they left.
- [ ] Designs cover the media reply composer, including which media types are accepted and how limits are communicated before the user commits.
- [ ] Every state includes its copy, written for a non-technical user.
- [ ] Designs cover both desktop and mobile web, and are handed to the Flutter story in a form the mobile app can follow.
- [ ] Designs use existing design library components wherever one fits, and any genuine gap is named explicitly rather than drawn around.
- [ ] Designs are reviewed against the answer to the approvals behaviour question in the epic description, so they describe how the queue actually works rather than how we assume it works.

### Mock-ups

This story produces them.

### Impact on existing data

None.

### Impact on other products

The pending replies concept may be worth reusing if another network gains moderation later. Design it as a Threads feature, but do not design it in a way that could only ever be Threads.

### Dependencies

- Answers to questions 2 and 3 in the epic description, on approvals behaviour and media processing time. Design can start before these land but cannot be finalised without them.

### Global quality checklist

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 2

### Title

`[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`

### Description

The inbox service has no concept of Threads. This story adds it: ingestion of replies and mentions, storage in the existing comment structure, the actions a user can take on a Threads reply including media replies and approvals, and the realtime push that makes all of it appear without a refresh.

Threads returns better data than Instagram does for the same job. Parent pointers are given rather than inferred, author details arrive with the reply rather than needing a separate lookup per author, hidden state is server truth rather than something we track ourselves, and there is a field that tells us a reply is ours rather than making us compare account identifiers. Take advantage of that instead of copying Instagram's workarounds.

Threads conversations are arbitrarily deep by design. The existing comment storage already handles arbitrary depth, so nothing new is needed there, but the depth is real and must survive ingestion intact.

**Three things that are easy to get wrong and expensive to find later:**

**Realtime updates.** Some existing network handlers publish the realtime event that makes a new comment appear in an open inbox, and some do not. If Threads is modelled on one of the handlers that does not, the result is an inbox that only updates when the user refreshes the page, which will pass a casual test and be reported as a bug weeks later. Model the realtime behaviour on LinkedIn, which does this correctly. **Add a regression test that asserts a realtime event is published when a Threads reply is ingested.**

**Platform lists.** There is no platform registry. Around fourteen separate places in the service enumerate supported platforms by name, and **most of them fail silently when one is missed**: a rejected request, a job that skips its work and still reports success, or a helper that returns nothing and logs. Only one raises at startup. Threads must be added to every list it belongs in, and deliberately **kept out of the direct-message paths**, because Threads has no DMs and adding it there produces runtime failures. Treat this as a checklist to work through, not a detail.

**Post resolution.** The generic path for matching an incoming comment to its post looks for a platform identifier that does not exist for Threads. Threads needs its own explicit branch. Without it, replies arrive and are silently discarded.

**Security.** The new webhook endpoint must verify the payload signature before doing anything with the body. Existing Meta webhook endpoints do not verify signatures, which means anyone who can reach the service can inject inbox content for any workspace. That is pre-existing and out of scope to fix everywhere, but **the Threads endpoint must not inherit it.**

**The missed-interaction guarantee.** Webhooks are best effort. A periodic reconciliation poll, bounded to the last 7 days, is what actually guarantees nothing is lost. Because comments are stored keyed on their platform identifier, the poll and the webhook deduplicate against each other by construction, so the poll is safe to run frequently. Ingestion runs on its own dedicated consumer groups so a backlog on Threads cannot slow down any other network.

**Media replies are the one genuinely new mechanism.** Every other action in the inbox completes inside the request that triggered it. A media reply does not: the media is staged, Threads processes it for an unpredictable interval, and only then can it be published. This needs a background job, a stored in-flight state the frontend can display and poll, a bounded number of retries, and a terminal failure state with a reason the user can act on. It must not block the request, and a reply left in flight must never be silently lost.

**Reply approvals are a new element state.** Replies Threads holds for review must be ingested and distinguishable from ordinary replies, with approve and reject actions that reach Threads and update state.

### Workflow

**A reply arrives**

1. Someone replies to a Threads post published from a connected ContentStudio account.
2. The reply appears in the user's inbox within seconds, with the author's name and avatar, the text, any attached media, the timestamp, and its position in the conversation thread.
3. If the user had that inbox open, it appears without them refreshing.
4. The reply counts toward the unread count and toward inbox notifications exactly as a Facebook or LinkedIn comment does.

**A mention arrives**

1. Someone mentions the connected Threads account in a post of their own.
2. The mention appears in the inbox as an item the user can read and reply to.
3. Mentions from private accounts are not delivered by Threads at all. Nothing in the product should imply that they are.

**A reply is held for review**

1. Threads holds a reply pending the account owner's review.
2. It appears in the inbox as pending, clearly distinct from a published reply.
3. The user approves it and it becomes publicly visible on Threads.
4. The user rejects it and it does not.
5. The outcome is recorded against the user who decided it.

**A user replies with text**

1. The user types a reply in the inbox and sends it.
2. The reply posts to Threads and appears in the conversation, marked as ours.
3. The reply is recorded against the post and the user who sent it, for reporting and for the activity log.

**A user replies with media**

1. The user attaches an image or video and sends the reply.
2. The reply is accepted immediately and shown as sending. The user is not made to wait on the request.
3. The media is staged and processed, and the reply publishes when processing completes.
4. On success the reply becomes an ordinary reply in the thread.
5. On failure the reply is marked failed with a specific reason and can be retried.
6. A reply still in flight when the user leaves and returns is still in flight, not lost.

**A user hides, unhides or deletes**

1. The user hides a reply. It stays visible in ContentStudio, marked as hidden, and is hidden on Threads.
2. The user unhides it and it becomes visible again on Threads.
3. The user deletes a reply the account itself wrote. It is removed from Threads and marked as deleted in ContentStudio.
4. The user cannot delete a reply written by someone else, because Threads does not permit it. The action is not offered.

**Something is missed**

1. A webhook is dropped or delayed.
2. The reconciliation poll picks up anything from the last 7 days that did not arrive by webhook.
3. The user sees the interaction. They never see it twice.

**Automation**

1. Existing inbox automation, saved replies, assignment rules and tagging apply to Threads with no separate configuration.

### Acceptance criteria

**Ingestion**

- [ ] Replies on posts published from a connected Threads account are ingested and appear in the inbox.
- [ ] Mentions of a connected Threads account are ingested and appear in the inbox.
- [ ] Reply threading is preserved at the depth Threads reports, with no depth ceiling applied during ingestion.
- [ ] Author name and avatar are stored from the data the reply already carries, without an extra lookup per author.
- [ ] A reply written by the connected account is identified as ours using the field Threads provides for it.
- [ ] A reply's hidden state reflects what Threads reports, not a state we maintain separately.
- [ ] Media attached to an inbound reply is stored, covering image, video, carousel, GIF and poll, so the inbox can display it.
- [ ] Ingestion runs on dedicated consumer groups, so Threads volume cannot delay any other network.

**Realtime**

- [ ] Ingesting a Threads reply publishes the same realtime event that LinkedIn publishes, so an open inbox updates without a refresh.
- [ ] A regression test asserts that this event is published on Threads reply ingestion.

**Post resolution**

- [ ] An incoming Threads comment is matched to its post using an explicit Threads branch, not the generic identifier path.
- [ ] A comment that cannot be matched to a known post is logged with enough detail to diagnose it, and does not silently disappear.

**Platform coverage**

- [ ] Threads is added to every platform list where it belongs, verified list by list rather than by testing the happy path.
- [ ] Threads is **not** added to any direct-message path.
- [ ] Adding Threads to the consumer group configuration does not break service startup.

**Webhook security**

- [ ] The Threads webhook endpoint verifies the payload signature before processing the body.
- [ ] A request with a missing or invalid signature is rejected and logged, and no inbox content is created.
- [ ] The verification failure is observable in monitoring, so a misconfiguration is visible rather than silent.

**Reconciliation**

- [ ] A scheduled job reconciles Threads interactions from the last 7 days.
- [ ] An interaction delivered by both webhook and poll appears exactly once.
- [ ] A webhook redelivered by Threads does not create a duplicate.

**Text actions**

- [ ] Reply, hide, unhide and delete-own succeed against Threads and are reflected in the inbox.
- [ ] Delete is not offered for replies the account did not write.
- [ ] A failed action surfaces a specific reason to the user rather than a generic failure.
- [ ] An action attempted with an expired or revoked token surfaces as a reconnect prompt, not as an unexplained error.
- [ ] Exhausting the daily reply quota surfaces as a quota message with the real number, not a generic failure.

**Sending media replies**

- [ ] A reply with an image or an image and text publishes to Threads.
- [ ] A reply with a video and optional text publishes to Threads.
- [ ] The request that submits a media reply returns immediately. It does not wait for processing.
- [ ] The reply is stored in an in-flight state the frontend can display and poll.
- [ ] The reply transitions to published once Threads finishes processing, and a realtime event fires so an open inbox updates.
- [ ] Processing failure produces a terminal failed state with a specific reason, not a silent disappearance and not an indefinite spinner.
- [ ] A failed media reply can be retried without the user re-uploading.
- [ ] Retries are bounded, and a reply that exhausts them lands in the failed state.
- [ ] A reply left in flight across a service restart is picked back up, not orphaned.
- [ ] Media that exceeds Threads' accepted types, sizes or durations is rejected before staging, with a reason.

**Reply approvals**

- [ ] Replies Threads holds for review are ingested and stored in a pending state, distinguishable from published replies.
- [ ] Approving a pending reply publishes it on Threads and updates its state.
- [ ] Rejecting a pending reply is applied on Threads and updates its state.
- [ ] The decision is recorded against the user who made it, for the activity log.
- [ ] A reply decided directly in the Threads app, outside ContentStudio, reconciles to the correct state rather than sitting pending forever.
- [ ] A failed approve or reject surfaces a specific reason and leaves the reply pending rather than in an unknown state.

**Counts**

- [ ] Repost and like counts are read for Threads posts and stored alongside the existing engagement counts.
- [ ] This work is gated on the insights permission scope, which must be submitted and approved. It is in scope for v1 and is not treated as droppable.

**Automation**

- [ ] Auto-replies, assignment, tagging, notes and saved replies work on Threads items with no Threads-specific configuration.

### Mock-ups

None for the backend itself. The in-flight and pending states it stores are designed in `[Design] Reply approvals and media reply states for the Threads inbox`, and the state names should match what that story specifies.

### Impact on existing data

No migration for replies. Threads replies are stored as an existing element type in the existing comment structure, which already supports arbitrary depth.

Media replies and pending approvals introduce reply states that do not exist today. Adding them must not change how any other network's replies are read or displayed.

Historical Threads interactions from before this ships will not be backfilled beyond the reconciliation window, so the inbox starts roughly 7 days deep on first sync.

### Impact on other products

- The web inbox needs the companion frontend story to render any of this.
- The mobile app consumes whatever the inbox API returns, and is covered by its own story.
- The public API, MCP, CLI and AI assistant are covered by the full-stack story.
- **Blocked by the Threads token refresh fix.** Until tokens are refreshed, accounts connected more than 60 days ago cannot authenticate, so no amount of correct inbox code will produce an inbox for them.

### Dependencies

- Threads token refresh fix, tracked separately, must ship first.
- Approved permission scopes for reading replies and managing mentions. Both are already granted.
- **The insights permission scope must be submitted to App Review.** Counts are in scope for v1 and cannot ship without it. This needs an owner and a date before the epic starts.
- Answers to the five verification questions in the epic description, particularly whether replies are returned for posts not published through ContentStudio, and how the approvals queue actually behaves.

### Global quality checklist

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 3

### Title

`[FE] Add Threads as a network in the unified inbox`

### Description

The inbox does not offer Threads anywhere: not in the network filter, not in the account list, not in the empty states. This story adds it, so a user with a connected Threads account sees Threads interactions alongside every other network and can act on them.

Most of this is following the pattern that already exists for the other networks. Four things are not.

**Reply depth.** Comment threads already render to whatever depth the data contains, so Threads conversations will display correctly with no work. What is capped is the **reply button**, which is only offered at the top level for most networks. Threads conversations are deep by nature, and the whole point of a Threads inbox is being able to answer anywhere in the conversation. Without lifting that cap for Threads, a user sees a six-level-deep thread and can only reply at the top. **This will look completely correct in testing and be wrong in use.**

**Media in both directions.** People reply to Threads posts with images, videos, carousels, GIFs and polls, and all of it must render in the thread. Users can reply with media too, but a media reply does not post instantly the way a text reply does. It is staged and processed first, so the reply needs a visible sending state, a failure state with a retry, and behaviour that survives the user navigating away. Those states are specified in the design story.

**Reply approvals.** Threads holds some replies for review. This is a moderation surface that does not exist anywhere else in the inbox, and it is specified in the design story rather than invented here.

**Avatars.** Threads author avatar URLs expire. The avatar must be read from the stored channel image, not from the short-lived URL that came with the platform payload, or avatars will render on the day of testing and be broken images a week later.

Threads has real limitations users will hit, and they need to be explained in the interface rather than discovered. No direct messages. Only replies the account itself wrote can be deleted. Mentions from private accounts never arrive.

### Workflow

1. The user opens the inbox in a workspace with a connected Threads account.
2. Threads appears in the network filter with its icon, and the user can filter to it like any other network.
3. Threads accounts appear in the account selector.
4. The user selects a Threads item and sees the full conversation thread, with each reply's author, avatar, text, any attached media and timestamp, nested as it is on Threads.
5. A reply containing an image, video, carousel, GIF or poll displays that media inline, the same way media displays for other networks.
6. The user clicks reply on any reply in the thread, at any depth, and answers it.
7. The user attaches an image or video to their reply and sends it. The reply appears immediately in a sending state, then becomes an ordinary reply once it publishes.
8. If the media reply fails, the user sees why and can retry without re-uploading.
9. The user sees that replies are waiting for review, opens them, and approves or rejects each one.
10. The user hides a reply. It stays in the thread, visibly marked as hidden, with an unhide action available.
11. The user opens the actions on a reply the account did not write. Delete is not offered.
12. The user assigns the item, tags it, adds an internal note, or uses a saved reply, all exactly as they would on any other network.
13. A new Threads reply arrives while the inbox is open. It appears in place, without a refresh.
14. In a workspace with no Threads account connected, nothing about Threads is shown.

### Acceptance criteria

**Network presence**

- [ ] Threads appears in the inbox network filter with the correct icon and label.
- [ ] Threads accounts appear in the account selector for the inbox.
- [ ] Threads items appear in the unified list and in the unread count.
- [ ] Nothing Threads-related is shown in a workspace with no connected Threads account.

**Conversation rendering**

- [ ] A Threads conversation renders at the full depth returned, with correct nesting.
- [ ] Author name, avatar and timestamp render for every reply.
- [ ] Avatars are read from the stored channel image, not from the expiring platform-supplied URL.
- [ ] A reply written by the connected account is visually identified as ours.
- [ ] A hidden reply is visually marked as hidden.
- [ ] An inbound reply containing an image, video, carousel, GIF or poll renders that media inline, consistent with how the inbox renders media for other networks.
- [ ] A media type we do not render falls back to a readable placeholder rather than an empty bubble.

**Reply affordance**

- [ ] The reply action is available at every depth of a Threads conversation, not only at the top level.
- [ ] Replying to a nested reply posts the answer to that specific reply, not to the top of the thread.

**Composer and media replies**

- [ ] The Threads reply composer accepts an image or a video attachment.
- [ ] Accepted types, sizes and durations are communicated before the user commits, and an unsupported file is rejected with a specific reason.
- [ ] Sending a media reply shows it immediately in a sending state, matching the design story, so the user does not send it twice.
- [ ] The reply transitions to a normal published reply when it completes, without a refresh.
- [ ] A failed media reply shows a specific reason and a retry that does not require re-uploading.
- [ ] Navigating away and returning shows the reply in the correct state, still sending, published or failed.
- [ ] Character behaviour matches Threads' own reply limit, with the usual counter treatment.

**Reply approvals**

- [ ] The user can tell that Threads replies are waiting for review without hunting for them, per the design story.
- [ ] Pending replies show the author and the content, with approve and reject.
- [ ] Approving or rejecting updates the item immediately and reflects the result.
- [ ] A failed decision shows a specific reason and leaves the reply pending.
- [ ] The empty state when nothing is pending matches the design story.

**Actions**

- [ ] Hide and unhide are available and reflect the result immediately.
- [ ] Delete is offered only on replies the connected account wrote.
- [ ] A failed action shows a specific message. Suggested copy for an expired connection: `We could not reach Threads. Reconnect this account in Settings and try again.`
- [ ] Hitting the daily reply limit shows a specific message rather than a generic failure. Suggested copy: `You have reached the daily reply limit for this Threads account. Try again tomorrow.`

**Limitations copy**

- [ ] The direct messages area explains that Threads does not support messaging. Suggested copy: `Threads does not offer direct messages, so no conversations will appear here.`
- [ ] The empty inbox state for a Threads-only filter explains what will appear. Suggested copy: `No replies or mentions yet. New activity on your Threads posts will show up here.`
- [ ] Where mentions are explained, the private-account limitation is stated. Suggested copy: `Mentions from private accounts are not shared by Threads.`
- [ ] Where a user might expect to see a quote of their post, the limitation is stated. Suggested copy: `Threads only shares quotes of your posts when the author mentions your account.`
- [ ] All new strings in this story are added to all supported locales.

**Realtime and parity**

- [ ] A new Threads reply arriving while the inbox is open appears without a refresh.
- [ ] Assignment, tagging, internal notes, saved replies, filters and search all work on Threads items with no Threads-specific handling visible to the user.

### Mock-ups

From `[Design] Reply approvals and media reply states for the Threads inbox`, which covers the pending replies surface and the media reply sending, failed and retry states.

Everything else uses the existing inbox layout with no new surface. The only other new visual elements are the network icon and the limitation strings above.

### Impact on existing data

None. Read-only against existing inbox data structures.

### Impact on other products

Depends entirely on `[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`. Until that ships there is nothing to render.

Lifting the reply-depth cap must be scoped to Threads. It should not change behaviour for any other network.

### Dependencies

- `[Design] Reply approvals and media reply states for the Threads inbox`
- `[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`
- Threads network icon available in the shared icon set.

### Global quality checklist

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 4

### Title

`[Flutter] Verify and complete Threads inbox support in the mobile app`

### Description

The mobile inbox already renders post and comment items generically, and already handles this exact shape for YouTube and Google Business Profile. The research concludes that basic Threads replies may therefore work on mobile with no change at all.

**That conclusion is an inference. It has not been verified against a real Threads account, and it is not safe to ship on.** This story is verification first, then whatever the verification turns up.

It is also certainly incomplete now, because media replies and reply approvals did not exist when that conclusion was drawn. Neither has any mobile equivalent today.

The work is: connect a Threads account, generate real replies and mentions including a deep nested thread, media replies and a pending reply, and walk the mobile inbox end to end. Then close whatever gaps are found, and build the two surfaces that certainly do not exist.

### Workflow

1. The user opens the inbox in the mobile app with a connected Threads account.
2. Threads appears in the network filter with its icon.
3. Threads replies and mentions appear in the list alongside other networks.
4. The user opens a Threads item and sees the full conversation thread nested correctly.
5. A reply containing an image, video, carousel, GIF or poll displays that media inline.
6. The user replies to a nested reply, at any depth, and the answer posts to that reply.
7. The user attaches an image or video from their device and sends it. The reply shows as sending, then publishes.
8. If it fails, the user sees why and can retry.
9. The user sees that replies are waiting for review, and approves or rejects them.
10. The user hides and unhides a reply.
11. Delete is offered only on replies the account wrote.
12. A push notification for a new Threads reply opens the correct item in the inbox.
13. Assignment, tagging and internal notes behave as they do for any other network.

### Acceptance criteria

**Verification, done first**

- [ ] A Threads account is connected in a test workspace and real replies and mentions are generated against it, including a thread at least four levels deep, a reply with an image, a reply with a video, and a reply held for review.
- [ ] Every point in the workflow above is walked on both iOS and Android, and the result is recorded per point as works, broken, or missing.
- [ ] The findings are written into this story before any code is changed, so the actual scope is visible rather than assumed.

**Then close the gaps**

- [ ] Threads appears in the mobile inbox network filter with the correct icon and label.
- [ ] Threads items appear in the unified list, in filters, in search and in the unread count.
- [ ] Conversations render at full depth with correct nesting.
- [ ] Author name, avatar and timestamp render for every reply, using the stored channel image rather than an expiring platform URL.
- [ ] An inbound reply containing an image, video, carousel, GIF or poll renders that media inline.
- [ ] The reply action is available at every depth, not only at the top level.
- [ ] The composer accepts an image or video from the device camera or gallery, with unsupported files rejected with a reason.
- [ ] A media reply shows sending, published and failed states matching the design story, and survives backgrounding the app.
- [ ] A failed media reply can be retried without re-selecting the file.
- [ ] Pending replies are reachable, with approve and reject, following the design story.
- [ ] Hide, unhide and delete-own work, and delete is not offered on replies the account did not write.
- [ ] The same limitation strings as the web app are present and localised: no direct messages, mentions not shared by private accounts, quotes only when mentioned, and the Threads-only empty state.
- [ ] Push notifications for Threads replies deep link to the correct inbox item.
- [ ] Nothing Threads-related is shown in a workspace with no connected Threads account.
- [ ] Anything found in verification that is not covered above is either fixed or raised as a follow-up with a clear reason.

### Mock-ups

From `[Design] Reply approvals and media reply states for the Threads inbox`, for the pending replies surface and the media reply states. Everything else uses the existing mobile inbox layout.

### Impact on existing data

None.

### Impact on other products

Depends on the backend story. The limitation copy should match the web app word for word so the two products say the same thing.

### Dependencies

- `[Design] Reply approvals and media reply states for the Threads inbox`
- `[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`
- A test workspace with a connected Threads account, available before verification starts.

### Global quality checklist

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 5

### Title

`[Full Stack] Expose Threads inbox through the public API, MCP, CLI and the AI assistant`

### Description

Every network in the inbox is reachable from outside the web app: through the public API, through the MCP server, through the CLI, and through the AI assistant's inbox tools, which let a user ask about their inbox and act on it in conversation.

If Threads is added to the inbox and nowhere else, it becomes the one network that only exists in the web UI. A user who asks the assistant to summarise their inbox gets a summary with a hole in it, and worse, no indication that anything is missing. That is more damaging than an honest gap, because the answer looks complete.

This story closes that. It is the only part of this epic that the engineering research did not cover, so it needs its own look at each surface.

Two parts need thought rather than pattern-matching. **Deep threads:** the assistant's inbox tools were shaped around shallower networks, and fetching, summarising and replying into a six-level conversation all need to behave sensibly. **Asynchronous replies:** every inbox action these surfaces expose today completes when the call returns. A media reply does not, so the API needs to hand back something a caller can check, and the assistant needs to tell the user their reply is on its way rather than claiming it is done.

### Workflow

**Public API**

1. A developer lists inbox items and filters by Threads as a platform, and gets Threads replies and mentions back.
2. They fetch a single Threads conversation and get the full thread with parent relationships intact, including media attached to inbound replies.
3. They reply to a specific reply at any depth.
4. They reply with an image or video, get an immediate response with the reply's in-flight state, and can check whether it has published.
5. They list pending replies and approve or reject them.
6. They hide, unhide or delete an own reply.
7. Attempting a direct message operation on Threads returns a clear error explaining that Threads has no messaging, rather than an empty result.

**MCP and CLI**

1. Threads is listed wherever inbox platforms are enumerated, with the same capabilities and the same limitations described.
2. Every Threads inbox operation available over the API is available here, including media replies and approvals.

**AI assistant**

1. A user asks the assistant to summarise their inbox. Threads replies and mentions are included, and the summary counts them.
2. A user asks the assistant to find Threads items matching something, and gets them.
3. A user asks the assistant to show a Threads conversation. The assistant returns it with the thread structure readable rather than flattened into an unordered list, and notes where a reply carried an image or video.
4. A user asks the assistant to reply to a specific reply in a deep thread. The reply goes to that reply, not to the top of the thread.
5. A user asks the assistant to reply with an image. The assistant sends it and tells the user it is being processed, rather than claiming it is already live.
6. A user asks the assistant what is waiting for review, and can approve or reject from the conversation.
7. A user asks the assistant to hide a reply, tag an item, assign it or add a note. All work on Threads.
8. A user asks the assistant to send a Threads direct message. The assistant explains that Threads does not offer messaging rather than failing or inventing a result.
9. Any Threads action that changes something goes through the same confirmation the assistant already requires for inbox actions on other networks.

### Acceptance criteria

**Public API**

- [ ] Threads is accepted as a platform filter on inbox listing endpoints and returns Threads items.
- [ ] Fetching a Threads conversation returns the full thread with parent relationships intact at any depth.
- [ ] Fetching a Threads conversation returns media attached to inbound replies, covering image, video, carousel, GIF and poll.
- [ ] Replying to a Threads item supports targeting a specific reply at any depth, not only the top of the thread.
- [ ] A reply with media is accepted, returns immediately with the reply's in-flight state, and exposes a way for the caller to learn whether it published or failed.
- [ ] A media reply that fails exposes a specific reason, and retry is available.
- [ ] Pending Threads replies can be listed, approved and rejected.
- [ ] Hide, unhide and delete-own are available for Threads items.
- [ ] Delete on a reply the account did not write is rejected with a reason.
- [ ] A direct message operation on Threads returns a specific error stating Threads does not support messaging, not an empty success.
- [ ] Threads appears wherever the API documents or enumerates supported inbox platforms, with its limitations and its asynchronous reply behaviour documented.

**MCP and CLI**

- [ ] Threads is listed wherever inbox platforms are enumerated in the MCP server and the CLI.
- [ ] Every Threads inbox capability available over the public API is available through both, including media replies and approvals.
- [ ] The Threads limitations and the asynchronous media reply behaviour are stated in tool and command descriptions rather than only surfacing at runtime.

**AI assistant**

- [ ] Inbox summaries include Threads replies and mentions, and count them correctly.
- [ ] Inbox search and filtering return Threads items.
- [ ] Fetching a Threads thread preserves and conveys the conversation structure rather than flattening it.
- [ ] Media attached to inbound Threads replies is visible to the assistant, so a summary can mention that someone replied with an image.
- [ ] Replying targets the specific reply the user meant, at any depth.
- [ ] Sending a media reply reports it as being processed, and the assistant does not claim it is published until it is.
- [ ] Pending Threads replies can be surfaced, approved and rejected in conversation.
- [ ] Hide, unhide, delete-own, tagging, assignment and internal notes all work on Threads items.
- [ ] Asking for a Threads direct message produces an explanation that Threads has no messaging API, not an error and not a fabricated result.
- [ ] Threads actions that change something require the same user confirmation already required for inbox actions on other networks. Approving a pending reply and sending a media reply both count as changes.
- [ ] Nothing in the assistant's Threads handling depends on the connected accounts in a way that makes it silently skip Threads. If a user has no Threads account connected, the assistant says so rather than returning an empty summary.

**Consistency**

- [ ] The Threads limitation wording is the same across the API errors, the MCP and CLI descriptions and the assistant's explanations, and matches the web app's copy.

### Mock-ups

None.

### Impact on existing data

None.

### Impact on other products

Depends on `[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`.

The deep-thread handling in the assistant's inbox tools may improve behaviour for other networks that also support nesting. Any change there must be checked against Facebook and Google Business Profile so nothing regresses.

The asynchronous reply pattern introduced here is the first of its kind in these surfaces. It should be shaped so a future network with the same behaviour can reuse it.

### Dependencies

- `[BE] Ingest, sync and action Threads replies, mentions, media and approvals in the unified inbox`
- Threads support present in the public API before the MCP server, CLI and assistant work can be completed, since all three sit on it.

### Global quality checklist

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
