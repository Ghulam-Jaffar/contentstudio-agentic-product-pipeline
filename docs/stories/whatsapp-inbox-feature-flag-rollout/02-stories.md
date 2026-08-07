# Story: Feature-flag the WhatsApp inbox integration and release to production

## [FE/BE] Gate the WhatsApp connection flow and inbox integration behind a feature flag, then ship to production

### Description

The WhatsApp Business integration is already built end to end: the connection flow (Meta OAuth connect and callback) and the inbox integration (incoming conversations and replies) across the backend, the social inbox service, and the frontend. WhatsApp is an inbox-only platform. It must not appear in Composer, Planner, or Analytics.

We want to merge this work and deploy it to production without exposing it to users yet. Put the entire WhatsApp surface behind a single feature flag (proposed name `whatsapp`), served in the same per-user feature-flags payload the app already uses for Meta Ads. In production the flag is OFF by default, so nothing changes for anyone. We then turn the flag on per workspace or per user for a controlled rollout. Making WhatsApp generally available to everyone is a later, separate decision.

The gate must be applied on both sides: the frontend hides WhatsApp from the connect flow and Inbox when the flag is off, and the backend blocks the WhatsApp connect, callback, and inbox endpoints for users who do not have the flag, so the feature cannot be reached by URL or API when it is off.

Mirror the existing Meta Ads pattern (`meta_ads` flag) rather than inventing a new mechanism.

### Workflow

1. A user without the WhatsApp flag (the default in production) sees no WhatsApp anywhere: not in the social connect modal or platform lists, and not in the Inbox. Their experience is identical to today.
2. An admin enables the WhatsApp flag for a workspace or user.
3. That user now sees WhatsApp as a connectable channel in the social connect flow. They connect their WhatsApp Business account through the Meta redirect and complete the callback.
4. The connected account shows with its phone number and an editable channel name.
5. Incoming customer conversations appear in that workspace's Inbox. The user can reply with text and media inside Meta's 24-hour service window.
6. If the flag is later turned off for that user, WhatsApp is hidden again in the UI, per the agreed behaviour for already-connected accounts (see acceptance criteria).

### Acceptance criteria

- [ ] A single feature flag (proposed `whatsapp`) controls the entire WhatsApp surface. It is delivered in the same per-user `feature_flags` payload already used for `meta_ads`.
- [ ] The flag can be turned on per workspace or per user without a deploy, so we can roll out gradually.
- [ ] Flag OFF (production default): WhatsApp does not appear in the social connect modal, platform lists, or the Inbox, and the app behaves exactly as it does today.
- [ ] Flag OFF also blocks WhatsApp on the server: the connect, callback, and WhatsApp inbox endpoints reject or hide WhatsApp for users without the flag, so it cannot be reached by direct URL or API.
- [ ] Flag ON: the connect and callback flow works end to end, the account is created with its phone number and editable channel name, incoming conversations land in the Inbox, and replies send within the 24-hour window.
- [ ] WhatsApp stays inbox-only whether the flag is on or off: it never appears in Composer, Planner, Analytics, or `trigger_platform_job` payloads.
- [ ] All WhatsApp code is merged to the default branch and deployed to production with the flag OFF. No WhatsApp code path runs for users without the flag, and there is no user-visible change from the deploy itself.
- [ ] Existing non-WhatsApp connection and inbox behaviour is unaffected.
- [ ] Behaviour when the flag is turned off for an account that already connected WhatsApp is defined and implemented (recommended default: hide WhatsApp in the UI while leaving the connected account and its data intact; confirm with product before build).
- [ ] Production Meta app configuration is verified: production WhatsApp Business credentials, webhook callback URL, and verify token are set so the connect flow and webhooks work when the flag is enabled.

### UI copy

None. WhatsApp UI strings already exist. This story only gates existing screens, it does not add new copy. If a placeholder or empty state is ever needed, none is in scope here.

### Impact on existing data

None. This is gating only. No schema or data migration. Any WhatsApp accounts connected during testing remain as they are.

### Impact on other products

- **Social inbox service (`social-inbox-manager`):** WhatsApp workers, strategy, and webhook handling are already implemented. Confirm they only act on connected WhatsApp accounts, so with the flag off (and none connected for a workspace) they stay inert. No change to other platforms' processing.
- **Mobile apps ([Flutter]):** WhatsApp in the mobile Inbox is out of scope for this story. If the mobile Inbox consumes the same feature-flags payload, confirm it also hides WhatsApp when the flag is off. Flag any mobile follow-up separately.
- **Chrome extension:** N/A.

### Dependencies

- The WhatsApp integration code across backend, `social-inbox-manager`, and frontend must be complete and ready to merge (per the implemented WhatsApp Business backend and inbox work).
- Production Meta app and webhook configuration ready (credentials, callback URL, verify token).

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- **Flag mechanism (mirror Meta Ads):** frontend `useAccount().featureFlag(name)` calls `useProfileStore.hasFeatureFlag(name)`, which reads `profile.feature_flags`. Meta Ads is gated with `featureFlag('meta_ads')` in `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/composables/useSocialPlatforms.js` and in the social connect modal. Add WhatsApp to the connect flow and Inbox behind `featureFlag('whatsapp')` the same way.
- **Backend gate + connect flow:** generic `/{platform}/connect` route via IntegrationBuilder to IntegrationConnector to `app/Strategy/Integrations/WhatsappConnector.php`; controller `app/Http/Controllers/Integrations/Platforms/Social/WhatsappController.php`; config `config/integrations.php` (whatsapp block and channel lists); `app/Repository/Integrations/Platforms/Social/WhatsappRepo.php`; routes `routes/web/integrations.php`. Enforce the `whatsapp` flag server-side on connect, callback, and WhatsApp inbox endpoints, and make sure the per-user `feature_flags` payload can include `whatsapp`.
- **Inbox service:** `social-inbox-manager/app/social_sync/whatsapp_strategy.py`, `app/workers/whatsapp_inbox_worker.py`, `app/api/webhooks/whatsapp_webhook_actions.py`, `app/database/mongo/repository/whatsapp_account_repository.py`. Already implemented; ensure processing is scoped to connected accounts.
- **Scope contract:** `contentstudio-backend/docs/prd/PRD-WhatsApp-Backend-Inbox.md` (inbox-only, no historical sync, 24-hour window, strip WhatsApp from `trigger_platform_job`).
