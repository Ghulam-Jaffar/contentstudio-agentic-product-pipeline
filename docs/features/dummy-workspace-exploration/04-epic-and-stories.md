# Epic + Stories - Sample Workspace Exploration

## Epic

**Title:** Sample Workspace Exploration

**Description:**

Give each newly onboarded eligible full-suite user a second private sample workspace filled with realistic demo data so they can understand ContentStudio faster without polluting their own real workspace. The release covers sample-workspace metadata and eligibility, asynchronous provisioning, date-relative seeded scenarios across Planner, AI, Inbox, Analytics, media, approvals, and team structure, plus backend guardrails so sample workspaces never trigger live sync, publishing, or billing side effects.

On the frontend, the release adds a post-onboarding start screen, persistent demo-workspace identification through banner and badge treatment, sample-aware guidance across key modules, and clean handoff paths between the sample workspace and the user’s real workspace. V1 keeps the real workspace empty and trustworthy while making the sample workspace feel rich, current, and safe to explore.

---

## Local Source Docs

- Research: [01-research.md](/home/casper/code/contentstudio-agentic-product-pipeline/docs/features/dummy-workspace-exploration/01-research.md)
- Workflow: [02-workflow.md](/home/casper/code/contentstudio-agentic-product-pipeline/docs/features/dummy-workspace-exploration/02-workflow.md)
- PRD: [03-prd.md](/home/casper/code/contentstudio-agentic-product-pipeline/docs/features/dummy-workspace-exploration/03-prd.md)

---

## Stories

---

### Story 1: [BE] Add sample workspace metadata, ownership, and eligibility rules

**Description:**
As a newly onboarded full-suite user, I want ContentStudio to create a private sample workspace that stays separate from my real workspace and does not count against my workspace limits so that I can explore the product safely without affecting my actual setup.

**Workflow:**
1. User signs up and completes onboarding for an eligible full-suite account.
2. ContentStudio keeps the user’s real workspace as their real working area.
3. ContentStudio determines whether the user is eligible for a private sample workspace.
4. If eligible, ContentStudio creates one private sample workspace owned by that user only.
5. User later sees both workspaces in the app, but the sample workspace does not consume the normal workspace allowance or trigger upgrade pressure.

**Acceptance criteria:**
- [ ] Workspace records support first-class sample-workspace metadata, including whether the workspace is a sample workspace, who owns it, seed status, seed version, seeded timestamp, and anchor date
- [ ] Eligibility rules exist for the v1 cohort and prevent sample workspaces from being created for ineligible users
- [ ] A user can have at most one private sample workspace in v1
- [ ] Sample workspaces are excluded from normal workspace-limit checks, billing counts, and upgrade prompts
- [ ] Sample workspaces are only returned to the owning user and are not exposed to other users by default
- [ ] Workspace-fetch responses include the sample-workspace metadata needed by the web app for chooser, banner, and badge behavior
- [ ] Sample workspaces support a provisioning state model that covers at least pending, ready, and failed
- [ ] Creation and fetch logic remain backward-compatible for normal workspaces

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- Adds sample-workspace metadata to workspace records
- Existing normal workspaces remain unchanged in behavior

**Impact on other products:**
- Web app will use this metadata for sample-workspace UX
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- None

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 2: [BE] Build sample workspace provisioning and date-relative scenario generation

**Description:**
As a newly onboarded user, I want my sample workspace to be prepared automatically after onboarding so that I can start exploring rich product data without waiting through a blocking setup flow.

**Workflow:**
1. User finishes signup onboarding.
2. ContentStudio marks onboarding complete and starts preparing the sample workspace in the background.
3. User can continue into their real workspace immediately while the sample workspace is being prepared.
4. Once preparation finishes, the sample workspace becomes available for exploration with data anchored to the current date and workspace timezone.

**Acceptance criteria:**
- [ ] Onboarding completion triggers an asynchronous provisioning job for eligible users instead of blocking the post-onboarding experience
- [ ] Provisioning is idempotent so duplicate onboarding completion events do not create duplicate sample workspaces or duplicate sample data
- [ ] Provisioning stores and updates sample-workspace readiness state so the frontend can show pending, ready, and failed states
- [ ] Scenario generation uses the workspace timezone and current date as the anchor for all generated timelines
- [ ] The scenario engine supports at least the following relative windows: previous month, current month, next month, and month after next
- [ ] Provisioning can record failure reason and retry safely without corrupting existing sample data
- [ ] Provisioning duration and outcome are logged for observability
- [ ] Provisioning can accept onboarding context such as business type for future scenario tailoring, even if v1 keeps most scenarios generic

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- Adds provisioning-state data for sample workspaces
- No changes to normal onboarding data beyond triggering provisioning for eligible users

**Impact on other products:**
- Enables frontend readiness messaging after onboarding
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 3: [BE] Seed sample accounts, planner calendar, approvals, team, media, and supporting records

**Description:**
As a user exploring the sample workspace, I want to see a realistic publishing calendar, connected-looking accounts, approvals, team members, media, and supporting setup records so that the product feels like an actual marketing operation instead of a demo stub.

**Workflow:**
1. User opens the sample workspace for the first time.
2. User sees sample social accounts already present across relevant channels.
3. User opens Planner and sees a believable spread of past, current, and upcoming content.
4. User opens posts and sees approvals, comments, media, categories, labels, campaigns, and team ownership already in place.
5. User understands how the planning and collaboration side of ContentStudio works without first building everything manually.

**Acceptance criteria:**
- [ ] Sample workspaces are seeded with non-live sample account records that look valid in the product but cannot be used as real connected accounts
- [ ] Planner data spans the previous month, current month, and next two months using the generated anchor date
- [ ] Seeded planner data includes a realistic mix of published posts, scheduled posts, draft posts, in-review posts, and rejected posts with feedback
- [ ] Sample approval records include approvers, comments, and status states that are visible in normal planner/post views
- [ ] Sample team members and roles are created so collaboration and approvals feel populated
- [ ] Supporting records are seeded where needed, including categories, labels, folders, campaigns, and media assets
- [ ] Seeded records are scoped only to the sample workspace and do not pollute the user’s real workspace
- [ ] No real platform credentials, tokens, or external account references are stored for sample accounts
- [ ] Re-running the seed process does not create uncontrolled duplicate records for the same sample workspace version

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- Creates sample workspace-scoped records across planner, approvals, media, team, and supporting collections
- Normal user workspaces and live account records remain unchanged

**Impact on other products:**
- Makes existing planner, approvals, media, and collaboration UI immediately useful in the sample workspace
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**
- Depends on: **[BE] Build sample workspace provisioning and date-relative scenario generation**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 4: [BE] Seed AI chats, generated assets, and content-library scenarios for sample workspaces

**Description:**
As a user evaluating ContentStudio’s AI capabilities, I want the sample workspace to include realistic AI chats, generated captions, images, videos, and saved content so that I can understand the AI workflow without setting up brand data and prompts from scratch.

**Workflow:**
1. User opens AI surfaces in the sample workspace.
2. User sees existing chat threads and generated outputs that look like real prior work.
3. User browses saved captions, images, videos, and post drafts tied to the sample workspace.
4. User understands how AI creation and content reuse work before they begin creating real content in their own workspace.

**Acceptance criteria:**
- [ ] Sample workspaces include seeded AI chat threads with realistic prompts and responses across common use cases
- [ ] Seeded AI outputs include at least captions, post variants, images, videos, and saved content artifacts where the current product supports them
- [ ] AI chat messages and saved outputs are scoped only to the sample workspace
- [ ] AI-generated sample assets appear in the normal media/content-library surfaces for the sample workspace
- [ ] Seeded AI examples reflect a believable marketing context rather than placeholder lorem ipsum
- [ ] Existing AI actions performed by the user inside the sample workspace continue saving back into the sample workspace only
- [ ] Seeded AI data does not appear in the user’s real workspace

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- Creates sample workspace-scoped AI chats, AI outputs, and content-library records
- Does not modify existing real AI history or content-library records

**Impact on other products:**
- Makes existing AI web flows immediately understandable in the sample workspace
- AI is web-only; no mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**
- Depends on: **[BE] Build sample workspace provisioning and date-relative scenario generation**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 5: [BE] Seed sample inbox conversations and isolate sample inbox behavior

**Description:**
As a user exploring Inbox, I want to see realistic conversations, comments, assignments, and done/archived states in the sample workspace so that I can understand inbox workflows without needing real social activity.

**Workflow:**
1. User opens Inbox inside the sample workspace.
2. User sees conversations and comments already populated across sample accounts.
3. User can inspect unread, assigned, pending, archived, and done states.
4. User understands how triage and collaboration work in Inbox without requiring live platform traffic.

**Acceptance criteria:**
- [ ] Sample workspaces are seeded with sample inbox details, messages, comments, assignments, and status mixes that appear correctly in Inbox
- [ ] Inbox sample data is scoped to the sample workspace only
- [ ] Sample inbox records include a realistic distribution of unread, assigned, pending, archived, and done items
- [ ] Sample inbox data does not trigger any live sync, fetch, reply, or notification behavior against external platforms
- [ ] Sample inbox records can be safely viewed and manipulated in the product without affecting live inbox data
- [ ] Seeding logic is versionable and idempotent for repeated provisioning/retry scenarios

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- Creates sample workspace-scoped inbox records
- Does not modify or mix with existing live inbox records

**Impact on other products:**
- Makes the current web Inbox experience immediately explorable for new users
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**
- Depends on: **[BE] Build sample workspace provisioning and date-relative scenario generation**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 6: [BE] Serve synthetic analytics for sample workspaces and sample accounts

**Description:**
As a user exploring Analytics, I want the sample workspace to show believable performance charts and account-level metrics so that I can understand what the analytics module looks like without requiring live data ingestion.

**Workflow:**
1. User opens Analytics in the sample workspace.
2. User sees overview and account-level metrics that match the sample accounts and recent content timeline.
3. User changes date ranges and still gets coherent sample analytics.
4. User understands the analytics experience without expecting that the data is live.

**Acceptance criteria:**
- [ ] Analytics responses for sample workspace and sample account IDs are served from a synthetic sample-data path rather than the normal live-ingestion path
- [ ] Sample analytics cover the date ranges needed by the current analytics UI and return coherent responses for common filters and overview requests
- [ ] Sample analytics align logically with the sample accounts and seeded timeline used elsewhere in the sample workspace
- [ ] Synthetic analytics responses are clearly distinguishable internally so they can be debugged and supported
- [ ] Live analytics refresh or sync paths are not invoked for sample workspaces
- [ ] Normal analytics behavior for real workspaces remains unchanged

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- No live analytics ingestion changes for normal workspaces
- Adds synthetic analytics response handling for sample workspaces

**Impact on other products:**
- Makes the current analytics web experience explorable for new users
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**
- Depends on: **[BE] Build sample workspace provisioning and date-relative scenario generation**
- Depends on: **[BE] Seed sample accounts, planner calendar, approvals, team, media, and supporting records**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 7: [BE] Block live actions from sample workspaces and track blocked attempts

**Description:**
As a user exploring the sample workspace, I want the system to stop me from connecting real accounts, syncing live data, or publishing real content from the demo environment so that I can explore safely without accidental external side effects.

**Workflow:**
1. User explores the sample workspace normally.
2. User tries an action that would affect live systems, such as connecting a real account, syncing live data, or publishing to an external platform.
3. ContentStudio blocks the live action at the backend level, even if the frontend path is hit directly.
4. The app can then show a clear handoff back to the user’s real workspace for live work.

**Acceptance criteria:**
- [ ] Live-action guardrails exist for sample workspaces at the backend level and do not rely only on frontend hiding
- [ ] Real social account connection and credential-save paths are blocked for sample workspaces
- [ ] Real publishing, posting, or scheduling paths that would hit external platforms are blocked for sample workspaces
- [ ] Live inbox reply/sync and analytics refresh/sync paths are blocked for sample workspaces
- [ ] Sample workspaces and sample accounts are excluded from background jobs, webhooks, automations, and other live-processing paths that should only run for real workspaces
- [ ] Blocked attempts are logged with enough context for support and product review
- [ ] Blocking behavior does not affect equivalent actions in the user’s real workspace

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- Prevents sample workspace records from entering live external workflows
- Adds blocked-action telemetry/logging for sample-workspace attempts

**Impact on other products:**
- Supports safe web-app exploration flows
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 8: [FE] Add the post-onboarding chooser and sample-workspace readiness states

**Description:**
As a newly onboarded user, I want to choose between my own workspace and a prepared sample workspace right after onboarding so that I can either start real setup immediately or explore the full product with realistic data first.

**Workflow:**
1. User completes the final onboarding step.
2. Instead of going directly to Home, user lands on a new start screen with two choices.
3. User can click **Open My Workspace** to continue with their real workspace immediately.
4. User can click **Explore Sample Workspace** if the sample workspace is ready.
5. If the sample workspace is still being prepared, user sees progress and can continue into their real workspace while waiting.
6. If preparation later finishes, user sees a non-blocking success notification and can enter the sample workspace from the chooser or workspace switcher.

**UI copy and components:**
- Use `Button`, `Badge`, `Progress`, `Alert`, `Loader`, and `Icon` from `@contentstudio/ui`
- Page eyebrow: `You're all set`
- Page title: `Where would you like to start?`
- Page description: `Your own workspace is ready for real work. You also have a private sample workspace filled with demo content so you can explore ContentStudio faster.`
- Real workspace card badge: `Your Workspace`
- Real workspace card title: `Start with your own workspace`
- Real workspace card body: `Connect your real accounts, create your own content, and set up your team when you're ready.`
- Real workspace primary CTA: `Open My Workspace`
- Sample workspace card badge: `Sample Workspace`
- Sample workspace card title when ready: `Explore a fully set-up example`
- Sample workspace card body when ready: `Open a private demo workspace with planned posts, inbox activity, AI content, approvals, and analytics.`
- Sample workspace primary CTA when ready: `Explore Sample Workspace`
- Sample workspace loading title: `Preparing your sample workspace`
- Sample workspace loading body: `We're generating realistic demo data for you. This usually takes less than a minute.`
- Sample workspace loading CTA: `Preparing...`
- Failure alert title: `We couldn't finish your sample workspace yet`
- Failure alert body: `You can continue with your own workspace now and try the sample workspace again later.`
- Failure primary CTA: `Open My Workspace`
- Failure secondary CTA: `Try Again`
- Ready toast: `Your sample workspace is ready to explore`

**Acceptance criteria:**
- [ ] Eligible users land on the chooser screen after onboarding instead of being sent directly to Home
- [ ] The chooser shows the exact page title, description, card labels, and CTA copy defined in this story
- [ ] Clicking `Open My Workspace` routes the user into their real workspace
- [ ] Clicking `Explore Sample Workspace` routes the user into the sample workspace when provisioning status is ready
- [ ] When provisioning status is pending, the sample card shows `Preparing your sample workspace`, the explanatory copy, and a visible `Progress`/`Loader` state instead of a working CTA
- [ ] When provisioning later finishes, the user sees the toast `Your sample workspace is ready to explore`
- [ ] When provisioning fails, the chooser shows the exact failure alert title/body and the CTAs `Open My Workspace` and `Try Again`
- [ ] The chooser works on desktop and mobile web layouts
- [ ] The chooser fires analytics events for chooser viewed, real workspace selected, and sample workspace selected

**Mock-ups:**
- New onboarding-completion chooser screen required
- Use existing ContentStudio visual language; no new component required

**Impact on existing data:**
- No new user content data
- Consumes sample-workspace readiness metadata from backend

**Impact on other products:**
- Web-only onboarding change
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**
- Depends on: **[BE] Build sample workspace provisioning and date-relative scenario generation**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 9: [FE] Add persistent demo-workspace banner, badges, and sample-aware visual treatment

**Description:**
As a user inside the sample workspace, I want the app to clearly show that I am in a demo environment so that I do not confuse generated sample data with my own real workspace data.

**Workflow:**
1. User enters the sample workspace from the chooser or workspace switcher.
2. User sees a persistent top banner telling them they are in a demo workspace.
3. User sees a `Sample` badge anywhere the workspace is identified in the workspace switcher and header contexts.
4. User sees lightweight sample-aware guidance on key sample surfaces such as Dashboard, Planner, AI, Inbox, and Analytics.
5. User can switch back to their real workspace at any time from the banner CTA.

**UI copy and components:**
- Use `Alert`, `Badge`, `Button`, `Icon`, and `ActionIcon` from `@contentstudio/ui`
- Persistent banner badge: `Sample Workspace`
- Persistent banner title: `Demo workspace for exploration`
- Persistent banner body line 1: `This workspace contains generated sample posts, chats, inbox threads, approvals, and analytics so you can explore ContentStudio safely.`
- Persistent banner body line 2: `Live account connections, syncing, and publishing are limited here.`
- Persistent banner primary CTA: `Switch to Your Workspace`
- Workspace switcher badge label: `Sample`
- Inline sample callout title on key surfaces: `You're viewing sample data`
- Inline sample callout body: `Use this workspace to explore how ContentStudio looks when it is fully set up. Switch to your own workspace when you're ready to work with real accounts and content.`
- Inline sample callout CTA: `Switch to Your Workspace`

**Acceptance criteria:**
- [ ] Every sample workspace page shows the persistent top banner with the exact badge, title, body copy, and CTA defined in this story
- [ ] The sample workspace is labeled with a `Sample` `Badge` in the workspace switcher
- [ ] The sample banner is persistent and does not disappear permanently after page navigation
- [ ] Dashboard, Planner, AI, Inbox, and Analytics each show the inline sample callout with the exact title, body, and CTA defined in this story
- [ ] Clicking `Switch to Your Workspace` from the banner or inline callout switches the user to their real workspace
- [ ] Normal workspaces never show the sample banner, sample badge, or inline sample callout
- [ ] The banner and callouts use theme-aware components and styling that work on default and white-label domains

**Mock-ups:**
- New banner treatment required
- Reuse existing layout patterns; no alternate app theme/skin is needed

**Impact on existing data:**
- No new content data
- Consumes sample-workspace metadata from backend

**Impact on other products:**
- Web-only UI treatment in v1
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Add sample workspace metadata, ownership, and eligibility rules**
- Depends on: **[FE] Add the post-onboarding chooser and sample-workspace readiness states**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 10: [FE] Add sample-workspace handoff UX for empty states and blocked live actions

**Description:**
As a user moving between the sample workspace and my real workspace, I want clear prompts that tell me when to explore the sample workspace and when to switch back to my own workspace for live actions so that I always know where to go next.

**Workflow:**
1. User enters their real workspace and sees normal empty states.
2. Empty states now include a clear path to explore the sample workspace when the user wants to see a fully populated example.
3. User later enters the sample workspace and tries a live action such as connecting a real account, publishing real content, or refreshing live data.
4. Instead of confusion or silent failure, user sees a clear handoff message telling them to switch to their own workspace for live actions.
5. User can switch immediately or stay in the sample workspace and continue exploring.

**UI copy and components:**
- Use `Alert`, `Modal`, `Button`, `Badge`, and `Icon` from `@contentstudio/ui`
- Real workspace dashboard discovery callout title: `Want to see a fully set-up example?`
- Real workspace dashboard discovery callout body: `Explore your private sample workspace to preview planned content, inbox activity, AI outputs, approvals, and analytics before setting up your own workspace.`
- Real workspace dashboard CTA: `Explore Sample Workspace`
- Real workspace planner secondary empty-state CTA: `See a sample calendar`
- Real workspace inbox secondary empty-state CTA: `See a sample inbox`
- Real workspace analytics secondary empty-state CTA: `See sample analytics`
- Blocked-action modal title: `Use your real workspace for live actions`
- Blocked-action modal description: `This demo workspace is for exploration only. To connect real social accounts, sync live data, or publish real content, switch to your own workspace.`
- Blocked-action primary CTA: `Switch to Your Workspace`
- Blocked-action secondary CTA: `Stay in Demo Workspace`
- Blocked-action helper line for account connection: `Real social accounts can only be added in your own workspace.`
- Blocked-action helper line for publishing: `Real publishing is turned off in the demo workspace.`
- Blocked-action helper line for sync or refresh: `Live syncing is turned off in the demo workspace.`
- Workspace-switch failure toast: `We couldn't switch workspaces right now. Please try again.`

**Acceptance criteria:**
- [ ] The real workspace dashboard shows the exact discovery callout title, body, and CTA defined in this story when a sample workspace is available
- [ ] Real workspace Planner, Inbox, and Analytics empty states each expose the exact secondary CTA labels defined in this story
- [ ] Clicking any of these discovery CTAs switches the user into the sample workspace
- [ ] When a user attempts a blocked live action from the sample workspace, a `Modal` opens with the exact title, description, CTA labels, and contextual helper line defined in this story
- [ ] The same modal pattern is used for at least account-connection, publishing, and sync/refresh handoff cases
- [ ] Clicking `Switch to Your Workspace` from the modal switches the user into their real workspace
- [ ] Clicking `Stay in Demo Workspace` closes the modal and keeps the user in the sample workspace
- [ ] If workspace switching fails, the user sees the toast `We couldn't switch workspaces right now. Please try again.`
- [ ] Normal real-workspace flows are not blocked by this handoff UX

**Mock-ups:**
- Reuse existing empty-state layouts and add secondary CTA treatment
- New blocked-action modal required

**Impact on existing data:**
- No new content data
- Consumes sample-workspace availability and blocked-action context from backend

**Impact on other products:**
- Web-only UI behavior in v1
- No mobile app changes in v1
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Block live actions from sample workspaces and track blocked attempts**
- Depends on: **[FE] Add the post-onboarding chooser and sample-workspace readiness states**
- Depends on: **[FE] Add persistent demo-workspace banner, badges, and sample-aware visual treatment**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
