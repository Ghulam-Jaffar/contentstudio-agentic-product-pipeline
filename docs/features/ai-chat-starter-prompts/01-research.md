# Research: AI chat feature pills and starter prompts

**Date:** 2026-09-14
**Pattern reference:** an established competitor pattern, feature pills that open a list of prompts
**Deliverable:** prototype, then stories

---

## 1. The ask

Two changes to the AI chat new-chat screen:

1. **Make the hero shared.** The centered greeting, chat box and prompt row currently appear in AI Studio only. The modal should get the same treatment.
2. **Replace the three static chips with feature pills.** Clicking a pill reveals a list of prompts for that area. Clicking a prompt **sends it** immediately, the hero collapses and the pills disappear, which is the existing flow. A close control on the list returns to the pills.

---

## 2. Current state

- The hero lives in `contentstudio-frontend/src/modules/AI-tools/ChatBox.vue`, gated by `isHeroState`, which requires `props.type === 'studio'`, zero messages, no loader and no active tool.
- **There are only two chat surfaces, not three.** `AIChatModal.vue` passes `type="modal"`; `AiStudioChatView.vue` passes `type="studio"`. `AIChatWidget.vue` is a 95-line floating launcher button that opens the modal, not a surface of its own.
- Only **two** studio-only gates exist in the file: `isHeroState` at line 711, and an unrelated dashboard composer hand-off at line 1346 that must stay studio-only.
- The three chips are a hardcoded computed, `heroChips`, with copy at `dashboard.ai_assistant.suggestion_chips.*`. Clicking one **fills the input** rather than sending.
- Chip styling uses hardcoded hex values, so it is not theme-aware on white-label domains.
- The hero title reads "Plan, create & schedule smarter with AI Studio", which would be wrong in the modal.
- A commented-out "Quick prompts and Favorite Prompts" block remains in the file from an earlier iteration.
- Separately, `Prompts.ts` holds 559 lines of text and image quick prompts used by the tools surface, plus a saved-prompts modal. Architecture notes say to build on that surface rather than beside it.

---

## 3. What the assistant can actually do

This is the constraint that decides the prompt list. A prompt that maps to no tool produces a failure, and the coordinator will not always route it to the one agent that knows how to say "not supported".

The chat is one coordinator team with **11 member agents** and six ContentStudio toolkits plus asset and web-research tools.

### Covered, and safe to build prompts on

| Area | Tools |
|---|---|
| **Planner and calendar, read** | `posts_list` with status, period, date range, label, campaign, content category, author and approval filters |
| **Publishing, write** | `posts_create` (scheduled, draft, queued, content category), `posts_update`, `posts_delete`, `posts_create_from_plan` |
| **Approvals** | `posts_approval`, `workspace_list_approval_workflows`, plus approval filters on `posts_list` |
| **Post plans** | `plan_posts`, `generate_planned_posts` |
| **Analytics** | `analytics_fetch_metrics` with period comparison, `analytics_compare_accounts`, `analytics_top_content`, `analytics_insights`, `analytics_list_available_metrics` |
| **Best time to post** | `analytics_best_times`, **Facebook and Instagram only** |
| **Inbox** | `inbox_summary`, `inbox_list_elements`, `inbox_find`, `inbox_get_thread`, `inbox_update_elements`, `inbox_tags`, `inbox_add_note`, `inbox_reply`, `inbox_moderate_comment` |
| **Image and video generation** | `generate_image`, `generate_image_set`, `transform_image`, `generate_video`, `generate_alt_text`, `enhance_image_prompt`, `analyze_image`, `describe_image` |
| **Media library** | `media_list` read, `media_import` from a URL |
| **Workspace and team** | account list with token state, team members, groupings, connectable platforms, workspace create and update, member invite and remove, `account_connect_start` |
| **Open web research** | Exa answer, contents, find similar |

### Not covered, and therefore forbidden as prompts

This list matters more than the one above.

| Area | Status |
|---|---|
| **Competitor analytics** | No tool at all |
| **Saved, scheduled or PDF reports** | No tool at all |
| **Ads analytics, Meta and Google** | Agents exist, zero chat reach |
| **Content Discovery, curated feeds, trending content** | Entire module unreachable from chat. Open-web research is the only substitute |
| **Social Listening, keyword monitoring** | HTTP service only, zero chat tools |
| **Automations: evergreen, RSS, bulk CSV** | Services exist, not registered on any member |
| **Brand knowledge** | Injected as context, no read or write tool. A user cannot inspect or edit it from chat |
| **Content category creation** | Explicitly impossible, no API endpoint |
| **Billing, plans, credit balance** | No tool |
| **Tasks or to-dos** | No such concept |
| **User-configurable agents** | Documented as a future phase, not built |
| **Media upload from chat, folders, delete** | Import from URL only |
| **Team role change, inbox assignment, workspace switching, Bluesky connect** | Deliberate refusals |
| **The eight dedicated image tools** (remove background, upscale, headshot, face swap, outfit swap, product image, motion control, lip sync) | Real HTTP routes, **not chat tools** |

### Live defects that shape prompt wording

- **Carousels are broken.** The create payload carries no platform option block, so the call fails **after** the user approves it. No carousel prompt.
- **Relative weekday resolution is wrong.** Nothing resolves phrases like "next Tuesday" on the publishing path, and a real case resolved to the wrong week. Avoid relative-weekday scheduling prompts.
- **A turn producing two assistant texts keeps only the last**, so compound "write captions and schedule them" prompts can lose the captions.
- Writes are human-in-the-loop gated: 19 tools require explicit confirmation. Prompt copy must not imply anything happens instantly.

### Skills are a spec, not code

The 19-skill catalog in `docs/features/ai-chat-skills/default-skills-catalog.md` is unbuilt. `src/skills/` does not exist. So pills must send **natural language prompts** that route through tools, not slash commands. The catalog is still the best guide to what a good prompt set looks like, and its grouping maps almost exactly onto the proposed pills.

Note its rule 5: nothing publishes, sends, deletes or changes anything without explicit confirmation.

---

## 4. Decisions taken during prototyping

An interactive prototype was built and reviewed. What it settled:

1. **Six pills:** Post, Write, Media, Inbox, Reporting, Workspace. Twenty-nine prompts across them.
2. **The panel floats, it does not push.** Anchored under the composer at the full width of it, not to the individual pill. This is the decision that makes one behaviour work on all three surfaces: the home widget has page content directly beneath the composer, so a panel that pushes would shove the page down on every click. A floating panel covers it temporarily, which is what any dropdown does.
3. **The pills stay visible** while a list is open, so a user can move between sections without closing first. The close control still closes the list, and tapping outside does too.
4. **Phone width gets a bottom sheet** rather than the anchored panel. Anchored overlays fight positioning, dismissal and the keyboard on a phone.
5. **The panel has a height ceiling and the list scrolls inside it.** Without that, adding one more prompt to a section silently breaks the layout on the shortest surface.
6. **Clicking a prompt sends it**, rather than filling the input as today's chips do.
7. **Gating is plan feature and role permission only.** Connected accounts are deliberately **not** a gate. Every prompt shows regardless, and the assistant explains what it can do from there. It already derives a capability list from its live tool registry, so it answers "nothing is connected yet" more usefully than a missing prompt would, and a new workspace still sees what the product can do.
8. **The greeting becomes the modal's**, being a welcome plus "How may I help you today?". It is the only one of the three that does not name AI Studio, which reads wrong in the modal and on home.
9. **The modal's three action buttons are replaced**, since the composer moves into that space.
10. **Where the pill and prompt catalogue lives is an implementation decision** for the developer, not something this spec prescribes.
11. **Mobile is not in this batch**, but has a story so it is tracked rather than forgotten.

### The approver case, which is the best argument for gating

A default approver cannot create posts, cannot reach the inbox, and cannot reach analytics. Gated, they get a **Post pill with one prompt**, "What's waiting for my approval?", which is exactly their job. Ungated they would see six pills of things they cannot do.

Note that `approverCanCreatePost` is a **per-member opt-in**, not a property of the role, so the check has to read the permission rather than the role name. An approver whose admin granted creation rights should see the create prompt.
