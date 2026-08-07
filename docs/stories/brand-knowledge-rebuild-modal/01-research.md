# Research: Brand Knowledge rebuild modal (one-time per user)

Both mechanics this ticket needs already exist and should be mirrored:

- **One-time-per-user modal:** `contentstudio-frontend/src/views/DashboardNew.vue` shows the `new-left-navigation-modal` once per user and never again, gated by the user preference `profile.preferences.new_left_navigation_modal` (default true, set false on dismiss). Modal is opened with `$cstuModal.show(id)`. Preferences live on the user profile, so the flag is user-level and persists across workspaces and devices.
- **Home "brand knowledge missing" banner already exists:** in the same file, `showBanner = getProfileAttempted.value && !AIUserProfile.value` (AI/brand profile comes from `useSetup()`, evaluated per active workspace). The banner CTA calls `handleBrandKnowledgeSetup()` which routes to `brand-settings`. This is the surface to use for the "switched to a workspace with no brand knowledge" case, so we do not re-show the modal.

Brand knowledge page route: `brand-settings` (with the active workspace slug).

So: add a new user preference (proposed `brand_knowledge_rebuild_modal`) for the modal, mirror the new-navigation-modal gating, and reuse the existing home banner for workspaces without brand knowledge.
