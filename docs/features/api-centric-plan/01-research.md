# Research — API-Centric Plan

## What Is This Feature?

The API-Centric Plan is a new fourth pricing tier for ContentStudio, purpose-built for users who interact with the platform primarily through the API and third-party integrations (Zapier, Make, Pabbly, n8n). These users need API access, publishing tools, and integration capabilities — without the overhead of the full social media management suite (Analytics, AI Studio, Social Inbox, Discovery).

The plan includes: API credentials management, publishing via API, content library, workspace management, team members, integrations, and campaign management. It excludes: Home dashboard, AI Studio, Analytics, Social Inbox, Discovery, Automation workflows, Blog publishing, Brand Knowledge, White Label, and SSO.

### Why Users Want This

- **Developers and automation builders** are forced into the full-suite experience, paying for features they don't use
- **Onboarding is mismatched** — the current onboarding assumes manual social media managers, not API-first users
- **Churn from mismatched expectations** — users who only need API access are underwhelmed or confused by the full platform
- **Market opportunity** — only 2 of 11 competitors (Late/Zernio and Ayrshare) offer dedicated API-first plans

---

## Competitor Analysis

### getlate.dev (Late / Zernio) — Deep Dive

Late/Zernio is a **pure API-first social media platform** — no traditional dashboard UI. Purpose-built for developers and automation users.

**Supported Platforms (14):** Twitter/X, Instagram, Facebook, LinkedIn, TikTok, YouTube, Pinterest, Reddit, Bluesky, Threads, Google Business, Telegram, Snapchat, and one additional.

**Core Capabilities:**
- Unified REST API for posting (text, images, video, carousels, Stories, Reels)
- Scheduling and queue management
- Analytics (likes, shares, reach, impressions, clicks, views) unified across platforms
- Unified inbox (DMs, comments, reviews) with programmatic reply
- Media upload/management with automatic per-platform validation
- Automatic rate limit handling with exponential backoff
- OAuth-based social account connection; Bearer token (API key) auth

**Developer Experience:**
- Official SDKs in 8 languages: Node.js, Python, Go, Ruby, Java, PHP, .NET, Rust
- API key format: `sk_` prefix + 64 hex chars
- Full documentation at docs.getlate.dev
- No-code integrations: n8n, Make, Zapier
- 99.7% uptime SLA; 2M+ posts delivered

**Late/Zernio Pricing:**

| Plan | Monthly | Annual | Posts/mo | Profiles | Rate Limit | Users |
|---|---|---|---|---|---|---|
| Free | $0 | $0 | 20 | 2 | Basic | Unlimited |
| Build | $19 | $16/mo | More | More | 120 req/min | Unlimited |
| Accelerate | $49 | $41/mo | More | More | 600 req/min | Unlimited |
| Unlimited | $999 | — | Unlimited | Unlimited | 1,200 req/min | Unlimited |

No setup fees, no hidden costs, 30-day money-back guarantee.

### Full Competitor API Access Summary

| Competitor | Has API Plan? | API Pricing Model | Key API Capabilities | Rate Limits | Developer Experience |
|---|---|---|---|---|---|
| **Late/Zernio** | Yes — API-first standalone | Free / $19 / $49 / $999/mo | Publish, schedule, analytics, inbox, media — 14 platforms | 120–1,200 req/min by tier | Excellent: 8 SDKs, full docs, no-code integrations |
| **Ayrshare** | Yes — API-first standalone | Free / $99 / $499+/mo | Publish, schedule, analytics, hashtags, RSS — 15+ platforms | Unlimited API calls on paid; post limits apply | Good: REST API, docs, multi-platform |
| **Buffer** | Yes — bundled (beta) | Included with all plans ($0–$10/channel/mo) | Publish, schedule, analytics | 60 req/user/min | Limited: new API in beta, legacy deprecated |
| **Hootsuite** | Yes — bundled (dev portal) | Free API access with approved dev account; requires $99+/mo plan | Publish, schedule, analytics, URL shortening (enterprise only) | Not documented | Moderate: dev portal with approval, REST docs |
| **Publer** | Yes — bundled (Business+) | Included with Business ($21+/mo + $7/account) | Create/schedule posts, analytics | 50 req/min | Basic: API exited beta 2025, no SLA |
| **Sprout Social** | Yes — bundled (Advanced) | Requires Advanced at $399/user/mo | Custom integrations, automated workflows | Not documented | Enterprise-only |
| **Agorapulse** | Yes — bundled (Advanced) | Requires Advanced at $149–199/user/mo | Publishing, analytics, integrations | Not documented | Limited docs |
| **Sendible** | Yes — bundled (higher tiers) | Included in higher plans ($29+/mo) | Custom integrations | Not documented | Limited |
| **SocialBee** | **No API** | N/A | N/A | N/A | N/A |
| **Loomly** | **No API** | N/A | N/A | N/A | N/A |
| **Metricool** | Yes — bundled (Advanced) | Requires Advanced ($54+/mo) | Reporting, custom integrations | Not documented | Limited: API is add-on feature |

### Common Patterns

1. **Bundled vs. Standalone:** Most traditional tools (Hootsuite, Sprout Social, Agorapulse, Sendible, Metricool, Publer) bundle API access into higher-tier plans. Only Late/Zernio and Ayrshare offer API-first standalone plans.
2. **API as Afterthought:** Most incumbents built dashboards first, added APIs later. SocialBee and Loomly still have no API.
3. **Price Gating:** Competitors gate API behind expensive tiers — Sprout Social at $399/user/mo, Agorapulse at $149+/user/mo.
4. **Rate Limits Rarely Published:** Most platforms don't publicly document rate limits.
5. **Post-Based vs. Request-Based Pricing:** Market split — Ayrshare caps posts, Late/Zernio differentiates by rate limits.

### Differentiators Worth Noting

- **Late/Zernio's Free Tier:** API access at $0 (20 posts/mo, 2 profiles) — powerful developer acquisition strategy
- **SDK Breadth:** Late/Zernio supports 8 languages — far more than any competitor
- **No-Code Bridge:** Late/Zernio targets both developers and non-technical automation users
- **Ayrshare's Usage-Based Model:** Pricing by active social profiles, not user seats

### Key Takeaways for ContentStudio

1. **The market gap is real.** Only 2/11 competitors offer dedicated API-first plans. Clear opportunity.
2. **Price Late/Zernio as the primary benchmark.** $19/mo Build plan is the closest competitor. Starting at $19–29/mo would be competitive.
3. **A free tier or generous trial drives developer adoption.** Late's free plan (20 posts, 2 profiles) is a powerful acquisition funnel.
4. **Rate limits are a key differentiator.** Developers evaluate this heavily — publish clear, competitive limits.
5. **SDK support matters.** At minimum provide Node.js and Python SDKs.
6. **ContentStudio's advantage: existing infrastructure.** Unlike Late (API-only), ContentStudio has the full suite. API plan leverages existing scheduling, content library, team management — potentially offering more than pure API tools.
7. **No-code integrations are table stakes.** Zapier/Make/n8n support is expected.
8. **Consider both post-based AND request-based limits** for clarity.

---

## Codebase Analysis

### Existing Related Code

#### Backend (`contentstudio-backend/`)

**Billing & Plans:**
- **Model**: `app/Models/Account/Subscription.php` — Collection: `subscription_plans` with fields: `slug`, `display_name`, `name`, `price`, `paddle_id`, `limits`, `features`
- **Controllers**: `app/Http/Controllers/Accounts/SubscriptionController.php` (user subscription details, stackable/addon limits), `app/Http/Controllers/Billing/PlanController.php` (billing plan management)
- **Helpers & Services**: `app/Helpers/Billing/PlanHelper.php` (price calc, addon management, billing cycle), `app/Services/PaddleBillingService.php` (Paddle integration), `app/Repository/Account/SubscriptionRepository.php`, `app/Repository/Billing/Subscriptions/BasePlanSubscriptionRepo.php`

**API Key Management:**
- **Model**: `app/Models/ApiKey.php` — Collection: `api_keys` with fields: `key`, `user_id`, `revoked`, `last_used_at`, `ai_creation`. Methods: `isValid()`, `revoke()`, `findByKey()`, `getUserKey()`
- **Controller**: `app/Http/Controllers/ApiKeyController.php` — Routes: `index()`, `store()`, `revoke()`, `regenerate()`. One key per user enforced via `userHasApiKey()`
- **Repository**: `app/Repository/ApiKeyRepo.php` — `generateApiKey()`, `createApiKey()`, `formatApiKeyData()`, rate limiting support

**Feature Gating:**
- **Middleware**: `app/Http/Middleware/SubscriptionMiddleware.php` — Feature access control based on plan features array

#### Frontend (`contentstudio-frontend/`)

**Navigation & Module Access:**
- **TopHeaderBar.vue** (`src/components/layout/TopHeaderBar.vue`) — Imports `useFeatures()`, conditionally renders nav items based on: `inboxAccess`, `blogPostAccess`, `contentFeedAccess`, `influencersAccess`, `apiKeyAccess`
- **Sidebar**: `src/components/UI/Sidebar/CstSidebar.vue` — Slot-based, flexible nav component used across modules

**Billing/Plan Store & Composables:**
- **Vuex Store**: `src/modules/setting/store/states/plan.js` — State includes `subscription` (with `slug`, `features`, `limits`, `paddle_billing`), `used_limits`, `trial_overs_in`, `is_rs_customer`, `type`. Getter: `getPlan`. Action: `fetchPlan()` POST to `/api/plan`
- **useFeatures()**: `src/modules/billing/composables/useFeatures.js` — Exports: `hasFeature(feature)`, `canAccess(feature, limitKey)`, `getLimit()`, `getRemainingLimit()`, `isLimitReached()`. Used extensively across TopHeaderBar, planner, analytics, inbox
- **useBilling()**: `src/modules/billing/composables/useBilling.js` — Plan checking, pricing logic, Paddle integration

**Feature Flags in Code:**
- `canAccess('social_inbox')`, `canAccess('blog_publishing')`, `canAccess('content_discovery')`, `canAccess('influencer_discovery')`, `canAccess('api_access')`

**Settings/Billing Pages:**
- `src/modules/setting/components/billing/` — `PlanDetailsCard.vue`, `UsageLimitsCard.vue`, `BillingHistoryCard.vue`, `UpgradePlanComponent.vue`, `ChangeTrialPlanDialog.vue`, `CancelPlanDialog.vue`

**Onboarding:**
- `src/modules/account/views/onboarding/` — Existing flow with branching (BusinessType.vue, SocialConnect.vue). Can add new API-first path.

### Reusable Components/Services

**Backend:**
1. `PlanHelper` class — plan-related business logic (pricing, tier calc, billing cycle)
2. `ApiKeyRepo` — key generation, storage, retrieval
3. `Subscription` model + Repository pattern — clean DB abstraction
4. `SubscriptionMiddleware` — feature access guards already implemented
5. `PaddleBillingService` — fully abstracted Paddle integration

**Frontend:**
1. `useFeatures()` composable — feature gate abstraction (canAccess, hasFeature, limit checks)
2. `CstSidebar` component — flexible navigation sidebar
3. `TopHeaderBar` — already conditionally renders features based on `canAccess()`
4. Plan store (Vuex) — centralized plan state management
5. `billing/components/` — reusable plan tiles, dialogs, modals
6. `usePermissions()` composable — role/permission checks alongside features
7. `@contentstudio/ui` — consistent styling system

### Integration Points

1. **Database:** Add new plan slug (`api-centric`) to `subscription_plans` collection with `features: { social_inbox: false, ai_studio: false, analytics: false, discovery: false, api_access: true, content_publishing: true, ... }` and `limits` object for API-specific limits
2. **Backend API:** Extend `PlanController` to return `user_type: 'api'` flag. Enhance `SubscriptionMiddleware` for API-first rules. Extend `ApiKeyController` for tier-specific rate limiting
3. **Frontend Store:** Extend plan state with `user_type` field. Add `isApiCentricPlan` computed. Feature gates auto-work via existing `features` object
4. **Navigation:** TopHeaderBar's `canAccess()` checks will auto-hide modules when `features.social_inbox: false` etc. — minimal changes needed
5. **Onboarding:** New flow path in `src/modules/account/views/onboarding/`. Branch on `user_type` or signup source flag
6. **API Credentials Page:** Existing endpoints in `ApiKeyController` — enhance UI with usage display per tier

### Technical Considerations

- **No breaking changes needed** — new plan type fits existing architecture
- **Feature gating already implemented** — `useFeatures()` + `canAccess()` is the exact abstraction needed
- **Plan slug pattern** — follow existing convention: `api-centric`, `api-centric-annual`
- **No explicit `user_type` field currently stored** — inferred from plan slug/features. May want explicit field for routing clarity
- **Paddle billing** — frontend checks `paddle_billing` flag. New plan uses same integration
- **Preview mode** needs new UI work — overlay/blur on locked modules (not removal), which is a new pattern
- **One API key per user** currently enforced — may need to support multiple keys for API-centric tier

### Critical Files for Implementation

| Component | File Path | Purpose |
|---|---|---|
| Plan Model | `backend/app/Models/Account/Subscription.php` | Define plan, features, limits |
| API Keys | `backend/app/Models/ApiKey.php` + Controller | Manage credentials |
| Feature Gate | `frontend/src/modules/billing/composables/useFeatures.js` | Module visibility |
| Plan State | `frontend/src/modules/setting/store/states/plan.js` | Store plan data |
| Navigation | `frontend/src/components/layout/TopHeaderBar.vue` | Conditional nav items |
| Sidebar | `frontend/src/components/UI/Sidebar/CstSidebar.vue` | Navigation structure |
| Onboarding | `frontend/src/modules/account/views/onboarding/` | New API-first flow |
| Billing UI | `frontend/src/modules/setting/components/billing/` | Plan switching UI |
| Routes | `frontend/src/router.js` | Route guards for locked modules |
