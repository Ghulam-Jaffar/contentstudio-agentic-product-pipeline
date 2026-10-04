# AI Chat capabilities promo (HyperFrames build)

- **Feature:** ContentStudio AI Chat (web), full capability tour
- **Length:** 59.2s, 1920x1080, voiceover + music + synced captions
- **Output:** `ai-chat-capabilities-promo-16x9.mp4`
- **Created:** 2026-10-02

Same angle as `../ai-chat-promo`, rebuilt from scratch with the **HyperFrames** framework (`heygen-com/hyperframes`, plugin v0.8.112) and its `product-launch-video` workflow, so the team can compare it with the hand-built HTML videos. It is the first video in this folder with a voiceover.

Unlike the other videos it does **not** use `_tooling/render.js`. It's a HyperFrames project: `index.html` assembles 11 sub-compositions from `compositions/frames/`, plus audio and caption tracks, and the HyperFrames CLI renders it.

## How it was made

| Step | Output |
|---|---|
| Brief | `BRIEF.md` (capability tour, 60s, 16:9, automation, no storyboard review) |
| Source material | `capture/` assembled by hand (no website crawl): real logo and the real Nano Banana Pro / Veo 3.1 outputs from `../ai-chat-teaser/media/`, UI labels from `contentstudio-frontend/src/locales/en/` |
| Design system | `frame.md`, the `blue-professional` preset recoloured to #1f6bff, Inter Tight (`assets/fonts/`) |
| Story | `STORYBOARD.md` (11 frames, Demo Loop arc, one shot blueprint per frame) and `SCRIPT.md` |
| Voice | Kokoro `af_heart`, local (no HeyGen login) → `assets/voice/` |
| Music | MusicGen small, local → `assets/bgm/track.wav`, bed at 0.12 under the voice |
| Captions | word timings from Whisper small.en (Kokoro gives none), corrected to the script → `compositions/captions.html` |
| Frames | one sub-agent per frame, built in parallel → `compositions/frames/NN-*.html` |
| Checks | `hyperframes lint` (0 errors) and `hyperframes check` (passed), contact sheets in `snapshots/` |

## Scenes

| Time | Frame | On screen |
|---|---|---|
| 0 to 5.8s | 01 hook | "Your day:" cycles calendar, image studio, inbox, analytics, then the AI Chat composer crashes in |
| 5.8 to 12.9s | 02 intro | Chat home: logo, "Good afternoon, Alex!", composer, starter pills |
| 12.9 to 18.6s | 03 publishing | Week strip, empty slots drafted into a post plan, Approve, "3 posts scheduled" |
| 18.6 to 24.1s | 04 images | Nano Banana Pro · 1:1 · Cinematic, 2x2 grid of the real flat lays, "22 image models · 31 styles", Add to Composer |
| 24.1 to 29.2s | 05 @ mentions | @Image1 bottle + @Image2 logo, the prompt, the real edited result |
| 29.2 to 33.4s | 06 video | Veo 3.1 · 8s · 16:9 · Start frame, the real Veo clip in a player, Schedule with this video |
| 33.4 to 38.3s | 07 carousels | 6-slide LinkedIn carousel in the slide viewer, "Slide 3 · Template" |
| 38.3 to 45s | 08 inbox | 3 reviews / 5 DMs / 4 comments, a Google review with a drafted reply, Approve, "Reply posted" |
| 45 to 50s | 09 analytics | Reach 184.2K, Engagement 9.6K, Followers +1,240, trend line, one insight |
| 50 to 55s | 10 skills | `/` menu (My Skills, Workspace Skills, Built by ContentStudio), `/content-plan for October`, the week's plan |
| 55 to 59.2s | 11 close | "Plan it. Make it. Post it. Measure it.", logo + ContentStudio, "Just ask.", "Try it now" |

## Rebuild

```bash
export HYPERFRAMES_PYTHON=$HOME/.venvs/hyperframes/bin/python   # Kokoro, MusicGen, Whisper live here
npx hyperframes lint && npx hyperframes check
npx hyperframes preview                                          # Studio, scrub the timeline
npx hyperframes render --quality high --output renders/video.mp4
```

Edit a single frame in `compositions/frames/NN-*.html`. Don't re-run `assemble-index.mjs`: it already hoisted frame 06's two video clips into `index.html`, and re-running would drop them. The music `<audio id="el-bgm">` was also added to `index.html` by hand, because the workflow only fetches music from HeyGen.

## Caveats

- Screens are styled mock-ups, not product screenshots. The brand (Bean & Co.), people and numbers are made up. The images and the video clip are real ContentStudio AI outputs (see `../ai-chat-teaser/README.md`).
- Carousel and inbox are shown on the PO's 2026-09-30 confirmation (code on the `features` branch).
- @ mentions cover uploaded images only. Best time to post isn't shown.
- The voice is the local Kokoro model. Signing in to HeyGen (`npx hyperframes auth login`) and re-running the audio step would give a HeyGen voice and a library music track.
- `hyperframes check` leaves 12 colour-contrast warnings on small labels (eyebrows, @ badges, placeholder text).
