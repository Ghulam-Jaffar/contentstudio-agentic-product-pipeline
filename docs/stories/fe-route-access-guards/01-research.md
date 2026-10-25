# FE route-level access guards (RBAC + white-label) · Research

**Ask (2026-10-22):** refactor frontend RBAC so every role check and white-label check is enforced by the route guard from one source, instead of manual checks inside each page. Pages get revamped and the manual check gets dropped, so a hidden page stays reachable by URL or the back button.

**Scope:** Web frontend only (`contentstudio-frontend/`). No backend, no mobile, no visual change.

**Read from:** `contentstudio-frontend` `origin/develop` @ `76daaeff59` (2026-10-22). This is a static code reading. Nothing was tested in a running app.

---

## Verdict: the problem is real

Access rules live in four kinds of places today, and they don't agree:

1. **The global guard** (`src/router.ts` `router.beforeEach`, step 3.5) only knows about one role: approvers. It calls `hasRoutePermission()`, which has a hardcoded allowlist of ~15 route names for approvers and returns `true` for every other role.
2. **Per-route `beforeEnter`**, added one route at a time, mostly for white-label: `api` (router.ts), `white-label` + `sso-domains` (`resellerSettingsWhiteLabelCheck` in `setting.ts`), `analytics_usermaven` (`userMavenWhiteLabelCheck` in `competitor.ts`), `listening`.
3. **`onMounted` / setup-time `router.push` inside page components**: billing, team, white-label, SSO, approval workflows, workflow builder, inbox.
4. **Navigation visibility** (what the menus hide). This is the most complete rulebook, and it is written three times: `SettingSidebar.vue` expanded menu, `SettingSidebar.vue` collapsed rail (its own comment says *"visibility guards are duplicated from the expanded template"*), and `HomeSettingsDropdown.vue`. Primary nav is the desktop navigation rail (`DesktopNavigationRail.vue`, mounted inside the legacy-named `TopHeaderBar.vue`, desktop only) plus the mobile drawer. Both render the item list from `useHeaderNavigation.ts`, which is where the lock and visibility rules live.

The menus hide more pages than the router blocks. Any rule that exists only in a menu is a URL-reachable gap.

### The history matches the complaint

The same fix keeps landing one route at a time:

| Commit | Date | Fix |
|---|---|---|
| `6376c8c882` | 2026-10-21 | CONT-4437 block the API page on white-label domains |
| `f7562ad422` | 2026-10-15 | block the Usermaven route on white-label domains |
| `3ad81099e2` | 2026-09-29 | lock paused accounts in the global guard |
| `6cda3b1ba5` | 2026-09-25 | correct approver access |
| `460d00aea4` | 2026-09-18 | block back-button re-entry into a locked workspace |

Code comments say the same thing: *"hiding them from the sidebar does not stop a typed URL"* (`setting.ts`, `competitor.ts`, router.ts `api`).

---

## Current state: route-by-route access matrix

Rules taken from what the nav hides. "Enforced at" is where a direct URL is actually blocked.

### Settings (`src/modules/setting/config/routes/setting.ts`)

| Route | Rule the nav applies | Enforced at | Gap |
|---|---|---|---|
| `plan`, `subscription` | `can_see_subscription` | `EnrolledPlanView.vue:435` onMounted, goes to `profile` | Page-level only |
| `team` | `can_view_team` | `Team.vue:158`, setup-time push, goes to `profile` | Page-level only. Runs during setup, before mount |
| `white-label` | `can_see_subscription` + not WL domain + not API plan + add-on | WL domain: route. API plan: global guard. Role + add-on: `WhiteLabelMain.vue:167` onMounted, goes to `home` | One page, three places |
| `sso-domains` | same as `white-label` | same split, `DomainVerification.vue:614` | One page, three places |
| `approval-workflows` | `manage_workflow` + add-on | `ApprovalWorkflowsList.vue:61` onMounted, goes to `profile` | Page-level. **Re-implements `manage_workflow` inline** (`canManageWorkflows`, line 47) instead of using `hasPermission` |
| `approval-workflows-builder` | `manage_workflow` + add-on | `ApprovalWorkflowBuilder.vue:327` onMounted | Page-level. Checks the add-on **without** `waitForPlan()`, unlike the list page, so a cold load can misjudge it |
| `emailNotificationStatus` | hidden when `shouldShowWhiteLabelData` | **Nowhere** (approver allowlist even includes it) | **Open by URL on white-label** |
| `refer&earn` | hidden on WL, and for non-`team` membership (clients) | **Nowhere** | **Open by URL on white-label and for clients** |
| `affiliate` | no menu entry at all | **Nowhere** | Orphan route, open by URL to everyone except approvers, including on WL. Confirm whether it is still needed |
| `brand-settings`, `integrations` | not API plan | Global guard (API-plan blocklist) | OK |
| `basicSetting`, `social`, `content_categories`, `miscellaneous` | `can_access_workspace_setting` wraps the section | Approvers: allowlist. Others: always true | OK today, only because the permission is true for every non-approver |

### Top-level and workspace routes

| Route | Rule | Enforced at | Gap |
|---|---|---|---|
| `api` | not approver, not WL | WL: `beforeEnter`. Approver: allowlist | OK, two places |
| `ai_studio*`, `media-library` | approver needs `can_create_post` / `can_access_media_library` | Global guard | OK |
| `analytics_usermaven` | not WL | `beforeEnter` | OK |
| `inbox-revamp` | `social_inbox` plan feature | `InboxView.vue:678` in-component | Plan gate, page-level |
| `manageLimits`, `manageTeam`, `createWorkspace` | none found | Nowhere | Confirm the intended audience |

### Dead code

`automationFeatureLockCheck`, `influencerFeatureLockCheck`, `analyticsFeatureLockCheck` are exported from `router.ts` and have **zero live references**. The analytics ones are commented-out `beforeEnter` lines in `analytics.ts` / `competitor.ts`.

---

## Root causes

1. **Two permission composables with the same logic.** `src/composables/usePermission.ts` (18 importers) and `src/modules/common/composables/usePermissions.ts` (100 importers) both define `hasPermission` and `hasRoutePermission`. The approver route allowlists are identical today but kept by hand. 375 `hasPermission` call sites.
2. **Three ways to detect white-label:** `window.isWhiteLabelDomain` (set in `index.html`), `isWhiteLabelDomain()` in `src/config/api-utils.ts`, and the `useWhiteLabelApplication().isWhiteLabelDomain` computed. On top of that there is `shouldShowWhiteLabelData` (WL domain **and** WL workspace). The menus use the domain rule for White Label / SSO and the domain + workspace rule for Email Status / Refer & Earn. 170 files reference one of these.
3. **The route guard has no way to say who may open a route.** Route `meta` holds only `title`, `guest`, `analyticsPageView` and similar. Nothing declares access, so every new page invents its own check.
4. **Plan load race.** The plan loads without blocking during boot. Pages that need plan data (`WhiteLabelMain`, SSO, approval workflows, inbox) call `waitForPlan()` in `onMounted`. That is part of why checks drifted into components. The global guard already waits for the plan at one point (analytics entry, step 4.6), so the guard can do this too.
5. **Inconsistent fallbacks.** A denied user lands on `profile`, `home`, `planner_list_v2`, `analytics_overview_v3` or `feed_view`, depending on which page denied them.
6. **Workspace slug vs active workspace.** `hasRoutePermission` reads the role from the *active* workspace. `listening/routes.ts` already resolves the workspace from `to.params.workspace`, because a typed URL for another workspace can arrive before the switch. A central guard has to settle which workspace's role applies.

---

## What needs to change

- **One declaration per route** of who may open it: allowed roles, required permission (e.g. `can_see_subscription`, `can_view_team`, `manage_workflow`), and a white-label rule (block on WL domain, or block when WL data is shown). Child routes inherit from their parent.
- **One global guard** evaluates that declaration on every navigation, including back/forward and cold deep links. It waits for profile, workspace and plan where needed, and sends a denied user to one consistent fallback.
- **One permission source:** fold `usePermission.ts` into `usePermissions.ts` (or the reverse), so the guard, menus and pages ask the same function.
- **One white-label source** for both "WL domain" and "show WL data".
- **Menus read the same route rules**, so "hidden in the menu" and "blocked by URL" cannot drift apart: `SettingSidebar` (both renderings), `HomeSettingsDropdown`, `useHeaderNavigation`.
- **Delete the page-level redirects** listed above once the route covers them.
- **Close the gaps:** `emailNotificationStatus`, `refer&earn`, `affiliate` on WL; `refer&earn` for clients; confirm `manageLimits` / `manageTeam` / `createWorkspace`.
- **Delete the dead lock helpers.**
- **Regression net:** a test that fails when an authenticated route has no access declaration, plus guard tests per role × WL (`tests/router/` already has guard test files to extend).

### Open questions for the PO / tech lead

1. **Plan and add-on gates** (inbox, discovery, White Label / SSO / approval-workflows add-ons, Looker Studio, listening) are a third kind of check, separate from role and white-label. Move them into the same layer now, or leave them out of scope for a follow-up?
2. **Denied fallback:** one destination for everyone (e.g. workspace home with a "You don't have access to that page" toast)? Or keep approvers on the planner list, as today?
3. **`affiliate` route:** remove it, or keep it and block it on WL?
4. **`manageLimits` / `manageTeam` / `createWorkspace`:** who is meant to reach them?

---

## Suggested split (for discussion)

No `[Design]` story: nothing visual changes. No `[BE]` story: the backend already authorizes its own APIs, and this work only stops the app from opening pages a user shouldn't see.

**Recommendation: 2 `[FE]` stories, plus an optional third.**

1. **`[FE]` Enforce role and white-label page access from the route guard.** Build the single declaration + global guard, merge the permission composables, unify white-label detection, declare every authenticated route, close the gaps, consistent fallback, guard tests. This is the change that fixes the bug class.
2. **`[FE]` Remove page-level access checks and drive menus from the route rules.** Delete the `onMounted` redirects, point `SettingSidebar` (both renderings), `HomeSettingsDropdown` and `useHeaderNavigation` at the same rules, remove the dead helpers, and add the "every route declares access" test. Depends on story 1. Kept separate because it touches many view files and is easier to review and QA as its own PR.
3. **(Optional) `[FE]` Move plan and add-on gates onto the route guard.** Only if question 1 is answered "yes, but later".

**Why not one story:** a single story touches the router, both composables, ~10 page views and 3 menus. QA has to walk the whole role × white-label × page matrix in one pass, and if one part regresses the whole thing gets reverted. **Why not more:** splitting per module (settings / analytics / publisher) brings back the problem we're fixing, a partial rollout where some pages are on the new guard and some aren't.

---

## PO decisions (2026-10-23)

1. Plan and add-on gates are in this phase. Epic with 3 stories (see `02-epic-and-stories.md`).
2. Denied access auto-redirects as today, no new message. Existing notices (API plan toast, WL "no active workspace" alert) unchanged.
3. Affiliate page is blocked on white-label.
4. Manage Team / Manage Limits follow the pages they manage (Team Members, Social Accounts), and stay in sync both ways. Create Workspace: whoever can create today keeps that.

### Follow-up findings behind decision 4

- **Manage Team and Manage Limits exist twice:** standalone routes (`/manage_team`, `/manage-limits`) and tabs on `/workspaces?tab=manage_team|manage_limits`. The tabs are a query param, not a route, so a route guard on the route name alone won't cover them.
- **Workspace switcher menu** (`src/composables/useWorkspaceSwitcherMenu.ts`): Manage Team shows for `can_view_team` and not collaborator/approver. Manage Limits shows for `can_access_workspace_setting` and not collaborator/approver, which in practice means super admin + admin.
- **Workspaces page tabs** (`src/modules/setting/components/workspace/Listing.vue:35-62`) only check that org members exist. **No role check**, so a collaborator can open `?tab=manage_limits` by URL.
- **Manage Limits actions** (remove workspace, team member, social account, automation in `manage-limits/tabs/*`) have **no per-action permission check**. That's safe today only because the page lists `filteredWorkspacesOwnedByMe` (owner = super admin). Giving admins access means the list must widen to workspaces where they can act, or they see an empty page. Written into the story.
- **Permission sources to sync with:** Team page actions = `can_manage_member` (`Team.vue:99`, falls to the default branch, so true for super admin + admin only). Social account removal = `can_save_social` (`useAccountListing.ts:205`). Workspace removal = backend-resolved `workspace.policy.can_remove` (owner, or admin with `canDeleteWorkspace`, `WorkspaceActiveTile.vue:283`).
- **Create Workspace** today = `can_save_workspace` (`useWorkspaceSwitcherMenu.ts`), route itself unguarded.
- **Approvers and workflow management (corrected 2026-10-23):** the Team member permission picker (`team-permission-options.ts`) offers "Manage approval workflows" to **collaborators only**. Approvers get edit posts, add planner notes, create posts and shared folder. The approver branch reading `allow_workflow_management` in both permission composables has no UI path to set it. Unverified whether the flag survives a collaborator being changed to approver (role change only clears `canDeleteWorkspace`, commit `0444ec7815`). Candidate dead code for story 1 cleanup.
- There is also `src/composables/useRouteGuard.ts`. Despite the name it only decides chrome visibility (top bar, sidebar) by route name. It's another hand-kept route list.

### Plan / add-on gate inventory (for story 3)

| Gate | Where enforced today |
|---|---|
| `social_inbox` | `InboxView.vue:678`, in-page redirect home |
| `content_discovery` | Desktop rail lock only (item list in `useHeaderNavigation.ts:77`, rendered by `DesktopNavigationRail.vue`). **No route or page check found** in `discovery_v2`. Confirm in app |
| Listening | Route `beforeEnter` |
| `white_label_addon`, `saml_sso_addon` | `WhiteLabelMain.vue` / `DomainVerification.vue` onMounted (SSO also shows an upgrade overlay when not unlocked) |
| `approval_workflow` | List + builder onMounted. Builder skips `waitForPlan()` |
| `grid_view_planning` | `GridView.vue:94` in-page |
| `looker_studio` | Route `beforeEnter` (`competitor.ts`) |
| API plan, sample workspace | Global guard |
| `competitor_analytics`, `campaign_label_analytics`, automations (`rss_scheduling`, `evergreen_recycling`, `bulk_upload_csv`) | In-page upgrade prompt / modal on action. Deliberate, keep |

## Files involved

- `contentstudio-frontend/src/router.ts`: global `beforeEach` (step 3.5), `api` / `home` `beforeEnter`, dead lock helpers
- `contentstudio-frontend/src/composables/usePermission.ts` and `src/modules/common/composables/usePermissions.ts`
- `contentstudio-frontend/src/config/api-utils.ts`, `src/modules/setting/composables/whitelabel/useWhiteLabelApplication.ts`, `index.html` (`window.isWhiteLabelDomain`)
- `contentstudio-frontend/src/modules/setting/config/routes/setting.ts`
- `contentstudio-frontend/src/modules/analytics/config/routes/analytics.ts`, `competitor.ts`
- `contentstudio-frontend/src/modules/listening/routes.ts`, `src/modules/ai-studio/config/routes.ts`, `src/modules/inbox-revamp/routes.ts`, `src/modules/publisher/config/routes/publisher.ts`
- Page checks to remove: `EnrolledPlanView.vue`, `Team.vue`, `WhiteLabelMain.vue`, `DomainVerification.vue`, `ApprovalWorkflowsList.vue`, `ApprovalWorkflowBuilder.vue`, `InboxView.vue` (plan gate, only if question 1 is in scope)
- Menus: `SettingSidebar.vue`, `src/components/layout/HomeSettingsDropdown.vue`, `src/components/layout/useHeaderNavigation.ts`
- Tests: `tests/router/*.test.ts`, `src/modules/setting/config/routes/__tests__/setting.spec.ts`, `src/composables/usePermission.spec.ts`
