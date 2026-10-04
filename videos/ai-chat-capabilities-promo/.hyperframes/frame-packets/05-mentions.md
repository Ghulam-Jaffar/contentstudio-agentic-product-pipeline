# Frame packet: 05-mentions

## Project inputs

- Project: /home/casper/code/contentstudio-agentic-product-pipeline/videos/ai-chat-capabilities-promo
- Design tokens: /home/casper/code/contentstudio-agentic-product-pipeline/videos/ai-chat-capabilities-promo/frame.md
- RULES_DIR: /home/casper/.claude/plugins/cache/hyperframes/hyperframes/0.8.112/skills/hyperframes-animation/rules

## Assigned storyboard block

## Frame 5 — @ mentions

- scene: Two uploaded images sit as chips @Image1 (blank amber bottle) and @Image2 (Bean & Co. logo); prompt "Put the logo from @Image2 on the label of @Image1" types with the mentions highlighted; the two inputs slide together and the real result (bottle with the logo on its label) springs up between them
- voiceover: "Upload your own images and mention them. Put the logo on the label. Done, in seconds."
- duration: 5.035s
- transition_in: push-slide LEFT
- status: outline
- src: compositions/frames/05-mentions.html
- type: feature_showcase
- persuasion: Friction reduction
- beat: awe
- blueprint: comparison-split (Adapt)
- asset_candidates: assets/bottle.png — blank amber bottle, @Image1; assets/brand-logo.png — Bean & Co. logo, @Image2; assets/bottle-with-logo.png — real edit result, bottle with logo
- focal: assets/bottle-with-logo.png
- roles: bottle = supporting (@Image1 card) · brand-logo = supporting (@Image2 card) · bottle-with-logo = cutout (result hero)
- sfx: pop, sparkle

narrativeRole: precise edits by referencing your own images.
keyMessage: combine your own assets with one sentence.

Adapt: keep the mirrored split-tilt entry + inner-edge badges; add a third beat where the two cards converge into the result.
Scene 1 (0–1.41s): eyebrow @ MENTIONS + headline "Use your own images."; two cards enter from opposite wings with mirrored book-open tilts: left the blank bottle, right the Bean & Co. logo; badges "@Image1" and "@Image2" pop on their inner edges.
Scene 2 (1.41–3.02s): a composer strip below types "Put the logo from @Image2 on the label of @Image1" with both mentions rendered as blue chips.
Scene 3 (3.02–5.04s): on "Done" both cards slide inward and tuck behind as the result image (bottle with the logo on its label) springs up centered between them, larger; hold.

## Selected blueprint: comparison-split

# comparison-split — Comparison Split-Cards

**intent**: Two paired items of equal weight shown side-by-side with mirrored 3D "book-open" tilts — the eye reads them as a balanced comparison, then a pill badge lands at each card's inner edge to punctuate. The motion IS the symmetry: two cards arriving from opposite wings into a held spread.

**roles served**

- Key_Feature (from `comparison-split-cards`): when two complementary features / capabilities of equal weight should be presented **simultaneously, not sequentially** — an A/B, a "X + Y together," paired concepts the viewer must weigh side-by-side. Not for >2 items (use `grid-card-assemble`) or sequential steps.

**duration**: 4–6s

**shot structure** (a `[bg]` canvas carrying two faint ambient glow blooms — `[accent A]` near 30%, `[accent B]` near 70% — so each side owns a color identity across a 50% symmetry axis; equal-width cards under one shared perspective parent)

- **Scene 1 (0.0–~0.8s) — title sets the concept.** A centered `[title line]` with an `[accent keyword]` slides DOWN into place from just above (a short smooth settle). The downward arrival is deliberate: it forms a non-conflicting T-shape against the cards, which arrive from the sides next.
- **Scene 2 (~0.4–1.9s) — the split-tilt entry (signature move).** Two equal-width feature cards arrive from opposite wings — `[left card]` from the left, `[right card]` from the right ~0.2s behind — each carrying a **mirrored 3D `rotateY` tilt** (left faces right, right faces left, opening like a book) and scaling ~0.85→1 as it lands. The entry overlaps the title's tail so the whole thing reads as ONE arrival, not two beats. Each card holds `[image / label / subtitle]`; box-shadows fall **outward** from the tilt (left shadow right, right shadow left).
- **Scene 3 (~1.9–end) — badges punctuate, then hold.** A pill `[badge]` lands at each card's **inner edge** (left then right, ~0.3s apart), overlapping its card ~15% so it reads as attached, not orbiting. This is the lone overshoot in the shot — it earns the punctuation. Settles and holds.

**motion vocabulary**: title slide-down from above; mirrored opposite-wing card entry; static book-open `rotateY` tilt (`+tilt` left, `−tilt` right); tilt-matched outward box-shadow; inner-edge badge spring-pop; gentle phase-opposed idle float (left vs right, never synchronized) registered as subtle jitter; dual side-glow ambient.

**rule mapping**

- two cards entering from opposite wings with mirrored `rotateY` tilts + tilt-matched shadow → `split-tilt-cards` (the signature; keep the two-layer split so the entry `x`/`scale` and the idle never collide on one alias)
- title slide-down settle → `gsap-effects` (translate + opacity on a long-tail `power3`)
- inner-edge pill badge pop (the one overshoot) → `spring-pop-entrance` (overshoot register — earns the punctuation)
- phase-opposed idle float on the pair → `sine-wave-loop` (low-amplitude register — subtle jitter, NOT lazy breathing; left `sin(t)`, right `sin(t+π)` so they never conveyor-belt)
- the two faint side glows behind the cards → `ambient-glow-bloom` (un-triggered soft bloom, one per accent)

**camera modifier**: camera-static by default — the symmetry is the subject and a move would break the balance.
