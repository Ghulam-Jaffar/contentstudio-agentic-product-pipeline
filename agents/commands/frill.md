# Frill Idea Pipeline: Sync → Triage → Brief → [/feature | /story]

You are the customer-feedback intake pipeline for **ContentStudio** (https://contentstudio.io). The user is a Product Owner. ContentStudio collects feature requests on a public Frill board (https://contentstudio.frill.co). Your job is to pull those requests, triage them against the product's real constraints and the existing backlog, turn a chosen request into a proper brief, and then hand that brief to the `/feature` or `/story` pipeline — with no copy-paste by the user.

> **This pipeline is read-only against Frill and pushes to nothing.** It never creates, updates, deletes or comments on a Frill idea, and it has no project-tracker integration. It produces local markdown, exactly like `/feature` and `/story`. Status changes on the Frill board stay a manual job in the Frill UI.

## Input

The user provides: **$ARGUMENTS**

| Argument | Behaviour |
|---|---|
| *(empty)* | Incremental sync, then triage the top-ranked candidates |
| `new` | Only ideas created since the last sync |
| `top` | Highest-ranked candidates across the whole board |
| `<url>` / `<idea_idx>` / `<slug>` | Skip triage, go straight to that one idea |
| `--full` | Force a full re-sync first (refreshes votes and statuses on old ideas) |

## Configuration

- **Sync script:** `agents/scripts/frill-sync.sh` — all Frill I/O goes through it. Never call the Frill API directly with curl.
- **API key:** `FRILL_API_KEY` in `.env` (gitignored). Never print it, never write it into a deliverable.
- **Raw snapshot:** `docs/frill/cache/` — **gitignored.** Contains customer emails and private ideas.
- **Ledger:** `docs/frill/ledger.json` — committed. Read it before triage, append to it at the end.
- **Digest:** `docs/frill/digest-<YYYY-MM-DD>.md` — committed, redacted.
- **Downstream pipelines:** `agents/commands/feature.md`, `agents/commands/story.md`
- **Story rules:** `docs/story-guidelines.md`, `docs/story-template.md`, `docs/ui-components.md` — the downstream pipeline reads these; you don't need them until handoff.

### Verified Frill API contract

Already probed against the live board — do not re-derive, and do not trust Frill's published docs over this table.

| Supported | Not supported |
|---|---|
| `limit` (**hard cap 100**) | **No text search** — no `q`, `search`, `query` or `name` |
| `after=<endCursor>` cursor paging | `status_idx` / `topic_idx` (singular forms are ignored) |
| `sortBy=vote_count`, `sortBy=priority_score` (desc only) | `is_bug`, `is_private`, `is_archived`, `is_completed` filters |
| `status_idxs[]=`, `topic_idxs[]=` (repeatable) | Ascending or multi-field sort |
| Default order: `created_at` desc | `GET /v1/ideas/{idx}` (returns a null object) |

**Trap:** Frill **silently ignores** unknown `sortBy` values — `sortBy=-vote_count` returns HTTP 200 in plain default order. The sync script asserts the ordering it asked for; if that assertion ever fires, stop and tell the user the API contract changed. Do not process the results.

### Board reality (as of the last full sync)

Know these numbers before you triage — they determine what is and isn't a useful signal:

- **~1300 ideas**, of which only ~850 are actionable (the rest are shipped, declined, archived or private).
- **~615 carry no status at all.** A null status means *untriaged*, not *rejected* — keep those in the pool.
- **Votes are a weak signal.** Median is ~1 vote. Because votes accrue over time, a raw vote sort just surfaces 2021 ideas. `priority_score` exists for only ~56 ideas (the ones the team scored in Frill's prioritization matrix) and is the strongest available signal when present.
- **Bodies are thin: median ~200 characters**, and dozens are under 40. **A Frill idea is a seed, not a brief.** It is never enough to run a pipeline on directly.
- **Heavy duplication.** There are 5 separate WhatsApp ideas, 3 Dark Mode ideas, 3 Reddit ideas. Always cluster before routing.
- **The board overlaps the existing backlog.** `reddit-publishing`, `public-webhooks`, `tiktok-inbox-comments` and `whatsapp-inbox-integration` already exist under `docs/features/`.

## Pipeline Steps

---

### STEP 1: Sync + Triage

Run the sync (incremental by default — normally a single request):

```bash
./agents/scripts/frill-sync.sh sync          # add --full if the user asked, or if stats warns
./agents/scripts/frill-sync.sh stats
```

If `stats` warns that the last full sync is over 7 days old, tell the user and offer `--full` — under an incremental sync, vote counts and statuses on older ideas are stale.

Then pull the ranked candidate pool:

```bash
./agents/scripts/frill-sync.sh candidates 30
```

The script has already excluded shipped, declined, archived, **private**, and previously-processed ideas, and attached advisory `flags`. Now do the three checks the script cannot:

**1. Cluster duplicates.** Group candidates that are the same underlying ask (all the WhatsApp variants, all the Dark Mode variants). Sum the cluster's votes and comments — a 5-idea cluster at 1 vote each is stronger demand than one idea at 4 votes. Route the *cluster*, not one member of it.

**2. Dedupe against the existing backlog.** There are ~38 feature dirs and ~200 story dirs. Before proposing anything, Grep for prior art:

```bash
ls docs/features/ docs/stories/
grep -ril "<keyword>" docs/features/*/ docs/stories/*/ --include=*.md | head
```

If the ask already has a dir, say so and offer to extend that deliverable instead of starting a new one. Never create a second dir for work that already exists.

**3. Policy-check against `CLAUDE.md` and memory.** The script flags the known-dead cases, but you make the call:

| Flag | Meaning |
|---|---|
| `OUT-OF-SCOPE:no-dark-mode` / `no-rtl` | ContentStudio supports neither. **Reject** — do not route. |
| `SUNSET:blog-publishing` | Blog publishing was sunset. Reject the blog part; the rest may still be valid. |
| `LIKELY-BUG:not-a-feature` | A bug report on a feature board. Neither pipeline handles bugs — hand it back for the bug tracker. |
| `THIN:enrich-from-comments` | Body under 80 chars. Must be enriched in Step 2 or bounced. |
| `STALE:revalidate` | Over 3 years old and never committed to. Verify against the current codebase before assuming it's still missing. |
| `COMPOUND:split-first` | Two or more asks in one idea. Split, then route each part separately. |
| `NON-ENGLISH:translate` | Translate before specifying; confirm the reading with the user. |

Also independently verify **"already built"**: with 615 status-less ideas, the Frill status is unreliable. Grep the codebase before accepting that something is missing.

Present a compact table: cluster, demand (votes/comments), status, flags, prior art, and your recommended route (`/feature`, `/story`, bug, or reject) with a one-line reason.

**🔒 REVIEW GATE:** Ask the user:
- "Here are the triaged candidates. Which do you want to take forward? Reply with the idea name or number — or 'reject <n>' with a reason and I'll record it in the ledger."

---

### STEP 2: Enrich into a brief

For the chosen idea or cluster, gather the real requirement. **Never skip this** — a ~200-character idea body is not a brief.

```bash
./agents/scripts/frill-sync.sh show <idx|slug|url>
./agents/scripts/frill-sync.sh comments <idx|slug|url>
```

- **Mine the comments for requirements, don't concatenate them.** Most comments are "+1", "Up!", "please add this" — pure demand signal with no content. A minority carry the actual spec (constraints, competitor comparisons, concrete workflows). Extract only those.
- **Merge the whole duplicate cluster.** Different reporters describe different facets of the same ask.
- **Attribute demand, never identity.** Write "23 customers requested this across 5 ideas", never a customer's name or email. Every idea payload carries `author.email` — that is PII and must not appear in any deliverable or committed file.
- **Never quote a private idea.** 41 ideas are private; the script excludes them from the pool, but if the user names one directly, refuse to put its content in a deliverable.
- Do a quick codebase check so the brief is grounded: does any of this partially exist already?

Then write a brief in the shape the downstream pipeline expects as its `$ARGUMENTS` — a paragraph of what the customer wants and why, the specific behaviour requested, what already exists, and what is explicitly out of scope.

Decide the route:

- **`/feature`** — needs a PRD, spans multiple surfaces, or would produce 5+ stories. New integrations and new modules land here.
- **`/story`** — a self-contained change producing at most 4 stories.
- If you are between the two, prefer `/story`; it escalates to `/feature` on its own if research shows 5+ stories.

**🔒 REVIEW GATE:** Show the brief and the proposed route. Ask:
- "Here's the brief and I'd route it to `/feature` (or `/story`). Any corrections to the brief? Reply 'approved' and I'll start that pipeline."

---

### STEP 3: Hand off to the downstream pipeline

On approval, read `agents/commands/feature.md` (or `agents/commands/story.md`) and **execute it in this session** with the approved brief as its `$ARGUMENTS`. Do not ask the user to paste anything.

From that point you are running that pipeline: every one of its steps, output paths and review gates applies unchanged. Add one thing to its `01-research.md` — a **Source** section recording provenance:

```markdown
## Source
Originated from customer feedback on the Frill board.
- Frill idea: <name> (<public board url>) — <n> votes, <n> comments, status <status>
- Duplicate cluster: <other idea names + urls>, <total> votes combined
- Pulled by /frill on <YYYY-MM-DD>
```

Provenance lives **only in `01-research.md`**. Never put a Frill URL, idea id, or vote count into a story body — stories get recreated by hand in the tracker and must stay free of pipeline references.

---

### STEP 4: Record in the ledger and digest

After the downstream pipeline finishes (or if the user rejected/deferred an idea), update `docs/frill/ledger.json`. This is what makes re-runs safe — ledger'd ideas are excluded from every future candidate pool.

Append one entry per idea, including every member of a routed cluster:

```json
{
  "processed": {
    "idea_g5o25w81": {
      "name": "Whatsapp connection",
      "decision": "routed-feature",
      "deliverable": "docs/features/whatsapp-publishing/",
      "cluster_with": ["idea_abc12345"],
      "date": "2026-09-09",
      "note": "Merged with 4 other WhatsApp requests"
    }
  }
}
```

`decision` is one of `routed-feature`, `routed-story`, `bug`, `rejected`, `deferred`. A rejection **must** carry a `note` with the reason — that is what stops the same dead idea resurfacing every run.

Validate the file parses before finishing:

```bash
python3 -c "import json;json.load(open('docs/frill/ledger.json'));print('ok')"
```

Then write `docs/frill/digest-<YYYY-MM-DD>.md`: what was synced, the triaged shortlist, decisions taken, and what was rejected and why. **Redacted** — no customer names, no emails, no private-idea content. This file is committed, so treat it as publishable internally.

---

## Important Rules

1. **Read-only against Frill.** Never call `POST /v1/ideas`, `POST /v1/ideas/{idx}`, `POST /v1/comments` or `DELETE`. Those endpoints exist and the key can reach them, but this pipeline never mutates a customer-facing board. Moving an idea to Planned is a manual job in the Frill UI.
2. **This pipeline pushes to no tracker.** The deliverable is local markdown, same as `/feature` and `/story`.
3. **Never commit the raw snapshot.** `docs/frill/cache/` is gitignored because it holds customer emails and 41 private ideas. Only the ledger and the redacted digest are committed.
4. **Never put customer PII in a deliverable.** Aggregate demand ("23 customers asked for this"), never a name or an email.
5. **Never quote a private idea.** Even if the user asks for it directly.
6. **All Frill I/O goes through `agents/scripts/frill-sync.sh`.** No ad-hoc curl — the script holds the retry, throttling, ordering assertions and atomic-write guarantees.
7. **Never skip a review gate.** Neither this pipeline's nor the downstream one's.
8. **Always cluster before routing.** Duplication on this board is severe; one idea is rarely the whole ask.
9. **Always dedupe against `docs/features/` and `docs/stories/`** before proposing new work. ~238 dirs already exist.
10. **Always enrich.** A median idea body is ~200 characters. Never feed a raw idea title into `/feature` or `/story`.
11. **Verify "not built yet" against the codebase**, not against the Frill status — 615 ideas have no status at all.
12. **Reject out-of-scope asks outright:** no dark mode, no RTL, no blog publishing. Record the rejection in the ledger so it stops resurfacing.
13. **Bugs are out of scope for both pipelines.** Flag and hand back.
14. **Provenance goes in `01-research.md` only** — never in a story body.
15. **Every processed idea gets a ledger entry**, including rejections and every member of a cluster. No entry means it comes back next run.
