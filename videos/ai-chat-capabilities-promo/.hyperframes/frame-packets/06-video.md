# Frame packet: 06-video

## Project inputs

- Project: /home/casper/code/contentstudio-agentic-product-pipeline/videos/ai-chat-capabilities-promo
- Design tokens: /home/casper/code/contentstudio-agentic-product-pipeline/videos/ai-chat-capabilities-promo/frame.md
- RULES_DIR: /home/casper/.claude/plugins/cache/hyperframes/hyperframes/0.8.112/skills/hyperframes-animation/rules

## Assigned storyboard block

## Frame 6 — Video

- scene: Video mode chips Veo 3.1, 8s, 16:9, Start frame (flat lay thumbnail); prompt "Slow push in, steam rising" submits; a player card plays the real Veo clip; it slides aside for a "Schedule with this video" button
- voiceover: "Love that shot? Turn it into an eight second video, without leaving the chat."
- duration: 4.245s
- transition_in: crossfade
- status: outline
- src: compositions/frames/06-video.html
- type: feature_showcase
- persuasion: Show-don't-tell proof
- beat: excitement
- blueprint: video-text-pivot (Adapt)
- asset_candidates: assets/latte-pushin.mp4 — [video] real Veo 3.1 push-in over the latte flat lay; assets/flatlay-1.png — start frame thumbnail
- focal: assets/latte-pushin.mp4
- roles: latte-pushin = cutout (video hero in a player card) · flatlay-1 = supporting (Start frame thumbnail chip)
- sfx: whoosh

narrativeRole: motion content, the expensive format, made by asking.
keyMessage: video is one prompt away.

Adapt: keep the video-yields-and-slides-aside signature; the "stat" is the schedule action, the closing text is the headline.
Scene 1 (0–1.13s): eyebrow VIDEO + headline "Now make it move."; composer with chips Veo 3.1 · 8s · 16:9 and a "Start frame" chip showing the flat lay thumbnail; "Slow push in, steam rising" types and sends.
Scene 2 (1.13–2.97s): a large rounded player card scales up center playing the real Veo clip (3–4s section), a thin blue progress bar along its bottom.
Scene 3 (2.97–4.25s): the player slides aside left and down a little as a blue "Schedule with this video" button and a "8s · 1080p" meta chip pop in the vacated right space; hold.

## Selected blueprint: video-text-pivot

# video-text-pivot — Video → Text Pivot

**intent**: A product video holds center and claims attention, then slides aside to hand its weight to a hero stat in the space it vacates, then both clear and kinetic text types into the center — accent words carrying the meaning the video used to carry — sealed by a gradient pill. The arc is "show → yield → pivot → stamp," and each handoff pairs an exit with a same-anchor entrance so two beats read, not four.

**roles served**

- Product_Intro (from `metric-video-text-pivot`): when the open is "see the feature" then "see the impact" and the `[product video]` must stay visible through the stat reveal — it slides, it doesn't cut.
- Key_Feature: a feature clip that yields to a frame-filling metric and a typographic impact line.

**duration**: 6–8s

**shot structure** (a `[bg]` canvas; one `[product video]` as a real muted `.mp4` clip, a hero stat, then kinetic text — each pair shares a screen anchor so the handoff reads as a weight-transfer)

- **Scene 1 (0.0–~1.6s) — the video shows.** The `[product video]` lands centered on a smooth scale-up and breathes (a small y-bob), claiming full attention. Camera static.
- **Scene 2 (~1.6–3.2s) — yield + stat (signature move).** The video SLIDES aside (x + scale down) **into the very space** the `[hero stat]` now fills as the stat pops in with 3D-depth type — one weight-transfer reading as a single event, not two. The stat breathes within this window.
- **Scene 3 (~3.2–5.0s) — pivot to text.** Both video and stat clear out and kinetic `[impact text]` TYPES into the vacated center, character by character; its `[accent words]` carry the meaning the video used to carry.
- **Scene 4 (~5.0–end) — stamp.** A gradient `[pill]` snaps shut around the closing line (`scaleX` 0→1), its glow halo resolving a beat behind so the silhouette reads before the bloom — sealing the statement as one graphic. Holds.

**motion vocabulary**: video scale-in + small breath; weight-transfer slide (video x + scale-down handing off to the stat at the same anchor); 3D-depth stat type; character-stream typing; gradient pill scaleX-snap; glow-halo bloom trailing the silhouette.

**rule mapping**

- video entrance (smooth) and the weight-transfer slide → `gsap-effects` (scale/opacity then x + scale on a long-tail `power3`); the video itself is a muted `<video class="clip">` direct child of the root
- hero stat's frame-filling 3D type → `3d-text-depth-layers` (static-depth variation — layers built at setup, no cascade fighting the entry)
- the same-anchor video-exit ↔ stat-entry handoff (if treated as a morph) → `scale-swap-transition` (shared center)
- character-by-character impact typing through segmented spans → `dynamic-content-sequencing` (clean character stream) or `discrete-text-sequence`
- pill `scaleX` snap + trailing glow halo → `gsap-effects` (scaleX) + `ambient-glow-bloom` (the halo, resolving a beat behind)
- video / stat breath within their windows → `sine-wave-loop` (low-amplitude register — subtle jitter, gated to each element's window, never a forever loop)

**camera modifier**: camera-static — all motion is element-space (the video translates), so the "pivot" is the elements moving, not a camera.
