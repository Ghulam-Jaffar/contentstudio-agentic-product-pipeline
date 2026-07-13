# Epic D — Architecture for MCP + CLI/Skills + AI chat (research, plan, update)

## Problem

ContentStudio now exposes AI-operable surfaces through several paths that grew independently and overlap:
- A **Node MCP server** published to npm (`contentstudio-mcp`).
- A **Laravel php-mcp server** inside the backend (`app/Mcp/Tools`, served at `mcp.contentstudio.io/mcp`).
- The **CLI + agent skill** (`contentstudio-cli`, `contentstudio-agent`).
- The **in-app AI chat**, backed by the `contentstudio-ai-agents` platform, which itself connects to the MCP server.

Tool definitions, auth handling, and the action contract are duplicated across these, which risks drift (a tool exists in one surface but not another, or behaves differently). We need a deliberate, shared architecture before we expand AI control (Epic E).

## Goal

Research the current state, map the overlaps and drift risks, and propose and then implement a unified architecture where the CLI, MCP, and AI chat share a single source of truth for tool/action definitions, a consistent auth model, and a stable contract, so new capabilities land everywhere at once.

## Stories (see `02-stories.md`)

1. `Research` Map the current MCP, CLI/skills, and AI-chat architecture, overlaps, and drift risks
2. `Plan` Propose a unified architecture (single tool-definition source, shared client/contract, consistent auth, migration path)
3. Implementation stories — defined after the plan is approved

## Sequencing

Research first, then plan, then a review gate before any implementation stories are written. This epic is the foundation for Epic E (AI tools & skills for full control).

## Out of scope

- Building new end-user AI capabilities (that is Epic E). This epic is about the plumbing they will share.



---

---


# Epic D stories — AI surfaces architecture

## [Research] Map the current MCP, CLI/skills, and AI-chat architecture

### Description
Produce a clear map of how ContentStudio exposes actions to AI today across the Node MCP server, the Laravel php-mcp server, the CLI + agent skill, and the in-app AI chat (via the AI-agents platform). Identify overlaps, duplication, drift risks, and inconsistencies in tool definitions, auth, and contracts.

### Acceptance criteria
- [ ] A document lists every AI-operable surface and where its code lives, with the exact set of tools/commands each exposes.
- [ ] Overlaps and gaps are identified (tools present in one surface but missing/different in another).
- [ ] Auth models are compared across surfaces (API key handling, workspace scoping, permissions).
- [ ] Drift risks are called out (for example the same action defined twice with different behavior).
- [ ] The current data/flow is diagrammed (chat -> AI-agents -> MCP -> public API; CLI -> public API; MCP server(s) -> public API).
- [ ] Open questions and recommendations are captured to feed the planning story.

### Implementation references
*Pointers from research, not a contract.*
- Node MCP: `contentstudio-mcp` (npm, `github.com/d4interactive/contentstudio-mcp`).
- Laravel MCP: `contentstudio-backend/app/Mcp/Tools`, `config/mcp.php`, `routes/mcp.php`, `app/Http/Controllers/CustomMcpController.php`; served at `mcp.contentstudio.io/mcp`.
- CLI + skill: `contentstudioio/contentstudio-agent` (`contentstudio-cli`, `SKILL.md`).
- AI chat: `contentstudio-frontend/src/modules/AI-tools` and the `contentstudio-ai-agents` platform (`src/agents`, `src/orchestration`, MCP client in `src/agents/tools/contentstudio_mcp.py`).

---

## [Plan] Propose a unified AI-surfaces architecture

### Description
Based on the research, propose a target architecture where the CLI, MCP, and AI chat share a single source of truth for tool/action definitions, a consistent auth model, and a stable contract, plus a migration path from today's state.

### Acceptance criteria
- [ ] The plan defines a single source of truth for tool/action definitions that all surfaces consume.
- [ ] The plan resolves the dual-MCP situation (Node vs Laravel) with a clear recommendation.
- [ ] A consistent auth and workspace-scoping model is specified across surfaces.
- [ ] A migration path from the current state is described, with sequencing and risks.
- [ ] The plan states how new capabilities (Epic E) will be added once and appear on every surface.
- [ ] Reviewed and approved before implementation stories are written.

### Implementation references
- Feeds Epic E. Keep the contract compatible with the public API as the backend source of truth.

---

## Implementation stories

To be authored after the plan is approved (kept out of this draft deliberately, since scope depends on the chosen architecture).
