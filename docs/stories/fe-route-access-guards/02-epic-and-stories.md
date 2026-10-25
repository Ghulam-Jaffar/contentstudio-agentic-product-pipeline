# Epic + stories: Route-level page access (roles, white-label, plan)

3 stories. Nothing is pushed to any tracker until the PO approves the push.

---

## Epic: Route-level page access for the web app

Today the web app decides who may open a page in several places at once: one global check that only knows about approvers, a handful of route checks added one page at a time, redirects written inside individual pages, and the menus, which hide pages without blocking them. Each of these has its own copy of the rules. When a page is rebuilt, its in-page check is easy to lose, and a page that is hidden in the menu stays reachable by typing its URL or pressing Back.

This epic moves every page access rule to the route, where one rule set is checked on every navigation: menu clicks, typed URLs, bookmarks, shared links, refreshes and browser Back/Forward. The menus read the same rules, so "hidden in the menu" and "blocked by URL" cannot drift apart again.

Three kinds of rules move:
1. **Role rules**: what a super admin, admin, collaborator or approver may open, including per-member permissions such as billing access, team access and workflow management.
2. **White-label rules**: ContentStudio-only pages that never show on a white-label domain.
3. **Plan and add-on rules**: pages that need a feature from the user's plan or a purchased add-on.

**Denied access behaves the way it does today:** the user is redirected, with no new message. Existing messages (the API plan notice, the white-label "no active workspace" notice) stay as they are.

### Out of scope

- Backend authorization. The backend already authorizes its own APIs. This epic only controls which pages the app opens.
- Changing who has which permission. Every role keeps the access it has today, except where a story calls out a deliberate change.
- Hiding or showing individual buttons inside a page, beyond the Manage Limits and Manage Team actions named in the first story.
- Mobile app and Chrome extension.

### Stories

1. [FE] Enforce role and white-label page access from a single route-level rule set
2. [FE] Drive menus from the route access rules and remove page-level access checks
3. [FE] Move plan and add-on page locks onto the route-level access rules

---

## [FE] Enforce role and white-label page access from a single route-level rule set

### Description

As a workspace member, I want every page I'm not allowed to open to stay closed no matter how I reach it, so that hidden pages can't be opened by typing a URL, following an old link or pressing Back.

As a developer, I want each page's access rule written once, next to the page's route, so that rebuilding a page can't drop its access check and a new page can't ship without one.

Today only approvers are checked on every navigation. Every other role rule and most white-label rules live inside individual pages or only in the menus. Recent fixes (the API page and the Usermaven page on white-label, back-button re-entry into a locked workspace) each closed one page at a time. This story replaces that with one rule set checked on every navigation, and closes the pages that are open by URL today.

### Workflow

**A member opens a page they can't access**
1. A collaborator without team access types the Team Members settings URL, or clicks an old bookmark to it.
2. The app redirects them to their profile settings. The Team Members page never appears, not even for a moment.
3. They press the browser's Back button. They do not land on Team Members again.

**A white-label user opens a ContentStudio-only page**
1. A user on a white-label domain pastes a link to Refer & Earn or Affiliate Program.
2. The app redirects them to their profile settings. Nothing ContentStudio-branded appears.

**A user opens a link to another workspace**
1. A user is working in Workspace A, where they are an admin. A teammate sends them a link to the Billing page in Workspace B, where they are a collaborator.
2. The app checks their role in Workspace B, the workspace in the link, and redirects them because collaborators can't open Billing.

**An approver opens a page outside their allowed list**
1. An approver types the Analytics URL.
2. The app redirects them to the planner list view, as it does today.

### Access rules

Every page the user reaches after logging in follows one row in this table. "Everyone" means every role, including approvers, unless the row says otherwise. Approvers keep today's short allowlist (planner views, notifications, profile, email notification status, password, workspaces, plus AI Studio and Content Library when they have the matching permission). Every page not on that list redirects an approver.

| Page | Who can open it | White-label |
|---|---|---|
| Your Profile, Notifications | Everyone | Allowed |
| Billing & Plan | Super admin, and admins given billing access | Allowed |
| Email Notification Status | Everyone | Hidden when a white-label workspace is opened on its white-label domain |
| Refer & Earn | Team members (not client members) | Hidden when a white-label workspace is opened on its white-label domain |
| Affiliate Program | Everyone except approvers | **Blocked on every white-label domain (new)** |
| White Label, SSO Domains | Super admin, and admins given billing access. Not on the API plan | Blocked on every white-label domain |
| Approval Workflows, Workflow Builder | Super admin, admins, and collaborators given workflow management | Allowed |
| Team Members | Super admin, admins, and collaborators given team access | Allowed |
| Basic Settings, Social Accounts, Content Categories, Miscellaneous | Everyone except approvers | Allowed |
| Brand Knowledge, Other Integrations | Everyone except approvers. Not on the API plan | Allowed |
| Workspaces list | Everyone | Allowed |
| Create Workspace | Unchanged: whoever can create a workspace today | Allowed |
| Manage Team (page and Workspaces tab) | **Same as Team Members (changed, see below)** | Allowed |
| Manage Limits (page and Workspaces tab) | **Members who can remove team members, social accounts or the workspace (changed, see below)** | Allowed |
| API | Everyone except approvers | Blocked on every white-label domain |
| Usermaven web analytics | Everyone with access to Analytics | Blocked on every white-label domain |
| AI Studio, Content Library | Everyone. Approvers only with create-post / Content Library permission | Allowed |
| All other workspace pages (Home, Publisher, Planner, Inbox, Analytics, Discovery, Listening, Automations) | Everyone except approvers. Plan rules are handled in "[FE] Move plan and add-on page locks onto the route-level access rules" | Allowed |

**Manage Team and Manage Limits follow the pages they manage.** Their rules must match the pages whose data they change:
- **Manage Team** opens for everyone who can open Team Members. Changing roles and removing members is available only to those who can do the same on Team Members (super admin and admins). A collaborator with team access sees the grid read-only.
- **Manage Limits** opens for anyone who can do at least one of its actions. Each action follows the page it comes from:
  - Remove a workspace: the workspace owner, and admins given the delete-workspace permission. The default workspace can't be removed, as today.
  - Remove a team member: the same people who can remove members on Team Members.
  - Remove a social account: the same people who can remove accounts on Social Accounts.
  - Remove an automation: the same people who can delete that automation on the automation pages.
- Manage Limits lists every workspace where the user can take at least one of these actions, not only workspaces they own. Otherwise an admin who is allowed in would see an empty page.

### Where a denied user lands

No message is shown, and the redirect replaces the history entry so Back does not return to the blocked page.

| Situation | Redirect to |
|---|---|
| Approver, any page outside their allowlist | Planner list view (as today) |
| Any other role, a Settings page | Your Profile settings |
| Any other role, any other page | Workspace Home |
| Manage Team or Manage Limits tab on the Workspaces page | Manage Workspaces tab |

### Acceptance criteria

**One rule set, checked on every navigation**
- [ ] Every page in the access rules table is checked on menu clicks, typed URLs, bookmarks, links from other pages, page refresh and browser Back/Forward.
- [ ] A denied user never sees the blocked page's content, not even briefly, on a cold load straight to its URL.
- [ ] The check uses the user's role in the workspace named in the URL, not the workspace they had open before.
- [ ] On a cold load, the decision waits until the user's profile, workspace and plan have loaded, so allowed users are never wrongly redirected while data is still loading.
- [ ] Each page's rule is written once, next to that page's route. No page keeps its own separate copy of the rule.
- [ ] Role and permission questions ("is this user an approver", "does this user have billing access") have one answer in the app. Today two separate copies of this logic exist. After this story there is one.
- [ ] "Is this a white-label domain" and "is a white-label workspace being shown on its domain" each have one answer in the app, used by every rule.

**Pages that are open by URL today**
- [ ] Email Notification Status redirects on a white-label workspace opened on its white-label domain.
- [ ] Refer & Earn redirects on a white-label workspace opened on its white-label domain, and for client members.
- [ ] Affiliate Program redirects on every white-label domain.
- [ ] Manage Team and Manage Limits redirect users who don't meet their rules, both as standalone pages and as tabs on the Workspaces page (for example `?tab=manage_limits`).
- [ ] Create Workspace redirects users who can't create a workspace today. Users who can create one today still can.

**Manage Team and Manage Limits**
- [ ] A collaborator with team access can open Manage Team and sees members and roles, but can't change roles or remove anyone.
- [ ] An admin can open Manage Limits and sees every workspace where they can remove at least one thing.
- [ ] On Manage Limits, the remove action for workspaces, team members, social accounts and automations appears only where the user has that permission on the matching page.
- [ ] A user with none of these permissions can't open Manage Limits and doesn't see its tab.

**Redirects**
- [ ] Denied users land on the destination in the "Where a denied user lands" table.
- [ ] No new toast or message is shown on a redirect. The existing API plan notice and white-label "no active workspace" notice are unchanged.
- [ ] After a redirect, pressing Back does not reopen the blocked page.

**Nothing else changes**
- [ ] Every role still opens every page it can open today, except the pages named above as blocked or changed.
- [ ] Approver access is identical to today.
- [ ] Trial-expired, paused-account, locked-workspace, email-verification, plan-selection, onboarding and sample-workspace redirects behave exactly as today.

**Tests**
- [ ] Automated tests cover each row of the access rules table for super admin, admin, collaborator and approver, on a normal domain and on a white-label domain.
- [ ] Tests cover cold load, browser Back after a redirect, and a link into a workspace where the user has a different role.

### Mock-ups

None. No visual change.

### Impact on existing data

None. No data or permissions are changed. The rules read the permissions members already have.

### Impact on other products

None. Mobile app and Chrome extension are not affected. White-label domains are affected as described in the access rules table.

### Dependencies

None. This is the first story of the epic.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [FE] Drive menus from the route access rules and remove page-level access checks

### Description

As a workspace member, I want the menus to show exactly the pages I can open, so that I never click something that bounces me away and never miss a page I'm allowed to use.

As a developer, I want the menus and the route to read the same access rules, and no page to carry its own access check, so that a rule changed in one place changes everywhere.

Today the settings menu rules are written three times: the expanded settings sidebar, the collapsed settings rail and the settings dropdown. The desktop navigation rail (and its mobile drawer), the workspace switcher's manage menu and the Workspaces page tabs each have their own conditions as well. Several pages also redirect from inside the page after it loads. After "[FE] Enforce role and white-label page access from a single route-level rule set" ships, those in-page checks are duplicates, and the menu copies are the next thing to drift.

### Workflow

1. A collaborator without team access opens Settings. The expanded sidebar, the collapsed rail and the settings dropdown all leave out Team Members, because the route would redirect them.
2. An admin opens the workspace switcher. Manage Team and Manage Limits appear, because the admin can open both.
3. An admin with billing access opens Settings on a white-label domain. White Label, SSO Domains and Affiliate Program are not listed anywhere. Billing & Plan is listed.
4. A developer adds a new settings page and forgets to give it an access rule. The automated check fails before the change can merge.

### Acceptance criteria

**Menus match the rules**
- [ ] For every role on a normal domain and on a white-label domain, each of these shows exactly the pages that role can open under the access rules, no more and no fewer:
  - the expanded settings sidebar
  - the collapsed settings rail
  - the settings dropdown
  - the desktop navigation rail and its mobile drawer
  - the workspace switcher's manage menu
  - the tabs on the Workspaces page
- [ ] Items that are shown with a lock and open an upgrade prompt (plan or add-on not included) keep doing so. That is a deliberate upsell, not a menu mismatch.
- [ ] The expanded sidebar and the collapsed rail never disagree about which settings items appear.
- [ ] Changing a page's access rule changes both the route and every menu that lists it, with no second edit.

**No page-level access checks**
- [ ] No page redirects the user from inside the page because of role or white-label. The route already decided before the page loaded.
- [ ] Opening an allowed page shows it directly, with no flash of a different page and no second redirect.
- [ ] The old, unused feature-lock helpers in the router are removed.

**Regression net**
- [ ] An automated check fails when any page reachable after login has no access rule.
- [ ] An automated check fails when a menu item points to a page that the menu's role can't open.

**Nothing else changes**
- [ ] Menu order, labels, icons, badges and styling are unchanged.
- [ ] Every page each role can open today still opens, and still appears in the same menu place.

### Mock-ups

None. No visual change.

### Impact on existing data

None.

### Impact on other products

None. Mobile app and Chrome extension are not affected.

### Dependencies

Depends on: **[FE] Enforce role and white-label page access from a single route-level rule set**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)

---

## [FE] Move plan and add-on page locks onto the route-level access rules

### Description

As a user whose plan doesn't include a feature, I want pages for that feature to behave the same way however I reach them, so that a locked menu item can't be bypassed by typing its URL.

As a developer, I want plan and add-on rules to live in the same route-level rule set as role and white-label rules, so that every reason a page can be closed is in one place.

Plan and add-on locks are spread out today. Listening, Looker Studio, the API plan and sample workspaces are checked on the route. Inbox, White Label, SSO Domains, Approval Workflows and the planner grid view redirect from inside the page after it loads. Content Discovery is locked in the desktop navigation rail, but no check was found on the page or its route, so a user without the feature can likely open it by URL. This needs confirming in the app.

Some pages deliberately open and show an upgrade prompt instead of redirecting, for example competitor analytics. Those are upsell surfaces and keep that behavior.

### Workflow

**Feature not in the plan, page redirects**
1. A user on a plan without Content Discovery types the Discovery URL.
2. The app redirects them to workspace Home. The page never opens, matching the locked menu item.

**Add-on not purchased, page redirects**
1. An admin with billing access, on a plan without the Approval Workflows add-on, opens an old link to the Workflow Builder.
2. The app redirects them to workspace Home, as the Approval Workflows list does today.

**Feature not in the plan, page shows an upgrade prompt**
1. A user without competitor analytics opens a competitor report URL.
2. The page opens with its upgrade prompt, exactly as today.

### Plan and add-on rules

| Page | Needs | When missing |
|---|---|---|
| Inbox | Social inbox | Redirect to workspace Home (as today) |
| Content Discovery | Content discovery | **Redirect to workspace Home (new)** |
| Listening | Listening supported and unlocked | Redirect to workspace Home, with the listening upgrade prompt (as today) |
| White Label | White-label add-on | Redirect to workspace Home (as today) |
| SSO Domains | SSO add-on | As today: redirect when the plan doesn't allow it, upgrade overlay when allowed but not yet unlocked |
| Approval Workflows, Workflow Builder | Approval workflows | Redirect to workspace Home (as today) |
| Planner grid view | Grid view planning | As today |
| Looker Studio | Looker Studio | Redirect to Analytics overview (as today) |
| Pages blocked on the API plan | Not on the API plan | Redirect to API page with the existing API plan notice (as today) |
| Sample workspace hidden sections | Not a sample workspace | As today |
| Competitor analytics, campaign and label analytics, automations | Matching feature | Page opens and shows its upgrade prompt or locked state (as today, unchanged) |

During the work, the developer confirms this list against the app and adds any other page whose menu item is locked by plan. Any page found open by URL while its menu item is locked is reported to the PO before its behavior is decided.

### Acceptance criteria

- [ ] Every page in the plan and add-on rules table behaves as listed, on menu clicks, typed URLs, bookmarks, refresh and browser Back/Forward.
- [ ] Content Discovery redirects to workspace Home when the plan doesn't include it.
- [ ] The Workflow Builder no longer misjudges add-on access on a cold load. It waits for the plan, like the Approval Workflows list.
- [ ] On a cold load, no user with the feature is redirected while the plan is still loading, and no user without it sees the page first.
- [ ] Pages that show an upgrade prompt today still open and show the same prompt.
- [ ] Each page's plan or add-on rule is written once, next to its route, alongside its role and white-label rules.
- [ ] Pages no longer redirect from inside the page because of plan or add-on.
- [ ] Locked menu items still appear with a lock and still open the upgrade prompt when clicked.
- [ ] Automated tests cover each row of the table with the feature included and not included.

### Mock-ups

None. No visual change.

### Impact on existing data

None.

### Impact on other products

None. Mobile app and Chrome extension are not affected.

### Dependencies

Depends on: **[FE] Enforce role and white-label page access from a single route-level rule set**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, nothing API-facing changes)
