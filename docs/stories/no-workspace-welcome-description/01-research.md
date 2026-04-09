# Research: No-Workspace Welcome Page Description

## Current State

`contentstudio-frontend/src/modules/setting/components/workspace/StartTrial.vue` shows:
- "Welcome to ContentStudio" heading
- YouTube video embed
- "Start Your Trial" button
- Sign out link

No description text exists between the heading and video — user has no context for why they're on this page.

## What Needs to Change

- Add a description paragraph between the `<h2>` and the video embed

## Files Involved

- `contentstudio-frontend/src/modules/setting/components/workspace/StartTrial.vue`
- All locale directories under `src/locales/` (new i18n key)
