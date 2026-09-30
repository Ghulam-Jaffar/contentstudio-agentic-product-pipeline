# Stories: Carousel generation in the Flutter AI chat

Epic: **AI Carousel post generation** (existing)

1. `[Flutter] Generate, review and schedule AI chat carousels in the mobile app`

---

# [Flutter] Generate, review and schedule AI chat carousels in the mobile app

### Description

As a ContentStudio mobile user, I want to ask AI chat for a carousel on my phone, swipe through the slides in order, and schedule it as a carousel post, so that I get the same carousel experience I get on the web.

AI chat on the web can plan a carousel, show the price for approval, generate an ordered set of slides with one caption, edit a named slide, and schedule it. The mobile app has no carousel handling of its own. It shows the slides as a loose set of images, not necessarily in order, with no single caption and no way to schedule them as one carousel post. This story brings the mobile app to parity. It also prepares it for slides arriving one at a time, which is coming with **AI Carousel 2.0: Carousel Maker and improvements**.

---

### Workflow

1. User opens AI chat on their phone and types "Make a 6-slide carousel about our spring sale".
2. The AI shows the plan and the price with **Generate**, **Refine** and **Cancel**, as for other paid tasks.
3. User taps **Generate**.
4. A carousel card appears in the answer. The slides appear in order as they're ready, and the user can swipe between them, with a "Slide 2 of 6" counter.
5. Under the slides, the user sees the one caption for the whole carousel.
6. User types "Make slide 3 more punchy". Only slide 3 changes, and the carousel shows the new version in place.
7. User taps a slide to see it full screen and swipes through the rest there.
8. User taps **Schedule**. The Composer opens with all 6 slides in order as one carousel post and the caption filled in.

---

### Acceptance criteria

**Generating**

- [ ] Carousel requests show the plan and price approval card with Generate, Refine and Cancel. Nothing is charged until the user taps Generate
- [ ] While the carousel is being made, the step list shows progress as for other tasks
- [ ] If the server sends the number of slides before the slides themselves, the carousel card shows one placeholder per slide and fills each in as it arrives. If it doesn't, the card appears when the first slide arrives

**Showing the carousel**

- [ ] A generated carousel shows as one carousel card, never as separate images
- [ ] Slides always appear in their intended order, including when they arrive out of order
- [ ] The user can swipe between slides, with a "Slide {current} of {total}" counter and dots
- [ ] The carousel's one caption shows once under the slides, with copy and edit actions as for other generated text
- [ ] Tapping a slide opens a full-screen viewer that swipes through all slides in order
- [ ] The slide the user is viewing doesn't jump when another slide arrives or changes

**Editing**

- [ ] When the user asks for a change to a named slide ("slide 3"), only that slide updates in place, and the others stay the same
- [ ] An edited carousel keeps its caption unless the user asked to change it

**Using it**

- [ ] **Schedule** and **Save as draft** open the Composer with all slides in order as one carousel post, and the caption filled in
- [ ] For a selected platform that doesn't support carousels, the Composer shows its existing message for that platform, not an error
- [ ] Carousels made on the web show correctly when the chat is opened on mobile, and the other way round

**General**

- [ ] Works on iOS and Android, on small and large phones
- [ ] All copy is translated in every supported language

---

### Mock-ups:

Follow the carousel card from the web design (`[Design] Carousel slide composition and the AI chat carousel experience`), adapted to the app. If the mobile AI chat redesign lands first, follow **[Design] Design mobile AI chat to feel like a native LLM app**.

---

### Impact on existing data:

None. Uses the same carousel data as web.

---

### Impact on other products:

- **Web:** none. This matches **[FE] Review, edit and schedule a generated carousel from AI chat**.
- **AI Carousel 2.0:** slide-by-slide delivery from **[BE] Send each carousel slide to the chat as soon as it is ready** is handled by the placeholder behaviour above. The 2.0 template picker shows as a plain text question on mobile until a mobile picker is scoped.

---

### Dependencies:

- **[BE] Generate carousels as an ordered, themed slide set with one caption**
- **[BE] Edit a carousel by regenerating the slide the user names**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage: N/A, nothing API-facing changes
