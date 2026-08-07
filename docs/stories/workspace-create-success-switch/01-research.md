# Research — Workspace-created success state with "switch to it now" prompt

## Goal

After a user creates a new workspace from the in-app "Create a new workspace" modal,
replace the current toast-and-close behavior with an in-modal success state that lets
them jump straight to the new workspace or stay where they are.

## Current State

- Modal: `create-workspace-modal` — `src/modules/account/components/CreateWorkspaceModel.vue`
  renders the fields form `src/modules/setting/components/workspace/WorkspaceFields.vue`
  (name, timezone, super admin, logo). Title key `workspace.create_modal.title`.
- Create button lives in `WorkspaceFields.vue` (label `workspace.fields.save_button`),
  calls `saveWorkspace(true, true)`.
- Create logic: `saveWorkspace()` in `src/composables/useWorkspaceCore.ts`. On success
  (create path) it currently:
  - fires the `workspace_created` Usermaven event,
  - shows a success toast `alertMessage(t('common.workspace_mixin.workspace_saved'), 'success')`,
  - hides the modal (`$cstuModal.hide('create-workspace-modal')`), refetches the
    workspace list, and emits `fetch-latest-workspaces`.
  - It does **not** switch to the new workspace (except in the onboarding route, where it
    routes to `home` in the new workspace).
- **Loading state already exists:** the create button is already bound
  `:loading="getSettingLoaders.saveWorkspace"` and `:disabled="…saveWorkspace"`, so it
  already shows a spinner and is disabled while the request is in flight.
- **Switching already exists:** `changeWorkspace(...)` (from `useWorkspaceSwitcher`, via
  `useWorkspaceCore`) is the function that moves the user into another workspace; the new
  workspace's slug comes back as `response.slug`.
- Existing failure handling in `saveWorkspace` (keep as-is): plan limit reached → opens
  `workspace-limits-dialog`; trial expired / upgrade / generic error / read-only (403) →
  error toast or the relevant dialog. Network/unknown errors fall to the catch block.

## What Needs to Change

- On **successful creation**, instead of just toasting and closing, transition the modal
  content to a **success state**: green check icon, heading, subtext, a primary
  "Go to [new workspace]" button, and a secondary "Stay in [current workspace]" button.
- **Primary button** → switch into the new workspace (reuse `changeWorkspace` with the
  new `response.slug`).
- **Secondary button** → close the modal, stay in the current workspace, and reset the
  modal back to the empty form for next time.
- **Remove the success toast** (`workspace_saved`) on create — the modal now confirms it.
  Keep all error toasts/dialogs unchanged. Keep the `workspace_created` Usermaven event.
- Keep the existing button loading state during creation.

## UX Reference

Common "resource created → switch now / stay" pattern (e.g. Slack/Notion after creating
a new workspace/team). Nothing exotic — an in-modal success panel replacing the form.

## Decisions (from user)

1. **In-app modal only** — the success state applies to the "Create a new workspace"
   modal (workspace dropdown / settings). The onboarding "create your first workspace"
   flow is unchanged (it already auto-routes into the new workspace).
2. **Web only** — no iOS/Android stories.

Implication: the success state must **not** fire on the onboarding route
(`route.name === 'onboardingWorkspace'`), which keeps its current auto-redirect.

## Mobile Context

N/A — web only per the decision above.

## Files Involved

- `src/modules/setting/components/workspace/WorkspaceFields.vue` — modal body; add the
  success-state view + the two buttons
- `src/composables/useWorkspaceCore.ts` — `saveWorkspace()` success branch: drop the
  success toast, surface success + the new slug/name to the modal instead of auto-hiding
- `src/modules/account/components/CreateWorkspaceModel.vue` — modal shell (title/close)
- `src/composables/useWorkspaceSwitcher.ts` — `changeWorkspace` used by the primary button
- `src/locales/{en,fr,de,it,es,el,zh,pl}/*.json` — new success-state copy (workspace namespace)
