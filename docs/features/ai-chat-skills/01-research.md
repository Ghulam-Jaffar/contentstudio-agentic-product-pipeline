# 01 — Research: Skills in AI Chat

**Feature:** Skills — reusable, user-editable instruction sets attached to an AI chat prompt via slash command.
**Date:** 2026-09-07
**Sources:** Competitor/industry research (WebSearch) + codebase analysis (contentstudio-frontend, contentstudio-backend, contentstudio-ai-agents, contentstudio-flutter).

---

## Part 0 — What we are building (and what we are not)

A **skill** is a named, reusable instruction set — `name` + `description` + a markdown `instructions` body — that a user attaches to an AI chat prompt so the assistant performs a specific task the same way every time.

**This is not the CLI/MCP agent-skills work** already in the backlog (`docs/features/contentstudio-public-cli-agent-skills/`, `docs/stories/clawhub-skill-publish/`). Same word, different concept, different audience. That work packages skills for external agents consuming the public API; this work is an in-product, frontend, prompt-attached feature for ContentStudio users in AI chat. The naming collision needs an explicit internal call — see Open Questions.

**Why users want it** (validated across the research):

1. **Prompt fatigue.** Output quality in every AI writing assistant is a function of prompt quality — every vendor's help doc says "provide a detailed prompt." Skills move that burden from per-message to once.
2. **Consistency.** A skill is a determinism device: same method, same output shape, every time.
3. **Expertise transfer.** A skill encodes what the team's best operator knows so a junior operator gets the same result.
4. **Capability discovery.** Most users don't know what an AI chat assistant can do. A `/` menu of ~18 named skills is the fastest capability tour a chat UI can offer — it turns a blank text box into a menu.

---

# PART A — Competitor & Industry Research

## A1. Social media management tools

| Competitor | Has Feature? | Key Capabilities | Pricing Tier | UX Approach | Unique Differentiator |
|---|---|---|---|---|---|
| **Vista Social ("Ask Vista")** | **Yes — near-exact match.** Literally called "Skills." | 20+ built-in skills pre-wired to accounts. Users author custom skills "in plain English, no code." Skills are "written in plain, readable text you can open and edit." Custom skills importable from external sources. Skills inherit workspace brand voice, profiles, analytics, scheduling data. Skills can also be attached to scheduled Agents. | Bundled with Ask Vista; no tier gate found publicly — **uncertain**. | **Slash commands in the chat box** — `/hook-maker`, `/triage-inbox`, `/post-postmortem`, `/best-time-to-post`, `/generate-post-ideas`. Custom skills appear as "saved, editable documents." | The skill↔agent pairing: *"An agent without a skill is a tireless generalist, and a skill without an agent is an expert you have to summon."* A skill fires manually **or** bolts onto a scheduled agent. |
| **Buffer (AI Assistant)** | **No.** | Generate from prompt, repurpose, adjust tone, expand/shorten, translate. Brand voice guidelines. | All plans incl. free, no usage limits. | In-composer prompt box + quick-action buttons. No library, no slash menu. | Free/unlimited AI is the differentiator, not customization. Buffer punts customization to external AI tools via its API. |
| **Hootsuite (OwlyWriter / OwlyGPT)** | **No.** Prompt *suggestions*, not saved skills. | Caption generation by tone/channel, repurposing top posts. Brand voice trained on 10–15 example posts + past content + Talkwalker mention data. Writing frameworks (AIDA, WIIFM). | Professional and above. | Pre-built query suggestions — a starter-prompt tray. One-way, not editable or savable. | Brand voice grounded in *external* signal (brand mentions), not just own posts. |
| **Sprout Social (AI Assist / Trellis)** | **Partial / adjacent.** Not user-authored skills in chat. | AI Assist: brand voice dropdown + free-text prompts. Trellis (agentic, expanded May 2026) spans Publishing, Listening, Smart Inbox, Reporting. **Trellis Studio** = environment to build custom AI workflows. | AI Assist on paid plans; **Trellis is rep-gated / enterprise**. | Trellis Studio is a separate builder, not an in-chat slash command. | **Their answer to reusable prompts is editorial, not product**: a 135+ prompt library published as marketing content, and a best-practice doc telling customers to *"build a shared internal playbook that captures your best prompts"* — i.e. maintain it in a doc, outside the tool. A direct, quotable admission of the gap. |
| **Publer (AI Assist)** | **Partial — closest non-Vista competitor.** "Brand Voices" ≈ a persona, not a task skill. | Named, tailored AI personalities per client/workspace/account. Each customizable — name it, feed it prompts, upload brand guidelines or sample content. Multiple voices, switchable. Plus saved AI chats. | Paid add-on / credit model (**uncertain**). | Named entity + switcher when generating. No slash command. | The name + instructions + switchable shape is structurally very close to a skill — scoped to *voice* rather than *task*. |
| **Later (Caption Writer)** | **No.** | 3 caption options from a prompt; auto-learns tone from previously scheduled posts. | Paid plans, credit-metered. Web + iOS + Android. | Single prompt box. | Implicit brand voice learned from post history — zero configuration, no artifact for the user to manage. |
| **Agorapulse** | **No.** | Create from scratch or optimize existing copy via **quick actions** or custom prompts. Brand tone/goals/audience set once, applied everywhere. | Bundled with paid plans. | Quick actions = fixed vendor-authored button set. Closest thing to a skill menu, but not user-extensible. | A single global custom-instruction applied across surfaces. |
| **SocialBee (Copilot)** | **No user-authored skills.** | Scans your website and auto-extracts business name, tagline, services, tone, audience, language. **1,000+ AI prompts** by industry/goal/post type, fill-in-the-blank + tone picker. | Paid plans. | A large vendor-curated prompt catalog. Users can edit a prompt per-generation, but not save to a library. | Auto-onboarding (brand profile inferred from website). Also the largest vendor prompt catalog — which is itself a discovery problem. |
| **Loomly** | **No.** Weakest of the set. | "AI detects brand voice and learns from past performance." | Uncertain. | Nothing surfaced. | None found. |
| **Sendible** | **No evidence.** | AI writing exists; nothing on saved prompts or skills surfaced. | Uncertain. | Not documented. | None found. |
| **Metricool** | **No.** | Platform-specific generation, hashtag suggestions, "prompt customization so the writing style better fits your brand," text-transform shortcuts. | **5–35 AI credits/month per brand** — very constrained. | Shortcut buttons for text transforms. | Nothing reusable; the credit ceiling actively discourages a skills-style workflow. |

**Bottom line:** exactly one competitor — **Vista Social** — has shipped this. Everyone else has either a *brand voice* object (a persona, applied globally) or a *vendor prompt catalog* (marketing content or fill-in templates). **Nobody except Vista Social has user-authored, named, slash-invoked, per-message task skills.**

## A2. Adjacent AI products (more mature prior art)

| Product | Has Feature? | Key Capabilities | UX Approach | Unique Differentiator |
|---|---|---|---|---|
| **Anthropic — Claude Skills** | **Yes — the reference implementation.** | `SKILL.md` with YAML frontmatter: `name` (≤64 chars), `description` (≤200 chars), then a markdown body; optional reference files and scripts. Built-in skills shipped (pptx, xlsx, docx, pdf). Creation: describe it and Claude builds it, write directly, or upload a folder. | **Customize → Skills**, a list with a **toggle switch per skill**. When iterating in chat, "the skill files open beside the conversation" for direct editing. | **Progressive disclosure** — only `name` + `description` sit in context; the body loads on selection. Plus the **admin-provisioned + user-override** model: org-pushed skills enabled by default, but individual users can toggle any of them off. |
| **OpenAI — Custom GPTs** | Yes (2023-era answer). | System prompt + knowledge files + tools, packaged and shareable. | You **switch into** a GPT — it replaces your chat context. Distribution via private/link/GPT Store. | Distribution/marketplace was the bet. It failed on discovery. |
| **OpenAI — Custom Instructions** | Yes (global). | Persistent facts applied to every chat. | One settings text box, always on. | Zero-friction, zero-choice baseline. |
| **OpenAI — Projects** | Yes (scoped). | Per-project system prompt + files + project-only memory. | A container you work *inside*. | Scoping by workspace rather than by task. |
| **OpenAI — ChatGPT Skills** | **Yes — newly shipped; OpenAI has adopted the term.** | "Reusable workflows telling ChatGPT how to complete one specific task using the same method every time." Editable once installed. Portable plain-text skills. | Sidebar → Plugins → Skills. "ChatGPT can pull one in automatically when a task matches" — auto-routing. | The cleanest conceptual model found: **Skills = *how* work gets done. Projects = *where* it lives. Custom GPTs = *what* you share.** (Rollout moving fast — **treat specifics as uncertain**.) |
| **Notion — Custom Agents** (3.3) | Yes. | Configurable instructions, triggers, schedules, model selection, activity logs. Build once, share with workspace. | **Agent Directory** — browsable and searchable. Sharing to individuals, groups, or the whole workspace. Full Access permission gates who can edit. | The **directory as a first-class discovery surface**, plus per-run activity logs ("every run is logged so changes are visible and reversible"). Sharing is a real permission model, not a boolean. |
| **Cursor — Rules** | Yes — **best invocation model in the market**. | `.cursor/rules/*.mdc` with frontmatter controlling activation. | **Four explicit activation modes**: Always Apply / Auto Attached (glob) / Agent Requested (model decides from `description`) / Manual (`@rule-name`). | They refused to pick auto-vs-manual and made it a per-rule property. Community guidance: *"a healthy rules directory is mostly Auto Attached and Agent Requested, with a small handful of Always rules on top."* |
| **GitHub Copilot** | Yes — **the cleanest split**. | `.github/copilot-instructions.md` = always-on rules. `.github/prompts/*.prompt.md` = **reusable prompts invoked as `/command-name`**. | Two distinct primitives with a stated rule of thumb: *"Use a prompt file when you need a slash command for one repeated task. Use custom instructions when you want rules applied to all requests."* | The rationale for slash-invoked over always-on: prompt files *"can contain task-specific steps that would be noisy in a global instruction file."* |
| **Linear** | Yes (instructions, not skills). | Workspace-level + team-level guidance. **Team guidance takes priority.** | Settings-based, hierarchical, always-on. | Scope hierarchy with explicit precedence. |
| **Intercom — Fin** | Yes, across three generations. | Custom Answers (static) → **Guidance** (plain-language rules) → **Procedures** (multi-step natural-language automation). | Authored in plain language in settings. | **The migration path is the lesson**: when the flexible natural-language primitive arrived they **deprecated** the rigid one (Custom Answers explicitly *not supported* with Guidance; new Task creation disabled March 2026). They chose deprecation over maintaining both. |

## A3. Common patterns

1. **Name + description + markdown body.** Claude (`SKILL.md`), Copilot (`.prompt.md`), Cursor (`.mdc`), Vista, Publer. Nobody uses a structured form or a builder wizard — **free-form markdown won**. The `description` is universally load-bearing.
2. **Vendor ships defaults; users author more.** Claude built-ins, Vista 20+, SocialBee 1,000+, Sprout 135+. **An empty library is a dead feature.**
3. **Per-item enable/disable toggle.** Universal — Claude, ChatGPT, Notion. It exists specifically because always-on instructions become noise.
4. **Slash command is the dominant invocation UI** when invocation is explicit. Vista, Copilot, Notion. Cursor uses `@`. Settled convention.
5. **Two-tier scoping: personal vs. team/workspace.** Claude, Notion, Linear, Copilot. **Nobody ships a single flat visibility.**
6. **Plain English, no code — marketed as such.** Vista ("no code"), Intercom ("no technical expertise required"), Notion, Publer.
7. **"Ask the AI to build the skill" is table stakes**, not a delighter. Claude, Vista, Windsurf all have it.
8. **Skills are one layer, not the whole story.** Every mature product pairs a *task* primitive with a *persistent context* primitive: ChatGPT (Skills + Projects + Custom Instructions), Copilot (prompt files + custom instructions), Cursor (Manual + Always rules), Vista (Skills + brand voice). **Skills sit on top of an always-on brand voice; they never replace it.**

## A4. Differentiators worth stealing

1. **Cursor's per-rule activation modes** — the most important idea in this research. It sidesteps the auto-vs-manual argument by making it a property of the skill. Relevant to us as a **v2 path**, not v1.
2. **OpenAI's three-way split** — *Skills = how, Projects = where, Custom GPTs = what you share.* Steal as internal positioning discipline so Skills don't collide with brand voice.
3. **Claude's progressive disclosure** — only name + description in context; body loads on selection. What makes a large library technically viable.
4. **Claude's provisioned + user-override model** — shared skills enabled by default, individually toggle-off-able. Exactly the shape our "Everyone" visibility should take.
5. **Vista's skill↔agent composition** — designing the skill as a standalone object (not a chat-only artifact) preserves this option at zero cost.
6. **Notion's Agent Directory** — a browsable/searchable listing *separate* from the invocation surface. Slash menus are for **recall**; a directory is for **discovery**. Different jobs; Notion built both.
7. **Notion's per-run activity logs** — trust signal for shared skills.
8. **Intercom's deprecation discipline** — a warning against shipping Skills alongside three other overlapping prompt features.

## A5. User expectations

**Table stakes** (ship or the feature reads as broken):

| Expectation | Evidence |
|---|---|
| A useful set of defaults on day one | Vista 20+, Claude built-ins, SocialBee 1,000+. Empty libraries don't get adopted. |
| Slash-command invocation from the chat box | Vista, Copilot, Notion — settled convention. |
| Plain-markdown instructions, no code, no wizard | Universal. |
| Per-skill enable/disable toggle | Claude, ChatGPT, Notion — universal. |
| Edit a skill without breaking it, undoably | Claude opens skill files beside chat; Vista skills are "open and edit." |
| Clear indicator that a skill is active | A removable pill matches how every chat UI signals attached context. |
| Personal vs. team visibility | Claude, Notion, Linear, Copilot. |
| Search/filter in the picker | Command-palette convention; autocomplete typically kicks in after ~3 chars. |
| "Ask the AI to write the skill for me" | Claude, Vista, Windsurf. **No longer a differentiator.** |

**Delighters:**

| Delighter | Why |
|---|---|
| **Copy-on-write fork of a default** | Nobody in social documents it. Claude's closest equivalent (admin skill + user toggle-off) is not a fork. **Genuinely differentiating**, and the right answer to "the user edited the default and broke it." |
| Categorized/grouped slash menu with section headers | Search-UX guidance: header categories aid scanning. |
| Skill preview before firing | Builds trust in shared skills; nobody in social does it. |
| Usage counts / last-used surfacing | Directly attacks library bloat — the "trust signal" prompt-library literature asks for. |
| Import/export as markdown | Vista supports importing custom skills. Lets a workspace bootstrap from public libraries. |
| Skills that read workspace context automatically | Vista's claim that "all skills automatically incorporate the workspace's established brand voice" is their strongest capability claim. |

## A6. Documented failure modes

These are observed failures, not speculation.

1. **Routing failure — the skill that never fires.** In the Claude ecosystem, when total skill descriptions exceed a character budget (~8,000 chars against a ~15,000-char system-prompt list cap), descriptions are **silently truncated or dropped** and those skills are never considered. Second cause: vague descriptions — *"a skill with a bad description is invisible because the model can't match it to the right moment."* Diagnostic: *"If manual invocation works but automatic triggering doesn't, the description is your problem or you've hit the character budget."*
   > **Implication for us:** because our design is **explicit user selection via slash command, one skill active at a time**, we sidestep this entire failure class. That is a real architectural advantage. **Do not add auto-routing later without accounting for it.**
2. **Library bloat past ~20 items.** OpenAI: **3M+ custom GPTs created, only ~159K published — ~94% never published.** Store commentary: *"most GPTs in the public store are half-baked, built once and abandoned."* Prompt-library literature: *"the problem isn't that your prompt library doesn't exist — it's that it isn't searchable."*
3. **Overlapping primitives confusing users.** The single most-discussed ChatGPT complaint: users set Custom Instructions, build a GPT, add Project Instructions, and *"wonder which one is actually being used. They overlap, sometimes conflict, and most people end up using just one and ignoring the others."* **ContentStudio already has brand voice + saved prompts — this is our highest-probability failure.**
4. **Users editing defaults and losing them.** Not well-solved anywhere. A forked default that never receives vendor improvements silently rots.
5. **Model changes silently breaking saved skills.** Custom GPTs stopped working when the underlying model was swapped without warning.
6. **Duplicate/near-duplicate proliferation.** Shared prompts edited by multiple people cause *"silent regressions and edit conflicts"* without versioning.
7. **Ownership/trust ambiguity in a flat list.** Fix: *"separating different types of prompts (private, community, official, saved) keeps the library from becoming one flat list where ownership and trust are unclear."*
8. **Hidden features don't get used.** A slash-only entry point with no visible affordance is discovered by a small minority.
9. **Deprecation pain.** Intercom's three-generation path. Shipping a v1 skill format means committing to migrating it.

## A7. Naming verdict

**Call it "Skills."** Evidence is one-directional: Anthropic established the term (Agent Skills / `SKILL.md`), **OpenAI has adopted it** (ChatGPT Skills as a distinct surface alongside Projects and Custom GPTs), Microsoft shipped "Agent Skills in Visual Studio," and **Vista Social — our direct competitor — already uses it** with slash-command invocation.

**Rejected:**
- **"Agents"** — actively harmful. Across Vista, Notion and Sprout, "agent" now means *autonomous, scheduled, runs without you*. Vista draws the line explicitly: *"Skills are reusable one-time commands you can trigger manually, while agents run automatically on a recurring schedule."* Using it here sets the wrong expectation and burns a word we want later. **Reserve "Agents" for scheduled autonomous work.**
- **"Custom GPTs"** — vendor-locked and carries the taint of the abandoned GPT Store.
- **"Prompts" / "Prompt library"** — undersells; reads as a snippet you paste, not a procedure the assistant follows. Also collides with the marketing prompt libraries Sprout and SocialBee publish.
- **"Recipes" / "Playbooks"** — no market adoption. "Playbook" appears only in Sprout's advice to keep prompts in an *external doc*.

## A8. Competitive read

ContentStudio would be the **second** social media management platform with this feature, behind Vista Social by roughly one product cycle. Buffer, Hootsuite, Later, Agorapulse, Metricool, Loomly and Sendible have nothing comparable and are not signalling movement. Sprout's own documentation tells customers to keep their prompt playbook in an external document.

**Copy-on-write forking of defaults, plus a clean workspace-visibility model, would put us ahead of Vista's documented behaviour** — though their exact edit/visibility/fork mechanics are **not publicly documented (uncertain)** and are worth a hands-on trial before finalising.

**Marked uncertain:** Vista's pricing tier for Skills and their exact fork/visibility mechanics (support KB returned 403). ChatGPT Skills availability and invocation specifics are rolling out rapidly. Sendible and Loomly returned so little that "nothing found" is fair but not a confirmed absence.

---

# PART B — Codebase Analysis

## B1. Existing related code

### Frontend — AI chat (`contentstudio-frontend/`)

| Concern | Path |
|---|---|
| Renderless engine host | `src/modules/AI-tools/AIChatMain.vue` (74 lines — thin slot host over `createAiChatEngine`) |
| Chat engine (state + provides) | `src/modules/AI-tools/composables/useAiChatEngine.ts` (734) |
| Message list + send orchestration | `src/modules/AI-tools/ChatBox.vue` (1541) |
| **The actual chat input** | `src/components/dashboard/ChatInput.vue` (2432) — note: **not** in the AI-tools module |
| Compact toolbar (small screens) | `src/components/dashboard/ChatInputCompactToolbar.vue` |
| Settings dropdown / media pills | `src/components/dashboard/ChatInputSettingsDropdown.vue`, `ChatInputMediaModePills.vue` |
| TipTap editor wrapper | `src/modules/AI-tools/components/CstTextEditor.vue` (238) |
| @-mention TipTap extension | `src/modules/AI-tools/extensions/imageMention.ts` (245) |
| @-mention dropdown UI | `src/modules/AI-tools/components/ImageMentionDropdown.vue` (76) |
| Mention chips on user bubble | `src/modules/AI-tools/components/UserMessageMentions.vue` |
| Outgoing payload builder | `src/modules/AI-tools/utils/buildChatStreamPayload.ts` (223) |
| SSE send / stream | `src/modules/AI-tools/composables/useAIChatStream.ts` (927) |
| Markdown renderer | `src/modules/AI-tools/components/StreamingMarkdown.vue` (193) |
| Brand voice chip | `src/modules/AI-tools/components/BrandVoiceSelector.vue` (90) |
| Reference-file chips | `src/modules/AI-tools/components/ReferenceFilesAttachment.vue` (95) |
| App-level modal host | `src/components/AuthenticatedAppExtras.vue` (mounts `SavedPromptsModal` once app-wide) |
| API layer | `src/api/ai-chat.ts`; URLs in `src/config/api-utils.ts` |
| TanStack query keys | `src/modules/AI-tools/queries/keys.ts` |

### Backend (`contentstudio-backend/`)
- `app/Models/AiChat/AiCustomPrompts.php`, `app/Repository/AiChat/AiPromptRepo.php`
- `app/Http/Controllers/AI/AIController.php` — `fetchCustomPrompts` @305, `saveCustomPrompts` @316, `removeCustomPrompts` @329, `processAgnoAgent` @975
- `app/Http/Requests/AI/{GetAiCustomPromptsRequest,StoreAiCustomPromptsRequest,RemoveAiCustomPromptRequest}.php`
- `routes/web/ai.php` @27–29
- `database/seeders/AICustomPromptsSeeder.php` (1486 lines, ~16 sections)
- `app/Helpers/Ai/AiChatHelper.php` → `buildWorkflowInputs()` @339 (the PHP→Python payload)
- **Visibility pattern:** `app/Models/SocialPostTemplate.php`, `app/Repositories/SocialPostTemplateRepository.php`, `app/Http/Controllers/Planner/SocialTemplateController.php`, routes `routes/api.php` @127–134
- **Brand knowledge:** `app/Models/Ai/AiContentLibrary/AiContentLibraryProfile.php`, routes in `routes/web/ai.php` (`aiContentLibrary/profile/*`)
- **Workspace scoping:** `app/Http/Middleware/PermissionMiddleware.php`, `app/Support/RequestWorkspaceId.php`, `app/Services/WorkspaceAccess/*`, `config/workspace_access.php`

### AI agents (`contentstudio-ai-agents/`)
- `src/orchestration/agents.py` (~3000+ lines — all member agents; `assistant_agent` @2663, `content_ops_agent` @2776, `workspace_data_agent` @2854, Analytics Analyst @2966)
- `src/orchestration/{team.py, team_hooks.py, capabilities.py, guardrails.py}`
- `src/api/routers/streaming_router.py` — `StreamRequest` @613, `stream_generate` @2485, brand block @1267–1310, member-context substitution @1872/@1898
- `src/integrations/contentstudio/toolkits/{analytics,publishing,publishing_write,workspace,organisation_write}.py`
- `src/integrations/contentstudio/{bootstrap,session_state,scope,client,analytics_registry}.py`
- `src/orchestration/post_plan/` (planner, steps, workflow, state)
- `src/agents/tools/general_assistant.py` — **legacy standalone agent, NOT what the chat stream uses.** Do the Skills work in `agents.py` + `streaming_router.py`.

---

## B2. The existing saved-prompts feature (the migration target)

### Data model
MongoDB collection **`ai_custom_prompt`**, model `AiCustomPrompts`. Fillable:

```
prompt, title, section, section_id, is_default, workspace_id,
user_id, color_code, is_favourite, is_custom, prompt_id,
created_at, updated_at
```

FE type (`src/api/ai-chat.ts:63`):
```ts
interface AIPromptData {
  _id?: string; title: string; prompt: string; type: string
  is_favourite?: boolean; workspace_id?: string; [key: string]: unknown
}
```

**Two record classes share one collection:**

1. **System defaults** — `is_default: true`, **no `workspace_id`, no `user_id`** → they are **global**, not per-workspace. Grouped by `section` / `section_id` across 16 sections (Social media post caption, Instagram caption, Facebook, LinkedIn, Pinterest, GMB, Twitter ideas, Continue Writing, Extend/Expand, Improve, Rewrite, Shorten, Summarize, Paragraph, Inspirational, Pros and Cons List). Seeded by `AICustomPromptsSeeder.php` via bare `AiCustomPrompts::create()` — **no idempotency**, no per-workspace seeding.
2. **User custom prompts** — `is_custom: true`, scoped by `workspace_id` **AND** `user_id`. **There is no visibility field at all** — every custom prompt is effectively "only me," with no way to share one with the workspace.

### The favourite mechanism is already a partial copy-on-write fork
`AiPromptRepo::savePrompt()` + `findOrCreate()`: starring a **default** prompt copies `_id → prompt_id` and creates a **new row** with `is_custom: false`, `is_favourite: true`, `user_id`, `workspace_id`, and the copied `title`/`prompt`. `getPrompts()` re-joins those rows onto the defaults by `prompt_id`.

**This is the closest thing to copy-on-write forking anywhere in the codebase** — a per-user overlay row pointing back at a system record. Skills' fork-on-edit is a strict superset: add `forked_from`, let the body diverge, add an enabled flag.

### Storage / API
| Op | FE fn | URL | BE |
|---|---|---|---|
| list | `fetchCustomPromptsApi(workspaceId)` | `POST ai/fetchCustomPrompts` | `AIController@fetchCustomPrompts` → `AiPromptRepo::getPrompts` |
| create/update/favourite | `saveCustomPromptApi(payload)` | `POST ai/saveCustomPrompts` | `AIController@saveCustomPrompts` → `AiPromptRepo::savePrompt` |
| delete | `deleteCustomPromptApi(wsId, id)` | `POST ai/removeCustomPrompts` | `AIController@removeCustomPrompts` → `AiPromptRepo::removePrompt` |

URLs `src/config/api-utils.ts:208-210`; routes `routes/web/ai.php:27-29` — **auth + `set.locale` only, no `PermissionMiddleware`**. `workspace_id` is validated to exist but **membership is never checked**. Only `ai/fetchCustomPrompts` is registered in `config/workspace_access.php`. **This is a pre-existing authorization gap Skills must not inherit.**

Response envelope: `{ status, data: { default_prompts: [{title, prompts: [...]}], custom_prompts: [...] } }`. `saveCustomPrompts` returns the *single saved prompt* under `data` — the FE type is wrong and both modals carry a comment + cast about it.

No Vuex/Pinia persistence, no localStorage. TanStack key `aiToolsKeys.customPrompts(workspaceId)` (`queries/keys.ts:19`) is used **only for invalidation**; `SavedPromptsModal.vue` fetches imperatively into local `reactive()` state.

### How a saved prompt is applied to a message
1. `SavedPromptsModal` is mounted **once app-wide** in `AuthenticatedAppExtras.vue:128`, opened via `$cstuModal.show('saved-prompts-modal')` from `ChatInput.vue:2371` (also from `CustomGenerateForm.vue:394` in the AI content library).
2. Click a prompt → `handlePromptSelect()` emits `prompt-selected`.
3. `AuthenticatedAppExtras.handlePromptSelected` → `useAIChatStore` → `promptInputApi.applyPrompt` → `ChatBox.vue:1239 onPromptSelectedBus({prompt, source})`, gated on `source === 'ai-chat'`.
4. `chatInputRef.setContent(promptText)` + `.focus()`.

> **The single most important fact for the Skills design: the prompt body is pasted verbatim into the editor as user text. It never reaches the backend as a separate field.** Today "apply a prompt" literally means "type this text for me." Skills is a genuinely different contract — a `skill_id` travelling with the message.

Secondary path: `useAiChatEngine.ts:457 handleCustomPrompts(value, type)` with `'append'` / `'save'`.

### Migration implications
- Field mapping is nearly 1:1 — `title → name`, `prompt → instructions`, `section/section_id → category`, `is_favourite → enabled` (semantics differ), `prompt_id → forked_from`.
- **Missing and must be added:** `description`, `visibility`, `slug`/slash trigger, workspace scoping on system records, enable/disable, import/export.
- The ~1486 lines of seeded defaults are **bracket-placeholder prompt text, not instruction sets**. A mechanical import produces 100+ junk "skills." **Recommendation: migrate user custom prompts (`is_custom: true`) to personal skills; retire the seeded defaults in favour of the curated set.**
- **Three call sites** open the modal: chat input full toolbar, compact toolbar, and **`CustomGenerateForm.vue:391` in the AI content library — which is not a chat surface and has no `/` input.** Skills replacing the modal must either keep something working there or explicitly de-scope it.

---

## B3. Reusable components / services

**Directly reusable (frontend):**
- **`StreamingMarkdown.vue`** — wraps `streamdown-vue` with `parseIncompleteMarkdown`, RAF-paced reveal, `rehypeAnimateWords`. **Yes, reusable for the skill instruction preview** — pass the full string in one go (`onMounted` snaps pre-hydrated content, so a non-streaming mount renders instantly with no typewriter effect).
- **`CstTextEditor.vue`** — TipTap 3 with a stripped `StarterKit`, `Placeholder`, `imageMention`. Exposes `focus/clear/setContent/getContent/serialize/pruneOrphanMentions/editor`. **A `/`-command extension plugs in here as a third extension.**
- **`extensions/imageMention.ts`** — the template for a slash extension: `@tiptap/extension-mention` with `char: '@'`, a dedicated `PluginKey`, `VueRenderer` + `@floating-ui/dom` positioning (`placement: 'top-start'`, flip/shift), `items({query})` filter. A `/`-trigger is a near-copy with `char: '/'`, its own `PluginKey`, and a skills source.
- **`ImageMentionDropdown.vue`** — the popup list UI to clone for the skill picker.
- **`TemplateModal.vue:31-60`** — the only-me/everyone visibility dropdown block; copy the markup wholesale.
- **`AddCustomPromptModal.vue`** — the CRUD modal shape (TextInput + Textarea + validation + `$cstuModal.hide`) to evolve into the skill editor.
- `@contentstudio/ui`: `CstuModal, TextInput, Textarea, Button, Switch, Icon, Dropdown, DropdownItem, ListItem, Collapsible, SearchInput`.

**Backend reusable:**
- `HasTypedResponse` trait (`typedFlatSuccess` / `typedFlatError` / `typedFlatResponse`).
- `PermissionMiddleware` — workspace resolution + demo-workspace read-only guard + optional action permission.
- `BaseFormRequest` / `FormRequest`.
- `LogsBuilder` for audit trail.
- **`php artisan typescript:transform`** — DTOs in `app/Data/Ai/` generate FE types into `@generated/ai`. Skills should ship as `app/Data/Ai/Skills/*Data.php` so FE types are generated, not hand-written (`ai-chat.ts` already imports from `@generated/ai`).

---

## B4. Integration points

### B4a. Frontend chat

**Slash trigger — greenfield.** There is **no `/` handling anywhere today.** No `char: '/'`, no `slashCommand`, nothing in `extensions/` or `utils/`. The only `'/'` hits are string-splits in `botMessageTooltips.ts` and `useReferenceFiles.ts`. Build by cloning `imageMention.ts`.

> **Gotcha:** `CstTextEditor.vue`'s `handleKeyDown` checks `imageMention.pluginKey.getState(view.state).active` before emitting `enter`. **A slash extension must be added to that check or Enter will submit the message while the skill dropdown is open.**

**Active-skill pill.** Two existing patterns:
- `BrandVoiceSelector.vue` sits **inline in the toolbar row** (`ChatInput.vue:530`) after a `h-[23px] w-px` separator and the saved-prompts button — icon + label + `Switch`.
- `ReferenceFilesAttachment.vue` sits **above the editor** (`ChatInput.vue:266`) — `<li>` chips with icon + label + truncated name + `X` remove button, `rounded-lg border px-2.5 py-1.5`.

> **The active-skill pill should follow `ReferenceFilesAttachment`, not BrandVoice.** It is attached state with a remove affordance, exactly like a reference clip — which matches the decided UX (an `x`, not a "stop").

**Payload.** `buildChatStreamPayload.ts` — add alongside the existing conditional-inclusion idiom used for `brandVoice`/`style`:
```ts
...(ctx.activeSkill ? { skill_id: ctx.activeSkill.id } : {}),
```
Current fields: `protocol_version, workspace_id, content, chat_id, images, videos, ai_library_posts, mentions?, confirm?, brand_voice_id?, style_id?, analytics_context?, model_choice, blog_id?`.

Selected state reaches the payload today via `useAIChat()` refs (`selectedBrandVoiceId`, `selectedStyleId`). **`activeSkillId` should live in the same place.**

### B4b. Backend CRUD
Mirror the composer-template stack:
- Model `app/Models/Ai/AiSkill.php` (collection `ai_skills`).
- Repository `app/Repository/AiChat/AiSkillRepo.php` or `app/Repositories/AiSkillRepository.php`.
- Controller `app/Http/Controllers/AI/AiSkillController.php`, `use HasTypedResponse`.
- Requests `app/Http/Requests/AI/Skills/{Store,Update,Delete,List}SkillRequest.php`.
- **Routes: RESTful under `routes/api.php` inside the `PermissionMiddleware` group — NOT `routes/web/ai.php`**, whose AI-prompt routes lack permission middleware. Register every non-GET URI in `config/workspace_access.php`.
- Visibility query, from `SocialPostTemplateRepository::getTemplates`:
  ```php
  ->where('workspace_id', $workspaceId)
  ->where(fn($q) => $q->where('user_id', $userId)->orWhere('is_private', false))
  ```
  > **Do not copy `getById` (~line 95): it filters on `is_public`, which is not a field on the model, so a non-owner can never fetch a shared template by id. That is a bug.**

### B4c. Chat stream — PHP → Python
`AiChatHelper::buildWorkflowInputs()` @339 builds `['chat_history','message','protocol_version','confirm','attachments','videos','reference_*','mentions','image_meta','metadata'=>[...,'brand_guidelines' /* injected @424/@428 */],'user_id','session_id','credits_limits']`.

**Skills injects here:** resolve `skill_id` → `{name, description, instructions}` and set `$workflowInputs['metadata']['active_skill']`.

> **Gotcha:** add `'skill_id' => 'nullable|string'` to `app/Http/Requests/AI/AiChatRequest.php` (beside `brand_voice_id` @89 / `style_id` @90). The file carries its own warning that **an unruled key is silently dropped by `validated()`** — forgetting it produces a silent no-op, not an error.

### B4d. ai-agents prompt assembly
`StreamRequest` (`streaming_router.py:613`) has `model_config = ConfigDict(extra="ignore")` and a generic `metadata: dict[str, Any] | None`, so **`metadata.active_skill` arrives with zero Python schema changes.**

The substitution mechanism mirrors brand voice exactly:
1. `streaming_router.py:1267-1310` reads `meta.get("brand_guidelines")`, runs `BrandVoiceParser.to_prompt_context()`, produces `brand_voice_block`.
2. `:1872` / `:1898` pass `{"brand_voice_block": ..., "target_platforms_block": ..., "language_block": ...}` into the team run.
3. Agno substitutes into `{brand_voice_block}` placeholders in member `instructions` lists (`agents.py:1834` caption, `:2153` image, `:2497` video, `:2696` assistant; also `post_plan/steps.py:95`).

**Skills mirrors this:** build `active_skill_block` in the router; add `{active_skill_block}` to member instruction lists under the "# Dynamic context for THIS turn (auto-substituted)" heading.

> **Prompt-caching caveat (documented in `bootstrap.py`):** per-session data must sit at the **END** of the instruction list — anything before it is what stays cacheable. Placeholders today sit 29–127 tokens in, below Anthropic's 1024-token minimum, so members cache nothing. A verbose skill block placed early would permanently poison prompt caching. **Place `{active_skill_block}` last**, or in the dynamic-context group where `{brand_voice_block}` already sits.

> **Routing caveat:** the coordinator **routes to a member**. A skill saying "analyse my queue" must not fight the routing description. Either inject the block into every member (simple, safe) or let a skill declare a preferred member (v2).

---

## B5. Toolkit inventory vs. the 18 proposed skills

### Inventory — 32 tools across 5 toolkits
All in `contentstudio-ai-agents/src/integrations/contentstudio/toolkits/`. Design constraint in `__init__.py`: **"no agent exceeds ~13 model-visible tools (D-07)."**

**`WorkspaceToolkit`** (6, read) — `workspace_list_workspaces` · `workspace_list_accounts` (incl. token state / `needs_reconnect`) · `workspace_list_groupings` (labels/campaigns/content categories) · `workspace_list_team_members` · `workspace_list_approval_workflows` · `workspace_list_connectable_platforms`

**`PublishingToolkit`** (3, read) — `posts_list` (id, status, scheduled time, target accounts, **truncated caption**; filters by date/status/label/campaign/category/author/approval) · `posts_internal_notes_list` · `media_list`

**`PublishingWriteToolkit`** (6, write, HITL-gated) — `posts_create` · `posts_update` · `posts_create_from_plan` · `posts_delete` · `posts_approval` · `posts_add_internal_note`

**`AnalyticsToolkit`** (6, read) *(not exported from `toolkits/__init__.py`; imported directly at `agents.py:2966`)* — `analytics_fetch_metrics` (reach/engagement/audience/video by period, over-time, period-over-period) · `analytics_compare_accounts` (ranks on ONE metric; check `ranking_complete`) · `analytics_top_content` (ranked **within** an account, never across) · `analytics_insights` (**no X/Twitter**; slow) · `analytics_best_times` (**Facebook + Instagram only**) · `analytics_list_available_metrics`

**`OrganisationWriteToolkit`** (11, write, gated) — `groupings_create/update/delete` · `workspace_create/update/delete` · `team_member_invite/remove` · `account_connect_start` · `account_add_facebook_group` · `media_import`

**Also available:** Exa web tools (`tool_exa_answer`, `tool_get_contents`, `tool_find_similar`), image/video generation tools, and the **Post Plan workflow** (`src/orchestration/post_plan/`, enabled by default) for compound multi-post plans.

> **Notably absent: there is no inbox toolkit.** `src/api/routers/inbox_router.py` exposes `POST /inbox/process` backed by `src/agents/content/inbox_reply_agent.py`, but it is a standalone endpoint driven by the Inbox UI — **the chat team cannot call it.** Public API v1 does have inbox (`routes/api/v1.php:497-542`: `inbox/elements/search`, `inbox/elements/summary`, `reviews`, `PUT reviews/{id}/reply`), all proxied to social-inbox-manager.

### The verdict table

| # | Skill | Data needed | Backing tool today? | Verdict |
|---|---|---|---|---|
| 1 | `/content-plan` | Pillars/topics, accounts, calendar gaps, cadence | **Yes** — Post Plan workflow + `posts_list` + `workspace_list_accounts` + `workspace_list_groupings` + brand block | **Ships v1** |
| 2 | `/find-content-pillars` | Brand knowledge, historical top posts | **Partial** — `analytics_top_content` yes; pillars arrive via `metadata.brand_guidelines`, not as a queryable tool | **Ships v1** (brand block already in context) |
| 3 | `/generate-post-ideas` | Brand voice + pillars, recent posts, web research | **Yes** — brand block + `posts_list` + Exa + `caption_agent` | **Ships v1** |
| 4 | `/hook-maker` | Nothing external | **N/A** — pure instruction | **Ships v1** |
| 5 | `/caption-polisher` | User text + brand voice | **Yes** — brand block + `post_improver_agent` | **Ships v1** |
| 6 | `/repurpose-this` | A source post + target platform rules | **Partial** — `posts_list` returns truncated previews only | **Ships v1 as paste-in**; add `posts_get(id)` for "repurpose post X" |
| 7 | `/queue-snapshot` | Scheduled/draft/approval-pending posts, gaps | **Yes** — `posts_list` explicitly owns gaps + counts | **Ships v1** |
| 8 | `/best-time-to-post` | Audience-online by day/hour | **Yes** — `analytics_best_times` | **Ships v1 with a hard caveat — Facebook + Instagram only.** The skill body must state this or it will hallucinate times for LinkedIn/TikTok. |
| 9 | `/prep-this-post` | One post's full content, media, targets, approval state | **Partial** — no full-post read | **Needs tool work** — `posts_get(id)` returning full caption + per-platform settings |
| 10 | `/weekly-performance` | Period metrics + prior-period delta, top posts, narrative | **Yes** — `analytics_fetch_metrics(compare_to_previous=True)` + `analytics_top_content` + `analytics_insights` | **Ships v1 — the strongest-backed analytics skill** |
| 11 | `/post-postmortem` | One post's metrics vs. baseline | **Partial** — no per-post metric lookup by id | **Needs tool work** — `analytics_post_metrics(post_id)`. Degrades badly today. |
| 12 | `/explain-this-metric` | Metric definitions + platform semantics | **Yes** — `analytics_list_available_metrics` + `analytics_registry.py` (553 lines of metric metadata) | **Ships v1** |
| 13 | `/account-growth` | Follower/audience over time, per-account comparison | **Yes** — `analytics_fetch_metrics(metric_set="audience", over_time=True)` + `analytics_compare_accounts` | **Ships v1** |
| 14 | `/triage-inbox` | Unread comments/DMs/mentions, priority | **No** — zero inbox tools on the chat team | **Blocked — needs an `InboxToolkit`. Do not ship v1.** |
| 15 | `/reply-to-this` | One conversation thread + brand voice | **No** as a chat tool | **Paste-in-only for v1**; full version needs `InboxToolkit` |
| 16 | `/respond-to-review` | Review text, rating, business context | **No** as a chat tool | **Paste-in-only for v1**; needs `InboxToolkit` for the real flow |
| 17 | `/account-health` | Token/reconnect state, cadence, failures | **Mostly** — `workspace_list_accounts` returns `needs_reconnect`, `SessionContext.unhealthy_accounts` already computed in `bootstrap.py`; `posts_list` gives cadence. **Failed/errored posts not clearly exposed.** | **Ships v1 (reduced)** — reconnects + cadence yes, failure history no |
| 18 | `/workspace-setup-check` | Accounts, categories, labels, campaigns, approval workflows, team, brand profile | **Yes** — the entire `WorkspaceToolkit` (6/6) plus `SessionContext` (pre-loads all of it into `cs_context` at zero prompt cost) | **Ships v1 — best-backed skill of all 18** |

**Summary: 11 ship clean · 3 ship reduced/caveated (6, 8, 17) · 2 need tool work (9, 11) · 3 blocked on an inbox toolkit (14, 15, 16 — two can ship paste-in-only).**

> **Headline:** the inbox trio is 3 of 18 and has **no backing at all** in the chat agent. The data exists behind public API v1 → social-inbox-manager, and `ContentStudioClient` already speaks to that API, so an `InboxToolkit` (~3-4 tools: `inbox_list_conversations`, `inbox_get_thread`, `inbox_list_reviews`, gated `inbox_reply`) is a well-defined follow-up — but it is real work, and it pushes a member past the D-07 13-tool ceiling, so **it needs its own member agent** (like Analytics Analyst got).

---

## B6. Technical considerations

### Proposed data model
```
ai_skills (MongoDB)
  _id
  workspace_id      string|null   # null ⇒ system skill (global, like is_default prompts today)
  user_id           string|null   # owner; null for system
  name              string
  slug              string        # the "/" trigger
  description       string        # one-liner in the "/" dropdown — hard char cap
  instructions      string        # markdown body
  category          string|null   # mirrors `section` on ai_custom_prompt
  is_system         bool          # replaces is_default
  is_private        bool          # ← reuse the composer-template field name verbatim
  forked_from       ObjectId|null # ← generalises prompt_id
  version           int           # bump on edit; lets a fork detect an upstream change
  surface_scope     string[]      # ['chat'] for v1; future: composer, inbox, analytics
  created_at / updated_at / updated_by_id / updated_by

ai_skill_states
  workspace_id, user_id, skill_id, enabled: bool
```

> **Per-user enable/disable must be a separate row, not a field on the skill** — a system skill has no per-user row to write to. This is exactly how the current favourite overlay works (`prompt_id` + `user_id` row), so the pattern is already proven in this codebase.

**Slug uniqueness is the sharpest design question.** Two users in a workspace can both fork `/hook-maker`. Proposal: slug is *not* globally unique; the `/` dropdown disambiguates by owner, and fork resolution is **"my version wins."**

### API
- REST under `routes/api.php` behind `PermissionMiddleware`: `GET|POST /skills`, `GET|PUT|DELETE /skills/{id}`, `POST /skills/{id}/fork`, `POST /skills/{id}/toggle`, `POST /skills/import`, `GET /skills/{id}/export`.
- **Do not** copy the verb-in-URL style from `routes/web/ai.php` — no permission middleware, no membership check.
- DTOs in `app/Data/Ai/Skills/`; run `php artisan typescript:transform` so FE types land in `@generated/ai`.
- Import/export envelope: `{version, skills: [{name, slug, description, instructions, category}]}`. **`instructions` is untrusted markdown that becomes part of a system prompt** — needs a length cap (suggest 8–16 KB) and a prompt-injection guard. `src/orchestration/guardrails.py` already exists in ai-agents.

### Migration
1. Add `ai_skills`; do **not** touch `ai_custom_prompt`.
2. Backfill `ai_custom_prompt where is_custom: true` → `ai_skills` with `is_private: true`, `forked_from: null`, `name: title`, `instructions: prompt`, `category: section`.
3. Seed the curated system skills via an **idempotent** command. `AICustomPromptsSeeder` uses bare `create()` and duplicates on re-run — do not repeat that. Follow `ApprovalWorkflowRepository`'s `Cache::lock` pattern (`app/Repository/ApprovalWorkflowRepository.php:104-128`) if per-workspace uniqueness is needed.
4. Keep `SavedPromptsModal` behind a feature flag until Skills is out. **`CustomGenerateForm.vue:391` (AI content library) is a non-chat consumer that needs its own story.**
5. **There is no per-workspace seeded-default pattern anywhere in the codebase** — content categories, approval workflows and prompts are all either global or created on demand. Recommend keeping system skills **global** (`workspace_id: null`), with per-user overlay rows for state.

### Caching
- **FE:** add `skills: (workspaceId) => ({queryKey: ['skills', workspaceId]})` to `aiToolsKeys`. Workspace-keyed, so `cleanupPreviousWorkspaceQueries` sweeps it on workspace switch.
- **BE:** `SessionContext` in `bootstrap.py` has a 300s read-through TTL cache keyed on a hash of the API key, max 512 entries. **Do not put skill bodies in `cs_context`** — it is per-session and skills change on user edit; a 5-minute stale system prompt is a bad bug. Resolve the skill in PHP per-turn (a single Mongo read).
- **Prompt caching:** see B4d — block placement matters.

### Session state for the active skill
1. **Per-turn (recommended).** `skill_id` rides the request like `brand_voice_id`/`style_id`. The FE pill holds it; a new chat clears it. Zero backend state. Matches the existing brand-voice contract exactly.
2. **Per-conversation.** `session_state.py` gives `load/peek/remember/write_key(session_id, key, value)` over a Neon `workflow_state` document. But its docstring warns `load()` deliberately does not serve from cache and every read costs a Neon round trip (~150 ms). **Paying that on every chat turn for a hint is the wrong trade** — the router already avoids it for its routing hint (D-18).

> **Verdict: per-turn `skill_id`, with the FE owning stickiness.** This aligns exactly with the decided UX (sticky within a session, cleared on a new session). If persistence beyond that is ever wanted, store it on the `ai_chat` Mongo document in PHP, not in Python session state.

### Other
- `routes/web/ai.php` marks `chatWithStreaming` with `ForceJsonResponse` so validation failures are 422s, not 302s.
- **Credits:** `processAgnoAgent` requires ≥5 text credits before anything runs. A skill that fans out to several tools costs materially more than a chat turn, and **no per-skill credit accounting exists.**

---

## B7. Mobile impact

`contentstudio-flutter/lib/features/ai_assistant/` — 39 files, Riverpod + Dio, feature-complete for basic chat.

**Exists:** `ai_assistant_view.dart` / `ai_assistant_sheet.dart`, `ai_input_bar.dart` (plain text input), `ai_message_bubble.dart`, `ai_history_drawer.dart`, `ai_follow_up_chips.dart`, `ai_writing_options_sheet.dart`, `ai_generated_image.dart`, `ai_composer_bridge.dart`. SSE via `ai_sse_decoder.dart` + `ai_stream_event_mapper.dart`.

**Does NOT have:** saved prompts (zero references) · brand voice selection (no `brand_voice_id`/`style_id` anywhere) · mentions · reference files · media-generation options. **Protocol v1 only** — `streamingBody` (`ai_assistant_endpoints.dart:18-35`) sends only `{workspace_id, content, images, videos, ai_library_posts: null, model_choice}` + optional `chat_id`; no `protocol_version`. The input is a plain Flutter text field — **no rich editor, so no `/`-trigger infrastructure at all.**

**A Skills addition would touch:** `ai_assistant_endpoints.dart` (add `skill_id`), `ai_assistant_service.dart` + `_repository.dart` + `_controller.dart`/`_state.dart`, a new skills repository/service, a picker bottom sheet (following `ai_writing_options_sheet.dart`), and a chip in `ai_input_bar.dart`.

> **Recommendation: mobile is out of scope for v1.** Mobile is already two features behind web. The cheapest correct posture is a later **read-only skill picker** — a bottom sheet listing enabled skills that sends `skill_id`, deferring authoring, forking, visibility and import/export to web. Because `skill_id` is a single scalar on an existing endpoint that is roughly a day of work whenever chosen. **Nothing in the v1 design blocks it.**

---

# PART C — Open questions for the review gate

## C1. Product decisions

1. **The inbox trio.** `/triage-inbox`, `/reply-to-this`, `/respond-to-review` have no backing tools. Options: (a) ship 15 skills now, add the trio after an `InboxToolkit`; (b) ship `/reply-to-this` and `/respond-to-review` as paste-in-only and hold `/triage-inbox` (triage is meaningless without a real list); (c) hold the launch and build the toolkit first.
2. **`/prep-this-post` and `/post-postmortem`** both need small new tools (`posts_get(post_id)`, `analytics_post_metrics(post_id)`). In scope for this epic or a prerequisite?
3. **Naming collision** with the CLI/MCP agent-skills work. Are these one concept or two? If two, what is each called internally and externally?
4. **The AI content library surface** (`CustomGenerateForm.vue`) opens the saved-prompts modal and is not a chat surface. Does it keep prompts, get skills, or lose the feature?
5. **The 1486 lines of seeded default prompts** — delete, migrate, or keep both systems? (Recommendation: retire.)
6. **Workspace-level disable.** Decided spec is per-user enable/disable. A workspace admin will also expect to hide a skill from the whole team. Two different tables — in v1 or not?
7. **Admin approval before a skill goes "Everyone"?** A shared skill lets one member write text that lands in every teammate's system prompt.
8. **Credits.** A skill fanning out to 4 analytics calls costs materially more than a chat turn. Metered differently, or absorbed?
9. **Entitlement.** Is Skills plan-gated or available to all tiers?

## C2. Technical decisions

10. **Does an active skill replace or augment member instructions?** Brand voice augments. If `/hook-maker` should never route to Workspace Data, the coordinator's routing descriptions will fight the skill. Can a skill pin a member? (Recommend: v2.)
11. **Skill vs. brand voice precedence.** If a skill says "be punchy" and brand voice says "be formal", who wins? Needs an explicit rule written into the block text.
12. **Prompt-cache regression.** Members cache nothing today. A 500-token skill body inserted mid-list makes it permanently worse. Confirm the block is appended last and that cache hit rate is measured.
13. **Slug collisions across fork lineage.** Two users fork `/queue-snapshot`; a third sees both. Proposal: not unique, dropdown disambiguates, "mine wins."
14. **Fork drift.** When a system skill is updated, what happens to forks? `version` + an "update available" affordance, or silent divergence?
15. **Instruction body as an injection vector.** Cap, sanitise, route through `guardrails.py`, or all three?
16. **Platform coverage disclaimers.** `analytics_best_times` is Facebook+Instagram only; `analytics_insights` excludes X. Does the caveat live in the skill body or the tool layer? Getting this wrong produces confident hallucinated answers.
17. **The pre-existing authorization gap** on `routes/web/ai.php:27-29` (no workspace-membership check on prompt endpoints). In scope to close, or left as-is while Skills gets it right?

---

## C3. Recommended approach (synthesis)

**Positioning.** Ship "Skills" and state the split explicitly in the empty state and docs: **Skills define *how* a task is done; brand voice defines *how you sound*.** Skills sit on top of brand voice, never replace it. The overlapping-primitives confusion is the #1 documented failure in this space and it is a positioning problem, not a technical one — which is also the strongest argument for **absorbing saved prompts rather than running both**.

**What the decided design already gets right — protect these:**
- **Explicit slash invocation, one skill at a time** eliminates the entire routing-failure class that dominates the Claude ecosystem's complaints, and makes the feature debuggable for support.
- **A removable pill above the composer**, following `ReferenceFilesAttachment` — correct signal, matches how chat UIs indicate attached context.
- **Copy-on-write fork on edit** — best-in-class, and undocumented anywhere in the social category.

**What to add or firm up:**
1. **Cap the catalog at ~18.** Vista ships 20+. SocialBee's 1,000 prompts and the GPT Store's 159K are the cautionary tales.
2. **Enforce `description` hard**, even though we don't auto-route — it is what the user reads in the slash menu. Cap it (Claude uses 200 chars) and require it. Highest-leverage field in the object.
3. **Build two surfaces.** Slash menu = recall (fuzzy-matched, grouped, keyboard-navigable). Library page = discovery + management (toggles, search, filter, edit, delete, create). **Do not make the slash menu do both jobs.**
4. **Add a visible affordance** next to the composer, not slash-only. Hidden features don't get used.
5. **Defaults enabled by default, individually toggle-off-able per user** — Claude's provisioning pattern.
6. **Plan for bloat now:** search, last-used sort, usage count. Consider an archive state rather than only delete.
7. **"Ask the AI to build a skill" is table stakes.** The differentiator is that the generated skill lands in an editable markdown body the user actually reads and refines.
8. **Do not ship auto-routing in v1.** If wanted later, do it as a Cursor-style per-skill activation mode, not a global behaviour change — and budget for description truncation first.
9. **Reserve "Agents"** for scheduled autonomous work. Design the Skill as a standalone object so a future Agent can reference one; it costs nothing to leave that door open.
