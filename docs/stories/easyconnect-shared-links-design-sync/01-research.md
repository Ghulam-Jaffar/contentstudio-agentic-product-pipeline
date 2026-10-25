# EasyConnect shared links design sync: research

**Request (PO, 2026-10-21):** Planner and Analytics "Manage Shared Links" panels are already synced. The EasyConnect links panel is outdated. Update it so all three match. Assign to Abdul Wasay, current sprint.

## Codebase pointers (contentstudio-frontend)

- Target design (Analytics): `src/modules/analytics/components/competitor/ManageSharedLinksDrawer.vue` + `ManageSharedLinkCard.vue` (search via `SearchInput`, virtualized list, status badge, `Switch`, URL row + Copy, chips incl. Password, hover Edit/Delete). Opened from `src/modules/analytics/views/common/AnalyticsFilterBarWrapper.vue`.
- Planner uses the same panel design, opened from `src/modules/planner_v2/components/PlannerHeader.vue`.
- Copy keys: `analytics.common.manage_shared_links.*` in `src/locales/en/analytics.json`.
- Current EasyConnect panel: `src/modules/integration/components/dialogs/SharedConnectLinkModal.vue` (`CstuModal`, legacy `CstSwitch`, hardcoded `#EEF1F3` borders, lock icon next to name, chevron expand into `ConnectionLinkedAccounts` per platform). Copy keys under `settings.integrations.social.easy_connect.*`.
- Edit flow: `src/modules/integration/components/dialogs/CreateShareLinkModal.vue` (unchanged).

## Screenshots

`screenshots/planner-manage-shared-links.png`, `screenshots/analytics-manage-shared-links.png`, `screenshots/easyconnect-current.png`.

## Notes

- No [Design] companion story: the existing Planner/Analytics panel is the design.
- EasyConnect-only element to keep: per-link connected accounts (expandable section).
