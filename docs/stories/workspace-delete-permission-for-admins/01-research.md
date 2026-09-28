# Research: Workspace delete permission for admins

**Source:** Frill idea #1363 `workspace-deletion-permission-for-admins` (`idea_10wp3prg`), status "Planned ⌛", 1 vote, 0 comments, submitted 2026-09-17 by an internal user (d4interactive).

**Idea body (verbatim):**
> Allow Super Admins to grant workspace deletion access to specific Admins without making them Super Admin, so teams with multiple admins don't need to share full account ownership just to manage workspace cleanup.

**Backlog check:** no match in `docs/features/` or `docs/stories/`. Closest prior art is `shared-folder-team-permission` (added `accessSharedFolder`) and `social-account-manage-access`. Neither touches workspace deletion.

---

## 1. Confirmed current behaviour

**Admins cannot delete a workspace from anywhere in the UI.** Verified against `contentstudio-backend` @ `c876807d3` (branch `features`, merged from `develop` 2026-09-15) and the matching frontend. Only the workspace creator sees the option.

### Why, exactly (this is the important part)

The Remove action on the workspace tile is gated on the server-computed `workspace.policy.can_remove`:

- `contentstudio-frontend/src/modules/setting/components/workspace/reusable/WorkspaceActiveTile.vue:156-160` — `v-if="item.workspace.policy && item.workspace.policy.can_remove && !item.default"`
- `contentstudio-backend/app/Models/Settings/Workspace.php:33-40` — `policy` is in `$appends`; `:147-150` builds it via `new PolicyHelper('Common', $this, Auth::id())`

`CommonStrategy` then resolves the wrong workspace id:

`contentstudio-backend/app/Libraries/Permission/Strategy/CommonStrategy.php:37-40`
```php
$permission = app(PolicyPermissionCache::class)->permissionFor(
    $this->user,
    $this->document['workspace_id'] !== '' ? $this->document['workspace_id'] : $this->document['_id']
);
```

A `Workspace` document has **no `workspace_id` attribute** (it *is* the workspace), so `$this->document['workspace_id']` is `null`. `null !== ''` is **true** under strict comparison, so the `_id` fallback never fires and `null` is passed through. From there:

- `app/Libraries/Permission/PolicyPermissionCache.php:29-31` — `! $workspaceId` → `new PermissionHelper($userId, null)`
- `app/Libraries/Permission/PermissionHelper.php:37` — `if ($user_id && $workspace_id)` is false, so `$member` stays `[]`
- `app/Libraries/Permission/PermissionHelper.php:98` — `is_object([])` is false → `return false`

Net: `can_remove` collapses to `checkDocumentOwnership()` alone (`CommonStrategy.php:53-56`, `document.user_id == auth id`). That is the workspace creator, who is the only user ever assigned `super_admin` (`app/Http/Controllers/Settings/WorkspaceController.php:289`).

**This bug is specific to `Workspace`.** The three other models appending a `'Common'` policy (`RssAutomation`, `SocialIntegrations`, `EvergreenAutomation`) all carry a real `workspace_id`, so the strategy behaves correctly there. `Workspace` is the one case the `_id` fallback was written for and the one case it never reaches.

### Consequence for implementation

Adding a `permissions` flag and a case in `PermissionHelper`'s admin branch **will not make the button appear**, because for the workspace policy `PermissionHelper` is never consulted. The id resolution has to be fixed first (in `CommonStrategy`, or by overriding `getPolicyAttribute` on `Workspace`).

And fixing the id resolution **on its own is a regression**: `hasPermission('can_remove')` would then hit the admin branch's `default: return true` (`PermissionHelper.php:116`) and every admin would get the button. The id fix and the new gated flag must land in the same change, flag first.

### Secondary gap: the API is not protected by the same accident

`PermissionMiddleware` builds `new PermissionHelper(Auth::id(), $workspaceId)` with a **real** id (`app/Http/Middleware/PermissionMiddleware.php:61`), so the `remove_workspace` action resolves through the admin branch's `default: return true` and **passes today**. Both delete routes are affected:

- `routes/web/settings.php:99-101` — `POST /removeWorkspace`
- `routes/api/v1.php:84-88` — `DELETE /api/v1/workspaces/{workspace_id}`

So the capability is currently hidden by the UI, not denied by authorization. The same new flag has to gate the middleware action, not just the policy.

### Existing non-permission constraint that still applies

`app/Http/Controllers/Settings/WorkspaceController.php:685`
```php
if ((count($allWorkspaces) == 1 && $workspaces['default']) || ! $workspaces['default'])
```
`$allWorkspaces` is workspaces the caller **owns** (`WorkspaceRepo.php:428-431`); `$workspaces['default']` is the caller's own `workspace_team.default` flag, which is per-user (`WorkspaceTeamRepo.php:41-44`). Failing this returns `workspace.owner_cannot_leave` (`:727`, `lang/en/responses.php:193`). The frontend mirrors it with `!item.default` on the tile and `useWorkspaceCore.ts:452-460`.

**Open question for the PO:** an invited admin's only workspace is usually their default, so even with the permission granted they would still be blocked by this rule. Decide whether the rule should be relaxed for a permitted admin (it reads as a "don't strand yourself" guard for owners, not an authorization rule) or left as-is. The stories assume **left as-is**.

---

## 2. The pattern to copy: `hasBillingAccess`

The idea asks for "just like the billing permission". That pattern is:

**Backend**
- Default lives in `app/Models/Settings/WorkspaceTeam.php:42-64` (`$attributes['permissions']`)
- Declared on the DTO at `app/Data/Workspace/WorkspaceMemberPermissionsData.php:14-26`
- Sanitised on write in `app/Http/Controllers/Settings/Team/TeamController.php:438-441`:
  ```php
  $permissions['hasBillingAccess'] = false; // default value is false
  if ($newRole === 'admin' && $membership === 'team') {
      $permissions['hasBillingAccess'] = $payload['permissions']['hasBillingAccess'] ?? false;
  }
  ```
  Add path: `processSingleMemberAddition()` → `setPermission()` (`TeamController.php:215`, `:286`)
- Read in `PermissionHelper.php:110-111` (`case 'billing_access'`)

**Frontend, two registries (both must be edited)**
- `src/modules/setting/components/workspace/team/components/TeamMemberAccessSettings.vue:98-110` — the admin block (`v-if="getTeam.role === 'admin'"`) containing the billing `Checkbox`, gated by the `showBillingAccessCheckbox` prop
- `src/modules/setting/components/workspace/team/AddTeamMember.vue:242-245` — the gate itself:
  ```ts
  if (getUserRole.value !== 'Super Admin') return false
  return getTeam.value.role === 'admin' && getTeam.value.membership !== 'client'
  ```
- `src/modules/setting/components/organization/dialogs/AddMemberToOrganization.vue:54` (`ADMIN_PERM_KEYS`), `:138-149` (label map), `:152-160` (its own identical `showBillingAccessCheckbox`)

Note `team-permission-options.ts` has **no `admin` key** in `PERMISSION_OPTIONS_BY_ROLE` (`:84-152`) — admin permissions are hardcoded in the two components above, not driven by that registry. A new admin permission follows the billing path, not the collaborator/approver path.

## 3. The three invite/edit surfaces

| Surface | Entry component | Renders permissions via |
|---|---|---|
| `/workspaces?tab=manage_team` | `src/modules/setting/components/organization/ManageTeamMember.vue` (mounted at `workspace/Listing.vue:74-78`) | `organization/dialogs/AddMemberToOrganization.vue` (own registry) |
| Workspace → Settings → Team | `src/modules/setting/components/workspace/team/Team.vue` | `team/AddTeamMember.vue` → `TeamMemberAccessSettings.vue` |
| New workspace guided setup, invite step | `src/modules/onboarding/components/guided-workspace-setup/GuidedSetupInviteStep.vue:95-110` | `TeamMemberAccessSettings.vue` (same component) |

So: two code paths, three user-facing surfaces. Fixing `TeamMemberAccessSettings.vue` covers surfaces 2 and 3.

## 4. Suggested naming (not for the story body)

- Permission key: `canDeleteWorkspace`, matching the camelCase admin family (`hasBillingAccess`, `accessSharedFolder`) rather than the newer snake_case `allow_workflow_management`.
- `PermissionHelper` admin-branch cases needed: **both** `remove_workspace` (middleware) and `can_remove` (policy → the button).
- No backfill migration required: an absent key reads as `false` via `?? false`, which is the intended default. Contrast `2026_01_27_170000_add_content_category_access_to_workspace_team.php`, which needed one because its default was permissive.

## 5. Other findings

- **No Laravel policies or gates exist anywhere** in the backend (`app/Policies` does not exist). All authorization is `PermissionMiddleware` + the `PermissionHelper` switch.
- **Denials on internal SPA routes return HTTP 200** with `{status:false}` (`PermissionMiddleware.php:91-94`), so a rejected delete looks like a silent no-op rather than an error. Worth a QA note.
- **Mobile is unaffected.** `contentstudio-flutter/` has no workspace deletion and no team management UI; `lib/features/workspaces/` is selector plus controller only. Permission keys are read generically via `workspace_member.dart`'s `permissionFlag(key, orElse:)`, so a new key needs no Flutter model change. No `[Flutter]` story.
- **Usermaven:** per guidelines section 17, a permission toggle inside an existing settings form is not a tracked milestone. Recommending no new event; flagging for the PO rather than speccing one.
- **Deletion is destructive:** soft delete plus `delete_scheduled_at = now + 60 days` (`WorkspaceController.php:688`), then `DeleteWorkspaceDataJob` purges ~35 collections. There is **no restore endpoint** in the codebase (`is_restorable` is written at `:702-704` but nothing reads it outside manual ops), so UI copy must not promise self-serve recovery.
