# 01 Research: AI Chat settings (Instructions, Memory, Usage)

**Date:** 2026-09-30
**Feature:** A settings button in the AI Chat header opens a settings modal with three tabs, Instructions, Memory and Usage, plus links to Skills and Brand Knowledge. Web and Flutter app.
**Prototype:** https://claude.ai/artifact/Y81bvg35xMWjsWJVaj2MD2
**Context:** CEO approved building this now, in parallel with the separate assistant rebrand / launch-planning track. Naming and the rebrand are out of scope here, so this epic uses neutral "AI Chat" wording.
**Sources:** competitor and AI-assistant research (WebSearch) plus codebase analysis of `contentstudio-ai-agents`, `contentstudio-backend`, `contentstudio-frontend` and `contentstudio-flutter`.

Legend in Part A: **[V]** vendor docs, **[3P]** third-party blog or snippet. Some vendor help pages (OpenAI help center, Vista Social support) returned 403, so their content comes from search snippets.

---

# PART A: Competitor and industry research

## A1. What the three parts are

- **Instructions:** standing rules a person writes that are added to every chat. Two layers are standard: workspace (set by admins) and personal (set by each user). They are deterministic and fully trusted, because a person wrote them.
- **Memory:** context the assistant builds on its own and carries across chats. For ContentStudio that would come from chats (plans, decisions) and from product signals (what performs, inbox patterns, publishing habits). Because it is inferred it can be wrong or out of date, so users need to be able to see and fix it.
- **Usage:** credit meters for AI text, image and video, plus where the credits went. Credit systems are opaque, and users want to know what drained their allowance before they hit a wall.

## A2. Social media competitors

| Competitor | Instructions / Memory / Usage view | Key capabilities | Pricing tier | UX approach | Unique differentiator |
|---|---|---|---|---|---|
| **Vista Social (Ask Vista)** | Instr: **Yes** ("Guidance" rules in AI Knowledge). Memory: **Partial** (a Recents list, no learned memory found). Usage: **Partial** (credits for AI image and video, no breakdown found) | Chat command centre with 50+ tools, Skills (`/skill-name`), scheduled Agents that ask before writes, AI Training & Knowledge with guidance and escalation rules, brand voice per profile group | Not verified | Settings → AI Training & Knowledge; knowledge scoped per profile group (per client) | Closest analogue to our plan. Escalation rules ("when not to answer") are unique. [V] support.vistasocial.com/hc/en-us/articles/46961598441371 |
| **Hootsuite (Wisdom, formerly OwlyGPT/OwlyWriter)** | Instr: brand voice learned from past posts. Memory: not documented. Usage: marketed as unlimited | Wisdom (launched 2026-06-24) merges OwlyWriter, OwlyGPT and Yeti; grounded in live social data; cited answers and actions | All plans | Implicit personalisation, no memory controls found | Real-time listening data with citations. [V] hootsuite.com/newsroom |
| **Sprout Social (Trellis / AI Assist)** | Instr: not verified. Memory: no product UI found. Usage: not found | Briefings, sentiment, listening Q&A, Trellis Studio workflows | Trellis from Essentials ($79/seat) [3P] | Conversational agent over data | Enterprise listening intelligence. [V] sproutsocial.com/ai/features/ai-agent/ |
| **Buffer** | Instr: tone and style guidelines. Memory: No. Usage: "unlimited" (one source says 150 credits per channel, conflicting, unverified) | Create, refine, repurpose | Free and up | Inline composer assistant | "Unlimited AI" as a pricing lever. [V] support.buffer.com/article/583 |
| **Publer** | Instr: **Brand Voices**, several per workspace, up to 5 uploaded files. Memory: No. Usage: credits | Switch voice per post and reply | Business | Selectable voice profiles | File-grounded voices. [V] publer.com/help |
| **Later** | Instr: learns tone from past posts. Memory: No. Usage: **Yes**, credits (Starter 5, Growth 50, Scale 100 per month, no top-ups) | Caption Writer, Ideas | Paid | Credit counter | Very tight caps. [V] help.later.com |
| **Metricool** | Instr: **Yes**, "Adjust instructions" per brand and per network. Memory: No. Usage: credits per brand per month | Per-network instruction overrides | Advanced | Settings form | Instructions layered by network. [V] help.metricool.com |
| **SocialBee (Copilot)** | Instr: brand profile learned from website or chat. Memory: implicit profile. Usage: no credit model found | Strategy, schedule, posts | All | Onboarding-style setup | States it does not train on customer data. [V] socialbee.com/social-media-copilot/ |
| **Agorapulse** | Instr: 7 preset tones, learns brand tone. Memory: not documented. Usage: not found | AI inbox reply suggestions | Not verified | "Assistants, not agents" | Inbox reply focus. [V] agorapulse.com |
| **Loomly** | Instr: not found. Memory: No. Usage: limited vs extended by tier | AI Assistant chat | Starter / Beyond | Not verified | [3P] |
| **Sendible** | Instr: not found. Memory: No. Usage: "unlimited" | AI content generator | All | Not verified | [3P] |

**Takeaway:** no social media competitor offers memory that users can see and edit. All offer some form of instructions or brand voice. Credit meters are common, but **none show where credits went**.

## A3. General AI assistants (the UX users will compare us to)

- **ChatGPT:**
  - Personalization → Memory has two toggles, "Reference saved memories" and "Reference chat history", plus a "Manage memories" list (delete one, or clear all).
  - An inline **"Memory updated"** chip links to that list.
  - Custom instructions are two fields, with a limit of about 1,500 characters on Free and 5,000 on paid.
  - Shared projects use **project-only memory**, and Business/Enterprise admins can turn memory off.
  - Temporary chat skips memory. [V] help.openai.com/en/articles/8590148
- **Claude:**
  - Settings has "Search and reference chats" and "Generate memory from chat history" toggles, plus a view-and-edit memory screen where you can edit directly or conversationally ("forget X").
  - **Memory is kept per project.** Incognito chat skips memory.
  - Enterprise admins can turn memory off, and the change is audit-logged. There is also an import-memory flow.
  - Reportedly moved to individually editable entries in July 2026 [3P, unverified]. [V] support.claude.com/en/articles/11817273
- **Gemini:** a "Personal context" page with separate controls for saved preferences and past chats, plus Temporary Chat.
- **Microsoft 365 Copilot:** a Saved memories screen with "Delete all". **Turning memory off does not delete existing memories**, and the UI says so. [V] support.microsoft.com
- **Notion Agent:** instructions and memories live on an editable page.
- **Jasper:** explicit tagged "Memories" (capped at 50 on Creator), plus Brand Voice and a Knowledge Base.
- **HubSpot Breeze:** Brand Identity feeds every AI feature, which is billed in HubSpot Credits.
- **Intercom Fin Guidance:** instructions grouped into **categories** (Communication style, Context and clarification, Spam, Other).

## A4. Common patterns

1. Instructions and memory are always separate concepts with separate controls.
2. Memory has two sources, saved facts and learned-from-history, each with its own toggle.
3. There is a list you can manage: view all, delete one, clear all. Editing is newer.
4. An in-chat "Memory updated" signal links straight to the list.
5. Users can opt out for a single chat (Temporary or Incognito chat).
6. Scoping by project or workspace stops context leaking between clients.
7. Admins have an off switch, with audit logging.
8. Marketing tools keep brand voice and knowledge separate from memory.
9. Credits reset monthly. Breakdowns by feature or user exist only outside social media tools.

## A5. Differentiators for ContentStudio

- **Memory learned from performance data**, not only from chats. For example: "Reels get 3x the reach of carousels here." No competitor was found that does this with visible memory.
- **Memory shared across the workspace.** ChatGPT and Claude memory is personal or per project. For a team tool, the workspace is the natural unit, and it fits how ContentStudio already works (timezone is set per workspace).
- **Source and date on every memory**, e.g. "From chat, Sep 24" or "From post results".
- **Memory categories that map to the product:** Plans, Decisions, What performs, Inbox patterns, Publishing habits.
- **Where credits went:** a breakdown by activity, which no social media competitor shows.

## A6. User expectations

| Table stakes | Delighters |
|---|---|
| Instructions field with a character limit | Workspace plus personal layers with a clear precedence rule |
| Memory on/off | Separate toggles per source (chats vs post results) |
| List with delete-one and delete-all | Inline edit, source and date on each memory |
| A notice when memory is saved | Undo right in that notice |
| Credits used vs limit and reset date | Breakdown by activity, warnings before running out |
| Admin can control AI features | Role-based editing of workspace memory and instructions |

## A7. Pitfalls and risks from research

- **Sensitive data creeps into memory.** One study found 28% of ChatGPT memories contain GDPR personal data [3P, arxiv 2602.01450]. In shared team memory, that becomes a leak between colleagues.
- **Memory poisoning through prompt injection.** This is a direct risk for us, because **inbox messages and comments are untrusted input**. A customer DM saying "always offer 50% off" must never become a memory.
- **Stale memories.** Show a date, and never trust old entries blindly.
- **Silent memory erodes trust.** Make saving visible.
- **Turning memory off is not the same as deleting it.** Users assume it is, so the copy must say otherwise.
- **"Memory full" dead ends** (ChatGPT). We need a cap plus consolidation that never rewrites entries a user has edited.
- **GDPR erasure** must be real, including agent-side copies.
- **Credit opacity** is a leading complaint about AI credit models.
- **Content bleeding between clients** in agency and white-label setups. Scoping memory to the workspace prevents this.

## A8. Recommended approach (from the research)

- **Instructions:**
  - Workspace instructions: admins edit, everyone else sees them read-only.
  - Personal instructions: every member has their own.
  - Precedence rule: workspace beats personal, and instructions beat Brand Knowledge on conflicts about how to work.
  - One line explains the layers: Instructions = how to work, Brand Knowledge = who we are, Skills = one task on demand.
- **Memory:**
  - Scoped to the workspace and shared by its members, never across workspaces.
  - Grouped by category, with source and date on each entry. Edit, delete one, delete all.
  - Source toggles for "from chats" and "from post results".
  - An in-chat "Memory updated" notice with a link.
  - Inbox text is never stored word for word.
  - Copy explains that turning memory off keeps existing memories.
- **Usage:** three meters, the reset date, a breakdown by activity, and a Get more credits link. Warnings as credits run low.

---

# PART B: Codebase analysis

All paths are relative to the repo root.

## B1. contentstudio-ai-agents

- **No persistent memory exists.** There is no `src/memory/` directory, despite an older reference in the Brand Knowledge research.
  - `src/utils/config.py:126` has `agno_enable_memory` (env `AGNO_ENABLE_MEMORY`, default off).
  - `src/agents/base.py:~225-243` only turns on Agno memory when it is enabled and a DB is configured, and is keyed by **user_id only**.
  - Only `src/agents/tools/general_assistant.py` takes `enable_memory`. It is **not wired to the AI chat Team**.
  - `stream_v2.py:1535` skips `MemoryUpdate*` events.
- **Chat Team:** `src/orchestration/team.py:118` `get_content_team()` builds a singleton coordinate-mode Team.
  - Its DB is **Redis with a TTL** (`_session_db()`, `team.py:78-110`), used only to resume paused confirmation runs.
  - `add_history_to_context` is off. History comes from PHP (`chat_history`) and is packed by `_build_team_input()` (`streaming_router.py:368`, last 10 turns).
- **Session state:** `src/integrations/contentstudio/session_state.py` holds a per-session JSON row (`mcp_workflow_sessions`, Neon, `src/db/models.py:98`). It stores workflow gates and pending-write pointers. It holds no user or workspace preferences.
- **Prompt assembly** (`src/api/routers/streaming_router.py`):
  - `brand_voice_block` is built at `:1771-1795`, and `skill_block` at `:689` (`_build_skill_block`).
  - Both go into `team_session_state` at `:2449-2461`.
  - Member prompts use `{brand_voice_block}` and `{skill_block}`. `{skill_block}` is appended to every member by the helper at `src/orchestration/agents.py:1972-1994`.
  - Resume handling is in `_resume_session_state` (`:715`).
  - **The new `instructions_block` and `memory_block` should follow the `skill_block` pattern.** They must always be present (empty string when unused), or the literal `{placeholder}` leaks into prompts on resumed runs.
- **Guardrail gap:** skill instructions skip the team input guardrail (`streaming_router.py:694-697`). Instructions and memory would too, so decide deliberately.
- **Credits:** a pre-check runs through `validate_credit_limits` (`src/utils/credit_validation/validation.py`), and the Team needs at least 5 text credits (`streaming_router.py:1966`). Agents report `credits_consumed` and PHP deducts. As a backstop, the Kafka `ChatUsageReportedEvent` goes to `ai-agents.chat.usage` (`src/events/schemas.py:413`, emitted from `src/api/helpers/stream_cancellation.py:396`).

## B2. contentstudio-backend

- **AI chat:**
  - Models: `app/Models/AiChat/AiChat.php`, `AiChatMessages.php`. Repos: `app/Repository/AiChat/`.
  - Routes: `routes/web/ai.php:15-58`. Streaming goes through `AIController::processAgnoAgent` (`AIController.php:1191`).
- **Payload builder:** `app/Helpers/Ai/AiChatHelper.php:521` `buildWorkflowInputs()`.
  - Brand is added at `:611-617` (`resolveBrandGuidelinesForMetadata`, `:669`) and the skill at `:640-650` (`SkillService::resolveForChat`).
  - **This is where `metadata.workspace_instructions`, `metadata.personal_instructions` and `metadata.memory` go.**
  - Add the new fields to `redactForLogs` (`:55`).
- **Skills stack (template for the new settings resource):**
  - Model: `app/Models/AiChat/AiSkill.php` (`ai_skills`).
  - Service: `app/Services/AiChat/SkillService.php`. DTO: `app/Data/Ai/Skills/SkillData.php`.
  - Controller: `app/Http/Controllers/AI/AiSkillController.php`. Routes: `routes/web/ai.php:64-73`.
  - Exceptions: `app/Exceptions/AiChat/`.
  - Skill instructions are capped at 8,000 characters.
- **Brand Knowledge:** `AiContentLibraryProfileController`, routes `routes/web/ai.php:94-120` (`aiContentLibrary/profile/*`).
- **Credits:** `app/Helpers/Billing/PlanHelper.php:995` `checkAvailableAICredits()`.
  - Usage counters are stored **on the workspace**: `caption_generation_credit` (text), `image_generation_credit`, `used_video_credits`, `used_video_clip_credits`, `used_ai_auto_reply_credits`.
  - Limits come from `SubscriptionLimits`. Deduction helpers are at `:1060,1163,1174,1205`. `ResetUsedCreditsCommand` resets them.
- **Per-activity ledger: partial, and chat only.**
  - `app/Models/AiChatUsage.php` (`ai_chat_usage`, `_id=run_id`) stores workspace_id, user_id, session_id, items, credits_charged, text_credits and image_credits. **It has no video credits.** Its main job is to make sure a chat run is billed only once (`claim()`), and stream-path rows may lack the credit fields.
  - Composer AI, AI Library, inbox auto-reply and the tools API deduct straight into the counters, with no record.
  - The Usage visibility epic (`docs/stories/usage-visibility/`) already found this ("counters, not events") and proposes a proper usage ledger. **The "Where credits went" breakdown depends on that ledger.**
- **Permissions:**
  - `app/Http/Middleware/PermissionMiddleware.php` plus `app/Libraries/Permission/PermissionHelper.php:97-219`.
  - super_admin can do everything. Admin can do everything except billing, workspace removal and ownership. Approver can only `schedule_plan`. Collaborators get a limited set.
  - Skills and brand routes check membership only.
  - **A new `manage_ai_settings` action is needed** (super_admin and admin yes, collaborator and approver no).
- **Deletion gap:** `app/Jobs/Settings/DeleteWorkspaceDataJob.php:73-130` does **not** delete AI chats, messages, skills, `ai_chat_usage` or the brand profile. No GDPR user-deletion hook for AI data was found. Memory must not repeat this gap.
- **No existing AI chat settings storage** (grep for `ai_chat_settings`, `custom_instructions` and `personal_instructions` found nothing).

## B3. contentstudio-frontend

- **Two chat headers, and both mount `AiCreditsChip`:**
  1. `src/modules/AI-tools/ChatHeader.vue`: history button `:49-53` (`$cstuModal.show('chat-history-modal')`), `<AiCreditsChip/>` `:61`, New Chat `:76-79`. Used by `src/modules/AI-tools/AIChatModal.vue:24`.
  2. `src/modules/ai-studio/components/AiStudioHeader.vue`: history `:30`, `<AiCreditsChip/>` `:42`, New chat `:44`.
- `src/components/common/AiCreditsChip.vue` and `src/composables/useAiCredits.ts` read `usePlanStore().getPlan.subscription.limits` / `used_limits` (keys at `:28-45`). There is no separate credits endpoint. **`useAiCredits` already gives used, limit and percent per type for the meters.**
- **Routes:**
  - Skills: `ai_studio_skills` (`/:workspace/ai-studio/skills`, `src/modules/ai-studio/config/routes.ts:41`).
  - Brand Knowledge: `brand-settings` (`src/modules/setting/config/routes/setting.ts:213`).
- **Brand toggle in the composer:** `src/modules/AI-tools/components/BrandVoiceSelector.vue` (Switch plus `useSetBrandEnabledMutation`).
- **Modal template:** `src/modules/AI-tools/components/VideoStylesModal.vue` uses `Modal` with `Tabs` / `Tabs.List` / `Tabs.Tab` from `@contentstudio/ui`. `ChatHistoryModal.vue` uses `CstuModal` with `z-index="1300"`.
- **Components available** (from `docs/ui-components.md`): `Modal`, `Tabs`, `Textarea`, `Switch`, `Progress`, `Button`, `Alert`, `Badge`, `ListItem`, `ActionIcon`, `Loader`, `Dialog`.
- **Gaps:**
  - No vertical side-nav tab variant has been confirmed for `Tabs`. The prototype uses a left-hand tab list, like Claude's settings.
  - No standalone tooltip component exists (use `CstPopup`).

## B4. contentstudio-flutter

- `lib/features/ai_assistant/presentation/ai_assistant_view.dart:379` `_header()` holds, in order: `AiStudioBadge`, title, `_creditsPill` (`:407`, opens `_showCreditsInfo`), history (`AiHistoryPage`), new chat, close.
- Credits come from `AiLimits` (`domain/entities/ai_limits.dart`: used, available and total for text, image and video). It is fed only from stream events (`application/ai_assistant_controller.dart:365,539`), with no dedicated fetch.
- Brand toggle: `presentation/widgets/ai_brand_voice_toggle.dart`, using the endpoints in `data/ai_assistant_endpoints.dart:19-20`.
- API layer: `data/ai_assistant_endpoints.dart`, `data/ai_assistant_service.dart`, `domain/repositories/ai_assistant_repository.dart`, with Riverpod controllers in `application/`.
- The role is available on `lib/features/workspaces/domain/workspace_member.dart:36`, exposed via `workspaces/application/workspace_controller.dart:44`.
- Settings pattern: `lib/features/settings/presentation/*_screen.dart` plus `application/*_controller.dart` plus `data/settings_service.dart`.
- **There are no Skills or Brand Knowledge screens in Flutter**, so those links must open the web app.

## B5. Integration points

| Piece | Where |
|---|---|
| Settings button | Next to history in `ChatHeader.vue:~49` and `AiStudioHeader.vue:~30`; Flutter `ai_assistant_view.dart:393` |
| Remove the credits chip | `ChatHeader.vue:61`, `AiStudioHeader.vue:42`; Flutter `_creditsPill` |
| Instructions into prompts | PHP `buildWorkflowInputs()` → agents `instructions_block` near `streaming_router.py:1771`, `team_session_state` `:2449`, carry-over `:2425`, member helper `agents.py:1972-1994` |
| Memory read | Same path as instructions, as a `memory_block` |
| Memory write | New. Async extraction after a run (Kafka, like `ai-agents.chat.usage`) into a workspace-keyed store. Post-results memory needs a hook from analytics or published posts |
| Usage breakdown | The usage ledger from the Usage visibility epic, or extend `AiChatUsage` and add ledger writes in the `PlanHelper::deduct*` callers |
| Deletion | `DeleteWorkspaceDataJob`, member removal in `TeamController`, plus agent-side copies |

## B6. Technical considerations

- **Storage:**
  - Keep settings and memory in backend MongoDB, e.g. `ai_chat_settings` (workspace instructions plus memory toggles), per-user personal instructions, and `ai_workspace_memories`.
  - The agents' Redis session store is not durable, and Agno memory is keyed by user, so neither fits workspace memory as-is.
- **Prompt budget:** cap instructions (proposal: 3,000 characters for workspace and 1,500 for personal). Cap the memory injected per run, with the most relevant and newest first.
- **Untrusted input:** inbox and comment text must never be written to memory word for word. Only aggregate patterns are allowed.
- **Caching:** the plan store drives the meters. The breakdown needs a new endpoint and should refresh after each chat turn.
- **Scope:** instructions and memory apply to AI Chat first. Applying them to Composer AI, AI Library and inbox auto-replies is a later decision.

## B7. Open technical questions (carried into the workflow)

1. What does memory store and how is it written: an Agno MemoryManager (user-keyed) or a custom workspace store written after each run? Recommendation: a custom workspace store in the backend, written by the agents after a run.
2. What data does "Learn from post results" use (the analytics service or published post metrics), and how often does it refresh?
3. Should instructions and memory reach non-chat AI surfaces, or chat only?
4. Should the usage breakdown be backfilled from `ai_chat_usage`, or count only from a new ledger going forward?
5. What is the prompt precedence between workspace instructions, personal instructions, brand voice and an active skill?
6. Who can see the Usage tab, given that billing access is restricted?
7. Are the memory toggles set per workspace, per user, or both?
