# Follow-up prompt — fix the create + sync flow

> Paste everything below the line into the **same** Claude conversation that produced the deck. Scoped to the empty state, the "Add competitors" modal and the "Sync X competitor data" modal. Logic, pricing and copy only — the visual language stays exactly as it is.

---

Three fixes, in order of importance: **the report has 6 profiles, not 5**; **the sync modal opens over the new report instead of being step 2 of a wizard**; and **competitors are added by handle, not by search**. Everything else below follows from those. Keep the visual language, components and layout exactly as they are, and don't touch any other screen in the deck.

## Fix 1 — A report is 5 competitors **plus the user's own profile**

My original brief said "up to 5 profiles including your own". That was wrong. In ContentStudio's existing Facebook and Instagram competitor reports the user adds **up to 5 competitors**, and **their own connected profile is included automatically** as the baseline everything is compared against. It never occupies a competitor slot and the user never types it in.

So a full report syncs **6 profiles**, and every cost figure I gave you was ~20% too low. Use these and nothing else.

**As competitors are added (at the default 20 posts each):**

| Competitors | Profiles synced | Wallet | Credits |
|---|---|---|---|
| 1 | 2 | $0.26 | 14 credits |
| 2 | 3 | $0.40 | 20 credits |
| 3 | 4 | $0.53 | 27 credits |
| 4 | 5 | $0.66 | 34 credits |
| **5 (full report)** | **6** | **$0.79** | **40 credits** |

**A full report, as the post count changes:**

| Posts per profile | Total posts | Wallet | Credits |
|---|---|---|---|
| 10 | 60 | $0.43 | 22 credits |
| **20 (default)** | **120** | **$0.79** | **40 credits** |
| 30 | 180 | $1.15 | 58 credits |
| 50 | 300 | $1.87 | 94 credits |
| 80 | 480 | $2.95 | 149 credits |
| 100 | 600 | $3.67 | 185 credits |

Only two things are billable: **profile lookups** and **posts returned**. Everything else rides along free — likes, reposts, replies, quotes, bookmarks and impressions all arrive inside each post we already paid for, and bio, location, website, join date and verified status all arrive inside the profile lookup. So there is nothing else to itemise, and **no per-unit rate may ever appear in the UI** — no `$0.13 / profile`, no `$0.006 / post`, no percentages, nothing about what X charges us. The user sees one total for the sync in front of them. Put any economics explanation in the annotation text outside the frame.

## Fix 2 — The sync modal opens over the new report, not as step 2 of a wizard

Right now **Next** takes the user straight from the competitor list into the cost modal, which means someone who can't afford the sync has done all that work for nothing.

Change it to:

1. The user adds competitors and clicks **Create report** (rename the button from **Next**).
2. They land on the **report page for the report they just made** — their competitors listed, no data yet.
3. The **Sync X competitor data** modal opens over it automatically, once.
4. They either **Sync now**, or close it and the empty report sits there with a **Sync now** button in the header for later.

So please add one thing to the deck: the **empty report page** behind that modal. Report name in the header, an X logo chip, a **Sync now** button, the five competitors listed as rows with their handles and follower counts, and an empty area below reading **Sync this report to see followers, engagement and top posts.** Nothing else — no charts, no placeholder skeletons.

Because this is no longer a wizard, the sync modal has **no Back button and no "Step 2 of 2"**. It's the same standalone modal the user gets later from **Sync now**, which is the point — one modal, one pattern.

## Fix 3 — Competitors are added by handle, not by search

Step 1 of the empty state says *"Search by profile name or handle"* and the modal has a **Search a profile name or @handle** box. Neither is buildable: X has no cheap profile search, a search-as-you-type box would bill us on every keystroke, and searching by display name isn't available at all. Competitors are added by **exact handle, confirmed one at a time**.

## Screen 1 — Empty state

1. **The payment-model toggle changes nothing on this screen** — no currency appears anywhere. Anchor a real number in the intro, and fix the profile framing:
   - Wallet: **Add up to 5 competitors and compare them against your own X profile — followers, engagement rate and posting activity side by side. A full report costs about $0.79 each time you sync it.**
   - Credits: **…costs about 40 X credits each time you sync it.**
2. **Rewrite step 1** — title **Add competitor profiles**, body **Enter each competitor's X handle and we'll confirm the account exists. Up to 5 competitors per report.**
3. **Drop the words "read" and "fetch" from product copy.** Use this sentence wherever the point needs making, verbatim: **X charges us for the data in this report, so it updates only when you sync it — and you'll always see the cost first.** Apply it to step 2's body.
4. **Rename the header button** from **Create New** to **Create X report**.
5. **Add the platform switcher** — **Facebook · Instagram · X** above the page title, X selected — so it reads as a third platform in an existing module rather than a standalone tool.

## Screen 2 — "Add competitors" modal

1. **Split the list in two.** A **Your profile** heading with the connected X account pinned above a divider — labelled **(you)**, added automatically, **no remove button**, and **not counted** toward the 5. Then a **Competitors** heading with the rows the user added, counted as **5 of 5 competitors added**. Right now all six are jumbled together with the user's own profile carrying a delete button like everyone else.
2. **Replace the search box** with an input placeholder `@handle`, a **Verify** button beside it, and helper text **Enter the exact handle — we'll confirm the account exists before adding it. Checking a handle is free.**
3. **Show the six add states** through the segmented control:
   - **Verified** — green check row: `Buffer · @buffer · 198.5K followers`
   - **Not found** — **We couldn't find @bufffer on X. Check the spelling and try again.**
   - **Protected** — **@example is a protected account. X only shares data for public profiles, so it can't be tracked.**
   - **Suspended** — **@example is suspended on X and has no data to track.**
   - **Already added** — **@buffer is already in this report.**
   - **Limit reached** — input and **Verify** both **disabled**, with **You've added 5 of 5 competitors. Remove one to add another.** Right now the box stays enabled at the limit.
4. **Add a live cost line above the footer**, updating as rows are added and removed, worded so the user's own profile is visibly part of it: **Syncing your profile + 5 competitors will cost $0.79.** (Credits: **…will cost 40 X credits.**)
5. **Add the balance line** under the modal subtitle. Wallet: **X Wallet balance: $24.80** with a **Top up** link. Credits: **X credits remaining this month: 112 of 150**, plus **Shared with your X publishing credits.**
6. **Drop the footer counter** `0 of 5 slots left.` — it contradicts the header's `5 of 5 added`. Keep the header one.
7. **Rename Next to Create report.**
8. **Fix the info band copy** to the standard sentence above.

## Screen 3 — "Sync X competitor data" modal

1. **Posts per profile must be a dropdown**, not a text field — `10, 20, 30, 50, 80, 100`, default **20**. A free-text number invites `500` and an unpayable bill.
2. **The summary block must show all six profiles.** It currently reads `5 profiles × 20 posts / 100 posts`, which undercounts and hides the user's own profile. Make it **You + 5 competitors × 20 posts** on the left, **120 posts** on the right.
3. **Wire the dropdown up** so changing it live-updates the summary and the cost together, per the second table.
4. **Put the cost on the button** as well as in the cost block: **Sync now — $0.79**. This matches how X analytics already does it elsewhere in the product.
5. **Make all three cohort states actually render** — only the wallet one does right now:
   - **X Wallet** — cost **$0.79**, balance **X Wallet balance: $24.80** with **Top up**.
   - **X Credits** — cost **40 X credits**, balance **X credits remaining this month: 112 of 150**, plus **Shared with your X publishing credits.** No `$` anywhere on the screen, body copy included.
   - **Not enough balance** — cost **$0.79**, balance **$0.12**, **Sync now** disabled, warning band **Your X Wallet doesn't have enough balance for this sync. Top up to continue.** with a **Top up wallet** button. Credits variant: **You don't have enough X credits left this month. Buy more credits to continue.** → **Buy credits**.

   A workspace is on one payment model and sees one currency everywhere. `$` and `credits` must never appear on the same screen.
6. **Remove Back** — per Fix 2 this is a standalone modal, so the footer is **Cancel** / **Sync now — $0.79**.
7. **Fix the intro line** to avoid "read": **X charges us for each profile and each post in this report. Choose how many recent posts to include for each profile.**
8. **Fix the backdrop** — put the new empty report page behind it, dimmed, not the "Add your first competitor" empty state.

## No sync-frequency picker in this deck

The **Sync frequency: Daily** dropdown you had earlier is gone for a reason: an automatic schedule is a separate settings screen with its own recurring-cost preview, and it's out of scope for these mockups. Don't reintroduce it here, and don't put "updates daily" or any auto-refresh language anywhere in the product UI — the whole point of these screens is that X reports refresh only when the user chooses to pay for it.

## One line for the open questions panel

**At the default 20 posts, a full report costs 40 X credits — and a Standard plan only gets 45 a month.** Should the default drop to 10 posts (22 credits) so a legacy-credit workspace can sync more than once?
