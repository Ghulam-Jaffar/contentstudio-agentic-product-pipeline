# Using ContentStudio with Claude Code (CLI + Agent Skill)

A practical guide to running ContentStudio from Claude Code, so you can manage social media by chatting with Claude while it drives the ContentStudio CLI for you. Includes setup and a set of real, copy-pasteable use cases.

> Reflects what is live today: the `contentstudio-cli` npm package, the `contentstudio` agent skill (`contentstudioio/contentstudio-agent`), and the `contentstudio-mcp` server. All are built on the ContentStudio public API. Scope today is publishing: posts, media, approvals, comments, and account/workspace discovery. Analytics and inbox are not part of the CLI yet.

---

## How the pieces fit

Three things, each with a clear job:

- **CLI (`contentstudio-cli`)** is the actual tool Claude Code runs as bash commands (`contentstudio ...`).
- **Agent skill (`contentstudio-agent`)** is a `SKILL.md` that teaches Claude Code how to use the CLI well: which command to pick, using `--json`, previewing changes with `--dry-run`, confirming the workspace before anything is published, and handling pagination. Claude Code can run the CLI without it, but the skill makes it reliable and hands-free.
- **MCP server (`contentstudio-mcp`)** is an alternative to the CLI. Instead of shelling out, Claude Code talks to ContentStudio through native MCP tools. Use the CLI plus skill for the lighter, scriptable path, or MCP for the native-tool path. Either works.

All paths need two things: the tool installed, and a ContentStudio API key.

---

## Prerequisites

- Node.js installed (for `npm` / `npx`).
- A ContentStudio API key. Generate one in the ContentStudio dashboard: **Settings → API Keys → Generate** (it starts with `cs_`). You need a plan that includes API access.

---

## Setup

### 1. Install the CLI
```bash
npm install -g contentstudio-cli
```

### 2. Authenticate
```bash
# Option A: store the key in the CLI config (simplest for interactive use)
contentstudio auth:login --api-key cs_xxxxxxxx

# Option B: environment variable (best for headless / CI)
export CONTENTSTUDIO_API_KEY=cs_xxxxxxxx
```

### 3. Install the agent skill into Claude Code (recommended)
```bash
npx skills add contentstudioio/contentstudio-agent
```
This installs the skill so Claude Code auto-discovers it (it lands in `.agents/skills/` and is symlinked into `.claude/skills/`). Run it inside a project for that project, or globally to make it available everywhere.

### 4. Allow the command so Claude Code does not prompt every time
Add to `.claude/settings.json` (project) or `~/.claude/settings.json` (personal):
```json
{ "permissions": { "allow": ["Bash(contentstudio:*)"] } }
```

### 5. (Optional) MCP route instead of the CLI
```bash
claude mcp add contentstudio --env CONTENTSTUDIO_API_KEY=cs_xxxxxxxx -- npx -y contentstudio-mcp
```

### Verify it works
```bash
contentstudio auth:status
contentstudio --json workspaces:list
contentstudio workspaces:use <workspace_id>
```
Or just ask Claude Code: "Check my ContentStudio auth and list my workspaces."

---

## How to read the use cases

Each use case shows what **you say** to Claude Code and what **Claude Code does** under the hood. Exact flags are visible with `contentstudio <command> --help` and at the API guide (api.contentstudio.io/guide). The commands below are illustrative of the shipped, colon-style command surface.

The safety model is built in: before publishing or changing anything, Claude Code (guided by the skill) confirms which workspace it is acting on and previews the action with `--dry-run` first. It never treats the current page of a list as "everything" when more results exist.

---

## Use cases

### Everyday publishing

**1. See what you can post to**
> You: "What ContentStudio accounts can I post to right now?"

Claude Code runs discovery and summarizes:
```bash
contentstudio auth:status
contentstudio --json workspaces:current
contentstudio --json accounts:list
```

**2. Draft a post for review**
> You: "Draft a LinkedIn post announcing our new pricing and save it as a draft."

Claude writes the copy, then:
```bash
contentstudio posts:create -c "<generated copy>" --account <linkedin_id> --publish-type draft
```

**3. Schedule to several channels at once**
> You: "Schedule this to Facebook and Instagram for 9am tomorrow."

Claude confirms the workspace, previews, then creates it:
```bash
contentstudio posts:create -c "<copy>" -i <fb_id> -i <ig_id> \
  --publish-type scheduled --schedule "2026-07-11 09:00:00" --dry-run
# then the same command without --dry-run once you confirm
```

**4. Publish now**
> You: "Post this to our X account right now."
```bash
contentstudio posts:create -c "<copy>" --account <x_id> --publish-type now
```

**5. Put the link in the first comment**
> You: "Schedule the post, and add the blog link as the first comment."

Claude schedules the post, then adds the comment to the created post id:
```bash
contentstudio comments:add <post_id> "Full article: https://example.com/post"
```

### Agentic, multi-step workflows (where Claude Code shines)

**6. Turn a blog post into a week of social posts**
> You: "Read this blog post and turn it into a week of platform-specific posts, saved as drafts."

Claude reads the source (a file or URL), writes tailored variants for each network, then creates one draft per platform per day so you can review before anything goes live.

**7. Repurpose a newsletter into a LinkedIn post and an X thread**
> You: "Repurpose my latest newsletter into a LinkedIn post and an X thread, queue them as drafts."

Claude extracts the key points, drafts both formats, and creates the drafts via `posts:create --publish-type draft`.

**8. Bulk-upload media and pair with captions**
> You: "Upload every image in ./campaign-assets to my library, then draft a post for each with a matching caption."
```bash
# per file:
contentstudio media:upload --file ./campaign-assets/1.jpg
# then a draft post that references the uploaded media
```

**9. Audit what is scheduled and flag problems**
> You: "List everything scheduled this week and flag any post missing a link or an image."

Claude pulls the posts as JSON and analyzes them, walking pagination so nothing is missed:
```bash
contentstudio --json posts:list
```

**10. Clear the review queue with you in the loop**
> You: "Show me the posts waiting for approval, and approve the ones that look good."

Claude lists pending posts, you decide, and it approves or rejects by id:
```bash
contentstudio posts:list --json
contentstudio posts:approve <post_id>
contentstudio posts:reject <post_id>
```

### Agencies and multi-client ops

**11. Do the same thing across every client workspace**
> You: "Draft this announcement as a draft in all my client workspaces."

Claude lists workspaces and repeats the action, targeting each with `--workspace` so it never publishes to the wrong client:
```bash
contentstudio --json workspaces:list
contentstudio posts:create -c "<copy>" --account <id> --publish-type draft --workspace <client_id>
```

**12. Build a report of what is scheduled**
> You: "Give me a CSV of everything scheduled this month across all accounts."

Claude pulls posts as JSON across workspaces and formats a CSV for you.

### Developer and automation workflows

**13. Generate a scheduling script from a spreadsheet**
> You: "Write me a Node script that reads content.csv and schedules each row with the ContentStudio CLI."

Claude writes the script using `contentstudio posts:create`, including `--json` parsing and error handling.

**14. Auto-announce on release from CI**
> You: "Write a GitHub Action that drafts a social post whenever I publish a new blog post."

Claude scaffolds the workflow, using `CONTENTSTUDIO_API_KEY` as a secret and the CLI to create the draft.

**15. Wire approvals into a chatbot**
> You: "Sketch a Slack command that lists pending ContentStudio posts and approves one by id."

Claude scaffolds it on top of `posts:list`, `posts:approve`, and `posts:reject`.

---

## Safety model (good to know)

- **Workspace confirmation before changes.** For any publishing or mutating action, Claude confirms which workspace it is acting on, even if one is already active.
- **Dry run first.** Mutating commands are previewed with `--dry-run` before the real call.
- **Human in the loop.** This is an assistant, not a hands-off autopilot. You stay in control of what goes live.
- **JSON and exit codes.** Agent-facing output is structured (`{"ok": true, "data": ...}` or an error envelope) with non-zero exit codes on failure, so Claude can act on results reliably.

---

## Troubleshooting

- **"command not found: contentstudio"** — the global install did not land on your PATH. Re-run `npm install -g contentstudio-cli`, or use `npx contentstudio-cli ...`.
- **Auth errors** — run `contentstudio auth:status`. If `has_api_key` is false, run `contentstudio auth:login --api-key cs_...` or set `CONTENTSTUDIO_API_KEY`. In headless runtimes (CI, daemons, Docker), a shell `export` does not persist to a service process; set the variable in the service environment and restart.
- **No workspace selected** — run `contentstudio --json workspaces:list`, then `contentstudio workspaces:use <id>`.
- **Claude Code keeps asking permission** — add `"Bash(contentstudio:*)"` to the `allow` list in `.claude/settings.json`.
- **Truncated results** — list commands paginate; ask Claude to fetch all pages, or pass paging flags. The skill is set up to not stop at page one.

---

## Not supported yet (do not expect these)

- Analytics commands (reporting, metrics).
- Inbox, CRM, or broadcast commands.
- Fully autonomous publishing with no review step.

---

## Reference

**Command groups:** `auth`, `workspaces`, `accounts` (connect, add-bluesky, add-facebook-group), `posts` (list, create, delete, approve, reject), `comments` (list, add), `media` (list, upload), plus lookups for `campaigns`, `categories`, `labels`, `team`, and `platforms`.

**Platforms:** Facebook, LinkedIn, Twitter/X, Instagram, YouTube, TikTok, Pinterest, Threads, Tumblr, Bluesky, Google Business Profile.

**Links:**

| Asset | Where |
|---|---|
| CLI package | npm: `contentstudio-cli` (`npm i -g contentstudio-cli`) |
| Agent skill | `npx skills add contentstudioio/contentstudio-agent` |
| MCP server | npm: `contentstudio-mcp`; bundle at contentstudio.io/mcp-server |
| API guide and reference | api.contentstudio.io/guide |
| Source repo | github.com/contentstudioio/contentstudio-agent |
