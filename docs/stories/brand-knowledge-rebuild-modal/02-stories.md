# Story: Brand Knowledge rebuild modal, shown once per user

## [FE/BE] Show a one-time Brand Knowledge rebuild modal with Rebuild, Skip, and close

### Description

Show a modal on the home dashboard telling users that Brand Knowledge has been rebuilt and inviting them to rebuild their brand. It has two CTAs, "Rebuild my Brand" and "Skip", plus a cross (X) icon in the top right. After the user acts on any of these, the modal must never appear again for that user, in any workspace and on any device. The dismissal is stored at the user level, not per workspace.

The home "set up Brand Knowledge" banner is already in production. It shows per workspace wherever a workspace has no Brand Knowledge, and that behaviour stays as is. This ticket only adds the one-time modal and the user-level dismissal. The banner is the ongoing per-workspace prompt; the modal is a single user-level announcement. In every other case (other workspaces, or after the modal is dismissed), there is no modal, only the existing banner where Brand Knowledge is missing.

Mirror the existing one-time `new-left-navigation-modal` in `DashboardNew.vue`, which is gated by a user preference. Add a new user preference for this modal.

### Behaviour

- The modal shows on the home dashboard for a user who has not dismissed it. It is gated by a user-level preference, so it is not tied to the current workspace.
- **Rebuild my Brand:** mark the modal dismissed for the user, then navigate to the Brand Knowledge page (`brand-settings` route for the active workspace).
- **Skip:** mark the modal dismissed for the user and close the modal.
- **Cross (X) icon, top right:** same as Skip. Mark the modal dismissed for the user and close it.
- After any of the three, the modal never shows again for that user.
- **Banner stays separate and unchanged:** the existing home banner keeps showing per workspace wherever Brand Knowledge is missing. The modal never behaves per workspace. So when a user is in a workspace with no Brand Knowledge, they see the banner, not the modal.

### Acceptance criteria

- [ ] The modal appears on the home dashboard for a user who has not yet dismissed it, matching the provided design (heading, body, the two CTAs "Rebuild my Brand" and "Skip", and a cross icon in the top right).
- [ ] Dismissal is persisted at the user level via a new user preference (proposed `brand_knowledge_rebuild_modal`), served in the profile payload the app already consumes.
- [ ] "Rebuild my Brand" persists the dismissal and routes to the Brand Knowledge page.
- [ ] "Skip" persists the dismissal and closes the modal.
- [ ] The cross (X) icon in the top right does exactly what Skip does: persists the dismissal and closes the modal.
- [ ] Once dismissed by any of the three, the modal never shows again for that user, in any workspace, in a new session, or on another device.
- [ ] The modal is never shown per workspace. Switching workspaces does not re-trigger it once dismissed.
- [ ] The existing home "set up Brand Knowledge" banner is unchanged and still shows per workspace wherever Brand Knowledge is missing. This ticket does not rebuild or gate the banner.

### Implementation references

*Pointers from research, not a contract.*

- One-time modal pattern to mirror: `contentstudio-frontend/src/views/DashboardNew.vue` (`new-left-navigation-modal`, gated by `profile.preferences.new_left_navigation_modal`, opened via `$cstuModal.show(id)`). Add a parallel `brand_knowledge_rebuild_modal` preference and gate the same way.
- Existing banner (already in production, do not change): same file, `showBanner = getProfileAttempted.value && !AIUserProfile.value`, CTA `handleBrandKnowledgeSetup()` routing to `brand-settings`.
- Backend: add the new preference to the user profile preferences so it persists and returns in the profile payload, alongside the existing preferences like `new_left_navigation_modal`.
