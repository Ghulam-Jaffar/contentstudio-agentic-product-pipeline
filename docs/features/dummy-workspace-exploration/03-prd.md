# PRD: Dummy Workspace Exploration

**Author:** Product Team
**Last Updated:** 2026-04-14
**Status:** In Review
**Target Release:** Q2 2026

---

## 1. Overview

Dummy Workspace Exploration gives each newly onboarded full-suite user a second private workspace populated with realistic sample data across Planner, AI, Inbox, Analytics, media, approvals, and team structure. The goal is to let users understand the full product immediately instead of landing in an empty workspace and having to imagine the value. The user still keeps their own real workspace as the primary place for live work; the sample workspace is a clearly labeled exploration environment with generated data and sandboxed behavior.

---

## 2. Problem Statement

**What problem are we solving?**

Today, a newly signed-up user completes onboarding and lands in a mostly empty workspace. The product is broad and cross-functional, but an empty workspace makes Planner, Inbox, Analytics, approvals, AI, and collaboration look incomplete until the user connects accounts, creates content, builds categories, and accumulates history. That creates a comprehension gap exactly when the user is deciding whether ContentStudio is worth adopting.

**Who has this problem?**

- New full-suite trial users
- New paid users creating their first workspace from scratch
- Agency and brand teams evaluating the breadth of the platform before real setup

The problem is most visible in the first session and first few days after signup, when users have the least context and the highest risk of dropping off.

**What happens if we don't solve it?**

- Users see empty states instead of product value
- Full-suite capabilities look harder to understand than they actually are
- Trial users are less likely to explore multiple modules
- Conversion and activation suffer because the product asks users to build context before they can experience it
- Support and sales spend more time explaining "what it would look like once set up"

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Improve first-week product exploration | % of newly onboarded full-suite users who visit 3+ core modules in first 7 days | +25% | Product analytics by signup cohort |
| Improve activation quality | % of newly onboarded users who create at least one real post or connect at least one real account in their own workspace within 7 days | +15% | Workspace activity events |
| Increase understanding of full-suite value | % of newly onboarded users who enter Planner, AI, Inbox, or Analytics at least once from the sample workspace | 50%+ | Sample-workspace usage events |
| Guard rail: avoid confusion | % of support tickets / feedback mentioning confusion between sample and real workspace | <5% of new-user onboarding tickets | Support tagging and feedback review |
| Guard rail: avoid unsafe actions | Real publish/sync actions triggered from sample workspace | 0 | Backend event logging and blocked-action tracking |

---

## 4. Target Users

**Primary Persona:**
New Full-Suite Trial User — A marketer, brand manager, small business owner, or agency operator who wants to understand what ContentStudio can do before investing time in a full setup. They are interested in the whole suite, not just a narrow API or automation surface.

**Secondary Persona:**
Evaluation-Phase Team Lead — A decision-maker or operator setting up ContentStudio for a team or client account who needs to understand calendar planning, approvals, inbox workflows, analytics, and AI-assisted creation quickly.

**Non-Users (explicitly out of scope):**
- Existing mature workspaces with real data
- API-centric users in v1
- Shared team demo workspaces across multiple real users
- Public/open demo environments

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | New full-suite user | get a ready-made sample workspace after onboarding | I can understand the product immediately | Must Have |
| US-2 | New user | keep my own workspace separate and empty | I can start real work without demo data polluting it | Must Have |
| US-3 | New user | choose whether to enter my own workspace or the sample workspace after onboarding | I stay in control of where I begin | Must Have |
| US-4 | New user | see realistic calendar, approvals, inbox threads, and analytics in the sample workspace | I can understand how the suite works in practice | Must Have |
| US-5 | New user | see AI chats, generated assets, and saved content examples | I can evaluate AI-assisted workflows without setting everything up first | Must Have |
| US-6 | New user | clearly know when I am in a demo environment | I do not confuse sample data with my own data | Must Have |
| US-7 | New user | switch back to my own workspace at any time | I can move from exploration to real setup smoothly | Must Have |
| US-8 | Product team | seed sample data relative to the current date | the calendar feels current and believable for every signup | Must Have |
| US-9 | Engineering team | prevent sample workspaces from triggering real publishing, sync, or billing effects | the feature remains safe and supportable | Must Have |
| US-10 | Support/sales | use the sample workspace as the default exploration surface for new users | users need less explanation and more self-serve discovery | Should Have |

---

## 6. Requirements

### 6.1 Must Have (P0)

**Workspace model and eligibility**
- Create one additional private sample workspace for each newly onboarded eligible full-suite user in v1
- Keep the user's normal workspace intact as the real workspace
- Add first-class workspace metadata to identify sample workspaces, ownership, seed version, and seed timestamp
- Ensure the sample workspace does not count against normal workspace limits, billing limits, or upgrade prompts

**Provisioning and data ownership**
- Trigger provisioning after signup onboarding is completed, not during raw registration
- Provision sample data through backend/services and queued jobs, not frontend fixtures
- Seed realistic records across:
  - planner/calendar
  - approvals/comments
  - media library
  - content categories / labels / folders / campaigns as needed
  - AI chats / AI-generated outputs / AI content-library artifacts
  - inbox conversations / comments / assignments
  - team members and roles
- Generate seeded dates relative to the workspace timezone and current date:
  - previous month
  - current month
  - next two months

**Analytics**
- Serve synthetic but coherent analytics for sample workspace/account IDs
- Ensure sample analytics align logically with seeded sample accounts, posts, and date ranges
- Disable or hide refresh/sync behaviors that imply live platform analytics

**Accounts and safe behavior**
- Sample accounts must be explicitly treated as non-live sample accounts
- Sample accounts must not trigger real syncs, webhooks, platform jobs, inbox fetches, or publishing flows
- External/live actions attempted from the sample workspace must be blocked, redirected to the real workspace, or sandboxed depending on the action

**Frontend experience**
- Add a post-onboarding start screen with two options:
  - `Your Workspace`
  - `Explore Sample Workspace`
- If sample provisioning is still running, show a loading/progress state on the sample option without blocking entry into the real workspace
- Surface the sample workspace as a normal workspace in the existing switcher, with a clear `Sample` or `Demo` badge
- Hide the normal workspace onboarding widget in the sample workspace
- Add `Explore Sample Workspace` CTAs to relevant empty states in the real workspace

**Visual treatment**
- Add a persistent top banner in the sample workspace across key product pages
- Banner must clearly communicate:
  - this is a demo/sample workspace
  - data is generated for exploration
  - live account / sync / publishing behavior is limited here
- Add a clear sample badge in workspace-switcher and workspace-identification contexts
- Use light contextual callouts on key surfaces where needed, but do not create a separate product skin for v1

**Observability and resilience**
- Track provisioning success/failure, duration, and retry state
- If provisioning fails, the user must still be able to continue with their real workspace
- Make sample-workspace readiness visible enough for the user to discover once available

### 6.2 Should Have (P1)

- Tailor parts of the seeded scenario to business type collected during onboarding
- Show a non-blocking in-app notification when the sample workspace finishes provisioning after the user already entered their real workspace
- Include a few realistic approval states such as in-review, approved, rejected-with-feedback, and missed scheduling context where applicable
- Provide a lightweight dashboard callout explaining what to explore first inside the sample workspace

### 6.3 Nice to Have (P2)

- Offer multiple sample scenario packs by industry or use case
- Allow a user to reset/regenerate the sample workspace from the UI
- Let users copy selected sample posts or assets into their real workspace
- Backfill sample workspaces for existing full-suite users

### 6.4 Explicitly Out of Scope

- A public/shared demo environment
- A full alternate theme/skin for sample workspaces
- Real publishing, real syncing, or live inbox transport from the sample workspace
- API-centric plan support in v1
- Multi-user collaboration in the same sample workspace
- Full regeneration tooling in v1

---

## 7. User Flow (High Level)

1. User signs up and completes the existing full-page onboarding flow
2. On onboarding completion, the system marks signup onboarding complete and starts sample-workspace provisioning in the background
3. User sees a post-onboarding start screen with:
   - `Your Workspace`
   - `Explore Sample Workspace`
4. User can enter their real workspace immediately, even if sample provisioning is still running
5. Once the user enters the sample workspace, they see a persistent demo banner, sample badge, and realistic seeded data across major modules
6. User explores Planner, AI, Inbox, Analytics, approvals, media, and collaboration flows using sample data
7. If the user attempts a live action from the sample workspace, the system blocks it, redirects it, or keeps it sandboxed
8. User can switch back to their real workspace any time and begin real setup there

Detailed flow and edge cases live in [02-workflow.md](/home/casper/code/contentstudio-agentic-product-pipeline/docs/features/dummy-workspace-exploration/02-workflow.md).

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Each eligible user gets at most one private sample workspace in v1 | Keeps ownership clear and avoids workspace sprawl |
| BR-2 | The sample workspace is created only after signup onboarding completion | Avoids provisioning abandoned signups and allows better personalization |
| BR-3 | The sample workspace must not count against workspace limits or billing logic | It is a product exploration surface, not a purchased workspace |
| BR-4 | Sample data must be created and served by backend/services, not frontend mock fixtures | Ensures real product flows and consistent data behavior |
| BR-5 | The seeded calendar must be date-relative to workspace timezone, covering last month, current month, and next two months | Keeps the demo believable for every signup date |
| BR-6 | Sample accounts must never trigger real syncs, real publishing, webhooks, or external platform effects | Safety and supportability |
| BR-7 | Analytics for sample workspaces are synthetic by design in v1 | Real analytics ingestion is too heavy and brittle for this use case |
| BR-8 | The sample workspace must always be visually labeled with banner and badge treatment | Reduces confusion between sample and real data |
| BR-9 | The real workspace remains the user's real working area and should stay clean of seeded sample data | Preserves trust and setup clarity |
| BR-10 | The sample workspace should allow safe internal interactions, but not full live behavior | Exploration quality without external side effects |
| BR-11 | The standard onboarding widget must be hidden in the sample workspace | Seeded data already demonstrates the product |
| BR-12 | Provisioning failure must not block user access to the real workspace | New-user access cannot depend on sample generation succeeding |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| Exact user eligibility in v1? | Full-suite trials only / full-suite trials + paid / all non-API users | Product | Before story creation sign-off | Pending |
| How personalized should the seeded scenario be? | Generic scenario / business-type-aware scenario / industry packs later | Product + Engineering | Before implementation | Pending |
| What is the lifecycle policy for sample workspaces? | Permanent / auto-cleanup after X days / regenerate on demand later | Product + Engineering | Before implementation | Pending |
| Which actions are blocked vs. sandboxed vs. redirected? | Per-module action matrix | Product + Engineering | Sprint 1 | Pending |
| How should readiness be surfaced if provisioning is slow? | Polling card / toast / notification center / all of these | Product + Frontend | Sprint 1 | Pending |
| Should users be able to rename or edit the sample workspace metadata? | No / partial / yes | Product | Before implementation | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Users confuse the sample workspace with their real workspace | Medium | High | Persistent banner, sample badges, clear chooser, clean real workspace, blocked live actions |
| Sample provisioning is slow and harms the onboarding finish experience | Medium | Medium | Async provisioning, non-blocking chooser, loading state, retryable job design |
| Seeded data feels fake or low quality | Medium | High | Use scenario templates with realistic statuses, campaigns, comments, AI history, and date-relative planning |
| Sample accounts accidentally enter real sync/publish flows | Low | High | First-class sample flags, backend guardrails, explicit exclusion from jobs/webhooks/observers |
| Analytics look inconsistent with other seeded data | Medium | Medium | Serve synthetic analytics from a controlled response layer aligned to seeded accounts and dates |
| Feature adds storage and maintenance overhead | Medium | Medium | Keep one sample workspace per user, version seeds, scope v1 tightly, defer reset tooling |
| Too much UI differentiation creates unnecessary design complexity | Low | Medium | Use banner + badges + callouts only; avoid a full alternate theme in v1 |

---

## 11. Dependencies

**Internal:**
- Backend team for workspace model updates, provisioning service/job, seeding logic, safe-action guardrails, and workspace-limit exclusions
- Frontend team for chooser flow, banner/badge treatment, empty-state CTAs, and sample-aware UI behavior
- Inbox service changes for seeded sample inbox records and safe handling
- Analytics service changes for synthetic sample analytics responses
- Product analytics / event tracking for activation and confusion guard rails

**External:**
- None required for initial product definition

**Blockers:**
- Eligibility decision for v1
- Action safety matrix by module
- Sample workspace lifecycle policy

---

## 12. Appendix

- Research: [01-research.md](/home/casper/code/contentstudio-agentic-product-pipeline/docs/features/dummy-workspace-exploration/01-research.md)
- Workflow: [02-workflow.md](/home/casper/code/contentstudio-agentic-product-pipeline/docs/features/dummy-workspace-exploration/02-workflow.md)
- Relevant signup onboarding flow: [useOnboardingFlow.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/modules/onboarding/composables/useOnboardingFlow.ts)
- Relevant workspace switching flow: [useWorkspaceSwitcher.js](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/composables/useWorkspaceSwitcher.js)
- Relevant workspace creation path: [UsersRepository.php](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-backend/app/Repository/Account/UsersRepository.php)

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-04-14 | Product Team | Initial PRD draft |
| 2026-04-14 | Product Team | Added explicit sample-workspace visual treatment and backend/frontend ownership split |
