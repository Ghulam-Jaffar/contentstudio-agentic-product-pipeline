# Epics & Stories — Publishing API v1.9 through v2.2

---

## Epic 1: Publishing API v1.9 — Add Missing Filters & Statuses to Fetch Posts

Extend the `GET /posts` endpoint to support additional filters and statuses that are available in the ContentStudio web app but currently missing from the Publishing API. This brings the API's filtering capabilities closer to parity with the web planner, enabling API consumers and integration apps (Zapier, Make, n8n) to retrieve posts with the same granularity as the UI.

**New statuses to add:** `notification_sent`, `notification_declined`, `processing`

**New filters to add:** `labels[]`, `campaigns[]`, `content_category[]`, `created_by[]`, `comment_status` (resolved/unresolved)

### Stories:

**Story 1.1: [BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**

> **Description:**
> As an API consumer, I want to filter posts by labels, campaigns, content categories, creator, and comment resolution status — and also filter by additional post statuses (notification sent, notification declined, processing) — so that I can retrieve exactly the posts I need without manual filtering on my end.
>
> The `GET /api/v1/workspaces/{workspace_id}/posts` endpoint currently supports filtering by `status[]`, `date_from`, `date_to`, `approval_assigned_to[]`, `approval_requested_by[]`. This story adds the following:
>
> **New filter parameters:**
> - `labels[]` — filter posts by label IDs
> - `campaigns[]` — filter posts by campaign IDs
> - `content_category[]` — filter posts by content category IDs
> - `created_by[]` — filter posts by the user IDs who created them
> - `comment_status` — filter posts by comment resolution status: `resolved`, `unresolved`, or `all` (default)
>
> **New status values for existing `status[]` filter:**
> - `notification_sent` — posts where a push notification was sent to the user for manual publishing
> - `notification_declined` — posts where the push notification was declined
> - `processing` — posts currently being processed for publishing
>
> Update the API documentation (Swagger/OpenAPI) to reflect all new parameters and status values.
>
> ---
>
> ### Workflow:
>
> 1. API consumer sends `GET /api/v1/workspaces/{workspace_id}/posts?labels[]=abc123&status[]=notification_sent`
> 2. The API validates the filter parameters and returns only posts matching the specified labels and status
> 3. API consumer can combine any number of the new filters with existing ones (date range, approval filters, pagination)
> 4. If an invalid filter value is provided, the API returns a clear validation error
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] `labels[]` filter returns only posts tagged with the specified label IDs
> - [ ] `campaigns[]` filter returns only posts belonging to the specified campaigns
> - [ ] `content_category[]` filter returns only posts in the specified content categories
> - [ ] `created_by[]` filter returns only posts created by the specified user IDs
> - [ ] `comment_status=resolved` returns only posts where all comments are resolved
> - [ ] `comment_status=unresolved` returns only posts with unresolved comments
> - [ ] `status[]=notification_sent` returns posts with notification sent status
> - [ ] `status[]=notification_declined` returns posts with notification declined status
> - [ ] `status[]=processing` returns posts currently being processed
> - [ ] All new filters can be combined with existing filters (date range, approval, pagination)
> - [ ] Invalid filter values return appropriate validation error messages
> - [ ] API documentation (Swagger/OpenAPI) is updated with all new parameters and status values
>
> ---
>
> ### Mock-ups:
> N/A — backend only
>
> ---
>
> ### Impact on existing data:
> No changes to existing data. These filters query existing fields that are already stored on posts.
>
> ---
>
> ### Impact on other products:
> Zapier, Make.com, and n8n integration apps will be updated separately to expose these new filters.
>
> ---
>
> ### Dependencies:
> None.
>
> ---
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A, API returns data, no user-facing strings
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.2: [BE] Update Zapier app to support new fetch posts filters**

> **Description:**
> As a Zapier user, I want to use the new fetch posts filters (labels, campaigns, content categories, created by, comment status, and new statuses) when searching for posts in my Zaps, so that I can build more precise automations.
>
> Update the Zapier app's "Find Posts" / "List Posts" action to expose the new filter parameters added in **[BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**. Each new filter should appear as an optional input field in the Zapier action configuration.
>
> ---
>
> ### Workflow:
> 1. User configures a "Find Posts" action in a Zap
> 2. User sees new optional filter fields: Labels, Campaigns, Content Category, Created By, Comment Status
> 3. User sees the new status options (Notification Sent, Notification Declined, Processing) in the Status dropdown
> 4. User selects desired filters and the Zap fetches matching posts
>
> ---
>
> ### Acceptance criteria:
> - [ ] Labels filter is available as a multi-select field in the Zapier action
> - [ ] Campaigns filter is available as a multi-select field
> - [ ] Content Category filter is available as a multi-select field
> - [ ] Created By filter is available as a multi-select field
> - [ ] Comment Status filter is available as a dropdown (All, Resolved, Unresolved)
> - [ ] New statuses (Notification Sent, Notification Declined, Processing) appear in the Status filter options
> - [ ] All filters work correctly when used in combination
>
> ---
>
> ### Mock-ups:
> N/A — backend only
>
> ### Impact on existing data:
> None.
>
> ### Impact on other products:
> None — Zapier app only.
>
> ### Dependencies:
> Depends on: **[BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.3: [BE] Update Make.com app to support new fetch posts filters**

> **Description:**
> As a Make.com user, I want to use the new fetch posts filters (labels, campaigns, content categories, created by, comment status, and new statuses) when listing posts in my scenarios, so that I can build more targeted automations.
>
> Update the Make.com app's "List Posts" module to expose the new filter parameters added in **[BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**.
>
> ---
>
> ### Workflow:
> 1. User adds a "List Posts" module in a Make.com scenario
> 2. User sees new optional filter fields: Labels, Campaigns, Content Category, Created By, Comment Status
> 3. User sees the new status options in the Status filter
> 4. User configures filters and the module fetches matching posts
>
> ---
>
> ### Acceptance criteria:
> - [ ] All new filters (Labels, Campaigns, Content Category, Created By, Comment Status) are available in the module configuration
> - [ ] New statuses (Notification Sent, Notification Declined, Processing) appear in the Status filter
> - [ ] All filters work correctly in combination
>
> ---
>
> ### Mock-ups:
> N/A
>
> ### Impact on existing data:
> None.
>
> ### Impact on other products:
> None — Make.com app only.
>
> ### Dependencies:
> Depends on: **[BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.4: [BE] Update n8n node to support new fetch posts filters**

> **Description:**
> As an n8n user, I want to use the new fetch posts filters when listing posts in my workflows.
>
> Update the n8n node's "List Posts" operation to expose the new filter parameters added in **[BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**.
>
> ---
>
> ### Workflow:
> 1. User adds a ContentStudio node in n8n with "List Posts" operation
> 2. User sees new optional filter fields and new status options
> 3. User configures filters and the node fetches matching posts
>
> ---
>
> ### Acceptance criteria:
> - [ ] All new filters (Labels, Campaigns, Content Category, Created By, Comment Status) are available in the node configuration
> - [ ] New statuses (Notification Sent, Notification Declined, Processing) appear in the Status filter
> - [ ] All filters work correctly in combination
>
> ---
>
> ### Mock-ups:
> N/A
>
> ### Impact on existing data:
> None.
>
> ### Impact on other products:
> None — n8n node only.
>
> ### Dependencies:
> Depends on: **[BE] Add missing filters and statuses to fetch posts endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Epic 2: Publishing API v2.0 — X (Twitter) Threads Posting

Extend the Publishing API v1's post creation endpoint to support publishing X (Twitter) threads. Currently, when creating a post via the API, users can only publish a single tweet. This epic enables API consumers to compose and publish multi-tweet threads, where each tweet in the thread is posted as a reply to the previous one. Integration apps (Zapier, Make.com, n8n) should also be updated to allow users to compose threads in their automations.

### Stories:

**Story 2.1: [BE] Add X (Twitter) threads support to post creation endpoint in Publishing API v1**

> **Description:**
> As an API consumer, I want to publish X (Twitter) threads via the Publishing API so that I can automate multi-tweet thread publishing without logging into ContentStudio.
>
> When creating a post via `POST /api/v1/workspaces/{workspace_id}/posts`, allow the user to provide an array of thread tweets. Each tweet in the thread should be published as a reply to the previous one, forming a connected thread on X.
>
> The internal platform logic for Twitter threads already exists — this story exposes that capability through the public API.
>
> ---
>
> ### Workflow:
> 1. API consumer sends a POST request to create a new post with X (Twitter) selected as a platform
> 2. The request includes thread content — an ordered array of tweet texts (and optionally media for each)
> 3. The API validates each tweet's character limit and media requirements
> 4. The post is created and scheduled/published as a thread — the first tweet goes out, followed by each subsequent tweet as a reply
> 5. The API response includes the post details with all thread tweet IDs
>
> ---
>
> ### Acceptance criteria:
> - [ ] API consumers can provide multiple tweets in a single post creation request for X (Twitter)
> - [ ] Each tweet in the thread respects X's character limit
> - [ ] Media attachments can be included per individual tweet in the thread
> - [ ] Threads are published in the correct order, with each tweet replying to the previous
> - [ ] The API response includes details for all tweets in the thread
> - [ ] Validation errors clearly indicate which tweet in the thread has an issue
> - [ ] API documentation (Swagger/OpenAPI) is updated
>
> ---
>
> ### Mock-ups:
> N/A — backend only
>
> ### Impact on existing data:
> No changes to existing data structure. Thread data is stored using the existing internal thread mechanism.
>
> ### Impact on other products:
> Zapier, Make.com, and n8n will be updated separately to support thread composition.
>
> ### Dependencies:
> None.
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 2.2: [BE] Update Zapier app to support X (Twitter) threads posting**

> **Description:**
> As a Zapier user, I want to compose and publish X (Twitter) threads from my Zaps so that I can automate thread publishing.
>
> Update the Zapier app's "Create Post" action to allow specifying multiple tweets for a thread when X (Twitter) is selected as a platform.
>
> ---
>
> ### Acceptance criteria:
> - [ ] Users can add multiple tweet texts in the Zapier action configuration for X threads
> - [ ] Media can be attached per tweet in the thread
> - [ ] Thread is published correctly via the updated API
>
> ### Dependencies:
> Depends on: **[BE] Add X (Twitter) threads support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 2.3: [BE] Update Make.com app to support X (Twitter) threads posting**

> **Description:**
> As a Make.com user, I want to compose and publish X (Twitter) threads from my scenarios.
>
> Update the Make.com app's "Create Post" module to allow specifying multiple tweets for a thread when X (Twitter) is selected.
>
> ---
>
> ### Acceptance criteria:
> - [ ] Users can add multiple tweet texts in the Make.com module for X threads
> - [ ] Media can be attached per tweet
> - [ ] Thread is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add X (Twitter) threads support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 2.4: [BE] Update n8n node to support X (Twitter) threads posting**

> **Description:**
> As an n8n user, I want to compose and publish X (Twitter) threads from my workflows.
>
> Update the n8n node's "Create Post" operation to allow specifying multiple tweets for a thread.
>
> ---
>
> ### Acceptance criteria:
> - [ ] Users can add multiple tweet texts in the n8n node for X threads
> - [ ] Media can be attached per tweet
> - [ ] Thread is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add X (Twitter) threads support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Epic 3: Publishing API v2.1 — Meta Threads Threaded Posting

Extend the Publishing API v1's post creation endpoint to support publishing threaded posts on Meta's Threads platform. Currently, the API can publish a single post to Threads, but users cannot compose multi-post threads (where each post replies to the previous, forming a conversation). This epic enables API consumers to publish connected thread sequences on Threads via the API and integration apps.

### Stories:

**Story 3.1: [BE] Add Meta Threads threaded posting support to post creation endpoint in Publishing API v1**

> **Description:**
> As an API consumer, I want to publish threaded posts on Meta's Threads platform via the Publishing API so that I can automate multi-post thread publishing.
>
> When creating a post via `POST /api/v1/workspaces/{workspace_id}/posts` with Threads selected as a platform, allow the user to provide an array of thread posts. Each post in the sequence is published as a reply to the previous one, forming a connected thread on Threads.
>
> The internal platform logic for Threads threaded posting already exists — this story exposes it through the public API.
>
> ---
>
> ### Workflow:
> 1. API consumer sends a POST request to create a new post with Threads selected as a platform
> 2. The request includes an ordered array of thread post texts (and optionally media for each)
> 3. The API validates content requirements for each post in the thread
> 4. The post is created and published as a threaded sequence on Threads
> 5. The API response includes details for all posts in the thread
>
> ---
>
> ### Acceptance criteria:
> - [ ] API consumers can provide multiple posts in a single request for Threads
> - [ ] Each post in the thread respects Threads' content limits
> - [ ] Media attachments can be included per individual post in the thread
> - [ ] Posts are published in the correct order as a connected thread
> - [ ] Validation errors clearly indicate which post in the thread has an issue
> - [ ] API documentation (Swagger/OpenAPI) is updated
>
> ---
>
> ### Mock-ups:
> N/A — backend only
>
> ### Impact on existing data:
> No changes to existing data structure.
>
> ### Impact on other products:
> Zapier, Make.com, and n8n will be updated separately.
>
> ### Dependencies:
> None.
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 3.2: [BE] Update Zapier app to support Meta Threads threaded posting**

> **Description:**
> Update the Zapier app's "Create Post" action to allow composing threaded posts when Threads is selected as a platform.
>
> ### Acceptance criteria:
> - [ ] Users can add multiple post texts in the Zapier action for Threads
> - [ ] Media can be attached per post
> - [ ] Thread is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add Meta Threads threaded posting support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 3.3: [BE] Update Make.com app to support Meta Threads threaded posting**

> **Description:**
> Update the Make.com app's "Create Post" module to allow composing threaded posts when Threads is selected.
>
> ### Acceptance criteria:
> - [ ] Users can add multiple post texts in the Make.com module for Threads
> - [ ] Media can be attached per post
> - [ ] Thread is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add Meta Threads threaded posting support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 3.4: [BE] Update n8n node to support Meta Threads threaded posting**

> **Description:**
> Update the n8n node's "Create Post" operation to allow composing threaded posts when Threads is selected.
>
> ### Acceptance criteria:
> - [ ] Users can add multiple post texts in the n8n node for Threads
> - [ ] Media can be attached per post
> - [ ] Thread is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add Meta Threads threaded posting support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Epic 4: Publishing API v2.2 — Facebook Carousel Posting

Extend the Publishing API v1's post creation endpoint to support publishing Facebook carousel posts. Currently, carousel posting via the API is limited to Instagram, TikTok, and LinkedIn. This epic enables API consumers to publish carousel posts on Facebook Pages via the API and integration apps.

### Stories:

**Story 4.1: [BE] Add Facebook carousel posting support to post creation endpoint in Publishing API v1**

> **Description:**
> As an API consumer, I want to publish Facebook carousel posts via the Publishing API so that I can automate multi-image carousel content on Facebook Pages.
>
> When creating a post via `POST /api/v1/workspaces/{workspace_id}/posts` with Facebook selected as a platform and `carousel` as the post type, the API should accept multiple images (with optional individual link and headline per card) and publish them as a carousel post on Facebook.
>
> The internal carousel logic for Facebook already exists — this story exposes it through the public API.
>
> ---
>
> ### Workflow:
> 1. API consumer sends a POST request to create a new post with Facebook selected and post type set to `carousel`
> 2. The request includes an array of carousel cards, each with an image URL (and optionally a link and headline)
> 3. The API validates the carousel content (minimum/maximum cards, image requirements)
> 4. The post is created and published as a carousel on the Facebook Page
> 5. The API response includes the post details with carousel card information
>
> ---
>
> ### Acceptance criteria:
> - [ ] API consumers can create carousel posts for Facebook Pages
> - [ ] Each carousel card supports an image, optional link, and optional headline
> - [ ] Facebook's carousel requirements are validated (min/max cards, image specs)
> - [ ] Carousel is published correctly on the Facebook Page
> - [ ] Validation errors clearly indicate which card has an issue
> - [ ] API documentation (Swagger/OpenAPI) is updated
>
> ---
>
> ### Mock-ups:
> N/A — backend only
>
> ### Impact on existing data:
> No changes to existing data structure. Carousel data uses the existing internal carousel mechanism.
>
> ### Impact on other products:
> Zapier, Make.com, and n8n will be updated separately.
>
> ### Dependencies:
> None.
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 4.2: [BE] Update Zapier app to support Facebook carousel posting**

> **Description:**
> Update the Zapier app's "Create Post" action to allow composing Facebook carousel posts with multiple image cards.
>
> ### Acceptance criteria:
> - [ ] Users can select "carousel" as post type for Facebook in the Zapier action
> - [ ] Users can add multiple image cards with optional links and headlines
> - [ ] Carousel is published correctly on Facebook
>
> ### Dependencies:
> Depends on: **[BE] Add Facebook carousel posting support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 4.3: [BE] Update Make.com app to support Facebook carousel posting**

> **Description:**
> Update the Make.com app's "Create Post" module to allow composing Facebook carousel posts.
>
> ### Acceptance criteria:
> - [ ] Users can select "carousel" as post type for Facebook in the Make.com module
> - [ ] Users can add multiple image cards with optional links and headlines
> - [ ] Carousel is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add Facebook carousel posting support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 4.4: [BE] Update n8n node to support Facebook carousel posting**

> **Description:**
> Update the n8n node's "Create Post" operation to allow composing Facebook carousel posts.
>
> ### Acceptance criteria:
> - [ ] Users can select "carousel" as post type for Facebook in the n8n node
> - [ ] Users can add multiple image cards with optional links and headlines
> - [ ] Carousel is published correctly
>
> ### Dependencies:
> Depends on: **[BE] Add Facebook carousel posting support to post creation endpoint in Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
