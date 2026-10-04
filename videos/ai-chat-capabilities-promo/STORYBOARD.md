---
format: 1920x1080
duration: 60s
message: "Plan, create, publish and measure. All by asking one chat."
arc: Demo Loop with feature-benefit progression (hook → product intro → 8 demo cycles → CTA)
audience: social media managers, marketers and agencies on or evaluating ContentStudio
mode: autonomous
music: confident upbeat minimal tech, light pulse, modern SaaS launch
---

## Video direction

- **palette system** (from `frame.md`): canvas `bg` white #ffffff with very light cool-gray UI panels (#f5f5f7); `primary` ContentStudio blue #1f6bff is the only accent: send button, active chips, Approve, eyebrows, numerals, chart line, progress bar, @mention highlights. Headlines `text` #111827 near-black, body `text-muted` #6b7280. Cards are the preset's tinted cards (4% blue fill, 20% blue 1.5px border, 10–14px radius, no shadow) except the floating chat window itself, which may carry one soft neutral shadow so it reads as the product surface. `positive` green only for a posted/approved check and the up-chip on analytics.
- **type**: display ramp (Inter Tight 600–700, −0.02em) for headlines and numerals; body ramp (Inter Tight 400) for chat messages and UI text; eyebrows uppercase 0.08em in blue.
- **recurring surface**: every feature frame (3–10) is built around the same AI Chat window: a rounded white chat panel (~62% of frame width, centered or offset per frame) with the composer pinned at its bottom (placeholder "Type your message here, or / for a skill", blue round send button). Prompts type into the composer, then pop up as a right-aligned user bubble (light blue tint); the assistant answers left-aligned with a small "Working" shimmer, then "Worked for Ns". A per-frame eyebrow + short headline sits top-left outside the window (e.g. eyebrow PUBLISHING, headline "Fill the week."), revealed when the VO names the capability. This repeated surface is what binds the film.
- **motion grammar + reveal model**: long-tail `power3` settles everywhere; no bounce except the one playful logo pop on Frame 11 and the comparison badges. Reveal each piece on its spoken cue: the prompt types as the VO starts the ask, the answer arrives as the VO names the result, buttons press on "Approve". Nothing is on screen before the VO reaches it except the persistent chat window chrome.
- **rhythm / held frames**: Frame 2 ends on a held read of the full chat home (breather after the hook). Frame 9 ends held on the insight card. Frame 11 holds the lockup for ~1.5s at the end. Everything else reveals continuously to its VO.
- **negative list**: no purple/blue "AI" gradients, no bokeh, no floating particles, no dark backgrounds, no fake browser chrome or OS cursor; the only cursor is a small custom pointer used for clicks. Neither failure mode: no slideshow (front-load then freeze), no screensaver (many elements floating independently). No lazy breathing, no slow back-half pans. No em dashes in on-screen copy. Caption band (bottom ~17%) kept clear of key content.

## Frame 1 — One chat, not five tools

- scene: Tool names cycle in one accent slot (Calendar, Image studio, Inbox, Analytics), then the AI Chat composer crashes in and shoves the text aside
- voiceover: "A calendar. An image studio. An inbox. Analytics. Four tools, four tabs. Or, one chat."
- duration: 5.845s
- transition_in: cut
- status: animated
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


## Frame 2 — Meet AI Chat

- scene: The chat home: ContentStudio logo mark, "Good afternoon, Alex!", the composer with placeholder "Type your message here, or / for a skill", starter pills (Post, Image, Video, Carousel, Analytics, Inbox) pop in one by one as the VO lists the verbs
- voiceover: "Meet AI Chat in ContentStudio. Plan, create, publish and measure your social media, just by asking."
- duration: 7.019s
- transition_in: zoom-through
- status: animated
- src: compositions/frames/02-intro.html
- type: product_intro
- persuasion: Value stacking
- beat: clarity + ease
- blueprint: prompt-type-submit-generate (Adapt)
- asset_candidates: assets/contentstudio-logo.png — real ContentStudio logo mark (blue circle, white bars)
- focal: assets/contentstudio-logo.png
- roles: contentstudio-logo = supporting (small brand mark in the chat header)
- sfx: pop-soft
- handoff_in: element=AI Chat composer pill (white, 100px radius, 1056x112px, blue round send button at its right end, placeholder "Type your message here, or / for a skill"); center x=960 y=560; scale=1; opacity=1; motion=at rest, then it becomes the composer at the bottom of the expanding chat window

narrativeRole: lands the whole promise (the message) by beat 2; every later frame is evidence.
keyMessage: one chat runs your whole social workflow.

Adapt: Product_Intro composer variant; the composer from Frame 1 grows into the full chat home instead of a prompt typing.
Scene 1 (0–2.11s): the composer from Frame 1 (handoff) expands upward into the full AI Chat window (card morph); the small ContentStudio logo mark sits in the header with "AI Chat"; headline "Good afternoon, Alex!" types in above the composer as the VO says "Meet AI Chat in ContentStudio" — Centered, window ~62% width.
Scene 2 (2.11–5.38s): starter pills pop in under the composer one per spoken verb, staggered on the VO: "Post" (plan), "Image" + "Video" (create), "Carousel" (publish), "Analytics" (measure), then "Inbox".
Scene 3 (5.38–7.02s): on "just by asking" the caret blinks in the composer and the send button glows once; held read of the full home (breather).


## Frame 3 — Publishing

- scene: Prompt "What's scheduled for this week?" types and sends; a compact week table answers; second prompt "Draft posts to fill my empty slots"; "Working" status theater, then a post-plan card with three drafted posts (flat lay thumbnails) and Approve / Change / Cancel; Approve gets pressed and the slots fill
- voiceover: "Ask what's scheduled this week. Let it draft posts for the empty slots. Approve, and they're booked."
- duration: 5.696s
- transition_in: crossfade
- status: animated
- src: compositions/frames/03-publishing.html
- type: feature_showcase
- persuasion: Show-don't-tell proof
- beat: control
- blueprint: agent-progress-theater (Reproduce)
- asset_candidates: assets/flatlay-1.png — latte flat lay, post thumbnail; assets/flatlay-2.png — rosetta latte flat lay, post thumbnail; assets/flatlay-3.png — pumpkin latte flat lay, post thumbnail
- focal: assets/flatlay-1.png
- roles: flatlay-1 = supporting (post thumbnail) · flatlay-2 = supporting (post thumbnail) · flatlay-3 = supporting (post thumbnail)
- sfx: click, chime

narrativeRole: the core job, scheduling, done by asking.
keyMessage: an empty week fills itself, with you approving.

Reproduce (agent-progress-theater, conversation-thread variant).
Scene 1 (0–1.3s): eyebrow PUBLISHING + headline "Fill the week." top-left; chat window right 60/40; "What's scheduled for this week?" types into the composer and pops up as a user bubble.
Scene 2 (1.3–2.44s): assistant answers with a compact Mon–Sun week strip: four days with a post chip, three days empty and dashed.
Scene 3 (2.44–3.91s): on "draft posts for the empty slots" the second prompt "Draft posts to fill my empty slots" sends; "Working" shimmer; a post-plan card cascades in with three drafted rows (flat lay thumbnail, day + time, one-line caption), each popping on its own beat.
Scene 4 (3.91–5.7s): on "Approve" the custom cursor clicks the blue Approve button (press + ripple, next to Change and Cancel); the three dashed days in the week strip fill with blue post chips and a small green check reads "3 posts scheduled"; hold.


## Frame 4 — Images

- scene: Image mode: chips Nano Banana Pro, 1:1, Cinematic; prompt "Autumn latte flat lay for Bean & Co." submits; a 2x2 grid of real generated flat lays streams in tile by tile; hover shows "Add to Composer" and "Schedule with this image"; a stat pill "22 image models · 31 styles"
- voiceover: "Describe the image you want. Pick from twenty two models and thirty one styles. Ready to post."
- duration: 5.568s
- transition_in: push-slide LEFT
- status: animated
- src: compositions/frames/04-images.html
- type: feature_showcase
- persuasion: Feature-to-benefit translation
- beat: excitement
- blueprint: prompt-type-submit-generate (Reproduce)
- asset_candidates: assets/flatlay-1.png — real Nano Banana Pro flat lay; assets/flatlay-2.png — real flat lay; assets/flatlay-3.png — real flat lay; assets/flatlay-4.png — real flat lay
- focal: assets/flatlay-1.png
- roles: flatlay-1 = cutout (hero tile, top-left of the grid) · flatlay-2 = supporting · flatlay-3 = supporting · flatlay-4 = supporting
- sfx: whoosh-soft, sparkle

narrativeRole: creation, visuals made in the same conversation.
keyMessage: on-brand images on demand.

Reproduce (Key_Feature, sub-shape B).
Scene 1 (0–1.67s): eyebrow IMAGES + headline "Make the visual." top-left; chat window; mode chips Nano Banana Pro · 1:1 · Cinematic light up blue one by one; "Autumn latte flat lay for Bean & Co." types and sends.
Scene 2 (1.67–3.34s): "Working" shimmer then a 2x2 grid of the four real flat lays resolves tile by tile (blur-to-sharp, left-to-right) inside the assistant message.
Scene 3 (3.34–5.57s): on "twenty two models, thirty one styles" a blue stat pill "22 image models · 31 styles" springs beside the window; the first tile lifts slightly and its action row appears: "Add to Composer" and "Schedule with this image"; hold.


## Frame 5 — @ mentions

- scene: Two uploaded images sit as chips @Image1 (blank amber bottle) and @Image2 (Bean & Co. logo); prompt "Put the logo from @Image2 on the label of @Image1" types with the mentions highlighted; the two inputs slide together and the real result (bottle with the logo on its label) springs up between them
- voiceover: "Upload your own images and mention them. Put the logo on the label. Done, in seconds."
- duration: 5.035s
- transition_in: push-slide LEFT
- status: animated
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


## Frame 6 — Video

- scene: Video mode chips Veo 3.1, 8s, 16:9, Start frame (flat lay thumbnail); prompt "Slow push in, steam rising" submits; a player card plays the real Veo clip; it slides aside for a "Schedule with this video" button
- voiceover: "Love that shot? Turn it into an eight second video, without leaving the chat."
- duration: 4.245s
- transition_in: crossfade
- status: animated
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


## Frame 7 — Carousels

- scene: Prompt "Make a 6 slide LinkedIn carousel about our autumn menu"; a slide viewer card swipes through slides (flat lay images with headline overlays), counter "Slide 3 · Template"; slides fan out to show all six
- voiceover: "Need a LinkedIn carousel? Six slides, on brand, designed for you in one go."
- duration: 4.864s
- transition_in: push-slide LEFT
- status: animated
- src: compositions/frames/07-carousel.html
- type: feature_showcase
- persuasion: Value stacking
- beat: delight
- blueprint: device-surface-showcase (Adapt)
- asset_candidates: assets/flatlay-2.png — slide image; assets/flatlay-3.png — slide image; assets/flatlay-4.png — slide image
- focal: assets/flatlay-2.png
- roles: flatlay-2 = cutout (slide 1 image) · flatlay-3 = supporting (slide image) · flatlay-4 = supporting (slide image)
- sfx: swipe, whoosh

narrativeRole: multi-slide formats, usually a design job, handled in chat.
keyMessage: carousels design themselves.

Adapt (showcase-carousel variant): the surface is the chat's slide viewer.
Scene 1 (0–1.36s): eyebrow CAROUSELS + headline "Six slides. Zero design time."; "Make a 6 slide LinkedIn carousel about our autumn menu" types and sends.
Scene 2 (1.36–3.31s): a square slide viewer card appears showing slide 1 (flat lay with a big headline overlay "Autumn at Bean & Co."); it swipes to slide 2 and 3 (different flat lays, overlays "Pumpkin spice is back", "Cinnamon oat latte"), counter reads "Slide 3 · Template".
Scene 3 (3.31–4.86s): on "designed for you" the six slides fan out in a gentle arc behind the viewer (1–6 numbered); hold.


## Frame 8 — Inbox

- scene: Prompt "Summarize my inbox"; summary tiles (12 unread, 3 reviews, 5 DMs) count in; a Google review card appears with a drafted reply; Approve is pressed, the reply posts with a check
- voiceover: "Catch up on your inbox. Reviews, DMs and comments, each with a drafted reply. Approve it in a click."
- duration: 6.72s
- transition_in: crossfade
- status: animated
- src: compositions/frames/08-inbox.html
- type: feature_showcase
- persuasion: Friction reduction
- beat: relief
- blueprint: cursor-ui-demo (Adapt)
- asset_candidates:
- focal: the drafted review reply card
- roles: (typography/UI only)
- sfx: click, chime

narrativeRole: engagement, the daily grind, compressed into approvals.
keyMessage: answer your audience without opening another tool.

Adapt: locked static stage; element swaps plus one cursor click.
Scene 1 (0–1.57s): eyebrow INBOX + headline "Reply without switching tabs."; "Summarize my inbox" types and sends.
Scene 2 (1.57–3.36s): three summary tiles count up left to right as the VO says "reviews, DMs and comments": "3 reviews", "5 DMs", "4 comments".
Scene 3 (3.36–5.15s): a Google review card (5 stars, "Best oat latte in town!", Maya R.) slides in with a drafted reply under it ("Thank you, Maya! See you for the autumn menu.").
Scene 4 (5.15–6.72s): on "in a click" the cursor presses Approve; the reply collapses to a green "Reply posted" check; hold.


## Frame 9 — Analytics

- scene: Prompt "How did we do last month?"; three metric cards count up (Reach 184.2K, Engagement 9.6K, Followers +1,240); a line chart draws against the previous period; one insight card: "Reels drove 61% of reach. Post more of them on weekdays."
- voiceover: "Ask how last month went. Get your reach, engagement and growth, plus what to do next."
- duration: 4.992s
- transition_in: push-slide LEFT
- status: animated
- src: compositions/frames/09-analytics.html
- type: feature_showcase
- persuasion: Statistical proof
- beat: confidence
- blueprint: dataviz-countup (Adapt)
- asset_candidates:
- focal: Reach metric card
- roles: (typography/data only)
- sfx: tick, riser-soft

narrativeRole: measurement, closing the loop on plan, create, publish.
keyMessage: answers, not dashboards.

Adapt (Key_Feature montage-cut mode, light): locked frame, element-level count-ups, one chart draw.
Scene 1 (0–1s): eyebrow ANALYTICS + headline "Know what worked."; "How did we do last month?" types and sends.
Scene 2 (1–2.83s): three metric cards count up in sequence: Reach 184.2K (+12% green chip), Engagement 9.6K (+8%), Followers +1,240.
Scene 3 (2.83–3.83s): a line chart draws left to right in blue against a faint gray previous-period line.
Scene 4 (3.83–4.99s): on "what to do next" an insight card with a blue left rule lands: "Reels drove 61% of your reach. Post more on weekdays."; held read.


## Frame 10 — Skills

- scene: Typing "/" opens the skills menu with three groups (My Skills, Workspace Skills, Built by ContentStudio); "/content-plan" is picked and "for October" is added; a week of planned posts cascades in as rows
- voiceover: "Save your playbooks as Skills. Type a slash, pick one, and the whole plan runs for you."
- duration: 5.056s
- transition_in: crossfade
- status: animated
- src: compositions/frames/10-skills.html
- type: benefit_highlight
- persuasion: Future pacing
- beat: power
- blueprint: cursor-ui-demo (Adapt)
- asset_candidates:
- focal: the skills menu
- roles: (UI only)
- sfx: click, pop

narrativeRole: repeatability, the reason to make it a habit.
keyMessage: explain it once, run it every time.

Adapt: the menu pick is the trigger, the plan cascade is the payoff.
Scene 1 (0–1.35s): eyebrow SKILLS + headline "Explain it once."; the caret types "/" and a skills menu springs up above the composer with three group labels: My Skills, Workspace Skills, Built by ContentStudio, each with 2 rows.
Scene 2 (1.35–2.7s): the row "/content-plan" highlights and is picked; it becomes a blue skill chip in the composer and "for October" types after it; send.
Scene 3 (2.7–5.06s): on "the whole plan runs" a week of planned posts cascades in as rows (Mon–Fri: platform dot, title, time), each popping on a beat; hold.


## Frame 11 — Just ask

- scene: Four verbs hard-cut at center (Plan it, Make it, Post it, Measure it), then the real ContentStudio logo springs in with a soft blue glow and the wordmark; "Just ask." lands under it; a small composer types "Try it now"
- voiceover: "Plan it, make it, post it, measure it. ContentStudio. Just ask."
- duration: 4.117s
- transition_in: zoom-through
- status: animated
- src: compositions/frames/11-cta.html
- type: cta
- persuasion: Risk reversal
- beat: inevitability
- blueprint: logo-assemble-lockup (Adapt)
- asset_candidates: assets/contentstudio-logo.png — real ContentStudio logo mark
- focal: assets/contentstudio-logo.png
- roles: contentstudio-logo = cutout (hero mark)
- sfx: impact-soft, chime

narrativeRole: the brand lockup and the one action.
keyMessage: just ask.

Adapt (CTA, word-beat lead-in resolving into a settled logo bloom): no push-through; the verbs clear and the mark spring-blooms on the cleared stage.
Scene 1 (0–2.2s): white stage; four short words hard-cut in at center one per spoken verb, display ramp, near-black with the verb in blue: "Plan it." → "Make it." → "Post it." → "Measure it." (in-place token cycle, each swap on its VO word).
Scene 2 (2.2–3.0s): on "ContentStudio" the last word clears and the real ContentStudio logo springs in centered from zero (the one playful overshoot) with a soft blue glow blooming behind it; the wordmark "ContentStudio" slides out to its right to complete the lockup.
Scene 3 (3.0–4.12s): on "Just ask." the line "Just ask." lands large below the lockup, and a small composer pill fades up under it with "Try it now" typing and the blue send button glowing; everything holds still to the final frame.
