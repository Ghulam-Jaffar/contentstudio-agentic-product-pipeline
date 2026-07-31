# ContentStudio AI Chat — planning bundle

Everything from the claude.ai planning session, as landed in this repo.

## Contents

```
docs/ai-chat/
  00-bundle-context.md      ← the session's CLAUDE.md, NOT loaded as project
                              instructions (repo root CLAUDE.md governs)
  00-bundle-readme.md       ← this file
  01-AI-Chat-Master-Plan.md
  02-Tool-Catalog.md        ← the spec
  03-Use-Case-Catalog.md
  04-Agents-and-Skills.md
  05-API-Gap-Register.md
  06-As-Built-Baseline.md   ← READ FIRST. What actually exists in the codebase
  07-PRD.md                 ← the PRD. Planning deliverable for the AI tools & skills epic
  contentstudio-api-v1.json ← OpenAPI export, 118 endpoints. STALE — the live API
                              has 159. Prefer contentstudio-backend/routes/api/v1.php
```

**Docs 01–05 were written without codebase access and carry stale premises.** They were
corrected to v1.1 on 2026-07-28 against the live source; `06-As-Built-Baseline.md`
records the verification and wins any disagreement. Headline changes: 159 endpoints not
118, the Inbox API is live, and 14 platform tools already ship in the chat.

## Pulling in specific docs

None of these auto-load. Reference them explicitly so you're not burning context on all
seven every session:

```
> read @docs/ai-chat/06-As-Built-Baseline.md and @docs/ai-chat/02-Tool-Catalog.md,
  then scaffold the missing Phase 1 analytics tools

> using @docs/ai-chat/03-Use-Case-Catalog.md and docs/story-guidelines.md,
  write the Phase 1 stories

> compare @docs/ai-chat/02-Tool-Catalog.md against
  contentstudio-backend/routes/api/v1.php and list every endpoint not
  covered by a tool
```

Note the third one uses the **live routes file**, not the stale JSON export.

## Keeping it in sync

If you iterate on the plan back in claude.ai, re-download the changed doc and drop it
in the same path. Nothing else to update.

If you iterate inside Claude Code and want the reasoning back out:

```
/export docs/ai-chat/session-notes.md
```

That writes the full Claude Code transcript to disk as plain text.

## Note

There's currently no one-click "send to Claude Code" from claude.ai — it's an open
feature request. Files on disk is the working path.
