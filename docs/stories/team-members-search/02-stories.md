# Story: Search on the workspace Team Members page

## [FE] Add member search to the workspace Team page (Settings, Team)

### Description

The workspace Team page (Settings, Team) lists team members but has no way to search them. When a workspace has many members, finding one is slow. Add a search box that filters members as you type, matching the search already on the Social accounts settings page.

### Workflow

1. User opens Settings, Team.
2. User types a name or email into the search box.
3. The member list filters in real time to matches.
4. Clearing the search restores the full list.

### Acceptance criteria

- [ ] The workspace Team page shows a search box above the member list.
- [ ] Typing filters the list by member name and email in real time.
- [ ] Matching is case-insensitive and matches partial text.
- [ ] Clearing the search restores the full list.
- [ ] When nothing matches, an empty message is shown: "No members match your search."
- [ ] The search looks and behaves consistently with the search on the Social accounts settings page.

### UI copy

- Search placeholder: "Search" (reuse the shared search placeholder)
- No-results message: "No members match your search."

### Impact on existing data

None. Filtering only.

### Impact on other products

Web only. Mobile and Chrome extension: N/A.

### Dependencies

None.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Target page: the workspace Team page at `settings/team/`, rendered by `contentstudio-frontend/src/modules/setting/components/workspace/team/Team.vue` (confirmed to have no search today).
- Pattern to match: the search on the Social accounts settings page at `settings/social`, rendered by `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/Social.vue` (the search input may live in a child component; reuse the same `SearchInput` and filtering approach).
- Use the `SearchInput` component (see `docs/ui-components.md`).
