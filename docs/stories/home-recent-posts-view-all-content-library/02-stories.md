# Stories: Home recent posts "View All" opens Content Library AI Creations

## [FE] Send Home "View All" on recent AI posts to Content Library AI Creations

### Description:

As a ContentStudio user, I want the **View All** button on my Home dashboard's recent AI posts to open Content Library on the AI Creations tab, so that I land where all my AI generated content now lives and can keep browsing, scheduling and managing those posts alongside the rest of my content.

Today View All takes users to the separate AI Library page in Publisher, which is not where we want users to manage their AI creations anymore.

---

### Workflow:

1. The user opens **Home** and sees the recent AI posts section with the **Generate post** and **View All** buttons.
2. The user clicks **View All**.
3. **Content Library** opens with **AI Creations** selected in the left sidebar and the **AI Posts** pill selected, showing the same posts they saw on Home plus the rest of their AI posts.
4. The user can switch to the **AI Studio** or **Clips** pills, or any other Content Library section, as usual.

---

### Acceptance criteria:

- [ ] Clicking **View All** in the recent AI posts section on Home opens Content Library, not the AI Library page in Publisher
- [ ] Content Library opens with **AI Creations** highlighted in the sidebar and the **AI Posts** pill selected
- [ ] The posts shown on Home appear in the AI Posts list for the same workspace
- [ ] It opens the Content Library of the workspace the user is currently in
- [ ] Browser back returns the user to Home
- [ ] Opening View All in a new tab (middle click or Ctrl/Cmd click) still works, since it remains a normal link
- [ ] The button label stays **View All**, and **Generate post** is unchanged

---

### Mock-ups:

N/A. No visual change, only where the button leads.

---

### Impact on existing data:

None.

---

### Impact on other products:

- Mobile app (Flutter): N/A, this Home section is web only.
- Chrome extension: N/A.
- White-label: works the same on white-label domains, the link follows the current domain.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (N/A, no new copy)
- [ ] UI theming support (N/A, no visual change)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, no new or changed API)
