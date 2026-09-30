# Videos

Every video the pipeline generates lives here, one folder per video. Nothing video-related goes under `docs/`. A feature or story folder only carries a short `promo-video.md` that points back here.

## Layout

```
videos/
├── README.md                     this file
├── _tooling/                     shared renderer, used by every video
│   ├── render.js
│   └── package.json              puppeteer-core + ffmpeg-static (node_modules is gitignored)
└── <slug>/                       one folder per video
    ├── README.md                 what it is, which feature, scene list, caveats
    ├── source.html               the animation, the only thing you edit
    ├── <slug>-16x9.mp4           rendered output, one file per aspect ratio
    ├── <slug>-9x16.mp4           (optional) vertical cut
    └── stills/                   preview frames, gitignored
```

## Naming

- **Folder slug:** `<feature-or-story-slug>-<kind>`, where kind is `promo`, `demo`, `teaser` or `explainer`. Example: `ai-chat-skills-promo`.
- **Output file:** `<slug>-<ratio>.mp4`. Ratios: `16x9` (1920x1080), `9x16` (1080x1920), `1x1` (1080x1080), `4x5` (1080x1350).
- A second version of the same video replaces the MP4 in place. A different video, for example a demo of the same feature, gets its own folder.

## How a video is made

There is no video-generation model in this setup. A video is a scripted HTML animation rendered frame by frame in headless Chrome and encoded with ffmpeg. That makes it exact, repeatable and easy to edit.

`source.html` must:

1. Define `window.render(t)`, which draws the frame at `t` seconds. It must be **deterministic**: no timers, no `Date.now()`, no CSS animations or transitions. All motion is computed from `t`.
2. Define `window.DURATION` in seconds.
3. Lay out at a fixed viewport (1920x1080 unless you are rendering another ratio).
4. Load fonts from Google Fonts only, with a system fallback. The renderer waits for `document.fonts.ready`.

Opened in a normal browser, the page should autoplay on a loop (guard it with `if (!navigator.webdriver)`), so it can be previewed without rendering.

## Rendering

```bash
cd videos/_tooling
npm install                                          # first time only
node render.js <slug> --stills 3,12.5,20             # check key frames first
node render.js <slug>                                # 1920x1080 MP4
node render.js <slug> --size 1080x1920               # vertical cut, needs a layout that fits
node render.js <slug> --variant light                # opens source.html#light, writes <slug>-light-<ratio>.mp4
node render.js <slug> --range 22.6,30.2              # renders only that span to <slug>-<ratio>-range.mp4, to splice a fix into an existing render
```

**Splicing a `--range` render:** splice only into a clean full render, write to a new file (never over the input), and check the frame count before replacing anything (`ffmpeg -i out.mp4 -map 0:v -c copy -f null -`, which should equal duration x fps). Splicing into an already-spliced file broke its frame timing once, and the result silently came out half-length.

Chrome is expected at `/usr/bin/google-chrome`. Set `CHROME_PATH` to override. A 36-second video at 30 fps takes about 90 seconds.

## Content rules

- **Only claim what ships.** Check the feature's research and PRD before putting a capability on screen. Coverage limits apply (for example, best time to post is Facebook and Instagram only).
- **No em dashes** in on-screen copy.
- **Use real UI labels** from `contentstudio-frontend/src/locales/en/` where the video mimics the product.
- **Screens are styled mock-ups, not screenshots.** Say so in the video's README.
- **Audio is optional.** A video can have an original score made in code (see `ai-chat-promo/music.js`) that is mixed in after rendering. A voiceover is still added in an editor.

## Index

| Video | Feature | Length | Ratios |
|---|---|---|---|
| [ai-chat-skills-promo](ai-chat-skills-promo/) | AI Chat Skills | 36s | 16x9 |
| [ai-chat-promo](ai-chat-promo/) | AI Chat, full capability overview | 113s | 16x9 |
| [ai-chat-teaser](ai-chat-teaser/) | AI Chat, short cut of the promo (current and all-light themes) | 65s | 16x9, 16x9 light |
