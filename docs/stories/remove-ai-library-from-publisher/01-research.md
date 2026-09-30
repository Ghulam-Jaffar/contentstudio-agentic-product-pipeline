# Research: Remove the AI Library from the Publisher module

## Current State

The AI Library is two routes under Publisher, shown in the Publisher sidebar as **AI Posts** and **Brand Settings**:

- Routes: `ai-content-library` (redirects to posts), `ai-content-library-posts`, `ai-content-library-profile` (`contentstudio-frontend/src/modules/publisher/config/routes/publisher.ts:180-208`)
- Sidebar entries: `contentLibraryItems` (`src/modules/publisher/components/SidebarMain.vue:916-931`), labels `publisher.sidebar.ai_posts` and `publisher.sidebar.brand_settings`
- Shell: `src/modules/publisher/ai-content-library/views/AIContentMain.vue`. It shows the Brand Knowledge first-time setup when setup is incomplete, the empty-state `GeneratePostPopup`, and mounts `CustomGenerateModal` and `PostSettingsModal`

### What the AI Library does today, and where each piece lives otherwise

| AI Library capability | Available elsewhere? | Where |
|---|---|---|
| Generate posts in bulk | Yes | AI Studio, **Bulk schedule** tool (`bulk-schedule-ai-modal`, `useAiToolCatalog.ts:202`), also Home "Generate post" and Composer's compose dropdown (`ComposeActionsDropdown.vue:237`). The bulk modal has its own configure step |
| Brand Knowledge setup and editing | Yes | Settings, `brand-settings` route (`modules/setting/config/routes/setting.ts:214`). The AI Library's own "Brand Knowledge" button already routes there (`AIPostHeader.vue:498-503`). AI Studio sends users without setup there too (`useAiToolLauncher.ts:81-87`) |
| Browse AI posts | Yes | Content Library, AI Creations, AI Posts pill (`ContentLibraryAiPosts.vue`), same data source (`useAiPostsInfiniteQuery`) |
| Search AI posts | Yes | Content Library search (`route.query.search`) |
| Preview, prev/next | Yes | Content Library uses the same `PostPreviewModal` |
| Delete single + bulk, bulk schedule, bulk save as draft | Yes | Content Library `PostSelectionBar` + delete confirmation |
| Card menu: Schedule with AI, open in composer, save as draft | Yes | Same `AIPostCard` / `AIPostCardMenu` |
| **Per-post "Schedule" and "Add to queue"** (card menu + preview modal) | **No, gap** | Content Library does not pass `show-schedule-cta` or handle `@schedule` / `@add-to-queue` on `AIPostCard` (`ContentLibraryAiPosts.vue:47-57`) or `PostPreviewModal` (`:77-84`). Home does (`RecentPostCreations.vue:48-70`, via `useSchedulePost`) |
| **"Default post generation settings"** (the page the sidebar calls "Brand Settings") | **Read only, gap** | `AIPostSettings.vue` → `PostSettingsForm.vue` (platform, language, post type, posts count, caption length, emoji, hashtags, image style, aspect ratio, image model) saved via `savePostSettings` / `savePostSettingsApi` (`useAIPostSettings.ts:298-400`, `api/ai-content-library.ts:125`) into `AIUserProfile.post_generation_settings`. The Bulk schedule modal's `CustomGenerateForm` has the same fields (`customPostGeneration`) and is **pre-filled from the saved defaults** via `initializeFormFromProfile` (`useAIPostSettings.ts:287-290`), but cannot save them. PO decision: saving moves into Bulk schedule |
| Grid/list toggle | Partly | Content Library AI Posts is grid only |
| Text/image credit usage in header | Elsewhere | Billing usage. Not needed on the posts view |

### Other references to the AI Library routes

- `src/modules/publish/components/media-library/MediaLibraryMain.vue:970-980`: `handleEmptyStateAction` pushes to `ai-content-library-posts`. The CTA ("Generate Content", `publisher.media_library.empty_state.ai_cta`) only renders in the media grid, so it shows on the **AI Studio** and **Clips** pills and **My AI Creations**, never on **AI Posts** (`showEmptyStateCta`, `:463-468`). The AI Posts pill has its own empty state in `ContentLibraryAiPosts.vue:20-30` with no button. Decision: media pills → AI Studio route (`modules/ai-studio/config/routes.ts:12`); AI Posts pill → new "Generate posts" button opening `bulk-schedule-ai-modal` with the `requiresBrandSetup` gate, reuse the launcher logic in `useAiToolLauncher.ts:81-96`.
- `src/components/dashboard/RecentPostCreations.vue:120-123`: Home "View All". Covered by CONT-4154.
- `src/modules/billing/composables/useApiCentricPlan.ts:31-32`: route names in the API Plan hidden list. Clean up.
- `src/modules/publisher/ai-content-library/composables/useAIPostGeneration.ts:340`: `AI_POSTS_ROUTES` gates the AI posts query to AI Library routes. Check nothing outside the AI Library relies on it before removing.
- `src/modules/AI-tools/composables/useBotMessageView.ts:517`: `isAiContentLibraryRoute` switches the AI chat "add to editor" behavior on the AI Library page. Becomes dead code.
- No links to `ai-content-library` in `contentstudio-backend/` (emails, notifications), `contentstudio-ai-agents/` or `contentstudio-flutter/`.

### What must NOT be deleted

`src/modules/publisher/ai-content-library/` is shared code. Content Library, Home, AI Studio bulk modal, AI chat, Planner, Composer and Brand Knowledge import its components, composables and queries (`AIPostCard`, `PostPreviewModal`, `useAIPostGeneration`, `useSetup`, `useRecentPosts`, `useAiPostsInfiniteQuery`, `BrandKnowledgeEditor`, types). Remove only the routes, the views (`AIContentMain`, `AIPosts`, `AIProfile`), the sidebar entries and anything left with no importer. Leave the folder move to a later refactor.

## What Needs to Change

1. Add per-post **Schedule** and **Add to queue** to Content Library AI Posts (card menu + preview), same as Home.
2. Add a "Save as my default settings" checkbox to the Bulk schedule setup step. When ticked, call `savePostSettings` with the `customPostGeneration` values before generating. Does not block generation on failure.
3. Empty states: AI Posts pill gets a "Generate posts" button → Bulk schedule; media pills' "Generate Content" → AI Studio.
4. Remove the AI Posts and Brand Settings items from the Publisher sidebar.
5. Replace the three routes with redirects: posts → Content Library AI Creations / AI Posts, profile → Brand Knowledge settings, bare `ai-content-library` → Content Library AI Posts. Keeps bookmarks and old links working.
6. Clean up route-name references (API Plan list, AI chat route check, query gate).
7. Remove unused views, including `PostSettingsModal` and `AIPostSettings` once saving lives in Bulk schedule (keep `PostSettingsForm` only if still imported).

## PO decisions (2026-09-29)

- Default post generation settings are managed from the Bulk schedule window, not the AI Library.
- Empty states: pill-aware (AI Posts → Bulk schedule, AI Studio/Clips/My AI Creations → AI Studio).

## Files Involved

- `contentstudio-frontend/src/modules/publisher/config/routes/publisher.ts`
- `contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/ContentLibraryAiPosts.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/MediaLibraryMain.vue`
- `contentstudio-frontend/src/modules/billing/composables/useApiCentricPlan.ts`
- `contentstudio-frontend/src/modules/AI-tools/composables/useBotMessageView.ts`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/composables/useAIPostGeneration.ts`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/views/*`
- `contentstudio-frontend/src/modules/AI-tools/bulk-schedule/BulkScheduleAIModal.vue`, `src/modules/publisher/ai-content-library/components/form/CustomGenerateForm.vue`
