# Epic + Stories — API-Centric Plan

## Epic

**Title:** API-Centric Plan
**Description:**

Introduce a fourth pricing tier to ContentStudio, purpose-built for developers and automation-first users who need API access and publishing tools without the full social media management suite. This epic covers the end-to-end implementation: new plan type in the database, signup flow with smart routing, shortened onboarding for API users, a dedicated API Dashboard, reduced navigation that hides irrelevant modules, preview mode for locked features, billing/plan switching with auto-calculated add-ons, and website updates to surface the new plan.

The API-Centric Plan captures a growing user segment currently underserved by the existing plans, reduces churn from mismatched expectations, and positions ContentStudio competitively against API-first tools like Late/Zernio and Ayrshare. See the linked Research and PRD documents for full competitor analysis, codebase analysis, and detailed requirements.

---

## Stories

---

### Story 1: [BE] Create API-Centric plan type with feature flags and limits in the subscription system

**Description:**
As a backend developer, I want to add a new API-Centric plan type to the subscription system so that the platform can distinguish API-first users from full-suite users and gate features accordingly.

**Workflow:**
1. A new user signs up and selects the API Plan (or arrives via `?mode=api`)
2. The system creates their account with `plan = api-centric-trial` and `user_type = api`
3. Their subscription record contains `features` with `social_inbox: false`, `ai_studio: false`, `analytics: false`, `discovery: false`, `blog_publishing: false`, `automation: false`, `white_label: false`, `sso: false`, `brand_knowledge: false`, and `api_access: true`, `content_publishing: true`, `content_library: true`, `integrations: true`
4. Their subscription `limits` contain API-specific values: `api_calls_per_month`, `social_accounts`, `workspaces`, `users`, `api_keys`
5. The existing `SubscriptionMiddleware` reads these features and correctly gates access
6. Existing `useFeatures()` composable on the frontend automatically hides modules based on these feature flags

**Acceptance criteria:**
- [ ] New plan slugs exist in `subscription_plans` collection: `api-centric` (monthly) and `api-centric-annual` (annual)
- [ ] Plan record contains correct `features` object with all excluded features set to `false` and included features set to `true`
- [ ] Plan record contains `limits` object with API-specific limits (api_calls_per_month, social_accounts, workspaces, users, api_keys)
- [ ] `user_type` field added to user/workspace model — values: `api` or `standard` (default: `standard`)
- [ ] `SubscriptionMiddleware` correctly denies access to excluded features for API plan users
- [ ] API endpoint `GET /api/plan` returns the full plan object including `user_type` for API plan users
- [ ] Trial variant exists: `api-centric-trial` with same feature set and trial-specific limits
- [ ] Paddle plan IDs configured for monthly and annual billing (or equivalent billing system config)
- [ ] Database migration/seed script creates the new plan records

**Mock-ups:** N/A — backend only

**Impact on existing data:**
- New records added to `subscription_plans` collection
- New `user_type` field on user/workspace model (existing users default to `standard`)
- No changes to existing plan records

**Impact on other products:**
- Frontend will use the `features` and `user_type` to gate navigation (separate FE story)
- Mobile apps: no impact — API plan is web-only

**Dependencies:** None — this is the foundational story that other stories depend on.

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — N/A, backend only
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 2: [BE] Handle signup mode flag routing and API user account creation

**Description:**
As a backend developer, I want to handle the `?mode=api` and `?mode=suite` signup URL parameters so that the system can route new users to the correct onboarding flow and assign the appropriate plan type during account creation.

**Workflow:**
1. User arrives at signup page with `?mode=api` in the URL (from website API CTA)
2. User submits the signup form (email, password, business name)
3. Backend detects the `mode=api` parameter
4. System creates the account with `plan = api-centric-trial`, `user_type = api`, business name as workspace name
5. Backend returns a response indicating the user should skip plan selection and go to API onboarding
6. If `?mode=suite` or no flag: account created with `plan = advanced-trial`, `user_type = standard` (existing behavior)

**Acceptance criteria:**
- [ ] Signup endpoint accepts optional `mode` parameter (`api`, `suite`, or absent)
- [ ] When `mode=api`: account created with `plan = api-centric-trial` and `user_type = api`
- [ ] When `mode=suite` or absent: account created with existing default behavior (`advanced-trial`, `user_type = standard`)
- [ ] Business name from signup form is used as default workspace name (existing behavior preserved)
- [ ] Signup response includes `user_type` and `skip_plan_selection` boolean so frontend knows which flow to enter
- [ ] Mode flag is validated (only `api` and `suite` accepted; other values treated as absent)
- [ ] Trial period begins on account creation (same duration for both plan types)

**Mock-ups:** N/A — backend only

**Impact on existing data:**
- No changes to existing signup behavior when mode is `suite` or absent
- New `user_type` field populated during account creation

**Impact on other products:**
- Frontend signup flow will read the `skip_plan_selection` flag (separate FE story)
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Create API-Centric plan type with feature flags and limits in the subscription system**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — N/A, backend only
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 3: [BE] Support plan switching from API Plan to full-suite with auto-calculated add-ons

**Description:**
As a backend developer, I want to support plan switching from the API Plan to Advanced or Agency Unlimited plans so that API users can upgrade to the full suite while preserving their current usage and auto-calculating add-ons for any excess.

**Workflow:**
1. API user clicks "Change Plan" from the billing page
2. System returns available upgrade plans (Advanced and Agency Unlimited only — no Standard)
3. User selects a target plan
4. System compares current API usage (social accounts, workspaces, users) against target plan's base limits
5. For any usage exceeding base limits, system calculates required add-ons and pricing
6. User confirms the switch with a full pricing breakdown
7. System processes the plan change via Paddle, applies add-ons, changes `user_type` to `standard`, updates `features` and `limits`

**Acceptance criteria:**
- [ ] Endpoint returns only Advanced and Agency Unlimited as upgrade options for API plan users (Standard excluded)
- [ ] Pricing breakdown endpoint compares current usage vs. target plan base limits and returns: base price, add-on items with quantities and prices, total monthly price
- [ ] Plan switch endpoint processes the upgrade: updates subscription slug, features, limits, and changes `user_type` from `api` to `standard`
- [ ] Connected social accounts are preserved during the switch
- [ ] Workspace data, team members, and content are preserved during the switch
- [ ] Paddle billing is updated correctly (new plan + add-ons)
- [ ] If user's current usage fits within target plan base limits, no add-ons are calculated
- [ ] Plan switch is atomic — if billing fails, no partial state changes occur

**Mock-ups:** N/A — backend only

**Impact on existing data:**
- Subscription record updated with new plan slug, features, limits
- User type changes from `api` to `standard`
- All existing data (social accounts, content, team members) preserved

**Impact on other products:**
- Frontend billing page will call these endpoints (separate FE story)
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Create API-Centric plan type with feature flags and limits in the subscription system**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — N/A, backend only
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 4: [BE] Add API usage tracking and activity logs endpoints

**Description:**
As a backend developer, I want to track API usage metrics and expose endpoints for retrieving usage statistics and activity logs so that the API Dashboard can display real-time usage data.

**Workflow:**
1. Every API request made by a user increments their usage counter
2. Each API request is logged with: endpoint, HTTP method, status code, timestamp, response time
3. API Dashboard frontend calls endpoints to retrieve: current usage vs. limits, remaining requests, rate limit status, recent activity logs

**Acceptance criteria:**
- [ ] API usage counter tracks calls per user per month, resetting on billing cycle
- [ ] `GET /api/usage` returns: `calls_used`, `calls_limit`, `calls_remaining`, `rate_limit_status`, `billing_cycle_reset_date`
- [ ] `GET /api/activity-logs` returns paginated list of recent API activity: `endpoint`, `method`, `status_code`, `timestamp`, `response_time_ms`
- [ ] Activity logs support filtering by: date range, endpoint, status code
- [ ] Activity logs are retained for at least 30 days
- [ ] Usage counter is accurate and handles concurrent requests safely (atomic increment)
- [ ] Rate limit status returns current request rate vs. allowed rate

**Mock-ups:** N/A — backend only

**Impact on existing data:**
- New collection/table for API activity logs
- Usage counters added to existing API key or subscription tracking

**Impact on other products:**
- Frontend API Dashboard will consume these endpoints (separate FE story)
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Create API-Centric plan type with feature flags and limits in the subscription system**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support — N/A, backend only
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 5: [FE] Build plan selection screen for post-signup flow

**Description:**
As a new user who signed up without a mode flag, I want to see a clear plan selection screen so that I can choose between the Full Suite and the API Plan before starting onboarding.

**Workflow:**
1. User creates their account on the signup page (email, password, business name)
2. If no `?mode=api` or `?mode=suite` flag was present, the system shows the Plan Selection screen as the very first step
3. User sees two visually polished cards side by side:
   - **Card 1 — Full Suite:** Title "Full Suite", tagline "The complete social media management platform", description listing key features (AI Studio, planner, analytics, inbox, discovery), CTA button "Get Started with Full Suite"
   - **Card 2 — API Plan:** Title "API Plan", tagline "Build automations and publish via API", description highlighting API access and integrations, CTA button "Get Started with API"
4. User clicks one of the two CTAs
5. System updates the account with the selected plan type and routes to the appropriate onboarding flow
6. If user came with `?mode=api` or `?mode=suite`, this screen is skipped entirely

**Acceptance criteria:**
- [ ] Plan selection screen appears as the first step after signup when no mode flag is present
- [ ] Two cards displayed side by side with clear visual differentiation
- [ ] Full Suite card shows: title "Full Suite", tagline "The complete social media management platform", feature list, CTA "Get Started with Full Suite"
- [ ] API Plan card shows: title "API Plan", tagline "Build automations and publish via API", feature list, CTA "Get Started with API"
- [ ] Clicking "Get Started with Full Suite" routes to existing full-suite onboarding
- [ ] Clicking "Get Started with API" routes to shortened API onboarding (3 steps)
- [ ] Screen is skipped when `?mode=api` flag present (user goes directly to API onboarding)
- [ ] Screen is skipped when `?mode=suite` flag present (user goes directly to full-suite onboarding)
- [ ] Page is responsive — cards stack vertically on mobile

**UI Copy:**

Page heading: "How would you like to use ContentStudio?"
Page description: "Choose the experience that fits your workflow. You can always switch later."

Card 1 — Full Suite:
- Title: "Full Suite"
- Tagline: "The complete social media management platform"
- Description: "Plan, create, schedule, analyze, and engage across all your social channels. Includes AI Studio, content planner, analytics, social inbox, discovery, and more."
- CTA: "Get Started with Full Suite" (use `Button` component, primary variant)

Card 2 — API Plan:
- Title: "API Plan"
- Tagline: "Build automations and publish via API"
- Description: "Access publishing, content management, and integrations through the ContentStudio API. Connect your tools, automate workflows, and manage everything programmatically."
- CTA: "Get Started with API" (use `Button` component, secondary variant)

**Components:** Use `Button` from `@contentstudio/ui` for CTAs. Cards are custom layout — no existing card component in the catalog for this purpose; use Tailwind layout with `bg-white`, `border-gray-200`, `rounded-lg`, `shadow-sm` styling, `hover:border-primary-cs-200` for hover state.

**Mock-ups:** See PRD section 7 — Design team to create high-quality mockups for this screen.

**Impact on existing data:** None — new screen in the onboarding flow.

**Impact on other products:**
- Existing full-suite onboarding flow is unchanged — this screen is inserted before it
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Handle signup mode flag routing and API user account creation**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 6: [FE] Build shortened 3-step onboarding flow for API users

**Description:**
As a new API Plan user, I want a shortened onboarding flow that collects only the information I need so that I can get to my API Dashboard quickly without irrelevant steps.

**Workflow:**
1. After selecting the API Plan (or arriving via `?mode=api`), user enters a 3-step onboarding flow
2. **Step 1 — Profile:** User sees heading "Set up your profile" with fields for full name, phone number, and time zone. User fills in the fields and clicks "Continue"
3. **Step 2 — Role:** User sees heading "What best describes you?" with the existing role selection options. User selects a role and clicks "Continue"
4. **Step 3 — Social Accounts:** User sees heading "Connect your social accounts" with options to connect Facebook, Instagram, LinkedIn, X, YouTube, TikTok, Pinterest, Threads. User connects accounts or clicks "Skip for now"
5. After completing Step 3 (or skipping), user is redirected to the API Dashboard

**Acceptance criteria:**
- [ ] API onboarding has exactly 3 steps: Profile → Role → Social Accounts
- [ ] Step 1 shows fields: Full Name (required), Phone Number (required), Time Zone (required with auto-detect default)
- [ ] Step 2 shows existing ContentStudio role options (same as full-suite onboarding)
- [ ] Step 3 shows social platform connection buttons for: Facebook, Instagram, LinkedIn, X, YouTube, TikTok, Pinterest, Threads
- [ ] Step 3 has a "Skip for now" option that allows skipping without connecting any accounts
- [ ] After completing Step 3, user is redirected to the API Dashboard (not Home)
- [ ] Progress indicator shows 3 steps (not the full-suite step count)
- [ ] Steps that exist in full-suite onboarding but are irrelevant for API users are skipped (e.g., feature tour, discovery prompts, Getting Started widget)
- [ ] All form validation works: required fields show errors, phone number format validated

**UI Copy:**

Step 1:
- Heading: "Set up your profile"
- Description: "Tell us a little about yourself so we can personalize your experience."
- Full Name field — Label: "Full Name", Placeholder: "e.g., Sarah Johnson"
- Phone Number field — Label: "Phone Number", Placeholder: "+1 (555) 123-4567"
- Time Zone field — Label: "Time Zone", Placeholder: auto-detected value
- Validation — Full Name empty: "Please enter your full name", Phone empty: "Please enter your phone number"
- CTA: "Continue" (use `Button` component, primary variant)

Step 2:
- Heading: "What best describes you?"
- Description: "This helps us tailor your experience."
- Role options: existing role cards (same as full-suite onboarding)
- CTA: "Continue" (use `Button` component, primary variant)

Step 3:
- Heading: "Connect your social accounts"
- Description: "Connect accounts to start publishing through the API."
- Skip text: "You can skip this step and connect accounts later."
- Primary CTA: "Finish Setup" (use `Button` component, primary variant)
- Secondary CTA: "Skip for now" (use `Button` component, ghost variant)

**Components:** Use `TextInput` for name/phone, `Dropdown` for time zone, `Button` for CTAs, `CstAccountCheckBox` for social account selection (legacy component — no `@contentstudio/ui` equivalent for social account cards).

**Mock-ups:** See PRD section 7 — Design team to create mockups.

**Impact on existing data:** None — reuses existing onboarding data models.

**Impact on other products:**
- Existing full-suite onboarding is unchanged
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Handle signup mode flag routing and API user account creation** and **[FE] Build plan selection screen for post-signup flow**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 7: [FE] Build API Dashboard page with credentials, usage, and activity logs

**Description:**
As an API Plan user, I want a dedicated API Dashboard that shows my API credentials, usage metrics, and activity logs so that I can manage and monitor my API integrations from one place.

**Workflow:**
1. User logs in and lands on the API Dashboard (default page for API Plan users)
2. On first visit, user sees a welcome message with prompt to generate their API key
3. User clicks "Generate API Key" — key and secret are displayed
4. User sees three main sections:
   - **API Credentials:** Shows API Key (masked by default, click to reveal) and API Secret. Buttons to generate new key and revoke existing
   - **API Usage:** Progress bar showing calls used / total limit, remaining requests count, rate limit status indicator
   - **Activity Logs:** Table of recent API requests with columns: Request, Endpoint, Status, Timestamp
5. User can click quick links to access: API Documentation, Integration Guides, Test API Endpoint

**Acceptance criteria:**
- [ ] API Dashboard is the default landing page for API Plan users after login
- [ ] Welcome message shows on first visit only: heading, description, "Generate API Key" CTA, "View API Documentation" link
- [ ] Welcome message dismisses after first API key is generated
- [ ] API Credentials section shows key (masked with dots, click/toggle to reveal) and secret
- [ ] "Generate New Key" button creates a new API key
- [ ] "Revoke Key" button revokes the current key with a confirmation dialog
- [ ] API Usage section shows: progress bar (calls used / total), remaining count, rate limit status (green/yellow/red indicator)
- [ ] Activity Logs table shows: request method + description, endpoint path, status code (color-coded: 2xx green, 4xx yellow, 5xx red), timestamp
- [ ] Activity Logs show last 20 entries by default with "View All" link
- [ ] Quick Links section shows: "View API Documentation", "View Integration Guides", "Test API Endpoint"
- [ ] Page is accessible from the "API" item in the top navigation bar

**UI Copy:**

Welcome message (first visit only):
- Heading: "Welcome to ContentStudio API"
- Description: "Your workspace is ready. Generate your API key to start publishing and automating workflows."
- Primary CTA: "Generate API Key" (use `Button` component, primary variant)
- Secondary link: "View API Documentation" (text link, `text-primary-cs-500`)

API Credentials section:
- Section title: "API Credentials"
- Key label: "API Key"
- Secret label: "API Secret"
- Reveal toggle tooltip: "Click to show your API key. Keep it secure — never share it publicly."
- "Generate New Key" button (use `Button` component, secondary variant)
- "Revoke Key" button (use `Button` component, ghost variant with `text-red-600`)
- Revoke confirmation dialog: Title: "Revoke API Key?", Body: "This will immediately disable your current API key. Any integrations using this key will stop working. You can generate a new key afterward.", CTA: "Revoke Key" / "Cancel" (use `Dialog` component)

API Usage section:
- Section title: "API Usage"
- Progress label: "{used} / {limit} API calls this month"
- Remaining: "{remaining} requests remaining"
- Rate limit: "Rate limit: {current_rate} / {max_rate} requests per minute"
- Status indicator: green dot = "Healthy", yellow dot = "Approaching limit", red dot = "Limit reached"
- Warning at 80%: "You've used 80% of your monthly API calls. Need more? Increase your limit." (use `Alert` component, warning variant)

Activity Logs section:
- Section title: "Recent Activity"
- Table headers: "Request", "Endpoint", "Status", "Timestamp"
- Empty state — Heading: "No API activity yet", Subtext: "Once you start making API requests, they'll appear here. Generate your API key to get started.", CTA: none (use existing empty state pattern)
- "View All" link: "View all activity →" (text link, `text-primary-cs-500`)

Quick Links section:
- Section title: "Quick Links"
- Link 1: "View API Documentation" with external link icon
- Link 2: "View Integration Guides (Zapier, Make, Pabbly, n8n)" with external link icon
- Link 3: "Test API Endpoint" with external link icon

**Components:** Use `Button`, `Dialog`, `Alert`, `Progress` from `@contentstudio/ui`. Use `Switch` for key visibility toggle. Activity logs table uses standard HTML table with Tailwind styling (no dedicated table component in catalog). Status indicators use Tailwind utility classes (`bg-green-500`, `bg-yellow-500`, `bg-red-500` for dots — these are semantic, not primary-cs theme colors).

**Mock-ups:** See PRD section 7 — Design team to create API Dashboard layout.

**Impact on existing data:** None — new page consuming new API endpoints.

**Impact on other products:**
- Settings → API Key page still exists for consistency (both API and full-suite users)
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Create API-Centric plan type with feature flags and limits in the subscription system** and **[BE] Add API usage tracking and activity logs endpoints**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 8: [FE] Implement reduced navigation and module gating for API Plan users

**Description:**
As an API Plan user, I want to see only the modules relevant to my workflow in the navigation so that I'm not overwhelmed by features I don't have access to.

**Workflow:**
1. API Plan user logs in
2. Top navigation bar shows only: Publisher, API, Content Library, Notifications, Settings
3. Home, AI Studio, Analytics, Social Inbox, and Discover are hidden from navigation
4. In Publisher sidebar: AI Library section, Automations section are hidden. Compose dropdown does not show "Blog Post" or "Bulk Schedule via AI" options
5. In Settings: White Label, SSO, Brand Knowledge, Blogs & Websites, Other Integrations are hidden. Basic Settings hides "Instagram Posting Method For Automation" and "Enable Onboarding Widget". Miscellaneous shows only Campaigns & Labels
6. Getting Started widget is hidden
7. Brand Knowledge shortcut in top bar is hidden
8. If user navigates directly to a hidden module URL (e.g., `/analytics`), they are redirected to the API Dashboard with a toast notification

**Acceptance criteria:**
- [ ] Top nav for API users shows: Publisher, API, Content Library, Notifications, Settings
- [ ] Top nav hides: Home, AI Studio, Analytics, Social Inbox, Discover
- [ ] Publisher sidebar hides: AI Library (AI Posts, Brand Settings), Automations (Bulk Schedule, Recycle Posts, RSS Auto-Post)
- [ ] Compose dropdown hides: "Blog Post" option, "Bulk Schedule via AI" option
- [ ] Settings → Account hides: White Label, SSO
- [ ] Settings → Workspace hides: Brand Knowledge, Blogs & Websites, Other Integrations
- [ ] Settings → Workspace → Basic Settings hides: "Instagram Posting Method For Automation", "Enable Onboarding Widget"
- [ ] Settings → Workspace → Miscellaneous shows only: Campaigns & Labels (all other options hidden)
- [ ] Getting Started widget does not appear for API users
- [ ] Brand Knowledge shortcut in top bar is hidden for API users
- [ ] "Explore Full Suite" CTA button appears in the top bar (implemented in a separate story)
- [ ] Direct URL navigation to hidden modules redirects to API Dashboard
- [ ] Toast notification on redirect: "This feature is not included in your API Plan. Click 'Explore Full Suite' to preview it."
- [ ] Feature checks use `useFeatures()` composable — no hardcoded plan slug checks
- [ ] Full-suite users see no changes to their navigation

**UI Copy:**
- Redirect toast: "This feature is not included in your API Plan. Click 'Explore Full Suite' to preview it." (use `CstToast` component)

**Components:** Uses existing `TopHeaderBar.vue` with `useFeatures()` composable for conditional rendering. No new components needed — this story modifies visibility logic in existing components.

**Mock-ups:** N/A — follows existing navigation patterns with items hidden.

**Impact on existing data:** None — purely visibility changes.

**Impact on other products:**
- Full-suite users: no changes
- Chrome extension: no impact (uses its own navigation)
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Create API-Centric plan type with feature flags and limits in the subscription system**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Story 9: [FE] Build Explore Full Suite preview mode with locked overlays

**Description:**
As an API Plan user, I want to preview the full ContentStudio suite so that I can evaluate whether upgrading is worthwhile — without leaving my current plan.

**Workflow:**
1. User sees a persistent "Explore Full Suite" button in the top bar
2. User clicks the button
3. Navigation expands: previously hidden items (Home, AI Studio, Analytics, Social Inbox, Discover) become visible
4. Top bar CTA changes to "Exit Preview | Back to API Dashboard"
5. User clicks on a preview module (e.g., Analytics)
6. The module page loads but displays a locked overlay/banner:
   - Heading: "This is a preview of Analytics"
   - Body: "Upgrade to a full-suite plan to unlock Analytics and start tracking your performance across all platforms."
   - Primary CTA: "Upgrade Plan" (takes user to billing page)
   - Secondary CTA: "Go Back" (returns to API Dashboard)
7. User can browse any preview module — all show locked overlays
8. User clicks "Exit Preview" to return to normal API navigation (hidden items disappear again)
9. User can also exit by clicking any of their active API modules (Publisher, API, Content Library, Settings)

**Acceptance criteria:**
- [ ] "Explore Full Suite" button visible in top bar for API Plan users only
- [ ] Clicking button reveals hidden nav items: Home, AI Studio, Analytics, Social Inbox, Discover
- [ ] Top bar CTA changes to "Exit Preview" and "Back to API Dashboard" while in preview mode
- [ ] Navigating to any preview module shows a locked overlay/banner covering the module content
- [ ] Overlay is non-interactive — user cannot click through to module content
- [ ] Overlay shows module-specific heading (e.g., "This is a preview of Analytics", "This is a preview of AI Studio")
- [ ] "Upgrade Plan" CTA navigates to Settings → Billing & Plan
- [ ] "Go Back" CTA navigates to API Dashboard
- [ ] Clicking "Exit Preview" returns to normal API navigation (hidden items removed)
- [ ] Navigating to an active API module (Publisher, Content Library, etc.) also exits preview mode
- [ ] Preview mode state is not persisted across page refreshes (refreshing exits preview mode)
- [ ] Full-suite users never see the "Explore Full Suite" button

**UI Copy:**

Top bar button (normal): "Explore Full Suite" (use `Button` component, secondary variant, with icon)
Top bar button (preview mode): "Exit Preview" (use `Button` component, ghost variant) and "Back to API Dashboard" (text link)

Locked overlay (per module):
- Heading: "This is a preview of {Module Name}" (e.g., "This is a preview of Analytics")
- Body copy per module:
  - Analytics: "Upgrade to a full-suite plan to unlock Analytics and start tracking your performance across all platforms."
  - AI Studio: "Upgrade to a full-suite plan to unlock AI Studio and create content with AI-powered tools."
  - Social Inbox: "Upgrade to a full-suite plan to unlock Social Inbox and manage all your conversations in one place."
  - Discover: "Upgrade to a full-suite plan to unlock Discover and find trending content for your audience."
  - Home: "Upgrade to a full-suite plan to unlock the Home dashboard with AI chat, quick shortcuts, and content suggestions."
- Primary CTA: "Upgrade Plan" (use `Button` component, primary variant)
- Secondary CTA: "Go Back" (use `Button` component, ghost variant)

**Components:** Use `Button` from `@contentstudio/ui` for CTAs. Locked overlay is a new component — a semi-transparent overlay (`bg-white/80 backdrop-blur-sm`) with centered content card. No existing overlay component in the catalog. *Requires new component: LockedModuleOverlay — a full-page overlay with heading, body text, and CTA buttons. Not currently in `@contentstudio/ui` — can be built as a local component in the shared module.*

**Mock-ups:** See PRD section 7 — Design team to create preview mode locked overlay mockups.

**Impact on existing data:** None — purely visual/state change.

**Impact on other products:**
- Full-suite users: no changes
- Mobile apps: no impact

**Dependencies:** Depends on **[FE] Implement reduced navigation and module gating for API Plan users**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 10: [FE] Build billing page and plan switching flow for API Plan users

**Description:**
As an API Plan user, I want to see my plan details on the billing page and be able to switch to a full-suite plan so that I can upgrade when I'm ready for more features.

**Workflow:**
1. User navigates to Settings → Billing & Plan
2. Billing page shows: current plan (API Plan), trial status (if applicable), renewal date, API calls limit and usage, social accounts limit and usage, API keys limit and usage
3. User clicks "Increase Limits" to purchase add-ons (social accounts, API calls, etc.)
4. User clicks "Change Plan" to switch to a full-suite plan
5. Modal appears showing two plan cards: Advanced ($69/mo) and Agency Unlimited ($139/mo). Standard is not shown.
6. User selects a plan
7. Pricing breakdown modal appears: comparison table showing current usage vs. target plan base limits, auto-calculated add-ons for any excess, base price + add-on total = total monthly price
8. User clicks "Confirm Switch" — plan is upgraded, navigation unlocks, user sees full-suite experience

**Acceptance criteria:**
- [ ] Billing page shows API Plan details: plan name, trial/paid status, renewal date
- [ ] Usage section shows: API calls (used/limit with progress bar), social accounts (used/limit), API keys (used/limit)
- [ ] "Increase Limits" button is present and functional
- [ ] "Change Plan" button opens a modal with plan cards
- [ ] Only Advanced and Agency Unlimited plan cards are shown (no Standard)
- [ ] Each plan card shows: plan name, price, base features, base limits
- [ ] Selecting a plan opens pricing breakdown modal
- [ ] Pricing breakdown shows: comparison table (current → target), add-on items if any, base price + add-ons = total
- [ ] "Confirm Switch" button processes the upgrade and redirects to the updated billing page
- [ ] "Cancel" button closes the modal without changes
- [ ] After successful switch: navigation updates to full-suite, toast confirms "Plan upgraded successfully"
- [ ] If user's usage fits within target plan base limits, pricing breakdown shows no add-ons

**UI Copy:**

Billing page:
- Page title: "Billing & Plan"
- Current plan label: "Current Plan: API Plan"
- Trial badge (if applicable): "Trial — {X} days remaining"
- API calls: "API Calls: {used} / {limit} this month" (use `Progress` component)
- Social accounts: "Social Accounts: {used} / {limit}"
- API keys: "API Keys: {used} / {limit}"
- "Increase Limits" button (use `Button` component, secondary variant)
- "Change Plan" button (use `Button` component, primary variant)

Change Plan modal:
- Title: "Change Your Plan"
- Description: "Upgrade to unlock Analytics, AI Studio, Social Inbox, and more."
- Card 1 — Advanced: "$69/mo", key features listed
- Card 2 — Agency Unlimited: "$139/mo", key features listed
- (Use `Modal` component from `@contentstudio/ui`)

Pricing breakdown modal:
- Title: "Your Plan Summary"
- Subtitle: "Switching from API Plan to {Selected Plan}"
- Comparison table headers: "Feature", "Current", "{Plan} Base", "Add-on"
- Example row: "Social Accounts | 30 | 25 | +5 accounts"
- Base price: "{Plan} Plan: ${price}/mo"
- Add-on total: "Add-ons: ${total}/mo"
- Total: "Total: ${grand_total}/mo"
- "Confirm Switch" button (use `Button` component, primary variant)
- "Cancel" button (use `Button` component, ghost variant)
- Success toast: "Plan upgraded successfully! You now have access to the full ContentStudio suite." (use `CstToast`)

**Components:** Use `Modal`, `Button`, `Progress` from `@contentstudio/ui`. Reuse existing `BillingPlanTile.vue` for plan cards. Use `CstToast` for success notification.

**Mock-ups:** See PRD section 7

**Impact on existing data:**
- Subscription record updated on plan switch
- All existing data preserved

**Impact on other products:**
- Full-suite billing page: no changes
- Mobile apps: no impact

**Dependencies:** Depends on **[BE] Support plan switching from API Plan to full-suite with auto-calculated add-ons** and **[BE] Create API-Centric plan type with feature flags and limits in the subscription system**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 11: [FE] Update website with API Plan pricing, landing page changes, and navigation

**Description:**
As a website visitor interested in API access, I want to discover the API Plan on the ContentStudio website so that I can understand the offering and sign up through the correct funnel.

**Workflow:**
1. User visits contentstudio.io and sees a top announcement banner: "New: ContentStudio API Plan — Build automations and publish via API. Learn more →"
2. User scrolls to the product tabs section and sees a new "API" tab alongside existing tabs
3. Clicking the "API" tab shows: heading, description, and CTA linking to the API page
4. User visits the pricing page and sees the API Plan card alongside Standard, Advanced, and Agency Unlimited
5. User visits the API page (`/social-media-api`) and sees: existing content + new pricing card + comparison table + FAQ
6. User clicks "Start Free Trial" on any API Plan CTA — redirected to `app.contentstudio.io/signup?mode=api`

**Acceptance criteria:**
- [ ] Main landing page: dismissable top banner with "New: ContentStudio API Plan" copy, linking to `/social-media-api`
- [ ] Product tabs section: new "API" tab with heading, description, and CTA
- [ ] Pricing page: API Plan card alongside existing three plans, showing plan name, price (TBD), limits, feature list, "Start Free Trial" CTA
- [ ] "Start Free Trial" CTA links to `app.contentstudio.io/signup?mode=api`
- [ ] API page (`/social-media-api`): new pricing card section, API Plan vs. Full Suite comparison table, FAQ section
- [ ] FAQ answers: switching plans, API limits, free trial, social accounts
- [ ] Product menu in main navigation: "API" link pointing to `/social-media-api`
- [ ] All CTAs for API Plan include `?mode=api` parameter in signup URL

**UI Copy:**

Banner:
- Text: "New: ContentStudio API Plan — Build automations and publish via API. Learn more →"

Product tab — API:
- Tab label: "API"
- Heading: "Automate publishing and workflows with our API"
- Description: "Build custom integrations, automate content scheduling, and connect ContentStudio to your existing tools using our developer-friendly API."
- CTA: "Explore the API Plan →"

Pricing card:
- Plan name: "API Plan"
- Price: TBD/month (placeholder until business team finalizes)
- Features included: API access, publishing via API, content library, campaign management, integrations (Zapier, Make, Pabbly, n8n), workspace management, team members
- Features excluded (shown as unavailable): Analytics, AI Studio, Social Inbox, Discovery
- CTA: "Start Free Trial" (linking to `signup?mode=api`)

API page FAQ:
- Q: "Can I switch to a full plan later?" A: "Yes, you can upgrade to Advanced or Agency Unlimited at any time from your billing settings. Your connected accounts and data are preserved."
- Q: "What happens when I hit my API limits?" A: "You can purchase additional API request capacity from your billing page, or upgrade your plan for higher base limits."
- Q: "Do I get a free trial?" A: "Yes, the API Plan includes a free trial with full access to all API features."
- Q: "Can I connect social accounts?" A: "Yes, the API Plan includes social account connections so you can publish via the API to all supported platforms."

**Components:** Website uses its own CMS/component system (not `@contentstudio/ui`). Design team to provide website mockups.

**Mock-ups:** Design team to create: banner design, product tab content, pricing card, API page sections.

**Impact on existing data:** None — website content updates only.

**Impact on other products:**
- Signup page receives `?mode=api` parameter (handled by separate story)
- No impact on app functionality

**Dependencies:** None — website changes can be built independently, but should launch coordinated with the app changes.

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 12: [Design] Create mockups for API-Centric Plan UI screens

**Description:**
As a product designer, I want to create high-quality mockups for all new UI screens in the API-Centric Plan so that the frontend team has clear visual specs to implement.

**Workflow:**
1. Designer reviews the PRD and workflow design documents
2. Designer creates mockups for each new screen and state
3. Mockups are shared in Figma and linked to the relevant stories

**Acceptance criteria:**
- [ ] Plan selection screen: two cards (Full Suite vs. API Plan) with clear visual differentiation
- [ ] API onboarding: 3-step flow (Profile → Role → Social Accounts) with progress indicator
- [ ] API Dashboard: credentials section, usage section with progress bars, activity logs table, quick links, welcome message (first visit)
- [ ] API Dashboard empty state: no key generated yet
- [ ] Reduced navigation: top bar showing API plan nav items only
- [ ] Preview mode: "Explore Full Suite" CTA, expanded navigation, locked module overlay
- [ ] Locked module overlay: semi-transparent overlay with heading, body text, upgrade CTA, go back button — one example per module (Analytics, AI Studio, Inbox, Discover, Home)
- [ ] Billing page: API Plan details, usage display, Change Plan modal, pricing breakdown modal
- [ ] Website: banner, product tab, pricing card, API page sections
- [ ] All mockups follow ContentStudio's existing design system and use `@contentstudio/ui` patterns
- [ ] Responsive variants provided for key screens (plan selection, API Dashboard)

**Mock-ups:** This story IS the mockup deliverable.

**Impact on existing data:** None.

**Impact on other products:** None — design deliverable only.

**Dependencies:** None — can start immediately from PRD and workflow documents.

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — designs include responsive variants
- [ ] Multilingual support — designs account for text expansion in other languages
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
