# 01 Research — MCP authentication and ChatGPT / Claude connector submission

Date: 2026-08-23
Scope: how the ContentStudio MCP server authenticates callers today, and what stands between that and being listed as a connector in ChatGPT and in Claude.

This doc is the local grounding for the epic. Nothing here goes into the story body.

---

## 1. Where MCP lives today

Two MCP servers exist, which is itself part of the problem:

1. **Laravel MCP server**, inside `contentstudio-backend`:
   - Tools: `app/Mcp/Tools/` — `CreatePostTool`, `DeletePostTool`, `FetchPostsTool`, `FetchWorkspacesTool`, `FetchSocialAccountsTool`, `ValidateTokenTool`, `HelpTool`, `PingTool`
   - Resource: `app/Mcp/Resources/WorkspacesResource.php`, prompt: `app/Mcp/Prompts/GettingStartedPrompt.php`
   - Transport: `routes/mcp.php` and `app/Http/Controllers/CustomMcpController.php`, exposing `api/mcp`, `api/mcp/sse`, `api/mcp/message`
   - Config: `config/mcp.php`
2. **Node MCP server** published to npm as `contentstudio-mcp`, plus the CLI and agent skill. Not mounted in this project.

The existing epic `ai-surfaces-architecture` already covers the overlap between these two servers, the CLI and the in-app AI chat, and its first story is `[Research] Map the current MCP, CLI/skills, and AI-chat architecture`. **This epic must not re-litigate which server survives.** It handles authentication and the directory submissions, and takes the "which server" answer from that epic.

## 2. How authentication works today, and why it blocks a connector listing

The Laravel MCP routes are registered with only the `api` middleware group. There is no authentication middleware on them. Instead, **each tool takes the API token as a tool parameter**:

```php
#[ToolParameter(description: 'ContentStudio API token (required)')]
```

with a `mcp_tools.api_key_required` error when it is absent.

That has three consequences:

1. **The credential is a model-visible argument.** The token travels through the model's context as tool input, appears in tool-call logs, and is at the mercy of whatever the client does with conversation history. A long-lived API key is the worst possible thing to put there.
2. **There is no user-level authorization boundary.** `app/Http/Middleware/ApiKeyMiddleware.php` exists for the public API and does resolve the key to a user and merge the workspace from the route, but the MCP routes do not use it. The tools do their own token handling.
3. **It is not what a connector directory accepts.** Both ChatGPT and Claude connect to remote MCP servers on the user's behalf, and both expect the server to run a standards-based authorization flow so the end user grants access in a browser. A server that expects an API key pasted into a tool argument cannot be listed, regardless of how well the tools work.

There is no authorization server today. `grep` finds no Passport, no OAuth authorization endpoints, no discovery metadata documents. Every `oauth2` hit in config is ContentStudio acting as an OAuth *client* against a social network. Authentication for the public API is API keys only, in the `api_keys` Mongo collection.

## 3. What the research ticket has to establish, not assume

The MCP authorization specification and both directories' listing requirements have moved repeatedly, so **the ticket must confirm the current published requirements at the time of the work rather than relying on this document**. What to confirm:

- The current MCP authorization specification: the authorization flow required of a remote MCP server, the discovery metadata a client fetches, whether client registration must be dynamic, and how the resource the token is issued for is identified.
- ChatGPT: the current path for a third-party connector or app, the technical requirements, the review and submission process, and any verification or policy requirements on the publisher.
- Claude: the same for the Claude connectors directory.
- Where the two differ, so we build once and satisfy both.
- Whether either directory requires specific behaviour of the tools themselves: naming, descriptions, read-only versus write tools, confirmation before destructive actions, rate limits.

## 4. Design questions specific to ContentStudio

1. **Workspace selection.** A ContentStudio user usually belongs to several workspaces. An OAuth grant is per user. So either the grant covers the workspaces the user can access and the tools take a workspace argument, or the user picks a workspace during the connect flow and the grant is scoped to it. This decision shows up in every tool signature.
2. **Roles and permissions.** The main API enforces the user's role per workspace. The MCP tools currently do their own thing. A connector acting for a user must not be able to do more than that user can do in the app, and this is the same gap the service authorization epic is addressing on other services.
3. **Write actions.** `CreatePostTool` and `DeletePostTool` are destructive from a user's point of view. Scope granularity, and whether a connector can publish or only draft, is a product decision, not only a technical one.
4. **Existing API keys.** Integrations use API keys today, and the CLI and npm MCP server rely on them. Adding a user-authorization flow must not break them, so the server needs to accept both for a period.
5. **Consumption and limits.** MCP calls hit the same features that consume credits. Existing related work: the AI usage visibility epic, and `x-pay-per-use-credits`.
6. **White-label.** White-label customers run under their own domains. Whether they get their own connector listing, or are excluded, is a product decision.

## 5. Adjacent existing deliverables (do not duplicate)

- `ai-surfaces-architecture` — owns the "which MCP server, one tool definition source" question
- `ai-tools-and-skills-platform` — owns the catalogue of capabilities exposed to AI
- `contentstudio-public-cli-agent-skills`, `docs/technical/public-cli-strategy-2026-04-08.md`
- `docs/technical/ai-agents-architecture-and-platform-findings-2026-04-02.md`
- Memory note: the public API base URL, npm packages and MCP endpoint that are live today are recorded in the reference memory on public dev surfaces.
