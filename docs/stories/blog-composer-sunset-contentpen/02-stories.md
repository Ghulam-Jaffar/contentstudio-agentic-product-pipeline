# Stories: Blog Composer Sunset → ContentPen Migration

---

## Story 1: [FE] Sunset blog composer for old users with phased ContentPen migration

### Description:

As the ContentStudio product team, we want to sunset the blog composer for old users (those with `redirect_to_content_pen: false` or unset) and migrate them to ContentPen.

**Users with `redirect_to_content_pen: true` are already handled** — they see the ContentPen CTA modal when creating new blog posts, and go directly to `app.contentpen.ai` if linked. No changes needed for these users.

This story only affects **old users** where the redirect flag is not set. Two phases:

**Phase 1 (Now → April 30, 2026):**
- Old users can still **create new** and **edit** blog posts — no blocking yet
- A **persistent non-dismissable sunset banner** is shown at the top of the blog composer (for both new and edit) warning them the feature is moving to ContentPen
- Banner has a CTA that opens the ContentPen CTA modal

**Phase 2 (April 30, 2026+):**
- Old users can **no longer create new** blog posts — the "create new" entry points now follow the same flow as `redirect_to_content_pen: true` (show ContentPen CTA modal)
- Old users can still **edit existing** blog posts from planner — but with an updated, stronger banner
- A date check (`new Date() >= new Date('2026-04-30')`) controls the phase transition — no deployment needed on April 30

**Scheduled/published blog posts are unaffected in both phases** — they continue to work normally.

**Entry points that need changes (only when `redirect_to_content_pen` is false):**
- `Blog.vue` — add sunset banner (Phase 1 and Phase 2 versions)
- `useComposerPost.js` — after April 30, block "create new" for old users (same as redirect flow)
- `ComposerDropdown.vue` — after April 30, "Composer Article" option shows ContentPen modal for old users
- `useNotificationHandler.js` — after April 30, blog notification clicks for new posts redirect to modal

**Entry points already handled (no changes needed):**
- All entry points for users with `redirect_to_content_pen: true` — already showing ContentPen modal
- `SidebarMain.vue`, `WelcomeRow.vue`, `useAIPostGeneration.ts` — already go through `useComposerPost.js`

**Components to use:**
- `Alert` from `@contentstudio/ui` — sunset banner in blog composer
- `Button` from `@contentstudio/ui` — CTA in banner
- `Icon` from `@contentstudio/ui` — warning icon in banner
- Existing `ContentPenCTAModal` — already handles super admin vs team member experience

---

### Workflow:

**Phase 1 (Now → April 30) — old users only:**
1. User clicks "create blog post" from any entry point — blog composer opens normally
2. User sees a persistent amber banner at the top of the blog composer about the upcoming sunset
3. User can create and save the blog post as usual
4. User clicks "Learn more about ContentPen" in the banner — ContentPen CTA modal opens
5. User opens an existing blog post to edit — same banner appears, editing works normally

**Phase 2 (April 30+) — old users only:**
6. User clicks "create blog post" from any entry point — instead of opening the blog composer, the ContentPen CTA modal opens (same behavior as users with `redirect_to_content_pen: true`)
7. Super admin sees "Try ContentPen" CTA button; team member sees contact-admin message
8. User opens an existing blog post to edit from planner — blog composer opens with an updated, stronger banner
9. User can still edit and save the existing post

---

### Acceptance criteria:

**Phase 1 — Banner on all blog composer usage (before April 30):**
- [ ] When an old user (`redirect_to_content_pen` is false/unset) opens the blog composer (new or edit), a persistent non-dismissable banner appears at the top
- [ ] The banner uses `Alert` component with amber/warning styling
- [ ] Phase 1 banner copy: "Blog composer is moving to ContentPen on April 30, 2026. We recommend switching to ContentPen now for a better blog writing experience with AI-powered SEO, internal linking, and more."
- [ ] Banner has a CTA button: "Try ContentPen" that opens the `contentpen-cta-modal`
- [ ] The banner cannot be dismissed or closed
- [ ] Creating and editing blog posts still works normally — no functionality is blocked
- [ ] Users with `redirect_to_content_pen: true` are not affected — their flow stays the same

**Phase 2 — New posts blocked, edit still allowed (April 30+):**
- [ ] A date check (`new Date() >= new Date('2026-04-30')`) controls the phase transition
- [ ] After April 30, old users clicking "create new blog post" from any entry point see the ContentPen CTA modal instead of the blog composer — same behavior as the existing `redirect_to_content_pen: true` flow
- [ ] After April 30, old users can still open existing blog posts for editing from planner
- [ ] The edit banner updates to Phase 2 copy: "Creating new blog posts is no longer available in ContentStudio. Use ContentPen for all new blog content. Editing existing posts will also be removed soon."
- [ ] Phase 2 edit banner CTA: "Switch to ContentPen" that opens the `contentpen-cta-modal`
- [ ] Editing and saving existing blog posts still works
- [ ] No deployment needed on April 30 — the date check handles the transition automatically

**General:**
- [ ] The existing ContentPen CTA modal continues to work correctly — super admin sees CTA button, team member sees contact-admin message
- [ ] Scheduled and published blog posts continue to work without any disruption in both phases
- [ ] All new user-facing strings use i18n keys — add to all locale directories
- [ ] No changes to blog publishing backend
- [ ] Colors use theme-aware classes — banner uses standard warning/amber styling

---

### Mock-ups:

N/A — use existing `ContentPenCTAModal` for all redirects. Banner in blog composer uses `Alert` component with warning variant.

---

### UI Copy:

**Phase 1 banner (before April 30, inside blog composer for old users):**
- Icon: `AlertTriangle` (amber)
- Text: "Blog composer is moving to ContentPen on April 30, 2026. We recommend switching to ContentPen now for a better blog writing experience with AI-powered SEO, internal linking, and more."
- CTA: "Try ContentPen" → opens `contentpen-cta-modal`

**Phase 2 banner (after April 30, inside blog composer when editing existing posts):**
- Icon: `AlertTriangle` (amber)
- Text: "Creating new blog posts is no longer available in ContentStudio. Use ContentPen for all new blog content. Editing existing posts will also be removed soon."
- CTA: "Switch to ContentPen" → opens `contentpen-cta-modal`

**No new copy needed for the ContentPen CTA modal** — it already has all the copy for super admin and team member flows.

---

### Impact on existing data:

No data or schema changes. Blog posts in the database are unaffected. Scheduled and published posts continue to work. This is a frontend-only access control change.

---

### Impact on other products:

- Web App: Blog composer entry points across publisher, dashboard, header, planner, notifications
- Mobile apps: No impact — mobile apps do not have blog composer
- Chrome extension: No impact

---

### Dependencies:

None — the ContentPen CTA modal and all required infrastructure already exist.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (banner should display correctly on narrower widths inside the blog composer)
- [ ] Multilingual support (banner copy uses i18n — add keys to all locale directories)
- [ ] UI theming support (use `Alert` component and theme-aware classes for banner)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
