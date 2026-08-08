# Research — Finalize the Automations wizard UI

Item 16 of the 7 Aug 2026 backlog batch.

## Current state

There are three automation create flows, all step-based, all living in `contentstudio-frontend/src/modules/automation/`:

| Flow | Entry component |
|---|---|
| RSS Auto-Post | `components/rss/SaveRSSAutomation.vue` |
| Bulk Schedule (CSV) | `components/csv/BulkUploadAutomationSave.vue` |
| Evergreen / Recycle | `components/evergreen/create/EvergreenMain.vue` with `components/evergreen/create/TabsContent.vue` |

### The concrete inconsistency: two different step indicators

RSS and Bulk CSV both use the shared component:

- `SaveRSSAutomation.vue:14` imports `StepIndicator` from `src/components/common/StepIndicator.vue` and renders it at line 73 with `:steps="stepIndicatorSteps"` and a `@step-click` handler.
- `BulkUploadAutomationSave.vue:233` renders the same `StepIndicator` the same way.

Evergreen does not. `components/evergreen/create/TabsContent.vue` hand-rolls its own three-step header, and it uses **hardcoded hex colors, a different one per step**:

- Step 1 active: `#7272FF` border and text
- Step 2 active: `#4CCE88`
- Step 3 active: `#FF9472`
- Inactive border: `#EAF0F6`, inactive text: `#4A4A4A`

So the same product concept (a three-step automation wizard) renders with a shared neutral indicator in two flows and a bespoke multi-coloured one in the third. The hardcoded colors are also a white-label problem: they cannot follow a customer's primary color, which the shared `StepIndicator` handles through its `getStepColorClass` / `getStepBorderClass` / `getStepBackgroundClass` helpers.

### Other structural differences

- **Step state naming.** RSS and CSV track step state as named booleans on a store-ish object (`getRssAutomationTabStatus.first`, `getCsvAutomationTabStatus.first`). Evergreen uses an `activeTab` string with the values `'first'`, `'second'`, `'third'`. Same three steps, three different representations.
- **Step visibility mechanism.** RSS uses `v-show`, CSV uses `:class="{ hidden: ... }"`, Evergreen switches on `activeTab`. Three ways to do one thing, with different implications for whether hidden steps stay mounted.
- **Step validation.** Evergreen has `isValidStep` plus explicit `processStepFirst` / `processStepSecond` / `processStepThird` functions. RSS and CSV validate differently, with `resetAccountValidation` and `checkValidForApprove` respectively.
- **Step click behaviour.** RSS and CSV both delegate to a `handleStepClick`. Evergreen inlines the guard in the template (`@click.prevent="activeTab !== 'first' && callStep(2)"`), so which steps are clickable is expressed differently.
- **Exit behaviour.** RSS has an `exitConfirmation` call with a hardcoded label string (`' RSS Auto-Post'`, note the leading space) passed at the call site. Worth checking whether the other two flows confirm on exit at all.
- **Evergreen has more sub-steps than the other two**: `EvergreenAccountSelection`, `AddEvergreenPost`, `EvergreenBulk`, `EvergreenOptimization`, `EvergreenFinalization`, plus `URLVariationModal` and three separate removal-confirmation dialogs. Its shape is genuinely bigger, which is part of why it drifted.

### Legacy styling markers

`TabsContent.vue` also carries Tailwind important-overrides (`!border-…`, `cursor-default!`) mixed with the hardcoded hex values. The repo's standards forbid hardcoded brandable colors and prefer component props over utility overrides, so this file is out of line with current conventions independent of the consistency question.

## What needs to change

- One step indicator for all three flows. Evergreen adopts the shared `StepIndicator`, which also removes the hardcoded colors.
- One consistent step chrome: header, step labels, progress affordance, which steps are clickable when, and the next/back/save button placement and labels.
- Consistent validation feedback: where a step's error appears and what it looks like when a user tries to advance with something missing.
- Consistent exit behaviour across all three flows, with the flow name coming from translation rather than a literal at the call site.
- Consistent empty, loading and error states within steps.
- No hardcoded brandable colors anywhere in the wizard chrome.

## Explicitly not in scope

Changing what any automation does, its step order, or its fields. This is a UI consistency pass. If a step's content is wrong, that is separate work.

## Related existing work, all distinct from this

- `docs/stories/automations-empty-states/` — shipped empty states for the three automation **listings**, not the create wizards.
- `docs/stories/bulk-automation-fixes/` — functional fixes in the bulk flow.
- `docs/stories/reusable-composer-in-evergreen/` — replaces the editor inside the Evergreen modal with the composer EditorBox and MediaBox. That story changes step content while this one changes wizard chrome, so they touch the same files and should be sequenced deliberately.

## Files involved

- `contentstudio-frontend/src/components/common/StepIndicator.vue` (the shared component to adopt)
- `contentstudio-frontend/src/modules/automation/components/rss/SaveRSSAutomation.vue`
- `contentstudio-frontend/src/modules/automation/components/csv/BulkUploadAutomationSave.vue`
- `contentstudio-frontend/src/modules/automation/components/evergreen/create/TabsContent.vue`
- `contentstudio-frontend/src/modules/automation/components/evergreen/create/EvergreenMain.vue`
- `contentstudio-frontend/src/modules/automation/components/evergreen/create/{EvergreenAccountSelection,AddEvergreenPost,EvergreenBulk,EvergreenOptimization,EvergreenFinalization}.vue`
- `contentstudio-frontend/src/locales/*/automation.json`

Note: the `automation` module has no module-level `CLAUDE.md` despite large files, so there is no local convention doc to defer to. The repo-level standards apply.

## Mobile

None. Automation create flows are web only.
