# Research: Blog Composer Sunset → ContentPen Migration

## Current State

Blog composer access is conditionally controlled by two flags on the super admin's user record:
- `redirect_to_content_pen` — if true, "create blog post" entry points show ContentPen CTA modal instead of opening composer
- `linked_to_content_pen` — if true (AND redirect is true), goes directly to `app.contentpen.ai`

**Old users** have `redirect_to_content_pen: false` (or unset), so they still get the regular blog composer. **New users** already get the ContentPen modal.

## Entry Points That Open Blog Composer

1. **Publisher sidebar** (`SidebarMain.vue`) — compose split button → "Blog Post" dropdown
2. **Dashboard** (`WelcomeRow.vue`) — "Blog Post" quick action
3. **Top header** (`TopHeaderBar.vue`) — "Create Post" → "Blog Post" / EventBus `open-new-blog-composer`
4. **Composer dropdown** (`ComposerDropdown.vue`) — "Composer Article" option
5. **AI Content Library** (`useAIPostGeneration.ts`) — "Create blog post" from AI post
6. **Planner post click** — editing existing blog post on calendar → routes to `composerBlog`
7. **Notification click** (`useNotificationHandler.js`) — blog-related notification → routes to `composerBlog`
8. **Blog Websites settings** (`TopHeaderBar.vue`) — conditionally hidden via `shouldRedirectBlogToContentPen`

Entries 1-5 already have ContentPen redirect logic for new users. Entry 6-7 go directly to the blog composer route.

## Key Files

- `contentstudio-frontend/src/composables/useComposerPost.js` — main create/edit routing logic with ContentPen checks
- `contentstudio-frontend/src/composables/useWorkspaceMembers.js` — `shouldRedirectBlogToContentPen`, `shouldDirectLoginToContentPen`, `isLinkedToContentPen`
- `contentstudio-frontend/src/components/ContentPenCTAModal.vue` — existing modal (super admin CTA vs team member message)
- `contentstudio-frontend/src/modules/composer/components/blog/Blog.vue` — blog composer component
- `contentstudio-frontend/src/modules/common/components/dropdowns/ComposerDropdown.vue` — "Composer Article" option
- `contentstudio-frontend/src/components/layout/TopHeaderBar.vue` — header blog entry points
- `contentstudio-frontend/src/components/dashboard/WelcomeRow.vue` — dashboard blog entry point
- `contentstudio-frontend/src/modules/publisher/components/ComposeActionsDropdown.vue` — publisher sidebar blog option
