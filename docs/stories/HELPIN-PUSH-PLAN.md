# Helpin push plan — backlog batch items 12 to 28

Execution spec for pushing the 7 Aug 2026 batch into Helpin. Written before the MCP connection existed, so the next session can execute without re-deriving anything.

Source of truth for content: `docs/stories/BACKLOG-2026-08-07.md` and the 13 folders it indexes.

## Connection

- Server: `https://mcp.helpin.ai/mcp`, configured in the repo's `.mcp.json`.
- Remote MCP with OAuth. Runs as the signed-in user, with that user's role and permissions.
- **Default access is read-only.** Write on Projects & tasks must be granted in Helpin's MCP settings (Manage permissions) or every create silently isn't possible.
- Capabilities Helpin advertises: Workspace context, Projects & tasks, Docs, Helpin agents.

## What the PO specified

| Setting | Value |
|---|---|
| Target space | **Engineering** |
| Epic coverage | **Every story must belong to an epic.** Create epics where a story set doesn't already have one. |
| Workflow state | **Ready** for all stories |
| Sprint / iteration | **10 August 2026** |
| Research docs | Go in as **research docs** (Helpin Docs). **Unresolved:** whether this means the 13 `01-research.md` files become Docs linked to their epic (48 stories stay as stories), or whether the 5 `[Research]` tickets themselves become Docs instead of stories (43 stories). **Ask before creating.** |
| `[Design]` stories | Assigned to **Fasih** |

## Volume

- **13 epics** (7 already authored as epics, 6 to be created — see below)
- **49 stories**
- **13 research docs**

## Epics that already exist in the deliverables

These folders are already authored with an epic wrapper (Problem / Goal / Scope / Rules / Sequencing / Out of scope). Use that content as the epic description.

| Item | Folder | Epic title | Stories |
|---|---|---|---|
| 12 | `media-library-ai-creations-and-request-state` | AI Creations in the Content Library, and Media Library request state | 5 |
| 13 | `public-api-best-time-to-post` | Best Time to Post across the public developer surfaces | 10 |
| 14 | `public-api-ads-analytics` | Ads Analytics across the public developer surfaces | 11 |
| 18 | `brand-knowledge-inbox-autoreply-docs` | Manage Inbox auto-reply help docs from Brand Knowledge | 4 |
| 19 | `analytics-chart-ui-standardization` | Standardize chart UI across Analytics | 3 |
| 23, 24 | `analytics-network-research-bluesky-threads` | Analytics network research, Bluesky and Threads | 2 |
| 26, 27, 28 | `inbox-network-research-telegram-bluesky-threads` | Inbox network research, Telegram, Bluesky and Threads | 3 |

## Epics to create

These were authored as standalone story sets. Each needs an epic so no story is left without one. Proposed titles and descriptions; the description content is the intro prose already at the top of each `02-stories.md`.

| Item | Folder | Proposed epic title | Stories |
|---|---|---|---|
| 15 | `upgrade-plan-modal-pricing-sync` | Sync the in-app Upgrade Plan UI with the website pricing page | 2 |
| 16 | `automations-wizard-ui-finalization` | Finalize the Automations wizard UI | 2 |
| 20 | `analytics-metrics-consistency-audit` | Analytics metric consistency | 2 |
| 21 | `post-previews-inbox-analytics` | Extend the standardized post previews to Inbox and Analytics | 3 |
| 22 | `social-listening-fixes-and-web-blog-mentions` | Social Listening fixes and Blog and Website Mentions to production | 1 |
| 25 | `ai-chat-frame-reference-upload-gate` | AI Chat upload gate blocks reference images with no frame set | 1 |

## The 6 `[Design]` stories to assign to Fasih

1. `[Design] AI Creations pills, counts and empty states` (item 12)
2. `[Design] Reconcile the in-app Upgrade Plan UI with the website pricing page` (item 15)
3. `[Design] Define the Automations wizard standard across all three create flows` (item 16)
4. `[Design] Brand Knowledge document row with the Inbox auto-reply control` (item 18)
5. `[Design] Define the Analytics chart standard` (item 19)
6. `[Design] Published-post preview treatment for Inbox and Analytics` (item 21)

## Must resolve in Helpin before creating anything

Do not guess any of these. Read them from the workspace, and stop and ask if one is missing or ambiguous.

1. **The Engineering space** — its identifier, and whether epics and stories both live in it.
2. **The Ready state** — Helpin's exact state name and identifier for "Ready". Also the correct state for an *epic*, which may differ from a story's.
3. **The 10 August 2026 sprint** — its identifier. **Confirm it already exists.** If it doesn't, ask before creating a sprint.
4. **Fasih** — resolve to a Helpin user. If more than one match, ask.
5. **Description rendering.** Stories contain Mermaid fenced blocks, checkbox lists and tables. Verify on the first story whether Helpin renders Mermaid. If it does not, flag it: the Workflow diagrams will show as code blocks. Do not silently strip them.
6. **Epic-to-story linking** — whether a story takes an epic reference at creation or has to be linked afterwards.
7. **Doc attachment** — whether a Doc can be linked to an epic, and whether Docs live in the same Engineering space.

## Execution order

1. Read the tool schemas. Map Helpin's model (projects, tasks, docs) onto epic and story before creating anything.
2. Resolve all seven unknowns above.
3. **Dry run on item 25** — smallest unit, 1 epic + 1 story + 1 doc. Stop. Have the PO check it in Helpin, specifically that the description renders correctly.
4. **Then item 12** — 5 stories, containing a `flowchart TD` and a `stateDiagram-v2`, so it is the real test of Mermaid rendering. Stop again if anything looks wrong.
5. Then the remaining 11 folders, epic first, then its stories, then its research doc.
6. Write returned URLs back to `docs/stories/<slug>/03-helpin-links.md` per folder, so local docs and Helpin stay cross-referenced.

## Rules for the push

- **Never invent an identifier.** If a state, sprint, space or user can't be resolved, stop and ask.
- **Story bodies go in as authored.** Do not re-summarize. The sections (Description, Workflow, Acceptance criteria, Mock-ups, Impact on existing data, Impact on other products, Dependencies, Global quality and compliance checklist, Implementation references) are the deliverable.
- **Leave every checklist box unchecked.** They are for devs during implementation.
- **No estimates, no labels.** The team handles both.
- **Dependencies are expressed by full story title**, as authored. If Helpin supports real dependency links, add them *in addition to* the text, not instead of it.
- **Stop after the dry run.** Do not proceed from item 25 to the rest without the PO confirming.
- If a create fails partway through a folder, record what was created before stopping. Do not retry blindly and risk duplicates.

## Scope notes the PO already accepted

- **Item 17 is not in this push.** It duplicates two stories in the existing `brand-knowledge-media-assets` epic.
- **Item 20 is 2 stories, not 4.** Report widget selection and the Top Posts / Account Insights counts are already authored in `analytics-report-sections-and-overview-cards`.
- **Item 22 is one combined story**, as instructed, with a recommended 3-way split noted inside the story for the PO to take or leave.

## Unrelated but blocking-adjacent

A Shortcut **read-write** API token was embedded in a permission rule in `.claude/settings.local.json`. The rule has been removed from the working tree, but the token remains in git history (commits `d69ef9c`, `16a8197`, `9f4f3af`), so it is still recoverable by anyone with repo access. **Revoking it in Shortcut is the only real fix** and is still outstanding.
