# Prototype UI prompt — Webhooks (paste into Lovable / v0 / similar)

> Tool-agnostic prompt to generate a high-fidelity clickable prototype of the Webhooks UI. It's scoped to v1 (publishing-lifecycle events) and matches the current ContentStudio API page. The real build will be re-skinned with `@contentstudio/ui`; this is for visual/UX validation only.

---

Design a high-fidelity, clickable web UI prototype for a new **Webhooks** section inside ContentStudio's existing **API** page. ContentStudio is a social media management SaaS; this is a developer-facing settings area.

**Visual style (match the existing API page):** Light theme only — no dark mode, no RTL. Clean and minimal: light-gray page background, white cards with subtle gray borders and ~12px rounded corners, a blue primary action color (#157FFF), gray-900 headings, gray-500/600 secondary text, small UPPERCASE section labels. Inter / system sans-serif for UI text; a monospace font for URLs, secrets, headers, event names, and JSON. Desktop-first but responsive.

**Page context:** The API page has a header card ("API — Manage your API key, monitor usage, and connect integrations" with a "View API Docs" link), an "API Requests" usage meter, and a tab bar with **"API Key"** and **"Request Logs"** tabs. **Add a third tab: "Webhooks".** Everything below lives under this tab. Inside the Webhooks tab, add a **sub-toggle with two views — `Endpoints` and `Deliveries`** (a smaller segmented control, not new top-level tabs): **Endpoints** = the webhook list + create/edit (screens 1–4); **Deliveries** = a global feed of recent deliveries across all webhooks, filterable by webhook, event, and status (screen 5). Opening a specific webhook from Endpoints shows its detail with that webhook's own deliveries (the Deliveries feed pre-filtered to it).

Design these screens:

**1. Empty state (no webhooks yet)** — a rich, centered setup card (like a "getting started" screen, NOT a bare empty state):
- Centered icon (webhook/plug) in a soft circle.
- Title: "Set up your first webhook"
- Body: "Get notified the moment things happen in your workspace. Register a URL, choose the post events you care about, and we'll send a signed request to your endpoint in real time — no polling required."
- A "How it works" row of **three step cards** (Step 1 / Step 2 / Step 3), each with a small icon, a title, and a one-line body:
  - **Step 1 — Create a webhook:** "Add your endpoint URL, an optional signing secret, and pick the events you want."
  - **Step 2 — We send events to your URL:** "When a post is created, scheduled, published, or fails, we POST a signed payload to your endpoint."
  - **Step 3 — Verify and handle deliveries:** "Check the signature, return a 2xx, and track every delivery in the log — failed ones retry automatically."
- A centered primary button **"Create your first webhook"** and a **"View Webhook Docs"** link directly below it.
- (This mirrors ContentStudio's existing "Set up your first approval workflow" screen — same three-step centered layout.)

**2. Webhooks list (when one or more exist)**
- A list of rows/cards. Each shows: the **Name**, the destination **URL** (monospace, truncated), a few small **event tags** (e.g. `post.published`, `post.failed`, `+1`), a **status pill** (Active = green, Paused = gray), a **last-delivery** indicator (green check or red dot + relative time like "2m ago"), and a **"⋮" menu** (Edit, Send test event, View deliveries, Pause/Resume, Delete).
- "Add Webhook" button top-right. Show a small counter like "3 of 5 webhooks used".

**3. Create / Edit webhook (slide-over panel or modal)** — title "New Webhook", subtitle "Configure a new webhook endpoint":
- **Name** — placeholder "My Webhook"; helper "A label to recognize this webhook."
- **Payload URL** — placeholder "https://myapp.com/webhooks/contentstudio"; helper "We'll send a POST request here." (required; must be https)
- **Secret (optional)** — placeholder "your-secret-key"; helper "Used to sign each delivery with an HMAC-SHA256 signature in the X-ContentStudio-Signature header, so you can verify the request came from ContentStudio." Include a small **"Generate"** button that fills a random secret.
- **Custom Headers (optional)** — an **"Add Header"** button that appends key/value input rows (e.g. `Authorization` / `Bearer …`); helper "Sent with every delivery — useful if your endpoint needs its own authentication."
- **Events** — a grouped checklist. Group header **"Posts"** with a "Select all" toggle. Five checkboxes, each with a one-line description in muted text:
  - `post.created` — "A post is created (draft or scheduled)."
  - `post.scheduled` — "A post is scheduled for a future time."
  - `post.published` — "A post is successfully published."
  - `post.failed` — "A post fails to publish on all platforms."
  - `post.partial` — "A post publishes on some platforms but fails on others."
  - Below the active group, show a **grayed-out, disabled** group labeled "Accounts · Inbox · Comments · Reviews" with a "Coming soon" badge (roadmap hint; not selectable).
- Footer: primary **"Create Webhook"**, secondary **"Cancel"**. Validation: URL required + https; at least one event required (disable Create until both are satisfied).

**4. Secret reveal (one-time, right after creating)**
- A confirmation card showing the signing secret (monospace) with a **Copy** button and a warning: "Copy your secret now — for security you won't be able to see it again."

**5. Deliveries view (the "Deliveries" sub-toggle) + per-webhook detail**
- This is the same **deliveries table** in two contexts: globally under the **Deliveries** sub-toggle (all webhooks), and inside a single webhook's detail (pre-filtered to that webhook).
- **Per-webhook detail header** (when opened from Endpoints): the webhook Name + URL, an Active/Paused toggle, a **"Send test event"** button, Edit/Delete, and the subscribed events as tags.
- A **deliveries table**: columns **Event**, **Status** (2xx = green "Success", else red "Failed"), **Attempt** (e.g. "1 of 7"), **Response code**, **Duration**, **Timestamp**. Each row expands to show the JSON **payload sent** and the **response received**. A **"Resend"** action per row.
- Filters: by **webhook** (only in the global Deliveries view), **event type**, and **status**; an "Export CSV" button (mirror the Request Logs tab).
- Empty state: "No deliveries yet. Events will appear here once they're sent."

**6. Send test event (small modal)**
- Pick an event type (default `post.published`) → **"Send test"** → a sample delivery appears at the top of the deliveries table.

**Sample payload to render inside the payload preview (so it looks real):**
```json
{
  "id": "evt_9f8c2a1b4d",
  "event": "post.published",
  "timestamp": "2026-06-27T12:00:03Z",
  "workspace_id": "601b773d2149273f48039ec2",
  "post": {
    "id": "p_88231",
    "status": "published",
    "scheduledFor": "2026-06-27T12:00:00Z",
    "publishedAt": "2026-06-27T12:00:03Z",
    "content": "Big news! Our summer sale starts today 🎉",
    "platforms": [
      { "platform": "facebook",  "status": "published", "platformPostId": "123_456", "publishedUrl": "https://facebook.com/123/posts/456", "error": null },
      { "platform": "instagram", "status": "published", "platformPostId": "178900", "publishedUrl": "https://instagram.com/p/abc123/",   "error": null }
    ]
  }
}
```

**Do not** include: dark-mode toggle, RTL, per-delivery pricing/credits (webhooks are free), or events outside the Posts group (those are "coming soon").
