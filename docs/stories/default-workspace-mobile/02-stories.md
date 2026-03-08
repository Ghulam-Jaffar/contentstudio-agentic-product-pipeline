# Stories: Set Default Workspace on iOS & Android

---

## Story 1: [iOS] Add "Set as Default" workspace option in the workspace list

### Description:

As a ContentStudio user on iOS, I want to be able to mark any of my workspaces as my default so that whenever I open the app, it automatically loads that workspace without me having to manually switch every time.

---

### Workflow:

1. User opens the app and taps the workspace switcher from the side menu.
2. User sees the workspace list screen. Their current default workspace is shown at the top with a "Default" badge next to its name.
3. For every other workspace in the list, user sees a "Set as Default" button (star icon or labelled button) next to the workspace row.
4. User taps "Set as Default" on any non-default workspace.
5. A loading indicator appears while the request is in progress.
6. On success, the previously-default workspace loses its "Default" badge and the selected workspace now shows the "Default" badge. A success toast appears: "Default workspace updated."
7. On next app launch, the app automatically opens into the workspace the user set as default.

**Edge cases:**
- If the workspace is **locked** (has a payment issue), the "Set as Default" button is hidden. The locked badge is already visible; no separate action is available.
- If the user has **only one workspace**, no "Set as Default" button is shown — it is already the default by definition.
- If the **network call fails**, the UI reverts to its previous state and shows an error toast: "Unable to update your default workspace. Please try again."

---

### Acceptance criteria:

- [ ] Each non-default workspace row shows a "Set as Default" button/icon
- [ ] The current default workspace displays a "Default" label/badge; the "Set as Default" button is not shown for it
- [ ] Tapping "Set as Default" sends `POST /changeDefaultWorkspace` with `{ workspace_id: <id> }` and handles the `Bool` response correctly (`status == true`, not `status == 1`)
- [ ] On success: the previously-default workspace loses the "Default" badge; the newly selected workspace gains it; a toast shows "Default workspace updated."
- [ ] On API failure: UI reverts; toast shows "Unable to update your default workspace. Please try again."
- [ ] Locked workspaces do not show the "Set as Default" button
- [ ] A workspace with only one entry in the list does not show the "Set as Default" button
- [ ] On fresh app launch, the app loads the workspace where `default: true` is set in the `/fetchWorkspaces` API response
- [ ] The existing response parsing bug is fixed: `respDic["status"] as? Int == 1` is corrected to `respDic["status"] as? Bool == true` in `WorkspaceViewController.setDefaultWorkspace()`

---

### Mock-ups:

N/A — to be designed by the design team if needed. See the web workspace switcher for UX reference.

---

### Impact on existing data:

No data model changes. The `default` boolean field already exists on the `WorkspaceTeam` model. The API endpoint already exists. Only client-side code changes.

---

### Impact on other products:

- **Web:** No impact — already implemented.
- **Android:** Separate story — [Android] Add "Set as Default" workspace option in the workspace list.
- **Chrome extension:** No workspace switching exists in the Chrome extension.

---

### Dependencies:

None. The backend API (`POST /changeDefaultWorkspace`) is already live.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, this is a native iOS screen
- [ ] Multilingual support — success/error toast strings and button labels must use the existing localization system; add keys for "Set as Default", "Default", "Default workspace updated.", and "Unable to update your default workspace. Please try again."
- [ ] UI theming support — N/A, iOS app does not use the web theming system
- [ ] White-label domains impact review — N/A for native iOS
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Story 2: [Android] Add "Set as Default" workspace option in the workspace list

### Description:

As a ContentStudio user on Android, I want a clear, explicit option to set any workspace as my default so that I always open into the right workspace when I launch the app, without the current ambiguous tap-to-switch-and-also-set-default behaviour.

---

### Workflow:

1. User opens the app and navigates to the workspace list screen (via the side menu or workspace switcher).
2. User sees all their workspaces listed. The current default workspace shows a "Default" badge/chip near its name.
3. For every non-default workspace, user sees a clearly labelled "Set as Default" button (e.g., a star icon button or a text button inside the workspace card).
4. User taps "Set as Default" on a workspace.
5. A progress indicator appears while the API call is in progress.
6. On success:
   - The previously-default workspace loses its "Default" badge.
   - The selected workspace gains the "Default" badge.
   - A snackbar appears: "Default workspace updated."
   - The newly selected workspace is also set as the active workspace (`Constants.activeWorkspace` updated, `PrefManager.setWorkspaceId()` called).
7. On next app launch, the splash/onboarding flow automatically opens the workspace marked `default: true` in the `/fetchWorkspaces` response.

**Edge cases:**
- If the workspace is **locked** (`has_payment_issue: true`), the "Set as Default" button is hidden on that card. The existing lock indicator already shows on the card.
- If the user has **only one workspace**, no "Set as Default" button is shown — it is already the default.
- If the **network call fails**, the UI reverts to its previous state and a snackbar shows: "Unable to update your default workspace. Please try again."
- Tapping the workspace card itself (outside the "Set as Default" button) switches the active workspace for the current session but does NOT change the default. These two actions must be clearly separated.

---

### Acceptance criteria:

- [ ] Each non-default, non-locked workspace card shows an explicit "Set as Default" button
- [ ] The current default workspace displays a "Default" badge; no "Set as Default" button is shown for it
- [ ] Tapping "Set as Default" sends `POST /changeDefaultWorkspace` with `{ workspace_id: <id> }` via `ServiceManager.changeDefaultWorkspace()`
- [ ] On success: previous default loses badge; new default gains badge; snackbar shows "Default workspace updated."
- [ ] On API failure: UI reverts; snackbar shows "Unable to update your default workspace. Please try again."
- [ ] Locked workspace cards do not show the "Set as Default" button
- [ ] A list with only one workspace does not show the "Set as Default" button
- [ ] Tapping the workspace card body (not the "Set as Default" button) does NOT call `changeDefaultWorkspace` — it only switches the active workspace for the current session
- [ ] On fresh app launch, `WorkspaceActivity.fetchAllWorkspaces()` reads the `default: true` field and correctly initialises `Constants.activeWorkspace` and `PrefManager.workspaceId` to that workspace
- [ ] `AdapterWorkspace` is updated so the "Set as Default" click is handled separately from the card click

---

### Mock-ups:

N/A — to be designed by the design team if needed. See the web workspace switcher for UX reference.

---

### Impact on existing data:

No data model changes. The `default` boolean is already stored in the `WorkspaceTeam` document and returned by `/fetchWorkspaces`. Only client-side changes in the Android app.

---

### Impact on other products:

- **Web:** No impact — already implemented.
- **iOS:** Separate story — [iOS] Add "Set as Default" workspace option in the workspace list.
- **Chrome extension:** No workspace switching in the Chrome extension.

---

### Dependencies:

None. The backend API (`POST /changeDefaultWorkspace`) is already live.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, this is a native Android screen
- [ ] Multilingual support — button labels and snackbar strings must use the existing localization system (`L.menu()` or string resources); add string resources for "Set as Default", "Default", "Default workspace updated.", "Unable to update your default workspace. Please try again.", and "This workspace is locked and cannot be set as default."
- [ ] UI theming support — N/A, Android app does not use the web theming system
- [ ] White-label domains impact review — N/A for native Android
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
