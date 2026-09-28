# Frill changelog template

The house structure for ContentStudio announcements, derived from the three most recently published ones (Aug 2026). **Follow this every time.** Do not invent a new shape.

Source of truth: `agents/scripts/frill-sync.sh announcements [N]` prints the latest published announcements. Re-read them if this file ever looks stale.

## Announcement metadata

| Field | How we set it |
|---|---|
| `name` | Comma-joined list of the headline items, sentence case. e.g. `Bluesky analytics, Ads analytics API, other improvements` |
| `content` | Markdown, in the body structure below |
| `category_idxs` | Pick from the existing set: **New Feature**, **Improvement**, **Fix**, **Integrations**. Use every one that genuinely applies |
| `idea_idxs` | Link the originating Frill ideas when the shipped work came from the board. Historically left empty, but we now have provenance in `ledger.json`, so use it |
| `author_idx` | **Required**, despite Frill's docs implying otherwise. Creating without it returns `422 The author idx field is required`. ContentStudio's announcements are authored by `user_8nn7ykn8` (Ghulam Jaffar); confirm with `frill-sync.sh announcements 1` |
| `published_at` | **Always `null`.** Every changelog is created as a draft and published by a human from the Frill admin UI |

## Body structure

One section per headline feature, separated by `---`, then a catch-all section at the end.

```markdown
### <Feature Name>: Now Available

<Benefit-led opening paragraph. Second person, conversational, no jargon. Says what
the customer can now do and why it matters, not what was built.>

**Here's what's included:**

- **<Short bold lead-in>.** <One or two sentences of plain explanation.>
- **<Short bold lead-in>.** <One or two sentences of plain explanation.>
- **<Short bold lead-in>.** <One or two sentences of plain explanation.>

![<screenshot alt>](<frill S3 image url>)

**For more details, check:** [<Doc title>](https://docs.contentstudio.io/...)   <- only when the help doc actually exists

---

### <Next Feature>: Now Supported in ContentStudio

...

---

### Other Improvements and Fixes

- <Plain one-line bullet, customer-visible outcome only.>
- <Plain one-line bullet, customer-visible outcome only.>
```

### Heading conventions

- `### <Feature>: Now Available` for a new capability
- `### <Feature>: Now Supported in ContentStudio` for new platform/network support
- `### API Updates: New Endpoints` for public API work
- `### Other Improvements and Fixes` always last, plain bullets, no bold lead-ins

### Lead-in line

Varies by section type, always bold:
- `**Here's what's included:**` for a feature with sub-capabilities
- `**Here's how it works:**` for a feature with a user flow
- `**Here's what's new:**` for API/endpoint batches

## Voice

- **Second person, benefit first.** "Bluesky just got a whole lot more measurable", not "We added Bluesky analytics."
- **Say what the customer can now do**, never what the team implemented.
- Light personality is on-brand: a rhetorical opener ("Why guess when you can test?"), the occasional emoji. Do not overdo it.
- **No em dashes. Anywhere.** Not in headings, not in prose. Use a colon in a heading (`Feature: Now Available`) and rewrite prose to avoid them. Older published announcements do use them; that is legacy, not a pattern to copy. This matches the same rule for story and UI copy.

## What never goes in a changelog

- **Feature-flagged work.** If it is not on for general users, it is not announced. It waits for the release where the flag comes off.
- **Technical debt and internal migrations.** A realtime-service swap is invisible to customers and reads as noise.
- **Anything with no customer-visible outcome**: schema changes, env vars, cron config, deployment notes.
- **Internal identifiers.** No Helpin task keys, no epic names, no `docs/...` paths, no team member names.
- Raw release-ticket phrasing. The release ticket is written for engineers; the changelog is written for customers. Always rewrite.
- **Links to help docs that do not exist yet.** Never invent or guess an article URL, and do not point at a bare `docs.contentstudio.io`. If the doc has not been written, omit the "For more details" line entirely. A dead link in customer-facing copy is worse than no link.

## How to create one

All Frill writing goes through `agents/scripts/frill-announce.sh`, which is deliberately separate from the read-only `frill-sync.sh`:

```
frill-announce.sh draft --name "<title>" --content-file <path> --author <user_idx> \
    [--categories idx,idx] [--ideas idx,idx] [--dry-run]
```

The script hardcodes `published_at: null` and has **no publish flag, by design**. It also refuses any body containing an em dash. Always `--dry-run` first.

Category idxs: `frill-sync.sh announcement-categories`.

## Process

1. Fetch the release ticket from Helpin.
2. Split the items: customer-facing vs feature-flagged vs internal.
3. **Verify each customer-facing item against the local docs and the actual code** before describing it. The release ticket is a summary and can be optimistic.
4. Draft in this structure.
5. **PO approves the draft.**
6. Create on Frill with `published_at: null`. The PO publishes.
