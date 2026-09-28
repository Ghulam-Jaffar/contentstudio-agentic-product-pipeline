# Research: Approver access to the Content Library

**Date:** 2026-09-21
**Trigger:** Reported inconsistency in the approver role's "Access to shared folder" permission.

---

## The reported problem

1. The "Access to shared folder" permission does nothing for an approver on its own, and should do nothing, because an approver has no reason to reach the media library.
2. But an approver who has composer access (can create or edit posts) automatically gets the media library inside the composer, so they should be able to reach the standalone Content Library too.

Both halves verified below. Both are accurate.

---

## Finding 1: the permission is inert for an approver with no post permissions

### Backend: only the media library reads it

`access_shared_folder` is consumed in exactly two controllers and nowhere else:

- `contentstudio-backend/app/Http/Controllers/Storage/MediaLibrary/MediaLibraryFolderController.php` lines 102, 195, 293, 389. Line 389 builds `$excludeGlobal` and passes it into `MediaLibraryFoldersRepo::findFolders()`, which drops `is_global` folders from the listing response.
- `contentstudio-backend/app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php` lines 100, 379, 692, 1129, 1664. These block upload / move / delete / download against a global folder.

There is **no role gate** on any media library endpoint. An approver's calls succeed; the only filter applied is the shared-folder one.

### Frontend: two gates keep the approver out of the media library

**Router guard.** `contentstudio-frontend/src/router.ts` lines 935 to 950 call `hasRoutePermission()` on every navigation, including browser back/forward. For an approver that resolves to a hard allowlist in `contentstudio-frontend/src/composables/usePermission.ts` lines 272 to 296:

```
planner_feed_v2, planner_list_v2, planner_calendar_v2, planner_tiktok_grid_view,
planner_instagram_grid_view, feed_view, list_plans, calender_plans, notifications,
profile, emailNotificationStatus, grid_tiktok, grid_instagram, setPassword,
workspaces, ai_studio, ai_studio_chat, ai_studio_history, ai_studio_skills
```

`media-library` is not on it, so a direct URL redirects to `planner_list_v2`.

**Nav rail.** `contentstudio-frontend/src/components/layout/useHeaderNavigation.ts` line 148: when `canAccessPrimaryNavigation` is false, `basePrimaryNavItems` returns a single `view-content` item and nothing else. It is always false for an approver, because `hasPermission('can_access_top_header')` falls through the approver switch to `return false` in `contentstudio-frontend/src/modules/common/composables/usePermissions.ts` lines 196 to 208. The Content Library item built at lines 328 to 343 is never reached for an approver.

So with no composer access the approver has zero code path to the folder or asset endpoints, and the checkbox at `contentstudio-frontend/src/modules/setting/components/workspace/team/team-permission-options.ts` line 146 toggles a value that nothing can observe.

### The default is on anyway

- `contentstudio-backend/app/Libraries/Permission/PermissionHelper.php` line 133 returns `$this->member->permissions['accessSharedFolder'] ?? true`.
- `contentstudio-frontend/src/modules/setting/types/team.ts` line 85 sets `accessSharedFolder: true` in `getDefaultPermissionsObject()`.

So an approver who has never been configured already has shared folder access, dormant.

---

## Finding 2: composer access grants the whole media library, modal only

An approver gets the composer through `approverCanCreatePost`, or `approverCanEditPost` on a post whose status is `reviewed` (`usePermissions.ts` lines 196 to 208). That unlocks the create-post path in the planner (`contentstudio-frontend/src/modules/planner_v2/composables/useCalendarView.ts` line 342) and the composer's own surfaces (`ComposerScheduleApproval.vue` line 6, `AccountSelectionAside.vue` lines 18 and 40).

Inside the composer the media library opens as a **modal, not a route**, so the router allowlist never fires:

```
useComposerDialogs().openMediaLibrary()
  -> EventBus.$emit('show-media-library-modal')
  -> UploadMediaModal
```

- `contentstudio-frontend/src/modules/composer/composables/useComposerDialogs.ts` line 97.
- `UploadMediaModal` is mounted globally and **unconditionally** at `contentstudio-frontend/src/Home.vue` line 1049. No permission `v-if`.
- The media library module has **no role checks at all**. The only `role ===` comparison is a super-admin name lookup in `UploadMediaModal.vue` line 433.

Result: an approver with composer access can browse, upload, create folders, and delete, but only through the composer picker. The standalone page at `/:workspace/publish/media/` stays blocked.

### Shared folder hiding is backend-only

`contentstudio-frontend/src/modules/publish/components/media-library/components/Folder.vue` line 323 gates the shared folder on `canAccess('media_library_shared_folder')`, which is a **plan feature**, not the member permission. Member-level hiding relies entirely on the backend dropping `is_global` folders from the listing. Worth knowing if the standalone page is opened up: there is no frontend fallback.

---

## Finding 3: the Flutter app has the identical gap

`contentstudio-flutter/lib/features/workspaces/domain/workspace_access.dart` line 122:

```dart
bool get showMediaLibrary => !isApprover;
```

Its own doc comment spells out the same split the web has: the Menu entry is gated, the composer's media-source sheet deliberately is not, because "an approver with `approverCanCreatePost` may draft a post, and attaching an existing asset is part of that."

The class already parses `approverCanCreatePost` and `approverCanEditPost` (lines 22, 23) and already exposes `canCreatePosts` (line 92) and `canEditPosts` (line 82), so the permission-aware version needs no new model plumbing.

Consumed by `showMediaLibraryProvider` in `contentstudio-flutter/lib/features/media_library/application/media_library_providers.dart` line 248, and the Menu entry in `lib/app/shell/home_shell.dart`.

**A second mobile difference worth a decision:** `canManageMediaProvider` at line 235 of the same file is `!isApprover`, so on mobile an approver cannot upload, delete, move, or manage folders **even inside the composer picker**. On web they can do all of that. The Flutter story below keeps that rule as-is and only opens the browsable entry point, because widening mutation rights is a separate product decision. Flag it to the PO.

---

## Existing stories in this backlog

| Story dir | Covers | Built? |
|---|---|---|
| `docs/stories/shared-folder-team-permission/` | Created this permission. Backend filtering plus the settings toggle. Never specified which roles can reach the media library, which is the gap. Shipped in commit `144c918`. | Yes |
| `docs/stories/approver-sidebar-publisher-access/` | Show the publisher sidebar to approvers without create-post | Yes. `isApprover()` is in `canAccessSidebar`, `PublisherMain.vue` lines 36 to 42 |
| `docs/stories/approver-sidebar-publisher-label/` | Rename "View Content" to "Publisher" | Yes. `useHeaderNavigation.ts` lines 151 to 156 now use `header.nav.publisher` |
| `docs/stories/approver-merged-sidebar/` | Workspace switcher at top, Settings and notifications at the foot, one sidebar, no rail | **No.** `SidebarMain.vue` has no workspace switcher and no settings foot, and `basePrimaryNavItems` still returns the single rail item |

The merged sidebar story is tracked at https://app.helpin.ai/w/contentstudio/pm/tasks/432660c4-260e-49b3-80b8-4b70a1195a9b?task=CONT-3932

### How the two interact

The merged sidebar story's rule is "when a user can reach exactly one module, there is no rail". Giving composer-enabled approvers the Content Library makes it two modules for them. Per the product decision on this work, that is intended: **approvers with create or edit post see two modules in the desktop rail, Publisher and Content Library. Approvers without either keep the single-module merged sidebar.** The merged sidebar story needs no rewrite, because its rule is already phrased on module count rather than on role.

---

## Existing copy

| Key | Current value |
|---|---|
| `header.nav.media_library` | "Library" (the rail label) |
| `header.nav.content_library` | "Content Library" (the rail tooltip) |
| `settings.workspace.team.permissions.approver.access_shared_folder` | "Access to shared folder" |
| `settings.workspace.team.tooltips.access_shared_folder` | "Allow this team member to view and use the shared folder in the Content Library. The shared folder is a common space where all permitted team members can upload, organize, and access media files together. If turned off, the shared folder won't appear in this member's Content Library at all." |

The nav item needs no new strings. The tooltip key is currently shared by admin, collaborator and approver, so splitting the approver variants means new keys rather than editing this one in place.

---

## Where the permission checkbox renders

Four surfaces show the approver permission list, all fed by `PERMISSION_OPTIONS_BY_ROLE.approver`:

- `contentstudio-frontend/src/modules/setting/components/workspace/team/components/TeamMemberAccessSettings.vue` lines 192 to 209, the `DropdownItem` loop. This is the shared renderer.
- `contentstudio-frontend/src/modules/setting/components/workspace/team/AddTeamMember.vue` lines 178 to 184, add and edit member.
- `contentstudio-frontend/src/modules/setting/components/organization/dialogs/AddMemberToOrganization.vue` lines 118 to 140, its own local copy of the option list.
- `contentstudio-frontend/src/modules/onboarding/components/guided-workspace-setup/GuidedSetupInviteStep.vue` lines 99 and 332.

`PermissionOption` in `team-permission-options.ts` has no disabled concept today, so the disabled state needs a new field on that type plus handling in the loop. `AddMemberToOrganization.vue` builds its list independently and needs the same treatment.

`APPROVER_PERM_KEYS` at line 38 drives "select all". `accessSharedFolder` is in it, so "select all" must learn to skip a disabled option.

---

## Gotchas for implementation

1. **`hasRoutePermission` is duplicated.** `contentstudio-frontend/src/composables/usePermission.ts` line 268 and `contentstudio-frontend/src/modules/common/composables/usePermissions.ts` line 214 hold two copies of the approver allowlist. Both carry "keep in sync" comments. The router uses the first, other callers use the second. Adding `media-library` means editing both, or collapsing them into one.

2. **The rail has no per-item approver path.** `basePrimaryNavItems` early-returns a hardcoded single-item array for approvers. Supporting two items means restructuring that branch rather than adding a condition to the existing Content Library push at line 328.

3. **`showApiNavItem` is the pattern to copy.** Lines 104 to 110 show how a nav item is conditionally included, though it currently excludes approvers by role name.

4. **No Usermaven event.** Section 17 of the story guidelines excludes read-only navigation and sidebar interactions. Page views are already tracked globally in the router. Nothing to instrument here.

5. **Backend needs no change.** The media library endpoints already have no role gate and already honour `access_shared_folder`. This is frontend-only on web, plus the Flutter provider.

6. **No new UI components.** The rail item, `Checkbox`, `DropdownItem`, `Icon` and `v-tooltip` all exist. `docs/ui-components.md` notes there is no standalone tooltip component, and the permissions panel already uses the `v-tooltip` directive on an `Icon`, so the disabled-state tooltip follows the existing pattern.

---

# Addendum: 2026-09-22, after the marketing and product meeting

The permission model changed. The Content Library is no longer derived from the post permissions. It becomes its own permission, and the AI chat widget is brought into scope.

## The agreed model

| Approver state | Content Library module | Composer | AI chat widget | AI Studio | "Access to shared folder" |
|---|---|---|---|---|---|
| No post permission, no library access | hidden | no | hidden | hidden | inert, greyed out |
| No post permission, library access on | **visible** | **no**, and no way in from the library | hidden | hidden | active |
| Can create or edit posts | visible, **locked on** | yes | **visible** | hidden | active |

Decisions taken in the meeting and confirmed after:

1. A new per-member permission, **"Access to Content Library"**, grants the module on its own. It does not grant the composer and it does not grant AI.
2. When an approver can create or edit posts, that permission is **forced on and cannot be unticked**, because they get the Content Library through the composer regardless.
3. An approver who has the library but no post permission must have **no route into the composer from the Content Library**, so the compose actions on assets have to go.
4. The AI chat widget follows the post permissions, not the role.
5. **AI Studio stays hidden from approvers whatever their permissions are.** See the discrepancy below.

## AI Studio is reachable today, contrary to the product's understanding

The product position is that AI Studio is not visible to approvers. In practice that is half true. The approver's nav rail collapses to a single item, so there is no entry point in the interface. But the route allowlist in `contentstudio-frontend/src/composables/usePermission.ts` lines 291 to 294 explicitly permits `ai_studio`, `ai_studio_chat`, `ai_studio_history` and `ai_studio_skills`, with a comment reading "approvers can use the modal AI chat, so the module mirrors that access."

So an approver who types or bookmarks the AI Studio address gets the full module today. Under the agreed rule those four entries should come out of both copies of the allowlist. Worth calling out to the PO as a behaviour change rather than a no-op.

## Surfaces the new rules touch

**AI chat widget.** `contentstudio-frontend/src/Home.vue` lines 1058 to 1075 mount `AIChatWidget` (the bottom-right launcher) and `AIChatModal` inside `AiChatEngineProvider`. Both are gated only on `!isApiCentricPlan` plus a route check. There is no role or permission gate anywhere on them.

**Compose actions inside the Content Library.** Three separate entry points, all of which have to be closed for a library-only approver:

- `contentstudio-frontend/src/modules/publish/components/media-library/components/Asset.vue` line 156, the `canAddToComposer` computed. Drives the per-asset actions whose copy is `publisher.media_tabs.asset.tooltips.compose_post` ("Compose Post") and `publisher.media_tabs.asset.tooltips.add_to_composer` ("Add to Composer").
- `contentstudio-frontend/src/modules/publish/components/media-library/components/MediaUploadFooter.vue` lines 42 to 50, the `showComposerButton` prop and its "Add to Composer" button, copy key `publisher.media_library.add_composer_button`.
- `contentstudio-frontend/src/modules/publish/components/media-library/MediaLibraryMain.vue` line 1119, `shareItem()`. This is the handler all three call sites reach, bound at lines 18, 57 and 143 including the asset preview modal. It ends in `$cstuModal.show('composer-modal')`. Closing the buttons without closing this leaves the composer reachable from the preview modal.

**New permission key.** `TeamController::addTeamMember()` and `updateTeamMember()` persist the whole `permissions` object from the request, so storing a new key needs no backend change. Reading it does: `PermissionHelper` needs a case, and the derived rule has to live there so the backend and frontend cannot disagree. The rule is:

```
access_media_library = accessMediaLibrary OR create_post OR edit_post
```

The media library endpoints currently carry **no role gate at all**, so this is also the first time the module is enforced server side rather than only hidden. Treat that as hardening rather than a like-for-like change, and check that nothing else was relying on the endpoints being open.

**Default for existing approvers.** `getDefaultPermissionsObject()` in `contentstudio-frontend/src/modules/setting/types/team.ts` and the backend's `?? ` fallbacks both need a considered default for the new key. Unlike `accessSharedFolder`, this one should default to **false** for approvers, otherwise every existing review-only approver silently gains the Content Library on deploy. Non-approver roles are unaffected either way, since the permission is only read for approvers.

## Mobile

The agreed AI rule was scoped to web. The Flutter app's AI assistant is untouched by this epic. The Flutter Content Library entry moves from the role check to the same derived rule as web.
