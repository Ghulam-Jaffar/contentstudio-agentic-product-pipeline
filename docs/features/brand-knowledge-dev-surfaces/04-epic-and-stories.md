# Epic + stories — Brand knowledge on the public API

**Priority: P1.** 4 stories. Nothing is pushed to any tracker.

---

## Epic: Brand knowledge on the public API

Brand Knowledge is the store of everything that makes AI output sound and look like a customer's brand: their brand profile, their brand style, their brand voice, the source materials the brand was built from, and their brand assets. It is the setting that decides whether generated content is usable or generic, and it works in exactly one place, the browser.

The public API cannot read it and cannot change it. So an agency onboarding fifty client workspaces fills five tabs by hand fifty times. A customer whose brand guidelines already live in another system retypes them, and retypes them again whenever they change. A developer who drives the rest of ContentStudio from the API hits a wall at the one feature that determines content quality. And an AI assistant connected over MCP can create posts for a workspace without being able to find out what that workspace sounds like, so it writes generic copy or asks the user to paste the brand in every session.

This epic exposes the whole feature. Everything the Brand Knowledge editor can do becomes an API capability, and from there a CLI command, an MCP tool, a connector action and a no-code step. A brand toolkit in the AI agents service follows, so the in-product assistant can read the brand and change it too.

It is worth noting what this buys beyond closing a gap. Research across the category found that no social media management platform exposes brand voice or brand kit over any developer surface. Buffer, Hootsuite, Sprout, SocialBee and Publer all have brand-voice features and all of them keep that layer inside the dashboard. The vendors doing it well, Jasper, Frontify and Canva, are not competing for the same customer.

### Out of scope

- Changing how brand knowledge behaves. This is exposure, not redesign. The Brand Knowledge editor and its five tabs are the revamp's epic, not this one.
- Any frontend or mobile work. Nothing in the web app or the Flutter app changes.
- Multiple brands per workspace. One brand per workspace is settled by the revamp.
- Accepting brand content inside a generation request. Generation keeps resolving the brand server-side from the stored record, so no caller can steer output with content the workspace never approved.
- Webhooks on brand change, and a job and polling model for long rebuilds. Both deferred.

### Stories

1. `[BE] Add brand knowledge read and editing to the public API`
2. `[BE] Add brand source materials and brand assets to the public API`
3. `[BE] Expose brand knowledge on every developer surface`
4. `[BE] Add brand knowledge tools to the AI agents service`

---

## [BE] Add brand knowledge read and editing to the public API

### Description

As a developer or an agency operator, I want to read and edit a workspace's brand knowledge over the API, so that I can set up and maintain a brand from a script instead of filling five tabs in the browser for every client.

The brand has three editable parts that a person fills in: the brand profile, which holds who the business is and how it is positioned; the brand style, which holds the logo, colours and fonts; and the brand voice, which holds tone, audience and character. A caller needs to read all of it at once, and to change one part without disturbing the others, because two integrations that each own a different part must not overwrite each other.

The workspace can also switch its brand on or off for AI generation, set its post generation defaults, and delete the brand entirely. All of that is reachable in the browser today and none of it is reachable over the API.

### Workflow

1. The developer uses the API key they already have for posts, media and analytics.
2. They read the workspace's brand and receive all of it: the profile, the style, the voice, and a summary of the source materials and brand assets.
3. They read one part on its own when they only need that part.
4. They update the brand profile, the brand style or the brand voice. Only the fields they send change, and a write to one part never alters another.
5. They turn the brand on or off for AI generation.
6. They read and update the post generation defaults.
7. They delete the brand when a client leaves.
8. Every change appears immediately in the Brand Knowledge editor, because it is the same brand.

### Acceptance criteria

- [ ] A caller can read a workspace's whole brand in one request, returning the brand profile, the brand style, the brand voice, the list of source materials and the list of brand assets.
- [ ] A caller can read the brand profile, the brand style and the brand voice individually.
- [ ] A caller can update the brand profile, the brand style and the brand voice individually, and a write to one leaves the other two unchanged.
- [ ] A partial update changes only the fields supplied and leaves every other field as it was.
- [ ] A caller can read and set whether the brand is applied to AI generation.
- [ ] A caller can read and update the post generation defaults.
- [ ] A caller can delete the brand, and the response states what was removed.
- [ ] Reading a workspace that has no brand returns an empty brand marked as not set up, not an error.
- [ ] Writing to a workspace that has no brand creates one.
- [ ] Every response states when the brand was last changed.
- [ ] Tone, emotion, character and language accept free text, matching the product, and the documentation lists the values the product commonly uses.
- [ ] Every documented limit is enforced and returns a typed error naming the limit and the current value, including the maximum number of brand colours and the maximum length of a single tag or list entry.
- [ ] Authorization matches the dashboard: a role that cannot change brand settings in the browser cannot change them over the API, and the refusal names the role required.
- [ ] Requests follow the existing public API conventions for workspace scoping, the response envelope, error codes and API credit consumption.
- [ ] The existing brand status endpoint continues to behave exactly as it does today, and an integration relying on it is unaffected.
- [ ] Generation continues to resolve the brand server-side from the stored record. No endpoint in this story accepts brand content that is applied directly to a generation request.
- [ ] The public schema for the brand is versioned independently of the internal storage document, so a later change to storage does not break integrations.
- [ ] OpenAPI annotations and the public documentation cover every endpoint, with a worked example of setting up a brand from scratch.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

No schema change and no migration. The endpoints read and write the brand record that already exists. Brand records created or edited over the API are identical to ones created in the browser.

### Impact on other products

The Brand Knowledge editor shows any change made over the API immediately, because it is the same record. AI generation in the composer, the inbox, RSS and Evergreen is unaffected, because it keeps resolving the brand server-side. The existing brand status endpoint is unchanged.

### Dependencies

- Depends on the Brand Knowledge revamp's data model if the decision is to ship against the revamped five-part shape rather than the current one. Confirm before starting.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (API error messages are customer-facing and follow the existing localization of the public API)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review (white-label customers use the same API)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add brand source materials and brand assets to the public API

### Description

As a developer or an agency operator, I want to manage a workspace's source materials and brand assets over the API, so that I can build a brand from a website or a document rather than typing it out, and keep logos and imagery in step with the brand system we already maintain elsewhere.

Source materials are what the brand is derived from: a website address, an uploaded document, pasted text, or a connected social account. Adding one rebuilds the brand from every active source, which is the single highest-value thing a caller can automate, because it turns brand setup from a form-filling exercise into pointing at a website.

Brand assets are the logos and imagery the brand uses. They are stored as media, and today they are hidden from the normal media list, so a caller has no way to reach them at all.

Two of the product's current behaviours are wrong at the edges and this story fixes them rather than copying them: reaching the source material limit reports that the profile could not be found, and adding more website sources than the system will process drops the extras without saying so.

### Workflow

1. The developer lists the workspace's source materials and sees each one's type, name and current state.
2. They add a source: a website address, a document, pasted text, or a connected social account.
3. The brand is rebuilt from every active source, the same as it is when done in the browser, and the response returns the updated brand so the caller can see what changed.
4. If a source cannot be read, it is marked unreachable with a reason and the brand is left exactly as it was.
5. They rename, update or remove a source.
6. They rebuild the brand from the existing sources on demand, choosing between the rebuild that re-derives the brand and the one that blends new material into it without discarding what is there.
7. They set whether a document is used for inbox auto-replies, and are told that doing so takes time to take effect.
8. They list the workspace's brand assets, upload a new one as a file or by giving a web address to fetch, and remove ones they no longer want.

### Acceptance criteria

- [ ] A caller can list a workspace's source materials, each showing its type, name, current state and when it was last synced.
- [ ] A caller can add a source material of each type the product supports: website address, document, pasted text and connected social account.
- [ ] Adding a source rebuilds the brand from all active sources, matching the browser behaviour exactly, and the response returns the updated brand.
- [ ] A caller can read, rename, update and delete an individual source material.
- [ ] A caller can trigger a rebuild from existing sources, and can choose between the rebuild that re-derives the brand from scratch and the one that blends without discarding existing content.
- [ ] When a source cannot be read, it is marked unreachable with a reason, and the brand is unchanged.
- [ ] A source that failed can be retried without being removed and re-added.
- [ ] Reaching the source material limit returns a typed error naming the limit and the current count. It no longer reports that the profile could not be found.
- [ ] When more website sources are supplied than the system will process, the response names which were used and which were skipped, and why. They are no longer dropped silently.
- [ ] A caller can read and set the per-source auto-reply flag, and the documentation states that enabling it starts indexing and takes time to take effect.
- [ ] A caller can list the workspace's brand assets.
- [ ] A caller can upload a brand asset either as a file or by supplying a web address for ContentStudio to fetch, matching how media upload already works.
- [ ] A caller can upload a brand logo.
- [ ] A caller can read and delete an individual brand asset.
- [ ] Upload size limits and accepted formats are enforced and documented, and a rejected upload returns a typed error naming the limit or the accepted formats.
- [ ] Reaching the brand asset limit returns a typed error naming the limit and the current count.
- [ ] Brand asset uploads consume the workspace's media storage allowance, and a workspace that is out of storage is refused with the existing storage-full error.
- [ ] The documentation states plainly that adding a website source can take several minutes, so a caller sets its timeouts accordingly.
- [ ] Authorization matches the dashboard, and the refusal names the role required.
- [ ] Requests follow the existing public API conventions for workspace scoping, the response envelope, error codes and API credit consumption.
- [ ] OpenAPI annotations and the public documentation cover every endpoint, including a worked example of building a brand from a website address.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

No schema change and no migration. Source materials and brand assets are read and written in the form they already take. Fixing the misleading not-found response and the silent skipping of extra website sources changes what callers are told, not what is stored.

### Impact on other products

Brand assets are media, so uploads count against the workspace's storage allowance and appear wherever brand assets already appear. Enabling the auto-reply flag on a document feeds the inbox auto-reply feature, which is existing behaviour now reachable from a second place. The Brand Knowledge editor reflects every change immediately.

### Dependencies

- Depends on `[BE] Add brand knowledge read and editing to the public API` for the brand resource these hang from.
- The open question of whether brand assets stay marked on the media record or move into a Media Library folder sits with the Brand Knowledge revamp. The API exposes brand assets as their own list either way, so this story is not blocked, but confirm before building.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (API error messages are customer-facing and follow the existing localization of the public API)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review (white-label customers use the same API)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose brand knowledge on every developer surface

### Description

As a developer using ContentStudio through the CLI, an AI assistant or a no-code tool, I want brand knowledge available there too, so that I am not pushed onto the REST API for the one part of a workflow I otherwise drive from a single place.

This matters more for brand knowledge than for most capabilities, because the surface that benefits most is the assistant. An AI assistant that can create a post for a workspace but cannot find out what that workspace sounds like will write generic copy, and the user will blame the product rather than the assistant. Every vendor that has solved this reached for the agent surface first.

### Workflow

1. The developer manages brand knowledge from the CLI, including building a brand from a website address in one command.
2. An AI assistant, through the MCP server, can read a workspace's brand before it writes anything, and can change the brand on the user's behalf.
3. The developer builds a no-code automation in Zapier, Make or n8n that keeps ContentStudio's brand in step with the system where their brand guidelines actually live.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Brand knowledge management is available in the CLI, covering reading the brand, updating the profile, style and voice, enabling and disabling the brand, managing source materials, rebuilding from sources, and managing brand assets.
- [ ] The CLI shows progress while a rebuild from a website source is running, rather than appearing to hang.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes the same capabilities as tools, so an assistant can read a workspace's brand and change it on a user's behalf.
- [ ] MCP brand tools return a compact summary by default and the full detail of one part on request, so a large brand does not exhaust the assistant's context.
- [ ] MCP source material reads return short extracts rather than whole documents.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Zapier, Make and n8n expose the capabilities appropriate to each: at minimum reading the brand, updating the profile, style and voice, and adding a source material.
- [ ] Where a surface deliberately omits a capability, for example deleting the brand from a no-code tool, that omission is recorded with its reason rather than left as a silent gap.
- [ ] The public documentation covers the capability on every surface, with a worked example of building a brand from a website end to end.
- [ ] The parity check passes for brand knowledge across every surface.
- [ ] Every surface enforces the same authorization and plan gating as the REST endpoints. No surface is a weaker path to the same action.
- [ ] Customer-facing text on each surface, including command help, tool descriptions and error messages, is written for a developer reader and translated where the surface supports it.

### Mock-ups

None. No graphical UI in this story. Command help and tool descriptions are specified in the acceptance criteria.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on both endpoint stories in this epic.
- Depends on the developer surface parity contract epic. Without it this story is six separate hand-copied changes rather than one.
- Takes from the MCP authorization epic how the ChatGPT and Claude connectors authenticate.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review (white-label customers use the same surfaces)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add brand knowledge tools to the AI agents service

### Description

As a customer using ContentStudio's own AI assistant, I want it to already know my brand and to be able to change it when I ask, so that I stop explaining my brand voice in every conversation and stop leaving the chat to edit a setting.

Today the assistant receives the brand only as an opaque block of guidance pushed in behind the scenes, and only for the members that write content. It cannot be asked what the brand is, it cannot report what it used, and it cannot change anything. A user who asks what tone their workspace is set to gets no useful answer.

Brand changes made this way are destructive, because rebuilding from sources regenerates fields a person may have written by hand. So every brand write goes through the same confirmation step the assistant already uses for creating posts and managing labels: the user sees exactly what will change before anything is saved.

### Workflow

1. The user asks the assistant about their brand, for example what tone the workspace uses or what the brand colours are.
2. The assistant reads the brand and answers, stating what it drew on so the user can tell the answer is grounded rather than invented.
3. The user asks the assistant to change something in the brand.
4. The assistant shows what it is about to change, with the current value and the proposed value, and waits.
5. The user approves, and the change is saved. Or the user declines, and nothing changes.
6. The next piece of content generated in that workspace reflects the change, because generation reads the stored brand.

### Acceptance criteria

- [ ] The assistant can read a workspace's brand profile, brand style and brand voice, and answer questions about them.
- [ ] The assistant can search the workspace's source materials and answer from them, returning short extracts rather than whole documents.
- [ ] The assistant can list the workspace's brand assets.
- [ ] Brand reads return a compact summary by default and the full detail of one part on request, so a large brand does not exhaust the assistant's context.
- [ ] When the assistant answers using brand knowledge, it states what it used.
- [ ] When a workspace has no brand, the assistant says so plainly and points the user to where they set one up, rather than inventing an answer.
- [ ] When a workspace's brand is switched off for AI generation, the assistant says so, so the user is not told content is on-brand when the product will not apply it.
- [ ] The assistant can update the brand profile, the brand style and the brand voice, and can add and remove source materials.
- [ ] Every brand write is presented for confirmation before it is applied, using the same confirmation mechanism the assistant already uses for other writes.
- [ ] The confirmation names exactly what will change, showing the current value and the proposed value.
- [ ] A confirmation for rebuilding the brand from sources additionally states that derived fields will be regenerated from all active sources.
- [ ] A confirmation that the user declines saves nothing.
- [ ] A confirmation that is no longer valid, because the brand changed after it was shown, is refused rather than applied to a value the user did not see.
- [ ] Brand tools act only on the workspace the user is in, using the user's own credentials, with no elevated access.
- [ ] Authorization matches the dashboard: a user whose role cannot change brand settings in the browser cannot change them through the assistant.
- [ ] Brand tool responses follow the service's existing response contract, including its size limits and its truncation behaviour, so a large brand degrades gracefully rather than failing.
- [ ] Rebuilding a brand from a website source is handled so that a slow rebuild does not appear to the user as a failed request.
- [ ] The assistant behaves identically on web and in the mobile app, because both render the same conversation components. No change is needed in either client.

### Mock-ups

None. This story adds no new conversation component. Brand answers and confirmations use the components the assistant already renders on web and mobile.

### Impact on existing data

None. Brand records are read and written through the public API in the same form as any other caller.

### Impact on other products

The in-product assistant gains brand awareness on both web and mobile without any client change, because both already render the same conversation components. The brand guidance already pushed into content generation is unchanged. A brand change made through the assistant appears in the Brand Knowledge editor immediately.

### Dependencies

- Depends on both endpoint stories in this epic. The service reaches brand knowledge through the public API, so the endpoints must exist first.
- Depends on the existing confirmation mechanism the service already uses for write tools.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — the assistant renders on mobile using existing components, so verify brand answers and confirmations display correctly there
- [ ] Multilingual support (assistant answers and confirmations follow the user's language, and brand content is returned in the language it was written in)
- [ ] UI theming support — N/A, no new interface elements are introduced
- [ ] White-label domains impact review (white-label customers use the same assistant)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
