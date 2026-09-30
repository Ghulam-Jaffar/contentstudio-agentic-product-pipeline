# AI Chat promo

- **Feature:** ContentStudio AI Chat (web), full capability overview
- **Length:** 113s, 60 fps, with music
- **Outputs:** `ai-chat-promo-16x9.mp4` (video + music), `ai-chat-promo-music.mp3` (music only)
- **Created:** 2026-09-30

The whole video follows one fictional coffee brand, Bean & Co., so the scenes connect. The flat lay the chat generates becomes the video's start frame, and the carousel, inbox and analytics scenes all belong to the same brand.

## Scenes

| Time | Scene |
|---|---|
| 0 to 6.6s | Dark open: "Plan it. Make it. Post it. Measure it." then "Now just ask." A circle wipe reveals the chat |
| 5.2 to 13.7s | AI Chat home: "Good afternoon, Alex!", composer, starter pills. The Post pill opens and "What's scheduled for this week?" is picked |
| 13.3 to 26.2s | **Publishing:** the Working timeline, a table of this week's posts, "Draft posts to fill my empty slots", a post plan confirmation card and Approve |
| 25.8 to 35.2s | **Images:** Image mode with Nano Banana Pro, 1:1, Cinematic. A 2x2 grid of generated images, with Add to Composer and Schedule with this image. "22 image models, 31 styles" |
| 34.8 to 43s | **@ mentions:** two uploaded images, "Put the logo from @Image2 on the label of @Image1", then the result |
| 42.6 to 52s | **Video:** Video mode with Veo 3.1, 8s, 16:9, audio on, start and end frames. The video generates and plays, with a ticker of video models |
| 51.6 to 64.6s | **Carousel maker** (dark): a 6-slide LinkedIn carousel in the slide viewer, swiped through, then five templates fanned out in 3D |
| 64.2 to 75.4s | **Inbox:** inbox summary tiles, a Google review, an Instagram DM and a Facebook comment. A drafted reply to the review is approved and posted |
| 75 to 85.4s | **Analytics:** reach, engagement and follower tiles, a line chart against the previous period, engagement by account, and one insight |
| 85 to 93.2s | **Skills:** the `/` menu (My Skills, Workspace Skills, Built by ContentStudio), `/content-plan for October`, and a week of planned posts |
| 92.8 to 101.6s | **And more:** voice input, inviting a teammate, reconnecting an expired account, the Brand toggle |
| 101.2 to 106.4s | Word montage: Publishing. Carousels. Inbox. Images. Video. Analytics. Skills. Workspace. |
| 106 to 113s | End card: "Just ask." ContentStudio AI Chat, with the composer |

## Music

`music.js` is an original score, synthesized in code with no samples or licensed audio, so there is nothing to license. It runs at 120 BPM from the chat reveal, and every hit is placed on a timestamp from `source.html`: a whoosh on each scene cut, clicks on each Approve press, a chime on each success, a drop when the carousel templates fan out, a stab on each montage word, and the final chord on "Just ask."

If you change the timings in `source.html`, update the times in `music.js` to match, then rebuild:

```bash
node music.js /tmp/music.wav
../_tooling/node_modules/ffmpeg-static/ffmpeg -y -i /tmp/music.wav -af loudnorm=I=-14:TP=-1.5 -b:a 192k ai-chat-promo-music.mp3
# render.js writes a silent MP4, so add the music back after every re-render
../_tooling/node_modules/ffmpeg-static/ffmpeg -y -i ai-chat-promo-16x9.mp4 -i ai-chat-promo-music.mp3 -map 0:v -map 1:a -c:v copy -c:a aac -b:a 192k -shortest out.mp4 && mv out.mp4 ai-chat-promo-16x9.mp4
```

## Caveats

- The screens are styled mock-ups, not product screenshots. The labels come from `contentstudio-frontend/src/locales/en/` and the chat components: the placeholder, Working / Worked for, Approve / Change / Cancel, Add to Composer, Schedule with this image/video, "Slide N · Template", the skill menu groups and the starter prompts.
- The brand, people, posts and numbers are made up.
- **Carousel and inbox** are shown because the PO confirmed on 2026-09-30 that they're in scope. The code for both is on the `features` branch of `contentstudio-ai-agents` and `contentstudio-frontend`, not on `main`/`master`.
- **@ mentions** are shown for uploaded images only (`@Image1`, `@Image2`), because that is all the chat supports.
- Model and style counts (22 image models, 31 styles) come from `locales/en/ai_tools.json` and `imageGenerationStyles.ts`. Update them if those lists change.
- Best time to post isn't shown, because it only covers Facebook and Instagram.
- The primary blue (`#1f6bff`) is approximate, the same value used in the Skills promo.
