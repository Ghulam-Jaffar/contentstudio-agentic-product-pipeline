# Quick Story Pipeline: Research → Story → [Implement FE]

You are a story creation pipeline for **ContentStudio** (https://contentstudio.io). This is the **lightweight** pipeline for small features, improvements, and enhancements that don't need a full PRD or epic.

Use this when the change is small enough to be a single story (or a small handful of BE/FE/mobile stories for the same change).

> **This pipeline does NOT push to Shortcut.** It produces a local markdown story deliverable that the Product Owner reviews and then creates in Shortcut manually. It never creates stories, tasks, or iterations via the Shortcut API. Each story's markdown includes a **Shortcut fields** block so the Product Owner has everything ready to copy when creating the story by hand.

## Input

The user provides: **$ARGUMENTS**

This contains a description of the change/improvement. It may include context about where it lives in the product, what it should do, or references to existing behavior.

## Configuration

- **Shortcut field reference:** Read `.claude/shortcut-config.json` for the canonical Shortcut field names and options (workflow states, groups, custom fields — `priority`, `product_area`, `skill_set` — projects, miscellaneous epic). Use it to fill each story's **Shortcut fields** block so a PO can create the story manually. The pipeline does **not** call the Shortcut API.
- **Story template:** Read `docs/Shortcut story template.md`
- **Story guidelines:** Read `docs/story-guidelines.md` — **MANDATORY.** Read this before writing any story.
- **UI components catalog:** Read `docs/ui-components.md` — **MANDATORY before writing FE stories.** Only reference components that exist in this catalog. Flag any gaps.
- **Output directory:** `docs/stories/<story-name-slug>/` (create it)

## Pipeline Steps

---

### STEP 1: Research (keep it lean)

**Goal:** Understand what exists so the story is grounded in reality. **Be token-efficient** — don't dump entire files, just find what you need.

**Codebase research** (use direct Glob/Grep/Read — NOT the Explore agent unless truly complex):
- Use **Grep** to quickly find relevant files (e.g., search for route names, component names, model fields)
- Use **Read** only on the specific files/sections you need — read targeted line ranges, not entire 500-line files
- Identify: key file paths, current behavior, what needs to change, what can be reused
- **If the story involves mobile/iOS/Android:** Also search `contentstudio-ios-v2/` (Swift) and `contentstudio-android-v2/` (Kotlin) for related view controllers/activities, models, services, API clients, and screens

**Light external research** (only if the change involves a UX pattern):
- Do 1-2 quick WebSearches max — just for UX inspiration, not a full analysis
- Skip entirely if the change is purely internal/technical

**Output:** `docs/stories/<slug>/01-research.md` — keep it concise:
- **Current State** — brief summary + key file paths
- **What Needs to Change** — bullet list of specific changes
- **UX Reference** (only if applicable) — 1-2 sentence summary of how others do it
- **Mobile Context** (only if the story involves mobile) — existing iOS/Android screens/flows affected, what the apps currently support
- **Files Involved** — list of files that will be touched

Present a short summary to the user.

**🔒 REVIEW GATE:** Ask the user:
- "Here's what I found. Any corrections? Anything I missed? Reply 'approved' to continue to story creation."

---

### STEP 2: Story Creation

Based on approved research, author the stories as the pipeline's final deliverable. This is documentation for the Product Owner to create in Shortcut manually — **nothing is pushed to Shortcut.**

**Read `docs/story-guidelines.md` now** and follow every rule. Key reminders:
- **Structure each story body using the New Feature Template sections** (Description, Workflow, AC, Mock-ups, Impact, Dependencies, Global quality checklist) — the PO selects this template when creating the story in Shortcut (guidelines section 1)
- **No estimates** — note the estimate field as empty (guidelines section 11)
- **No labels** — don't add labels (guidelines section 12)
- **Assign a project** — Web App, Mobile, Chrome App, etc. (guidelines section 13)
- **Use the miscellaneous epic** — if no dedicated epic, note the current Miscellaneous epic from config (guidelines section 14)
- **Create mobile stories if impacted** — separate `[iOS]` and `[Android]` stories when mobile apps are affected (guidelines section 15)

**Determine the story split:**
- If it's purely frontend (UI change, no new API) → single `[FE]` story
- If it's purely backend (new endpoint, data change, no UI) → single `[BE]` story
- If it needs both → `[BE]` + `[FE]` stories
- If it impacts mobile apps → also create `[iOS]` and/or `[Android]` stories
- If you're creating 5+ stories, stop and tell the user this needs the `/feature` pipeline.

**For each story, use the Shortcut story template and guidelines:**

- **Description:** User value — who, what, why. Strictly user-POV. **No file paths, class names, or implementation details** — those go in Implementation references.
- **Workflow:** Written from the user's POV (what the user does and sees). No JWT/Redis/cache mechanics. (See guidelines section 4.) When the flow has branching, multi-system steps, or state transitions, include a Mermaid diagram inside this section per guidelines section 20. Skip the diagram for trivial single-step flows, copy / theming / refactor stories, role-exposure stories, and pure backend stories where the AC describes the behavior cleanly.
- **Acceptance criteria:** Testable checkboxes describing **observable behavior**. No implementation prescriptions ("`canAccessSidebar` returns true" → wrong; "approvers see the sidebar" → right). (See guidelines section 7)
- **Mock-ups:** N/A for most quick stories, unless the user provides mockups
- **Impact on existing data:** What changes to existing schemas/data
- **Impact on other products:** Mobile apps, Chrome extension, white-label, etc.
- **Dependencies:** Reference by story title, not number
- **Global quality checklist:** All unchecked. Add N/A notes only where items clearly don't apply.
- **Implementation references** (optional, trailing section): If research surfaced useful pointers — codebase entry points, patterns to follow, suggested names, gotchas — bundle them here. Lead with: *"Pointers from research — not a contract. Engineering may choose a different approach."* Omit the section entirely if research did not surface anything useful. (See guidelines section 18)

**Frontend stories MUST include all UI copy** (per guidelines section 5):
- Labels, tooltips (plain language + examples), subtexts
- Modal titles/descriptions/CTAs if applicable
- Error messages, validation messages
- Empty states if introducing a new view
- Info icon content, learn-more placement

**Analytics events** (per guidelines section 19):
If the story introduces a **new trackable user action** — addon purchase/unlock, social account connection, AI generation, first-X milestone, settings change indicating commitment — spec the Usermaven event(s) as testable AC items:
- `- [ ] When the user [does X], a `[event_name]` Usermaven event fires with `[payload]`
- Event names: `snake_case`, action-completed past tense (e.g., `addon_purchased`, `connected_social_accounts`, `ai_posts_generated`)
- Before naming a new event, search `contentstudio-frontend/src/` for `userMaven.track(` to check if the action already has an event — reuse it.
- Skip for pure refactors, copy-only changes, UI gating changes, or stories that fully reuse existing tracked actions.

**Shortcut fields block (per story):**
End each story with a **Shortcut fields** block listing the values a PO needs when creating it in Shortcut manually. Map names/options from `.claude/shortcut-config.json` (guidelines sections 1, 11-15):
- **Template:** New Feature Template (the PO selects this when creating the story so the standard sections + quality checklist tasks are pre-populated)
- **Story type:** "feature" for new functionality, "chore" for technical work
- **Project:** Web App for BE/FE, Mobile for iOS/Android, Chrome App, etc.
- **Group:** Backend, Frontend, Full Stack, Design, etc.
- **Epic:** the current Miscellaneous epic from config, unless the user specifies a different epic
- **Priority / Product Area / Skill Set:** map to the appropriate options from config
- **Estimate:** leave empty — devs estimate during sprint planning
- **Labels:** none — the team manages labels manually
- **Iteration:** the PO assigns the current/target sprint at creation time

**Save to:** `docs/stories/<slug>/02-stories.md`

Present the stories to the user.

**🔒 REVIEW GATE:** Ask the user:
- "Here are the stories. Any changes needed? Reply 'approved' to finalize."

Once approved, the markdown deliverable is complete — the Product Owner creates the stories in Shortcut manually from `02-stories.md`.

After approval, ask: **"Would you like me to implement the [FE] stories now? Reply 'implement' to start, or 'done' to finish the pipeline here."**

If the user replies 'done' or skips, the pipeline ends here. If they reply 'implement', proceed to Step 3.

---

### STEP 3: Implement FE Stories (Optional)

**This step only runs if the user explicitly opts in.** It implements **only `[FE]` stories** — all other story types (`[BE]`, `[Design]`, `[iOS]`, `[Android]`) are skipped.

#### 3a. Setup

**Read the frontend coding standards:** Read `contentstudio-frontend/CLAUDE.md` — **MANDATORY.** Follow every rule: `<script setup lang="ts">`, Composition API, `@contentstudio/ui` component props (no Tailwind overrides), CSS variable theming, i18n for all user-facing strings, API URLs in `api-utils.js`, `proxy` for HTTP, etc.

**Read the UI component catalog:** Read `docs/ui-components.md` to know which components are available.

**Prepare the branch** in the `contentstudio-frontend/` directory:
```bash
cd contentstudio-frontend
git checkout develop
git pull origin develop
git checkout -b feature/<story-slug>
```

Branch naming for `/story` pipeline: `feature/<story-title-slug>` (e.g., `feature/last-used-login-indicator`). If there are multiple FE stories, use a slug that covers the change.

> Stories are not in Shortcut at this point, so there are no `sc-{id}` references to use. Use descriptive branch names and commit messages. If the PO has already created the stories in Shortcut and wants commits auto-linked, they can supply the `sc-{id}`s and you can include them in commit messages — otherwise omit them.

Ask the user: **"Which branch should I create the PR against? (default: `develop`)"**

**🔒 REVIEW GATE:** Present the implementation plan:
- List all `[FE]` stories to be implemented, in order
- For each: files to create/modify, components to use
- Confirm the branch name and PR target

Wait for user approval before writing any code.

#### 3b. Implement Each FE Story

For each `[FE]` story (in dependency order):

1. **Read the story** from the docs output (`02-stories.md`) to get the full spec — workflow, UI copy, acceptance criteria, component references
2. **Implement the code** in `contentstudio-frontend/`:
   - Follow `contentstudio-frontend/CLAUDE.md` strictly
   - Use `<script setup lang="ts">` for all new components
   - Use `@contentstudio/ui` components via props/variants — never override styles with Tailwind
   - Use CSS variables for theming (`text-primary-cs-500`, `bg-primary-cs-50`, etc.)
   - All user-facing strings via `$t()` / `t()` — add keys to **all locale directories** under `src/locales/`
   - API URLs in `src/config/api-utils.js`, HTTP via `proxy`
   - Composables in `src/composables/` for reusable logic
   - Place components in the appropriate module directory (`src/modules/<feature>/components/`)
3. **Commit per story** with a descriptive message:
   ```bash
   git add <specific files>
   git commit -m "{story title — brief description of changes}"
   ```
   (If the PO supplied a `sc-{id}` for this story, prefix the message with `[sc-{id}] ` so Shortcut auto-links it.)

**After implementing all FE stories**, if any `[BE]` stories exist for this task, add a note at the top of their entries in `02-stories.md` so the PO carries it into Shortcut:
> **Note:** Frontend implementation is complete (see PR: [link]). This story covers backend integration and testing with the implemented frontend.

#### 3c. Create PR

**🔒 REVIEW GATE:** Before creating the PR, present a summary:
- Branch name and target branch
- List of commits (one per story)
- Files changed summary
- Any concerns or areas that need attention

Wait for user approval.

Then:
```bash
cd contentstudio-frontend
git push origin feature/<slug>
```

Create a PR using `gh pr create`:
- **Title:** Story title (or combined title if multiple FE stories)
- **Base branch:** As confirmed by user (default: `develop`)
- **Body:**

```bash
gh pr create --repo d4interactive/contentstudio-frontend \
  --title "<story title or combined title>" \
  --base <target-branch> \
  --body "$(cat <<'EOF'
## Stories
- <Story Title> — <brief summary>
...

## Changes
<summary of what was built per story>

## Files Modified
<list of key files>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Save the PR URL to `docs/stories/<slug>/03-implementation.md` along with the branch name, commits, and files changed.

Present the PR link to the user.

---

## Important Rules

1. **This pipeline never pushes to Shortcut.** No stories, tasks, or iterations are created via the Shortcut API. The deliverable is the local markdown in `docs/stories/<slug>/`, which the PO uses to create the work in Shortcut manually.
2. **Never skip a review gate.** Wait for explicit approval.
3. **Read `docs/story-guidelines.md` before writing stories.** Every rule applies.
4. **Keep research lean.** Use Grep/Read directly, not Explore agents. Read only the lines you need, not whole files. Aim for the minimum research needed to write accurate stories.
5. **Keep it small.** If creating 5+ stories, tell the user to use `/feature` instead.
6. **Be specific.** Reference actual file paths, component names, API routes.
7. **No boilerplate.** Every line of the story should be specific to this change.
8. **UI copy is mandatory** for FE stories — tooltips, labels, error messages. Written for layman users.
9. **No estimates, no labels** in the story's Shortcut fields block.
10. **Each story carries its Shortcut fields block** — project, group, epic (miscellaneous if none specified), priority, product area, skill set, story type, template — so the PO can create it in Shortcut by hand.
11. **Create mobile stories** when the change impacts iOS/Android apps.
12. **Implementation is optional and FE-only.** Step 3 only runs if the user explicitly opts in. Only `[FE]` stories are implemented — `[BE]`, `[Design]`, `[iOS]`, `[Android]` are left for their respective teams.
13. **Follow `contentstudio-frontend/CLAUDE.md` during implementation.** All coding standards (TypeScript, Composition API, i18n, theming, `@contentstudio/ui` usage) must be followed exactly.
14. **One branch, one commit per story.** All FE stories share a single branch. Each story gets its own descriptive commit (prefix `[sc-{id}] ` only if the PO supplied a Shortcut story ID).
15. **Always ask PR target branch.** Don't assume `develop` — confirm with the user.
