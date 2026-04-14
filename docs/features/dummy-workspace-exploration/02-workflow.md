# Workflow Design: Dummy Workspace Exploration

**Date:** 2026-04-14

---

## 1. Feature Placement

### Navigation & Entry Points

| Entry Point | Location | Purpose |
|---|---|---|
| **Provision trigger** | Signup onboarding completion flow | Create the user's private sample workspace after onboarding is finished |
| **Post-onboarding start screen** | New screen shown immediately after onboarding completion | Let the user choose between their real empty workspace and the sample workspace |
| **Workspace switcher** | Existing workspace switcher | Persistent way to move between real and sample workspaces |
| **Real workspace empty states** | Dashboard, Planner, Inbox, Analytics entry surfaces | Remind users they can explore the product with the sample workspace |
| **Sample workspace banner** | Global top banner in sample workspace | Keep the user aware they are in a sample environment |

### Recommended Placement Rules

- The sample workspace should be created only for new full-suite users in v1.
- It should be provisioned after the user completes signup onboarding, not during raw account registration.
- The real workspace remains the user's primary workspace.
- The sample workspace appears as a normal workspace in the switcher, but with a clear `Sample` badge.
- The sample workspace should suppress the normal in-app onboarding widget and instead show a dedicated sample banner.

### Visual Treatment Recommendation

- The sample workspace should feel clearly different, but not like a separate product skin.
- Recommended v1 treatment:
  - persistent top banner across product pages
  - `Sample` or `Demo` badge in workspace switcher and workspace header contexts
  - lightweight informational callout style on key entry surfaces such as Dashboard, Planner, Inbox, AI, and Analytics
- Recommended banner content:
  - this is a demo workspace for exploring ContentStudio
  - data is generated sample data
  - live account, sync, and publishing behavior is limited here
  - primary CTA: `Switch to Your Workspace`
- Do not create a full alternative theme for v1. A clear banner-plus-badge system is enough to reduce confusion without adding unnecessary design and implementation weight.

### Relevant Existing Integration Points

- Signup onboarding flow:
  - [onboardingFlow.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/modules/onboarding/config/routes/onboardingFlow.ts)
  - [flow.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/modules/onboarding/config/flow.ts)
  - [useOnboardingFlow.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/modules/onboarding/composables/useOnboardingFlow.ts)
- Workspace loading and switching:
  - [useWorkspaceCore.js](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/composables/useWorkspaceCore.js)
  - [useWorkspaceSwitcher.js](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/composables/useWorkspaceSwitcher.js)
  - [useWorkspaceStore.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/stores/setting/useWorkspaceStore.ts)
- Workspace onboarding status:
  - [useWorkspaceOnboarding.js](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/composables/useWorkspaceOnboarding.js)
  - [WorkspaceController.php](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-backend/app/Http/Controllers/Settings/WorkspaceController.php)

---

## 2. User Flows

### Flow A: Complete Signup Onboarding and Prepare Sample Workspace

**Happy Path:**

1. User signs up and enters the full-page onboarding flow.
2. User completes the onboarding steps for profile, business type, social-connect choice, and brand website where applicable.
3. On the final onboarding step, ContentStudio marks signup onboarding as completed and immediately starts a background job to create and seed a private sample workspace.
4. ContentStudio keeps the user's real workspace intact as their primary workspace.
5. Instead of dropping the user directly onto the dashboard, ContentStudio shows a new post-onboarding start screen with two clear options:
   - **Your Workspace**: Start with your real workspace
   - **Explore Sample Workspace**: Open a fully populated sample workspace
6. If sample provisioning has already finished, both options are available immediately.
7. If sample provisioning is still running, the sample card shows a loading state such as:
   - **Preparing your sample workspace**
   - **We’re generating realistic posts, chats, inbox threads, and analytics for you**
8. The user can either:
   - enter their real workspace immediately, or
   - wait a moment and open the sample workspace as soon as it is ready

**Alternative Flows:**

- **A1: Provisioning finishes instantly** → user sees the sample option immediately and can open it without delay.
- **A2: Provisioning is still running** → user enters their real workspace first and gets a non-blocking notification when the sample workspace is ready.
- **A3: Provisioning fails** → user continues into their real workspace; the product shows a retry-ready message rather than blocking the account.
- **A4: Ineligible plan** → for v1, API-centric or otherwise ineligible users skip this flow and continue with their normal post-onboarding route.

---

### Flow B: First Entry into the Sample Workspace

**Happy Path:**

1. User chooses **Explore Sample Workspace** from the post-onboarding start screen.
2. ContentStudio switches the active workspace to the sample workspace using the existing workspace-switching flow.
3. The user lands on the normal product dashboard, but with a persistent banner and workspace badge that clearly say they are inside a sample environment.
4. The banner explains:
   - this workspace contains generated sample data
   - it is private to this user
   - it is intended for product exploration
   - live external actions are limited or sandboxed
5. The dashboard loads with realistic seeded cards and counts instead of empty states.
6. The workspace switcher now visibly shows:
   - the user's real workspace
   - the sample workspace with a `Sample` badge
7. The standard in-product onboarding widget does not appear in the sample workspace.

**Alternative Flows:**

- **B1: User skips sample workspace on day 1** → the sample workspace remains available in the switcher and in empty-state CTAs inside the real workspace.
- **B2: User returns later** → the sample workspace remains selectable without re-running onboarding.

---

### Flow C: Explore Planner and Publishing in the Sample Workspace

**Happy Path:**

1. User opens Planner in the sample workspace.
2. User sees a realistic calendar populated across:
   - the previous month
   - the current month
   - the next two months
3. The calendar includes a believable mix of:
   - published posts
   - scheduled posts
   - posts in review
   - rejected posts with comments
   - posts grouped into campaigns/content themes
4. User opens a post and can inspect:
   - content
   - media
   - approval state
   - comments
   - labels/categories
   - linked accounts
5. User can safely interact with internal product workflows such as:
   - filtering views
   - opening composer on an existing sample post
   - changing review/approval state
   - browsing media and post details
6. If the user attempts an action that would touch a real external platform, ContentStudio explains that sample workspaces are for exploration and offers a **Switch to Your Workspace** CTA.

**Alternative Flows:**

- **C1: User creates a new post in sample workspace** → allowed only if the resulting action stays inside the sandbox; live external publishing is blocked in v1.
- **C2: User attempts to connect a real social account inside sample workspace** → system redirects the user to their real workspace before allowing account connection.

---

### Flow D: Explore AI in the Sample Workspace

**Happy Path:**

1. User opens the AI surfaces from dashboard, chat, or content-library entry points.
2. User sees pre-seeded chat history and generated assets that make the workspace feel already in use.
3. Seeded AI examples include:
   - captions
   - social post variants
   - image generations
   - video generations
   - saved brand/content artifacts
4. The user can continue interacting with AI tools normally because these are internal, non-platform actions.
5. Newly generated AI content saves into the sample workspace's own AI folders and content-library records.

**Alternative Flows:**

- **D1: User wants to test AI with their own brand context** → they can switch to their real workspace and use real brand setup there.

---

### Flow E: Explore Inbox in the Sample Workspace

**Happy Path:**

1. User opens Inbox in the sample workspace.
2. User sees seeded conversations, comments, review items, and assignments across dummy accounts.
3. User can inspect realistic operational states such as:
   - unread
   - assigned
   - pending
   - archived
   - marked done
4. User can interact with safe inbox actions that do not require real platform connectivity.
5. Reply/composer actions inside the sample workspace are treated as sandbox behavior rather than live social replies.

**Alternative Flows:**

- **E1: User tries to use a live platform reply path** → system keeps the action inside sample mode or blocks it with a clear explanation, depending on the endpoint.

---

### Flow F: Explore Analytics in the Sample Workspace

**Happy Path:**

1. User opens Analytics in the sample workspace.
2. User sees complete-looking overview and per-account charts for dummy accounts.
3. The data is consistent with the sample workspace's campaigns, posts, and channels, but it is synthetic.
4. Date-range changes still work and return coherent sample analytics.
5. Any manual sync or platform refresh action is hidden or disabled in the sample workspace.

**Alternative Flows:**

- **F1: User expects live synced analytics** → UI copy clarifies that analytics are sample data in this workspace and points the user to their real workspace for live account analytics.

---

### Flow G: Switch Back to the Real Workspace and Start Real Work

**Happy Path:**

1. User clicks **Switch to Your Workspace** from the sample banner or workspace switcher.
2. ContentStudio loads the user's real workspace.
3. The real workspace remains intentionally empty or near-empty, preserving a clean starting point for real work.
4. Empty states inside the real workspace include a lightweight CTA:
   - **Explore Sample Workspace**
5. The user can now:
   - connect real accounts
   - configure real categories/team/data
   - use live publishing flows

**Alternative Flows:**

- **G1: User wants to compare both** → switching remains available anytime through the workspace switcher.

---

## 3. Key Design Decisions

### Decision 1: When should the sample workspace be created?

| Option | Pros | Cons |
|---|---|---|
| **A: During raw account signup** | Fastest possible availability | Wastes provisioning on abandoned signups, low personalization, creates extra cost early |
| **B: After signup onboarding completion (Recommended)** | Uses real onboarding context, avoids abandoned-signup waste, matches user mental model | Requires async completion handling |
| C: Lazy-create only when user clicks explore sample | Avoids provisioning users who never use it | Adds delay at the exact moment the user wants to explore |

**Recommendation: Option B.**

Create the sample workspace after signup onboarding is completed. This aligns best with the user journey and gives us business context for better sample data.

---

### Decision 2: Where should the user land after onboarding?

| Option | Pros | Cons |
|---|---|---|
| A: Always land in real workspace | Simple, preserves ownership mental model | Most users will never discover the sample workspace |
| B: Always land in sample workspace | Maximizes feature discovery | Users may mistake the sample data for their actual business data |
| **C: Show a post-onboarding chooser (Recommended)** | Clear mental model, high discovery, low confusion | Requires one extra lightweight screen |

**Recommendation: Option C.**

Use a short post-onboarding start screen with two choices: real workspace or sample workspace. This keeps the sample visible without corrupting the user's understanding of what belongs to them.

---

### Decision 3: How interactive should the sample workspace be?

| Option | Pros | Cons |
|---|---|---|
| A: Read-only sample | Lowest risk | Feels fake and shallow; weak product exploration |
| **B: Safe sandboxed interactions (Recommended)** | Feels real while staying safe; users can learn core workflows | Requires branching for risky actions |
| C: Fully live behavior | Maximum realism | Too risky with dummy accounts and synthetic data |

**Recommendation: Option B.**

Allow internal, safe interactions in the sample workspace, but block or sandbox actions that depend on live connected accounts, syncs, or external publishing.

---

### Decision 4: How visually distinct should the sample workspace be?

| Option | Pros | Cons |
|---|---|---|
| A: No visual distinction beyond workspace name | Lowest design effort | High confusion risk; users may mistake sample data for their own |
| **B: Banner + badges + contextual callouts (Recommended)** | Clear and lightweight; enough differentiation without a parallel UI system | Requires cross-module UI touchpoints |
| C: Full alternate theme/skin | Most obvious distinction | High cost, high maintenance, unnecessary for v1 |

**Recommendation: Option B.**

Use a persistent top banner plus sample badges and a few contextual callouts. That creates strong clarity without turning the demo workspace into a design branch of the product.

---

## 4. Integration with Existing Features

### Backend vs Frontend Ownership

- The dummy workspace data itself should be created and stored by backend/services.
- The frontend should not own the sample dataset through static fixtures or hardcoded mock JSON.
- Backend responsibilities:
  - create the sample workspace
  - flag it as a sample workspace
  - seed planner, AI, inbox, media, team, and supporting records
  - return synthetic analytics for sample workspace/account IDs
  - enforce sandbox restrictions for risky actions
- Frontend responsibilities:
  - show the post-onboarding chooser
  - label the sample workspace in the workspace switcher
  - show the persistent sample banner
  - render sample data through normal workspace-scoped APIs
  - surface `Explore Sample Workspace` CTAs from real-workspace empty states

### Signup Onboarding

- The trigger should attach to the completion of the signup onboarding flow, not the older workspace-creation onboarding route.
- Relevant flow ownership lives in:
  - [useUserOnboarding.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/modules/account/composables/useUserOnboarding.ts)
  - [useOnboardingFlow.ts](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/modules/onboarding/composables/useOnboardingFlow.ts)
  - [router.js](/home/casper/code/contentstudio-agentic-product-pipeline/contentstudio-frontend/src/router.js)

### Real Workspace Onboarding Widget

- The real workspace should keep the existing onboarding widget behavior.
- The sample workspace should not use that widget, because seeded data already demonstrates the product.

### Workspace Switching

- This feature should reuse the normal workspace list, fetch, and switch behavior.
- The main addition is metadata and UI treatment for identifying the sample workspace.

### Planner / Composer / Approvals

- The sample workspace should feel populated and useful in Planner immediately.
- Approval and comment flows should feel explorable.
- External publish paths should remain protected.

### AI

- AI is the best place to allow rich interaction in the sample workspace because it is internal and not dependent on real social accounts.

### Inbox

- Inbox should feel operational, not empty.
- Safe operational actions should work, but live platform transport should remain sandboxed or blocked.

### Analytics

- Analytics should be explicitly labeled as sample data in the sample workspace.
- Manual sync/refresh behavior should not pretend to be live.

### Empty-State Strategy in the Real Workspace

- The real workspace should stay clean.
- Dashboard and product empty states should include an **Explore Sample Workspace** CTA so users can re-enter the sample environment later.

---

## 5. Scope Recommendation

### Include in v1

- Create one private sample workspace for each newly onboarded full-suite user
- Trigger provisioning after signup onboarding completion
- Add a post-onboarding start screen
- Add `Sample` labeling in workspace switcher and a persistent sample banner
- Seed realistic planner, AI, inbox, team, media, labels, categories, and related data
- Generate planner dates relative to the provisioning date and workspace timezone
- Provide synthetic analytics for the sample workspace
- Keep risky actions sandboxed or blocked with clear copy
- Add sample-workspace discovery CTAs in empty states of the real workspace

### Defer to v2

- Reset/re-generate sample workspace from the UI
- Multiple sample scenarios per business type or industry pack selection
- Backfilling sample workspaces for all existing users
- Copy sample posts/assets directly into the real workspace
- Sample workspace support for API-centric plans
- Multi-user collaboration inside the sample workspace

### Explicit Recommendation

V1 should optimize for:

- clarity
- safe exploration
- fast product comprehension

It should not try to make the sample workspace indistinguishable from a real live workspace. The user should feel they are exploring a high-quality sandbox, not managing live accounts.
