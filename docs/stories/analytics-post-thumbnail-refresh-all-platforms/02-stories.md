# Stories: Analytics Post Thumbnail Refresh — All Platforms

---

## [BE] Extend post thumbnail refresh to all analytics platforms

**Epic:** https://app.shortcut.com/contentstudio-team/epic/24551

---

### Description:

As a ContentStudio user, I want post thumbnails in analytics to display correctly for all connected social accounts so that I can clearly identify each post when reviewing performance data.

We already refresh post thumbnails for Facebook — this story extends that same behavior to all remaining platforms supported in analytics: Instagram, LinkedIn, Pinterest, TikTok, Twitter, and YouTube.

Post thumbnail URLs stored in analytics can expire over time. Without a refresh mechanism, analytics views end up showing broken images for posts. The fix is to ensure the same refresh job that exists for Facebook is also running for every other platform.

---

### Workflow:

1. User opens Analytics and navigates to any report that shows individual posts (e.g., Posts Overview, platform-specific reports).
2. User sees post thumbnails displayed correctly for all platforms — Facebook, Instagram, LinkedIn, Pinterest, TikTok, Twitter, YouTube.
3. Even for older posts (published weeks or months ago), thumbnails load and display properly — no broken images.

---

### Acceptance criteria:

- [ ] Post thumbnails are refreshed and display correctly in analytics for Instagram posts
- [ ] Post thumbnails are refreshed and display correctly in analytics for LinkedIn posts
- [ ] Post thumbnails are refreshed and display correctly in analytics for Pinterest posts
- [ ] Post thumbnails are refreshed and display correctly in analytics for TikTok posts
- [ ] Post thumbnails are refreshed and display correctly in analytics for Twitter posts
- [ ] Post thumbnails are refreshed and display correctly in analytics for YouTube posts
- [ ] Older posts (20+ days) no longer show broken/expired thumbnail images across all platforms
- [ ] The refresh mechanism runs automatically without requiring any manual intervention

---

### Mock-ups:

N/A

---

### Impact on existing data:

Stored thumbnail URLs in ClickHouse for the affected platforms will be updated with fresh, valid URLs. No data is deleted — only the thumbnail URL fields are updated where expired.

---

### Impact on other products:

Analytics is web-only. No impact on mobile apps or Chrome extension.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing strings
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
