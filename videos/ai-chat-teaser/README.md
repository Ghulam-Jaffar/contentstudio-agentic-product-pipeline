# AI Chat teaser

- **Feature:** ContentStudio AI Chat (web), the short cut of `ai-chat-promo`
- **Length:** about 66s, 60 fps, with music
- **Outputs:** `ai-chat-teaser-16x9.mp4` (current theme: dark intro, carousel and end card), `ai-chat-teaser-light-16x9.mp4` (all-light theme), both with music, plus `ai-chat-teaser-music.mp3` (music only)
- **Created:** 2026-09-30

## How it's built

`source.html` is a copy of `../ai-chat-promo/source.html` with a teaser layer added at the end. The `CUT` table lists which promo scenes play, their start and end in promo time, and a playback speed. Consecutive scenes overlap by 0.3s so their own fades crossfade. Two scenes are left out: the home screen and the "And more" cards. Every other scene plays at 1.6x to 1.7x speed. The intro, the word montage and the end card play at normal speed.

A fix to a scene in the promo doesn't carry over to this copy automatically. Copy it across, keeping the teaser layer: the `ONLY` / `MULTI` guards in `vis()` and in the cursor code, and everything after `/* ---------- teaser cut`.

## Variations and branding

- **Two themes from one source.** `source.html` renders the current theme by default. `source.html#light` renders the all-light version: no dark backgrounds, dark ink for the text, and deeper colours for the gradient word and the montage words. Render it with `node render.js ai-chat-teaser --variant light`.
- **Logo and CTA end card.** The ContentStudio logo is the real file, `contentstudio-frontend/src/assets/img/logo/logo-blue.png`, copied here as `logo.png` and embedded in `LOGO()`. Don't redraw it, because `cs_logo.svg` in the same repo draws the bars smaller. Put it before the name anywhere ContentStudio appears as a brand, and never follow the name with "AI Chat". The end card has a large logo with a soft glow, then "Just ask.", then the chat composer with "Try it now" typing into it and the cursor clicking a glowing send button (with a click in the music), then "Plan, create, publish and measure. All from one chat.", then a small logo and ContentStudio lockup at the bottom.
- **Copy.** The headlines were rewritten from a marketing angle and no longer match the promo's. For example, "Plan the week in one sentence.", "On-brand visuals, on demand.", "Carousels that design themselves." and "New look? Just say so."

## Scenes (teaser time)

| Time | Scene |
|---|---|
| 0 to 5.3s | Intro: "Plan it. Make it. Post it. Measure it." then "Now just ask." |
| 5 to 12.8s | Publishing: calendar table, filling empty slots, approving the plan |
| 12.5 to 18.2s | Images |
| 17.9 to 23s | @ mentions |
| 22.7 to 28.6s | Video |
| 28.3 to 36.2s | Carousel maker, with the template fan-out at about 32.4s |
| 35.9 to 42.5s | Inbox |
| 42.2 to 48.3s | Analytics |
| 48 to 53.1s | Skills |
| 52.8 to 58s | Word montage, in scene order: Publishing, Images, Video, Carousels, Inbox, Analytics, Skills |
| 57.7 to 64.7s | CTA end card: logo, "Just ask.", a composer with "Try it now" typed and sent, ContentStudio lockup |

## Music

`music.js` is an original score, synthesized in code with no samples and no licensed audio. It keeps the promo score's arrangement and sound design, which the PO liked:

- A 120 BPM groove
- A whoosh on every cut
- Clicks, success chimes and reveal sparkles (no send sound, the PO found it too much)
- The riser and drop on the carousel fan-out
- A stab on each montage word, and the final hit on "Just ask."

The instruments that carry the tune are warmer, because the promo's synth arp and pad sounded too synthetic. A felt piano model (inharmonic partials, two detuned strings per note, slightly varied timing and touch) plays the broken chords and the melody. Warm strings replace the pad, and the bass is rounder. Cues are read from the same `CUT` table, and promo moments that were cut out (the home screen and the "And more" cards) are skipped.

If you change `CUT`, copy the new table into `music.js`, then rebuild:

```bash
node music.js /tmp/teaser.wav
../_tooling/node_modules/ffmpeg-static/ffmpeg -y -i /tmp/teaser.wav -af loudnorm=I=-14:TP=-1.5 -b:a 192k ai-chat-teaser-music.mp3
# render.js writes a silent MP4, so add the music back after every re-render
../_tooling/node_modules/ffmpeg-static/ffmpeg -y -i ai-chat-teaser-16x9.mp4 -i ai-chat-teaser-music.mp3 -map 0:v -map 1:a -c:v copy -c:a aac -b:a 192k -shortest out.mp4 && mv out.mp4 ai-chat-teaser-16x9.mp4
```

## Real AI generations

The image and video results are real outputs from ContentStudio AI, generated on 2026-10-01 through the staging MCP server (`CS-Staging` connector, **jenna benna** workspace, 55 staging credits). They're kept in `media/` as local working files only. `videos/*/media/` is gitignored, because the rendered MP4s are the deliverable. The originals stay in that workspace's media library on staging. To re-render later, download them back into `media/` (base URL `https://storage.googleapis.com/contentstudio-media-library-nearline/media_library/6920d8857de69e43c50e3377/uncategorized/original/`) and rebuild the frames with `ffmpeg -ss 3 -t 3.8 -i media/latte-pushin.mp4 -vf "fps=24,scale=1120:630" -q:v 3 media/vframes/f%03d.jpg`:

| File | Model | Used in |
|---|---|---|
| `flatlay-1.png` to `flatlay-4.png` (`YEmDGyFsbBPjyRr`, `AHEJEocN8RXitKD`, `I0kd9lBeROrMYn5`, `v2G6mMIJoFbLAiw`) | Nano Banana Pro, 1024x1024 | Images grid, post-plan thumbnails, carousel slides, video start frame |
| `bottle.png` (`uM3lPezf8XznZy5`), `brand-logo.png` (`N9mXF7N9uEFpou4`) | Nano Banana Pro | @ mentions inputs (Image1, Image2) |
| `bottle-with-logo.png` (`JKPf5imMsyvca7y`) | Nano Banana Pro, both inputs as reference images | @ mentions result |
| `latte-pushin.mp4` (`6mOOfmmLRUbUTFo`) | Veo 3.1, image to video from `flatlay-1.png`, 8s, 1080p, no audio | Video player. Frames 3s to 6.8s are extracted to `media/vframes/` (24 fps) and play at 1x speed |

The End frame slot shows the product's empty-state placeholder (a cloud-upload icon and "End Frame", captioned "Optional") and is never filled. AI Chat supports an end frame (`end_frame_url`), but the public API's video generation, which the MCP server uses, only takes a start image. The procedural drawings the scenes used before are still in `source.html` as `flatlayDrawn()`, `bottleDrawn()` and `logoArtDrawn()`, but nothing calls them.

## Caveats

Same as the promo: the screens are mock-ups built from real UI labels, and the brand and numbers are made up. The image and video results are real (see above). Carousel and inbox are live code on the `features` branch, shown on the PO's confirmation. @ mentions cover uploaded images only.
