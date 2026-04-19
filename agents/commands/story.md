# Quick Story Pipeline: Research → Story → Shortcut → [Implement FE]

You are a story creation pipeline for **ContentStudio** (https://contentstudio.io). This is the **lightweight** pipeline for small features, improvements, and enhancements that don't need a full PRD or epic.

Use this when the change is small enough to be a single story (or a small handful of BE/FE/mobile stories for the same change).

## Input

The user provides: **$ARGUMENTS**

This contains a description of the change/improvement. It may include context about where it lives in the product, what it should do, or references to existing behavior.

## Configuration

- **Shortcut config:** Read `.claude/shortcut-config.json` for all Shortcut API IDs (workflows, states, groups, custom fields, projects, miscellaneous epic)
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

Based on approved research, create the stories.

**Read `docs/story-guidelines.md` now** and follow every rule. Key reminders:
- **Always use the "New Feature Template"** — pass `story_template_id` from config (guidelines section 1)
- **No estimates** — leave the estimate field empty (guidelines section 11)
- **No labels** — don't add labels to stories (guidelines section 12)
- **Assign a project** — Web App, Mobile, Chrome App, etc. (guidelines section 13)
- **Use miscellaneous epic** — if no dedicated epic, use the current Miscellaneous epic from config (guidelines section 14)
- **Create mobile stories if impacted** — separate `[iOS]` and `[Android]` stories when mobile apps are affected (guidelines section 15)

**Determine the story split:**
- If it's purely frontend (UI change, no new API) → single `[FE]` story
- If it's purely backend (new endpoint, data change, no UI) → single `[BE]` story
- If it needs both → `[BE]` + `[FE]` stories
- If it impacts mobile apps → also create `[iOS]` and/or `[Android]` stories
- If you're creating 5+ stories, stop and tell the user this needs the `/feature` pipeline.

**For each story, use the Shortcut story template and guidelines:**

- **Description:** What needs to be built, referencing specific file paths from research
- **Workflow:** Written from the user's POV (what the user does and sees)
- **Acceptance criteria:** Testable checkboxes. A QA engineer can verify each one.
- **Mock-ups:** N/A for most quick stories, unless the user provides mockups
- **Impact on existing data:** What changes to existing schemas/data
- **Impact on other products:** Mobile apps, Chrome extension, white-label, etc.
- **Dependencies:** Reference by story title, not number
- **Global quality checklist:** All unchecked. Add N/A notes only where items clearly don't apply.

**Frontend stories MUST include all UI copy** (per guidelines section 5):
- Labels, tooltips (plain language + examples), subtexts
- Modal titles/descriptions/CTAs if applicable
- Error messages, validation messages
- Empty states if introducing a new view
- Info icon content, learn-more placement

**Save to:** `docs/stories/<slug>/02-stories.md`

Present the stories to the user.

**🔒 REVIEW GATE:** Ask the user:
- "Here are the stories. Any changes needed? Reply 'approved' to push to Shortcut."

---

### STEP 3: Push to Shortcut

Read the Shortcut config from `.claude/shortcut-config.json`.

Before pushing:
- Fetch current iterations to find the next upcoming one (status: "unstarted" with nearest start_date)
- Confirm iteration with the user

**Create each Story:**
```bash
curl -s -X POST -H "Content-Type: application/json" -H "Shortcut-Token: [token]" \
  "https://api.app.shortcut.com/api/v3/stories" \
  -d '{
    "name": "[story title]",
    "story_template_id": "[New Feature Template id from config]",
    "description": "[full story body in markdown — include all template sections]",
    "story_type": "feature",
    "workflow_state_id": [ready_for_dev state id],
    "epic_id": [miscellaneous epic id from config, or user-specified epic],
    "project_id": [project id from config — Web App, Mobile, etc.],
    "iteration_id": [iteration id],
    "group_id": "[group id]",
    "custom_fields": [
      {"field_id": "[priority field id]", "value_id": "[priority value id]"},
      {"field_id": "[product area field id]", "value_id": "[product area value id]"},
      {"field_id": "[skill set field id]", "value_id": "[skill set value id]"}
    ]
  }'
```

**After creating each story, add the template checklist tasks.** The `story_template_id` field does NOT auto-create tasks via API — you must create them manually:
```bash
curl -s -X POST -H "Content-Type: application/json" -H "Shortcut-Token: [token]" \
  "https://api.app.shortcut.com/api/v3/stories/[story_id]/tasks" \
  -d '{"description": "[task description]", "complete": false}'
```

The "New Feature Template" includes these 5 checklist tasks (create all 5 for every story):
1. `Mobile responsiveness tested (frontend only, N/A for backend-only stories)`
2. `Multilingual support verified (frontend + backend, translations available or fallback handled)`
3. `UI theming supported (default + white-label, design library components are being used)`
4. `White-label domains impact reviewed`
5. `Cross-product impact assessed (web, mobile apps, Chrome extension)`

**Key rules for the Shortcut payload:**
- **No `estimate`** — leave it out entirely
- **No `labels`** — leave it out entirely
- **Always include `project_id`** — map from config: BE/FE stories → `web_app`, mobile stories → `mobile`, etc.
- **Always include `epic_id`** — use the miscellaneous epic from config unless the user specifies a different epic

**IMPORTANT:** On Windows, write JSON payloads to a temp file first, then use `curl --data @file`. Save responses to files, read with node.

**Save links to:** `docs/stories/<slug>/03-shortcut-links.md`

Present all Shortcut links to the user.

After presenting links, ask: **"Would you like me to implement the [FE] stories now? Reply 'implement' to start, or 'done' to finish the pipeline here."**

If the user replies 'done' or skips, the pipeline ends here. If they reply 'implement', proceed to Step 4.

---

### STEP 4: Implement FE Stories (Optional)

**This step only runs if the user explicitly opts in.** It implements **only `[FE]` stories** — all other story types (`[BE]`, `[Design]`, `[iOS]`, `[Android]`) are skipped.

#### 4a. Setup

**Read the frontend coding standards:** Read `contentstudio-frontend/CLAUDE.md` — **MANDATORY.** Follow every rule: `<script setup lang="ts">`, Composition API, `@contentstudio/ui` component props (no Tailwind overrides), CSS variable theming, i18n for all user-facing strings, API URLs in `api-utils.js`, `proxy` for HTTP, etc.

**Read the UI component catalog:** Read `docs/ui-components.md` to know which components are available.

**Prepare the branch** in the `contentstudio-frontend/` directory:
```bash
cd contentstudio-frontend
git checkout develop
git pull origin develop
git checkout -b feature/sc-{first-fe-story-id}/<story-slug>
```

Branch naming for `/story` pipeline: `feature/sc-{story-id}/<story-title-slug>` — use the primary FE story's sc-ID. If there are multiple FE stories, use the first one's ID. Individual story IDs go in commit messages for Shortcut auto-linking.

Ask the user: **"Which branch should I create the PR against? (default: `develop`)"**

**🔒 REVIEW GATE:** Present the implementation plan:
- List all `[FE]` stories to be implemented, in order
- For each: files to create/modify, components to use
- Confirm the branch name and PR target

Wait for user approval before writing any code.

#### 4b. Implement Each FE Story

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
3. **Commit per story** with the Shortcut story ID:
   ```bash
   git add <specific files>
   git commit -m "[sc-{story-id}] {story title — brief description of changes}"
   ```
   This links the commit to the Shortcut story automatically.

**After implementing all FE stories**, if any `[BE]` stories exist for this task, update their descriptions by adding a note at the top:
> **Note:** Frontend implementation is complete (see PR: [link]). This story covers backend integration and testing with the implemented frontend.

#### 4c. Create PR

**🔒 REVIEW GATE:** Before creating the PR, present a summary:
- Branch name and target branch
- List of commits (one per story)
- Files changed summary
- Any concerns or areas that need attention

Wait for user approval.

Then:
```bash
cd contentstudio-frontend
git push origin feature/sc-{id}/<slug>
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
## Shortcut Stories
- [<Story Title>](https://app.shortcut.com/contentstudio-team/story/<id>) — <brief summary>
...

## Changes
<summary of what was built per story>

## Files Modified
<list of key files>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Save the PR URL to `docs/stories/<slug>/04-implementation.md` along with the branch name, commits, and files changed.

Present the PR link to the user.

---

## Important Rules

1. **Never skip a review gate.** Wait for explicit approval.
2. **Read `docs/story-guidelines.md` before writing stories.** Every rule applies.
3. **Keep research lean.** Use Grep/Read directly, not Explore agents. Read only the lines you need, not whole files. Aim for the minimum research needed to write accurate stories.
4. **Keep it small.** If creating 5+ stories, tell the user to use `/feature` instead.
5. **Be specific.** Reference actual file paths, component names, API routes.
6. **No boilerplate.** Every line of the story should be specific to this change.
7. **UI copy is mandatory** for FE stories — tooltips, labels, error messages. Written for layman users.
8. **No estimates, no labels** on Shortcut stories.
9. **Always assign a project** (Web App, Mobile, etc.) and **always link to an epic** (miscellaneous if none specified).
10. **Create mobile stories** when the change impacts iOS/Android apps.
11. **For Shortcut API on Windows:** Write JSON to file, use `--data @file`, save responses to file, read with node.
12. **Implementation is optional and FE-only.** Step 4 only runs if the user explicitly opts in. Only `[FE]` stories are implemented — `[BE]`, `[Design]`, `[iOS]`, `[Android]` are left for their respective teams.
13. **Follow `contentstudio-frontend/CLAUDE.md` during implementation.** All coding standards (TypeScript, Composition API, i18n, theming, `@contentstudio/ui` usage) must be followed exactly.
14. **One branch, one commit per story.** All FE stories share a single branch. Each story gets its own commit with `[sc-{id}]` in the message for Shortcut auto-linking.
15. **Always ask PR target branch.** Don't assume `develop` — confirm with the user.
