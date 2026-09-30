# Epic: App shell refresh, with top bar, global search and Quick add

## Epic description

The desktop navigation rail has run out of room. The team wants more items and faster navigation in it, but there's no space left, and the sidebar itself looks weak next to the rest of the product.

This epic rebuilds the app shell on desktop. The page a user is working on sits inside one clean card. The navigation rail on the left and a new top bar are joined around that card and separated from it by the card's edge, so the navigation reads as one frame and the work reads as one surface. The top bar gives the rail its breathing room back and adds two things ContentStudio doesn't have today:

- **Global search**: type to jump straight to any module, page, setting or AI tool (for example "inbox", "billing", "image to video"), from anywhere, including with a keyboard shortcut.
- **Quick add**: one button to create the things people add most, like a post, a social account or a team member, without first navigating to the right settings page.

**Status:** skeleton. The PO is sharing competitor references and design inspiration, not finished designs, so the designer produces the actual design. The stories will be detailed once that design exists. What moves from the rail to the top bar, the exact layout and the final copy are settled in the design story.

### Scope

In:

- Card-based desktop layout with a top bar, and the rail rebalanced around it
- Global search to jump to any module, page, setting or AI tool
- Quick add menu in the top bar

Out:

- The Health Center (its own epic: **Workspace Health Center**)
- The mobile web header and drawer, which stay as today unless the designs say otherwise
- The mobile app

### Stories

1. `[Design] Design the card-based app shell with a top bar, global search and Quick add`
2. `[FE] Move the app into a content card with a desktop top bar`
3. `[FE] Add global search to jump to any module, page, setting or AI tool`
4. `[FE] Add a Quick add menu to the top bar`

---

# [Design] Design the card-based app shell with a top bar, global search and Quick add

### Description

As a designer, I want to turn the team's direction for the new app shell into final designs, so that devs rebuild the desktop layout, the top bar, search and Quick add from one agreed reference, and the rail finally has room for what it needs.

The PO has competitor references and design inspiration to share. They are a starting point, not a design. This story produces the actual design.

---

### Workflow

1. Designer reviews the competitor references and inspiration the PO shares and today's desktop rail (workspace switcher, modules, More menu, Customize sidebar, notifications, approvals, theme picker, profile).
2. Designer designs the shell: the content card, the rail and the top bar joined around it, and how they meet at the card's edge.
3. Designer decides what lives in the top bar and what stays in the rail, and what the rail gains with the space freed.
4. Designer designs global search: the field in the top bar, the open results panel, result groups, keyboard navigation and the empty and no-results states.
5. Designer designs the Quick add button and its menu.
6. Designer places the trial, billing and notice banners in the new layout.
7. Designer checks the shell against every rail theme and a white-label theme, and at common laptop sizes.

---

### Acceptance criteria

- [ ] Final designs cover the desktop shell: content card, rail and top bar, at 1280, 1440 and 1920 widths
- [ ] A clear list of what sits in the top bar and what stays in the rail
- [ ] Global search is designed in its closed, open, typing, results, no-results and keyboard-focused states
- [ ] Quick add is designed with its final list of actions and how items the user can't use are handled
- [ ] Banners (trial, billing, notices) have a defined place
- [ ] The shell works with every existing rail theme and with white-label colours and logos
- [ ] Every element is mapped to an existing `@contentstudio/ui` component, or flagged as a gap. Likely gap: a search results panel (command palette) component
- [ ] Designs are handed off to every `[FE]` story in this epic

---

### Mock-ups:

Competitor references and inspiration from the PO, to be attached. Final designs to be attached by the designer.

---

### Impact on existing data:

None.

---

### Impact on other products:

Web app on desktop. Mobile web and the mobile app are not in scope.

---

### Dependencies:

None. Blocks every `[FE]` story in this epic.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, design only, nothing API-facing changes

---

# [FE] Move the app into a content card with a desktop top bar

### Description

As a ContentStudio user on desktop, I want the page I'm working on to sit in one clear card, framed by the navigation rail and a top bar, so that the app feels calmer and more polished, and the navigation has room for everything I use instead of hiding things under More.

Today the desktop app has only the left rail. Everything else, like notifications, approvals, the changelog and the profile menu, is packed into it, and items that don't fit spill into More. This story adds the top bar, moves the agreed items into it, and places every page inside the content card.

**Detail to follow once the design story is done:** the final split of items between rail and top bar, spacing, and copy.

---

### Workflow

1. User logs in on a desktop browser.
2. User sees the navigation rail on the left and a top bar across the top, joined into one frame around a content card.
3. The page they opened (for example Planner) sits inside the card.
4. User finds notifications, approvals and their profile where the designs place them in the top bar.
5. User sees more of their modules directly in the rail, with fewer items tucked under More.
6. User switches workspaces, opens Composer, and moves between modules exactly as before. Only the frame around them has changed.

---

### Acceptance criteria

- [ ] On desktop, every page renders inside the content card, with the rail and top bar framing it as designed
- [ ] The top bar holds the items the design story assigns to it. They work exactly as they do in the rail today
- [ ] The rail shows the items the design story assigns to it. Customize sidebar (show, hide, reorder) keeps working, and each user's saved rail layout is kept
- [ ] Items only overflow into More when the window is genuinely too short, and fewer items overflow than today at the same height
- [ ] Trial, billing and notice banners appear where the design places them and never cover the content
- [ ] The top bar leaves room for search and Quick add (see **[FE] Add global search to jump to any module, page, setting or AI tool** and **[FE] Add a Quick add menu to the top bar**)
- [ ] Every existing rail theme and white-label theme and logo displays correctly in the new frame
- [ ] Full-screen pages that hide the rail today (for example some onboarding and editor views) keep behaving as they do
- [ ] Mobile web keeps its current header and drawer
- [ ] No page scrolls sideways at 1280px wide
- [ ] Copy for any new or moved labels and tooltips comes from translation keys

---

### Mock-ups:

From **[Design] Design the card-based app shell with a top bar, global search and Quick add**.

---

### Impact on existing data:

None. Saved rail layouts (order and hidden items) are kept.

---

### Impact on other products:

- **White-label domains:** the frame must follow white-label colours and logos.
- **Mobile web:** unchanged.
- **Mobile app and Chrome extension:** no impact.

---

### Dependencies:

**[Design] Design the card-based app shell with a top bar, global search and Quick add**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Add global search to jump to any module, page, setting or AI tool

### Description

As a ContentStudio user, I want to type what I'm looking for into one search box, like "inbox", "billing" or "image to video", and jump straight there, so that I don't have to remember which menu, tab or settings page holds it.

ContentStudio has many modules, settings pages and AI tools, and today the only way to reach them is to click through the rail and menus. This story adds a search box to the top bar that finds any place in the app the user has access to and takes them there. A keyboard shortcut opens it from anywhere.

**Detail to follow once the design story is done:** exact result groups, whether v1 includes content (posts, media) as well as places, and final copy.

---

### Workflow

1. User clicks the search box in the top bar, or presses Ctrl+K (Cmd+K on Mac) anywhere in the app.
2. The search opens with a short list of suggestions, for example recently visited pages.
3. User types "bill". Results update as they type: "Billing & plans" under Settings.
4. User moves through results with the arrow keys and presses Enter, or clicks one.
5. ContentStudio opens that page, and the search closes.
6. User types something that matches nothing, like "zzz", and sees "No results for "zzz"".
7. User presses Esc to close the search without going anywhere.

---

### Acceptance criteria

- [ ] A search field sits in the top bar with placeholder "Search or jump to..." and a hint showing the shortcut (Ctrl K or ⌘K)
- [ ] Ctrl+K on Windows and Linux, and Cmd+K on Mac, opens search from any page, including while focus is elsewhere. It doesn't open while the user is typing in Composer or AI chat if the design says so
- [ ] Results cover every module, the main pages inside each module, every settings page, and every AI Studio tool
- [ ] Results update as the user types, match partial words and common synonyms (for example "billing" and "plan", "team" and "members"), and ignore capitalisation
- [ ] Results are grouped by type (for example Pages, Settings, AI tools) with an icon for each
- [ ] Only places the user can open in the current workspace, with their role and plan, are shown
- [ ] Up and down arrows move through results, Enter opens the highlighted one, and Esc closes search
- [ ] Opening a result goes to that page and closes search
- [ ] Before typing, search shows the user's recently visited pages
- [ ] No-results state: "No results for "{query}"" with the text "Try a module name like "Planner", or a setting like "Billing"."
- [ ] Search works in every supported language, matching the translated page names
- [ ] Using search fires a `global_search_result_opened` Usermaven event with `{ result_type, result_id }`. The typed query is not sent

---

### Mock-ups:

From **[Design] Design the card-based app shell with a top bar, global search and Quick add**.

---

### Impact on existing data:

None, unless recent pages are remembered across devices, in which case they're stored with the user's preferences.

---

### Impact on other products:

- **White-label domains:** results must never show ContentStudio-branded pages that white-label users can't see.
- **Mobile app and Chrome extension:** no impact.

---

### Dependencies:

- **[Design] Design the card-based app shell with a top bar, global search and Quick add**
- **[FE] Move the app into a content card with a desktop top bar**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes

---

# [FE] Add a Quick add menu to the top bar

### Description

As a ContentStudio user, I want one button in the top bar that lets me add a post, a social account, a team member and other common things from wherever I am, so that I don't have to go looking for the right settings page every time I want to add something.

Today each "add" lives in its own place: new posts in Composer, social accounts under Settings, team members under team settings, workspaces in the workspace dropdown. This story adds a **Quick add** button to the top bar that opens a short menu of these actions. Each one opens the same flow it opens today.

**Detail to follow once the design story is done:** the final list and order of actions, and final copy.

---

### Workflow

1. User clicks **+** (Quick add) in the top bar.
2. A menu opens with:
   - Create post
   - Connect social account
   - Invite team member
   - Create workspace
   - Create label
   - Create campaign
3. User clicks **Connect social account**. The Connect Social Accounts modal opens on top of the page they were on.
4. User connects an account and closes the modal, and they're still on the same page.
5. A user who isn't allowed to invite team members doesn't see **Invite team member** in the menu.

---

### Acceptance criteria

- [ ] A Quick add button sits in the top bar where the design places it, with the tooltip "Add something new: a post, a social account, a team member and more."
- [ ] Clicking it opens a menu (`Dropdown` / `DropdownItem` from `@contentstudio/ui`) with the final list of actions from the design story, each with an icon
- [ ] Each action opens the same flow it opens today, without leaving the current page where the flow is a modal:
  - Create post opens Composer
  - Connect social account opens the Connect Social Accounts modal
  - Invite team member opens the invite flow
  - Create workspace opens the create-workspace flow
  - Create label and Create campaign open their create forms
- [ ] Actions the user isn't allowed to do in the current workspace (for their role, or for approvers and collaborators) are hidden
- [ ] Actions that the user's plan limit blocks (for example no more social account slots) show the existing upgrade prompt instead of failing
- [ ] The menu can be used with the keyboard (arrow keys, Enter, Esc)
- [ ] Choosing an action fires a `quick_add_used` Usermaven event with `{ action }`
- [ ] All labels and tooltips come from translation keys in every supported language

---

### Mock-ups:

From **[Design] Design the card-based app shell with a top bar, global search and Quick add**.

---

### Impact on existing data:

None. Reuses existing create flows.

---

### Impact on other products:

- **White-label domains:** the menu must follow white-label theming.
- **Mobile app and Chrome extension:** no impact.

---

### Dependencies:

- **[Design] Design the card-based app shell with a top bar, global search and Quick add**
- **[FE] Move the app into a content card with a desktop top bar**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes
