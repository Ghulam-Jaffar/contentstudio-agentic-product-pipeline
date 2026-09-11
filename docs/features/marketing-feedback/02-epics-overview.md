# Marketing feedback: epics overview

**Date:** 2026-09-11
**Source:** marketing team feedback batch, six asks
**Research:** `01-research.md`

Six separate epics, not one. Each is independently shippable and independently valuable, and each is handed to the Product Owner as its own epic. Every story in every one of them carries a `[Marketing Feedback]` tag at the end of its title so the batch stays traceable once the stories live in the tracker.

All six asks are now specced.

---

## The epics

| # | Epic | File | Stories |
|---|---|---|---|
| 1 | AI chat conversation context | `03-epic-ai-chat-context.md` | 5 |
| 2 | Ask AI about a highlighted passage | `04-epic-quote-reply.md` | 3 |
| 3 | AI chat images in the Content Library | `05-epic-ai-images-in-content-library.md` | 4 |
| 4 | Feature CTAs in AI chat | `06-epic-feature-ctas.md` | 3 |
| 5 | Bulk media download | `07-epic-bulk-media-download.md` | 3 |
| 6 | Multiple first comments with scheduling | `08-epic-multiple-first-comments.md` | 4 |

**22 stories across six epics.** Each epic carries its own `[Design]` story, so design work is scoped to the epic it serves rather than pooled.

---

## Story title tag

Every story title ends with `[Marketing Feedback]`:

> `[BE] Widen and enrich the AI chat history sent to the model [Marketing Feedback]`
> `[FE] Let users highlight text in AI chat and ask about that passage [Marketing Feedback]`

This is the only addition to the standard title convention. Stories still carry no metadata block of any kind and still end at the global quality checklist.

---

## What can run in parallel

All five epics are independent of each other, with one exception: **epic 2 depends on epic 1**, because a quoted passage travels to the model inside the same context contract epic 1 reshapes. Building epic 2 first would mean doing that integration twice.

Epics 3, 4 and 5 have no dependencies at all and can start immediately.

Within each epic, the backend stories lead and the `[Design]` story should start before the frontend work.

---

## Two research findings that changed the scope of the ask

**Chat pagination already exists on the backend.** The message history endpoint already accepts a cursor and reports whether more messages exist and where to continue from. The frontend already stores those values and then never reads them. So the infinite-scroll ask is **client-side only**, on web and in the Flutter app, and epic 1 contains no backend story for it. This is the cheapest high-visibility win in the whole batch.

**A feature-navigation mechanism already exists in AI chat, and it is unsafe.** Assistant replies can already carry a navigation action, and the app navigates to whatever address the model produced, with no check that the address is real and no check that the user's plan or role allows it. So epic 4 is half new feature and half correcting a live dead-link path.

---

## Epic 6: decisions taken

This was held for finalization and is now specced. The decisions that unblocked it:

- **"Conditional" means a condition per comment, measured against the item above it.** Comment one is timed from the post, every other comment from the comment above, so reordering can never make the list contradict itself. A condition is an operator, an amount and a unit, matching the condition pattern used elsewhere in the product. No condition means post immediately. Per-platform comment text was considered and **not** taken, so one account selection applies to every comment. Engagement-triggered conditions were ruled out, since nothing in ContentStudio polls post engagement for publishing decisions.
- **Platforms stay as they are today:** Facebook Pages, Instagram when posting through the API, LinkedIn, YouTube for public non-kids videos, and Telegram.
- **Full Flutter parity**, with the existing account-selection and post-loading gaps fixed first, because without that fix a user editing a post in the app silently destroys comments they set up on web.

**The cap needs a technical lead's read.** Three is committed and needs no sign-off. Five is wanted if nothing technical blocks it, and the epic carries four specific questions about rate limits, queued-job volume, payload headroom and anything that treats a published post as finished. The cap stays configuration either way.

Also still open, and flagged in the epic: the proposed mobile scope and workflow, whether the days unit should be capped since relative conditions accumulate, whether a duplicated comment should reset to immediately, and whether automations get multi-comment in a later release.
