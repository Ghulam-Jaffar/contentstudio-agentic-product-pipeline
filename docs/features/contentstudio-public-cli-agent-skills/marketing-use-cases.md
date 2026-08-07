# ContentStudio CLI and Agent Skills: Use Cases for Marketing

A plain-language briefing for the marketing team. It covers what we shipped, who it is for, concrete use cases with real examples, positioning angles, content ideas, and accuracy guardrails so we do not over-claim.

> Everything below reflects what is live today: the `contentstudio-cli` npm package (v1.0.10), the `contentstudio` agent skill, and the `contentstudio-mcp` server. All three are public and grounded in the ContentStudio public API.

---

## What we shipped (in one minute)

We turned ContentStudio's public API into three developer-and-AI-friendly surfaces:

1. **A command-line tool (CLI).** Install with `npm install -g contentstudio-cli`, run the `contentstudio` command, and manage ContentStudio straight from a terminal or a script. No custom code needed.
2. **An AI agent skill.** One command (`npx skills add contentstudioio/contentstudio-agent`) teaches a shell-capable AI agent how to operate ContentStudio in plain English.
3. **An MCP server.** `contentstudio-mcp` connects ContentStudio to MCP-compatible AI apps such as Claude Desktop, so users can publish and manage posts by chatting, no terminal required.

**What it can do today:** discover workspaces and connected accounts, upload media, create and schedule posts, save drafts, list posts, approve or reject posts, add comments and internal notes, and audit campaigns, labels, categories, and team members. Across **11 platforms**: Facebook, LinkedIn, Twitter/X, Instagram, YouTube, TikTok, Pinterest, Threads, Tumblr, Bluesky, and Google Business Profile.

**Why it is credible for automation:** every command can return clean JSON, every risky action supports a `--dry-run` preview, and results paginate so nothing is silently truncated. It is built for scripts, cron jobs, CI pipelines, and AI agents, with a human-in-the-loop safety model.

---

## The headline positioning

Pick the line that fits the channel:

- "ContentStudio now runs in your terminal, and inside your AI agent."
- "Publish and manage social media from the command line, or let an AI agent do it for you."
- "The social media tool your scripts and your AI assistant can both use."

Differentiator to lean on: most major schedulers (Buffer, Hootsuite, Later) do not ship a public CLI, agent skill, and MCP server. This is a genuine "first-class developer and AI surface" story.

---

## Use cases

Each use case has: who it is for, the scenario, and a concrete example. Example commands are illustrative; exact flags are in `contentstudio --help` and the API guide at api.contentstudio.io/guide.

### For developers and automation users

**1. Schedule a whole content calendar from a script**
A marketer keeps their calendar in a spreadsheet or CMS. A short script loops through the rows and schedules every post at the right time, across every platform, in one run. No manual copy-paste per platform.
```bash
contentstudio posts:create -c "Weekly tip: batch your content 📅" \
  --account <fb_id> --publish-type scheduled --schedule "2026-05-01 09:00:00"
```

**2. Cross-post one message to many channels at once**
Announce something everywhere in a single command instead of re-entering it per network.
```bash
contentstudio posts:create -c "Big launch today 🚀" \
  -i <fb_id> -i <li_id> -i <tw_id> --publish-type scheduled --schedule "2026-05-01 09:00:00"
```

**3. Auto-announce from your CMS or CI pipeline**
When a new blog post or release ships, the pipeline automatically drafts or schedules the matching social posts. Publishing becomes part of the release, not an afterthought.

**4. Bulk-upload media to the library**
Seed the ContentStudio media library with a folder of images or videos in one pass, ready to attach to posts later.
```bash
contentstudio --json media:upload --file ./hero.jpg
```

**5. Audit and report with JSON**
Pull accounts and posts as JSON and pipe into dashboards, spreadsheets, or a weekly report. Great for ops and agencies who need visibility across many accounts.
```bash
contentstudio --json accounts:list
contentstudio --json posts:list
```

### For agencies and operations teams

**6. Manage many client workspaces from one place**
Switch between client workspaces, or target a specific one per command, so a single script can operate across an entire book of clients.
```bash
contentstudio --json workspaces:list
contentstudio posts:list --workspace <client_workspace_id>
```

**7. Run approvals without leaving the shell (or wire it into Slack)**
Reviewers approve or reject queued posts from the terminal, or through a small Slack bot built on top of the CLI. The whole publish-and-review lifecycle stays automatable.
```bash
contentstudio posts:approve --post <post_id> --action approve
```

**8. Add the first comment or an internal note automatically**
Automatically drop the "link in the first comment" or leave a review note for a teammate, as part of the posting script.
```bash
contentstudio --json comments:add <post_id> "Link in comments 👇"
```

### For AI-agent users (the standout story)

**9. Let an AI assistant run your social media in plain English**
Install the skill once, then just ask. The agent figures out the right commands, previews the action, and executes it.
```bash
npx skills add contentstudioio/contentstudio-agent
```
> Example ask to the agent: "Draft a LinkedIn post about our new pricing and schedule it for 9am tomorrow." The agent confirms the workspace, previews with `--dry-run`, then schedules it.

**10. Publish by chatting inside Claude Desktop, Cursor, or Copilot (MCP)**
Non-terminal users connect the `contentstudio-mcp` server to their AI app and manage ContentStudio through conversation. A downloadable Claude Desktop bundle is available at contentstudio.io/mcp-server.

**11. Always-on, headless automation agents**
The tool is designed for unattended runtimes such as OpenClaw, cron daemons, and Docker containers. Set the API key as an environment variable and an agent can manage posting around the clock.

---

## Messaging angles marketing can use

- **AI-agent-ready.** Works as an installable skill and as an MCP server, compatible with Claude, Cursor, Copilot, and headless agent runtimes. This is a fresh, on-trend narrative.
- **Automation-safe by design.** JSON output, dry-run previews, non-zero exit codes on failure, and a human-in-the-loop approval model. Serious teams can trust it in production.
- **Developer-led growth.** A 2-minute path from `npm install` to a first successful post opens ContentStudio to a technical audience the browser UI never reached.
- **Category differentiation.** Public CLI plus agent skill plus MCP is rare among social schedulers. Good fuel for comparison content.
- **API plan pull.** Every use case is a reason to be on a plan that includes API access.

---

## Content and campaign ideas

- Launch blog post: "Automate ContentStudio from your terminal, or hand it to an AI agent."
- Tutorial: "Schedule a month of posts with a 10-line script."
- Demo video or GIF: type one `posts:create` command, show the post going live.
- AI demo: "Use Claude to run your social media" (skill or MCP walkthrough).
- Developer distribution: npm, GitHub, Product Hunt, dev.to, Hacker News, relevant subreddits and communities.
- Comparison page: "ContentStudio API plus CLI plus MCP vs schedulers with no developer surface."
- Docs-as-marketing: the quickstart at api.contentstudio.io/guide as a shareable, linkable asset.

---

## Proof points and links

| Asset | Where |
|---|---|
| CLI package | npm: `contentstudio-cli` (`npm i -g contentstudio-cli`) |
| Agent skill | `npx skills add contentstudioio/contentstudio-agent` |
| MCP server | npm: `contentstudio-mcp`, bundle at contentstudio.io/mcp-server |
| API guide and reference | api.contentstudio.io/guide |
| Source repo | github.com/contentstudioio/contentstudio-agent |

---

## Accuracy guardrails (please do not over-claim)

- **Publishing-focused today.** v1 covers posting, media, approvals, comments, and account/workspace discovery. It does **not** include analytics, inbox, or CRM commands yet. Do not message "analytics from the CLI" or "manage your inbox from the terminal."
- **Human-in-the-loop, not fully autonomous.** The agent skill deliberately confirms the target workspace before any change and previews mutations first. Position it as a safe assistant, not a hands-off autopilot.
- **Requires an API key.** Users generate a key in ContentStudio under Settings, API Keys, and need a plan with API access. Keep this clear so trials do not create confusion.
- **11 supported platforms** as listed above. Do not imply networks we do not support.
