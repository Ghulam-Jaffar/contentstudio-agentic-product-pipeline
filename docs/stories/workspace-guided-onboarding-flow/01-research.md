# Research — Guided onboarding modal for a newly created workspace

## Goal

When a user lands on a freshly created (empty) workspace, launch a single guided modal
that walks them through 3 skippable steps — invite team → connect social accounts → add
brand URL — one shell that swaps its body per step, with Back navigation and a persistent
footer. Closing (X) dismisses the flow; it doesn't auto-reappear, and the remaining steps
stay reachable from their normal homes.

## Current State (what already exists to reuse)

- **Per-workspace onboarding state already exists.** Workspaces carry an
  `onboarding_steps` object (`src/types/generated/workspace.ts`) with statuses for exactly
  these steps: `invite_team.status`, `connect_social_account.status`,
  `setup_brand_knowledge.status` (plus `content_category`, `create_first_post`). It's
  already written by the invite flow, the save-social-accounts flow, and brand setup —
  so completion is tracked server-side today. This is the backbone for "which steps are
  done" and for the re-entry point.
- **Dashboard "Getting Started" widget** (`src/modules/common/components/widgets/GettingStarted.vue`)
  already reads `onboarding_steps` and lists these setup tasks — the natural re-entry
  point for anything skipped.
- **Step 1 — Invite team** already exists as `AddTeamMember.vue`
  (`src/modules/setting/components/workspace/team/`), with `TeamMemberBasics.vue` (email +
  membership) and `TeamMemberAccessSettings.vue` (role + granular permissions):
  - Membership: **Team** / **Client** (`settings.workspace.team.membership.*`)
  - Role: **Administrator** / **Collaborator** / **Approver** (`settings.workspace.team.roles.*`)
  - "Additional settings" disclosure (`showAdditionalSettings`) — collapsed by default;
    **Collaborator** has the largest permission set (add social accounts, view team, add
    topics, reschedule queue, approve/publish, change FB group publish method, access
    shared folder, manage approval workflows); **Approver** has a smaller set.
  - Invite is currently **one member at a time** via `useTeam().saveTeamMember`; success
    fires the `team_member_invited` Usermaven event (`useOrganizationQueries.ts`). The
    guided step needs **repeatable rows** (per-row permissions) and a batch submit that
    reports per-email results.
- **Step 2 — Connect social accounts** already exists as the connect flow
  (`src/modules/integration/components/SocialConnectHost.vue` / `Social.vue`; onboarding
  variant `account/views/onboarding/SocialPlatform.vue`; state via `useSocialAccountsModal`).
  Platform list + EasyConnect "create new link" already supported. Saving accounts already
  updates `onboarding_steps.connect_social_account`.
- **Step 3 — Brand** already exists in `publisher/ai-content-library/` (brand knowledge:
  `BrandKnowledgeSetupEditor.vue`, `useBrandKnowledgeQueries.ts`, and
  `account/composables/useOnboardingBrandGeneration.ts` — URL → crawl → Brand Profile page
  with build in progress). Saving updates `onboarding_steps.setup_brand_knowledge`.

## What Needs to Change

- **New guided modal shell** that launches automatically when a user lands on a
  newly created workspace: one modal, header ("Step X of 3" + per-step title/subtitle),
  body swaps per step, persistent footer (Back hidden on step 1; Skip + primary CTA on
  the right). Skip advances to the next step (not out). X dismisses the whole flow.
- **Step 1 body:** adapt the invite fields into **repeatable rows** with **per-row**
  membership/role/permission state (no comma-separated bulk email). "Add another member"
  appends a row inheriting the previous row's settings; rows individually removable;
  submit sends an array and reports partial failures per email.
- **Step 2 body:** reuse the connect flow but re-laid-out to fit the shell width as a
  **scrollable two-column grid**, using the **shared guided footer** (not its own), with a
  live "connected" count in the footer.
- **Step 3 body:** single website URL input + a short summary of what gets generated;
  "Build my brand" kicks off the crawl, closes the flow, lands on the Brand Profile page.
- **Persistence:** resume step index / progress on refresh mid-flow; completed actions
  already persist via `onboarding_steps`. Dismissed (X) → don't auto-reappear.
- **Analytics:** track completion vs skip per step (to see where users drop off).
- **Re-entry:** skipped steps remain reachable from their normal homes; surface a re-entry
  point via the existing dashboard Getting Started widget.

## UX Reference

Standard product onboarding checklist/stepper (Slack/Notion/Canva post-create setup) — a
single stepped modal with skip + back, backed by a persistent "getting started" checklist.

## Mobile Context

N/A — this is a web-app flow (workspace creation + these three setup flows are web).

## Decisions (from user)

1. **Additional in-app workspaces only** — fires for workspaces created via the in-app
   "Create a new workspace" flow (2nd workspace onward). The signup / first-workspace
   onboarding keeps its existing flow (no overlap).
2. **Re-entry placement: TBD** — skipped steps remain reachable from their normal homes
   regardless; the explicit re-entry point (dashboard Getting Started vs settings) is left
   open, to be pinned down with the mockups.

## Files Involved

- New guided-modal component + step bodies (likely `src/modules/account/` or
  `src/modules/onboarding/`) — reusing the existing invite/connect/brand components
- `src/modules/setting/components/workspace/team/AddTeamMember.vue` + `TeamMemberBasics.vue`
  + `TeamMemberAccessSettings.vue` — source of the invite fields/roles/permissions (adapt to rows)
- `src/modules/integration/components/SocialConnectHost.vue` / `Social.vue` — connect step
- `publisher/ai-content-library/**` + `account/composables/useOnboardingBrandGeneration.ts` — brand step
- `src/modules/common/components/widgets/GettingStarted.vue` — re-entry point (reads `onboarding_steps`)
- `onboarding_steps` on the workspace (`types/generated/workspace.ts`) — step status persistence
- `src/locales/{en,fr,de,it,es,el,zh,pl}/*.json` — new guided-flow copy
- Backend: workspace onboarding state (step index/dismissed) + batch-invite result handling
