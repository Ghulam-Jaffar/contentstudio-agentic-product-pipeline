# Research - Dummy Workspace Exploration

Date: April 14, 2026

Note: This research is codebase-only per request. Competitor research is intentionally skipped.

## What We Are Planning

After signup and onboarding, each user should have:

1. Their normal real workspace
2. A second private dummy workspace, visible only to that user

The dummy workspace should let the user explore the full suite with realistic seeded data:

- connected social accounts
- planned, approved, and in-review posts
- AI chat history and generated content
- inbox conversations, comments, and assignments
- analytics charts
- team members
- supporting data such as folders, labels, categories, and media

The seeded content must be date-relative, not static. The calendar should feel like an actual marketing team planned it:

- last 1 month
- current month
- next 2 months

## Executive Take

This is not just "create one more workspace".

It is a cross-product provisioning system. The right architecture is:

1. Create the normal workspace as today
2. Create a second workspace flagged as dummy/demo
3. Run a dedicated provisioning job that seeds each product area from templates
4. Use relative dates anchored to the user/workspace timezone at provisioning time
5. Keep dummy data isolated from real billing, sync, automation, and external platform side effects

The main constraint is analytics. Publishing, AI, media, and inbox can be seeded directly into existing workspace-scoped collections. Analytics is more complex because the product reads many platform-specific endpoints and account-level sections. For analytics, a direct fake-data response layer or precomputed synthetic dataset keyed to dummy workspace/account IDs is likely safer than trying to emulate the real ingestion pipeline.

## Key Findings

### 1. Signup currently creates exactly one default workspace

The current creation path is:

- `contentstudio-backend/app/Repository/Account/UsersRepository.php`
- `contentstudio-backend/app/Models/Settings/Workspace.php`
- `contentstudio-backend/app/Http/Controllers/Settings/WorkspaceController.php`

Relevant behavior:

- `UsersRepository::registerAccount()` creates the user
- after save, it builds the default workspace payload with `Workspace::getDefaultWorkspacePayload()`
- then it calls `WorkspaceController::createWorkspace()`

This is the primary insertion point for auto-creating the dummy workspace.

Important constraint:

- `WorkspaceController::createWorkspace()` checks workspace limits before creating a workspace
- if dummy workspaces are created through the same path without a bypass, they may count against plan limits immediately

Implication:

- dummy workspace creation must either bypass standard workspace-limit checks, or
- dummy workspaces must be excluded from workspace-count/billing logic

### 2. There is already old onboarding-copy infrastructure we can learn from

Relevant files:

- `contentstudio-backend/app/Http/Controllers/Integrations/OnboardingController.php`
- `contentstudio-backend/app/Jobs/Integrations/OnboardingPlansJob.php`
- `contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php`
- `contentstudio-backend/app/Repository/Publish/Planner/PostingRepository.php`
- `contentstudio-backend/app/Repository/Publish/Planner/BlogPostRepository.php`
- `contentstudio-backend/app/Repository/Publish/Automation/RssAutomationRepository.php`
- `contentstudio-backend/app/Repository/Publish/Automation/EvergreenAutomationRepository.php`
- `contentstudio-backend/app/Repository/Settings/WorkspaceTeamRepo.php`
- `contentstudio-backend/app/Repository/Settings/UtmRepo.php`
- `contentstudio-backend/app/Repository/Settings/HashtagRepo.php`
- `contentstudio-backend/app/Repository/Publish/Composer/FolderRepo.php`

What this system does:

- copies accounts, folders, labels, UTMs, hashtags, automations, blog posts, plans, and postings
- marks copied records with `migration = true`
- stores `generated_from_id` on copied items

This is useful because:

- it proves the product already has a precedent for cross-workspace seeded data
- it shows how copied planner data is linked and rewritten per workspace

But it is not the right core for this feature:

- it is built for migration/import flows, not for evergreen product demos
- it assumes copying from existing real onboarding accounts/workspaces
- it does not solve dynamic dates or realistic full-suite demo scenarios

Recommendation:

- reuse patterns, not the whole flow
- build a new dedicated dummy-workspace provisioning service/job

### 3. The frontend already handles multiple workspaces cleanly

Relevant files:

- `contentstudio-frontend/src/composables/useWorkspaceCore.js`
- `contentstudio-frontend/src/composables/useWorkspaceSwitcher.js`
- `contentstudio-frontend/src/stores/setting/useWorkspaceStore.ts`
- `contentstudio-frontend/src/Home.vue`
- `contentstudio-frontend/src/views/DashboardNew.vue`
- `contentstudio-frontend/src/router.js`

What matters:

- workspaces are already fetched as a list, then one is loaded as the active workspace
- the UI is heavily driven by `activeWorkspace._id`
- switching workspaces triggers fresh data loading across modules
- the standard dashboard and API-centric dashboard both already depend on active workspace state

Implication:

- if the dummy workspace is a normal workspace with valid workspace-scoped data, most of the frontend should "just work"
- the main product question is not whether the frontend can support this, but how we want to expose/select the dummy workspace after onboarding

### 4. Planner/publishing already supports realistic seeded calendars

Relevant files:

- `contentstudio-backend/app/Models/Publish/Planner/Plans.php`
- `contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php`
- `contentstudio-backend/app/Repository/Publish/Planner/PostingRepository.php`
- `contentstudio-frontend/src/stores/planner/usePlannerStore.ts`

Why this is promising:

- `Plans` already supports status, approval, labels, members, account selection, media details, platform-specific sharing details, first comments, and more
- there is already a separate `Posting` layer for downstream posting state
- existing onboarding jobs already know how to create copied plans and postings

Implication:

- the planner is one of the safest areas to seed directly
- the "proper social media calendar" requirement is achievable with scenario templates plus relative date generation

### 5. AI data can be seeded directly into existing workspace-scoped stores

Relevant files:

- `contentstudio-backend/app/Repository/AiChat/AiChatRepo.php`
- `contentstudio-backend/app/Repository/AiChat/AiChatMessagesRepo.php`
- `contentstudio-backend/app/Models/Ai/AiContentHistory.php`
- `contentstudio-backend/app/Repository/Ai/AiContentLibrary/AiContentLibraryProfileRepo.php`
- `contentstudio-backend/app/Repository/Ai/AiContentLibrary/AiContentLibraryPostRepo.php`
- `contentstudio-backend/app/Repository/Storage/MediaLibraryFoldersRepo.php`

What exists already:

- AI chats are workspace-scoped
- AI chat messages are workspace-scoped and can include post/video job metadata
- AI content history is workspace-scoped
- AI content library profiles and posts are workspace-scoped
- media library already supports dedicated AI folders such as `My AI creations` and `AI Video Clips`

Implication:

- seeded AI chats, generated captions, images, videos, and content-library items are straightforward compared to analytics

### 6. Inbox data is also seedable, but it lives in a separate service

Relevant files:

- `social-inbox-manager/app/database/mongo/repository/inbox_details_repository.py`
- `social-inbox-manager/app/database/mongo/repository/inbox_comments_repository.py`
- `social-inbox-manager/app/database/mongo/repository/inbox_messages_repository.py`
- `contentstudio-frontend/src/modules/inbox-revamp/composables/useInbox.js`

What exists:

- inbox uses Mongo collections `inbox_details`, `inbox_comments`, and `inbox_messages`
- records are explicitly keyed by `workspace_id`
- comments and messages already preserve assignment / marked-done / archived state

Implication:

- realistic inbox views can be seeded directly if we provision those collections
- we must ensure dummy inbox data never triggers real sync workers or external notifications

### 7. Analytics is the hardest domain

Relevant files:

- `contentstudio-frontend/src/api/analytics.ts`
- `contentstudio-frontend/src/modules/analytics/queries/keys.ts`
- `contentstudio-social-analytics-go/src/api/analytics/router.go`

What the codebase shows:

- frontend analytics is split into many section-level requests
- most sections are both workspace-scoped and account-scoped
- the Go service exposes a large number of platform-specific routes and overview variants

Implication:

- analytics is not a single summary blob we can fake once
- trying to seed the full real analytics pipeline would be expensive and brittle for a demo workspace feature

Recommendation:

- use a dedicated dummy analytics strategy for demo workspaces
- preferred v1 direction: serve precomputed synthetic analytics for dummy workspace/account IDs at the analytics response layer
- avoid building fake ingestion/fetcher/sync jobs just to make demo charts appear

### 8. There is already a trial cleanup precedent

Relevant file:

- `contentstudio-backend/app/Console/Commands/CleanupTrialInboxData.php`

What it shows:

- the product already has logic for cleaning trial-related inbox data after a retention window

Implication:

- lifecycle handling for dummy workspaces is not a foreign concept
- we should plan cleanup/regeneration rules up front instead of leaving dummy data permanent and unmanaged

## Reusable Pieces

The best reusable pieces for this feature are:

- normal workspace creation flow
- workspace switching and active-workspace hydration
- planner models and posting relationships
- AI chat/content-library repositories
- inbox Mongo collections and workspace scoping
- old onboarding migration patterns for copied data relationships

The least reusable piece is analytics, because its current shape is too distributed and platform-specific.

## Recommended Architecture

### Core approach

Build a dedicated `DummyWorkspaceProvisioningService` plus queued job.

Recommended responsibilities:

1. Create dummy workspace metadata
2. Mark workspace as dummy/demo
3. Seed foundational supporting records
4. Seed realistic scenario data per product domain
5. Anchor all seeded timestamps to the workspace timezone and current date
6. Store seed version + provisioning timestamp for future regeneration

### Data model additions

Add explicit workspace-level metadata, for example:

- `is_dummy_workspace`
- `dummy_workspace_version`
- `dummy_seeded_at`
- `dummy_anchor_date`
- `dummy_owner_user_id`

This matters because many downstream behaviors should branch on "dummy workspace" status.

### Seeding strategy

Use scenario templates plus relative date offsets, not copied static fixtures.

Example template behavior:

- scheduled posts spread through current month and next 2 months
- recently published posts in last month
- some posts in review with approvals
- some rejected posts with comments
- AI chat history tied to seeded assets
- inbox conversations with a mix of unread, assigned, pending, archived, replied
- team roles that feel realistic for a marketing org

### Domain-by-domain recommendation

#### Real direct seeding

Good fit for v1:

- workspace team
- content categories
- labels
- folders
- media library
- planner plans
- planner postings
- planner comments / approvals
- AI chats
- AI messages
- AI content history
- AI content library
- inbox details
- inbox comments
- inbox messages

#### Special handling

Needs separate treatment:

- analytics
- connected accounts if current observers/webhooks assume real external credentials

For accounts, the safest design is likely:

- create explicit dummy account records or a dedicated `is_dummy`/`is_demo` account mode
- ensure dummy accounts never enter real sync, observer, webhook, or billing flows

### Guardrails required

- dummy workspace must not count against user workspace limits
- dummy accounts must not trigger real syncs or platform jobs
- dummy automations must not actually execute
- dummy data must be excluded from billing/usage logic where needed
- dummy workspace should be clearly identifiable internally for support/debugging

## Main Product Decisions To Resolve In Workflow Step

These are the important decisions still open:

1. Should the user land in the real workspace after onboarding, the dummy workspace, or a chooser?
2. Should the dummy workspace be re-creatable/resettable by the user?
3. Should dummy data be editable freely, or partly locked to preserve demo quality?
4. Should dummy accounts appear visually distinct from real connected accounts?
5. Should analytics be fully synthetic in v1, while other modules are truly seeded?

## Recommended V1 Scope

Recommend v1 includes:

- auto-create one dummy workspace per signup
- realistic planner/calendar data with dynamic dates
- seeded supporting data: categories, labels, folders, team members, media
- seeded AI chat + AI content-library data
- seeded inbox conversations/comments/messages
- synthetic analytics responses for dummy workspaces
- clear UI treatment so users understand this is an exploration workspace

Recommend v1 excludes:

- copying dummy content into real workspaces
- multi-theme dummy workspace variants
- user-controlled scenario generation
- regeneration from the frontend unless it is a simple reset action

## Recommendation

Proceed, but treat this as a platform feature, not a content fixture script.

The clean implementation path is:

1. Introduce a first-class dummy workspace concept
2. Provision it through a dedicated backend service/job
3. Seed most domains directly into existing workspace-scoped stores
4. Handle analytics separately with synthetic response data for dummy workspaces
5. Decide the post-onboarding entry UX in the next step

If we try to do this as a loose collection of seed scripts without workspace-level dummy semantics, it will become fragile quickly.
