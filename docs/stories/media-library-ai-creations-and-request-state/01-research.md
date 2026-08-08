# Research — AI Creations in the Content Library, and Media Library request state

Item 12 of the 7 Aug 2026 backlog batch. Corrected after PO review: the original draft framed this as async AI generation job tracking. That was wrong. The real problem is that there is no server-side way to ask for AI-generated media, so the frontend asks for everything and filters it on the client, re-requesting pages in a loop until it has enough.

## Fault 1: the API cannot filter by AI-generated

The frontend does send the filter. `useMediaLibraryFetch.ts` → `applyLibraryScopeToPayload()`:

```
if (normalizedType === 'all_uploads')    payloadFilters.is_ai_generated = false
if (normalizedType === 'ai_creations' ||
    normalizedType === 'my_ai_creations') payloadFilters.is_ai_generated = true
if (ai_creations && pill === 'clips')     payloadFilters.type = 'video'
```

The backend drops it.

- `contentstudio-backend/app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php::fetchMediaAssets()` (line 653) adds only `workspace_ids` and `user_id` to `$filters` before calling the repository. `is_ai_generated` is never mapped. Every other occurrence of `is_ai_generated` in that controller (lines 70, 170, 337, 459, 2234) is in an **upload/create** path, setting the flag on new media, not reading it as a filter.
- `contentstudio-backend/app/Repository/Utilities/MediaRepository.php::findMediaAssets()` applies exactly: `brand_asset`, `fetch_archived`, `folder_id`, `search` (name LIKE), `type` (mime_type LIKE), `sort`, `usage_filter`, `page`, `limit`. **There is no `is_ai_generated` clause anywhere in the list path.** The only three occurrences in the whole repository file are line 588 and line 1161 (both *setting* the flag) and line 701 (the stats count).

So a request for AI Creations returns the workspace's entire media library, unfiltered, and the client throws most of it away.

Note the partial exception: the **clips** pill also sends `type: 'video'`, which *is* honoured (`mime_type LIKE '%video%'`). So clips gets a server-narrowed set of all videos, then filters to AI ones client-side. It is the only pill with any server-side narrowing at all.

## Fault 2: client-side filtering forces a page-refetch loop

`useMediaLibraryFetch.ts::fetchMediaAssets()` filters each returned page:

```
const scopedData = filterAssetsByUploader(filterAssetsBySection(media.data))
```

Because a server page of 40 can yield very few or zero in-scope items, the function then **recursively re-requests further pages** until it has enough (lines 439 to 475):

```
const targetScopedCount = clear ? limit : mediaCountBeforeFetch + 1

if (fillScopedPage && (isSectionScopedType(...) || selectedUploaderIds.length > 0) && ...) {
  while (generation === fetchGeneration && mediaAssets.value.length < targetScopedCount
         && !mediaMeta.value.completed && fillPages < MAX_FILL_PAGES && emptyStreak < 2) {
    await fetchMediaAssets(false, folderId, filters, page, limit, false)
    ...
  }
}
```

`MAX_FILL_PAGES = 5` (line 75) and `emptyStreak < 2` are the only things bounding it. Both are scar tissue: the comment block at lines 415 to 426 documents two prior failure modes in the completion logic, in its own words —

> - 0 total results: `to` is null, so `total === to` never matched and **the fill loop fetched empty pages forever**.
> - `total` over-reporting the real row count (e.g. it counts unfiltered/stale rows while the AI/scoped query returns fewer): `to` never reaches `total`, so **pagination ran past the end**.

That second bullet is the backend gap described from the other side. `total` is the count of *all* media because the server never applied the AI filter, while the rendered list is the filtered subset, so the two can never agree.

**This affects three sections, not just AI Creations.** `isSectionScopedType()` returns true for `all_uploads`, `ai_creations` and `my_ai_creations`. All uploads sends `is_ai_generated: false`, equally ignored, so a workspace with lots of AI media pays the same cost on its default view.

`my_ai_creations` is worse again: ownership is filtered purely client-side via `isOwnedByCurrentUser()` (comparing `created_by` / `user_id` against the current user) with no server-side equivalent in that path, so it stacks a second client-side narrowing on top of the first.

## Fault 3: pagination metadata describes a different set than the one displayed

`mediaMeta` carries `current_page`, `count`, `total`, `completed`, where `total` comes straight from the server's unfiltered pagination. The `completed` flag had to be rewritten to lead on `rawPageCount < limit` precisely because `to`/`total` cannot be trusted:

```
completed: rawPageCount < limit || (media.to ?? 0) >= (media.total ?? 0)
```

Any UI that wants to show "showing X of Y" for a scoped section has no honest Y today.

## Fault 4: the AI Creations sidebar count covers one data source of three

AI Creations has three pills. `components/FiltersBar.vue` (lines 169 to 193):

```
type AiCreationsPill = 'ai_studio' | 'clips' | 'ai_posts'   // default 'ai_studio'
```

They are served by **two different systems**:

| Pill | Source | How it is filtered |
|---|---|---|
| AI Studio | `Media` collection, via the media library list endpoint | client-side `isAiGeneratedAsset()` |
| Video Clips | `Media` collection, same endpoint | server `type=video`, then client-side `isAiGeneratedAsset()` |
| AI Posts | **A different collection entirely** — `useAiPostsInfiniteQuery` from `@modules/publisher/ai-content-library/queries`, rendered by `ContentLibraryAiPosts.vue` | not the media endpoint at all |

`useMediaLibraryFetch` explicitly bails out for AI Posts (`isAiPostsLibrarySection()` early-returns with an empty list, lines 304 to 310), because that pill is somebody else's data.

The sidebar count comes from one number:

- `components/SideBar.vue`: `aiCreationsCount = toCount(props.filesCount?.aicreations) ?? 0`
- fed by `MediaRepository.php` line 701: `$aiCreations = (clone $query)->where('is_ai_generated', true)->count();`

That counts **media rows flagged AI-generated**. It therefore includes AI Studio and Video Clips, and **excludes AI Posts entirely**, because those live in another collection. The label says "AI Creations" and the number describes a subset.

Two knock-on effects:

1. **The pills do not partition.** `filterAssetsBySection()` returns *all* AI-generated media for the `ai_studio` pill and *AI-generated video* for `clips`. So clips is a strict subset of what AI Studio shows, and the same asset appears under both. Whatever the count is meant to mean, the three pills currently overlap.
2. **All uploads inherits the error.** `SideBar.vue` derives `allUploadsCount` as `all_uploads ?? uploads`, falling back to `Math.max(allCount - aiCreationsCount, 0)`. A wrong `aicreations` corrupts the All uploads count through that fallback.

## Fault 5: the empty state wins the race with the first request

Independent of the above, and still valid from the original research.

`isMediaFetching` starts `false` and only flips true inside `fetchMediaAssets()`. Nothing calls it during setup; every fetch is triggered by a non-immediate route watcher (`route.query.type`, `route.query.folderId`, `route.query.aiContentType`, `combinedFilters`, active workspace id), which fire after the first paint. Meanwhile `MediaLibraryMain.vue:66` gates the empty state on:

```
v-if="allMediaAssets.length === 0 && isMediaFetching === false"
```

Both conditions hold on the first frame, so the user sees "no media yet" and its call to action, then skeletons, then content. Every section, folder and filter change repeats it. There is no error branch in the template at all, only loading and empty.

The fill loop makes this worse: `isMediaFetching` is only cleared on the outer `clear` call (line 477), so the loading state spans the entire recursive run, and a slow fill keeps the user waiting with no indication that anything is progressing.

## What needs to change

**Backend**
- Honour an AI-generated filter on the media list endpoint, so a request for AI Creations returns AI media and nothing else.
- Distinguish the categories server-side, so AI Studio and Video Clips are separately requestable rather than overlapping client-side derivations.
- Return pagination metadata that describes the filtered set.
- Fix the AI Creations count so it reflects what the section actually contains across all three pills, including AI Posts from the other collection.

**Frontend**
- Consume the real filters and delete the fill loop, `MAX_FILL_PAGES`, `emptyStreak` and the client-side `filterAssetsBySection` narrowing.
- Fix the request/loading lifecycle so the loader leads and the empty state is only reachable after a settled request.
- Show a per-pill count that matches the section.

## Decisions needed before implementation

- **What does the AI Creations count mean?** Sum of all three pills, or a per-pill count next to each pill? Summing across two collections means one more query on a sidebar that renders on every page load.
- **Should AI Studio exclude clips?** Today it includes them. If the three pills are meant to partition, AI Studio needs a positive definition (generated by AI Studio) rather than "any AI media", which needs a field that distinguishes origin. Confirm whether `Media` already records which surface generated an asset, or whether that has to be added and backfilled.
- **Backfill.** If a new origin field is introduced, existing AI media has to be classified. Confirm whether existing rows can be attributed reliably or whether they land in a default bucket.

## Files involved

Frontend:
- `contentstudio-frontend/src/modules/publish/components/media-library/composables/useMediaLibraryFetch.ts`
- `contentstudio-frontend/src/modules/publish/components/media-library/MediaLibraryMain.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/SideBar.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/FiltersBar.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/ContentLibraryAiPosts.vue`
- `contentstudio-frontend/src/api/media-library.ts`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/queries/` (the AI Posts source)

Backend:
- `contentstudio-backend/app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php` — `fetchMediaAssets()` at line 653
- `contentstudio-backend/app/Repository/Utilities/MediaRepository.php` — `findMediaAssets()` at line 731, stats block at lines 690 to 730

## Mobile

None. The Media Library and AI Creations are web only.
