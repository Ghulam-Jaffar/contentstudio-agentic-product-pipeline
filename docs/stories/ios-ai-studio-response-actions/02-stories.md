### Description:
As a ContentStudio iOS user, I want the AI Studio response card to show `Copy` and `Replace` inside the response footer so the action buttons feel clearly attached to the generated text and not the follow-up action pills.

---

### Workflow:

1. User opens AI Studio in the iOS Composer and submits a prompt.
2. AI Studio returns generated text and renders the response card.
3. The follow-up action pills appear directly below the generated text inside the same card.
4. A thin divider separates the follow-up pills from the response action footer.
5. The `Copy` and `Replace` buttons appear in the bottom-right corner of the response card footer.
6. User taps `Copy` to copy the response text to the clipboard.
7. User taps `Replace` to replace the currently rendered response text using the existing replace flow.

---

### Acceptance criteria:

- [ ] The AI Studio response card includes an inline footer with `Copy` and `Replace` buttons.
- [ ] The footer is positioned at the bottom right of the generated text card.
- [ ] A thin separator line visually divides the follow-up pills from the footer button area.
- [ ] The `Copy` and `Replace` buttons remain inside the response card and do not appear below the follow-up pills outside the card.
- [ ] The button layout remains correct whether the generated text is short or spans multiple lines.
- [ ] Tapping `Copy` copies the full generated response text to the clipboard.
- [ ] Tapping `Replace` triggers the existing response replacement flow for the current AI Studio output.
- [ ] No new backend or data model work is required.

---

### Mock-ups:
N/A — design is a layout refinement of the existing iOS AI Studio response card.

---

### Impact on existing data:
No data changes. This is a UI-only iOS layout refinement.

---

### Impact on other products:
No impact on the web app or Android app. This is an iOS-only Composer UI update.

---

### Dependencies:
None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness
- [ ] Multilingual support
- [ ] UI theming support
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
