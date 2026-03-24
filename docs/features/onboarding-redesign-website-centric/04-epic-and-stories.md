# Epic + Stories — Onboarding Redesign (Website-Centric)

> **Prototype (source of truth for all UI copy, flow, and layout):** https://cs-onboarding-website-centric.lovable.app/
>
> All UI copy, labels, placeholders, button text, step ordering, and layout details must be taken directly from the prototype. The stories below describe the technical scope and acceptance criteria — the prototype defines the visual and copy spec.

---

## Epic

**Title:** Onboarding Redesign — Page-Based Flow with Website-Centric Onboarding

**Description:**
Redesign the ContentStudio onboarding experience from a modal-based flow to a dedicated full-page flow. Currently, onboarding steps (profile setup, role selection, social account connection) are presented as modals overlaying the app. This redesign converts all steps into standalone pages with their own routes, creating a more focused and immersive onboarding experience.

Additionally, a new fourth step is introduced: website-centric onboarding. Users can enter their website URL, review fetched website data, and approve it to trigger automatic social post generation — giving new users immediate value from day one. The video intro is relocated from the onboarding flow to the home page.

The complete prototype with all UI copy, flow, and interactions is available at: https://cs-onboarding-website-centric.lovable.app/

**Objective:** Q2 2026

---

## Stories

---

### Story 1: [BE] Update onboarding API to support page-based flow and website step

**Description:**
As a backend developer, I want to extend the onboarding API to support the new page-based flow and the website-centric step so that the frontend can save the user's website URL, fetch website metadata, and trigger social post generation during onboarding.

Currently, the onboarding API (`/onboarding/steps`, `/onboarding/status`) tracks 6 completion steps. This story adds:
- A new `website` field on the Workspace model to store the user's website URL
- A new onboarding step (`enter_website`) in the allowed steps list
- A new API endpoint to accept a website URL, fetch metadata/content from it, and return structured data (title, description, logo, key content)
- A new API endpoint to trigger social post generation from the fetched website data
- A new API endpoint to check the status/progress of post generation

**Relevant existing code:**
- Controller: `app/Http/Controllers/Settings/WorkspaceController.php` — `performOnboardingStep()` (line 1135), `setOnboardingStatus()` (line 1168)
- Routes: `routes/web/settings.php` (lines 157-163)
- Model: `app/Models/Settings/Workspace.php` — `onboarding_steps`, `onboarding` fields
- AI Brand Controller: `app/Http/Controllers/Onboarding/OnboardingBrandController.php` — reference for AI-driven onboarding patterns
- Jobs: `OnboardingSubmissionJob`, `OnboardingPlansJob` — reference for background job patterns

---

**Workflow:**
1. User enters their website URL on the onboarding website step page
2. System receives the URL via a new API endpoint, validates it, fetches metadata and content from the website
3. System returns structured data (site title, description, logo/favicon, key content snippets) to the frontend for preview
4. User reviews and optionally edits the fetched data, then approves it
5. System receives the approval and triggers a background job to generate social posts from the website content
6. Frontend polls a status endpoint to show generation progress
7. User can skip at any point — the website step and post generation are optional

---

**Acceptance criteria:**
- [ ] New `website` field added to the Workspace model (string, nullable)
- [ ] `enter_website` added to the allowed onboarding steps in `performOnboardingStep()`
- [ ] `POST /onboarding/website` endpoint accepts `{ workspace_id, url }`, validates the URL format, fetches website metadata, and returns structured data (title, description, favicon, content snippets)
- [ ] `POST /onboarding/website/generate` endpoint accepts `{ workspace_id, website_data }` and triggers a background job to generate social posts from the website content
- [ ] `GET /onboarding/website/generate/status` endpoint returns the current generation progress (pending, in_progress, completed, failed) and the count of generated posts
- [ ] Website fetch handles common edge cases: timeouts (10s max), invalid URLs, unreachable sites, sites with no parseable content — returns appropriate error messages
- [ ] Background post generation job is queued and does not block the API response
- [ ] All new endpoints require authentication and workspace validation
- [ ] Existing onboarding steps and status endpoints continue to work unchanged

---

**Mock-ups:**
N/A — backend only. See prototype for the frontend flow that consumes these APIs.

---

**Impact on existing data:**
- New `website` field added to the `workspace` MongoDB collection (nullable, no migration needed for existing documents)
- New `enter_website` step added to allowed onboarding steps — existing workspaces won't have this step in their `onboarding_steps` object, which is fine (treated as not completed)

---

**Impact on other products:**
- Generated social posts will appear in the user's drafts/composer — same as any AI-generated content
- No impact on mobile apps (onboarding redesign is web-only)
- No impact on Chrome extension

---

**Dependencies:**
- May depend on AI agent pipeline (`contentstudio-ai-agents/`) for website content parsing and post generation
- No dependencies on other stories in this epic — this can be developed in parallel with frontend stories

---

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 2: [FE] Convert onboarding from modal to page-based routing with guards

**Description:**
As a new user, I want the onboarding experience to be a focused full-page flow instead of a modal overlay so that I can complete setup without distractions.

Currently, onboarding steps are rendered as modals inside `OnBoardingMain.vue` in the account module (`src/modules/account/views/onboarding/`). This story converts the entire onboarding flow to dedicated page routes with a new full-page layout. It also implements route guards so that users with incomplete onboarding are always redirected to the appropriate onboarding step.

**All layout, styling, step ordering, and transitions must match the prototype exactly:** https://cs-onboarding-website-centric.lovable.app/

**Relevant existing code:**
- Current routes: `src/modules/onboarding/config/routes/onboarding.js` — only has `/onboarding/create-workspace`
- Modal views: `src/modules/account/views/onboarding/OnBoardingMain.vue`, `WelcomeOnboard.vue`, `BusinessType.vue`, `SocialConnect.vue`, `VideoIntro.vue`
- Composable: `src/composables/useWorkspaceOnboarding.js` — `isOnboardingComplete` computed
- App layout: `App.vue` — already detects onboarding routes via `route.name`
- Store: `src/modules/onboarding/store/onboarding.js`

---

**Workflow:**
1. New user signs up and creates a workspace
2. Instead of seeing a modal overlay, user lands on a full-page onboarding screen at `/onboarding`
3. User sees a clean, distraction-free page layout with a step indicator showing their progress (as shown in prototype)
4. User navigates through steps sequentially — each step is a separate route (`/onboarding/profile`, `/onboarding/role`, `/onboarding/connect`, `/onboarding/website`)
5. If user tries to navigate to the main app (e.g., `/home`) while onboarding is incomplete, they are redirected back to their current onboarding step
6. Once all required steps are completed (or skipped where allowed), user is redirected to the home page
7. Returning users who already completed onboarding go directly to the home page as normal

---

**Acceptance criteria:**
- [ ] New page-based routes created: `/onboarding/profile`, `/onboarding/role`, `/onboarding/connect`, `/onboarding/website` (and a base `/onboarding` that redirects to the first incomplete step)
- [ ] Full-page layout component created for onboarding — no app sidebar, no header, clean focused layout matching the prototype
- [ ] Step progress indicator visible on all onboarding pages, showing current step and completed steps (as per prototype design)
- [ ] Route guard implemented: if `workspace.onboarding === false`, user is redirected to `/onboarding` (which resolves to their first incomplete step)
- [ ] Route guard allows skipping forward only for steps that are skippable (social connect and website are skippable; profile and role are required)
- [ ] Back navigation between steps works correctly
- [ ] `App.vue` layout correctly hides sidebar/header when on any onboarding route
- [ ] Old modal-based onboarding components are deprecated (can be removed in a follow-up cleanup)
- [ ] `useWorkspaceOnboarding.js` composable updated to track the new step flow and page-based state
- [ ] All transitions and animations between steps match the prototype
- [ ] Page is responsive and works on tablet/mobile viewports

---

**Mock-ups:**
All layout and design specs are in the prototype: https://cs-onboarding-website-centric.lovable.app/

---

**Impact on existing data:**
- No data model changes — this story only restructures the frontend routing and layout
- Existing `onboarding_steps` data in workspace documents remains compatible

---

**Impact on other products:**
- No impact on mobile apps — mobile has its own onboarding flow
- No impact on Chrome extension
- Existing deep links or bookmarks to `/onboarding/create-workspace` should redirect appropriately

---

**Dependencies:**
- None — this is the foundational FE story that other FE stories build upon

---

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 3: [FE] Redesign onboarding profile and role selection pages

**Description:**
As a new user, I want to enter my full name and select my role on clean, dedicated onboarding pages so that ContentStudio can personalize my experience.

This story converts the existing modal-based profile setup (`WelcomeOnboard.vue`) and role/business type selection (`BusinessType.vue`) into full-page onboarding steps. All UI copy, labels, field placeholders, role options, and layout must be taken directly from the prototype.

**Prototype (source of truth for all copy and layout):** https://cs-onboarding-website-centric.lovable.app/

**Relevant existing code:**
- `src/modules/account/views/onboarding/WelcomeOnboard.vue` — Current profile setup modal
- `src/modules/account/views/onboarding/BusinessType.vue` — Current role selection modal
- `src/modules/account/views/onboarding/ProfileAvatarCard.vue` — Avatar component

**UI Components to use:**
- `TextInput` from `@contentstudio/ui` for name/field inputs
- `Button` from `@contentstudio/ui` for CTAs
- `Avatar` from `@contentstudio/ui` for profile picture display
- `Radio` or `CstCardCheckbox` for role selection cards (use card-style selection as shown in prototype)

---

**Workflow:**
1. User arrives at the profile setup page (`/onboarding/profile`) — the first onboarding step
2. User sees a clean page layout with fields to enter their information (as shown in prototype — all labels, placeholders, and helper text from prototype)
3. User fills in required fields and optionally uploads a profile photo
4. User clicks the continue button to proceed
5. User arrives at the role selection page (`/onboarding/role`)
6. User sees role/business type options displayed as selectable cards (as shown in prototype — all role names and descriptions from prototype)
7. User selects their role and clicks continue to proceed to the next step
8. Both steps save data via the existing onboarding API endpoints

---

**Acceptance criteria:**
- [ ] Profile page (`/onboarding/profile`) renders as a full page within the onboarding layout
- [ ] All form fields, labels, placeholders, helper text, and validation messages match the prototype exactly
- [ ] Profile photo upload works with preview (reuse existing `ProfileAvatarCard.vue` logic or refactor)
- [ ] Required field validation — continue button is disabled until required fields are filled; shows inline validation errors from prototype
- [ ] Role selection page (`/onboarding/role`) renders as a full page with selectable role cards
- [ ] All role options, card labels, and descriptions match the prototype exactly
- [ ] User must select a role before proceeding — continue button disabled until selection made
- [ ] Both pages save user data to the backend on continue (reuse existing API calls)
- [ ] Step progress indicator updates correctly on both pages
- [ ] Both pages use `@contentstudio/ui` components (`TextInput`, `Button`, `Avatar`) — no legacy `Cst*` components for elements that have modern equivalents
- [ ] All user-facing strings use i18n (`$t()` / `t()`) with keys added to all locale files
- [ ] Uses CSS variable theming classes (`text-primary-cs-500`, `bg-primary-cs-50`, etc.) — no hardcoded colors

---

**Mock-ups:**
See prototype: https://cs-onboarding-website-centric.lovable.app/ — Steps 1 and 2

---

**Impact on existing data:**
- No data model changes — profile and role data saves to the same fields as the current modal flow

---

**Impact on other products:**
- No impact on mobile apps or Chrome extension

---

**Dependencies:**
- Depends on: **[FE] Convert onboarding from modal to page-based routing with guards** (for the page layout and routing infrastructure)

---

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 4: [FE] Redesign onboarding social account connection page

**Description:**
As a new user, I want to connect my social media accounts on a dedicated onboarding page so that I can start managing my social presence right away, or skip this step if I prefer to do it later.

This story converts the existing modal-based social account connection (`SocialConnect.vue`) into a full-page onboarding step. All UI copy, platform options, button labels, and skip behavior must be taken directly from the prototype.

**Prototype (source of truth for all copy and layout):** https://cs-onboarding-website-centric.lovable.app/

**Relevant existing code:**
- `src/modules/account/views/onboarding/SocialConnect.vue` — Current social connect modal
- `src/modules/account/views/onboarding/ConnectedAccountsDropdown.vue` — Connected accounts display
- `src/modules/onboarding/store/onboarding.js` — Tracks connections per platform (Facebook, Twitter, LinkedIn, Pinterest, etc.)

**UI Components to use:**
- `Button` from `@contentstudio/ui` for connect and skip CTAs
- `Avatar` from `@contentstudio/ui` for connected account avatars
- `Badge` from `@contentstudio/ui` for connection status indicators

---

**Workflow:**
1. User arrives at the social connect page (`/onboarding/connect`) — third onboarding step
2. User sees available social platforms with connect buttons (as shown in prototype — all platform names, button labels, and descriptions from prototype)
3. User clicks a platform's connect button — OAuth flow opens in a popup/new tab
4. After successful connection, the platform shows as connected with the account name/avatar
5. User can connect multiple platforms or skip entirely
6. User clicks continue to proceed to the website step, or clicks skip to go to the next step
7. This step is optional — user can skip without connecting any accounts

---

**Acceptance criteria:**
- [ ] Social connect page (`/onboarding/connect`) renders as a full page within the onboarding layout
- [ ] All platform options, labels, descriptions, and button text match the prototype exactly
- [ ] Each platform shows a connect button that initiates the OAuth flow (reuse existing connection logic from `SocialConnect.vue`)
- [ ] Successfully connected platforms show the connected account name and avatar
- [ ] User can connect multiple platforms before proceeding
- [ ] Skip functionality works — user can skip this step without connecting any accounts
- [ ] Step progress indicator updates correctly
- [ ] Connection state persists if user navigates back and returns to this step
- [ ] Uses `@contentstudio/ui` components — no legacy `Cst*` for elements with modern equivalents
- [ ] All user-facing strings use i18n (`$t()` / `t()`)
- [ ] Uses CSS variable theming classes — no hardcoded colors
- [ ] Loading states shown during OAuth connection attempts

---

**Mock-ups:**
See prototype: https://cs-onboarding-website-centric.lovable.app/ — Step 3

---

**Impact on existing data:**
- No data model changes — social connections save to the same structure as the current modal flow

---

**Impact on other products:**
- No impact on mobile apps or Chrome extension
- OAuth callback URLs should remain unchanged

---

**Dependencies:**
- Depends on: **[FE] Convert onboarding from modal to page-based routing with guards** (for the page layout and routing infrastructure)

---

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 5: [FE] Build onboarding website input step with data preview and post generation

**Description:**
As a new user, I want to enter my website URL during onboarding so that ContentStudio can fetch my website data, show me a preview I can edit, and generate social media posts for me — giving me immediate value from day one.

This is the new fourth onboarding step. It introduces a website URL input, fetches and displays website metadata for user review/editing, and triggers automatic social post generation upon approval. Users can also skip this step entirely and go straight to the home page.

**All UI copy, field labels, placeholders, button text, loading states, progress indicators, success states, and skip behavior must be taken directly from the prototype:** https://cs-onboarding-website-centric.lovable.app/

**Relevant existing code:**
- Composable: `src/composables/useWorkspaceOnboarding.js` — needs new `enter_website` step
- Store: `src/modules/onboarding/store/onboarding.js` — needs website state tracking
- API utils: `src/config/api-utils.js` — add new onboarding/website endpoints

**UI Components to use:**
- `TextInput` from `@contentstudio/ui` for website URL input
- `Button` from `@contentstudio/ui` for submit, skip, and edit CTAs
- `Loader` from `@contentstudio/ui` for fetch/generation loading states
- `Progress` from `@contentstudio/ui` for generation progress indicator
- `Alert` from `@contentstudio/ui` for error states

---

**Workflow:**
1. User arrives at the website step (`/onboarding/website`) — fourth and final onboarding step
2. User sees a URL input field with supporting copy (as shown in prototype)
3. User enters their website URL and submits
4. System shows a loading state while fetching website data (as shown in prototype)
5. System displays the fetched data in an editable preview — site title, description, and other extracted content (as shown in prototype)
6. User reviews the data and can edit any field
7. User clicks approve/confirm to trigger social post generation
8. System shows a post generation progress page — with a progress indicator and status messages (as shown in prototype)
9. While posts are being generated, user can stay on the page to watch progress or skip to the home page
10. Once generation is complete, user sees a success state with the option to view generated posts or go to home
11. At any point before generation, user can skip this entire step and go directly to home

---

**Acceptance criteria:**
- [ ] Website step page (`/onboarding/website`) renders as a full page within the onboarding layout
- [ ] URL input field with all labels, placeholder text, and helper text matching the prototype
- [ ] URL validation — shows inline error for invalid URLs (validation message from prototype)
- [ ] Submit button triggers API call to fetch website data; shows loading state during fetch (as per prototype)
- [ ] Fetched data preview displays site title, description, and extracted content in editable fields (layout from prototype)
- [ ] User can edit any fetched field before approving
- [ ] Approve button triggers API call to start post generation; transitions to generation progress view
- [ ] Generation progress page shows progress indicator and status messages (as per prototype)
- [ ] User can skip to home at any point — skip button visible throughout (as per prototype)
- [ ] Generation completion shows success state with CTA to view posts or go to home (as per prototype)
- [ ] Error handling: if website fetch fails, show error state with retry option and option to skip (error copy from prototype)
- [ ] Error handling: if post generation fails, show error state with option to retry or skip to home
- [ ] `useWorkspaceOnboarding.js` updated to include `enter_website` step tracking
- [ ] New API endpoint URLs added to `src/config/api-utils.js`
- [ ] All user-facing strings use i18n (`$t()` / `t()`) with keys added to all locale files
- [ ] Uses CSS variable theming classes (`text-primary-cs-500`, `bg-primary-cs-50`, etc.) — no hardcoded colors
- [ ] Uses `@contentstudio/ui` components (`TextInput`, `Button`, `Loader`, `Progress`, `Alert`) — no legacy equivalents
- [ ] Loading, error, and empty states all implemented per prototype

---

**Mock-ups:**
See prototype: https://cs-onboarding-website-centric.lovable.app/ — Step 4 (website input → preview → generation → success)

---

**Impact on existing data:**
- New `enter_website` step will be tracked in `workspace.onboarding_steps` — existing workspaces without this step are unaffected (step treated as not started)
- Generated posts will appear in the user's composer/drafts

---

**Impact on other products:**
- Generated posts may appear in the Planner calendar if auto-scheduled
- No impact on mobile apps (onboarding redesign is web-only)
- No impact on Chrome extension

---

**Dependencies:**
- Depends on: **[FE] Convert onboarding from modal to page-based routing with guards** (for the page layout and routing infrastructure)
- Depends on: **[BE] Update onboarding API to support page-based flow and website step** (for the website fetch, generation, and status endpoints)

---

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 6: [FE] Relocate video intro from onboarding flow to home page

**Description:**
As a new user who just completed onboarding, I want to see the introductory video on the home page instead of during the onboarding flow so that I can start exploring the app and watch the video when I'm ready.

Currently, `VideoIntro.vue` is the final step of the modal-based onboarding. With the redesign, the video is no longer part of the onboarding step flow. Instead, it should appear on the home page when a user first lands there after completing onboarding.

**Prototype (source of truth for placement and behavior):** https://cs-onboarding-website-centric.lovable.app/

**Relevant existing code:**
- `src/modules/account/views/onboarding/VideoIntro.vue` — Current video intro modal
- `src/modules/common/components/widgets/GettingStarted.vue` — Dashboard getting started widget
- `src/composables/useWorkspaceOnboarding.js` — `onboardingStepsCompleted('watch_video')`

---

**Workflow:**
1. User completes the onboarding flow (or skips to home) and lands on the home page
2. User sees the introductory video prominently displayed (as a modal, banner, or inline widget — follow prototype placement)
3. User can watch the video or dismiss it
4. Dismissing or completing the video marks the `watch_video` onboarding step as complete
5. The video does not appear again on subsequent home page visits after being watched or dismissed

---

**Acceptance criteria:**
- [ ] Video intro removed from the onboarding step flow — no longer appears as an onboarding step
- [ ] Video appears on the home page for users who haven't watched it yet (check `watch_video` step status)
- [ ] Video placement and presentation match the prototype
- [ ] User can dismiss the video without watching — marks `watch_video` step as completed
- [ ] Watching the video to completion marks `watch_video` step as completed
- [ ] Video does not reappear after being watched or dismissed
- [ ] Existing `VideoIntro.vue` logic reused or refactored into the home page context
- [ ] All user-facing strings use i18n (`$t()` / `t()`)
- [ ] Works correctly for both new users (first time on home) and existing users who already watched the video

---

**Mock-ups:**
See prototype for video placement on home page: https://cs-onboarding-website-centric.lovable.app/

---

**Impact on existing data:**
- No data model changes — `watch_video` step in `onboarding_steps` continues to work the same way
- Users who already completed `watch_video` won't see it again

---

**Impact on other products:**
- No impact on mobile apps or Chrome extension
- The Getting Started widget (`GettingStarted.vue`) may need updating if it references the video step as part of onboarding

---

**Dependencies:**
- Depends on: **[FE] Convert onboarding from modal to page-based routing with guards** (onboarding flow must be updated first so the video is no longer part of it)

---

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Story Summary

| # | Story Title | Team | Group | Project | Priority | Skill Set | Product Area |
|---|---|---|---|---|---|---|---|
| 1 | [BE] Update onboarding API to support page-based flow and website step | Backend | Backend | Web App | High | Backend | Onboarding |
| 2 | [FE] Convert onboarding from modal to page-based routing with guards | Frontend | Frontend | Web App | High | Frontend | Onboarding |
| 3 | [FE] Redesign onboarding profile and role selection pages | Frontend | Frontend | Web App | High | Frontend | Onboarding |
| 4 | [FE] Redesign onboarding social account connection page | Frontend | Frontend | Web App | High | Frontend | Onboarding |
| 5 | [FE] Build onboarding website input step with data preview and post generation | Frontend | Frontend | Web App | High | Frontend | Onboarding |
| 6 | [FE] Relocate video intro from onboarding flow to home page | Frontend | Frontend | Web App | Medium | Frontend | Onboarding |

### Dependency Graph
```
[BE] Update onboarding API... ──────────────────────────────────┐
                                                                 ├──→ [FE] Build website input step...
[FE] Convert onboarding from modal to page-based routing... ────┤
                                                                 ├──→ [FE] Redesign profile and role pages
                                                                 ├──→ [FE] Redesign social account connection page
                                                                 └──→ [FE] Relocate video intro to home page
```
