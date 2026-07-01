# 01 — Research / Grounding: Facebook Reels Collaboration

> **Note:** The competitor/industry research step was skipped at the user's request. This doc captures the **codebase analysis** (frontend social composer + backend publishing) and the **Meta API verification** that ground the feature. It is not a gated deliverable — it exists for the paper trail.

---

## TL;DR — the scope reframe

ContentStudio **already publishes Facebook reels end-to-end** and **already exposes a "Reel" post type for Facebook in the composer**. The only missing piece is the **collaborator invite**, which exists for Instagram but not Facebook.

| Capability | Instagram | Facebook | Gap? |
|---|---|---|---|
| Reel publishing | ✅ | ✅ (`processVideoPost` — init → upload → poll → publish) | No |
| "Reel" post type in composer | ✅ | ✅ (`FacebookOptions.vue`) | No |
| Invite a collaborator | ✅ (`instagram_collaborators`) | ❌ | **This feature** |

So the feature = **bring an Instagram-style collaborator invite to Facebook reels**, adapted to Facebook's (quite different) collaborator API.

---

## Meta Facebook Reels API — verified facts

Source: [Meta — Publish a Reel](https://developers.facebook.com/docs/video-api/guides/reels-publishing/) (the link the user provided).

**Publishing flow (already implemented in ContentStudio):**
1. `POST /{page-id}/video_reels` with `upload_phase=start` → returns `video_id` + `upload_url`
2. Upload binary to `rupload.facebook.com/video-upload/{video-id}` (`application/octet-stream`, `offset` + `file_size` headers)
3. `POST /{page-id}/video_reels` with `upload_phase=finish`, `video_state=PUBLISHED`, `video_id` (+ optional `description`, `title`)

**Collaborator invite (the new part) — materially different from Instagram:**
- **Separate endpoint:** `POST /{video-id}/collaborators` — *not* a parameter on the publish/finish call.
- **Parameter:** `target_id` = the collaborator's **Facebook Page ID** (or delegate Page ID for New Page Experience).
- **Pages only:** "You can only publish Reels to other Facebook Pages." Personal profiles cannot be collaborators.
- **Acceptance required:** "When they accept the invitation, the reel will immediately be published on their Facebook Page." Until then, nothing appears on the collaborator's Page.
- **Rate limit:** "You can only send 10 collaborator invitations per Page per 24 hours."
- **Timing:** invite is sent *after* the reel is published (it references the published `video-id`).

> **Instagram vs. Facebook collaboration — the key difference:** Instagram passes an inline `collaborators` array of **usernames** on the media container in a single call; Meta resolves and notifies them. Facebook requires a **second call** to `/{video-id}/collaborators` keyed by **Page ID**, only Pages qualify, and there's a hard daily rate limit. Third-party tools (e.g. PostFast) let users type a username and resolve it to a Page behind the scenes — so a username-style input is feasible UX as long as we resolve to a Page ID server-side.

---

## Frontend (contentstudio-frontend, Vue 3) — findings

**Composer module:** `composer_v2`
- Entry/state monolith: `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue`
- Hub: `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue`
- Initial state defaults: `contentstudio-frontend/src/modules/composer_v2/views/composerInitialState.ts`

**Instagram collaborator UI — the pattern to mirror:**
- `contentstudio-frontend/src/modules/composer_v2/components/ChannelOptions/InstagramOptions.vue`
  - Collaborator input + saved-collaborators dropdown + tag list
  - Max **3** collaborators per post; not allowed on `story`
  - State: `instagramCollaborators: []` in `SocialModal`, emitted via `setInstagramCollaborators`
  - Saved collaborators: `workspace.social_settings.instagram.collaborators[]`
  - Feature-gated: `canAccess('insta_collab_post')` → lock icon + upgrade modal when not entitled
  - Sent to backend as `instagram_collaborators` in the publish payload
  - i18n: `composer.instagram_options.collaborators.*` in `src/locales/<locale>/composer.json`

**Facebook options today:**
- `contentstudio-frontend/src/modules/composer_v2/components/ChannelOptions/FacebookOptions.vue`
  - Already supports post types: `feed`, `reel`, `story`; share-to-story toggle; video title; device selection; reel/story eligibility notices ("only available for Pages…")
  - **No collaborator UI** — this is the gap
- Facebook composer state: `facebookOptions { posted_as, post_type, facebook_background_id, facebook_share_to_story }`

**Reuse:** the Instagram collaborator input/tag/dropdown block, add/remove/validation logic, feature-gating pattern, and saved-collaborators API pattern are ~80% reusable. **New:** `facebookCollaborators` state + handler, conditional render (Facebook + Page account + reel post type), payload mapping, a Page-resolution/validation step (Facebook-specific), and i18n keys.

**i18n:** `src/locales/en/composer.json` (namespace `composer`), mirrored to `de, fr, pl, es, it, el, zh`.

---

## Backend (contentstudio-backend, Laravel 10) — findings

**Facebook publishing service:** `contentstudio-backend/app/Libraries/Integrations/Platforms/Social/Facebook/FacebookPlatform.php`
- `performPosting()` → `executePosting()` → routes by post type → `reelPost()` → `processVideoPost()` (the 3-step init → upload → poll → publish flow — already matches Meta's reels API)
- Dispatched by `contentstudio-backend/app/Jobs/PlatformPostingJob.php`
- Graph client: Facebook PHP SDK; HTTP via `FacebookHelper::fbHttpPostRequest()` (adds `appsecret_proof`)
- Graph version: `env('FACEBOOK_GRAPH_VERSION')` (config/integrations.php); Page token from `FacebookAccounts.long_access_token`

**Instagram collaborator handling — pattern to mirror:**
- `InstagramPlatform::addCollaborators($plan, $payload)` — strips `@`, adds `collaborators` array to the media-container payload (inline, single call). Fire-and-forget; no acceptance tracking, no webhook.

**Post data model:** `contentstudio-backend/app/Models/Publish/Planner/Plans.php`
- `instagram_collaborators` (fillable), `facebook_options` (fillable); DTO `app/Data/Planner/PlanData.php`
- `facebook_options` shape in `config/socialPost.php`: `{ posted_as, post_type, facebook_background_id, facebook_share_to_story }` — **no `facebook_collaborators` yet**

**What's new on the backend (vs. Instagram's inline approach):**
- Add `facebook_collaborators` to `facebook_options` / Plan / DTO
- After `processVideoPost` publishes the reel and we have the `video_id`, make the **separate** `POST /{video-id}/collaborators` call(s) with `target_id` per collaborator
- Resolve a user-entered Page handle/URL → Page ID (Graph lookup), or require the caller to supply a Page ID
- Handle Facebook-specific failure modes: invalid/non-Page target, rate limit (10/Page/24h), reel published but invite failed (best-effort, non-blocking)
- **No webhook** infra exists for collaborator acceptance/rejection (Instagram doesn't track it either) — acceptance tracking would be a v2 effort

**Multi-step upload pattern available for reference:** `InstagramResumableUpload.php` and the existing `processVideoPost()` already model init → upload → poll → publish with retries/timeouts.

---

## Constraints that fall out of the API

- **Facebook collaboration is reel-only** (the collaborators edge lives on `video_reels`). Unlike Instagram (feed + reel + carousel), there's no feed/story collaboration on Facebook.
- **Publishing account must be a Facebook Page** (reels are Page-only) **and** the collaborator must be a **Page**.
- **Collaborator value is a Page**, not an arbitrary username — UX must make this clear or resolve a handle/URL to a Page ID.
- **Rate limit 10/Page/24h** must surface as a user-facing error when exceeded.
- **Best-effort invite:** the reel publishes regardless; a failed invite should not fail the post.
