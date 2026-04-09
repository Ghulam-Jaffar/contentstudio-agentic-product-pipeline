# Stories: No-Workspace Welcome Page Description

---

## Story 1: [FE] Add description text to the no-workspace welcome page

### Description:

As a user who lands on the welcome page with no active workspace, I want to understand why I'm here and what to do next so that I'm not confused by a page with just a heading, video, and button.

Currently, `StartTrial.vue` shows "Welcome to ContentStudio", a YouTube video, and a "Start Your Trial" button with no explanatory text. This story adds a short description between the heading and video.

**File:** `contentstudio-frontend/src/modules/setting/components/workspace/StartTrial.vue`

---

### Workflow:

1. User logs in but has no active workspace
2. User lands on the welcome page showing "Welcome to ContentStudio"
3. Below the heading, user sees a description: "You don't have an active workspace. Start a free trial to create your workspace and begin managing your social media."
4. User watches the video or clicks "Start Your Trial"

---

### Acceptance criteria:

- [ ] A description paragraph is added between the `<h2>` heading and the YouTube video embed
- [ ] Copy: "You don't have an active workspace. Start a free trial to create your workspace and begin managing your social media."
- [ ] Text is centered, muted color (`text-gray-500` or similar neutral), reasonable font size (`text-base` or `text-lg`)
- [ ] The description uses an i18n key — add to all locale directories
- [ ] Existing layout (heading, video, button, sign out link) remains unchanged

---

### Mock-ups:

N/A — single line of text added between existing heading and video.

---

### UI Copy:

- Description: "You don't have an active workspace. Start a free trial to create your workspace and begin managing your social media."

---

### Impact on existing data:

None.

---

### Impact on other products:

- Web App: No-workspace welcome page only
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (text should wrap properly on smaller screens)
- [ ] Multilingual support (new i18n key added to all locale directories)
- [ ] UI theming support — N/A, plain text with neutral color
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
