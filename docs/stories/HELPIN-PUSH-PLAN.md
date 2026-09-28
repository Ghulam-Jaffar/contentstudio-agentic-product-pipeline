# Helpin push spec

How pipeline deliverables get created in Helpin. Applies to **newly authored work only**.

## Scope

**Only work created from 2026-09-23 onward is pushed.** The existing backlog — the ~238 folders under `docs/features/` and `docs/stories/`, including the 7 Aug 2026 batch this document originally specced — is **not** being backfilled. PO decision, 2026-09-23. Those folders stay as local markdown records.

If the PO ever wants a specific old folder pushed, they will say so explicitly. Do not offer, and do not treat an old folder as pending work.

> The original 7 Aug 2026 batch plan (13 epics / 49 stories, with its epic tables and `[Design]` assignments) is preserved in git history at commit `7beb5e5` if it is ever needed.

## Approval

**Every push requires the PO's explicit approval in the moment.** This is not the same as approving the markdown.

- Approving `02-stories.md` or `04-epic-and-stories.md` approves **the authoring**. It is not consent to create anything in Helpin.
- Before any create call, show what will be created — titles, epic, target space, state, sprint, assignee, priority, labels — and wait for a clear yes.
- No standing approval. A yes covers the batch in front of the PO, not the next one.
- If the PO has not specified a field, **ask**. Never default one.

## Connection

- Server: `https://mcp.helpin.ai/mcp`, configured in the repo's `.mcp.json` (project scope, shared via git).
- Remote MCP with OAuth. Runs as the signed-in user, with that user's role and permissions.
- **Authenticated 2026-09-23.** `claude mcp get helpin` reports `OK Connected`.
- **Granted scopes:** `helpin.context.read`, `helpin.pm.read`, `helpin.pm.write`, `helpin.docs.read`, `helpin.docs.write`, `helpin.agents.read`, `helpin.agents.run`.
- **Not granted:** `helpin.crm.read`, `helpin.crm.write`, `helpin.support.read` — the server offers them; this pipeline has no business touching CRM.
- `pm.write` and `docs.write` are both present, so epic/story and Doc creation are permitted. If a create still fails, the cause is Helpin's in-app MCP permissions (Manage permissions), not the OAuth grant.
- Access tokens are short-lived (~15 min) with a refresh token; renewal is automatic.

## What the PO supplies per push

Never inferred, never defaulted:

| Field | Notes |
|---|---|
| Target space | e.g. Engineering |
| Sprint | Current or next — the PO says which |
| Workflow state | May differ between an epic and a story |
| Priority | |
| Labels | |
| Assignee | Per story, or a rule such as all `[Design]` stories to one person |
| Linked docs / linked stories | Set as real Helpin links where the API supports them, in addition to the authored text references |

## Resolved from the live workspace (2026-09-23)

Workspace **ContentStudio** `fb464f68-0c05-4cbd-b571-5254bc88211f`. Connected as Ghulam Jaffar, role `admin`, `read_only: false`, toolsets `context, pm, docs`.

| Thing | Value |
|---|---|
| Engineering **team** (PM work) | `f6b7c16a-8278-48f8-9d31-ffa89ee2cb87` |
| Engineering **Docs space** | `bc4df642-96cf-44b0-b345-7a3e5c4ac8b4` |
| Product Specs Docs space (system) | `f5607cef-4577-46de-85eb-cb2cd435be41` |
| Engineering workflow | `924473ee-2042-463a-8b5e-b47b9778397f` |
| **Ready** state | `32eac130-913d-423c-a08f-e522e2768be4` — `is_default: true`, so omitting `state_id` already lands in Ready |
| Current sprint | `a14040be-79fb-4d40-9340-54c2cd721f54` = **"Sep 21 - Oct 04 - 2026"** |
| Fasih Shaukat | member `835a1bed-679d-48f5-8048-b003ac0ea79d` (Engineering) |

**Teams and spaces are different axes.** PM objects (epics, tasks) belong to a **team**; Docs belong to a **space**. The push spec's earlier "target space" wording conflated them.

### Answers to the original unknowns

1. **Target** — Engineering *team* for epics/stories, Engineering *space* for Docs. Resolved.
2. **States** — `Ready` is the workflow default. Epics use a **separate** `epic_state_id` field; that set has not been enumerated yet.
3. **Sprint** — current sprint resolved. **There is no `list_sprints` tool and `search_workspace` returns nothing for `entity_type: sprint`.** Sprint IDs are only discoverable by reading `sprint_id` / `sprint_name` off an existing task. The *next* sprint cannot be resolved until a task is filed against it — ask the PO.
4. **Assignees** — resolved via `search_workspace` with `entity_types: ["workspace_member"]`.
5. **Rendering** — see the warning below. Docs support Mermaid; tasks are a different story.
6. **Epic-to-story linking** — `create_task` accepts `epic_id` **at creation**. No second call needed.
7. **Doc attachment** — `link_document_to_object` with `linked_object_type: "epic"`, `link_context: "attached"`. Confirmed.

### Still open

- **Epic state IDs** (`epic_state_id`) — not yet enumerated.

### Resolved 2026-09-28

- **Sprints are now listable.** `list_sprints` (with `team_id` and `status: "unstarted"`) returns upcoming sprints with names, dates and IDs. Use it to resolve "next sprint" instead of harvesting IDs off tasks. On 2026-09-28 the Engineering sprints after the current one were `Oct 05 - Oct 18 - 2026` (`8ae1115d-b22b-45b2-9bcb-ad0ecf77855d`) and `Week 43-44, 2026` (`82405ad6-bad4-4431-a47e-8bb4da942392`).
- **Labels are now listable.** `list_pm_labels` returns the full set. Besides the table below there is also a `design` label (`09404b9a-2b57-4170-8c6a-2ca740a10bf0`), distinct from `UIUX`, plus `infrastructure`, `admin-portal`, `Changelog` and the `Bug Reason - …` family.

## Reference data (fetched 2026-09-23)

### Workflow states — Engineering workflow `924473ee-2042-463a-8b5e-b47b9778397f`

| State | ID | Note |
|---|---|---|
| Ready | `32eac130-913d-423c-a08f-e522e2768be4` | `is_default: true`, `state_type: unstarted` |
| In Progress | `bdf0b9a5-8e7e-4c43-95f8-7bf78c62e7d1` | |
| In Review | `11b7eee2-a95d-4554-925c-88b01e2355d2` | |
| In Validation | `87de1af5-23db-4563-95e3-91cb61b8ffce` | |
| Ready to Ship | `0a6cc1b6-4c93-4fd8-9cf5-d588278329e3` | |
| Done | `d92d0375-20f2-4e9e-aeee-41482e02fdcb` | |

Omitting `state_id` lands in **Ready**, which is what the pipeline wants anyway.

### Labels (team-scoped to Engineering)

| Label | ID |
|---|---|
| frontend | `038efaf4-923a-4098-9f9e-770dca6116a9` |
| backend | `fc4d6be8-c95f-40d3-914b-ab30d946dda3` |
| ai | `6530eafe-bbbd-47aa-bb4c-6139247056c0` |
| UIUX | `e2d1df4a-463d-4ffe-9a01-50d2bf843b79` |
| sentry-issue | `114f9b27-baa4-4602-8cdb-5ac7d5baa4be` |
| Marketing team feedback | `dcbcd896-c2cf-4a59-aa52-038a8c3b3ffc` |

Harvested from live tasks — there is no `list_labels` tool, so this table may be incomplete.

### Sprints seen on live tasks

| Sprint ID | Name | Open tasks |
|---|---|---|
| `a14040be-79fb-4d40-9340-54c2cd721f54` | **Sep 21 - Oct 04 - 2026** (current) | 63 |
| `337d6fa3-be1b-4ab2-8c9e-d2ed54acd91e` | — | 19 |
| `5bc631f7-1e67-4d36-9c23-1661657e3b65` | — | 7 |
| `97627f84-5cb7-4db1-bdb2-09c27a0a7dec` | — | 4 |

`sprint_name` only comes back from `get_task`, not from `list_tasks`. There is no sprint list tool.

### Enums

- `priority`: `none`, `low`, `medium`, `high`, `urgent`
- `task_type`: `feature`, `bug`, `chore`
- `severity`: `none`, `minor`, `major`, `critical`

## Test run results (2026-09-23)

A full epic + story + doc + link was created and read back. Findings:

**Epic descriptions are NOT converted. Send HTML.** (Found on the first real push, 2026-09-24.) `create_epic` and `update_epic` store `description` verbatim, so markdown shows up in the epic Overview as one run-on paragraph full of literal `###`, `**` and `- `. Tasks and Docs convert markdown; epics don't. Always convert the epic description first with `agents/scripts/md-to-helpin-html.py <file.md>` and send the HTML it prints (`<p>`, `<h3>`, `<ul>/<ol>` with `<li><p>`, `<strong>`, `<code>`, which is the same shape Helpin's own task conversion produces). If an epic already went up as markdown, fix it with `update_epic` using the converted HTML.

**Helpin auto-labels some tasks.** Even with no `label_ids` sent, some `[Flutter]` stories came back with the `frontend` label (5 of 25 on 2026-09-24, with no clear pattern). After a push, check every created task's `labels`. If the PO asked for no labels, or for different ones, clear or set them with `update_task` and `label_ids`: `[]` clears.

**`create_task_batch` can't set sprint or assignee.** It takes no `sprint_id` or `owner_member_ids`. When the PO gives either, create stories one at a time with `create_task`.

**A task cannot link to another team's epic.** (Found 2026-09-25.) `create_task` on the Marketing team with an Engineering `epic_id` fails with a generic "could not complete the request" error, and nothing is created. Create the task without `epic_id`, and name the epic in its Dependencies.

**`update_task` does NOT convert markdown.** (Found 2026-09-25, quarterly billing push.) Unlike `create_task`, `update_task` stores `description` verbatim, so a markdown body sent on update comes back as raw `###` / `- [ ]` text. When editing a story's body after creation, send HTML in the same shape `create_task` produces (`<h3>`, `<p>`, `<ul>/<ol>`, `<table>`, `<hr>`, `<code>`, checkboxes stripped to plain `<li>`). Python-markdown with the `tables` extension works if nested indents are widened to 4 spaces and lists and tables get a blank line before and after.

**Markdown IS auto-converted on `create_task`.** The earlier HTML warning was a false alarm — sending markdown produces correct `<h3>`, `<ol>`, `<table>`, `<strong>`, `<code>`. Author story bodies as plain markdown, exactly as the template already does. Do **not** hand-write HTML.

**Markdown checkboxes are silently stripped — in tasks *and* docs.** `- [ ] Mobile responsiveness` becomes a plain `<li>Mobile responsiveness</li>`. The box is gone, so a dev has nothing to tick. This is why the already-pushed `[BE] Add repeat posts to the public API` shows its quality checklist as a plain bullet list.

**PO decision, 2026-09-23: leave it as body content anyway.** Helpin's native checklist (`checklist_items` on `create_task`, or `create_task_checklist_item`) was tested and works — items come back `completed: false` with real positions. It was **rejected**: the quality checklist is part of what the story *says*, not a set of sub-tasks to track on the board. So the 6 items stay as markdown in the description, rendering as a plain bullet list. That is the accepted outcome — **do not "fix" it by converting to checklist items.**

**Docs preserve Mermaid.** A ```mermaid fence survives as a `codeBlock` and Helpin Docs renders it. Tables survive as real `table` blocks. So Workflow diagrams are safe in a Doc; whether a task description renders Mermaid is still unverified.

**Epics do not require a state.** `epic_state_id` came back `null` on creation and that is accepted.

**Every mutation requires an `idempotency_key`** (8-128 chars), so a retry cannot double-create.

### Doc routing rule (PO decision, 2026-09-23)

**Route by document type, not product area.**

| Pipeline file | Collection |
|---|---|
| `01-research.md` | **Research** `e6181c6d-9a98-44ce-8afb-a7fe5e9d1199` |
| `02-workflow.md` | **PRDs & Feature Specs** `479ec25f-efbf-4623-804d-188886465255` |
| `03-prd.md` | **PRDs & Feature Specs** `479ec25f-efbf-4623-804d-188886465255` |
| Competitor research | **Competitor Analysis** `5e3ced6a-05ff-455e-a911-499191530502` |

The product-area collections (Analytics, Social Inbox, Billing, Infra, Mobile Application, SOPs, CS - AI Teammembers/Agent) are **curated by humans — the pipeline does not file into them.** A PRD is always findable in one place regardless of which area it covers.

**Never leave `collection_id` unset.** That is what "Uncategorized" is, and it already holds 500 documents.

## Nothing is ever deleted — archive is the only removal

**Helpin has no delete, at either layer.**

- **Over MCP:** no delete *or* archive tool exists for tasks, epics, documents or checklist items. `update_task` updates fields *"without deleting, archiving, or moving it between teams."* `create_epic` states *"Agent auto-run and archive controls are excluded."*
- **In the UI:** the only option a human gets is **Archive**, confirmed by the PO. There is no delete there either.

**Consequences:**

1. **The pipeline cannot remove anything from Helpin** — not even something it created seconds earlier. A bad push can add noise but can never destroy existing work.
2. **A mistaken push is not undoable, only archivable.** Archived clutter is permanent clutter. This is the real weight behind the approval gate: there is no clean way back.
3. **Cleanup is always a manual UI job, and it is archiving, not deleting.** Never promise to delete or archive something from Helpin — hand over the link and let the PO archive it.
4. Because retries are guarded by `idempotency_key` and nothing can be removed, **the safe failure mode is to stop and report, never to retry.**

### Docs collections — Engineering space `bc4df642-96cf-44b0-b345-7a3e5c4ac8b4`

**Every pushed doc must be filed into a collection.** `create_document` takes `collection_id`; omitting it dumps the doc into "Uncategorized", which is how that bucket reached 500 documents. If a doc is created without one, `move_document` can file it afterwards.

Top level:

| Collection | ID |
|---|---|
| PRDs & Feature Specs | `479ec25f-efbf-4623-804d-188886465255` |
| Research | `e6181c6d-9a98-44ce-8afb-a7fe5e9d1199` |
| Competitor Analysis | `5e3ced6a-05ff-455e-a911-499191530502` |
| Mobile Application | `825de60c-f41f-4c9a-b3bb-0718593ba5d2` |
| Analytics | `16713121-cf0a-4b28-8c69-b027ef819f8f` |
| Social Inbox | `5980d48f-db81-4cf1-b87d-f6889ff78114` |
| Billing | `df7207f6-f995-4cea-b3fd-cea6af34f25f` |
| Infra | `a99406b1-8e12-4078-823e-bf2911131a30` |
| SOPs | `5f4da566-fe99-4052-b56e-5bf855108ab2` |
| CS - AI Teammembers/Agent | `451286ad-5c55-4937-8422-a91eec8ce8a9` |

Nested (not visible in the sidebar at a glance):

| Sub-collection | Parent | ID |
|---|---|---|
| AI Studio | PRDs & Feature Specs | `760d6cca-4557-48d8-a4c0-f57a65a8febc` |
| Hootsuite | Competitor Analysis | `f5588eca-7127-450d-a2e3-08a77d16bab9` |
| Vista Social | Competitor Analysis | `765af748-a4ec-4c69-9b12-1fff74f75c00` |
| Onboarding | Competitor Analysis | `03a379c1-f908-46cc-9097-899d54e95715` |
| Discover module value addition & UI/UX | Competitor Analysis | `e02a4ca8-b60b-4204-bc9c-77ae3991031c` |
| AI Tools, Agents & Skills - Chat | CS - AI Teammembers/Agent | `5e64e995-3dbe-4be9-9222-5bff85c960a9` |
| Content creator tool | CS - AI Teammembers/Agent | `15858de8-f004-4bdb-b999-ce0ed6ced76f` |
| other tools | CS - AI Teammembers/Agent | `a1a09281-8c80-4286-997d-803f1ce680b5` |
| X (Twitter) - pay as you go model | Billing | `d0d61c88-67e6-4a6b-881f-d763a253337b` |

"Uncategorized" is not a real collection — it is simply `collection_id: null`.

### Web URL patterns (confirmed by the PO, 2026-09-23)

Base: `https://app.helpin.ai/w/contentstudio`

| Target | URL |
|---|---|
| Epic | `<base>/pm/epics/<epic_id>` |
| Story, opened on its epic | `<base>/pm/epics/<epic_id>?task=<TASK_KEY>` |

**The `/pm/` segment is required.** The `helpin://tasks/<id>` URI the API returns is *not* the web path — deriving the URL from it produces a dead link. PM objects live under `/pm/`, and a story is addressed by its **task key** (`CONT-4048`), not its UUID.

Not yet confirmed: the Docs URL, and how to link a story that has no epic. Ask the PO rather than guessing.

When writing `0N-helpin-links.md`, use these patterns so every link is clickable.

### Test objects created — delete when done

| Object | ID / key |
|---|---|
| Epic | `f135d3f1-32f1-4243-a17c-3f59309e01e9` |
| Story | **CONT-4048** `81cdd5b2-e5fa-4217-8792-5ef28bffc188` |
| Doc | `2ed8a411-8d5d-42e1-858a-bc23e20a6d98` (status `draft`) |
| Doc-to-epic link | `c2b7a368-e122-4cdc-9831-1307e5a43711` |

## Mapping

| Pipeline deliverable | Helpin object |
|---|---|
| `01-research.md`, `02-workflow.md`, `03-prd.md` | **Docs**, linked to the epic |
| The authored epic | **Epic** |
| Every authored story, including `[Research]` ones | **Story** |
| Returned URLs | written back to `<slug>/0N-helpin-links.md` |

## Execution order

1. Read the tool schemas. Map Helpin's model onto epic, story and doc before creating anything.
2. Resolve the seven unknowns above.
3. **Dry run on the smallest unit** — 1 epic + 1 story + 1 doc. Stop. Have the PO check it in Helpin, specifically that the description renders correctly.
4. **Then a story set containing Mermaid**, to prove diagram rendering. Stop again if anything looks wrong.
5. Then the rest, epic first, then its stories, then its docs.
6. Write returned URLs back to `<slug>/0N-helpin-links.md` so local docs and Helpin stay cross-referenced.

## Rules for the push

- **Never invent an identifier.** If a state, sprint, space or user can't be resolved, stop and ask.
- **Story bodies go in as authored.** Do not re-summarize to fit a field. The sections (Description, Workflow, Acceptance criteria, Mock-ups, Impact on existing data, Impact on other products, Dependencies, Global quality checklist) are the deliverable.
- **The quality checklist is body content, not native checklist items.** It renders as a plain bullet list because Helpin strips `- [ ]`. Intended.
- **No estimates.** The team sets those during sprint planning.
- **No metadata block in the body.** Sprint, state, labels, assignee and priority go in as Helpin fields, never as text.
- **Dependencies are expressed by full story title**, as authored. If Helpin supports real dependency links, add them *in addition to* the text, not instead of it.
- **A story that came from Frill carries its Frill idea link**, so a changelog can find it later.
- If a create fails partway through, **record what was created and stop.** Do not retry blindly and risk duplicates.
