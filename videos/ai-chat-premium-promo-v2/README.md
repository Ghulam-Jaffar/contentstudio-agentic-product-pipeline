# AI Chat premium promo, v2 (reference-driven)

- **Feature:** ContentStudio AI Chat (web), capability film
- **Length:** 68.25s, 1920x1080, 30 fps with true motion blur, music only
- **Output:** `ai-chat-premium-promo-v2-16x9.mp4`
- **Created:** 2026-10-03

Same story, copy and look as `../ai-chat-premium-promo` (v1). This version was rebuilt after studying the reference library in `claude video pipeline/repos/` and `.claude/skills/`. It exists to show what difference the references make, so keep both versions to compare.

## What changed from v1, and where each change came from

| Change | v1 | v2 | Reference |
|---|---|---|---|
| Timing | Seconds picked by hand | One beat grid in `timing.js` (68.57 BPM, 2-bar sections), read by both `source.html` and `music.js`. Cuts, camera landings, headlines and chords sit on beats | PDoomVideo `src/core.js` (BPM/OFF/B(n)), Austerlitz `film.js` buildTimeline, awesome-ai-motion beat-synced ad prompt |
| Camera | Continuous Catmull-Rom glide that never settles | Fast eased legs (power2 in-out repositions, power4-out dives), then holds where the camera sits still with a 2-sine drift (4 px / 3 px / 0.25°). Scale interpolates in log space | ClaudeAnimationBase ANIMATION_GUIDE ("fast actions, slow meanings"), HyperFrames `rules/3d-camera-flight.md`, Remotion `perceptual-scale` |
| Motion blur | CSS blur scaled by camera speed, which also softened text | True shutter blur: 4 sub-frames per frame across a 180° shutter, averaged (`render.js --shutter 4`) | awesome-ai-motion prompts (sub-frame tmix), HyperFrames `components/motion-blur`, Opus dataset (Fr_Sorrentino) |
| Headlines | Lines fade up | Kicker first. Words rise out of blur one at a time with a soft landing ease, the gradient line tracks in from wide letter-spacing, and the support line follows. Exits are faster than entrances | HyperFrames `per-word-rise`, `tracking-in`, `typography.md` (-0.045em display tracking) |
| Light | None | A light band sweeps across the window as each headline lands (screen blend, 115°, linear) | HyperFrames `light-sweep-pass`, `gloss-sweep` |
| Finish | None | Fine per-frame grain (overlay 6%), a vignette stretched horizontally, a very slight warm grade | Austerlitz `engine/shaders.js` finish pass |
| UI motion | Ease-out entrances, even typing | Critically damped springs on every card. Typing has a human rhythm with pauses at word breaks. The AI reply streams in word by word from grey to ink | HyperFrames springEase, `rules/discrete-text-sequence.md`, `components/streaming-text` |
| Ending | Logo and tagline | Ends on the composer again ("Try it now" typing under the lockup), echoing the opening | ClaudeAnimationBase ("rhyme the ending with the opening") |
| Score | Additive strings, generic | String sections with 4 detuned voices per note (±3/±9 cents), random phase, vibrato that fades in, band-limited and low-passed. Swells into each headline. Minor colours first, resolving to F major on "Just ask."; the reverb tail darkens. Still no sound effects. Content above 4 kHz is 39 dB below the mix | Austerlitz `tools/mix.py` (string recipe, darkening IR, minor → major arc) |
| Process | Contact sheet at the end | Hard moments checked as stills before the full render. Bugs caught that way: gradient words rendering invisible, and an over-warm grade | Opus dataset QA gates, ClaudeAnimationBase review budget |

## Rebuild

```bash
node music.js music.wav
cd ../_tooling && node render.js ai-chat-premium-promo-v2 --fps 30 --shutter 4
cd ../ai-chat-premium-promo-v2
../_tooling/node_modules/ffmpeg-static/ffmpeg -y -i ai-chat-premium-promo-v2-16x9.mp4 -i music.wav -map 0:v -map 1:a -c:v copy -af loudnorm=I=-16:TP=-1.5 -c:a aac -b:a 192k -shortest out.mp4 && mv out.mp4 ai-chat-premium-promo-v2-16x9.mp4
```

The full render with shutter blur takes about an hour, because every frame is captured 4 times. Change timings only in `timing.js`; both the picture and the music follow it.

## Caveats

Same as v1. The screens are mock-ups built from real UI labels. The brand and numbers are made up. The images and video clip are real ContentStudio AI outputs, loaded from `../ai-chat-teaser/media/` (gitignored).
