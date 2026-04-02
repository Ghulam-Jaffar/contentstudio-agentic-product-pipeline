# AI Agents Architecture and Platform Findings

Date: April 2, 2026

Audience: CTO, CEO, Engineering, Product Leadership

Scope: Static code analysis of `contentstudio-ai-agents`, with emphasis on architecture, request flow, data handling, streaming, model management, localization, performance, framework usage, duplication, and implementation gaps.

## Executive Summary

The current AI agents platform is not primarily limited by prompt quality or by any one model vendor. It is limited by architecture.

The codebase has grown into several parallel systems inside one service:

- a large custom orchestration layer for chat, content generation, and MCP workflows
- a separate streaming layer with its own event logic and execution branching
- a background media-job system using Postgres, Redis, and Kafka
- a Temporal-based reel workflow that is only loosely connected to the rest of the runtime model
- a large analytics surface built mostly through repeated per-agent and per-router implementations

At a high level, the product already offers broad capability, but the platform underneath it is inconsistent in the places that matter most:

- orchestration is custom and oversized
- true token streaming is partial, not a system-wide contract
- model selection is only partly centralized and partly ignored
- workflow and stream state are split across database JSON blobs, in-process caches, and global singletons
- localization assets exist, but runtime usage is inconsistent
- analytics and router code contain extensive duplication
- deprecated code and implementation leftovers remain in active runtime paths

This is why the team is seeing a mix of problems that look unrelated on the surface but share the same root causes:

- slow or inconsistent response times
- “streaming” that is not actually token streaming
- model behavior differences across features
- localization gaps and English-only fallbacks
- features that look implemented in API shape but not fully implemented in runtime behavior
- difficult debugging, hard-to-predict regressions, and rising maintenance cost

## Bottom Line

The AI agents service needs to be treated as a platform refactor and standardization effort, not as a sequence of isolated bug fixes.

The highest-leverage issue is architectural: the core runtime has no single, explicit execution contract for routing, streaming, model resolution, localization, and state management. Until those contracts are made real and enforced, feature work will continue to reintroduce inconsistency.

## Current Architecture

### 1. API Surface

`src/api/main.py` wires a large number of routers into one FastAPI application:

- unified content generation
- streaming generation
- inbox reply
- image and caption tools
- AI library flows
- analytics by platform
- workflow and job endpoints

This is a broad surface area for one service, but the important detail is that the runtime patterns behind those endpoints are not uniform.

### 2. Core Interactive Flow

The main interactive generation path is driven by `src/api/routers/streaming_router.py` and `src/teams/intelligent_router.py`.

Observed execution flow:

1. normalize request payload
2. optionally analyze newly attached images
3. load MCP workflow state for the session
4. run intent analysis
5. validate credits and optimize some media parameters
6. optionally run web research
7. execute the selected tool
8. translate results into SSE events

That is a large amount of work before the user receives a meaningful result.

### 3. State and Data Handling

The platform uses several different state layers at once:

- Postgres via SQLAlchemy for jobs, reel jobs, and MCP workflow sessions
- Redis and Kafka for background media lifecycle events
- in-memory global dictionaries for active stream cancellation and stream context
- in-memory singleton router and singleton analytics agents inside API workers
- request-scoped state cache inside the custom MCP agent
- external ContentStudio MCP server for platform operations

This means state is partly durable, partly ephemeral, and partly worker-local.

### 4. Workflow Handling

Workflow continuity is mostly managed through custom logic:

- `mcp_workflow_sessions.workflow_state` stores opaque JSON state
- `ContentStudioMCPAgent` loads and updates workflow state directly
- `IntelligentRouter` contains workflow continuation and guided-flow logic inline
- Temporal is used for reel workflows, but not as the primary orchestration model for the broader platform

### 5. AI Framework Usage

The repository presents itself as “built with the Agno framework,” but actual runtime usage is much narrower than that positioning suggests.

Observed use is mostly:

- `agno.agent.Agent`
- provider model wrappers
- `agno.tools.mcp.MCPTools`

The main orchestration layer is still handwritten application logic, not a cohesive Agno workflow/team architecture.

## Quantitative Indicators

Several complexity indicators support the conclusion that the orchestration surface is oversized and fragmented:

- `src/teams/intelligent_router.py`: 11,542 lines
- `src/api/routers/streaming_router.py`: 3,907 lines
- `src/agents/tools/contentstudio_mcp.py`: 1,876 lines
- `src/agents/image/image_generator.py`: 2,707 lines
- `src/agents/content/caption_writer.py`: 933 lines
- `src/agents/video/video_generator.py`: 838 lines
- analytics agent files: 91
- analytics router files: 9
- analytics agents overriding `_setup_agents()`: 82
- files directly importing provider model classes: 99
- files using localization helpers: 14
- files containing Agno `Team` or workflow-style constructs in `src/`: 1

These numbers do not prove correctness issues by themselves, but they do show that the platform has accumulated significant orchestration and duplication debt in its most critical layers.

## Main Findings

### 1. The orchestration layer is too large and owns too many concerns

The most important runtime logic is concentrated in a few oversized files:

- `IntelligentRouter` handles intent routing, workflow continuation, research enhancement, tool execution, model selection handoff, media composition flows, and error handling
- `streaming_router` handles request normalization, pre-analysis, workflow checks, intent analysis, credit validation, tool-specific SSE behavior, cancellations, and completion events
- `ContentStudioMCPAgent` handles connection management, state persistence, request caching, and MCP tool execution

This is not just a style problem. It creates concrete risks:

- changes are harder to reason about
- behavior differs across adjacent code paths
- feature additions increase regression risk disproportionately
- testing becomes weaker because most logic is integration-heavy and branch-heavy

### 2. Agno is being used mostly as a model wrapper layer, not as the true orchestration framework

The codebase and README position the system as Agno-based, but runtime orchestration is mostly custom.

What is actually visible:

- widespread use of `Agent`
- limited use of `MCPTools`
- almost no use of Agno workflow/team abstractions in active runtime code
- `router_team.py` exists, but it is not the main production orchestration path

This means the platform is paying the complexity cost of a custom orchestration layer while not fully benefiting from a framework-driven orchestration model.

### 3. State handling is fragmented across durable and non-durable layers

The service mixes persistent and worker-local state:

- jobs and reel jobs are stored in Postgres as JSON-heavy records
- MCP workflow state is stored as JSON blobs in `mcp_workflow_sessions`
- stream cancellation and stream context live in global in-memory dictionaries
- request-scoped workflow state is cached inside the MCP agent
- singleton router and singleton agent instances exist per worker process

This has two consequences:

- behavior is harder to make deterministic across multiple workers
- there is no single versioned source of truth for the full interactive runtime state

This is especially important because the README and startup flows explicitly support multi-worker API execution, while stream cancellation and some stateful behavior are process-local.

### 4. True token-based streaming exists only for part of the platform

The strongest true streaming implementation is in caption generation:

- `CaptionWriterAgent` streams text directly from the model
- `stream_events.stream_caption()` converts those chunks into SSE `caption_chunk` events

But this is not the platform-wide behavior.

Observed gaps:

- many assistant-style responses are not token streamed; they are generated first and then emitted word-by-word via `split_preserving_html()` with artificial delays
- image and video flows are status/progress streams, not token streams
- MCP flows are progress plus post-hoc assistant chunk simulation, not true model token streams
- the streaming endpoint is a mix of true token streaming, status streaming, and simulated typing

This means “streaming” currently describes multiple incompatible behaviors under one API surface.

### 5. There is at least one concrete streaming contract mismatch in combined flows

The combined writing flows do not agree on what a streamed caption looks like:

- `_execute_combined_tool()` collects string chunks and reconstructs a `CaptionOutput`
- `_execute_writing_and_video_tool()` assumes the stream yields `CaptionOutput` objects
- the caption writer’s streaming path actually yields text chunks

That mismatch is a concrete implementation gap and likely contributes to brittle behavior in some combined generation paths.

### 6. Slow response times are structurally likely

The latency problem is architectural, not only provider-related.

Before a user sees a useful result, the system can perform several of these steps:

- attachment analysis
- workflow state lookup
- intent analysis
- guided-flow detection
- credit validation
- parameter optimization
- pre-generation research
- content generation
- image or video generation

In other words, multiple expensive decisions and tool calls are stacked in series. This is visible in both `streaming_router` and `IntelligentRouter`.

The resulting issues are predictable:

- slow time-to-first-token
- highly variable latency by feature path
- more provider spend than necessary
- more failure points before the user sees value

### 7. Model architecture is only partially centralized and partly non-functional

The codebase does contain a central model configuration layer:

- `src/utils/config.py`
- `src/utils/model_registry.py`
- `src/utils/router_model.py`

But runtime adoption is inconsistent.

Observed problems:

- many files instantiate provider-specific model classes directly
- many analytics agents override `_setup_agents()` and choose their own provider/model logic
- model naming is inconsistent across defaults, legacy config fields, and registry keys
- `BaseAgent.run_async()` exposes `model_choice` and `fallback_choice`, but does not actually switch the active agent based on those values

This creates a platform where model-selection features exist in configuration and API shape, but not always in execution behavior.

### 8. Base abstractions exist, but leaf implementations bypass them too often

`BaseAgent` provides useful primitives:

- model setup
- fallback handling
- optional speed model usage
- memory support
- observability hooks

However, many leaf agents override setup and rebuild provider selection themselves. The analytics area is the clearest example.

This weakens the whole purpose of a base abstraction:

- fixes do not propagate cleanly
- model behavior drifts by feature
- tests and tracing become less comparable across agents

### 9. Localization assets are present, but runtime localization is inconsistent

The locale files themselves are in relatively good shape across supported languages. The real issue is runtime adoption.

Observed behavior:

- locale helper usage is sparse relative to codebase size
- many API routes and error paths still raise English-first `HTTPException` responses
- analytics often rely on prompt instructions such as “Generate all insights in {language}” instead of shared localization helpers
- streaming events use localization helpers in some paths, but other runtime messages remain hardcoded

So the localization problem is not “we have no translations.” It is “localization is not enforced as a runtime contract.”

### 10. Analytics is heavily duplicated and structurally expensive to maintain

The analytics area is one of the largest sources of duplication:

- 91 analytics agent files
- 9 analytics router files
- 82 custom `_setup_agents()` overrides
- repeated request/response models
- repeated singleton getters
- repeated prompt structure
- repeated provider fallback logic

This creates predictable problems:

- model behavior differs by analytics feature
- localization and error handling differ by analytics feature
- fixes have to be applied many times
- onboarding and debugging cost rise quickly

This area should be treated as a consolidation candidate, not as a long-term stable pattern.

### 11. Dead code, deprecated code, and implementation leftovers are still present in the runtime tree

The codebase still contains several signals of unfinished or superseded implementation:

- `contentstudio_mcp_backup.py`
- deprecated `image_editor.py`
- deprecated redirect logic from `image_edit_tool` to `creative_image_tool`
- a `/stream/test` SSE endpoint in the main streaming router
- runtime debug prints inside production paths
- TODO markers inside `reel_workflow.py` for still-incomplete workflow responsibilities

None of these items alone causes the platform issues, but together they show that the runtime tree still contains ambiguity about what is current, deprecated, or transitional.

### 12. The platform currently uses multiple runtime paradigms without one unified contract

The service currently contains all of the following:

- interactive SSE generation
- background media jobs with Redis and Kafka events
- MCP workflow continuity persisted in Postgres JSON
- a Temporal workflow for reels
- per-feature direct agent execution

These are all defensible individually. The problem is that they do not appear to share one unified event model, state model, or observability model for product behavior.

That is why two AI features in the same product can feel operationally very different.

### 13. Prompt and message handling are more implicit than they should be

There are several hidden prompt-shaping behaviors in foundational code:

- `BaseAgent.run_async()` prepends current server date and time to every request
- chat history is flattened into plain text in some base flows
- analytics localization is often handled as a prompt instruction rather than as a shared response contract
- some paths expect structured outputs while others expect plain strings and then reshape them later

This makes debugging harder because a model response depends not only on the visible feature code, but also on hidden prompt mutation in shared layers.

## Root Causes

The major root causes can be summarized as follows:

### 1. The platform grew feature-by-feature without hard platform contracts

Routing, streaming, localization, model resolution, and workflow state evolved in parallel rather than under one enforced architecture.

### 2. Reusable abstractions exist but are not strongly adopted

Base agents, model registry, localization helpers, and workflow persistence primitives exist, but many leaf modules bypass them.

### 3. Custom orchestration absorbed too much responsibility

Instead of smaller explicit services or framework-driven workflow primitives, too much behavior was accumulated inside `IntelligentRouter`, `streaming_router`, and the custom MCP agent.

### 4. Multiple runtime models coexist without one governing execution model

Chat SSE, async jobs, MCP workflows, analytics generation, and Temporal workflows are all valid patterns, but they are not operating under one clear platform contract.

### 5. Growth outpaced simplification

The codebase clearly shipped a lot of capability quickly. The simplification and standardization pass did not keep up.

## Business and Engineering Impact

If left as-is, the current architecture will continue to create:

- slower and more inconsistent user experience
- difficulty delivering truly realtime or truly streaming AI interactions
- higher debugging and support cost
- unpredictable model behavior between features
- incomplete localization despite translation investment
- higher provider cost because of stacked AI calls
- slower onboarding for new engineers
- higher regression risk when expanding agentic features or analytics

This is not only a code quality issue. It directly affects product trust, delivery speed, and the team’s ability to scale AI features confidently.

## Recommended Plan

### Phase 1. Define and Enforce Core Platform Contracts

Goal: make the runtime predictable.

Recommended actions:

- define one canonical interactive request/response contract
- define one explicit streaming contract with distinct modes: token stream, progress stream, completion, cancellation
- make model selection contract real end-to-end, not just configurable
- standardize localization contract for API errors, SSE events, and assistant-visible system messages
- remove or clearly isolate deprecated and backup runtime code

Expected outcome:

- less ambiguity in behavior
- faster debugging
- fewer “implemented in shape but not in reality” gaps

### Phase 2. Decompose the Orchestration Layer

Goal: shrink the blast radius of changes.

Recommended actions:

- split `IntelligentRouter` into smaller execution services
- separate intent routing, workflow continuation, research enhancement, and tool execution
- split `streaming_router` into normalization, orchestration, and transport/event layers
- move MCP workflow handling into a clearer service or state-machine abstraction
- reduce worker-local mutable state where possible

Expected outcome:

- better testability
- less branch-heavy code
- lower regression risk

### Phase 3. Build Real Streaming as a First-Class Capability

Goal: stop mixing true streaming with simulated streaming.

Recommended actions:

- start emitting user-visible output after minimal validation
- remove artificial word-by-word typing for paths that are not truly token streamed
- define which tools genuinely support token streaming and expose that explicitly
- unify combined caption/media flows around one streaming payload contract
- measure and optimize time-to-first-token across major flows

Expected outcome:

- materially better perceived speed
- clearer product behavior
- fewer frontend assumptions and fewer streaming edge cases

### Phase 4. Unify Model Resolution and Execution

Goal: make model behavior consistent and governable.

Recommended actions:

- require registry-backed model resolution for all text, image, and video selections
- stop per-agent ad hoc provider instantiation except where genuinely necessary
- remove or rewrite APIs that imply model choice but do not actually switch models
- standardize fallback behavior and speed tiers across the codebase

Expected outcome:

- fewer model inconsistencies
- simpler governance of cost, quality, and latency
- easier experimentation and rollout control

### Phase 5. Consolidate Analytics into a Reusable Framework

Goal: replace repetition with configurable reuse.

Recommended actions:

- create a generic analytics insight engine with shared prompt templates, model policy, and schema handling
- reduce per-platform routers through factories or shared handlers where possible
- centralize localization and response shaping for analytics
- eliminate repeated `_setup_agents()` implementations unless there is a real business need

Expected outcome:

- lower maintenance cost
- more consistent analytics quality
- much easier cross-platform improvements

### Phase 6. Clarify the Real Orchestration Strategy

Goal: decide what the platform actually is.

Recommended actions:

- decide when to use Agno, when to use Temporal, and when custom orchestration is justified
- if Agno is intended to be central, use its workflow/team capabilities more deliberately
- if custom orchestration remains primary, simplify claims and structure the codebase accordingly
- unify event/state/trace taxonomy across chat, jobs, and workflows

Expected outcome:

- clearer architectural direction
- less accidental complexity
- better alignment between platform narrative and runtime reality

## Suggested Success Metrics

To confirm improvement, the team should track:

- p50 and p95 time to first token for interactive text responses
- percentage of streaming requests using true token streaming vs simulated typing
- average number of AI/tool hops per interactive request path
- count of direct provider model imports over time
- count of files bypassing shared model-selection utilities
- count of files bypassing localization helpers for user-facing API messages
- cancellation success rate in multi-worker environments
- analytics module count and duplicated setup patterns over time
- response-contract validation failures in API and SSE payloads

## Representative Technical Examples

The following examples are especially representative of the current gaps:

- caption streaming is genuinely token-based in `caption_writer.py`, but many assistant and MCP responses in `streaming_router.py` are emitted after generation via word splitting
- `stream_cancellation.py` stores active streams and context in global in-memory dictionaries
- `BaseAgent.run_async()` exposes `model_choice` and `fallback_choice`, but does not switch the underlying agent instance based on those values
- `ContentStudioMCPAgent` combines DB-backed workflow state with request-scoped in-memory cache inside the same execution model
- `_execute_combined_tool()` and `_execute_writing_and_video_tool()` disagree on the streamed caption payload type
- analytics agents repeatedly override `_setup_agents()` with custom provider logic instead of consistently inheriting platform model behavior

## Representative Evidence

The findings above are based on the following code paths.

### Core Platform

- `contentstudio-ai-agents/README.md`
- `contentstudio-ai-agents/src/api/main.py`
- `contentstudio-ai-agents/src/api/routers/streaming_router.py`
- `contentstudio-ai-agents/src/api/helpers/stream_events.py`
- `contentstudio-ai-agents/src/api/helpers/stream_cancellation.py`
- `contentstudio-ai-agents/src/teams/intelligent_router.py`
- `contentstudio-ai-agents/src/teams/mcp/unified_intent_detector.py`
- `contentstudio-ai-agents/src/teams/router_team.py`
- `contentstudio-ai-agents/src/agents/base.py`
- `contentstudio-ai-agents/src/agents/tools/contentstudio_mcp.py`

### Configuration and Localization

- `contentstudio-ai-agents/src/utils/config.py`
- `contentstudio-ai-agents/src/utils/model_registry.py`
- `contentstudio-ai-agents/src/utils/router_model.py`
- `contentstudio-ai-agents/src/locales/helpers.py`

### Persistence, Jobs, and Events

- `contentstudio-ai-agents/src/db/models.py`
- `contentstudio-ai-agents/src/db/repositories/workflow_repository.py`
- `contentstudio-ai-agents/src/events/publisher.py`
- `contentstudio-ai-agents/src/events/schemas.py`
- `contentstudio-ai-agents/src/jobs/video_tasks_dramatiq.py`
- `contentstudio-ai-agents/src/workflows/reel_workflow.py`

### Analytics Surface

- `contentstudio-ai-agents/src/agents/analytics/...`
- `contentstudio-ai-agents/src/api/routers/analytics/...`
- `contentstudio-ai-agents/src/agents/analytics/facebook/facebook_page_engagement.py`
- `contentstudio-ai-agents/src/api/routers/analytics/linkedin_analytics_router.py`

## Notes and Limitations

This document is based on source-code analysis, not on live production traces, latency profiling, or database/event inspection in a running environment.

That limitation matters for exact magnitude, but not for direction. The core findings are architectural and are visible directly in the runtime design and implementation patterns.

The recommended next step is to pair this document with:

- request tracing for real production paths
- time-to-first-token measurement by feature
- provider and model usage telemetry
- error-rate breakdown by endpoint and tool path

## Recommendation for Leadership and Team

Treat the AI agents service as a shared platform that now needs standardization before more feature expansion.

A reasonable framing is:

- Phase 1 restores runtime predictability
- Phase 2 reduces architectural blast radius
- Phase 3 makes streaming and responsiveness credible
- Phase 4 and Phase 5 reduce long-term maintenance cost
- Phase 6 aligns the technical architecture with the product’s AI strategy
