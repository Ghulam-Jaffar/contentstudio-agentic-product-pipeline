# Epic F — Frontend TypeScript migration

## Problem

The frontend has a strong TypeScript core (APIs, stores, composables, analytics, planner_v2, newer modules), but legacy JavaScript islands remain: JS-script Vue SFCs and shared JS helpers. This blocks fully typed app-shell safety and leaves the highest-traffic shared code untyped. Named priority areas: AI library, video clips, billing, inbox.

## Current snapshot (at time of writing)

| Area | State |
|---|---|
| `src/**/*.ts` | ~1,028 files |
| `src/**/*.js` (under src) | ~140 files |
| `src/**/*.vue` | ~1,333 files |
| Vue SFCs already TS setup | ~1,019 |
| Vue SFCs still JS scripts | ~312 (65 Options API, 247 `<script setup>` without `lang="ts"`) |
| `any` tokens | ~951 across ~318 files |
| tsconfig | `strict: false`, `allowJs: true`, `checkJs: false` |

Largest JS SFC clusters: `modules/composer/components` (48), `modules/common/components` (37), `modules/publisher/ai-content-library` (33), `components/common` (28), `modules/integration/components` (20), `modules/setting/components` (19), `modules/inbox-revamp/components` (16), `AI-tools`/`video-clips` (24), `billing` (10), `discovery` (10).

## Goal

Migrate the remaining legacy JS (Vue SFC scripts and shared helpers) to TypeScript in a low-risk order, then tighten the compiler. No user-facing behavior change; this is technical debt reduction that improves safety and maintainability.

## Phased stories (see `02-stories.md`)

1. `[FE]` Phase 1 — shared foundation (`common/composables`, `common/lib`, `common/constants`, shared preview components)
2. `[FE]` Phase 2 — convert `<script setup>` JS SFCs to `<script setup lang="ts">` (batched by module)
3. `[FE]` Phase 3 — composer legacy slice (EditorBox, image generation, saved captions, sidebar assistant, MainComposer, SocialModal)
4. `[FE]` Phase 4 — feature islands: publisher AI content library (AI library), video clips, billing, inbox
5. `[FE]` Phase 5 — app shell (`main.js`, `router.js`, `i18n/index.js`, `config/api-utils.js`)
6. `[FE]` Phase 6 — tighten `tsconfig` (drop `allowJs`, then `strictNullChecks`, then move toward `strict`)

## Sequencing (per the migration analysis)

Shared foundation first (de-risks everything downstream), then the cheap `<script setup>` conversions, then the composer legacy slice, then the least-migrated feature islands, then the app shell, and only then tighten the compiler once the JS surface has dropped materially.

## Rules

- No behavior change. Each migrated file behaves exactly as before.
- New/edited files are written as if `strict` were on (no implicit `any`, guard nullables, no unexplained `!`).
- Do not enable global `strict` in a feature PR; that is Phase 6's controlled step.
- Do not net-grow `SocialModal.vue` / `MainComposer.vue`; peel off cohesive slices when touched (per the composer module guide).
- The gate is `type-check:staged` / `type-check:changed` staying clean for touched files; do not add new baseline errors or silence with `@ts-ignore`.

## Out of scope

- Rewriting logic or refactoring beyond what a JS-to-TS conversion requires.
- Enabling full `strict` mode before the JS surface is materially reduced.



---

---


# Epic F stories — Frontend TypeScript migration

Each phase is a `[FE]` chore. Shared acceptance criteria for every phase:
- No user-facing behavior change; the migrated code behaves exactly as before.
- `type-check:changed` (and `type-check:staged` on commit) is clean for the touched files; no new baseline type errors are introduced.
- No new implicit `any`; nullables are guarded; no unexplained non-null `!`; no `@ts-ignore` (use `@ts-expect-error` with a reason only where unavoidable).
- Existing unit tests pass; lint is clean for touched files.

---

## [FE] Phase 1 — migrate the shared foundation to TypeScript

### Description
Migrate the shared, cross-module JavaScript to TypeScript first, since composer, planner, publisher, and integration all depend on it. Covers common composables, common lib/helpers, common constants, and shared preview components.

### Acceptance criteria
- [ ] `src/modules/common/composables` JS files are migrated to TS.
- [ ] `src/modules/common/lib` (event bus, HTTP/common utilities, legacy helpers) is migrated to TS.
- [ ] `src/modules/common/constants` is migrated to TS (and used as a source of domain types where sensible).
- [ ] Shared preview components used across modules are migrated.
- [ ] Plus the shared acceptance criteria above.

### Implementation references
- `src/modules/common/composables` (~19 JS), `src/modules/common/lib` (~14 JS), `src/modules/common/constants` (~8 JS), `src/components/common` shared components.

---

## [FE] Phase 2 — convert `<script setup>` JS SFCs to `<script setup lang="ts">`

### Description
Convert the ~247 Vue SFCs that already use `<script setup>` but without `lang="ts"`. These are the cheapest wins because the structure is already correct. Batch by module.

### Acceptance criteria
- [ ] `<script setup>` SFCs without `lang="ts"` are converted to `<script setup lang="ts">`, batched by module (for example setting, publish, integration remaining setup-style files).
- [ ] Props/emits use typed `defineProps`/`defineEmits`; two-way bindings use `defineModel()` where it is a clean swap.
- [ ] Plus the shared acceptance criteria above.

### Implementation references
- Largest setup-style JS clusters: `modules/integration/components` (~20), `modules/setting/components` (~19), remaining `publish` SFCs. Prioritize modules that are otherwise "clean/nearly migrated".

---

## [FE] Phase 3 — composer legacy slice

### Description
Migrate the legacy composer components to TypeScript, the biggest single JS SFC cluster. Includes the old EditorBox, image generation, saved captions, sidebar assistant, and the large `MainComposer.vue` and `SocialModal.vue`.

### Acceptance criteria
- [ ] Legacy composer JS SFCs are migrated to `<script setup lang="ts">` (or typed Options API only where a full conversion is out of scope for that file).
- [ ] `SocialModal.vue` and `MainComposer.vue` do not net-grow; cohesive slices are peeled into focused typed components/composables when touched.
- [ ] Plus the shared acceptance criteria above.

### Implementation references
- `src/modules/composer/components` (~48 JS SFCs) and `src/modules/composer/composables` (~8 JS). Follow the composer_v2 module guide's debt-reduction rule for the monolith files.

---

## [FE] Phase 4 — feature islands: AI library, video clips, billing, inbox

### Description
Migrate the least-migrated feature areas: publisher AI content library, AI-tools video clips, billing, and inbox.

### Acceptance criteria
- [ ] `src/modules/publisher/ai-content-library` JS SFCs and logic files are migrated to TS.
- [ ] `src/modules/AI-tools` / `video-clips` remaining JS is migrated.
- [ ] `src/modules/billing` (JS Vue + logic/constants) is migrated to TS.
- [ ] `src/modules/inbox-revamp` remaining JS is migrated.
- [ ] Plus the shared acceptance criteria above.

### Implementation references
- `publisher/ai-content-library` (~33 JS SFCs + ~5 JS logic), `AI-tools`/`video-clips` (~24), `billing` (~10 JS Vue + ~11 JS logic/constants), `inbox-revamp/components` (~16).

---

## [FE] Phase 5 — app shell

### Description
Migrate the app bootstrap and config once the dependent route/config modules are typed.

### Acceptance criteria
- [ ] `main.js`, `router.js`, `src/i18n/index.js`, and `src/config/api-utils.js` are migrated to TypeScript.
- [ ] App boots and routes correctly with no behavior change.
- [ ] Plus the shared acceptance criteria above.

### Implementation references
- Do this after Phases 1-4 so dependent modules are already typed.

---

## [FE] Phase 6 — tighten the TypeScript config

### Description
Once the JS surface has dropped materially, tighten the compiler in controlled steps.

### Acceptance criteria
- [ ] `allowJs` is removed (no remaining JS under `src` that the build depends on), or the last stragglers are migrated first.
- [ ] `strictNullChecks` is enabled and the resulting errors are resolved.
- [ ] A path toward full `strict` is documented, with remaining `any` reduction tracked.
- [ ] The full `type-check` has no new errors attributable to this change.

### Implementation references
- Current `tsconfig.json`: `strict: false`, `allowJs: true`, `checkJs: false`. Tighten incrementally (allowJs off, then strictNullChecks, then strict) as a dedicated effort, not inside a feature PR.
