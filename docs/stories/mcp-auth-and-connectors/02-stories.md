# Epic + stories — MCP authentication and connector submission

**Scope of this doc:** the epic and **3 stories**: one research, two backend. Nothing is pushed to any tracker. The PO creates this by hand.

---

## Epic: MCP authentication and ChatGPT / Claude connector submission

ContentStudio already exposes a set of actions over MCP, so an AI assistant can list a user's workspaces and connected accounts, read their posts, and create or delete a post. What it does not have is a way for a user to grant that access safely.

Today the user's API token is passed to each MCP tool as an argument. That means the credential travels through the assistant's conversation as tool input, ends up in logs and history, and stays valid indefinitely. It also means there is no consent step: nothing shows the user what they are about to give access to, and nothing lets them take it back. And because the token is handled by the tools themselves rather than by the same authorization the rest of the product uses, there is no guarantee an assistant is limited to what that user can actually do in ContentStudio.

This is also what stops us from being listed. Both ChatGPT and Claude let users connect third-party services, and both expect the service to run a proper authorization flow, where the user signs in and approves access in their browser and can revoke it later. A server that expects an API key pasted into a tool argument cannot be listed in either directory, however good the tools are.

This epic fixes the authentication model for the MCP server and then gets ContentStudio listed as a connector in ChatGPT and in Claude. When it is done, a user connects ContentStudio from inside their assistant, signs in, sees exactly what they are approving, and can revoke it at any time, and everything the assistant does on their behalf is limited by the same permissions they have in the app.

**Related epic, do not duplicate:** the AI surfaces architecture epic owns the question of which MCP server is the one we keep and where tool definitions live. This epic takes that answer as an input and covers authentication and the directory submissions only.

### Out of scope

- Adding new MCP tools or changing what the existing tools do. This epic is about how a caller is authorized, not what they can call.
- Deciding which MCP server survives, or unifying tool definitions across MCP, the CLI and AI chat. That is the AI surfaces architecture epic.
- Removing API key access for existing integrations and the CLI. Those keep working.

### Stories

1. `[Research] Define the MCP authorization model and confirm what the ChatGPT and Claude directories require`
2. `[BE] Let users authorize an AI assistant to act on their ContentStudio account`
3. `[BE] Prepare and submit the ContentStudio connector to the ChatGPT and Claude directories`

### Sequencing

Research first, with a review gate. Then the authorization work. Submission only after the authorization flow is live and verified, because both directories review a working connector.

---

## [Research] Define the MCP authorization model and confirm what the ChatGPT and Claude directories require

### Description

As the engineering and product team, we want the current requirements for listing a connector in ChatGPT and in Claude confirmed from their published documentation, together with a decided authorization model for our MCP server, so that we build the flow once, satisfy both directories, and do not discover a blocking requirement after the work is done.

### Workflow

1. The team documents how MCP access is authenticated today and states plainly why it cannot be listed as a connector, so the problem being solved is on the record.
2. The team reads the current published requirements for connecting a third-party service to ChatGPT and to Claude, and writes down the technical requirements, the review process, and any publisher verification or policy requirements for each.
3. The team notes where the two differ, and defines a single approach that satisfies both.
4. The team decides how workspaces are handled, given a ContentStudio user usually has several: whether the user picks one while connecting, or the assistant chooses per action.
5. The team decides what an assistant is allowed to do on the user's behalf, in particular whether it may publish, and what the user is shown and asked to approve.
6. The team decides how an assistant's access is limited to what that user can already do in ContentStudio, including their role in each workspace.
7. The team decides how a user reviews and revokes access after granting it, and where in the product that lives.
8. The team confirms existing API key access for integrations and the CLI continues to work unchanged, and for how long both are supported side by side.
9. The team plans the submission: what each directory asks for, who owns the submission, and the realistic timeline including review.
10. The output is reviewed and approved. Implementation stories are refined from it.

### Acceptance criteria

- [ ] The document states how MCP access is authenticated today and why it cannot be listed as a connector in either directory.
- [ ] The current published requirements for listing a connector in ChatGPT are documented, including technical requirements, the submission and review process, and any publisher verification or policy requirements.
- [ ] The same is documented for Claude.
- [ ] Differences between the two are listed, with a single approach that satisfies both.
- [ ] The requirements are confirmed against each provider's current published documentation at the time of the ticket, with the date of confirmation recorded, because these requirements change.
- [ ] The document defines how workspaces are handled during and after the connect flow, and what happens for a user who belongs to several.
- [ ] The document defines exactly what an assistant may do on a user's behalf, distinguishing read actions from actions that create, publish or delete, and states what the user approves at connect time.
- [ ] The document defines how an assistant's access is constrained to the user's own permissions, including their role per workspace, so it can never exceed what that user can do in the app.
- [ ] The document defines how a user sees which assistants have access and how they revoke it, and where that lives in the product.
- [ ] The document confirms that existing API key access for integrations and the CLI continues to work, and states how long both authentication paths are supported together.
- [ ] The document states what happens to an assistant's access when the user's password changes, when they sign out everywhere, when they lose access to a workspace, or when their subscription lapses.
- [ ] The document states how actions taken by an assistant are attributed and recorded, so a workspace owner can tell what was done by an assistant rather than by a person.
- [ ] The document states whether actions taken through a connector consume credits and how that is surfaced to the user.
- [ ] The document states the position for white-label customers, either that they are covered or that they are explicitly excluded, with the reasoning.
- [ ] A submission plan lists what each directory requires from us, who owns it, and the expected timeline including review.
- [ ] The document states which MCP server the connector will point at, taking that decision from the AI surfaces architecture epic rather than deciding it here.
- [ ] Open questions and recommendations are captured for the review gate.
- [ ] The document is reviewed and approved before implementation begins.

### Mock-ups

None required for the research ticket. The connect and consent screens are specified as part of it and designed in the implementation story.

### Impact on existing data

None. Nothing is changed, migrated or deleted by this ticket.

### Impact on other products

The research must confirm the position for the CLI, the npm MCP package, the in-app AI chat and existing API integrations, since all of them authenticate against ContentStudio today and must not be broken by the chosen model.

### Dependencies

- Takes the "which MCP server do we keep" decision from the AI surfaces architecture epic.
- Overlaps the service authorization epic on the question of constraining a caller to the user's real permissions, and should reuse its conclusion rather than invent a second model.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, research ticket with no UI
- [ ] Multilingual support (the connect and consent screens will need translated copy, so the research must note it)
- [ ] UI theming support — N/A, research ticket with no UI
- [ ] White-label domains impact review (the document must state the white-label position explicitly)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Let users authorize an AI assistant to act on their ContentStudio account

### Description

As a ContentStudio user, I want to connect ContentStudio to my AI assistant by signing in and approving access, so that I never have to paste an API key into a chat, I can see exactly what I am granting, and I can revoke it whenever I want.

### Workflow

1. The user starts connecting ContentStudio from inside their AI assistant.
2. The user is taken to a ContentStudio sign-in page in their browser. If already signed in, they go straight to the approval step.
3. The user is shown what they are about to allow: the name of the assistant requesting access, what it will be able to do, and which workspace or workspaces it covers.
4. The user approves or cancels. On cancel, nothing is granted and the assistant reports that the connection was not completed.
5. On approval, the assistant is connected and can act on the user's behalf, limited to what that user can already do in ContentStudio.
6. The user can later see the connected assistant in their ContentStudio settings and revoke its access.
7. After revoking, the assistant can no longer act on the user's behalf, and the next attempt tells the user access was removed and offers to reconnect.

### Acceptance criteria

- [ ] A user can connect ContentStudio from an AI assistant without ever entering an API key or token into the assistant.
- [ ] Connecting takes the user through ContentStudio sign-in in a browser, and a user who is already signed in is not asked to sign in again.
- [ ] Before access is granted, the user sees an approval screen naming the requesting assistant, what it will be able to do, and the workspace scope. Approval is an explicit action, not implied by signing in.
- [ ] Cancelling grants nothing, and the assistant reports that the connection was not completed.
- [ ] No credential is ever passed to or returned through a tool argument, and no long-lived credential is held by the assistant.
- [ ] An assistant acting on a user's behalf can do no more than that user can do in the app: workspace access and the user's role in each workspace are both enforced on every action.
- [ ] An assistant cannot act on a workspace outside the granted scope, and attempting to returns a clear refusal rather than data.
- [ ] The user can see every connected assistant in their ContentStudio settings, with when it was connected and when it was last used.
- [ ] The user can revoke an assistant's access from that screen, and it stops working immediately.
- [ ] Access is invalidated when the user loses access to the workspace it was granted for, and when the user signs out of all devices or resets their password.
- [ ] Existing API key access continues to work unchanged for integrations, the CLI and the published MCP package, verified explicitly.
- [ ] Actions taken by an assistant are recorded and attributable, so a workspace owner can tell an assistant's action from a person's.
- [ ] The authorization flow matches the model agreed in `[Research] Define the MCP authorization model and confirm what the ChatGPT and Claude directories require`, and meets the requirements both directories publish.
- [ ] The flow is verified end to end against at least one real AI assistant client, not only against a test harness.
- [ ] All user-facing copy for the sign-in, approval, revocation and error states is defined and translated, and reads clearly to a non-technical user.
- [ ] Failure states are handled with clear messages: the approval times out, the user has no workspaces, the user's subscription does not allow it, or the connection is refused.

### Mock-ups

Required: the approval screen shown during connect, the connected assistants list in settings, the revoke confirmation, and the error states. To be produced under `[Design] Design the assistant authorization and connected assistants screens`, which the PO should create alongside this story once the research is approved.

### Impact on existing data

Introduces new records for granted access and the assistants a user has connected. No existing data is migrated or deleted. Existing API keys keep working and are not converted.

### Impact on other products

The connected assistants screen is a web addition. The mobile app and the Chrome extension are unaffected, and the research must have confirmed that. Existing integrations, the CLI and the published MCP package must keep authenticating exactly as they do today.

### Dependencies

- Depends on `[Research] Define the MCP authorization model and confirm what the ChatGPT and Claude directories require`.
- Depends on the AI surfaces architecture epic for which MCP server the flow is implemented on.
- Needs `[Design] Design the assistant authorization and connected assistants screens` before the frontend part of the flow is built.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (the sign-in and approval screens open in a browser and may be on a phone)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review (the approval screen must present the right brand for the domain the user signs in on)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Prepare and submit the ContentStudio connector to the ChatGPT and Claude directories

### Description

As ContentStudio, we want our connector listed in the ChatGPT and Claude directories, so that users can find and connect ContentStudio from inside the assistant they already use instead of setting it up manually.

### Workflow

1. The team assembles what each directory asks for: the connector's name, description, icon, category, the list of what it can do, support and privacy links, and any publisher verification.
2. The team checks the connector against each directory's technical and policy requirements and fixes anything that fails.
3. The team tests the connector as an end user would in each assistant: connect, approve, run each available action, revoke, reconnect.
4. The team submits to each directory and tracks the review.
5. The team responds to review feedback until both listings are live.
6. Once live, the team documents how the listing is updated when tools change, and who owns it.

### Acceptance criteria

- [ ] The connector meets every current technical requirement published by each directory, verified item by item against their documentation with the date of verification recorded.
- [ ] The listing content is prepared and approved: name, description, icon, category, the list of what the connector can do, and support and privacy links.
- [ ] Publisher verification or domain ownership requirements are satisfied for each directory.
- [ ] The full user journey is tested inside each assistant before submission: discover, connect, approve, use every available action, revoke, reconnect.
- [ ] Every action exposed through the connector is verified to work from each assistant, and any action deliberately not exposed is documented with the reason.
- [ ] Actions that create, publish or delete behave as agreed in research, including any confirmation the user is expected to see first.
- [ ] The connector behaves correctly for a user with several workspaces, for a user with a single workspace, and for a user with none.
- [ ] The connector behaves correctly for a user whose plan does not include the feature the assistant asked for, returning a clear message rather than an error.
- [ ] Both submissions are made, review feedback is addressed, and both listings are live.
- [ ] A documented owner and process exists for keeping each listing current when the connector's actions change.
- [ ] Monitoring is in place for connector traffic and failures, so a broken listing is noticed by the team rather than reported by users.

### Mock-ups

None beyond the listing assets, which are prepared as part of this story.

### Impact on existing data

None.

### Impact on other products

A public listing raises expectations of the underlying actions. The team should expect connector traffic against the same features the web app uses, and the monitoring criterion above exists for that reason.

### Dependencies

- Depends on `[BE] Let users authorize an AI assistant to act on their ContentStudio account`. Neither directory will review a connector without a working authorization flow.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no new UI in this story
- [ ] Multilingual support (listing content is in English, but the connect flow it links to must be translated)
- [ ] UI theming support — N/A, no new UI in this story
- [ ] White-label domains impact review (the listing represents ContentStudio, so the white-label position agreed in research must be honoured)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
