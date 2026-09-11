# Research: Marketing feedback epic

**Date:** 2026-09-11
**Source:** marketing team feedback batch, six asks
**Deliverable:** `02-epic-and-stories.md`

This doc stays local. Codebase paths, entry points, gotchas and open questions live here and never in a story body.

---

## 1. The six asks, as given

| # | Ask (verbatim intent) | Surface |
|---|---|---|
| 1 | AI chat should have the context of previous messages and the current chat history | AI chat (web + Flutter) |
| 2 | Users can highlight an AI response or part of their own text and reply to / address that specific passage, like Claude and ChatGPT | AI chat (web) |
| 3 | Images generated in AI Studio (tools and AI chat) should be uploaded to the user's media library | AI Studio, media library |
| 4 | When the AI mentions a ContentStudio feature, offer a CTA that deep-links the user straight to it | AI chat |
| 5 | Bulk download media from the Content Library | Media library |
| 6 | Multiple first comments, plus conditional scheduling of those comments, from the composer | Composer, publishing |

Ask 6 is deliberately held back: the user wants it finalized with a mockup/prototype before stories are written, and has their own ideas to feed in. Everything through ask 5 is specced now.

---

## 2. Backlog dedupe

Checked all 38 dirs in `docs/features/` and 202 in `docs/stories/` before writing anything new. Nothing in the backlog covers any of the six asks. The near misses, and why they are not duplicates:

| Existing deliverable | Relationship |
|---|---|
| `docs/stories/ai-chat-component-rendering/` | Renders **tool results** as components (stat tiles, charts, tables, account cards) via a component registry. Ask 2 and ask 4 layer on top of assistant **prose**, not tool results. Ask 4's inline feature CTA is a natural sibling to that registry and should reuse its fallback discipline, but no story there covers feature mentions or text selection. |
| `docs/stories/media-library-ai-creations-and-request-state/` | Assumes AI creations already exist as media rows and fixes the **filter, count and loading state** over them. Ask 3 is the upstream half: getting AI Studio output into the library as a media row in the first place. Ask 3 is a hard dependency of that epic's value, not a duplicate of it. |
| `docs/stories/ai-studio-composer-phase-1/`, `docs/features/ai-chat-skills/`, `docs/stories/ai-surfaces-architecture/` | AI surface plumbing and composer handoff. No conversation-memory, quote-reply, or feature-CTA scope. |
| `docs/stories/flutter-composer-web-parity/`, `docs/stories/ai-powered-composer-customization/`, `docs/stories/public-api-publishing-gaps/` | Mention first comment only as an existing single-comment field to reach parity with or expose. None of them add multiple comments or comment scheduling. |
| `docs/features/composer-v2-refactor/` | Touches first comment and bulk actions in passing as part of a refactor. Not the feature work in asks 5 and 6. |

---

## 3. Analytics event baseline (Usermaven)

Pulled the live catalog from `contentstudio-frontend/src` (`grep -rhoP "userMaven\.track\(\s*['\"][a-z0-9_]+['\"]" src`):

```
account_cancellation_feedback, addons_limits_updated, ai_post_compose, ai_post_feedback,
ai_post_regenerated, ai_posts_generated, audio_attached, brand_profile_created,
brand_voice_setup_cta_clicked, brand_voice_toggle_state, composer_sidebar_tab_clicked,
connected_social_accounts, customize_ai_caption_edited_after_generation,
customize_ai_generate_clicked, customize_ai_generate_failed, customize_ai_generate_success,
customize_ai_menu_opened, customize_ai_per_platform_regenerate, easy_connect_accounts_connected,
easy_connect_link_created, intent_dialog_shown, intent_dialog_submitted, lander_signup_completed,
lander_signup_intent_captured, language_change, media_note_added, pageview, team_member_invited
```

**Finding: AI chat is currently uninstrumented.** There is no `ai_chat_*` event in the catalog at all. The `ai_post_*` and `customize_ai_*` families cover the AI Content Library and composer customization, not chat. So every chat story in this epic introduces a genuinely new event name rather than reusing one, and the epic is the right moment to establish the `ai_chat_*` prefix.

New events this epic introduces (see the stories for payloads):

- `ai_chat_quote_reply_sent` — ask 2
- `ai_chat_feature_cta_clicked` — ask 4
- `ai_media_saved_to_library` — ask 3
- `media_bulk_download_requested` — ask 5

Deliberately **not** tracked: opening a new chat, scrolling history, hovering a selection, opening the bulk-action bar. Per guidelines section 17 these are trivial UI interactions and would only add noise.

---

## 4. Prototype surface for ask 6

`cs-prototypes/` is a Next.js 15 / React 19 / Mantine 8 / Tailwind 4 app with an established per-feature convention: one folder under `cs-prototypes/app/features/<slug>/`, a default-exported `page.tsx`, registration in `app/features/_meta.ts`, and a per-feature `prompt.md`. Existing prototypes include `ai-chat`, `planner`, `inbox-threads`, `listening`. This is where the ask 6 multi-comment composer mockup should be built once the behavior is finalized.

---

## 5. Competitor pattern for ask 2 (quote reply)

The marketing ask names Claude and ChatGPT explicitly, so the target pattern is well established rather than novel. Confirmed behavior of the ChatGPT web implementation:

- The user drags a selection across any part of a message. A small floating bubble with a quote icon appears anchored to the selection.
- Clicking it inserts the selected passage into the chat input as a visually distinct quoted block, above the user's cursor, and focuses the input.
- The user then types a follow-up that is understood to be about the quoted passage, and the quote travels with the message in the transcript.
- Selection works on assistant output; selecting your own earlier text works the same way.
- Known limitation worth designing around: in the ChatGPT macOS app a selection cannot span both prose and a fenced code block, and the selection is clipped at the block boundary. The web app does not have that limit. For ContentStudio the equivalent boundary risk is a selection that spans prose and a rendered tool-result component from the AI chat component registry.

Design implications carried into the story:

- The quote is a first-class part of the outgoing message, not a string prepended to the user's text, so it can be rendered as a block in the transcript and sent to the model as explicit reference material.
- Anchor by quoted text plus source message id. Character offsets into a markdown-rendered DOM are brittle across a re-render or a streamed update.
- Selection inside a rendered component (a chart, a stat tile, a table) should fall back to plain text extraction or be suppressed, never produce a broken anchor.

Sources: [ChatGPT Reply feature demo](https://x.com/ai_for_success/status/1786306348881190962), [Exploring ChatGPT's Quote feature](https://www.autoage.it/en/exploring-chatgpt-s-quote-feature), [macOS code-block selection limitation](https://community.openai.com/t/quoting-feature-in-macos-app/763816)

---

## 6. Ask 1 — AI chat conversation context: current state

### Architecture

Three tiers, and the important surprise is that **Laravel, not the agent tier, is the system of record for conversation history**.

- **Frontend** sends only the newest message plus a `chat_id`. No message array is sent. `contentstudio-frontend/src/modules/AI-tools/utils/buildChatStreamPayload.ts`.
- **Laravel** rebuilds a `chat_history` array from MongoDB on every turn and forwards it. `contentstudio-backend/app/Http/Controllers/AI/AIController.php` `processAgnoAgent()` (history assembly around L1240-1370), payload built by `app/Helpers/Ai/AiChatHelper.php` `buildWorkflowInputs()` L419.
- **ai-agents** flattens that array into a single labelled string and passes it as the Agno Team's `input=`. `contentstudio-ai-agents/src/api/routers/streaming_router.py` `_build_team_input()` L246.

Agno's native history (`add_history_to_context`) is **deliberately off**, for two documented reasons: passing `List[Message]` made the coordinator re-execute prior turns, and native history reads off the Team's `db=`, which is `RedisDb(expire=1800)`, so enabling it would give every conversation a 30-minute lifetime. The Redis session store exists only so a paused human-in-the-loop run can resume. Do not "fix" context by flipping that flag.

### Key paths

| Concern | Path |
|---|---|
| FE payload builder | `contentstudio-frontend/src/modules/AI-tools/utils/buildChatStreamPayload.ts` |
| FE SSE stream | `contentstudio-frontend/src/modules/AI-tools/composables/useAIChatStream.ts` |
| FE thread lifecycle | `contentstudio-frontend/src/modules/AI-tools/composables/useAiChatThread.ts` |
| FE engine | `contentstudio-frontend/src/modules/AI-tools/composables/useAiChatEngine.ts` |
| FE chat API | `contentstudio-frontend/src/api/ai-chat.ts` (`fetchChatByIdApi` supports `before`) |
| BE routes | `contentstudio-backend/routes/web/ai.php` |
| BE history read | `contentstudio-backend/app/Repository/AiChat/AiChatMessagesRepo.php` `getChatHistory()` |
| BE message window | `contentstudio-backend/app/Repository/AiChat/AiChatRepo.php` `getChatMessages()` (cursor paging) |
| Agent history packer | `contentstudio-ai-agents/src/api/routers/streaming_router.py` `_build_team_input()` L246 |
| Team construction | `contentstudio-ai-agents/src/orchestration/team.py` `get_content_team()`, `Team(...)` L122 |
| Contract reference | `contentstudio-ai-agents/tasks/agentic-architecture/INDEX.md` §4.6 — the `chat_history` contract, best single reference |

### Data model

- MongoDB `ai_chat` — `_id` (this **is** the `session_id` sent to ai-agents), `workspace_id`, `user_id`, `is_active`, `title`, `image_meta`, `analytics_context`, `updated_at`, soft deletes.
- MongoDB `ai_chat_messages` — `chat_id`, `workspace_id`, `user_id`, `role` (`user` or `assistant` only), `content` (array of typed blocks: `text`, `image_url`, `video_url`, `ai_library_post`, `asked`, `video_job`, `followup_actions`), `mentions`, `reference_clips`, `credits_consumed`, `post_id`, `job_id`, `job_status`, `has_error`, timestamps.
- Postgres in ai-agents has **no** chat message table. `mcp_workflow_sessions` is workflow state, not history.

### The actual gaps behind the marketing ask

The ask reads as "chat has no memory". It does have memory. What it has is a memory that quietly drops things, in six distinct ways:

1. **Hard 10-message window, double-capped, nothing summarized.** Laravel caps at 10 (`getChatHistory(..., 10)`, a literal in `AIController.php` around L1240), and `_build_team_input` caps again at `max_turns=10`. Turn 11 forgets turn 1 permanently. No config, no flag.
2. **Every history turn except the newest assistant turn is cut to 800 chars** (`HISTORY_TURN_CHARS = 800`; the newest assistant turn gets `HISTORY_LATEST_ASSISTANT_CHARS = 6000`). A long user brief, or the second-newest assistant turn holding three captions, reaches the model amputated. This has already caused production defects.
3. **Tool calls and tool results are never in history.** `role` is only `user` or `assistant`. Tool activity lives solely in Agno's Redis run record on a 30-minute TTL and is never replayed. The model cannot know it already fetched the user's accounts or already created a post — only whatever prose it happened to say about it.
4. **History is `role` + `content` only.** `getChatHistory()` selects just those two fields, so `mentions`, `reference_clips`, `post_id`, `job_id` and, critically, **timestamps** never reach the model. It cannot reason about elapsed time between turns.
5. **Assistant prose is lost when a turn produces two assistant texts** — Laravel assigns rather than appends (`$fullResponse = $content`), so the earlier text is dropped from the bubble *and* from the stored message, which corrupts every later turn's history. Tracked as `PROSE-OVERWRITE-12` in INDEX.md L360 and explicitly not fixed.
6. **The analytics context turn burns a history slot.** `AiChatHelper::analyticsContextTurn()` prepends a synthetic, never-persisted `role: USER` message as a workaround, consuming one of the 10 slots. `buildWorkflowInputs` has a comment conceding this should become first-class so the fake turn can be dropped.

Separately, on the "current chat history" half of the ask:

7. **The user cannot scroll back in a thread at all.** Server-side cursor paging fully exists — `AiChatRepo::getChatMessages($before)` returns `has_more` and `next_before`, and `fetchChatByIdApi` accepts `before`. `useAiChatThread.ts` writes `hasMore` / `nextBefore` into state and **nothing reads them**; `fetchChatById()` does not even set them. `ChatHistoryModal.vue`'s `loadMore` pages the chat *list*, not messages. So past `CHAT_MESSAGE_WINDOW = 30` (`AIController.php` L115) older messages are unreachable in the UI even though the API can serve them. This is the cheapest, highest-visibility fix in the whole epic: the backend work is already done.

### Other findings worth a dev's attention, not story scope

- **Prior-turn video is invisible to the model.** Nothing in ai-agents `src/` reads a history message's top-level `video` key. Images were fixed 2026-08-11; video was not.
- **Images cannot cross a HITL pause.** `acontinue_run` has no `images=` parameter, so a resumed turn gets URLs-as-text and the model is told it cannot see them (`_handles_without_vision()`).
- **Unbounded accumulated tool context within a single turn.** `MAX_RESULT_BYTES` caps each individual tool result; nothing caps the sum. Open question Q-K in INDEX.md L307, owned by Tech Lead.
- **Single-active-chat is server state.** `deactivateAllChats` runs before every send, so two tabs or two devices fight over which chat is active and `fetchActiveChat` can resume the wrong one.
- **Gate answers depend on a rendering helper.** A button press posts `content: ""` and `AiChatHelper::confirmAnswerText` synthesizes the bubble text (capped at `ANSWER_MAX_LENGTH = 2000`). Without it the coordinator re-asked the same question twice in one conversation.
- Existing tests to mirror: `contentstudio-backend/tests/Unit/AI/AiChatMessageWindowTest.php` (cursor paging, no dup / no gap), `AiChatHelperMapChatMessageImagesTest.php`, `AiChatTimelineTest.php`, and on the Python side `test_the_php_multi_image_history_shape_is_parsed_whole`, which asserts against verbatim PHP output so either side moving fails a test.

### Story shape for ask 1

Three stories, and they are independent enough to ship in any order:

- `[BE]` widen and enrich what reaches the model: make the window configurable and larger, raise/remove the per-turn truncation, carry timestamps and tool-result summaries, fix the prose-overwrite loss, and make analytics context first-class so it stops eating a slot. This spans Laravel and ai-agents, but it is one contract change, so splitting it by repo would create two stories that cannot be tested apart.
- `[FE]` wire up the load-older-messages paging that the API already supports.
- `[Flutter]` match the mobile thread to the same window and paging behavior, since AI chat is in scope for mobile.

---

## 7. Ask 6 — Multiple first comments and conditional scheduling: current state

Research is complete. **Stories are deliberately deferred** until the behavior is finalized and a prototype exists, per the request. Everything below is what the finalization conversation needs.

### What first comment is today

One toggle, **one shared message string**, and a per-account opt-in list. That is the whole feature.

`contentstudio-frontend/src/modules/composer/views/composerInitialState.ts` L81-91:

```ts
export interface FirstCommentState {
  has_first_comment: boolean
  first_comment_message: string
  first_comment_accounts: string[]
}
```

Backend mirrors it flat on the plan document (`app/Models/Publish/Planner/Plans.php` L93-96): `has_first_comment`, `first_comment_message`, `first_comment_status`, `first_comment_accounts`. Defaults in `config/socialPost.php` L541-543.

Selecting accounts across Facebook, Instagram and LinkedIn posts the **same string** to all three. There is no per-platform or per-account comment text anywhere in the system.

### Platforms

Offered on five (`composerInitialState.ts` L440-446): Facebook, Instagram, LinkedIn, YouTube, Telegram. Per-account eligibility in `useComposerSharing.ts` L645-660 — Facebook Pages only (Groups excluded in composer), Instagram only when posting method is `api`, YouTube only when `privacy_status === 'public'` and not made-for-kids, LinkedIn and Telegram unconditional. **No support at all** for X, Threads, Bluesky, TikTok, Pinterest, GMB/GBP, Tumblr.

Publish sites, all five duplicating the same gate and normalization:

| Platform | Entry point | API call |
|---|---|---|
| Facebook | `app/Libraries/Integrations/Platforms/Social/Facebook/FacebookPlatform.php` L1332-1356 → `addFirstComment()` L1481 | `POST {postId}/comments` |
| Instagram | `app/Libraries/Integrations/Platforms/Social/InstagramPlatform.php` L326 → `handleFirstComment()` L986 | `POST {mediaId}/comments` |
| LinkedIn | `app/Builders/Integrations/Platforms/Social/LinkedinBuilder.php` L257-265 → `addComment()` L306 | `POST /v2/socialActions/{postId}/comments` |
| YouTube | `app/Strategy/Planner/YoutubePosting.php` L279-285 → `addFirstComment()` | `commentThreads->insert` |
| Telegram | `app/Strategy/Planner/TelegramPosting.php` L101-104 → `postFirstComment()` L366 | `sendMessage` with `reply_to_message_id` |

### The critical finding: nothing is delayed today

**Every first comment fires immediately and synchronously, inline in the same publish call, right after the main post returns an id.** Grepping `comment_delay`, `comment_schedule`, `first_comment_time` across `app/Strategy/Planner/` and `app/Libraries/Publish/` returns nothing.

The same is true of every threaded-posting chain (X, Threads, Bluesky) — back-to-back inside one synchronous call, carrying `reply_to` forward.

So of the two halves of this ask, *multiple comments* is an extension of machinery that exists, and *conditional scheduling of those comments* is **genuinely new machinery**: it needs a queued job and a per-comment status, because `first_comment_status` is a single aggregated scalar today.

### The strongest precedent to reuse: threaded posting

Threaded posting already has the exact "parent post plus an ordered list of children" shape, with a naming convention that is `has_first_comment` / `first_comment_*` with an array instead of a scalar. `config/socialPost.php` L193-280:

```php
$twitter_options = [
    "has_threaded_tweets" => false,
    "threaded_tweets" => [ ["message" => "", "image" => [], "video" => $defaultVideo, ...] ],
    "threaded_tweets_accounts" => []
];
```

Sequential publish with the reply chain carried forward, `TwitterPlatform.php` L169-193 — note the `break` on insufficient credits, so a failed link stops the chain rather than orphaning later items.

What the threaded path hands over for free: the array shape, the sequential-chain publisher, per-item result recording (`threaded_tweets` / `multi_threads` arrays with per-item `link` / `error` / `error_message`, persisted at `PlansRepository.php` L2379-2381), the per-item validation loop (`useComposerSubmitValidation.ts` L340+), and multi-item previews (`PostPreview.vue` L476-492 already renders `getThreadedTweet()` / `getMultiThreads()` / `getMultiBluesky()` as lists, right next to the single-string `getComment()`). What it does not hand over is any notion of delay between items.

### Structural obstacle worth knowing before design

`firstComment` is neither a `sharing_details` section nor an `options` section. It is a third, flat, **scalar** section — `useComposerDraftSync.ts` L47 classifies it as `ScalarSection` alongside `publishTimeOptions` and `selection`, and `sharingDetailsTypes.ts` L160-181 makes it a sibling of both.

Going to multiple comments therefore means choosing between:

1. **Move it under `{platform}_options`**, matching how every other per-platform feature works, or
2. **Turn the scalar section into a keyed or array section**, which ripples through `RenderedDraftSnapshot`, `HydratePayload`, `useComposerDraftSync`, `PostPreview`, and all five backend normalizers.

There is already a **ready-made seam for per-platform text that has zero producers**: `buildValidationInputs.ts` L27 and L72-84 define and read `firstCommentByPlatform?: Record<string, string>`, and nothing in the repo ever populates it.

### Other constraints to carry into design

- **Character limits are inconsistent and silently wrong when mixing platforms.** `useFirstComment.ts` L69-91 returns the *minimum* limit across selected accounts, enforced as a bare `maxlength` with no visible counter: LinkedIn selected → 1250, else Instagram → 2200, else 8000. So LinkedIn's 1250 silently caps an Instagram comment's 2200. The public API v1 uses a **different** limit again, 2000 (`PostStoreRequest.php` L147-150).
- **Instagram counts caption plus first comment together** against the 30-unique-hashtag rule (`features/validation/rules/instagram.ts` L100-140). N comments multiply this.
- **A comment failure never fails the post.** Each call is try/caught and returns `['status' => false, ...]`; plan status stays `published` while `first_comment_status` independently becomes `failed`. Aggregation in `app/Libraries/Publish/Posting/Posting.php` L139-162 and `PostingTally.php`.
- **There is no "retry just the first comment" path.** `app/Jobs/RetryPostingJob.php` re-runs the entire posting for a channel, which would duplicate the main post. `PartialFailedDetail.vue` L13, L62 deliberately hides the retry affordance when the failure is comment-only.
- **The planner already conflates the two concepts.** `src/locales/en/planner.json` L712 reads `"first_comment_failed": "First comment/Threads failed"` — one status bucket for both.
- **Five duplicate normalizers.** `FacebookPlatform::getFirstCommentDetails`, `InstagramPlatform::getFirstCommentDetails`, `YoutubePosting.php` L113-118, `LinkedinBuilder.php` L156-162, `TelegramPosting.php` L337-356. Telegram has already drifted, carrying extra `apply_to_all_accounts` and `enabled`/`text` aliases. Any multi-comment work should consolidate these first or the drift multiplies by N.
- **Automation paths carry first comment too** and would need to follow: evergreen variations (`app/Libraries/Publish/Planning.php` L1547-1549), `RecyclePostsService`, `RssPlansJob`, `config/rssAutomation.php`, `config/csvAutomation.php`.
- **Billing gate** is the `auto_first_comment` feature key (`src/modules/billing/constants/featureList.ts` L9).
- **Mobile is a strict subset.** `contentstudio-flutter/lib/features/composer/domain/model/first_comment.dart` is only `{enabled, message}` — no account picker, and `plan_hydrator.dart` L183-188 **drops `first_comment_accounts` entirely**. Mobile eligibility also differs from web: Facebook Page **or Group**, Instagram, LinkedIn, and **no YouTube or Telegram**, with no Instagram posting-method gate. Multi-comment work will widen an already-divergent mobile gap.
- **Public API shape differs from storage.** v1 takes a nested `first_comment: {message, accounts}` (`PostStoreRequest.php` L147-150) and flattens it (`PostController.php` L1803-1809), reversing for reads in `PostResource.php` L339, L359-366. A multi-comment shape needs a v1 decision: extend the nested object to an array, or version the endpoint.

### Open questions for the finalization round

These are the decisions that must be made before stories can be written. None of them can be inferred from the codebase.

1. **What does "conditional" mean?** Three readings are all consistent with the ask: (a) time-delayed — comment 2 posts N minutes after comment 1; (b) condition-triggered — post a comment only if the post reaches some engagement threshold; (c) per-platform conditional — this comment only for LinkedIn, that one only for Instagram. These are very different features. (c) is nearly free given the existing unpopulated `firstCommentByPlatform` seam; (b) needs engagement polling that does not exist anywhere.
2. **How many comments, and is the cap per platform?** Threaded posting has its own caps; comment APIs are rate-limited differently per network.
3. **Ordering guarantee under delay.** If comment 2 is delayed and comment 1 fails, does comment 2 still post? Threaded posting `break`s the chain. Does the same rule apply, and is it configurable?
4. **Per-comment status and retry.** A per-comment status field is unavoidable for delayed comments. Does that also mean the product finally gets a comment-only retry, which it has never had?
5. **Which platforms are in scope?** The five that support it today, or is this the moment to add X/Threads/Bluesky, where replies are already implemented as threads rather than comments?
6. **Scheduled-post interaction.** For a post scheduled two weeks out, is a comment delay measured from publish time or from the scheduled time? These diverge whenever publishing is late or retried.
7. **Mobile.** Given mobile does not even have the account picker today, does it get multi-comment, read-only display of comments authored on web, or nothing?

### Recommended next step for ask 6

Build the composer prototype in `cs-prototypes/app/features/` (Next.js 15, React 19, Mantine 8, Tailwind 4; register in `app/features/_meta.ts`, default-export `page.tsx`, write a `prompt.md`). Prototype the multi-comment editor with per-comment platform targeting and a delay control, get it agreed, then write stories against the agreed behavior. Option (c) plus (a) is the combination the existing code most cheaply supports.

---

## 8. Ask 3 — AI Studio images into the media library: current state

### The headline: it already works for one surface and not the other

**AI Studio *tools* persist automatically. AI *chat* does not.** Two code paths, two different outcomes. The marketing ask reads as one gap; it is really one gap plus one inconsistency.

### AI Studio tools — already persisted

The tools surface (image-to-image, remove-background, upscale, headshot, face-swap, outfit-swap, product-image) persists on terminal success. `app/Services/AI/AiToolStreamingService.php` calls `AiToolMediaPersistenceService::persistToLibrary()`, which fabricates an internal request and reuses the normal link-upload path:

```php
$request = new Request([
    'workspace_id' => $workspaceId,
    'link' => [$imageUrl],
    'source' => "ai-tool:{$toolKey}",
    'is_ai_generated' => true,
    'user_id' => $userId,
]);
$data = $controller->uploadMediaByLink($request)->getData(true);
```

The bytes are copied off fal.ai onto GCS, a real `media` doc is written, `MediaAssetProcessingJob` is dispatched, and the terminal SSE event is enriched with `persisted`, `media_id` and `persist_error`. From the SPA this is **best-effort**: `AiToolsController::stream()` omits `$requirePersistence` so it defaults to false, persistence runs after credit deduction, and a failure still returns success with `persist_error` driving a warning banner in `ToolResultPanel.vue` L168-200. Async video tools reach the library the same way via `ChatMessageService::finalizeVideoJob()`. Public API v1 generation uses the **strict** variant (`AiImageSyncService::generate()`), which fails the whole request and rolls back on `MEDIA_STORAGE_LIMIT_EXCEEDED` / `MEDIA_PERSISTENCE_FAILED`.

### AI chat — not persisted

`app/Helpers/Ai/AiChatHelper.php` `persistImageAsset()` L63 re-hosts the fal image to GCS and stops:

```php
$uploadPath = 'ai-generated-images/'.date('Y/m/d').'/'.uniqid().'.'.$ext;
$gcsUrl = MediaLibrary::uploadOriginalsToGCSFromURL($uploadPath, $imageUrl);
```

No `media` document is ever created — grepping `MediaRepository`, `saveMedia`, `uploadMediaByLink`, `AiToolMediaPersistence` across `AiChatHelper.php` and `AIController.php` returns nothing. The URL lives only on the chat message content (`{type: 'image_url', image_url: {url}}`) and in the chat's `image_meta` array.

So the accurate framing for the story: **this is not an expiring link.** It is a durable public GCS object in the same nearline bucket the media library uses, with **no database row**. Consequences:

- Invisible to the Media Library, to AI Creations, to the composer picker, to search, and to folders.
- Invisible to storage-quota accounting, which counts `media` rows (`SubscriptionLimits::availableMediaLibraryLimits`).
- Orphaned when the chat is deleted. The only way to find it again is to scroll the chat.
- Its GCS path `ai-generated-images/Y/m/d/...` has **no workspace segment**, so unlike every real media object (`media_library/{workspace_id}/...`) these are not even partitioned by tenant on disk.

### Three partial escape hatches that exist today

1. **A per-image "Add to media library" button in chat** (`src/modules/AI-tools/components/BotMessageImageCard.vue` L105-165, `ImageAction = 'download' | 'library' | 'composer' | 'schedule'`) routing to `useAIChatMedia.addImageToMediaLibrary`. It does a **browser-side re-download and re-upload**:

   ```ts
   const response = await fetch(imageUrl)
   const blob = await response.blob()
   const file = new File([blob], `ai-generated-${Date.now()}.png`, {...})
   const uploadResult = await uploadImages([file], null, { isAiGenerated: true })
   ```

   This creates a **second copy of the bytes** on GCS. It sets `is_ai_generated: true` but **no `source`**, so the row cannot be attributed to AI chat. The name is a generic `ai-generated-<timestamp>.png`, and none of the generation metadata — model, prompt, seed, aspect ratio — reaches the media doc.

2. **"Add to composer" and "Schedule" silently perform the same upload** with `skipToast: true`. The code comment is explicit: `// AI image URLs are not durable publishable links — schedule the library copy.` So a user who never touches the library button still accumulates library rows, invisibly.

3. **An agent-side `media_import` tool** (`contentstudio-ai-agents/src/integrations/contentstudio/toolkits/organisation_write.py` L502, HITL-confirmed) whose docstring says *"including one just generated in this chat, so it survives beyond the conversation."* It posts to public API `POST workspaces/{id}/media`, and `Api/V1/MediaController::upload()` **does not set `is_ai_generated`** — so an agent-imported generation lands under **All uploads**, not AI Creations.

### Media document schema

`media` collection (`app/Models/Utilities/Media.php`), written by `MediaRepository::saveMedia()` L37: `_id`, `workspace_id`, `user_id`, `created_at`, `updated_at`, `type` (literal `"Media Library"`), `link`, `key`, `name`, `size`, `extension`, `mime_type`, `thumbnail`, `thumbnails`, `converted_link`, `converted_size`, `h`/`w`/`a`, `zapier`, `folder_id`, `source`, `is_processing`, `is_global`, `is_archived`, `is_ai_generated`, `super_admin_id`, `audio_track_id`, `audio_platform`, `note`, `note_added_by`, `planIds`, `deleted_at`, `canva_design_id`/`crello_design_id`, plus the Brand Knowledge v2 set `brand_asset`, `brand_used`, `linked_profile_id`, `ai_description`.

Note `is_ai_generated` is **not** a declared property on `MediaAssetData` — it reaches the wire only through the `HandlesAdditionalFields` pass-through. Worth fixing while in here.

### AI origin flag and filter

`is_ai_generated` (boolean) plus `source` as a secondary origin string (`ai-tool:{toolKey}`, `ai-reel-generator`, `brand_analysis`, `canva`). **Correction to the existing `media-library-ai-creations-and-request-state` research:** its fault 1, that the backend drops `is_ai_generated` and AI Creations returns the whole library, is **stale and now fixed** (commits `f7f10185b`, `f99205417`). `fetchMediaAssets()` maps both filters and `MediaRepository::findMediaAssets()` L991-1001 honours them, and the AI Creations count now spans both collections (L872, adding `AiContentLibraryPostRepo::getPostsCountByWorkspaceId`). Its faults 2, 3 and 5 — the FE client-side filter with `MAX_FILL_PAGES = 5` page refetch loop, the unfiltered `total`, and the empty state rendering before the first request — are all still true.

That epic also left open: *"Confirm whether `Media` already records which surface generated an asset."* **It does for tools** (`source = "ai-tool:{toolKey}"`, `ai-reel-generator`); it **does not for AI chat**, because chat writes no row and the manual button sets no `source`. That is precisely why that epic could not positively define its AI Studio pill. The two epics are complementary: this ask supplies the missing `source` on the write side.

### Ownership model

Workspace-primary with user as a secondary dimension. Every doc carries both `workspace_id` and `user_id`; GCS path is `media_library/{workspace_id}/uncategorized/original/...`. `findMediaAssets()` scopes by workspace, OR-ing in `is_global = true AND super_admin_id = <workspace owner>` for shared folders (gated on the `access_shared_folder` permission), and for the all-uploads view also OR-ing legacy rows with no `workspace_id`. Storage quota is per workspace subscription and checked before every upload, which is why `media_storage_full` is a distinct `persist_error` code.

### Story shape for ask 3

Two stories:

- `[BE]` persist AI chat images to the media library server-side on generation, reusing `AiToolMediaPersistenceService` so chat and tools share one path; stamp `source = 'ai-chat'`; derive a real filename; carry the generation metadata; declare `is_ai_generated` properly on the DTO; make the agent-side `media_import` set the AI flag; and make the existing FE button a no-op pointer at the already-persisted row instead of a second byte copy.
- `[FE]` show the persisted state in chat — the image card reflects that the image is already saved, the button becomes "View in media library", and the quota-failure case gets real copy.

Key decision to settle in the story: **best-effort or strict.** Tools-from-SPA is best-effort; public API v1 is strict. Chat should follow the SPA precedent (never fail a generation the user already paid credits for) and surface a quota warning instead.

---

## 9. Ask 5 — Bulk download from the Content Library: current state

### Multi-select already exists, so this story extends rather than builds

This is the cheapest of the five in FE terms. `MediaSelectionBar.vue` is a fixed bottom-centre bar shown whenever `selected.length`, already containing a "Bulk Actions" `Dropdown`. Adding a Download item is a small change.

Selection state lives in `src/modules/publish/components/media-library/composables/useDragDrop.ts` as plain refs, not Pinia and not the URL:

```ts
const selected = ref<string[]>([])
const isAllSelected = ref<AllSelectedState>({ visible: false, total: false })
```

Selection is deliberately cleared on section or route change (`MediaLibraryMain.vue` L530-539).

Bulk actions today: remove selected, archive, restore (archived section), move to folder via drag-drop onto the sidebar, compose post, export as CSV (metadata only, emailed), and select-all-visible / select-all-across-pages. There is no bulk download and no add-to-campaign.

**Select-all-total semantics that a zip endpoint must match** (`MediaLibraryMain.vue` L646-652):

```ts
payload.ids = []
payload.folder_id = normalizedSelectedType.value === 'folder' ? route.query.folderId : getSelectionScopeType()
```

The server re-resolves with `limit: 10000, page: 1` (`MediaLibraryAssetsController.php` L869-884). So a new endpoint accepts `{ids}` **or** `{select_all + folder_id + type + usage_filter}` — the shape `exportCSVMedia` / `fetchMediaAssetsById` already use.

### Single download today

One place only: the preview modal. Not in the asset card's `⋮` menu. `useMediaLibrary.ts` L259-283 does a client-side fetch, blob and synthetic anchor against the raw public GCS URL. There is **no backend download endpoint** and no `Content-Disposition` proxy for media. That this works proves GCS CORS already allows browser GETs from the app origin.

### Building blocks for the zip

- **No zip code anywhere.** Zero hits for `ZipArchive`, `ZipStream`, `archiver`, `jszip` in either repo. `jszip` / `file-saver` are **not** in `package.json`, so a client-side zip would mean a new dependency. Server-side, `php8.4-zip` **is** installed (`docker/BaseDockerfile` L31), so `ZipArchive` is available.
- **Closest precedent is the CSV media export**, and it is synchronous today: `MediaLibraryAssetsController::exportCSVMedia` L1502 → `createCsv` → `putFileAs` to GCS → persist record → email link.

  ```php
  $disk = Storage::disk('gcs-contentstudio-media-library-nearline');
  $gcsResponse = $disk->putFileAs('media_library_exports', $file, 'media_content_'.Carbon::now().'.csv', [...'public']);
  $file_url = $disk->url($gcsResponse);
  self::storeMediaExportCsv($user['_id'], $payload['workspace_id'], $user['email'], $file_url);
  ```

  The `media_library_exports/` GCS prefix and the `media_csv_export` Mongo collection already exist. The email template lives in an **external** service (`LUMOTIVE_EMAILS_API`), not this repo. The FE just optimistically toasts "An email will be sent to you shortly".
- **Queues:** base `app/Jobs/Job.php`; media lanes `media_library:processing_data` and `media_library:delete_media` already have workers, so reusing one avoids provisioning. Long-running export archetypes: `app/Jobs/Analytics/GenerateReportJob.php` (`$timeout = 1800; $tries = 3; $backoff = [60,120,300]`) and `app/Jobs/Analyze/ExportReportJob.php` (idempotency guard: skip when already completed with a URL). Best *media* archetype is `app/Jobs/Storage/ImportDriveMediaJob.php` — per-file isolation, Mongo placeholder row, LogsBuilder plus Sentry on failure, realtime broadcast on completion.
- **Notification channels, three and all in use:**
  1. **DB notification centre** — `NotificationRepository::save([... 'type' => 'report_generated', 'source' => 'analytics' ...])` via `app/Services/Analytics/ReportCompletionService.php` L780, with the FE routing on click in `src/composables/useNotificationHandler.ts` L414. A zip story adds a new `type` plus an FE route case.
  2. **Realtime is Centrifugo, not Pusher.** `App\Libraries\Realtime\Realtime::broadcast($channel, $event, $data)`, with channel builders in `app/Libraries/Realtime/Channels.php` mirrored byte-for-byte in `src/modules/common/services/realtime/channels.ts` and a golden fixture `channels.fixture.json`. A `media-import:{workspaceId}:{mediaId}` family already exists with a 45s polling fallback in `useMediaImportRealtime.ts` — the pattern to copy, and a new builder must be added in **both** files plus the fixture.
  3. **Email** — `App\Jobs\Emails\EmailJob` → `Helper::sendMail(...)` for in-repo templates, or the external `LUMOTIVE_EMAILS_API` route the CSV export uses.
- **No media downloads list UI exists.** Analytics has a whole download-reports surface (`src/modules/analytics/components/reports/download-reports/`, `useDownloadReports.ts`), but media CSV exports are email-only and the `media_csv_export` rows are never read back. So either email the link like CSV does, or build a net-new panel. Recommend following CSV (notification centre plus email) and not building a panel in v1.
- **Synchronous streamed precedent** if a small-selection fast path is wanted: `app/Http/Controllers/Settings/RequestLogController.php` L120-145 and `app/Http/Controllers/Billing/XWallet/UsageController.php` L54-80 return `StreamedResponse` with `Content-Disposition: attachment`.

### Constraints that shape the story

- **Storage is GCS, not S3.** `gcs-contentstudio-media-library-nearline` is primary, with `-standard` and legacy `lumotive-web-storage`. **Nearline is the hot bucket**, so bulk re-reads carry retrieval cost and latency — worth an explicit note.
- **URLs are public and unsigned.** Uploads are written `'public'` and `link` is `$disk->url(...)`. Signed URLs exist only for the temp-upload/audio lane. So a server-side zipper can stream objects by key without auth, and `key` is already stored on every doc alongside `link`, `name`, `size`, `mime_type`.
- **No CDN and no white-label media domain.** White-label affects app URLs and email branding only, not asset hosts. Conversely the zip link will be a bare `storage.googleapis.com` URL unless proxied.
- **Worst case is a 10k-item workspace** (select-all resolves at `limit: 10000`); grid pages at 40 with infinite scroll. Files can be large: `upload_max_filesize = 2048M`, resumable GCS uploads from 10 MiB, videos to ~150 MB discussed in the Drive-import code. Total zip size is unbounded in principle, so the story needs an explicit count and size cap, streaming with `writeStream` rather than buffering in PHP memory, and a job timeout in the 1800s range.
- **Skip placeholder rows.** Some docs are `is_processing: true` with **no `link`** (in-flight Drive imports).
- **Respect the global-folder exclusion.** Assets in global folders carry `is_global: true` and access is gated on `access_shared_folder`, which admins do **not** hold by default (`PermissionHelper.php` L105). A bulk scope must apply the same `$excludeGlobal` rule as `fetchFolders` / `fetchMediaAssets`.
- **Convention: all media endpoints are `POST`**, including fetches, returning flat `{status, message, ...}` bodies inspected manually via `rawPost`. A new endpoint should match rather than go REST.
- **Auditing is thin.** `App\Enums\Activity\ActivityAction` only covers note create/update/delete, so "who downloaded what" would be new.

### Story shape for ask 5

Two stories: a `[BE]` queued zip-export endpoint plus job and notification, and an `[FE]` story adding the Download bulk action, the caps and confirmation copy, and the ready/failed states. No Flutter story — the mobile app has no media-library bulk surface, and a zip download is not a mobile-shaped interaction.

---

## 10. Ask 2 — Highlight a passage and reply to it: current state

### Nothing like this exists

`grep -rn "window.getSelection\|getSelection()\|selectionchange" contentstudio-frontend/src` returns **zero hits repo-wide**. No selection popover, no quote reply, no `user-select` styling on messages. Copy-to-clipboard exists but **whole-message only** (`src/modules/AI-tools/composables/useBotMessageActions.ts` L111-115 `copyPrompt`, duplicated for the user bubble at `UserChatTemplate.vue` L374-375).

### How a message is rendered, and why it matters here

The AI chat does **not** use `marked`, `markdown-it` or `DOMPurify`. It uses **`streamdown-vue` v1.0.29** (pinned exact, `package.json` L115), which renders hast straight to Vue vnodes — so there is **no `v-html` anywhere in the message path**. Sanitization is URL-prefix allowlisting (`allowedImagePrefixes` / `allowedLinkPrefixes` = `['https://', 'http://']`), not HTML sanitizing. Wrapper is `src/modules/AI-tools/components/StreamingMarkdown.vue`. (`markdown-it@14.1.0` is in `package.json` but only `modules/AI-tools/composables/chat.ts` imports it — legacy, not in the live bubble path.)

Message list is one loop with a component switch, `src/modules/AI-tools/ChatBox.vue` L51-66: `message.role === 'user' ? UserChatTemplate : BotChatTemplate`.

| Role | Component |
|---|---|
| user | `src/modules/AI-tools/UserChatTemplate.vue` — right-aligned bubble, body via `components/UserMessageMentions.vue`, which also uses `StreamMarkdown` and overlays `@ImageN` mention chips |
| assistant | `src/modules/AI-tools/BotChatTemplate.vue` — left-aligned, and its root wrapper already carries a per-message DOM id built from the message id (L32-36), so **a stable anchor already exists** |

### Four real obstacles the story must handle

1. **Every word is its own inline-block element.** `src/modules/AI-tools/utils/rehypeAnimateWords.ts` wraps each word in a `span.sd-animate-word` and `StreamingMarkdown.vue` L112 sets `display: inline-block` on them. So a DOM `Range` over one sentence crosses dozens of element boundaries, and `range.toString()` comes back with missing or odd whitespace at those boundaries. Naive `toString()` will produce mangled quotes. Either strip the plugin while a selection is active, or extend it to emit offset `data-*` attributes.
2. **The assistant body has two mutually exclusive render paths** (`BotChatTemplate.vue` L44-60): `components/timeline/AiBlockTimeline.vue` for stream-protocol-v2 ordered blocks (`AiProse`, `AiDataTable`, `AiChart`, `AiStatTiles`, `AiPreviewCard`, `AiQuestionCard`, `AiSourceList`, `AiStepRow`, `AiAssetGrid`), and `StreamingMarkdown` only when `messageContent && !timelineHasProse`. A selection can therefore land in prose, in a table cell, in a chart, or in a human-in-the-loop confirm card. Prose-only is the sane v1 scope.
3. **The rendered DOM lags the store while streaming and is re-diffed every frame.** `StreamingMarkdown.vue` L25-76 reveals at `REVEAL_CHARS_PER_SECOND = 120` via `useRafFn` at 60fps, repairing truncated markdown each frame with `parseIncompleteMarkdown`. A selection made mid-stream will be destroyed or shifted. Gate the affordance on `!message.isLoading` — the same flag `BotMessageFooter` already uses (`BotChatTemplate.vue` L118-119).
4. **Anchor by text plus message id, never by character offset into the DOM.** Given points 1 and 3, offsets are not stable.

### Where the pieces go

- **The affordance joins an existing row.** `src/modules/AI-tools/components/BotMessageFooter.vue` is a row of `ActionIcon`s shown on group-hover (Sources, Replace text in editor, Add to editor, schedule, add-to-draft, copy-text). A floating selection bubble is new, but the fallback "reply to this message" action has a home.
- **The input already has a chip region, with a close precedent.** The composer is `src/components/dashboard/ChatInput.vue` (2432 lines, one shared instance across every chat surface, never remounted between hero and docked states). The text field is **TipTap 3.23.4** via `src/modules/AI-tools/components/CstTextEditor.vue`, not a textarea. Above it sit media-mode pills, frame slots, image and video thumbnail rows, and — the closest precedent — `src/modules/AI-tools/components/ReferenceFilesAttachment.vue`, a list of chips with an `Icon`, a truncated label, metadata and an `X` remove button emitting `@remove(index)`, including uploading-spinner and error states. A quoted-selection pill should look and behave like these.
- **Draft state has an owner.** `src/stores/core/useAIChatStore.ts` L26-63 holds a singleton `currentMessage` with `content`, `images`, `videos`, `referenceFiles`, `imageGeneration`, `videoGeneration`. A `quotedSelection` field belongs here.
- **The wire format has a precedent to copy exactly.** `src/modules/AI-tools/utils/buildChatStreamPayload.ts` already carries `mentions` ("only added when the user typed @ImageN chips") and `referenceClips` ("persisted alongside the message so the user's bubble can be rebuilt on reload"). Mentions are echoed back into the bubble via `UserChatTemplate.vue` L42-46. A quoted-selection field should follow that pattern so the quote survives a reload — which also means the Mongo `ai_chat_messages` doc needs to carry it.
- **The quote must reach the model explicitly.** Per the ask-1 findings, history is a hand-rolled labelled string built in `_build_team_input()`. A quote that is only rendered in the UI and not represented in that string will be invisible to the model, which defeats the feature. This makes ask 2 depend on the ask-1 contract work.
- **Repo rule:** `contentstudio-frontend/docs/stream-v2-frontend-sot.md` is the stream-v2 contract and decision log, and the frontend CLAUDE.md requires appending to it in the same commit as any stream-v2 change.

### Mobile

Out of scope for v1. `contentstudio-flutter/lib/features/ai_assistant/presentation/widgets/ai_html_message.dart` builds native widgets with `Text.rich` / `InlineSpan` buffers, which are **not selectable** — quote reply there would require moving to `SelectableText.rich` throughout. Note for the record and skip.

### Story shape for ask 2

Two stories: an `[FE]` story for selection capture, the floating bubble, the quoted pill in the input and the quote block in the transcript; and a `[BE]` story to persist and forward the quote so it reaches the model and survives reload. Prose-only scope for v1, with selections inside rendered components explicitly excluded.

---

## 11. Ask 4 — Feature mentions and redirect CTAs: current state

### A navigate mechanism already exists, unvalidated and ungated

This is the most important finding for this ask. Assistant messages can already emit follow-up actions that deep-link into the app. `src/modules/AI-tools/composables/useBotMessageActions.ts` L35-40, L163-175:

```ts
export interface FollowupAction { text?: string; type?: string; url?: string; [key: string]: unknown }
...
const handleFollowupAction = async (action: FollowupAction) => {
  if (!action) return
  const actionType = action.type || 'prompt'
  if (actionType === 'navigate' && action.url) {
    bridge.closeChat()
    await router.push(action.url)
    return
  }
  onFollowupPrompt(action)
}
```

`stream-frames.ts` L491 narrows it to `type?: 'prompt' | 'send' | 'navigate'`, and it renders as pill buttons in `components/BotMessageFollowups.vue`.

So the ask splits cleanly into two halves: **add inline CTAs when the assistant names a feature in prose**, and **harden the existing navigate path**, which today does `router.push()` on a **raw URL string produced by the model**, with no `hasRoute` check and no plan or permission gating. A model hallucinating a link-in-bio path currently produces a dead navigation, and a model naming Social Listening to a user whose plan lacks it currently produces a bounce.

### Workspace slug is mandatory in every in-app URL

Every destination is `/:workspace/<module>/...` carrying the workspace **slug** — no query param, no subdomain. Resolution (`components/layout/useHeaderNavigation.ts` L91-96) falls back from the active workspace's slug to the current route param. A CTA must build `{ name: '<routeName>', params: { workspace: workspaceSlug } }`. This is a second reason not to let the model emit raw URLs: it does not reliably know the slug. Exceptions with no workspace param: auth routes, `/workspaces`, `/manage_team`, `/createWorkspace`, `/manage-limits`, `/start_trial`, share links, `/api-connect/:sessionToken`, `/easy-connect/:id`.

### Registries to reuse rather than reinvent

**The nav registry is the right source of truth.** `src/components/layout/useHeaderNavigation.ts` (433 lines) already pairs a feature id, label key, icon, route target and gating in one object:

```ts
export interface HeaderNavigationItem {
  id: string; labelKey: string; tooltipKey?: string; iconName?: string
  to?: RouteLocationRaw
  landingValue?: string
  isActive: boolean; isVisible: boolean; isDisabled: boolean
  tooltip?: string; showLock: boolean
  featureKey?: FeatureKey
  disabledAction?: 'upgrade' | 'listening-upgrade' | 'none'
}
```

Item ids map to landing route names: `home`, `ai-studio`, `publisher`, `inbox`, `analytics`, `listening`, `media-library`, `api`, `discover`, `social-accounts`, `brand-knowledge`.

Critically, `isDisabled`, `showLock`, `disabledAction` and `tooltip` are **already computed per item**, and disabled items deliberately set `to: undefined`. Consuming this composable gets the paywall check for free and keeps CTAs consistent with what the nav rail shows.

`src/config/feature-announcements.ts` holds `FEATURE_ANNOUNCEMENTS` and `export type FeatureKey`, roughly 22 keys — but it is a new-badge registry (dates plus a Usermaven surface string) with **no route info**, and it does not cover Composer, Planner, Discovery or Analytics overview. It is the naming convention to mirror or the place to extend, not a ready map.

### Canonical routes for the CTA map

| Destination | Route name | Path |
|---|---|---|
| Create post / Composer | `social-modal` | `/:workspace/composer/:id?` — **frozen name and path**, backend emails depend on them; canonical create-post deep link |
| Planner | `planner`, or `planner_calendar_v2` / `planner_list_v2` | `/:workspace/publisher/planner[...]` |
| Analytics overview | `analytics_overview_v3` | `/:workspace/analyze/overview/` |
| Per-platform analytics | `facebook_analytics_v3`, `instagram_analytics_v3`, `twitter_analytics_v3`, `linkedin_analytics_v3`, `tiktok_analytics_v3`, `youtube_analytics_v3`, `pinterest_analytics_v3`, `threads_analytics_v3`, `bluesky_analytics_v3`, `gmb_analytics_v3` | `/:workspace/analyze/<platform>/:accountId?` |
| Competitor analytics | `competitor_analytics` | `/:workspace/analyze/competitor-analytics` |
| Reports | `my_report_v3`, `download_reports_v3`, `reports_setting_v3` | `/:workspace/analyze/...` |
| Social Inbox | `inbox-revamp` with `params.filter = 'unassigned'` | `/:workspace/inbox/:filter` |
| Automations | `evergreen-automation-listing`, `rss-automation-listing`, `csv-process-listing` | `/:workspace/publisher/automation/<type>` |
| Content Library | `media-library` | `/:workspace/publish/media/` — nav labels it "Content Library" |
| AI Content Library | `ai-content-library-posts` | `/:workspace/publisher/ai-content-library/posts` |
| Discovery | `discovery-v2` (names centralized in `modules/discovery_v2/constants/routes.ts`) | `/:workspace/discovery` |
| Social Listening | `listening` | `/:workspace/listening` |
| AI Studio | `ai_studio`, `ai_studio_chat`, `ai_studio_tool` | `/:workspace/ai-studio[/tools/:toolKey]` |
| Brand Knowledge | `brand-settings` | `/:workspace/settings/brand-settings/` |
| Social accounts | `social` | `/:workspace/settings/social/:id?` |
| Team | `team`, `groups` | `/:workspace/settings/...` |
| Billing | `plan`, `subscription` | `/:workspace/settings/...` |
| Content categories | `content_categories` | `/:workspace/settings/content_categories/` |
| Approval workflows | `approval-workflows` | `/:workspace/settings/approval-workflows/` |
| Hashtags | `miscellaneous` | `/:workspace/settings/miscellaneous/:id?` — **no dedicated route**; `Miscellaneous.vue` L18 renders the hashtags section |

Three destinations the map must **not** invent:

- **Link in Bio does not exist.** No route, no module, no component. The only "link in bio" strings in the repo are inside listening mocks and competitor dummy data.
- **Best Time to Post has no route.** It is an in-composer and scheduling widget (`modules/publish/components/posting/TimeRecommendation.vue`, `modules/common/components/schedule-post/ScheduleTimeStep.vue`, API `analytics/scheduling/optimal-times`). A CTA should point at the Composer or the relevant platform analytics page.
- **`ai_studio_history` is a dead name.** It appears in two gating allowlists but is not a registered route.

Also note the legacy `/:workspace/planner/*` paths all redirect into `/:workspace/publisher/planner/*`, and a second older analytics tree still exists at `/:workspace/analytics/*`. Prefer route **names** over hand-built paths.

### Five gating layers a CTA must respect

`meta` is almost entirely just `{ title }` — gating lives in guards, not meta.

1. **Global `beforeEach`** in `src/router.ts` (~L855+): trial-expired lockout redirects everything to `trial_expired`; a role allowlist via `usePermission().hasRoutePermission(to.name)` bounces to `planner_list_v2`; the API-centric plan block alerts *"This feature is not included in your API Plan."* and redirects to `api`; sample/demo workspaces block any route matching `discovery` or `ai_studio` to `home`, plus a blocked-analytics-route check. It also validates a saved landing page with `router.hasRoute(...)` before pushing — **the precedent to copy for CTA validation**.
2. **API-centric plan blocklist** — `modules/billing/composables/useApiCentricPlan.ts` L8-53, the only explicit blocked-route constant in the repo: path patterns for analyze, analytics, inbox, discovery, listening and ai-studio, plus a names set covering analytics, listening, AI studio, brand settings, integrations, white-label, SSO, AI content library and every automation route.
3. **Approver role allowlist** — `src/composables/usePermission.ts` L268-299: only planner views, notifications, profile, email-notification status, set-password, workspaces, and the AI studio names. Everything else bounces.
4. **Per-route `beforeEnter`** — only 23 exist and nearly all are `ifAuthenticated`. The one real feature gate is Listening (`modules/listening/routes.ts` L127-146), which evaluates access and redirects to `home` with a one-shot `?listening_upgrade=locked` query that makes the home page open the upgrade modal. Good pattern for a locked CTA.
5. **Component-level feature flags** — `modules/billing/composables/useFeatures.ts` `canAccess(featureKey)` returns `{ allowed, error?: { type: 'FEATURE_UNAVAILABLE' | 'LIMIT_REACHED', message } }` backed by the plan's `features`, `limits` and `used_limits` (a limit of 0 means unlimited), with copy in `FEATURE_MESSAGES`.

**AI Studio itself has no plan flag yet** — `modules/ai-studio/config/routes.ts` L1-11 says so explicitly, and specifies that when an `ai_studio` subscription feature ships the guard should be modeled on `listening/routes.ts` and **not** on the broken `shouldLockBasedOnSubAddons` in `router.js`.

### Story shape for ask 4

Two stories: a `[BE]` story establishing a server-owned feature registry so the model emits a stable feature **key** rather than a URL, and an `[FE]` story that resolves keys to gated route targets, renders inline CTAs in prose, and hardens the existing `navigate` follow-up path. The registry belongs on the backend so it can be updated without a model change and so the agent tier can be told which keys are valid.

Mobile is a natural follow-up — `lib/features/ai_assistant/presentation/widgets/ai_follow_up_chips.dart` already exists and navigation is `go_router` with `app_links` for deep links — but it depends on the same registry landing first, so it stays out of this batch.
