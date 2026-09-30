# Deferred draft: connect modal capability columns

**Status:** drafted 2026-09-29, **held back by the PO** from the "Reconnect and connect accounts from wherever they're used" epic. Not pushed to Helpin. Pick it up when the PO says so.

Design reference: design canvas row 3, https://claude.ai/artifact/APzGjGQbTCJn8rYJ5EHtur

When this is picked up:
- Add the modal to the design scope (columns at the new width, partial-support notes, the Inbox and Analytics highlighted variants, the Facebook and Instagram menu lines). It was taken out of the epic's [Design] story.
- Add back to **[FE] Add connect and settings shortcuts to every account picker**: the Inbox filter's `+` opens the modal in its Inbox view, and Analytics' `+` in its Analytics view.
- Verify the open capability cells in `01-research.md` section 5 (Threads inbox, Instagram direct vs via Facebook, TikTok types).

---

# [FE] Show where each platform works in the Connect Social Accounts modal

### Description

As a ContentStudio user connecting a social account, I want to see at a glance which parts of ContentStudio each platform works with, and what each Facebook or Instagram option gives me, so that I connect the right thing the first time instead of connecting a Facebook Profile and then wondering why it never appears in my Inbox or Analytics.

The modal keeps its current layout: the same rows, the same EasyConnect banner, the same `+` on each row. Platforms with one connection type still connect in one click. Only Facebook and Instagram open a menu, as they do today.

---

### Workflow

1. User opens the Connect Social Accounts modal from anywhere.
2. Above the list, three column headings read **Publish**, **Inbox** and **Analytics**. Each row shows a check where the platform works, a light dash where it doesn't, and a short note where it only partly works.
3. User sees that Tumblr has a check only under Publish, and that Facebook's Inbox and Analytics checks say "Pages only".
4. User hovers "Pages only" and reads "Comments and messages from your Facebook Pages show up in Inbox. Profiles and Groups don't."
5. User clicks Facebook's `+`. The menu shows each option with a line underneath: "Connect Facebook Profile: Publishing through a phone notification. No Inbox or Analytics."
6. User opens the modal from the Inbox filter's `+` instead. The Inbox column is highlighted, the platforms that work in Inbox are listed first, and the rest sit below a "Not available in Inbox" label, dimmed but still connectable.

---

### Acceptance criteria

**Columns:**

- [ ] The modal widens to fit three columns headed "Publish", "Inbox" and "Analytics", aligned with each row's cells, and the headings stay visible while the list scrolls
- [ ] Each cell shows a check (works), a light dash (doesn't work) or a check with a short note (partly works)
- [ ] Cell values match this table:

| Platform | Publish | Inbox | Analytics |
|---|---|---|---|
| Facebook | ✓ | ✓ Pages only | ✓ Pages only |
| Instagram | ✓ | ✓ | ✓ |
| Threads | ✓ | ✓ Comments | ✓ |
| X (Twitter) | ✓ | – | ✓ With X app |
| LinkedIn | ✓ | ✓ Pages only | ✓ |
| Pinterest | ✓ Boards only | – | ✓ |
| GBP | ✓ | ✓ Reviews | ✓ |
| YouTube | ✓ | ✓ Comments | ✓ |
| Tumblr | ✓ | – | – |
| TikTok | ✓ | – | ✓ |
| Bluesky | ✓ | – | ✓ |
| Telegram | ✓ | – | – |
| Meta Ads | – | – | ✓ Ads only |
| Google Ads | – | – | ✓ Ads only |
| WhatsApp | – | ✓ Messages | – |

- [ ] Cells follow the workspace's feature access: Threads Analytics shows "–" where Threads analytics isn't enabled, and LinkedIn Analytics shows "✓ Pages only" where LinkedIn profile analytics isn't enabled
- [ ] Each cell has a hover tooltip:
  - Works in Publish: "You can publish and schedule posts to {Platform}."
  - Works in Inbox: "Comments and messages from {Platform} show up in your Inbox so you can reply."
  - Works in Analytics: "See {Platform} analytics for the accounts you connect."
  - Doesn't work: "{Platform} doesn't work in {Publishing / Inbox / Analytics} yet."
  - Facebook Inbox "Pages only": "Comments and messages from your Facebook Pages show up in Inbox. Profiles and Groups don't."
  - Facebook Analytics "Pages only": "See analytics for your Facebook Pages. Profiles and Groups don't have analytics."
  - LinkedIn Inbox "Pages only": "Comments on your LinkedIn Page posts show up in Inbox. Personal profiles don't."
  - Threads Inbox "Comments": "Replies to your Threads posts show up in Inbox."
  - YouTube Inbox "Comments": "Comments on your videos show up in Inbox."
  - GBP Inbox "Reviews": "Google reviews for your locations show up in Inbox so you can reply."
  - WhatsApp Inbox "Messages": "Customer chats on your WhatsApp Business number show up in Inbox."
  - Pinterest Publish "Boards only": "Posts go to your Pinterest boards. Your profile is used for analytics."
  - X Analytics "With X app": "Analytics shows up when analytics is turned on for the X app you connect with."
  - Meta Ads and Google Ads Analytics "Ads only": "See how your ads perform. This connection doesn't publish posts or show up in Inbox."
- [ ] Long row subtitles, for example "(Profiles & Public Boards)", are cut off with "…" before they reach the columns

**Facebook and Instagram menus:**

- [ ] Each option keeps its current label and gets a line underneath:
  - Connect Facebook Page: "Publishing, Inbox and Analytics"
  - Connect Facebook Profile: "Publishing through a phone notification. No Inbox or Analytics."
  - Connect Facebook Group: "Publishing through a phone notification. No Inbox or Analytics."
  - Connect via Facebook Account: "Publishing, Inbox and Analytics"
  - Connect Directly with Instagram: "Publishing, Inbox and Analytics"
- [ ] Lines that mention a limitation are shown in the warning colour
- [ ] Instagram's existing help icons and their text stay

**Opened from Inbox or Analytics:**

- [ ] Opened from the Inbox filter's `+`: the Inbox heading and column are highlighted, platforms that work in Inbox are listed first, and the rest follow under the label "Not available in Inbox", dimmed to about half opacity but still connectable
- [ ] Opened from an Analytics `+`: the same with the Analytics column and the label "Not available in Analytics"
- [ ] Opened from anywhere else: the standard order with nothing highlighted

**Unchanged:**

- [ ] The EasyConnect banner, the "N Connected" pills, the Learn more link and Cancel behave as today
- [ ] Platforms with one connection type still connect directly from `+` with no menu
- [ ] The X (Twitter) row keeps its current lock and custom-app menu behaviour
- [ ] EasyConnect pages that reuse the modal show the same columns

**Consistency:**

- [ ] The column values come from one place in the code that describes each platform, not separate lists per screen, so that adding a platform or a capability updates the modal in one edit

---

### Mock-ups:

Design canvas, row 3: https://claude.ai/artifact/APzGjGQbTCJn8rYJ5EHtur. Final visuals come from **[Design] Design the Reconnect required message and the account picker shortcuts**.

---

### Impact on existing data:

None.

---

### Impact on other products:

The guided onboarding connect step uses the same modal body, so it gets the columns too. The mobile app has its own connect screen and is out of scope.

---

### Dependencies:

- **[Design] Design the Reconnect required message and the account picker shortcuts**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage — N/A, no API changes
