# Research: LinkedIn Profile Analytics access at the super admin level

## Current State

- LinkedIn Profile Analytics sits behind the `linkedin_profile_analytics` feature flag (see `docs/features/linkedin-profile-analytics-revival/`).
- The flag is **per user**: a user passes if the flag is in the `feature_flags` array on **their own** user document, or if their email ends in `@contentstudio.io` / `@d4interactive.io` (internal staff).
- Customers ask for access, send their email, and we add the flag to that one user by hand in Mongo. Teammates in the same workspace do not see the feature unless they are added one by one.
- The same per-user check is duplicated in three places that must agree:
  - Backend: `contentstudio-backend/app/Traits/Common/HasFeatureFlags.php` (`hasFeatureFlag`, constant `FLAG_LINKEDIN_PROFILE_ANALYTICS`), also a copy on `contentstudio-backend/app/Models/Account/User.php`; route gate `contentstudio-backend/app/Http/Middleware/FeatureFlagMiddleware.php`
  - Web app: `contentstudio-frontend/src/stores/core/useProfileStore.ts` (`hasFeatureFlag`), used via `LINKEDIN_PROFILE_ANALYTICS_FLAG` in `src/utils/linkedinEntity.ts`, `src/modules/analytics/components/common/composables/useAnalyticsUtils.ts`, `src/stores/integration/useSocialAccountStore.ts`
  - Analytics pipeline: `contentstudio-social-analytics-go` `db/mongodb/feature_flags.go` (`HasFeatureFlag`)
- Workspaces already carry the owner relation (`user_id`, `super_admin_state`, `super_admin_id` in `WorkspaceRepo.php`), so "who is the super admin of this workspace" is already resolvable.

## Discrepancy to confirm

The request says the other feature flags already work at the super admin level. In the code, **every** flag (`linkedin_profile_analytics`, `threads_analytics`, `whatsapp`) goes through the same per-user check above. Account-level inheritance exists for plans, add-ons and credits (resolved from the workspace owner), not for `feature_flags`. The story is written as "behave like plan/add-on access does: resolve from the super admin". Whether to switch only this flag or the shared check for all flags is a dev decision; the story scopes it to LinkedIn Profile Analytics.

## What Needs to Change

- Access resolves from the **super admin of the workspace being used**, not from the signed-in user.
- Super admin has the flag → every team member in every workspace they own gets the feature (analytics pickers, reports, and the reconnect-for-analytics flow).
- A team member who belongs to several workspaces sees it only in workspaces whose super admin has access.
- Existing per-user grants: keep honored for a transition, so nobody loses access on release. Grants to team members (not super admins) should be moved to their super admin.
- The authorization step that requests the extra LinkedIn analytics permission (keyed today on the connecting user's flag) must follow the same rule, or a team member would connect without the permission.
- Backend, web app and Go analytics service must agree.

## Files Involved

- `contentstudio-backend/app/Traits/Common/HasFeatureFlags.php`
- `contentstudio-backend/app/Models/Account/User.php`
- `contentstudio-backend/app/Http/Middleware/FeatureFlagMiddleware.php`
- LinkedIn connect/authorization path in `contentstudio-backend/app/Repository/Integrations/Platforms/Social/LinkedinRepo.php`
- `contentstudio-frontend/src/stores/core/useProfileStore.ts` (or the profile payload returns the resolved flag so the FE needs no change)
- `contentstudio-social-analytics-go/db/mongodb/feature_flags.go`
