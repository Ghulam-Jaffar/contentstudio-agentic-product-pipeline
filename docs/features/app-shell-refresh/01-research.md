# Research: App shell refresh (top bar, global search, Quick add)

**Source:** Team meeting, 2026-09-30. The desktop rail has run out of room for more items and quick navigation, and the sidebar looks weak. Direction: the main app sits in one card; the navigation (rail) and a new top bar are joined around it and separated from it by the card's border. The top bar holds a **global search** (jump to any feature or module) and a **Quick add** button (add social accounts, team members and similar). The PO will share competitor references and design inspiration (not finished designs). The **Health Center** was split into its own epic (`docs/features/workspace-health-center/`).

**Status:** skeleton epic. The PO asked for epics to be created now and detailed later. Layout, what moves from the rail to the top bar, and final copy all wait on the designs.

## Current shell (`contentstudio-frontend/src/`)

- Layout: `Home.vue`. Mounts `<TopHeaderBar v-if="showTopHeaderBar" @open-composer=...>` (~line 1031, computed line 373). Banners above it: `StickyBanner` (trial), `DashboardNotificationBanner`, `NoticeBanner` (expired card), `HeaderBillingNotifications`. `HeadNotificationSlider` is a fixed bottom-centre toast (~line 1000). `ComposerWidget` and `AiChatEngineProvider` wrap the `router-view` (~1050-1060).
- `components/layout/TopHeaderBar.vue` (1109 lines): on desktop renders **no top bar**, only `<DesktopNavigationRail>` (line 684). On mobile it renders a real top header with a drawer (`isMobileHeaderViewport`, ~725). Wires the Frill changelog, notifications, approvals and profile.
- `components/layout/DesktopNavigationRail.vue` (937 lines): workspace switcher at top, module list, an always-present **More** hover dropdown (562-660) with overflow and "Customize sidebar" (`data-cy="header-customize-sidebar"`, line 652), utility icons: approval notifications (699), notifications (733), theme picker (~807), profile/settings (859). `CustomizeSidebarModal` (890). Items that don't fit the height spill into More (185-196).
- Helpers: `RailNavItem.vue`, `railItemClasses.ts`, `railThemes.ts`, `useSidebarLayout.ts` (saved order and hidden modules), `CustomizeSidebarModal.vue`, `WorkspaceSwitcherDropdown.vue`, `HomeSettingsDropdown.vue`.
- Rail items (`components/layout/useHeaderNavigation.ts`): home (198), ai-studio (217), publisher (237), inbox (255), analytics (280), listening (302), media-library (329), discover (352, in More by default), api (177), more (409). Secondary group: social-accounts (375), brand-knowledge (392), which hides below 800px height. Utility ids: `approval-notifications | notifications | changelog | settings` (line 46).
- Related earlier work: `docs/stories/desktop-nav-rail-pin-unpin/` (Customize sidebar), Helpin epic "Left Sidebar Navigation" (completed 2026-05). Rail theming (`railThemes.ts`) must survive the redesign.

## Search and Quick add

- **No global search or command palette exists.** No Cmd/Ctrl+K handlers anywhere.
- **No Quick add menu.** The only create hook is the `open-composer` emit (TopHeaderBar line 66) and `ComposerWidget`.
- Existing create entry points Quick add could call: Composer (new post), Connect Social Accounts modal, invite team member (Settings > Team), create workspace (`WorkspaceSwitcherDropdown`), label/campaign create, AI Studio tools.
- Search index for v1 can be built client-side from the router and `useHeaderNavigation.ts` (modules, sub-pages, settings pages, AI Studio tools from `useAiToolCatalog.ts`), respecting permissions and plan. Content search (posts, media, conversations) would need backend and is a separate decision.

## Open questions for the PO (to settle with the designs)

- Which utilities move from the rail to the top bar: notifications, approvals, changelog, help, profile, workspace switcher?
- What the rail gains with the freed space (for example Health Center, Brand Knowledge, Social Accounts always visible)
- Search scope for v1: navigation only (modules, pages, settings, AI tools) or also recent items and content (posts, media, contacts)? Draft assumes navigation plus a few actions.
- Quick add list and order. Draft: Create post, Connect social account, Invite team member, Create workspace, Create label, Create campaign.
- Where the trial, billing and notice banners sit in the new layout
- Mobile web: keep today's mobile header and drawer, or adopt the top bar? Draft keeps mobile as is. The Flutter app is separate and not affected.
- White-label: the rail and top bar must follow white-label themes and logos.
