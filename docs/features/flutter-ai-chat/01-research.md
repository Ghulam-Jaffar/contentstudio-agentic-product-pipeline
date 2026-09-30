# Research: Flutter AI chat, a full LLM-app experience on mobile

**Source:** PO, 2026-09-30, following the team meeting. AI chat in the Flutter app should feel exactly like using Claude or ChatGPT on a phone. That is "actual big things", to be discussed further. Asks captured so far:

- Take AI chat out of the Composer and make it reachable from anywhere in the app
- The live elapsed timer on running steps, and smooth typing and streaming, same as web
- Carousel handling (filed separately in the existing **AI Carousel post generation** epic: `docs/stories/flutter-ai-chat-carousel/`)
- Everything else mobile is missing vs web
- A notification when the AI finishes: in the app, and when the app is minimised, except when the user already has that exact chat open

**Epic:** new, **Flutter AI chat: a full LLM-app experience on mobile**. Several stories are skeletons pending the PO's discussion and the design.

## Current state (`contentstudio-flutter/`)

**Entry points**
- Composer is the only entry: `lib/features/composer/presentation/widgets/composer_ai_button.dart:133` calls `showAiAssistantSheet`. The button sits in the caption toolbar (`composer_caption_toolbar.dart:104`). Its label launches "write"; its chevron opens `showAiWritingOptions` (write/improve/rephrase/shorten/lengthen/spellingGrammar plus "AI image", which is a separate `ComposerAiImageSheet` flow).
- `lib/features/ai_assistant/presentation/ai_assistant_sheet.dart`: `showModalBottomSheet(isScrollControlled: true)` with `AiLaunchContext` and a `ComposerAiBridge` that hands results back.
- No route in `lib/app/router.dart` (shell branches `/planner`, `/inbox`, `/notifications`, `/menu`), no bottom-nav entry (`lib/app/shell/home_shell.dart`, `CSBottomNavBar`), no FAB.
- `aiAssistantControllerProvider` is a non-autoDispose `NotifierProvider` (`application/ai_assistant_providers.dart:42`), so a chat survives reopening. Opening with no action resumes the workspace's active chat (`ai/activeChat`).
- Implication for the global entry: the Composer bridge (Add to post / Replace in composer) must keep working when chat is opened from the Composer, and must not be offered when chat is opened from elsewhere. Outside the Composer, "use this" means creating a post (open Composer prefilled).

**Streaming and rendering**
- SSE over Dio to `/ai/chatWithStreaming` (`data/ai_assistant_service.dart:128-202`, `receiveTimeout: Duration.zero`). Full v2 frame protocol: `data/ai_sse_decoder.dart` → `data/ai_stream_event_mapper.dart` → `domain/timeline/timeline_reducer.dart` (port of web `reduceStreamFrame.ts`).
- Riverpod 2. Stop via `ai_stream_cancel_controller.dart` (`POST ai/cancel`). If the stream drops before the run ends, `_reconcile` (controller ~706) re-reads the turn with back-off from 5s to 2m. **So runs continue server-side when the app is backgrounded**, which is what makes the "AI finished" push useful.
- Native markdown (markdown → html → widgets), `presentation/widgets/ai_html_message.dart`, `ai_prose_block.dart` reveals prose at a steady pace.
- Steps: `ai_step_row.dart`, grouped in `ai_block_timeline.dart` ("Working" shimmer / "Worked, N steps"). Durations formatted by `domain/timeline/timeline_format.dart`. **The backend does not persist `timeline.ms`, so reloaded turns lose durations.** Web has the same exposure.
- Block registry `presentation/widgets/timeline/ai_block_view.dart`: prose, step, table, chart, metrics, asset, preview, sources, question. Unknown components degrade to a prose summary.

**Gaps vs web**
- No `/` skills, no `@` mentions, no attachment or media upload in the input.
- No account or channel picker in chat. Draft and Schedule route through the Composer's publishing options.
- Videos still generating (`asset.video` with `job_id`, no url) draw nothing and show a "coming soon" alert. No job polling.
- `navigate` suggestions are hidden on mobile. No `analytics_context`.
- No regenerate, no edit message, no voice input.
- Image sets merge into a carousel view, but there is no dedicated carousel handling (see the carousel story).

**Input performance**
- `presentation/widgets/ai_input_bar.dart`: `TextField` (`maxLines: 5`). `onChanged` → `controller.setDraft` → `state.copyWith(draft: value)` on **every keystroke**.
- `ai_assistant_view.dart:232` watches the whole `aiAssistantControllerProvider` with no `select`, so each keystroke likely rebuilds the entire view including the message list. The top suspect for typing lag.
- Mitigations exist (settled bubbles reuse widgets, memoised resolutions ~10ms per event, `cacheWidth` on thumbnails). A device pass on keyboard insets and scroll feel is still owed (`docs/features/ai-studio-stream-v2.md` §6).

**Notifications**
- `firebase_messaging`, `flutter_local_notifications`, `app_links`. Sources `data/firebase_push_notification_source.dart` and `data/native_push_notification_source.dart` (iOS APNs over MethodChannel). Foreground display `data/flutter_local_notification_presenter.dart`. Registration `lib/app/push_registration_listener.dart`. Taps `lib/app/notification_tap_listener.dart`.
- `domain/notification_intent.dart` is a sealed type with only `OpenPostIntent` (manual-publish pending post) and `UnknownNotification`. `application/notification_routing_coordinator.dart` switches workspace and navigates. No AI intent exists.
- Helpin epic **Notifications architecture** (`ea0a4fe5-...`, research-first, local `docs/stories/notifications-architecture/`) is defining one notification system. The AI-finished notification should follow it, not invent a parallel path.

**Docs**
- `contentstudio-flutter/docs/features/ai-assistant.md` (port plan, 475 lines), `docs/features/ai-studio-stream-v2.md` (authoritative, 287 lines: deliberate gaps §5, QA bug table §8), `push-notifications.md`, `notifications.md`.

## Open questions for the PO

- Global entry: a bottom-nav tab, a floating button, a header icon, or all of these? Draft leaves it to the design and requires "one tap from any main screen".
- AI-finished notifications: for runs started on mobile only, or also runs started on web? Only for long runs (for example over 10 seconds), or always? Draft: runs started from the app, any length, suppressed only when that exact chat is on screen.
- Should failed or stopped runs notify too? Draft: failed yes ("couldn't finish"), stopped by the user no.
- Voice input: device dictation (keyboard mic) is free; in-app voice mode is a much larger build. Draft: an in-app mic button using on-device speech to text.
- AI Studio tools (image, video, carousel maker) on mobile: out of scope for this epic unless the PO says otherwise.

## Competitor research (2026-09-30): ChatGPT, Claude, Gemini, Grok, Perplexity mobile

**Table stakes in 2026** (all the major apps):

1. Opens straight into a new chat; history in a drawer with search, rename, delete, **pin**
2. Pill composer that grows, with one **"+"** sheet: camera, photos, files, tools (Gemini adopted exactly this in 2026)
3. **Dictation mic** separate from **real-time voice mode** (Claude voice free since early 2026, 18 languages, push-to-talk; Gemini Live now inline in chat)
4. Token streaming, stop button, collapsible "Thought for Ns", live tool-step lines
5. Mobile markdown: side-scrolling tables, code with copy; tappable citations/sources
6. Per-reply action row: copy, regenerate, thumbs up/down, read aloud, share
7. Long-press: select text, edit and resend, **branch into new chat** (ChatGPT, Dec 2025)
8. Inline images, full-screen, save; ChatGPT edits/comments on generated images (Sep 2026)
9. Scroll-to-bottom button, composer pinned above keyboard, swipe to dismiss keyboard
10. Runs continue server-side and reappear after backgrounding or kill
11. **Push when a long task finishes** (ChatGPT and Gemini deep research; Gemini on lock screen)
12. Inline retry on failure, offline banner, visible usage limits
13. Greeting empty state ("Good morning, Ali") with 3-4 starter chips
14. Haptics on send and completion, fluid motion (Gemini "Neural Expressive" redesign)
15. Visible, editable memory and custom instructions
16. Entry points outside the app: home and lock-screen widgets, Action Button, Siri/Assistant shortcuts, **share sheet into the app**
17. Live Activities / Dynamic Island for long voice or tasks (ChatGPT)

**Differentiators that fit ContentStudio:** reply to post draft; schedule from chat with a confirm card; per-network post preview in chat (with character limits); share-sheet intake (share a photo, link or reel into AI chat); background generation with push that opens the result or approval queue; workspace-aware memory (brand voice, hashtags, banned words); approve or act on pending posts and inbox items from chat; voice-to-post.

Sources: 9to5Google I/O 2026, ai-toolbox ChatGPT sidebar guide, ChatGPT Learn Projects, TechCrunch and Android Central (Gemini widgets), Perplexity help center, Anthropic support (voice), OpenAI on X (branching), MacRumors (Live Activities), Gemini Help (deep research notifications), OpenAI memory controls, Releasebot and ClickUp ChatGPT changelogs.

## PO decisions on the competitor gaps (2026-09-30)

- **Speech-to-text: yes, highest priority.** Voice to text only, not a back-and-forth voice mode. Follow what Claude and ChatGPT do; no design story needed. Own story, split out of CONT-4202 (which keeps regenerate and edit).
  - Reuse the existing backend transcription endpoint: `POST ai/transcribeAudio` (`contentstudio-backend/routes/web/ai.php:51`, `AIController::transcribeAudio`, `app/Services/AI/OpenAITranscriptionService.php`, model `gpt-4o-transcribe` via `services.openai.transcription_model`). `TranscribeAudioRequest`: webm, mp3, wav, m4a, mp4, mpeg, mpga, max 25 MB. Same auth group the app already uses for `/ai/*`.
  - Web reference: `contentstudio-frontend/src/modules/AI-tools/composables/useVoiceInput.ts` (MediaRecorder + waveform, live Web Speech transcript, OpenAI blob fallback, silence auto-pause, `MAX_DURATION = 120` s). The app should allow longer dictation (10 minutes fits well under 25 MB as m4a).
  - Flutter has no audio or speech package today (nothing in `pubspec.yaml`). Candidates: a recorder package for m4a capture plus the platform recogniser (`speech_to_text`) for the live preview.
- **Starter suggestions: already covered** by CONT-3916 `[Flutter] Add feature pills and starter prompts to the mobile assistant` (AI chat starter prompts epic). No new story.
- **Memory and preferences: already covered** by CONT-4214 `[Flutter] Add AI Chat settings with Instructions, Memory and Usage to the app` (AI Chat settings epic). No new story.
- **Haptics and smooth motion: yes**, own story.
- **Home-screen widget (iOS and Android): yes**, own story. Includes app-icon quick actions.
- **Not now:** pinned chats, thumbs up/down, branch a chat, share a chat, real-time voice mode, Live Activities.
