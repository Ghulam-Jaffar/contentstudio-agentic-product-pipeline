# Frame packet: 01-hook

## Project inputs

- Project: /home/casper/code/contentstudio-agentic-product-pipeline/videos/ai-chat-capabilities-promo
- Design tokens: /home/casper/code/contentstudio-agentic-product-pipeline/videos/ai-chat-capabilities-promo/frame.md
- RULES_DIR: /home/casper/.claude/plugins/cache/hyperframes/hyperframes/0.8.112/skills/hyperframes-animation/rules

## Assigned storyboard block

## Frame 1 — One chat, not five tools

- scene: Tool names cycle in one accent slot (Calendar, Image studio, Inbox, Analytics), then the AI Chat composer crashes in and shoves the text aside
- voiceover: "A calendar. An image studio. An inbox. Analytics. Four tools, four tabs. Or, one chat."
- duration: 5.845s
- transition_in: cut
- status: outline
- src: compositions/frames/01-hook.html
- type: hook
- persuasion: Negative contrast
- beat: overwhelm → curiosity
- blueprint: ticker-takeover (Adapt)
- asset_candidates:
- focal: the AI Chat composer pill (hero)
- roles: composer = cutout (hero) · lead-in text group = supporting
- sfx: whoosh, impact-soft
- handoff_out: element=AI Chat composer pill (white, 100px radius, 1056x112px, blue round send button at its right end, placeholder "Type your message here, or / for a skill"); center x=960 y=560; scale=1; opacity=1; motion=at rest (jitter settled to zero by the cut)

narrativeRole: names the tool sprawl the audience lives in, then replaces it with one object.
keyMessage: everything you juggle can be one chat.

Adapt: keep the ticker + collision signature; lead-in is the VO list itself, hero is the AI Chat composer pill, not a logo.
Scene 1 (0–0.7s): white canvas; a large display line seats left-of-center reading "Your day:" with an empty accent slot beside it — Centered-left, big type ~70% width.
Scene 2 (0.7–3.51s): the accent slot rolls vertically through "a calendar." → "an image studio." → "an inbox." → "analytics." in blue, one roll per VO word (vertical spring-ticker, smooth not bouncy).
Scene 3 (3.51–4.56s): on "Or," the AI Chat composer pill (white, rounded, blue send button, placeholder text) crashes in from off-screen right with a motion-blur streak and SHOVES the whole text group off to the left (reactive displacement), landing heavy at center.
Scene 4 (4.56–5.84s): on "one chat" the composer sits alone dead-center, ~55% width, a small eyebrow "ONE CHAT" fades in above it; holds with subtle jitter only.

## Selected blueprint: ticker-takeover

# ticker-takeover — Ticker Displace / Takeover

**intent**: A context phrase types in, an accent word cycles through options like a slot-machine to suggest "this could be many things," then a hero CRASHES in from off-screen and physically shoves the text aside — "actually, this is what it is." A collision, not a fade.

**roles served**

- Hook (from `takeover-ticker-displace`): when a static lead-in phrase + a cycling accent word should be **physically replaced** (not cross-dissolved) by a hero arriving with momentum, and the final frame is the hero alone. Reach for it when the takeover should read as an impact.
- Brand_Outro: the same collision used as a sign-off — options cycle, the brand mark crashes in and owns the frame.

**duration**: 5–7s

**shot structure** (a `[bg]` canvas; one text group on the left/center that gets ejected by an incoming hero)

- **Scene 1 (0.0–~1.4s) — context build.** A typewriter lays down a `[lead-in phrase]` character-by-character (smooth, no typos — selling confidence, not human chaos). Camera static.
- **Scene 2 (~1.4–3.0s) — the cycling beat.** An `[accent word]` slot inside the line ticks through 2–3 `[options]` on a vertical spring-roll (each click a new word), suggesting breadth — "many things this could be." (More than ~3 reads as filler.)
- **Scene 3 (~3.0–4.2s) — the collision (signature move).** A `[hero]` crashes in from off-screen with momentum and physically SHOVES the whole text group aside — the text reacts to the impact (gets displaced), it does not fade. The hero lands **heavy** — a longer settle, not a zip — so it reads as mass, not speed.
- **Scene 4 (~4.2–end) — the hero alone.** The hero settles dead-center and reads still. Holds.

**motion vocabulary**: smooth character typewriter; vertical spring-ticker word roll (2–3 steps); off-screen hero crash-in with momentum; reactive displacement of the struck text group; heavy long-tail landing (not bouncy); dual-axis subtle jitter on the resting hero.

**rule mapping**

- smooth single-phrase typewriter lead-in → `discrete-text-sequence` (smooth-slice / continuous `floor(progress)` form — no typo machinery)
- accent word slot-machine cycling through options → `vertical-spring-ticker` (`STEPS` = number of options the hero will replace; the rule's footer-reveal is unused — Scene 3 takes its place)
- hero shoves the text group aside on impact → `reactive-displacement` (the text is the displaced mass; express the hero's "heavy land" as a longer `power2` settle, not the rule's default `back.out`)
- hero's fast off-screen crash-in → `motion-blur-streak` (directional velocity blur resolving sharp as it lands)
- resting-hero aliveness → `sine-wave-loop` (low-amplitude dual-frequency register — scale + rotation jitter composing onto the hero's final landed scale; never a yoyo around 1)

**camera modifier**: camera-static — the displacement happens in element space (the hero moves the text), so there is no real camera move; the impact is the only motion.
