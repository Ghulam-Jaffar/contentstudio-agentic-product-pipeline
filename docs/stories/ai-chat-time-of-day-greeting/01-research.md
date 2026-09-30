# AI Chat: time-of-day greeting on the new chat screen — Research

## Current State

- The new-chat (hero) screen shows a static headline above the composer: **"Let's make something scroll-stopping"**.
- Rendered by `contentstudio-frontend/src/modules/AI-tools/components/NewChatHero.vue` (lines 6-11) as an `<h2>` with `text-xl font-normal text-cs-text-strong`, truncated to one line. It reads small, grey and flat against the dotted gradient background, which is the "typography isn't right" problem.
- Key: `ai_tools.starter_prompts.hero_title`, translated in all 10 locales (`en, es, it, el, fr, de, zh, sv, cs, pl`) under `src/locales/<lang>/ai_tools.json`.
- `NewChatHero` is used on two surfaces from `src/modules/AI-tools/ChatBox.vue` (`heroSurface`: `studio` = full AI Chat page, `modal` = chat modal/drawer) and on the dashboard via `src/components/dashboard/AIAssistantSection.vue`, which passes `:show-title="false"` and keeps its own "Plan, create & schedule smarter with AI Studio" headline. **The dashboard is out of scope.**
- The user's first name is available on the profile store (`profileStore.getProfile.first_name`, already used in `MediaGenerationOptions.vue` for preferences and elsewhere for names). First name is optional, so the greeting needs a no-name fallback.
- No existing greeting / time-of-day helper exists in the frontend (grep for `greeting`, `good_morning` returned nothing relevant).

## What Needs to Change

- Replace the static headline with a greeting built from **time of day on the viewer's device** + **first name**, chosen from a pool of variants per time band.
- Pick the variant once when the new-chat screen opens; it must not change while the user types or when the composer re-renders. A new chat can show a different variant.
- Name fallback: variants without the name when `first_name` is empty.
- Typography: larger (roughly 28-32px on desktop, 24px on mobile width), medium weight, strong text colour, with the name picked out in the primary colour (theme variable, so white-label works). Allow two lines instead of truncating.
- Translate every variant in all 10 locales. Translators should adapt, not translate literally ("night owl" won't survive every language).
- Remove the old `hero_title` key once replaced (only consumer is `NewChatHero`).

## Decisions / Proposals for the PO

- **Clock source: the viewer's device time**, not the workspace timezone. Workspace timezone drives publishing; a greeting is personal and should match the user's own morning.
- **Time bands:** Morning 5:00-11:59, Afternoon 12:00-16:59, Evening 17:00-21:59, Late night 22:00-4:59.
- **Small weekday flavour** (optional, one Monday and one Friday variant) to keep it feeling alive.
- Web only for now. The Flutter AI assistant screen has its own empty state; not mentioned in the request, so no `[Flutter]` story (flagged as follow-up).

## UX Reference

Claude.ai greets by time band and name ("Good afternoon, Sam", "Evening, Sam", "Hey there, night owl") in a large serif headline above the composer. ChatGPT and Gemini use a single large "Hello, Sam" / "What can I help with?" line. Common pattern: one large, warm line, name emphasized, nothing else competing with it.

## Files Involved

- `contentstudio-frontend/src/modules/AI-tools/components/NewChatHero.vue` — headline markup + styling
- New composable, e.g. `contentstudio-frontend/src/modules/AI-tools/composables/useChatGreeting.ts` — band + variant selection
- `contentstudio-frontend/src/locales/*/ai_tools.json` — new `starter_prompts.greeting.*` keys, remove `hero_title`
- `contentstudio-frontend/src/modules/AI-tools/ChatBox.vue` — no change expected (already passes `surface`)

## Update after PO review (2026-09-30)

- Long greetings are too big for phones. Each variant is now tagged Short or Long; narrow chat areas (under 600px) rotate Short only. Tagged per variant rather than measured, so translations behave predictably. Added "Still going strong, {name}?" (evening) and "Still up, {name}?" (late night) so every band has at least two Short options.
- Width should come from the chat container, not the viewport, because `ChatBox.vue` renders both the full page (`studio`) and the modal (`modal`) surfaces. A `ResizeObserver` or `useElementSize` on the hero container is the natural hook.
- Previewer: https://claude.ai/artifact/R8gfb4a2LzEK9JLwhdWszR
