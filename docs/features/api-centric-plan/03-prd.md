# PRD: API-Centric Plan

**Author:** Product Team
**Last Updated:** 2026-03-23
**Status:** In Review
**Target Release:** Q2 2026

---

## 1. Overview

The API-Centric Plan introduces a fourth pricing tier to ContentStudio, purpose-built for developers, automation builders, and integration-first users who need API access and publishing tools without the overhead of the full social media management suite. This plan captures a growing segment of users who sign up only for API access but are currently forced through an onboarding and product experience designed for manual social media managers. The plan includes a streamlined onboarding flow, a focused in-app experience with reduced navigation, a dedicated API Dashboard, preview mode for locked modules, and a seamless upgrade path to full-suite plans.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio's current three plans (Standard $29/mo, Advanced $69/mo, Agency Unlimited $139/mo) all assume users want the full social media management suite. A growing segment of users signs up solely for API access to build automations, connect third-party tools (Zapier, Make, n8n), or publish content programmatically. These users encounter:
- An onboarding flow that asks them to explore features they'll never use
- A navigation cluttered with modules irrelevant to their workflow (Analytics, AI Studio, Social Inbox, Discovery)
- No dedicated API experience — API keys are buried in Settings, with no usage dashboard or activity logs
- Pricing that bundles features they don't need, creating a perception of poor value

**Who has this problem?**

- **Developers and automation builders** who interact with ContentStudio exclusively via API
- **Agency tech leads** who use ContentStudio as a headless publishing backend for their own tools
- **Integration-first users** who connect ContentStudio to their workflow via Zapier/Make/Pabbly/n8n
- Estimated segment: Based on API key generation rates and support ticket analysis, this represents a meaningful and growing portion of signups

**What happens if we don't solve it?**

- Continued churn from users who feel the product doesn't match their needs
- Lost revenue from developer-focused users who choose API-first competitors like Late/Zernio ($19/mo) or Ayrshare ($99/mo)
- Missed positioning opportunity as the market moves toward automation-first workflows
- Support burden from API users confused by the full-suite interface

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Capture API-first user segment | New API Plan signups | 200+ in first 90 days | Subscription analytics (filter by plan slug `api-centric`) |
| Reduce mismatch churn | Churn rate for API Plan users vs. full-suite trial users | 15% lower 30-day churn | Billing data — compare cohorts |
| Drive upgrade to full-suite | API → full-suite plan conversion | 10% within 6 months | Billing data — plan change events |
| Improve developer satisfaction | NPS for API Plan users | >40 | In-app survey after 30 days |
| Guard rail: no full-suite cannibalization | Full-suite plan signups | <3% decrease | Compare pre/post launch signup rates for existing plans |

---

## 4. Target Users

**Primary Persona:**
Developer / Automation Builder — Technical users (developers, devops, tech leads) who want to integrate ContentStudio's publishing capabilities into their own tools, scripts, or workflows. Comfortable with APIs and documentation. Values: simplicity, clear rate limits, fast onboarding, no UI overhead.

**Secondary Persona:**
Integration-First Marketer — Non-developer power users who connect ContentStudio to other tools via no-code platforms (Zapier, Make, n8n, Pabbly). May not write code directly but builds automated workflows. Values: easy setup, reliable integrations, clear limits, affordable pricing for publishing-only needs.

**Non-Users (explicitly out of scope):**
- Full-suite social media managers who want Analytics, AI Studio, or Social Inbox — these users belong on existing plans
- Enterprise customers with complex SSO/White Label requirements — API plan explicitly excludes these
- Users who need Blog publishing — removed from API plan scope

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Developer | sign up and get API credentials within 5 minutes | I can start building integrations immediately | Must Have |
| US-2 | Developer | see my API usage, rate limit status, and activity logs in one dashboard | I can monitor and debug my integrations | Must Have |
| US-3 | Automation builder | connect my social accounts and publish via API | I can automate content scheduling from my own tools | Must Have |
| US-4 | Integration-first marketer | manage content categories and campaigns via the UI | I can organize the content that my automations publish | Must Have |
| US-5 | Developer | generate and revoke API keys from the dashboard | I can manage credentials securely | Must Have |
| US-6 | API user | see only the modules relevant to my workflow | I'm not overwhelmed by features I don't use | Must Have |
| US-7 | API user | preview the full ContentStudio suite | I can evaluate if upgrading is worthwhile | Should Have |
| US-8 | API user | upgrade to a full-suite plan with my current usage preserved | I don't lose connected accounts or capacity when switching | Must Have |
| US-9 | Website visitor | understand the API Plan offering and pricing on the website | I can decide if ContentStudio is right for my API needs | Must Have |
| US-10 | API user | see a clear breakdown of costs before switching plans | there are no pricing surprises | Must Have |

---

## 6. Requirements

### 6.1 Must Have (P0)

**Backend & Data:**
- New plan type slug (`api-centric`, `api-centric-annual`) in `subscription_plans` collection with appropriate `features` and `limits` objects
- `user_type` field on user/workspace model (`api` vs `standard`) to control routing and navigation
- Smart routing: handle `?mode=api` and `?mode=suite` signup URL parameters to set plan type and skip plan selection
- API key management: existing CRUD endpoints (generate, revoke, regenerate) continue working for API plan users
- Feature gating: set `features.social_inbox: false`, `features.ai_studio: false`, `features.analytics: false`, `features.discovery: false`, `features.blog_publishing: false`, `features.white_label: false`, `features.sso: false` for API plan
- Trial support: API Trial with same duration as existing trials
- Plan switching: support upgrade from API Plan to Advanced or Agency Unlimited with auto-calculated add-ons for excess usage

**Frontend — Signup & Onboarding:**
- Plan selection screen (post-signup, pre-onboarding): two cards — Full Suite and API Plan — with clear descriptions and CTAs
- Smart routing: if `?mode=api` flag present, skip plan selection and enter API onboarding directly
- Shortened 3-step onboarding for API users: Profile (name, phone, timezone) → Role → Social Accounts (with skip option)
- After onboarding, redirect to API Dashboard (not Home)

**Frontend — Navigation & Module Gating:**
- Reduced top navigation for API plan users: show Publisher, API, Content Library, Notifications, Settings. Hide: Home, AI Studio, Analytics, Social Inbox, Discover
- Publisher module restrictions: remove Blog and Bulk Schedule via AI from Compose dropdown. Hide AI Library and Automations sidebar sections
- Settings restrictions: hide White Label, SSO, Brand Knowledge, Blogs & Websites, Other Integrations. Miscellaneous shows only Campaigns & Labels
- Basic Settings: hide Instagram Posting Method For Automation, Enable Onboarding Widget
- Route guards: redirect API users to API Dashboard if they navigate to a hidden module URL

**Frontend — API Dashboard (new page):**
- Welcome message (first visit only): heading, description, "Generate API Key" CTA, "View API Documentation" link
- API Credentials section: display API Key and Secret, generate new key, revoke existing
- API Usage section: calls used / total limit, remaining requests, rate limit status
- Activity Logs: table of recent API activity (request, endpoint, status, timestamp)
- Quick Links: API Documentation, Integration Guides, Test API Endpoint

**Frontend — Preview Mode:**
- "Explore Full Suite" CTA button persistent in top bar for API users
- Clicking it: reveal hidden nav items (Home, AI Studio, Analytics, Inbox, Discover)
- Preview modules show locked overlay/banner with: heading ("This is a preview of [Module]"), body text, "Upgrade Plan" CTA, "Go Back" button
- "Exit Preview" button in top bar to return to normal API navigation
- No functional interaction with preview modules — display only

**Frontend — Billing:**
- Billing page shows API Plan details: current plan, trial status, renewal date, API calls limit/usage, social accounts limit/usage, API keys limit/usage
- "Increase Limits" button for API-specific add-ons
- "Change Plan" flow: modal shows only Advanced and Agency Unlimited (no Standard). Selecting a plan triggers pricing breakdown modal with auto-calculated add-ons for excess usage
- Pricing breakdown: comparison table (current vs. target), base price + add-on total = total monthly price, "Confirm Switch" / "Cancel" buttons

**Website:**
- Main landing page: top announcement banner ("New: ContentStudio API Plan — Build automations and publish via API. Learn more →") linking to `/social-media-api`
- Product tabs section: new "API" tab with heading, description, and CTA
- Pricing page: API Plan card alongside existing three plans with pricing, features, "Start Free Trial" CTA linking to `signup?mode=api`
- API page (`/social-media-api`): add pricing card, API Plan vs. Full Suite comparison table, FAQ section
- Navigation: add "API" link under Product menu pointing to `/social-media-api`

### 6.2 Should Have (P1)

- API usage warning notifications: in-app alert at 80% API call usage, email notification at 90%
- Activity logs pagination and filtering (by date, endpoint, status)
- API Dashboard responsive design for tablet/mobile viewing
- Welcome email for API plan users with quickstart guide and documentation links
- Empty state for API Dashboard when no key generated yet (illustration, headline, CTA)

### 6.3 Nice to Have (P2)

- API key copy-to-clipboard with auto-hide after generation (shown once, then masked)
- Quick API test tool in dashboard (send a test request and see the response)
- API status indicator (green/yellow/red) based on platform health
- Integration template gallery (sample Zapier zaps, Make scenarios, n8n workflows)

### 6.4 Explicitly Out of Scope

- **Multiple API keys per user/workspace** — deferred to V2
- **Webhook management UI** — deferred to V2
- **SDK downloads from dashboard** — deferred to V2
- **Downgrade from full-suite to API plan** — complex data/access implications, deferred to V2
- **API-specific add-on purchasing** (extra API calls packages) — deferred pending business decisions
- **Rate limit customization** — deferred to V2
- **Developer sandbox/testing environment** — deferred to V2
- **Dark mode** — ContentStudio does not support dark mode
- **RTL language support** — ContentStudio does not support RTL
- **Mobile app changes** — API plan is web-only, no iOS/Android impact

---

## 7. User Flow (High Level)

### Signup → Onboarding → API Dashboard

1. User arrives at signup (from website API CTA with `?mode=api`, or direct)
2. If `?mode=api`: system skips plan selection → API onboarding
3. If no flag: user sees Plan Selection screen → chooses Full Suite or API Plan
4. Account created: Plan = API Trial, User Type = API, business name → workspace name
5. Onboarding Step 1: Profile (full name, phone number, time zone)
6. Onboarding Step 2: Role (existing ContentStudio role options)
7. Onboarding Step 3: Social Accounts (connect or skip)
8. User lands on API Dashboard with welcome message
9. User generates API key and starts building integrations

### Daily Usage

1. User logs in → lands on API Dashboard
2. Reviews usage metrics, checks activity logs
3. Navigates to Publisher to review scheduled content or compose new posts
4. Manages media in Content Library
5. Manages team, social accounts, or billing in Settings

### Preview & Upgrade

1. User clicks "Explore Full Suite" in top bar
2. Browses preview modules with locked overlays
3. Clicks "Upgrade Plan" → Billing → selects Advanced or Agency
4. Reviews pricing breakdown with auto-calculated add-ons → confirms switch
5. Full navigation unlocks, user type changes to Standard

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | API Plan users cannot access Analytics, AI Studio, Social Inbox, or Discovery modules (except in preview mode) | These features are excluded from the API Plan to differentiate from full-suite plans |
| BR-2 | Standard plan is not shown as an upgrade option for API users | Standard doesn't support add-ons; API users with excess limits would hit a dead-end |
| BR-3 | Plan switching preserves all connected social accounts and data | Users should not lose access to anything during an upgrade |
| BR-4 | Only one API key per user (V1) | Existing system constraint; multi-key support deferred to V2 |
| BR-5 | API Trial has the same duration as existing full-suite trials | Consistent trial experience across all plan types |
| BR-6 | Smart routing flags (`?mode=api`, `?mode=suite`) are consumed once and not persisted in the URL after signup | Clean UX — flag drives initial routing then disappears |
| BR-7 | Blog option and Bulk Schedule via AI are removed from Compose dropdown for API users | These features are excluded from the API Plan scope |
| BR-8 | Preview mode is view-only — no data creation or modification in preview modules | Prevents confusion about feature access and avoids data integrity issues |
| BR-9 | Getting Started widget is hidden for API users | Irrelevant for API-first workflow; would reference features they can't access |
| BR-10 | Brand Knowledge shortcut is hidden from top bar for API users | Feature is excluded from API Plan |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| API Plan pricing (monthly and annual)? | TBD — benchmark: Late/Zernio $19–49/mo | Business / Revenue | Before dev start | Pending |
| API request limit for API Plan? | Options: 10K, 25K, 50K, 100K/mo | Business / Engineering | Before dev start | Pending |
| Social account limit for API Plan? | Options: 5, 10, 15, 25 | Business / Product | Before dev start | Pending |
| Workspace and user limits? | Options: 1/1, 2/2, 5/5, unlimited | Business / Product | Before dev start | Pending |
| Trial duration (same as existing or different)? | Same (14 days) vs. shorter (7 days) vs. longer (30 days) | Business | Before dev start | Pending |
| API Plan add-on availability and pricing? | Enable add-ons for social accounts, API calls, workspaces, users | Business / Product | Before dev start | Pending |
| Rate limiting strategy for API Plan? | Options: requests/min, requests/hour, requests/day | Engineering | Sprint 1 | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Full-suite plan cannibalization — existing users downgrade to cheaper API plan | Medium | High | API Plan excludes high-value features (Analytics, AI, Inbox). Monitor cohort signup data. Guard rail metric: <3% decrease in full-suite signups |
| Preview mode drives upgrade expectations but UX is confusing | Low | Medium | Clear "preview only" labeling, non-interactive overlays, prominent "this is not your plan" messaging. User test the preview flow before launch |
| API pricing set too low, leaving revenue on the table | Medium | Medium | Benchmark against Late/Zernio ($19–49) and Ayrshare ($99). Start mid-range, adjust based on adoption data. Consider annual discount to increase commitment |
| Onboarding too short — API users don't connect social accounts | Medium | Low | Social account connection is included (Step 3) but skippable. Dashboard prompts if no accounts connected. API docs reference account setup as prerequisite |
| Feature gating edge cases — some UI elements slip through and show API users features they shouldn't see | Medium | Medium | Comprehensive QA checklist for every module. Automated tests for feature flags. Use existing `useFeatures()` composable which is well-tested |
| Plan switching add-on calculation is incorrect | Low | High | Unit test all pricing scenarios. Show explicit breakdown before confirmation. Add "Contact Support" fallback if calculation seems wrong |

---

## 11. Dependencies

**Internal:**
- **Backend team:** New plan type in DB, `user_type` field, mode flag handling, subscription middleware updates, onboarding flow branching
- **Frontend team:** API Dashboard page, navigation gating, preview mode UI, billing flow updates, onboarding flow
- **Design team:** Plan selection screen mockups, API Dashboard layout, preview mode locked overlay, website banner/page designs
- **Business/Revenue team:** Pricing decisions, limit decisions, add-on pricing — MUST be finalized before development begins

**External:**
- **Paddle:** New plan configuration in Paddle billing system (plan ID, pricing, billing intervals)
- **Website CMS:** Landing page updates, pricing page card addition, API page sections — coordinated with marketing/web team

**Blockers:**
- Pricing, limits, and add-on structure must be finalized by Business team before Sprint 1
- Design mockups needed for: plan selection screen, API onboarding, API Dashboard, preview mode overlay, website changes

**Existing Code Dependencies:**
- `app/Models/Account/Subscription.php` — plan model with features/limits
- `app/Http/Middleware/SubscriptionMiddleware.php` — feature access control
- `app/Http/Controllers/ApiKeyController.php` — API key CRUD
- `src/modules/billing/composables/useFeatures.js` — frontend feature gating
- `src/components/layout/TopHeaderBar.vue` — navigation rendering
- `src/modules/setting/store/states/plan.js` — plan state management
- `src/modules/account/views/onboarding/` — onboarding flow components

---

## 12. Appendix

- **Original Proposal Document:** "ContentStudio API-Centric Plan — Product Proposal & Complete Flow Documentation" (March 2026)
- **Competitor Research:** See `docs/features/api-centric-plan/01-research.md` — detailed analysis of Late/Zernio, Ayrshare, and 9 other competitors
- **Workflow Design:** See `docs/features/api-centric-plan/02-workflow.md` — 6 user flows, 3 design decisions, edge cases, V1/V2 scope
- **Codebase Analysis:** See Research doc Section "Codebase Analysis" — critical file paths, integration points, reusable components

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-03-23 | Product Team (via Pipeline) | Initial draft based on API-Centric Plan proposal document, competitor research, and codebase analysis |
