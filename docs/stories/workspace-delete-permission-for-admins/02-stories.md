# Stories: Workspace delete permission for admins

**Frill request:** [Workspace Deletion Permission for Admins](https://contentstudio.frill.co/b/90pqo1vj/feature-requests/workspace-deletion-permission-for-admins)

---

## [BE][FE] Add a "Can delete this workspace" permission for admins

### Description:

As a Super Admin, I want to grant workspace deletion to specific Admins, so that teams with several admins can handle workspace cleanup without me promoting anyone to Super Admin or sharing full account ownership.

Today deletion is effectively limited to the workspace owner. This story adds a single per-member permission, "Can delete this workspace", that a Super Admin switches on for an individual Admin.

It works exactly like the existing billing access permission:

- **Off by default.** Nothing changes for any existing team member until a Super Admin turns it on.
- **Only a Super Admin can set it.** No other role can grant or revoke it, including admins who already hold it.
- **Admin role only**, and only for Team membership. It is not offered for Client membership, or for Collaborators and Approvers.

The option appears everywhere a team member can be invited or edited:

1. All Workspaces, Manage Team tab, when adding or editing a member
2. A workspace's own team settings, when adding or editing a member
3. The invite step shown after a new workspace is created

The permission must gate deletion in both directions: whether the Remove option is offered for that workspace, and whether the delete request is accepted by the server. An Admin without the permission is refused even if the request is sent directly, not only hidden in the interface. Super Admins are unaffected and keep the ability to delete regardless.

---

### Workflow:

1. A Super Admin opens any screen where team members are invited or edited, and selects or edits a member whose role is Admin and whose membership is Team.
2. Alongside the existing Admin permissions, the Super Admin sees an unchecked option labelled "Can delete this workspace", with an info icon explaining what it allows.
3. The Super Admin switches the option on and saves.
4. The Admin signs in, opens the workspace list, and now sees the option to remove that workspace.
5. The Admin removes the workspace, and deletion proceeds exactly as it does for the workspace owner today, including the notification sent to workspace members.
6. The Super Admin later switches the option back off. The Admin no longer sees the option to remove the workspace, and a delete request sent directly is refused.
7. If the Super Admin changes the member's role away from Admin, or their membership to Client, the option disappears and is not saved.

---

### Acceptance criteria:

**Permission behaviour**

- [ ] A new per-member permission for deleting the workspace exists on team members, and reads as false for every member that does not have it explicitly set, including all members that exist today
- [ ] The permission is only stored for members whose role is Admin and whose membership is Team. Changing a member to any other role, or to Client membership, clears it
- [ ] Only a Super Admin of the workspace can set or change it. A request from any other role that attempts to set it is rejected and the stored value is left unchanged
- [ ] An Admin with the permission on can delete the workspace through the web app and through the public API delete endpoint
- [ ] An Admin with the permission off, or not set at all, is refused by the server on both the web app delete request and the public API delete endpoint. The public API returns 403
- [ ] An Admin with the permission off is not offered the option to remove the workspace anywhere in the interface
- [ ] A Super Admin of the workspace can still delete it regardless of this permission, exactly as today
- [ ] Collaborators and Approvers cannot delete the workspace under any combination of settings
- [ ] Deleting with the permission produces the same outcome as an owner deleting today: the workspace is soft deleted, the 60 day permanent deletion is scheduled, member postings are paused, and workspace members receive the existing deletion notification
- [ ] The existing rule that a user cannot delete the workspace currently set as their own default is unchanged by this permission
- [ ] Granting or revoking the permission takes effect for the affected Admin without a Super Admin having to re-invite them

**Interface**

- [ ] The option appears in the Admin permissions group, directly below "Can access billing", using the `Checkbox` component from `@contentstudio/ui`
- [ ] It appears on all three surfaces: All Workspaces Manage Team add and edit, workspace team settings add and edit, and the invite step shown after a new workspace is created
- [ ] It is visible only when the signed-in user is a Super Admin of that workspace, and only when the member being invited or edited has the Admin role with Team membership
- [ ] It is unchecked by default on every new invite
- [ ] An info icon sits next to the label and shows this tooltip on hover: "Lets this admin delete this workspace, along with its posts, media and connected social accounts. This cannot be undone from inside ContentStudio. Leave it off if only the workspace owner should be able to delete it."
- [ ] The saved state is shown correctly when reopening an existing Admin for editing
- [ ] The permissions summary on the member row counts and displays this option the same way it handles the other Admin permissions
- [ ] Saving uses the existing team member success and error feedback. No new toast is introduced, and on failure the option keeps the value the Super Admin selected so the save can be retried
- [ ] All new copy is added to the translation files and renders through the translation layer, with no hardcoded strings
- [ ] Design system components and theme-aware colour classes only, with no hardcoded colour values

---

### Mock-ups:

No new layout. The option reuses the existing Admin permissions block exactly as billing access is presented today.

| Element | Copy |
|---|---|
| Checkbox label | Can delete this workspace |
| Info icon tooltip | Lets this admin delete this workspace, along with its posts, media and connected social accounts. This cannot be undone from inside ContentStudio. Leave it off if only the workspace owner should be able to delete it. |

---

### Impact on existing data:

No migration required. The permission is absent on every existing team member record, and an absent value reads as false, which is the intended default. No existing member gains or loses any ability when this ships.

---

### Impact on other products:

- **Mobile apps:** none. The Flutter app has no workspace deletion and no team management screens. It reads member permissions generically, so no app change is needed for the new key.
- **Chrome extension:** none.
- **Public API:** the delete workspace endpoint becomes correctly restricted. An Admin without the permission now receives 403 where the request previously succeeded. Worth calling out in API release notes.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
