# Research: "+ New" button in create-capable dropdowns

**Source:** Team meeting, 2026-09-30. Target epic: **Q3 - 2026: Miscellaneous** (`bd4d7f87-9f18-4297-a4f6-7cc63ec784e4`).

## Ask

Several dropdowns (labels, campaigns, templates) show a small circular plus icon for "create new". It is hard to see and doesn't look good. Replace it with a labelled "+ New" button, the same way the workspace dropdown was changed. The PO asked for a companion `[Design]` story even though the pattern exists: the designer finalises how it lands in each dropdown.

## Reference pattern

`contentstudio-frontend/src/components/layout/WorkspaceSwitcherDropdown.vue` (~lines 95-107): `<Button variant="light" size="xs">` with `<Icon name="Plus" :size="14"/>` and `t('header.workspace.add_new')` ("New"), tooltip `header.workspace.add_new_tooltip`. Originated in the `docs/stories/workspace-dropdown-new-pill/` story.

## In scope: dropdown headers with an icon-only `CirclePlus`

| Dropdown | File | Current tooltip key / copy | Action |
|---|---|---|---|
| Labels | `src/components/common/LabelAttachment.vue:184,193` (header), `:315` (empty state) | `common.label_attachment.add_new_label` "Add a new Label" | Opens inline create-label form (`showAddLabel`) |
| Campaigns | `src/components/common/CampaignAttachment.vue:117,126` (header), `:261` (empty state) | `common.campaign_attachment.add_new_campaign` "Add a new Campaign" | Opens inline create-campaign form (`showAddCampaign`) |
| Templates | `src/modules/composer/components/TemplateAttachment.vue:104,113` | composer mode: `composer.template_attachment.save_as_template` "Save post as template"; other mode: `create_new_template` "Create a new template" | Composer: save current post as template modal. Elsewhere: opens Composer |
| Inbox tags | `src/modules/inbox-revamp/components/TagsDropdown.vue:51,60` | `inbox.tags_dropdown.add_new_tag` "Add a new Tag" | Opens inline create-tag form |
| Planner custom views | `src/modules/publisher/components/PlannerCustomViewsSidebarDropdown.vue:13` | `planner.planner_custom_view.sidebar.create_tooltip` | Emits `create-view` |

Notes:
- Label, Campaign and Tags render a second, greyed `CirclePlus` while the inline create form is open (the `v-else` branch). The new button needs a disabled state for that.
- Tags uses a hardcoded `#8F8F8F` grey on the disabled icon; the new button should use component props, no hardcoded colours.
- Older square `Plus` buttons: `src/modules/composer/components/Labels.vue:5`, `Campaigns.vue:5`. Check whether these are still reachable; if so include them.

## Out of scope (already have visible text, or not dropdowns)

- `DocSelector.vue` (Inbox auto replies), `FolderPickerDropdown.vue` (Discovery), `HashtagSelection.vue` (Automations): the icon already sits beside a text label.
- Sidebars and nav with a `+`: `publisher/components/SidebarMain.vue`, `listening/components/nav/*`, media library `SideBar.vue`, `setting/.../ContentCategories.vue`. Not dropdowns; can be a follow-up if the PO wants consistency.

## Translations

English values live in `src/locales/en/common.json:704,723`, `composer.json:518-519`, `inbox.json:398`. The workspace story found several languages storing partial strings for `add_new`; check every supported language reads correctly as a standalone "New" button.
