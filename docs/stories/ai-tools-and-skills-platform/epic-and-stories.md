# Epic E — AI tools & skills (reusable AI capability across the product)

## Problem

ContentStudio has started adding AI actions (content generation, scheduling) but they are built ad hoc. To scale AI across the product, we need a deliberate set of **AI tools and skills** built as a reusable capability, not tied to any single feature. Today there is no shared catalog or framework, so every new AI feature risks re-implementing the same actions.

## Goal

Build a catalog and framework of AI tools & skills that power AI features across ContentStudio. This is the core work. AI chat is one consumer of these tools, not the point of them. The same tools and skills should be reusable across many surfaces, for example:

- AI chat (already uses content generation and scheduling)
- Inbox (AI-assisted replies, triage, actions)
- Analytics AI insights
- Account, workspace, and publishing management
- and other AI-driven workflows as they arise

Research and plan first, then build in phases, with each tool defined once and reusable everywhere.

## Dependency

Depends on **Epic D — Architecture for MCP + CLI/Skills + AI chat**. The tools/skills here plug into the shared tool-definition and contract Epic D defines, so a tool built once is available to every consumer (chat, inbox, analytics, MCP, CLI).

## Stories (see `02-stories.md`)

1. `Research` Inventory the capabilities to expose as AI tools/skills, and the surfaces that will consume them
2. `Plan` AI tools/skills catalog and phased rollout (reusable across chat, inbox, analytics insights, management, ...)
3. Implementation stories per tool/domain — defined after the plan

## Sequencing

Research and plan first (with a review gate), then implement in phases. Each phase adds a coherent set of tools that multiple surfaces can immediately reuse. Reuse the shared architecture from Epic D rather than wiring tools per surface.

## Out of scope (until planned)

- Fully autonomous actions without user confirmation (keep the human-in-the-loop model the CLI/skill already use).
- Tools for domains not yet stable in the underlying API (build on supported surfaces first).



---

---


# Epic E stories — AI tools & skills (reusable capability)

## [Research] Inventory capabilities to expose as AI tools/skills, and their consuming surfaces

### Description
Catalog the ContentStudio capabilities that should become reusable AI tools/skills, note what already exists (content generation, scheduling), and map both the candidate tools and the surfaces that will consume them (AI chat, inbox, analytics AI insights, management, and others). Prioritize by value, effort, and readiness.

### Acceptance criteria
- [ ] A catalog lists candidate AI tools/skills grouped by domain (content, publishing lifecycle, inbox, analytics/insights, account and workspace management, approvals, listening, ...).
- [ ] Each tool is marked as already available, ready to build, or blocked (with reason).
- [ ] Each tool maps to its backing surface (public API endpoint, internal service, etc.).
- [ ] The consuming surfaces are listed for each tool (for example a "reply" tool used by both inbox and chat; an "insights" tool used by analytics and chat).
- [ ] Tools are prioritized (value vs effort vs readiness), and API gaps needed to support a tool are flagged.

### Implementation references
*Pointers from research, not a contract.*
- Existing AI tooling: `contentstudio-frontend/src/modules/AI-tools`, `contentstudio-ai-agents` (`src/agents`, `src/tools`, `src/orchestration`), backend dedicated tools `contentstudio-backend/app/Mcp/Tools` and `contentstudio-ai-agents/src/tools/dedicated`.
- Consuming surfaces to consider: AI chat (`AI-tools`), inbox (`social-inbox-manager` + inbox-revamp), analytics AI insights (`analytics_v3` / meta-google insights), management (settings/team/workspace, publishing lifecycle via the public API).

---

## [Plan] AI tools/skills catalog and phased rollout

### Description
Turn the inventory into a concrete plan: the tool/skill catalog with contracts, how each tool plugs into the shared architecture (Epic D) so multiple surfaces reuse it, the phasing, and the guardrails.

### Acceptance criteria
- [ ] The plan defines each tool/skill with a clear input/output contract, reusable across surfaces.
- [ ] Each tool maps onto the shared architecture from Epic D (single source of truth; available to chat, inbox, analytics, MCP, CLI at once).
- [ ] A phased rollout is defined (which tools/domains ship in which order, and which surfaces light up per phase).
- [ ] Guardrails are specified: confirmation before mutations, permission checks, workspace scoping.
- [ ] Adoption success metrics are proposed.
- [ ] Reviewed and approved before implementation stories are written.

### Implementation references
- Reuse the CLI/skill human-in-the-loop conventions (confirm before mutating, preview where possible). Keep tool contracts compatible with the public API as the backend source of truth.

---

## Implementation stories (per tool/domain)

To be authored after the plan is approved. Expected phases add reusable tools by domain (for example inbox, analytics/insights, management), each usable immediately by every consuming surface rather than wired per feature.
