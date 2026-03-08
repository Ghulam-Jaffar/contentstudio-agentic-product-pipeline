# Research: Set Default Workspace on iOS & Android

## Current State

### Backend — API already exists
- `POST /changeDefaultWorkspace` in `contentstudio-backend/app/Http/Controllers/Settings/WorkspaceController.php:850`
- Accepts `{ workspace_id: string }` + authenticated user ID
- Sets all user workspace memberships to `default: false`, then sets the target to `default: true`
- Returns `{ "status": true }` (Boolean) on success
- No changes needed to the backend.

### Web Frontend — Already fully implemented
- `changeDefaultWorkspace()` in `contentstudio-frontend/src/composables/useWorkspaceSwitcher.js:434`
- The workspace switcher on web has a "Set as Default" affordance that calls `POST changeDefaultWorkspaceUrl` with `{ workspace_id }`
- Web is complete — no FE story needed.

### iOS — Broken implementation
- `contentstudio-ios-v2/ContentStudio/Controllers/Workspace/WorkspaceViewController.swift`
- `WorkspaceViewController` shows two UITableView sections: "Default Workspace" (current default) and "All Workspaces" (all workspaces)
- Tapping a row in "All Workspaces" triggers `setDefaultWorkspace(wsId:selectedIndex:)` → calls `ServiceManager.sharedInstance.SetDefaultWorkSpaces(id:completionHandler:)` → `POST /changeDefaultWorkspace`
- **Critical bug:** Response check at line 177 reads `respDic["status"] as? Int == 1` but the backend returns `Bool`, not `Int`. The cast always fails → the response handler never fires → workspace selection change never takes effect in the UI. From the user's perspective, tapping a workspace appears to do nothing.
- Also: `Constant.changeDefaultWorkspaceUrl` is already defined in `Constant.swift`

### Android — Working API call but poor UX
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Workspace/WorkspaceActivity.java` — workspace list screen
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Workspace/AdapterWorkspace.java` — adapter with `handleClick()`
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Network/ServiceManager.java:140` — `@POST("changeDefaultWorkspace")` Retrofit call exists
- Current UX: tapping any non-default workspace card calls `changeDefaultWorkspace` immediately with no confirmation or explicit "Set as Default" affordance — the intent is ambiguous (is it switching to this workspace or setting it as default?)
- No visible "Set as Default" button/label on workspace rows — the action is hidden/unclear

## What Needs to Change

### iOS
- Fix the `setDefaultWorkspace` response handler: change `as? Int == 1` → `as? Bool == true` (or `as? Int == 1` → recheck against the actual bool return from `status`)
- Add a clear "Set as Default" touch target per workspace row (e.g., a ☆/★ star icon or a contextual button) so the action is discoverable
- Show a success toast / confirmation message when default is set
- Handle error states: network failure, locked workspace
- On app launch, load the workspace marked as `default: true` from the API response automatically

### Android
- Add an explicit "Set as Default" button or long-press context action per workspace card in `AdapterWorkspace`
- Differentiate clearly between "switch to workspace now" vs "set as default on next launch"
- Show a success snackbar/toast when the default is set
- Handle edge cases: locked workspaces cannot be set as default (show dialog), already-default workspace shows visual indicator and "Set as Default" button is hidden/disabled
- On app launch, respect the `default: true` field from the workspaces API response

## Edge Cases

| Case | Expected Behavior |
|---|---|
| **Locked workspace** | "Set as Default" button hidden/disabled; tapping shows "This workspace is locked and cannot be set as default." |
| **Already default** | No "Set as Default" option shown; a visual badge/indicator ("Default") is displayed instead |
| **Only one workspace** | "Set as Default" button hidden — it's already the default |
| **Network failure** | Show error: "Unable to update your default workspace. Please try again." |
| **Approver role** | Can set default workspace (no restriction — they can switch workspaces already) |
| **Guest/pending invite** | N/A — pending invites don't appear in the workspace list |

## Files Involved

**iOS:**
- `contentstudio-ios-v2/ContentStudio/Controllers/Workspace/WorkspaceViewController.swift`
- `contentstudio-ios-v2/ContentStudio/Views/Workspace/Cell/WorkspaceTableViewCell.swift`
- `contentstudio-ios-v2/ContentStudio/HelperClasses/Constant.swift` (URL constant already exists)
- `contentstudio-ios-v2/ContentStudio/HelperClasses/ServiceManager.swift` (API call already exists)

**Android:**
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Workspace/WorkspaceActivity.java`
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Workspace/AdapterWorkspace.java`
- `contentstudio-android-v2/app/src/main/java/com/muneeb/lumotive/Network/ServiceManager.java` (API call already exists)
- Layout files: `row_layout_workspace.xml` (add "Set as Default" button)
