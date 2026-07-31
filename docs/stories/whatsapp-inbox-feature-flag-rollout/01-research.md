# Research: Feature-flag the WhatsApp inbox integration for production release

## Ask

The WhatsApp Business inbox integration (connection flow + inbox conversations and replies) is already built. Put it behind a feature flag so it can be merged and deployed to production while staying hidden by default, then enabled for a controlled rollout. One combined frontend + backend story.

## What is already built (do not rebuild)

WhatsApp Business is implemented across three codebases as an **inbox-only** platform. Reference contract: `contentstudio-backend/docs/prd/PRD-WhatsApp-Backend-Inbox.md` (status: Implemented). Key points from the PRD that constrain this story:

- **Inbox only.** WhatsApp must never appear in Composer, Planner, Analytics, or `trigger_platform_job` payloads. It only shows in the connect flow and the Inbox.
- **Connection flow** uses the generic `/{platform}/connect` OAuth redirect via the same IntegrationBuilder chain as YouTube/Threads (no bespoke connect UI).
- **Webhook driven**, no historical message sync. Inbox starts empty and fills from webhooks. Strict 24-hour reply window.
- **Account identifier** is a phone number with a user-editable channel name.

Existing implementation touch points:
- Backend connect/flow: `app/Strategy/Integrations/WhatsappConnector.php`, `app/Http/Controllers/Integrations/Platforms/Social/WhatsappController.php`, `app/Repository/Integrations/Platforms/Social/WhatsappRepo.php`, `config/integrations.php` (whatsapp block + channel lists), `routes/web/integrations.php`.
- Inbox service (`social-inbox-manager`): `app/social_sync/whatsapp_strategy.py`, `app/workers/whatsapp_inbox_worker.py`, `app/api/webhooks/whatsapp_webhook_actions.py`, `app/database/mongo/repository/whatsapp_account_repository.py`.
- Frontend: WhatsApp is not yet listed as a connectable channel in `useSocialPlatforms.js` (the platforms array), so the connect entry needs to be added behind the flag; inbox surfacing follows the same gate.

## Feature-flag mechanism (mirror Meta Ads)

- Frontend: `useAccount().featureFlag(name)` delegates to `useProfileStore.hasFeatureFlag(name)`, which reads the per-user `profile.feature_flags: string[]` from the auth payload.
- Existing precedent: **Meta Ads** is gated behind the `meta_ads` flag. `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/composables/useSocialPlatforms.js` conditionally adds the Meta Ads platform only when `featureFlag('meta_ads')` is true, and the social connect modal hides it the same way.
- Backend serves the per-user `feature_flags` array in the auth/profile payload the app already consumes.

So WhatsApp gets a new flag (proposed name `whatsapp`), enabled per user/workspace. Default OFF in production, enabled for the rollout list, promoted to everyone later as a separate product decision.

## Decision for the story

One `[FE/BE]` story: gate the whole WhatsApp surface (connect + inbox) behind the `whatsapp` feature flag on both the client and the server, merge and deploy to production with the flag OFF, and enable it per workspace/user for rollout. No functional change to the already-built flow.
