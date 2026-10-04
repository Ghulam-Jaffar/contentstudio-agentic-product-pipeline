# AI Chat premium promo

- **Feature:** ContentStudio AI Chat (web), capability film
- **Length:** 66s, 1920x1080, 60 fps, music only (no voiceover, no captions, no sound effects)
- **Outputs:** `ai-chat-premium-promo-16x9.mp4`
- **Created:** 2026-10-03

A premium, Apple-style cut of the AI Chat story. One continuous 3D camera travels through the chat window: it opens close and angled on the composer while the prompt types, follows the answer as it builds, pushes into the detail, then pulls back to a wide shot. There, a short marketing headline lands beside the window. The window alternates sides each section. A light motion blur and a soft depth-of-field falloff make the camera read like a lens.

It grew out of `../ai-chat-style-test`, the 12-second test of the Images section that the PO approved.

## Sections

| Time | Section | Headline |
|---|---|---|
| 0 to 4.6s | Intro | Plan. Create. Publish. Measure. → Just ask. |
| 4.6 to 11.6s | Publishing: week view, post plan, Approve, "3 posts scheduled" | Fill the week. / Ask for drafts. Approve. Done. |
| 11.6 to 18.6s | Images: Nano Banana Pro, 4 real flat lays, Add to Composer | On-brand visuals, on demand. / 22 image models. 31 styles. |
| 18.6 to 25.6s | Image editing: @Image1 + @Image2 → real edited result | Your assets. Your words. / Mention an upload. Change it in a sentence. |
| 25.6 to 32.6s | Video: Veo 3.1, start frame, the real Veo clip | Stills, set in motion. / Veo 3.1 and more, right in the chat. |
| 32.6 to 39.6s | Carousels: slide viewer, 6 slides | Six slides. Done for you. / LinkedIn carousels, designed in seconds. |
| 39.6 to 46.6s | Inbox: summary, Google review, drafted reply, "Reply posted" | Every reply, ready to send. / Reviews, DMs and comments, drafted for you. |
| 46.6 to 53.6s | Analytics: count-ups, trend line, next-step insight | Numbers that talk back. / Ask how you did. Hear what to do next. |
| 53.6 to 60.6s | Skills: `/` menu, `/content-plan for October`, the plan | Explain it once. / Save a Skill. Run it with a slash. |
| 60.6 to 66s | End card | Just ask. · ContentStudio logo · Plan, create, publish and measure. All from one chat. |

## Music

`music.js` is an original score, synthesized in code with no samples or licensed audio: dark warm strings, a soft low felt-piano pulse at 80 BPM and a round bass, with one chord per section. There are deliberately **no sound effects** (no whooshes, chimes, sparkles, risers or clicks) and **nothing in the high register**. Every bus runs through a low-pass so nothing sounds bright or "8-bit". The only accents are soft low chords on each intro word, when each headline lands, and on the final "Just ask."

If the section timings in `source.html` change (`INTRO`, `SL`), change the same constants in `music.js`, then rebuild:

```bash
node music.js music.wav
cd ../_tooling && node render.js ai-chat-premium-promo --fps 60
cd ../ai-chat-premium-promo
../_tooling/node_modules/ffmpeg-static/ffmpeg -y -i ai-chat-premium-promo-16x9.mp4 -i music.wav -map 0:v -map 1:a -c:v copy -af loudnorm=I=-16:TP=-1.5 -c:a aac -b:a 192k -shortest out.mp4 && mv out.mp4 ai-chat-premium-promo-16x9.mp4
```

## Caveats

- The screens are styled mock-ups built from real UI labels (`contentstudio-frontend/src/locales/en/`), not product screenshots. The brand (Bean & Co.), people and numbers are made up.
- The images, the edited bottle and the video clip are real ContentStudio AI outputs, loaded from `../ai-chat-teaser/media/` (gitignored, see that README to restore them).
- Carousel and inbox are shown on the PO's 2026-09-30 confirmation. @ mentions cover uploaded images only. Best time to post isn't shown.
