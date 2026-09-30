# Research: Home "Recent posts" View All goes to Content Library AI Creations

## Current State

- The Home dashboard's recent AI posts section ("Recent brand creations") has two header actions: **Generate post** (opens the bulk schedule AI modal) and **View All**.
- **View All** is a `router-link` to the `ai-content-library-posts` route, which is the standalone AI Library page under Publisher.
  - `contentstudio-frontend/src/components/dashboard/RecentPostCreations.vue:29-35` (link), `:120-123` (`libraryRoute` computed)
  - Copy key: `dashboard.recent_brand_creations.view_all` = "View All" (`src/locales/en/dashboard.json:176`)
- Content Library (route name `media-library`) already has an **AI Creations** sidebar section with three pills: **AI Studio** (default), **Clips**, **AI Posts**.
  - Section is selected by the query `type=ai_creations` (`SideBar.vue:397`, `MediaLibraryMain.vue:584-596`)
  - Pill is selected by `aiContentType` = `ai_studio` | `clips` | `ai_posts`, defaulting to `ai_studio` when missing or invalid (`FiltersBar.vue:169-221`)
  - `aiContentType=ai_posts` renders `ContentLibraryAiPosts.vue`, the same AI posts the Home section previews (`MediaLibraryMain.vue:347`, `:61-63`)

## What Needs to Change

- Point View All at Content Library, AI Creations section, AI Posts pill: `{ name: 'media-library', query: { type: 'ai_creations', aiContentType: 'ai_posts' } }` (plus workspace param if the route needs it, match how other links to `media-library` are built).
- Landing on AI Posts rather than the default AI Studio pill, because the Home cards are AI posts. Opening AI Studio would show generated media, not the posts the user was looking at.
- No copy change, no BE change, no new analytics event (navigation only).

## Files Involved

- `contentstudio-frontend/src/components/dashboard/RecentPostCreations.vue`
