# Cancellation and retention flow - Helpin links

Pushed 2026-09-28 by Ghulam Jaffar. Workspace **ContentStudio**, Engineering team.

## Epic

| | |
|---|---|
| **Cancellation and retention flow** | `0ba2734b-bf57-4b79-84d2-a76ccbc00ebe` |
| Team | Engineering `f6b7c16a-8278-48f8-9d31-ffa89ee2cb87` |
| Owner | none, by request |
| Epic state | not set, which Helpin accepts |

## Stories

Seven, all in sprint **Week 41-42, 2026** (`8ae1115d-b22b-45b2-9bcb-ad0ecf77855d`, 5 to 18 October), state **Ready**, priority **urgent**, **no assignee**.

| Key | Title | Label | Task ID |
|---|---|---|---|
| CONT-4128 | `[Design] Design the cancellation and retention flow screens` | UIUX | `f154af70-93d0-4fa4-a731-ec6b0a539dde` |
| CONT-4129 | `[BE] Build the save offer engine for the cancellation flow` | backend | `f83ad5dc-c833-4ced-8079-18628cb7d92f` |
| CONT-4130 | `[BE] Apply a permanent 20% save discount and record it against the account` | backend | `f8b5c879-ad73-44ae-925e-6c5c2f318a04` |
| CONT-4131 | `[BE] Add a fixed one to three month pause with a scheduled resume date` | backend | `611cea0f-a8a3-4f51-af90-6689afd86df1` |
| CONT-4132 | `[BE] Restore held scheduled posts and automations when a subscription resumes` | backend | `74496f1d-1ac1-473d-a996-3a8215db8c49` |
| CONT-4133 | `[FE] Rebuild the cancel plan dialog with the consequences, reason and offer screens` | frontend | `658dbfa8-9948-49c6-ba13-a8c20aebb5a4` |
| CONT-4134 | `[FE] Add a Pause plan button and pause picker to the billing page` | frontend | `6d34eb7d-a68e-4f38-8dd1-55d6dbf07bc4` |

## Docs

All three created in the Engineering space (`bc4df642-96cf-44b0-b345-7a3e5c4ac8b4`) and **attached to the epic**.

| Doc | Collection | Document ID |
|---|---|---|
| Cancellation and retention flow - workflow | PRDs & Feature Specs | `4e911a9e-efcb-4472-84b3-5d90799b69d1` |
| PRD: Cancellation and retention flow | PRDs & Feature Specs | `a8d1ac51-9d5c-45e8-828c-7238fe163708` |
| Cancellation and retention flow - research | Research | `2694363c-928f-4f01-8b52-dca50e1263af` |

## Artifacts referenced from every story

- **Prototype**, version 43: https://claude.ai/artifact/2RuyPekdyeQheZCdknWfQ8
- **Spec**, revision 89: https://claude.ai/code/artifact/0eff1388-eb6c-4b8c-b541-28d889edc0e1

### Flow canvas, added 2026-10-06

Built for meeting visibility: every screen on one pannable board, branch logic as wires, and a green build note on each part naming the ticket that delivers it.

https://claude.ai/artifact/DA2AkQqjRLFJzsJYj9frKz

**This one is private.** The prototype and spec are visible to the organisation; the canvas is not until it is shared from its Share menu. Do that before the meeting if anyone else needs to open it.

It reuses the locked prototype's own screen markup and design tokens, so the screens on the canvas are the screens in the prototype rather than a redrawn copy. A copy change in the prototype does not propagate automatically, though, since the canvas holds its own copy of the node list.

## Notes from the push

- **CONT-4133 needed a second call.** `create_task` returned its description as raw markdown rather than converting it, unlike the other six. It was repaired with `update_task` and hand-built HTML. The body is the longest of the seven, which is the likely cause. Check the returned `description` on any long story rather than assuming conversion happened.
- **`agents/scripts/md-to-helpin-html.py` gained table and fenced-code support** during this push. It previously flattened a markdown table into a paragraph of pipe characters, which would have silently mangled any epic description containing one.
- **`list_sprints` now exists** over MCP. The push spec said it did not, and that sprint IDs were only discoverable by reading them off an existing task. That is out of date. The next sprint resolved cleanly.
- **No auto-labelling this time.** Every task came back with exactly the one label sent. Verified after creation.
- **An adjacent epic already exists**: *Paddle Retain (ProfitWell) - Cancellation & Churn Reduction* (`fef7121d-41e2-4329-967f-bf557bf48d15`), from December 2025, six unstarted tasks, deadline long past. That is the buy-it-off-the-shelf approach to the same problem. Not touched. Worth a decision on whether to archive it now that this epic exists.

## What the PO still has to decide

Carried from the PRD's open questions, none of which block starting:

1. Are the four existing Paddle discount codes set to 20%, and should the flow reuse them or get its own code for reporting?
2. Does pausing suspend the jobs that refresh social tokens? This one gates offering a three month pause at all.
3. Can a team member without billing permission reach Cancel plan?
4. Keep the two current reasons this spec drops: *already paying for another account*, and *credit card issue*?
5. After a pause ends, how far ahead is the customer warned before billing restarts?
6. Which booking page does **Book a demo** open, and who picks the booking up? Nothing in the product owns one today.
7. Accounts already sitting in an indefinite pause have held posts and automations from before CONT-4132 exists. Restore those on resume, or leave them?
