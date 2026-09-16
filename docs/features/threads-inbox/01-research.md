# Research pointers: Threads inbox

**Date:** 2026-09-16
**Source:** engineering wrote two documents, a feasibility study and a PRD, both dated 2026-09-14. They live in the frontend repo's research and docs folders. This file is a pointer and a condensation, not a replacement.

This doc stays local. Codebase paths, entry points and gotchas live here and never in a story body.

---

## 1. Where the detail lives

- **Feasibility study**, `2026-09-14-threads-inbox-feasibility.md` in the frontend repo's research docs. Answers whether this is buildable, and at what cost. Tags every claim by provenance: read in our code, read in Meta's docs, third party, or inference.
- **PRD**, `PRD-threads-inbox-integration.md` alongside it. Answers how, against the code as of 2026-09-14, and corrects three things the feasibility study got optimistic about.

Both are thorough and should be read by whoever picks this up. The stories deliberately do not repeat their file lists, line numbers or code snippets.

---

## 2. Scope: locked 2026-09-14, widened 2026-09-16

**In:** replies to our posts, mentions of the account, reply and hide and unhide and delete-own as actions, repost and like counts.

**Out:** direct messages (no API exists for anyone), and any refactor of the strategy base class.

**Scope was widened by product on 2026-09-16.** Engineering's write-up deferred three things on cost grounds: sending media replies, reply approvals, and counts as a droppable phase. Product's decision is that anything the API supports ships in v1, so all three are in. Only genuine API impossibilities stay out.

Consequences of that decision, which the dev's documents do not reflect:

- **Sending media replies** needs the async container flow: stage, poll until processing completes, publish. That means a background job, an in-flight reply state, retries and a terminal failure state. The feasibility study explicitly scoped v1 as text-only to avoid exactly this. Note that *receiving* media was always free, since inbound replies already carry `media_url`, `thumbnail_url`, `children`, `gif_url` and `poll_attachment`.
- **Reply approvals** (`pending_replies`) are a moderation queue with no equivalent anywhere in our inbox. The feasibility study said this is the one thing that would need design. It now does, so the epic carries a `[Design]` story it originally did not need.
- **Counts** move from droppable-last-phase to on the critical path, which makes submitting the third scope a dated action with an owner rather than a maybe.

**App Review is already passed** for the two scopes the inbox needs. That removes the build-then-screencast-then-wait ordering that dominated the feasibility study's schedule risk. The **third** scope for counts has not been submitted, and under the widened scope it is now blocking rather than optional.

---

## 3. The five things that change how this is built

### 3.1 There is a live publishing bug, and it is not inbox work

Nothing refreshes Threads tokens. The long-lived token is exchanged once at connect and never again. Threads long-lived tokens expire at 60 days.

**Threads publishing is silently failing today for every account connected longer than that.** The inbox does not cause this, but it makes it loud, because a dead token turns an occasional failed post into a permanently empty inbox.

The PRD calls for this as its own ticket, shipping first. It is not one of the four tickets requested, and it is flagged in the epic.

### 3.2 Do not copy the YouTube strategy wholesale

YouTube's comment handler never publishes the realtime event that makes the inbox update without a refresh. Only Facebook, Instagram and LinkedIn do. Copying YouTube verbatim ships a Threads inbox that only updates on hard refresh, and it will pass a casual test.

Take the structure from YouTube and the realtime publish from LinkedIn, including two serialisation fixes that are easy to lose.

The PRD names this as the single most likely defect in the whole build and asks for a regression test that catches it by name.

### 3.3 Fourteen hardcoded platform lists, and only one fails loudly

There is no platform registry. Every list that mentions platforms names them explicitly, and **most fail silently** when one is missed: a 400, a skipped job that still returns success, or a null helper that logs and moves on. Only the Kafka consumer group map raises at startup.

Some of those lists are DM paths that Threads must **not** be added to, because Threads has no DMs and adding it there produces runtime failures.

The PRD carries the complete table with the failure mode for each. It is a checklist, not a footnote.

### 3.4 The frontend caps the reply affordance, not the render

Comment rendering is already arbitrary depth and matches the storage. What is capped is the reply button, which is gated at depth one for every platform except Facebook and Google Business Profile.

Threads' entire value is arbitrary-depth conversation. Without a Threads branch there, a user sees a six-deep thread and can only answer at the top. This is a one-line branch and it is easy to miss because everything renders correctly.

### 3.5 No Meta webhook payload is signature-verified today

None of the existing Meta webhook routers verify payload signatures. Anyone who can reach the service can inject inbox comments for any workspace.

This is pre-existing and out of scope to fix globally, but the PRD is explicit that **the Threads webhook must not inherit it**. Verifying the signature on the new router is about ten lines and is a trust boundary.

---

## 4. What fits with no change

Comment storage already handles arbitrary depth. A Threads reply is an existing element type, so no new type is needed. The upsert on comment id makes webhook redelivery and poll reconciliation deduplicate by construction. The auto-reply engine never branches on network above a single helper, so Threads is one branch. Action flags are set by the platform helper at parse time, so Threads controls its own affordances without touching shared code.

Threads also returns better data than Instagram: parent pointers are given rather than inferred, author details arrive inline rather than needing a second call per author, hidden state is server truth, and there is a field that identifies our own replies rather than making us compare ids.

---

## 5. Gaps in the research, relative to what was asked for

The two documents cover the inbox service and the web frontend thoroughly. They do **not** cover:

- **The public API, MCP tools or CLI.** No mention of exposing Threads inbox through any of them.
- **The AI assistant's inbox toolkit.** The assistant has inbox tools covering summary, listing, finding, threads, replies and moderation. Nothing in either document addresses Threads there.
- **Mobile.** The feasibility study concludes no mobile story is required for v1, on the grounds that the mobile inbox renders whatever element types the API returns and already handles this shape for YouTube and Google Business Profile. **That conclusion is tagged as an inference, not verified.** A Flutter ticket was requested, so it is written as verification plus the gaps that verification is likely to find.

---

## 6. Open questions to answer before estimating

Three are blocking, each answerable in under a day with a tester account:

1. **Does the conversation endpoint return replies on posts not published through us?** Highest value question in either document. If the answer is no, v1 covers only posts we published, and the headline requirement is unmeetable.
2. **What is returned for a reply from a private account?** Mentions from private accounts are documented as never returned. Whether replies are affected is unknown.
3. **Observed webhook latency and duplicate delivery rate.** Sizes the reconciliation poll cadence with data rather than a guess.

Two more were added by the widened scope:

4. **How the reply approvals queue actually behaves.** What is held, what the owner can do, whether decisions reverse, whether anything notifies us. This sizes the design story, so it blocks design, not just build.
5. **Real media processing time and failure rate.** Decides whether a sent media reply is a brief spinner or a genuine pending state the user can navigate away from.

Non-blocking: the poll cadence itself, how far back the first sync reaches, and who owns the third App Review submission.

---

## 7. Why there is now a design story

The feasibility study said everything in its recommended v1 renders in the current inbox without a new surface, and that a design story would be needed only if reply approvals were pulled into scope.

They have been. The design story covers the pending replies surface and the sending, failed and retry states for a media reply, which is the other thing the widened scope introduced that has no existing pattern.
