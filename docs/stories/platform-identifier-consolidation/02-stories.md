# Epic + stories — Platform identifier consolidation

**Scope of this doc:** the epic and **4 stories**: two backend, one frontend, one Flutter. No `[Design]` story, because there is no new or changed UI. Nothing is pushed to any tracker. The PO creates this by hand.

---

## Epic: Platform identifier consolidation

Every social account connected to ContentStudio now carries a single identifier field, the platform identifier, and it is populated and kept in sync for every platform we support. That was the whole point of introducing it: one field that identifies an account, whatever network it belongs to.

The problem is that we introduced it without retiring what came before. Facebook accounts are still looked up by a Facebook-specific key, Instagram by an Instagram-specific one, Twitter, LinkedIn and Pinterest each by their own, and Google Business Profile and Tumblr accounts by their name. Newer platforms use the platform identifier. So both systems are live at once, and every part of the product that touches a social account has to first work out which field to read for which network. Both the backend and the web app keep their own lookup table for exactly that, and the mobile app has the same mixture in its account parsing.

The cost shows up in three ways. Adding a new social network means updating a lookup table in three separate codebases before anything works. When one side reads the platform identifier and the other reads a network-specific key, a match that should succeed quietly fails, and the user sees an account that appears to be missing rather than an error anyone can act on. And nothing prevents new code from reintroducing an old key, so the situation slowly gets worse rather than better.

There is a related problem on the backend worth fixing in the same pass. All social accounts live in one place, but the backend still fetches them one network at a time, so a single screen that shows a mixed set of accounts triggers a separate query per network instead of one. That is a lot of repeated work for the composer, the planner, analytics and the inbox channel list.

This epic makes the platform identifier the one way an account is identified across the backend, the web app and the mobile app, collapses the repeated per-network account queries into one, and puts a guard in place so old keys cannot come back. This is a refactor: no screen, no wording and no user-visible behavior should change, and proving that is a large part of the work.

### Out of scope

- Removing legacy identifier fields from stored data. Code stops reading them in this epic. Whether the stored fields are then removed is a separate decision with its own migration.
- Changing any documented public API or webhook response field. If a legacy key appears in a published contract, it stays until that contract is versioned.
- Any change to how accounts are connected or authorized.

### Stories

1. `[BE] Make the platform identifier the single way a social account is identified`
2. `[BE] Fetch a workspace's social accounts in one query instead of one per network`
3. `[FE] Read social accounts by platform identifier only and retire the per-network key lookup`
4. `[Flutter] Read social accounts by platform identifier only in the mobile app`

### Sequencing

The backend identifier story goes first, because the frontend and mobile stories depend on the platform identifier being reliably present and correct in every response. The query consolidation story is independent and can run in parallel. The frontend and Flutter stories can then run in parallel with each other.

---

## [BE] Make the platform identifier the single way a social account is identified

### Description

As the engineering team, we want the backend to identify every social account by its platform identifier and nothing else, so that adding a new social network no longer requires teaching the product which field to read for it, and so account lookups stop failing silently when two parts of the product disagree about which key identifies an account.

### Workflow

1. The team produces a complete list of every place the backend identifies a social account by a network-specific key rather than the platform identifier, including the configuration that lists which field belongs to which network.
2. For every network still using a legacy key, the team confirms the platform identifier is present, correct and stable on every existing account of that network, and fixes any account where it is not before switching reads.
3. All reads are switched to the platform identifier, going through one shared way of asking "which account is this", so no caller needs to know the network to identify an account.
4. Where a network's situation is genuinely not a rename, such as a Pinterest board versus a Pinterest profile, the team records what each identifier means and confirms the switch preserves that distinction.
5. Legacy keys keep being written for now, so nothing that still reads them breaks mid-rollout.
6. A guard is added that fails the build or the test suite when a legacy identifier key is reintroduced into a read path.
7. The team verifies every feature that resolves a social account still behaves identically, network by network.

### Acceptance criteria

- [ ] A documented inventory lists every backend read path that identifies a social account, and states which identifier each one used before the change.
- [ ] Every social account of every supported network has a populated platform identifier, verified against production data before reads are switched, with any gap fixed or explicitly recorded as an exception.
- [ ] All backend read paths identify social accounts by the platform identifier, and none reads a network-specific identifier key.
- [ ] Identifying an account goes through one shared code path, so a caller does not need to know the network to resolve an account.
- [ ] The configuration that maps a network to its identifier field is no longer needed for identification, or contains the same field for every network.
- [ ] Networks where the legacy identifier is not a straight rename are documented individually, and their behavior is unchanged after the switch. This explicitly covers Pinterest boards versus Pinterest profiles, and Google Business Profile and Tumblr accounts identified by name.
- [ ] Legacy identifier keys are still written to new and updated accounts, so anything still reading them keeps working during rollout.
- [ ] Adding a new social network requires no new entry in any network-to-identifier mapping on the backend.
- [ ] A guard fails automatically when a network-specific identifier key is reintroduced into a read path.
- [ ] No documented public API or webhook response loses or renames a field.
- [ ] Every feature that resolves a social account is verified unchanged for every supported network: connecting and reconnecting an account, the composer account picker, publishing and republishing, the planner, approvals, analytics dashboards, shared analytics report links, PDF reports, social listening, the social inbox, and automations.
- [ ] Accounts in an invalid, expired or disconnected state still resolve and still display the same state as before.
- [ ] Test coverage exists for account resolution per supported network, so a future network cannot be added without it.

### Mock-ups

None. Backend-only refactor with no user-visible change.

### Impact on existing data

No data is deleted or migrated by this story. Legacy identifier keys stay on stored documents and keep being written. Where an existing account is found to be missing a platform identifier, that value is backfilled, and the story records how many accounts needed it. Whether the legacy fields are later removed from stored data is a separate decision outside this epic.

### Impact on other products

The web app and the mobile app both read social account responses, and both currently mix identifier keys. They are covered by `[FE] Read social accounts by platform identifier only and retire the per-network key lookup` and `[Flutter] Read social accounts by platform identifier only in the mobile app`. Because legacy keys keep being written, neither is blocked by this story shipping first. Integrations using the public API must see no change at all.

### Dependencies

- None. This story unblocks the frontend and Flutter stories in this epic.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing copy changes
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (white-label customers resolve accounts through the same paths)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Fetch a workspace's social accounts in one query instead of one per network

### Description

As a user opening any screen that lists my connected accounts, I want it to load quickly and consistently, so that the composer, the planner, analytics and the inbox do not slow down as I connect more networks. Today the backend fetches accounts one network at a time even though they all live in the same place, so a single screen triggers a separate query for every network we support.

### Workflow

1. The team lists every place the backend fetches social accounts for a workspace, and records how many separate queries each of those places currently makes.
2. Those are replaced with a single fetch that returns the workspace's accounts and filters by network only where a caller genuinely needs one network.
3. The team measures the number of queries and the response time for the heaviest screens before and after, and records both.
4. The team verifies the accounts returned are identical to before: same accounts, same order, same states, for a workspace with accounts on every supported network.

### Acceptance criteria

- [ ] A documented inventory lists every backend path that fetches a workspace's social accounts and the number of separate queries each made before the change.
- [ ] Fetching a workspace's social accounts issues one query, not one per network.
- [ ] Callers that genuinely need a single network's accounts can still ask for exactly that, without triggering a fetch of everything.
- [ ] For a workspace with accounts connected on every supported network, the set of accounts returned is identical to before the change: same accounts, same fields, same order, same connection states.
- [ ] Query count and response time are measured before and after for the composer account picker, the planner, the analytics overview and the inbox channel list, and both figures are recorded in the story.
- [ ] Response time for those screens does not regress, and the recorded query count is lower.
- [ ] Accounts in an invalid, expired or disconnected state are returned with the same state as before.
- [ ] Workspaces with no connected accounts still return an empty result and the affected screens still show their existing empty state.
- [ ] A workspace with a large number of connected accounts is included in verification, not only a small test workspace.

### Mock-ups

None. Backend-only refactor with no user-visible change.

### Impact on existing data

None. No data is created, changed, migrated or deleted. Only how accounts are read changes.

### Impact on other products

The mobile app and integrations read the same account responses. Because the returned data is unchanged, they should be unaffected, and verification must confirm that rather than assume it.

### Dependencies

- Independent of the rest of this epic, and can run in parallel with `[BE] Make the platform identifier the single way a social account is identified`.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing copy changes
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [FE] Read social accounts by platform identifier only and retire the per-network key lookup

### Description

As the engineering team, we want the web app to identify every social account by its platform identifier alone, so that account matching stops depending on a per-network lookup table that has to be updated for every new network and that silently fails when it disagrees with the backend.

### Workflow

1. The team produces a complete list of every place in the web app that identifies a social account by a network-specific key, including the shared channel definition that maps each network to its identifier field.
2. Every one of those is switched to the platform identifier.
3. The per-network identifier field is removed from the shared channel definition, so it stops being a lookup table.
4. Account matching goes through one shared helper, so no component identifies an account by itself.
5. A guard is added that fails when a network-specific identifier key is reintroduced.
6. The team walks the full product surface where accounts are selected, displayed or filtered, network by network, and confirms nothing changed for the user.

### Acceptance criteria

- [ ] A documented inventory lists every place in the web app that identified a social account by a network-specific key before the change.
- [ ] Every account lookup, comparison, filter and selection in the web app uses the platform identifier.
- [ ] The shared channel definition no longer carries a per-network identifier field.
- [ ] Account matching goes through one shared helper, and no view or component identifies an account independently.
- [ ] Adding a new social network requires no new entry in any network-to-identifier mapping in the web app.
- [ ] A guard fails automatically when a network-specific identifier key is reintroduced.
- [ ] No UI copy, label, tooltip, empty state or error message changes anywhere in this story. This is a refactor behind existing screens, and any wording change is a defect.
- [ ] Every surface where accounts are selected, displayed or filtered is verified unchanged for every supported network: the composer account picker and platform tabs, per-account previews and customization, the planner and its filters, custom views, approvals, all analytics dashboards and their account pickers, competitor sections, shared analytics report links, downloaded PDF reports, social listening, the social inbox channel list and filters, automations, content categories, and the account management and connection screens.
- [ ] Accounts in an invalid, expired or disconnected state show the same warning and the same reconnect path as before.
- [ ] Selecting the same account twice, or selecting accounts across several networks at once, behaves as before.
- [ ] A workspace with accounts on every supported network is used for verification, plus a workspace with none, plus a workspace with a large number of accounts.
- [ ] Existing automated tests are updated to the platform identifier rather than left asserting legacy keys, and account matching has test coverage per supported network.

### Mock-ups

None. No new or changed UI. This story deliberately introduces no visual change.

### Impact on existing data

None. No data is created, changed, migrated or deleted.

### Impact on other products

None from this story. The mobile app is covered separately by `[Flutter] Read social accounts by platform identifier only in the mobile app`.

### Dependencies

- Depends on `[BE] Make the platform identifier the single way a social account is identified`, so the platform identifier is confirmed present and correct on every account before the web app relies on it exclusively.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (verify the affected screens on small viewports, since account pickers and filters are touched across the product)
- [ ] Multilingual support (no copy changes, but verify no translation key is dropped by the refactor)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [Flutter] Read social accounts by platform identifier only in the mobile app

### Description

As a user of the ContentStudio mobile app, I want my connected accounts to appear and behave consistently everywhere in the app, so that an account never goes missing from a picker or a list because the app and the backend disagree about how that account is identified.

### Workflow

1. The team produces a list of every place the mobile app identifies a social account by a network-specific key, across account parsing, the composer, the planner, the inbox and push notification handling.
2. Every one of those is switched to the platform identifier.
3. The team verifies each affected area of the app on both platforms with accounts connected on every supported network.

### Acceptance criteria

- [ ] A documented inventory lists every place in the mobile app that identified a social account by a network-specific key before the change.
- [ ] Every account lookup, comparison and selection in the app uses the platform identifier.
- [ ] Account parsing has one shared way of reading an account's identifier, so no feature module identifies an account independently.
- [ ] Adding a new social network requires no new entry in any network-to-identifier mapping in the app.
- [ ] No UI copy or wording changes anywhere in this story. This is a refactor and any wording change is a defect.
- [ ] Every affected area is verified unchanged for every supported network: connecting and reconnecting accounts, the social channels list, the composer account selection and per-account previews, the planner and its filters, the social inbox channel list, approvals, and the publish reminder notification flow.
- [ ] Accounts in an invalid, expired or disconnected state show the same state and the same reconnect path as before.
- [ ] Verified on both iOS and Android from the single codebase, on a workspace with accounts on every supported network and on a workspace with none.
- [ ] Existing automated tests are updated to the platform identifier rather than left asserting legacy keys.

### Mock-ups

None. No new or changed UI.

### Impact on existing data

None on the server. Any locally cached account data in the app must still load, or be discarded and refetched cleanly rather than leaving the user with a broken account list after updating the app.

### Impact on other products

None. Web is covered separately by `[FE] Read social accounts by platform identifier only and retire the per-network key lookup`.

### Dependencies

- Depends on `[BE] Make the platform identifier the single way a social account is identified`.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (verify on small phone and tablet layouts)
- [ ] Multilingual support (no copy changes, but verify no translation key is dropped)
- [ ] UI theming support (default + white-label)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
