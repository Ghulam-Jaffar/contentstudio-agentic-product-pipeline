# Research: Refer & Earn page update (new affiliate program copy + affiliate dashboard state)

**Requirements source:** [Refer & Earn in-app copy, marketing](https://docs.google.com/document/d/18VUPrWFss092jpAH-DoXRoSkJlQhwznbUxhdSTAK6QU/edit?usp=sharing)

## Current State

The Refer & Earn page is a single Vue component with three runtime states, gated on whether the user has a stored FirstPromoter auth token.

**Entry points**
- Route: `refer&earn` → `contentstudio-frontend/src/modules/setting/config/routes/setting.ts:254`
- Nav: `contentstudio-frontend/src/modules/setting/components/SettingSidebar.vue:98` and `src/components/layout/HomeSettingsDropdown.vue:158`
- Visibility gate (already in place): hidden when `shouldShowWhiteLabelData` is true, and only shown when `getTeamMembership === 'team'` (account owner, not an invited team member) — `SettingSidebar.vue:96-103`
- Component: `contentstudio-frontend/src/modules/setting/components/ReferEarn.vue` (221 lines)

**State A — loading** (`ReferEarn.vue:24-31`): `Loader` while `useProfileQuery` is pending.

**State B — error** (`ReferEarn.vue:34-43`): `Alert color="danger"` when the profile fetch fails.

**State C — enrolled affiliate** (`ReferEarn.vue:52-60`): renders the **FirstPromoter dashboard in an `<iframe>`** at `https://contentstudio.firstpromoter.com//iframe?at=<auth_token>`, `height="850px"`. There is no native link card, no stats row, no tier display — everything the affiliate sees today comes from inside the FirstPromoter embed.

**State D — not enrolled** (`ReferEarn.vue:63-141`): the program pitch. Contains:
- `h1` "Earn 30% recurring commission" + description
- `Button` "Become an affiliate" → calls `becomeAffiliate()`
- 3-step strip (Join / Refer / Earn) with `Icon` circles
- `Alert color="info"` box headed **"Recurring means recurring"** with two bullets
- Footnote row with `ShieldCheck` icon: "Program eligibility and payout terms apply."

**Copy source**: all strings live under `settings.profile.refer_earn.*` in `contentstudio-frontend/src/locales/<locale>/settings.json`, present in all 8 locale dirs (`de`, `el`, `en`, `es`, `fr`, `it`, `pl`, `zh`).

**Enrollment mechanics (important)**
- `becomeAffiliate()` → `useCreatePromoterMutation` → `POST /createPromoter` (`contentstudio-backend/routes/web/accounts.php:26`)
- `AccountController::createPromoter()` (`contentstudio-backend/app/Http/Controllers/Accounts/AccountController.php:343-405`) calls FirstPromoter `POST /api/v1/promoters/create` with the user's email/name and `campaign_id: 6639`, then stores `first_promoter = { auth_token, ref_id, id }` on the user document and returns the refreshed profile.
- **Enrollment is instant and unreviewed.** The user goes from State D to State C in one click, with no application step.
- `first_promoter` is passed through to the FE profile payload (`app/Libraries/Account/Account.php:150`, `AccountController.php:65`), so `auth_token`, `ref_id` and the promoter `id` are already available client-side.

**Legacy duplicate**: `src/modules/setting/components/AffiliateProgram.vue` + route `affiliate` (`setting.ts:146`) is an older, pre-redesign version of the same iframe embed. Nothing in the nav links to it — it is orphaned but still routable.

**Mobile**: no Refer & Earn / affiliate surface exists anywhere in `contentstudio-flutter/lib/`. Nothing to change on mobile.

## Conflicts Between Marketing's Doc and What Exists

Decision taken: build the doc as written. Resolutions per item below.

1. **"Become an affiliate" destination vs. State 2 detection.** Marketing wants the primary button to open `contentstudio.firstpromoter.com/?utm_source=app...` in a new tab, and the trust line says "Applications reviewed in 3 business days". Today the button enrolls the user in-app via `POST /createPromoter` instantly. If the button becomes a plain external link, nothing ever writes `first_promoter.auth_token` to the user, so **the app can never detect that the user became an affiliate and State 2 would never render**. **Resolved:** the primary button does both — it fires the existing `POST /createPromoter` enrollment *and* opens the doc's FirstPromoter URL in a new tab. The enrollment call writes the token so State 2 still resolves, and the new tab lands the user on FirstPromoter to finish setting up their affiliate account, which is what the doc's link is for. `createPromoter` already handles the already-a-promoter case, so a user who signed up on FirstPromoter directly won't error.

**Still open for marketing:** "Applications reviewed in 3 business days" is shipping verbatim per the doc, but enrollment via `promoters/create` is instant and unreviewed. Either the campaign's approval setting needs to actually require review, or the line needs rewording. Flagged to the PO, not blocking the build.

2. **State 2 replaces the FirstPromoter iframe with native UI.** The link card, 5-metric stats row, tier strip and payout reminder are all data the app does not have. FirstPromoter's API can supply clicks, referral/customer counts and commission balances (`GET /promoters/show`), but **it has no "tier" field** — Partner/Silver/Gold/Elite/Agency Partner is a ContentStudio program construct. This needs a new backend endpoint, and the tier/progress line either needs a server-side derivation from active-referral count or has to ship rate-only per marketing's own fallback note.

3. **Affiliate link format.** `first_promoter.ref_id` is stored, but the full promotion URL is not, and nothing in either codebase builds one. FirstPromoter's `promoters/create` response also carries a default promotion link that is not currently persisted. The endpoint in item 2 should return the promotion link rather than have the frontend guess a `?fpr=` pattern.

4. **"Recurring means recurring" must be deleted, not reworded.** That maps to `settings.profile.refer_earn.intro.recurring.{title,body_1,body_2}` — removal across all 8 locale dirs. The existing header subtitle ("...while their account stays active") carries the same contradiction and is replaced by marketing's new subtext.

## What Needs to Change

**State 1 (not yet an affiliate) — copy and layout, no new data**
- Header title stays "Refer and Earn"; subtext replaced
- Main card heading → "Start at 15% recurring, grow to 40%", body replaced
- Add a **secondary** button "See full program details" (external) alongside the primary CTA — the card has only one button today
- Add a trust line under the buttons (new element)
- Retitle the 3 steps: Join/Apply in minutes, Refer/Share your link, Earn/15% to 40% recurring
- Replace the `Alert` box entirely with a "How the program works" facts box (6 bullets)
- Add an optional new-affiliate bonus line under the facts box (new element)
- Replace the footnote with the two-sentence version, with "Program eligibility and payout terms apply" as a link to the affiliate T&C page
- All external links open in a new tab, all carry the UTM strings from marketing's doc

**State 2 (approved affiliate) — new UI plus new data**
- Replace the iframe with: link card (read-only field + Copy button + toast + helper text), 5-metric stats row with a row-level empty state, tier strip with rate and progress line, primary "Open affiliate dashboard" / secondary "Get promo assets" / tertiary "Program details and terms" links, payout reminder line
- Needs a backend endpoint returning promotion link, clicks, signups, active referrals, pending commission, approved commission, and (if derivable) tier + rate + next-tier threshold

**Backend**
- New authenticated endpoint that fetches the current user's promoter record from FirstPromoter and returns the link + metrics, with caching (the page shouldn't hit FirstPromoter on every render) and a graceful degraded response so the FE can render the link card without stats

**Cleanup (optional, flag to PO)**
- Retire the orphaned `affiliate` route and `AffiliateProgram.vue`

## UI Components (from `docs/ui-components.md`)

Available and sufficient for State 1: `Button` (primary/secondary variants), `Icon`, `Alert`, `Loader`.

For State 2: `Button`, `TextInput` (read-only link field), `ActionIcon` or `Button` for Copy, `Icon`, `Loader`, `Progress` (optional for the tier progress line), `Badge` (optional for the tier name).

**Gaps:**
- **No card component** in `@contentstudio/ui` — the existing page hand-rolls cards with `rounded-2xl bg-white p-6`; the stats row and link card follow the same approach.
- **No standalone Tooltip** component (catalog line 137) — use `CstPopup` or `v-tooltip` if any metric needs a definition tooltip.
- **No skeleton loader** — stats load state uses `Loader`.
- Toast on copy: `useClipboard` composable already exists at `src/modules/setting/composables/useClipboard.ts` (wraps `@vueuse/core`) and is used by `CliCommandRow.vue` and `useDomainVerification.ts`. Reuse it; alerts go through `useAlertStore`.

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/setting/components/ReferEarn.vue` | State 1 rebuild, State 2 replaces iframe |
| `contentstudio-frontend/src/locales/*/settings.json` (8 dirs) | New `refer_earn` keys, delete `intro.recurring.*` |
| `contentstudio-frontend/src/api/profile.ts` | New affiliate-stats fetch fn (`createPromoter` at :150 stays) |
| `contentstudio-frontend/src/modules/core/queries/useProfileQueries.ts` | New query for affiliate stats |
| `contentstudio-frontend/src/config/api-utils` (URL config module) | New endpoint URL constant |
| `contentstudio-frontend/src/modules/setting/composables/useClipboard.ts` | Reuse as-is |
| `contentstudio-backend/routes/web/accounts.php` | New route beside `/createPromoter` (:26) |
| `contentstudio-backend/app/Http/Controllers/Accounts/AccountController.php` | New stats method beside `createPromoter()` (:343) |
| `contentstudio-frontend/src/modules/setting/components/AffiliateProgram.vue` + `setting.ts:146` | Optional retirement of the orphaned duplicate |

## Backlog Check

No existing story or feature under `docs/stories/` or `docs/features/` covers refer/affiliate/FirstPromoter work. This is net-new.

## Open Items for Marketing

1. **"Applications reviewed in 3 business days"** — copy ships verbatim, but in-app enrollment is instant. See conflict 1 above.
2. **"Get promo assets" destination** — the doc names an "affiliate resource library" but gives no URL. Needed before the affiliate view can ship.
3. **State 2 tertiary link has no UTMs** — the doc gives `https://contentstudio.io/affiliate-program` bare for "Program details and terms", while the same link in State 1 carries the full UTM set. Shipping as written; confirm whether that is deliberate.
4. **Tier source** — Partner/Silver/Gold/Elite/Agency Partner is not a FirstPromoter field. Confirm where tier is authoritative: derived from active-referral count server-side, or maintained manually in the campaign.

## Analytics Events

No existing Usermaven event covers affiliate actions (searched `userMaven.track(` across `contentstudio-frontend/src/` — closest is `team_member_invited`). Two new events specced:

- `affiliate_program_joined` — enrollment success, no payload
- `affiliate_link_copied` — Copy button success, no payload

Deliberately not tracked: clicks on the outbound dashboard / promo assets / program details links (navigation, already covered by global `pageview` tracking on the page itself).
