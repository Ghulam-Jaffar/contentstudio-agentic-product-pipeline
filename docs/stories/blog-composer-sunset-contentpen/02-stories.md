# Stories: Blog Composer Sunset → ContentPen Migration

---

## Story 1: [FE] Sunset blog composer with two-phase ContentPen migration

### Description:

As the ContentStudio product team, we want to sunset the blog composer and migrate all users to ContentPen so that blog functionality is consolidated into a dedicated, more capable product.

The blog composer is being deprecated in two phases:

**Phase 1 (Now → April 30, 2026):**
- **Creating new blog posts is blocked for all users.** Remove the `shouldRedirectBlogToContentPen` conditional — all "create blog post" entry points now show the existing ContentPen CTA modal (`contentpen-cta-modal`) regardless of the super admin's `redirect_to_content_pen` flag.
- **Editing existing blog posts is still allowed**, but a persistent non-dismissable sunset banner is shown at the top of the blog composer directing users to ContentPen.
- **Scheduled/published blog posts are unaffected** — they continue to work normally.

**Phase 2 (April 30, 2026+):**
- **Editing is also blocked.** All blog composer entry points (create and edit) redirect to the ContentPen CTA modal.
- A date check (`new Date() >= new Date('2026-04-30')`) controls the phase transition — no deployment needed on April 30.

**Entry points that need changes:**
- `useComposerPost.js` — main create/edit routing logic (remove `shouldRedirectBlogToContentPen` condition for create, add date check for edit)
- `ComposerDropdown.vue` — "Composer Article" option (add ContentPen redirect)
- `Blog.vue` — blog composer component (add sunset banner for Phase 1, block access for Phase 2)
- `useNotificationHandler.js` — blog notification click (add Phase 2 block)
- `TopHeaderBar.vue` — Blog Websites settings link (hide for all users, not just those with redirect flag)

**Entry points already handled (no changes needed):**
- `SidebarMain.vue`, `WelcomeRow.vue`, `useAIPostGeneration.ts` — already use `openBlogComposer` which goes through `useComposerPost.js`

**Components to use:**
- `Alert` from `@contentstudio/ui` — sunset banner in blog composer
- `Button` from `@contentstudio/ui` — CTA in banner
- `Icon` from `@contentstudio/ui` — warning icon in banner
- Existing `ContentPenCTAModal` — already handles super admin vs team member experience

---

### Workflow:

**Phase 1 (Now → April 30):**
1. User clicks any "create blog post" entry point (publisher sidebar, dashboard, header, composer dropdown, AI library)
2. Instead of opening the blog composer, the ContentPen CTA modal opens
3. Super admin sees "Try ContentPen" button that creates a session and redirects to `app.contentpen.ai`
4. Team member sees a message to contact their super admin
5. User navigates to an existing blog post to edit it (from planner, notifications, or direct URL)
6. Blog composer opens with a persistent amber banner at the top explaining the sunset
7. User can still edit and save the post
8. User clicks "Learn more about ContentPen" in the banner — ContentPen CTA modal opens

**Phase 2 (April 30+):**
9. User tries to edit an existing blog post
10. Instead of opening the blog composer, the ContentPen CTA modal opens
11. Optional toast: "Blog composer is no longer available. Use ContentPen to create and edit blog content."

---

### Acceptance criteria:

**Phase 1 — New posts blocked:**
- [ ] All "create new blog post" entry points show the ContentPen CTA modal for ALL users — not just users with `redirect_to_content_pen: true`
- [ ] The `shouldRedirectBlogToContentPen` condition is removed from create-new-post paths — all users get the modal
- [ ] `ComposerDropdown.vue` "Composer Article" option shows the ContentPen CTA modal instead of opening the blog composer
- [ ] Blog Websites settings link in the header is hidden for all users (not just those with the redirect flag)

**Phase 1 — Edit still allowed with banner:**
- [ ] When editing an existing blog post, a persistent non-dismissable banner appears at the top of the blog composer
- [ ] The banner uses `Alert` component with amber/warning styling
- [ ] Banner copy: "Blog composer is moving to ContentPen. Editing existing blog posts will no longer be available after April 30, 2026. Move to ContentPen for a better blog writing experience with AI-powered SEO, internal linking, and more."
- [ ] Banner has a CTA button: "Learn more about ContentPen" that opens the `contentpen-cta-modal`
- [ ] The banner cannot be dismissed or closed
- [ ] Editing and saving the blog post still works normally
- [ ] Scheduled and published blog posts continue to work without any disruption

**Phase 2 — Everything blocked (April 30+):**
- [ ] A date check (`new Date() >= new Date('2026-04-30')`) controls the phase transition
- [ ] After April 30, editing existing blog posts also shows the ContentPen CTA modal instead of opening the composer
- [ ] Blog notification clicks that would open `composerBlog` route also redirect to the ContentPen CTA modal after April 30
- [ ] No deployment needed on April 30 — the date check handles the transition automatically
- [ ] A toast is shown when the user is redirected: "Blog composer is no longer available. Use ContentPen to create and edit blog content."

**General:**
- [ ] The existing ContentPen CTA modal continues to work correctly — super admin sees CTA button, team member sees contact-admin message
- [ ] All new user-facing strings use i18n keys — add to all locale directories
- [ ] No changes to blog publishing backend — scheduled posts continue to go out
- [ ] Colors use theme-aware classes — banner uses standard warning/amber styling

---

### Mock-ups:

N/A — use existing `ContentPenCTAModal` for all redirects. Banner in blog composer uses `Alert` component with warning variant.

---

### UI Copy:

**Sunset banner (Phase 1, inside blog composer when editing):**
- Icon: `AlertTriangle` (amber)
- Text: "Blog composer is moving to ContentPen. Editing existing blog posts will no longer be available after April 30, 2026. Move to ContentPen for a better blog writing experience with AI-powered SEO, internal linking, and more."
- CTA: "Learn more about ContentPen" → opens `contentpen-cta-modal`

**Phase 2 toast (when edit is blocked):**
- "Blog composer is no longer available. Use ContentPen to create and edit blog content."

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
- [ ] Multilingual support (banner copy and toast use i18n — add keys to all locale directories)
- [ ] UI theming support (use `Alert` component and theme-aware classes for banner)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
