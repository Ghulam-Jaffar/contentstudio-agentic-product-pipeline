# Codebase Analysis — Onboarding Redesign (Website-Centric)

> Competitor research skipped per product owner request. Prototype serves as the spec: https://cs-onboarding-website-centric.lovable.app/

---

## Current Onboarding Architecture

### Frontend

**Entry Points & Routes:**
- `/src/modules/onboarding/config/routes/onboarding.js` — Defines `/onboarding` route, redirects to `/onboarding/create-workspace`
- Only one sub-route exists: `create-workspace` → `CreateWorkspaces.vue`
- After workspace creation, onboarding steps are handled as **modals** inside the account module

**Onboarding Step Views (Modal-Based):**
- `/src/modules/account/views/onboarding/OnBoardingMain.vue` — Root layout wrapper
- `/src/modules/account/views/onboarding/WelcomeOnboard.vue` — Step 1: Name/profile setup
- `/src/modules/account/views/onboarding/BusinessType.vue` — Step 2: Role/business type selection
- `/src/modules/account/views/onboarding/SocialConnect.vue` — Step 3: Connect social accounts (skippable)
- `/src/modules/account/views/onboarding/VideoIntro.vue` — Final: Video introduction
- `/src/modules/account/views/onboarding/CreateWorkspaces.vue` — Workspace creation
- `/src/modules/account/views/onboarding/UserCredentials.vue` — User credentials
- `/src/modules/account/views/onboarding/ProfileAvatarCard.vue` — Avatar selection
- `/src/modules/account/views/onboarding/ConnectedAccountsDropdown.vue` — Account dropdown

**State Management:**
- Vuex store: `/src/modules/onboarding/store/onboarding.js` — Tracks social account connections per platform, workspace list, loaders
- Workspace type: `/src/types/common/workspace.ts` — Defines `OnboardingSteps` interface with 6 steps + 1 bonus

**Composables:**
- `/src/composables/useWorkspaceOnboarding.js` — Primary composable
  - `isOnboardingComplete` — checks `workspace.onboarding === true`
  - `shouldShowOnboardingWidget` — shows if 0-99% progress and not hidden
  - `onboardingStepsCompleted(step)` — POST to mark step done
  - `onboardingCompleted(status)` — Mark entire onboarding done
- `/src/composables/useOnboarding.js` — Legacy/alternate, used in dashboard

**Post-Onboarding Widget:**
- `/src/modules/onboarding/components/OnboardingConfirmation.vue` — Modal to confirm/dismiss
- `/src/modules/common/components/widgets/GettingStarted.vue` — Dashboard widget tracking progress

**Route Guards & App.vue:**
- `App.vue` adjusts layout based on `route.name` containing "onboarding"
- Router comments warn: "CHANGING NAME OF ANY ROUTE CAN EFFECT 'App.vue' PAGE CONDITIONS"

### Backend

**API Routes** (`/routes/web/settings.php`, lines 157-163):
```
POST /onboarding/steps — Record step completion (WorkspaceController@performOnboardingStep)
POST /onboarding/status — Mark onboarding complete/incomplete (WorkspaceController@setOnboardingStatus)
POST /onboarding/widget/never — Hide widget preference (WorkspaceController@setNeverShowOnboardingWidget)
```

**Controller:** `/app/Http/Controllers/Settings/WorkspaceController.php`
- `performOnboardingStep()` (line 1135) — Validates workspace_id + step name, records completion
- `setOnboardingStatus()` (line 1168) — Sets `workspace.onboarding = status`
- `setNeverShowOnboardingWidget()` (line 1089) — Hides widget per workspace/user

**Additional Onboarding Controllers:**
- `/app/Http/Controllers/Integrations/OnboardingController.php` — Handles initial onboarding submission + social account migration
- `/app/Http/Controllers/Onboarding/OnboardingBrandController.php` — AI brand generation during onboarding
- `/app/Libraries/OnboardingHelper.php` — Utility for collecting social account IDs

**Background Jobs:**
- `OnboardingSubmissionJob`, `OnboardingPlansJob`, `SetOnboardingUserStatus`

**Data Models:**
- Workspace (`/app/Models/Settings/Workspace.php`): `onboarding` (boolean), `onboarding_steps` (object), `hide_onboarding_widget`
- User (`/app/Models/Account/User.php`): `show_onboarding_widget`, `business_name`, `business_type`, `signup_on_boarding`

**Current Onboarding Steps Tracked:**
1. `watch_video`
2. `connect_social_account`
3. `create_first_post`
4. `content_category`
5. `discover_content`
6. `invite_team`
7. `accounts_connection_modal_closed` (bonus, hidden)

### Key Finding: No Website Field Exists
- No `website` field on User or Workspace models
- `user_websites()` relationship exists but points to WordPress blog connections, not a generic website URL
- **New field required** on Workspace or User model for the website URL

---

## What Needs to Change (Modal → Page Conversion)

### Frontend Changes Required

1. **New page-based routes** — Replace modal flow with dedicated `/onboarding/step-1`, `/onboarding/step-2`, etc. routes
2. **New page layout** — Full-page onboarding layout (not modal overlay). `App.vue` already has onboarding route detection
3. **Route guards** — If `workspace.onboarding === false`, redirect user to `/onboarding` instead of home
4. **New Step 4: Website input** — Entirely new step with URL input, data fetching, preview/edit, and post generation
5. **Update composables** — `useWorkspaceOnboarding.js` needs new step tracking (`enter_website` or similar)
6. **Vuex/Pinia store updates** — Track website input state, fetched data, generation progress

### Backend Changes Required

1. **New field** — `website` on Workspace or User model
2. **New API endpoint** — Accept website URL, fetch metadata/content
3. **New API endpoint** — Trigger social post generation from website data
4. **New onboarding step** — Add `enter_website` to allowed steps in `performOnboardingStep()`
5. **Post generation job** — Background job to generate posts from website content (likely uses AI agents)

### Integration Points

- **AI Agents (`contentstudio-ai-agents/`)** — Website content parsing and post generation likely goes through the AI agent pipeline
- **Composer** — Generated posts may feed into the composer/drafts
- **Social Account Connection** — Step 3 remains but as a page instead of modal
- **Video intro** — Moves to home page (post-onboarding), no longer part of the step flow
