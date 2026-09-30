# Stories: "+ New" button in create-capable dropdowns

Epic: **Q3 - 2026: Miscellaneous**

1. `[Design] Finalize the "+ New" button for the label, campaign, template, tag and custom view dropdowns`
2. `[FE] Replace the "+" icon with a "+ New" button in the label, campaign, template, tag and custom view dropdowns`

---

# [Design] Finalize the "+ New" button for the label, campaign, template, tag and custom view dropdowns

### Description

As a designer, I want to finalise how the "+ New" button from the workspace dropdown carries over to every dropdown that lets a user create something, so that devs build one consistent button instead of five slightly different ones and the create action is easy to spot everywhere.

Today these dropdowns use a small, icon-only circular "+" next to the dropdown heading. It is easy to miss, its meaning only shows on hover, and it doesn't match the "+ New" button the workspace dropdown now uses. The pattern exists, but each dropdown has a different header layout (some have a search field in the same row, some have a settings icon), so the placement needs a designer's call for each one.

---

### Workflow

1. Designer reviews the workspace dropdown's "+ New" button as the reference.
2. Designer reviews the five dropdowns in scope: Labels and Campaigns (Composer and Planner), Templates (Composer and the templates picker outside Composer), Inbox Tags, and Planner Custom Views.
3. Designer places the "+ New" button in each dropdown header, keeping it on one line next to the heading and the search field.
4. Designer defines the disabled look used while the inline create form is already open.
5. Designer defines the empty-state version, for example when a workspace has no labels yet.
6. Designer confirms the button label for the Composer template dropdown, where the action saves the current post as a template instead of creating a blank one.

---

### Acceptance criteria

- [ ] Final designs show the "+ New" button in the header of all five dropdowns: Labels, Campaigns, Templates, Inbox Tags and Planner Custom Views
- [ ] The button matches the workspace dropdown's "+ New" button (same component, size, icon and colour treatment)
- [ ] Each header is shown with the longest supported translation of "New" and still fits on one line
- [ ] Default, hover, disabled (create form already open) and empty-state versions are designed
- [ ] The Composer template dropdown's button label and tooltip are confirmed
- [ ] Every element is mapped to an existing `@contentstudio/ui` component (`Button`, `Icon`), or flagged as a gap
- [ ] Designs are handed off to **[FE] Replace the "+" icon with a "+ New" button in the label, campaign, template, tag and custom view dropdowns**

---

### Mock-ups:

To be attached by the designer. Reference: the workspace dropdown header ("Your Workspaces" with the "+ New" button on the right).

---

### Impact on existing data:

None.

---

### Impact on other products:

Web app only. The mobile app and the Chrome extension have their own pickers.

---

### Dependencies:

None. Blocks **[FE] Replace the "+" icon with a "+ New" button in the label, campaign, template, tag and custom view dropdowns**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, design only, nothing API-facing changes

---

# [FE] Replace the "+" icon with a "+ New" button in the label, campaign, template, tag and custom view dropdowns

### Description

As a ContentStudio user adding a label, campaign or template to a post, I want the create action in the dropdown to be a clearly labelled "+ New" button, so that I can see straight away how to create one without hovering over a tiny icon to find out what it does.

The small circular "+" icon is hard to see, gives no hint on touch devices, and no longer matches the workspace dropdown, which already uses a "+ New" button. This story swaps the icon for that button in five dropdowns. What each button does stays exactly as it is today.

---

### Workflow

1. User opens Composer and clicks the Labels dropdown.
2. User sees the heading "Labels" with a **+ New** button at the right end of the header row.
3. User hovers the button and reads the tooltip explaining what a label is for.
4. User clicks **+ New** and the same inline "create label" form opens as before. While it is open the button shows as disabled.
5. User saves the label and the button becomes active again.
6. The same button appears in the Campaigns dropdown, the Templates dropdown, the Inbox Tags dropdown and the Planner Custom Views dropdown, and each one does what its "+" icon did before.

---

### Acceptance criteria

**Button**

- [ ] The circular, icon-only "+" no longer appears in the header of the Labels, Campaigns, Templates, Inbox Tags or Planner Custom Views dropdowns
- [ ] Each of those headers shows a **+ New** button built with the `Button` component from `@contentstudio/ui` (same variant and size as the workspace dropdown's "+ New" button) with the `Plus` `Icon`, with no colours overridden by hardcoded classes
- [ ] The button follows the workspace's primary theme colour, so on a white-label domain it renders in that domain's colour
- [ ] Clicking the button does exactly what the old icon did in that dropdown (opens the inline create form, opens the save-as-template modal, opens Composer, or starts a new custom view)
- [ ] While an inline create form is open (Labels, Campaigns, Inbox Tags) the button is disabled and cannot be clicked again
- [ ] The button is hidden for users who can't create that item today, and the header spacing stays correct without it

**Copy**

- [ ] Button label in every dropdown: "New", except the Composer template dropdown, which uses the label confirmed in the design story
- [ ] Labels tooltip: "Create a label to tag and filter your posts. For example, "Product launch" or "Customer stories"."
- [ ] Campaigns tooltip: "Create a campaign to group related posts, like "Black Friday 2026", and follow them together in the Planner and Analytics."
- [ ] Templates tooltip in Composer: "Save this post as a template so you can reuse its text and media in future posts."
- [ ] Templates tooltip outside Composer: "Create a new template in Composer that you can reuse in future posts."
- [ ] Inbox Tags tooltip: "Create a tag to sort conversations. For example, "Refund request" or "VIP customer"."
- [ ] Planner Custom Views tooltip: "Save your current Planner filters as a view so you can come back to them in one click."
- [ ] Each button's accessible label reads the same as its tooltip's first sentence

**Empty states**

- [ ] Where a dropdown has no items yet (for example no labels in the workspace), the empty-state message uses the same **+ New** button instead of the inline "+" icon

**Translations**

- [ ] All labels and tooltips come from translation keys, with no hardcoded English
- [ ] Every supported language has a translation of "New" that reads correctly as a standalone button
- [ ] In the language with the longest label, every header still fits on one line with no wrapping or overlap with the search field

---

### Mock-ups:

From **[Design] Finalize the "+ New" button for the label, campaign, template, tag and custom view dropdowns**.

---

### Impact on existing data:

None. Presentation and copy only.

---

### Impact on other products:

- **Mobile app (Flutter):** no impact, it has its own pickers.
- **Chrome extension:** check whether it reuses the Label or Campaign dropdown. If it does, it picks up the change and should be checked.
- **White-label domains:** must be checked on a domain with a non-default primary colour.

---

### Dependencies:

**[Design] Finalize the "+ New" button for the label, campaign, template, tag and custom view dropdowns**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes
