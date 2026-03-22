# Workflow Design — API-Centric Plan

## 1. Feature Placement

### Entry Points
- **Website:** New "API Plan" card on pricing page (`contentstudio.io/pricing`), updated API page (`contentstudio.io/social-media-api`) with pricing card, top announcement banner on landing page, new "API" tab in product tabs section
- **Signup:** Plan selection screen as the first post-signup step (if no mode flag); smart routing via `?mode=api` URL parameter skips plan selection
- **In-App (API users):** "API" becomes the primary section in left navigation, replacing Home. API Dashboard is the default landing page after onboarding
- **In-App (Full-suite users):** No change — existing experience is untouched
- **Settings:** API Key accessible from both API Dashboard and Settings → Account → API Key (consistency with full-suite experience)

### Navigation Structure (API Plan Users)

```
[Top Header Bar]
  Publisher | API | Content Library | Notifications | Settings
  [Explore Full Suite →]

[Publisher Sidebar]
  Compose (no Blog, no Bulk Schedule via AI)
  Planner > All Posts / Custom Views / Scheduled / Filters
  Calendar (Month/Week)
  Planner Settings > Content Categories / Social Accounts
  (AI Library — HIDDEN)
  (Automations — HIDDEN)

[API Dashboard]
  API Credentials
  API Usage
  Activity Logs
  Quick Links

[Settings]
  Account: Profile, Billing & Plan, Notifications, Email Status, Refer & Earn, API Key
  Workspace: Basic Settings, Social Accounts, Team Members, Content Categories, Miscellaneous (Campaigns & Labels only)
  (Hidden: White Label, SSO, Brand Knowledge, Blogs & Websites, Other Integrations)
```

---

## 2. User Flow — Happy Path

### Flow A: New User Signup (from API page on website)

1. User discovers the API Plan via the website (banner, pricing card, API page, or product tab)
2. User clicks "Start Free Trial" which links to `app.contentstudio.io/signup?mode=api`
3. User fills in email, password, and business name on the standard signup form
4. System detects the `?mode=api` flag — skips the plan selection screen
5. System creates the account with Plan = API Trial, User Type = API, business name → workspace name
6. **Onboarding Step 1 — Profile:** User enters full name, phone number, time zone
7. **Onboarding Step 2 — Role:** User selects their role from existing ContentStudio role options
8. **Onboarding Step 3 — Social Accounts:** User connects social accounts (Facebook, Instagram, LinkedIn, X, YouTube, TikTok, Pinterest, Threads) or clicks "Skip for now"
9. User lands on the API Dashboard with a welcome message (first visit only)
10. User clicks "Generate API Key" to get their API credentials
11. User begins building integrations using the API key, documentation, and quick links

### Flow B: New User Signup (direct — no flag)

1. User goes directly to `app.contentstudio.io/signup` (no mode flag)
2. User fills in email, password, and business name
3. System shows the **Plan Selection Screen** (two cards: Full Suite vs. API Plan)
4. User selects "Get Started with API" card
5. System creates account with Plan = API Trial, User Type = API
6. Onboarding continues from Step 6 of Flow A above

### Flow C: New User Signup (from full-suite pricing page)

1. User arrives at signup from full-suite pricing or general website (`?mode=suite` or no flag)
2. System skips plan selection and proceeds directly to full-suite onboarding
3. No changes from existing experience

### Flow D: API User — Daily Usage

1. User logs in and lands on the API Dashboard
2. User sees current API usage (calls used / total limit, remaining requests, rate limit status)
3. User reviews activity logs (recent API requests: endpoint, status, timestamp)
4. User manages API credentials (generate new key, revoke existing)
5. User navigates to Publisher to check scheduled posts or compose new content
6. User navigates to Content Library to manage media and text assets
7. User navigates to Settings to manage social accounts, team members, or billing

### Flow E: API User — Explore Full Suite (Preview Mode)

1. User clicks "Explore Full Suite" CTA in the top bar
2. Top bar changes: CTA becomes "Exit Preview | Back to API Dashboard"
3. Previously hidden nav items appear: Home, AI Studio, Analytics, Social Inbox, Discover
4. User clicks on any preview module (e.g., Analytics)
5. Module page loads with a **locked overlay/banner**:
   - Heading: "This is a preview of Analytics"
   - Body: "Upgrade to a full-suite plan to unlock Analytics and start tracking your performance across all platforms."
   - Primary CTA: "Upgrade Plan" (→ billing page)
   - Secondary CTA: "Go Back" (→ API Dashboard)
6. User can exit preview mode via "Exit Preview" button or by navigating to any active API module

### Flow F: API User — Plan Upgrade

1. User navigates to Settings → Billing & Plan
2. User clicks "Change Plan"
3. Modal shows two full-suite plan cards: Advanced ($69/mo) and Agency Unlimited ($139/mo). Standard is NOT shown (doesn't support add-ons).
4. User selects a plan
5. System compares current API usage against selected plan's base limits
6. **Pricing Breakdown Modal** appears:
   - Shows current limits vs. target plan base limits
   - Auto-calculates add-ons for any excess (e.g., 30 social accounts vs. 25 base = +5 add-on)
   - Displays: base plan price + add-on total = total monthly price
7. User clicks "Confirm Switch"
8. System upgrades the plan, user type changes to Standard, full navigation unlocks

---

## 3. Alternative Flows & Edge Cases

### Signup Edge Cases
- **User signs up with `?mode=api` but later wants full suite:** Can switch plans from Settings → Billing at any time. Plan selection is not shown again.
- **User signs up with `?mode=suite` but actually wants API:** Must subscribe to a full-suite plan first, then contact support or switch from billing settings (if API plan becomes available for existing users)
- **Trial expiration:** API trial expires same as full-suite trial. User is prompted to select a paid plan (API Plan or full-suite plans).

### API Dashboard Edge Cases
- **No API key generated yet:** Dashboard shows prompt to generate first key. Usage and logs sections show empty states.
- **API key revoked:** Dashboard shows "No active API key" with "Generate New Key" button.
- **Rate limit exceeded:** Banner warning: "You've reached your API rate limit. Requests will resume shortly. Need higher limits? Increase Limits."
- **API calls limit approaching:** Warning when at 80%: "You've used 80% of your monthly API calls. Increase your limit to avoid disruptions."

### Navigation Edge Cases
- **Deep link to hidden module:** If an API user navigates directly to `/analytics` or `/inbox`, redirect to API Dashboard with a toast: "Analytics is not included in your API Plan. Explore Full Suite to preview it."
- **Shared workspace with mixed users:** Each user's navigation is determined by their own plan type, not the workspace plan. If workspace has both API and full-suite users, each sees their own view.

### Billing Edge Cases
- **API user has more social accounts than target plan base:** System auto-adds add-ons. User sees exact pricing before confirming.
- **API user downgrading from full-suite back to API:** Not in v1 scope (would require removing access to analytics data, inbox history, etc.).
- **Trial user clicking "Explore Full Suite":** Preview mode works during trial. Upgrade CTAs lead to plan selection.

---

## 4. Key Design Decisions

### Decision 1: Plan Selection as Post-Signup Step vs. Pre-Signup Toggle

**Option A (Recommended — as per proposal):** Plan selection appears as the first screen after account creation, before onboarding. Two cards: Full Suite vs. API Plan.

**Option B:** Add a toggle or tab on the signup page itself (e.g., "I'm a developer" toggle).

**Recommendation: Option A.** Keeps the signup form clean and identical for all users. The plan selection screen provides more space for clear descriptions of each option. Smart routing via URL flags means most users skip this screen entirely anyway. Only users with unknown intent see it.

### Decision 2: Preview Mode Implementation — Overlay vs. Screenshot/Static

**Option A (Recommended):** Load the actual module page with a semi-transparent locked overlay/banner. Module content is visible but not interactive.

**Option B:** Show a static screenshot or marketing image of the module with an upgrade CTA.

**Recommendation: Option A.** Showing the real (empty) module with a lock overlay feels more authentic and gives users a genuine taste of the UI. It also requires less design work (no screenshots to maintain). The overlay approach is used successfully by tools like Figma (free tier viewing locked features).

### Decision 3: Standard Plan in Upgrade Options

**Decision:** Do not show Standard plan as an upgrade target for API users.

**Rationale:** Standard doesn't support add-ons. If an API user has more social accounts, workspaces, or users than Standard allows, there's no way to bridge the gap. Showing it would create a dead-end experience. Only Advanced and Agency Unlimited are offered.

---

## 5. Integration with Existing Features

### Publisher
- API users get full access to Compose (without Blog and Bulk Schedule via AI options), Planner, and Calendar
- Content scheduled via API appears in the Planner the same way as content scheduled via the UI
- Compose dropdown removes: "Blog Post" option, "Bulk Schedule via AI" option
- All other publisher features remain: custom views, filters, content categories, social account switching

### Content Library
- Full access to media management, captions, reusable text assets, templates, files
- API users can upload media via the UI and reference it in API calls
- No changes from full-suite content library experience

### Integrations
- Zapier, Make, Pabbly, n8n integrations remain available
- "Other Integrations" settings section is hidden (full-suite only)
- Integration connectors that work via API continue to function

### Notifications
- Publishing success/failure notifications
- API limit warnings (approaching/exceeded)
- Billing updates
- No inbox-related notifications (inbox is hidden)

### Team Members
- Full team management (invite, roles, permissions)
- Team members inherit the workspace's plan type for navigation
- No changes from existing team member management

---

## 6. Scope Recommendation

### V1 (This Release)
- New plan type in database with feature flags
- Signup flow with mode flag routing + plan selection screen
- Shortened 3-step onboarding for API users
- API Dashboard (credentials, usage, logs, quick links)
- Reduced navigation for API users (hide: Home, AI Studio, Analytics, Inbox, Discover)
- Publisher restrictions (remove Blog + Bulk Schedule via AI from Compose)
- Settings restrictions (hide: White Label, SSO, Brand Knowledge, Blogs & Websites, Other Integrations)
- Explore Full Suite preview mode with locked overlay
- Billing page showing API plan details + Change Plan flow (Advanced/Agency only)
- Website changes: pricing page card, API page pricing section + comparison + FAQ, landing page banner + product tab, navigation update

### V2 (Deferred)
- Multiple API keys per user/workspace
- API usage analytics dashboard (detailed charts, per-endpoint breakdown)
- Webhook management UI (configure callbacks for events)
- API rate limit customization (per-user override)
- SDK generation/download from dashboard
- Developer sandbox/testing environment
- Downgrade from full-suite back to API plan
- API-specific add-ons (extra API calls, extra keys)
- Team member role-specific API key scoping
- API changelog/status page integration
