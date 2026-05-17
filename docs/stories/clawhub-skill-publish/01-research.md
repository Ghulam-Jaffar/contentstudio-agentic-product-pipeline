# Research — Publish ContentStudio Skill to Clawhub

Lean research — this is a tracking story for distribution work; the underlying skill artifact already exists.

## Current state

- ContentStudio's public CLI and agent skill are tracked under the existing feature [ContentStudio Public CLI & Agent Skills](https://app.shortcut.com/contentstudio-team/epic/115952). Local docs: [docs/features/contentstudio-public-cli-agent-skills/](../../features/contentstudio-public-cli-agent-skills/).
- The standalone skill repo is covered by epic Story 9 — *[BE] Publish a standalone ContentStudio skill repo for direct agent installation* (sc-116139). That story delivers a public GitHub repo installable via `npx skills add contentstudio/contentstudio-agent`, with a `SKILL.md` declaring the `contentstudio` binary and the `CONTENTSTUDIO_API_KEY` env var.
- The backend work to make the skill exist is complete. The remaining task is **distribution**: publishing that skill into the Clawhub marketplace so users of *Openclaw* (and other Clawhub-consuming agent runtimes) can install ContentStudio with whatever the Clawhub-native install command is.

## What needs to change (scope of this story)

- Submit the existing ContentStudio skill artifact to Clawhub via whatever Clawhub uses for marketplace submission (PR to a registry repo, CLI publish command, or web submission form — TBD by the dev based on Clawhub's current docs).
- Verify the published listing renders correctly on Clawhub (metadata, description, install instructions, version pin).
- Verify the skill works end-to-end from inside **Openclaw** for the full ContentStudio command surface: workspaces, accounts, media upload, post creation, approvals, comments — the same surface the standalone skill repo already documents.
- Document a release/refresh process so future CLI versions are reflected on Clawhub without bit-rot.

## Why this is its own story (not part of epic Story 9)

- Epic Story 9 (sc-116139) is bounded to the GitHub-based `npx skills add` install path. Clawhub is a separate marketplace with its own submission, review, and update workflow.
- This story is **post-launch tracking** — it depends on Story 9 being live and on the skill having stabilised through real `npx skills add` usage.
- It also adds an *Openclaw end-to-end validation* requirement that doesn't exist in Story 9.

## Open questions (flagged for engineering, not blocking story creation)

- What is Clawhub's exact submission path? (PR to a registry repo vs. a `clawhub publish` CLI vs. a web form). Engineering to confirm against current Clawhub docs.
- Does Clawhub require a separate manifest format than `SKILL.md`, or does it consume the same manifest? If different, a thin adapter manifest may be needed.
- What is the exact list of Openclaw use cases to validate against? The story currently lists the same command surface as Story 9 (workspaces, accounts, media, posts, approvals, comments) — confirm with product before final sprint planning.

## Files / artifacts involved

- The standalone skill repo created by epic Story 9 (location TBD when that story is in flight — likely `github.com/contentstudio/contentstudio-agent` or similar).
- No `contentstudio-backend/`, `contentstudio-frontend/`, or `social-inbox-manager/` code changes expected — this is a distribution + verification story.
- Docs updates inside the standalone skill repo's README to reference the new Clawhub install path.
- Updates to the public CLI launch docs ([docs/features/contentstudio-public-cli-agent-skills/04-epic-and-stories.md](../../features/contentstudio-public-cli-agent-skills/04-epic-and-stories.md) Story 6 deliverables) to add Clawhub as a recognised install path.

## Mobile / FE impact

- None. Pure distribution / tooling.
