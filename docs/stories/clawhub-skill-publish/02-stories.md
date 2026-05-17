# Stories — Publish ContentStudio Skill to Clawhub

Single tracking story. `[BE]` (distribution / tooling). Epic assignment: follow-up to the existing **ContentStudio Public CLI & Agent Skills** epic (sc-115952), or filed under a quarterly miscellaneous epic if it lands after that epic closes — to be decided when the story is pulled into a sprint.

---

## Story 1 — `[BE]` Publish the ContentStudio skill to Clawhub and validate Openclaw compatibility

### Description

As an Openclaw user (and a user of any Clawhub-consuming agent runtime), I want to install the ContentStudio skill from Clawhub with that platform's native install path so that I can run the full ContentStudio command surface — workspaces, accounts, media, posts, approvals, comments — through my agent without manually cloning the standalone skill repo from GitHub.

This story is a distribution follow-up to the existing CLI & Agent Skills epic (sc-115952). The skill artifact is already produced by **[BE] Publish a standalone ContentStudio skill repo for direct agent installation** (sc-116139). What's left is publishing that artifact to the Clawhub marketplace and proving it works end-to-end inside Openclaw.

### Workflow

1. Engineering submits the ContentStudio skill to Clawhub using whatever submission path Clawhub currently supports (registry-repo PR, `clawhub publish` CLI, or marketplace web form — confirmed against Clawhub's current docs at sprint pull-in).
2. Clawhub review (if any) is completed. The listing goes live with the correct title, description, version pin, install command, and link back to the standalone GitHub skill repo.
3. An engineer installs the listing inside Openclaw using Openclaw's native install command (e.g. `openclaw skill add contentstudio`, exact syntax confirmed against Openclaw docs) on a clean machine, sets `CONTENTSTUDIO_API_KEY`, and runs the documented end-to-end smoke flow against a staging workspace.
4. The smoke flow covers every use case in the ContentStudio command surface: list workspaces, list connected accounts, upload media, create a post (saved as draft — not published), list approvals, list comments. Each call returns valid JSON via `--json`.
5. A release/refresh playbook is added to the standalone skill repo's README describing what to do when a new CLI version ships so the Clawhub listing does not bit-rot.
6. The CLI launch docs and the standalone skill repo README are updated to mention Clawhub as a recognised install path alongside `npx skills add`.

### Acceptance criteria

- [ ] The ContentStudio skill is live on Clawhub with the correct title (`ContentStudio`), a description matching the standalone skill repo's tagline, a version pin, and a working install command shown on the listing.
- [ ] The listing links back to the standalone GitHub skill repo created by sc-116139.
- [ ] Installing the skill inside **Openclaw** using Openclaw's native install path on a clean machine results in a working `contentstudio` binary and a readable `SKILL.md`/manifest equivalent for the agent.
- [ ] After installation, the following commands all succeed against a staging API key on a fresh box, each returning valid JSON when `--json` is passed:
  - `contentstudio workspaces list --json`
  - `contentstudio accounts list --workspace <id> --json`
  - `contentstudio media upload --workspace <id> --file <path> --json`
  - `contentstudio posts create --workspace <id> --account <id> --text "smoke test" --media <id> --status draft --json`
  - `contentstudio approvals list --workspace <id> --json`
  - `contentstudio comments list --workspace <id> --post <id> --json`
- [ ] If Clawhub requires a different manifest format than the bundled `SKILL.md`, a thin adapter manifest is committed to the standalone skill repo and Clawhub's listing points at it; the canonical `SKILL.md` remains unchanged.
- [ ] A `RELEASING.md` (or equivalent section in the standalone skill repo's README) documents: how to bump the version on Clawhub, who can publish, and the cadence (e.g. ship a Clawhub update whenever the CLI minor version changes).
- [ ] The standalone skill repo's README mentions Clawhub install as a recognised path, alongside the existing `npx skills add contentstudio/contentstudio-agent` path.
- [ ] CLI launch docs (deliverable from sc-116121 — *[BE] Publish public CLI docs, quickstart examples, and agent setup guides*) are updated to add Clawhub to the list of supported install paths.

### Mock-ups

N/A — distribution / tooling story.

### Impact on existing data

- No ContentStudio server-side data changes.
- No app-side data changes.

### Impact on other products

- Adds a third install path for the existing ContentStudio skill (npm CLI, GitHub-based `npx skills add`, now Clawhub). Strictly additive.
- No mobile app impact.
- No Chrome extension impact.
- No frontend impact.

### Dependencies

- Depends on: **[BE] Publish a standalone ContentStudio skill repo for direct agent installation** (sc-116139) — the artifact this story distributes.
- Depends on: **[BE] Package the CLI for AI-agent discovery with bundled SKILL.md and JSON-safe execution** (sc-116115) — provides the manifest and JSON contract Clawhub will consume.
- Benefits from: **[BE] Publish public CLI docs, quickstart examples, and agent setup guides** (sc-116121) — so the docs update in AC item 9 has a single landing place.

### Global quality & compliance

- [ ] Mobile responsiveness (N/A — distribution / tooling)
- [ ] Multilingual support (N/A — marketplace listing follows the marketplace's own locale conventions)
- [ ] UI theming support (N/A — distribution / tooling)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points**
- The standalone skill repo produced by sc-116139 (location TBD when that story lands — expected to live under the ContentStudio GitHub org).
- The bundled `SKILL.md` produced by sc-116115.

**Open questions for engineering to resolve at sprint pull-in**
- Exact Clawhub submission path: registry-repo PR vs. `clawhub publish` CLI vs. web form. Check current Clawhub docs.
- Whether Clawhub consumes the same `SKILL.md` shape or requires its own manifest format. If different, write a thin adapter manifest rather than forking the canonical `SKILL.md`.
- Exact Openclaw install command syntax for verifying the listing end-to-end.

**Existing patterns to copy**
- The Openclaw verification flow should mirror the `npx skills add contentstudio/contentstudio-agent` smoke flow used in sc-116139's verification step — same commands, same staging account, same expected JSON shapes. This keeps both install paths in lockstep.

**Gotchas**
- Bit-rot: if no `RELEASING.md` exists, the Clawhub listing will drift behind the npm package within one or two CLI releases. The AC enforcing a release playbook is the mitigation; do not skip it.
- Clawhub may auto-resolve binaries by name from PATH; confirm the skill manifest specifies `contentstudio` (matching the npm bin) and not a different alias.
