# Bring Your Own AI: Research

**Date:** 2026-10-02
**Trigger:** CEO request after Notion's launch on 29 Sep 2026 ([post](https://x.com/NotionHQ/status/2104994569964388738)): "Got ChatGPT Plus or Pro? That subscription now covers your AI usage in Notion. Just connect your plan, pick a GPT model, and go." Community Note: it needs a Notion Business or Enterprise plan as well.
**PO's proposed shape:** an "AI Setup" section in Settings. The user connects their own ChatGPT or OpenAI account (OpenAI sign-in), picks a GPT model, and AI chat then runs on the user's own subscription instead of ContentStudio AI credits.

Terms used here:
- **SIWC:** OpenAI's "Sign in with ChatGPT". This is the program that lets an app spend a user's ChatGPT Plus or Pro plan.
- **BYOK:** bring your own key. The user pastes a pay-as-you-go OpenAI API key into the app.

---

## PO Direction (2026-10-02, after first research pass)

Clarified scope from the PO. This overrides the earlier "API key first" recommendation below.

- **This is not "log in to ContentStudio with ChatGPT".** Users sign in to ContentStudio as they do today. Once inside, they **connect their ChatGPT plan** so our AI runs on it. This is the plan-usage half of OpenAI's program. The identity-only half is out of scope.
- **What runs on the user's plan:**
  - all of ContentStudio's text and reasoning AI, not just one screen
  - the AI chat orchestrator and coordinator
  - the general assistant
  - the agents and follow-ups behind chat
- **What stays on ContentStudio's own models and credits:**
  - caption generation
  - image generation
  - video generation
- **Model picker in AI chat:** shows the models available on the connected plan, and the user can switch between them. This mirrors Notion.
- **Reference UX is Notion's launch video.** Screenshots are saved in `references/`:
  1. `notion-01-model-picker-connect.png`: the chat model picker shows Auto (default, "Balances speed, effort, and cost"), a Recommended group, "Other models >", and a **Subscriptions** section with **"Connect ChatGPT plan"**. The picker itself is an entry point.
  2. `notion-02-chatgpt-account-chooser.png`: an OpenAI-hosted screen, "Continue to Notion / Choose an account to sign in with ChatGPT", with the signed-in account, "Log in to another account" and "Create account".
  3. `notion-03-chatgpt-consent.png`: an OpenAI-hosted consent screen, "Connect ChatGPT and Notion", with a toggle **"Use your ChatGPT plan: Allow Notion to consume usage from your ChatGPT Plan's limits"**, then Continue and Cancel. The footer says connections are managed in ChatGPT settings.
  4. `notion-04-model-picker-connected.png`: after connecting, the picker has a **"ChatGPT" section labelled "Enabled"**. It lists the plan's models (GPT-5.6 Luna, Terra, Sol, GPT-5.2, 5.4, 5.5, GPT-6 Luna "New"), and the selected one shows a check. The composer chip shows the chosen model with the OpenAI logo.
  5. `notion-05-chat-on-chatgpt-plan.png`: the assistant keeps its identity: "You're speaking with Notion AI, but I'm using your ChatGPT plan."

**What this means:**
- Screens 2 and 3 are OpenAI's own "Sign in with ChatGPT" consent pages, with the plan-usage scope. So the PO's target flow **is** the partner-gated OpenAI program described below. It can't be swapped for the open API-key route without changing the user experience. Users would have to create, fund and paste a key, and their ChatGPT plan wouldn't be used.
- Getting into OpenAI's commercial trial is therefore on the critical path. Delivery of the connect step depends on OpenAI issuing ContentStudio a client ID.

## PO Direction #2 (2026-10-02): one connection model for every provider

The PO widened the scope. ChatGPT subscription is no longer the only option. It becomes **one provider in a single "AI connections" model**, alongside API keys. The reference is Helpin's own "Add connection" modal, with screenshots in `references/helpin-0*.png`:

- **Modal:** "Add connection", subtitle "Personal connection · Only you" with an info icon.
- **Provider dropdown:** OpenAI API key, Anthropic API key, OpenRouter API key, ChatGPT subscription, Compatible endpoint.
- **Fields:**
  - Connection name, prefilled with the provider name.
  - API key, with an info icon, for key providers.
- **ChatGPT subscription:** "You will sign in to ChatGPT with a one-time code." The button is **Continue to ChatGPT**.
- **Compatible endpoint:** a dropdown of endpoints that support has provisioned. When there are none, it shows "No compatible endpoints are available. Contact Helpin support to request one."

**Decisions:**

| Decision | Choice |
|---|---|
| v1 providers | **OpenAI API key, Anthropic API key, Google Gemini API key, OpenRouter API key, ChatGPT subscription.** ChatGPT subscription is enabled only once OpenAI approves ContentStudio. |
| Compatible endpoint | **Deferred to v2** |
| Connections per user | **Many, each with a name.** For example "OpenAI" and "Claude" side by side. The picker groups models by connection. |
| Scope per connection | Personal. Only the user who added it can use it. |

**Why this resolves the gating problem:** API keys need no partner approval, so v1 can ship without waiting on OpenAI. The ChatGPT subscription option lights up in the same list once OpenAI approves us.

**Codebase fit:**
- `AG/utils/model_registry.py` `_create_model` already builds `Claude`, `OpenAIChat`, `Gemini` and `OpenRouter` models. The only gap is a per-request `api_key`, since keys are read from env today.
- An Anthropic key gives Claude, which is the model family our chat prompts are tuned for. It's the lowest-risk option on quality.

**Provider notes:**
- **Anthropic:** API keys are fine for commercial apps. Only subscription OAuth is banned.
- **OpenRouter:** hundreds of models, and tool calling varies. The picker must only list models that support tool calling, because chat needs tools.
- **Gemini:** supports tool calling. ContentStudio already runs Gemini for analytics insights.
- **ChatGPT subscription:** Helpin's copy mentions a one-time code, which is the device-code style sign-in the Codex CLI uses. Our research says hosted commercial apps need OpenAI partner approval for plan usage. **Open question:** check with the Helpin team how they got access before relying on the same route.

## TL;DR (first research pass)

1. **OpenAI's program for this is real, but ContentStudio can't use it yet.** OpenAI launched it at DevDay on 29 Sep 2026. For hosted commercial apps it is limited to "selected commercial partners through a limited trial". Joining means filling in an interest form, with no published timeline. Of the 16 launch partners, 11 can spend users' plans, and almost all of them are developer tools. Notion is the only mainstream app among them. No social media tool is a partner.
2. **Even after acceptance, two capabilities are unconfirmed for plan usage:**
   - Tool calling. AI chat depends on it to create posts, read analytics and draft skills.
   - Image generation.
   OpenAI's docs only show streaming text.
3. **What we can ship now is BYOK with an OpenAI API key.** Publer is the only social media competitor with any version of this feature, and it uses an API key, not ChatGPT sign-in. Zapier, Make, n8n, JetBrains, Raycast and TypingMind all do the same. A ChatGPT Plus subscription does **not** include API credits, so BYOK is a different promise from Notion's.
4. **This is real engineering work in ai-agents, not just a settings toggle.**
   - Every model client is built from server-side keys.
   - The chat Team is a process-wide singleton.
   - Chat currently runs on Claude, with GPT-5 only as a fallback.
   Running a turn on a user's OpenAI account means per-request model and key injection. It also means a provider switch to GPT for that user, with prompt and quality regressions likely.
5. **Only text is coverable.** Image, video and carousel photos run on fal, Veo and BytePlus, which ContentStudio pays for. Metering must split: text is skipped under BYO, image and video stay on ContentStudio credits.
6. **Recommended path:**
   - Submit OpenAI's interest form now.
   - Phase 1 ships BYOK with an OpenAI API key for AI chat text.
   - Phase 2 adds "Continue with ChatGPT" once OpenAI accepts us.
   - Both phases share one settings section, one model picker and one error model.

---

## Part A: Market and Competitor Research

### What is this feature?

Users who already pay for AI (ChatGPT Plus or Pro, or an OpenAI API account) connect it to a third-party app. The app's AI then runs on that account instead of the app's own credits.

**Why users want it:**
- They don't want to pay twice.
- They get the model they already trust.
- They no longer hit the app's credit limits.

**The two flavours are not the same promise:**

| | Subscription sign-in (SIWC) | API key (BYOK) |
|---|---|---|
| What the user connects | A ChatGPT Plus or Pro plan | A pay-as-you-go OpenAI API key |
| How they pay | Usage comes out of their existing plan allowance | Billed per token to their OpenAI API account |
| Setup | One click, OAuth | Create a key on platform.openai.com, add billing, paste it |
| Who can offer it | Selected partners only | Any app |

### What OpenAI offers third parties (verified)

**Two capabilities:**
- Identity-only sign-in (`openid profile email`). Sign-in-only partners: Airtable, Canva, GitLab, HubSpot, Supabase.
- Plan usage ("token sharing"). It needs the `chatgpt.tokens.use.direct` scope plus `offline_access` and `resource.invoke`.

**Access:**

| App type | Plan usage available? |
|---|---|
| Open-source and local apps | Yes. A dynamic client ID, no pre-registration. |
| Hosted commercial SaaS (this is ContentStudio) | Only through the "limited trial" via the interest form. No criteria or timeline published. |

**Launch partners with plan usage:** Amp Code, Conductor, Dactyl, Devin, Hermes Agent, Hyperagent, Kilo Code, Notion, Vercel, Vorflux, Warp. Lovable is coming soon.

**Technical model:**
- OAuth 2.0 authorization code with PKCE plus OIDC. Issuer `https://auth.openai.com`. Commercial client IDs start with `oaiapp_`. Exact callback URLs per environment.
- Access tokens last 1 hour. Refresh tokens last 30 days and rotate on every refresh.
- Inference goes through the Responses API using the user's access token.
- Models are discovered per connection. There is no fixed list.
- Reported constraints: `stream: true` and `store: false` are required, and sampling fields like `temperature` are rejected.
- Tool calling and image generation on plan usage are **not confirmed**.

**Errors apps must handle:**

| Error | Status | Required behavior |
|---|---|---|
| `subscription_sharing_usage_limit_exceeded` | 429 | Pause, link the user to ChatGPT Settings > Usage |
| `subscription_sharing_user_not_eligible` | 403 | Explain. Don't retry or loop OAuth. |
| `subscription_sharing_invalid_user` | 401 | Ask the user to reconnect |
| `invalid_grant` / `refresh_token_invalidated` / `refresh_token_reused` | n/a | Clear tokens and redo OAuth |

**OpenAI does not notify the app when a user disconnects.** The app only finds out from failed requests.

**User-side controls:**
- Usage draws on the existing "ChatGPT Work and Codex" allowance.
- Users can cap each app at 10 to 100% of weekly usage.
- Plus users share a 5-hour window across all connected apps.
- When the cap is hit, usage stops. It never falls back to the partner's billing.
- At a 100% cap, the app can spend the user's purchased ChatGPT credits. WorkOS flags this as a surprise-billing risk.
- Disconnecting happens in ChatGPT Settings > Security and login. Signing out of the app does not disconnect.

**Security rules:**
- Keep the client secret server-side.
- Never pass tokens, codes or PKCE verifiers to browser JS.
- Verify JWTs against JWKS.
- Use OpenAI's approved "Continue with ChatGPT" button branding.

**Commercials:** no revenue share terms or rate limits were found.

**Other providers:**
- **Anthropic** explicitly bans using Claude subscription OAuth in third-party tools (terms of 19 Feb 2026, enforced against Cursor, Cline, OpenCode and others).
- **Google** has no program.
- So any "bring your own Claude or Gemini" would have to be an API key.

### Notion deep dive

| Aspect | Notion's implementation |
|---|---|
| Requirements | Business or Enterprise workspace, a member (guests excluded), ChatGPT Plus or Pro |
| Level | **Per member.** Each person connects their own ChatGPT. |
| Entry points | Settings > Notion AI > Usage > Connected subscriptions > Connect. Also the model picker > Subscriptions > Connect ChatGPT plan. |
| Billing selection | Driven by the chosen model. A specific GPT model bills the ChatGPT plan. Auto or non-GPT models bill Notion credits. The picker shows which plan pays **before** sending. |
| Coverage | Notion Agent only. Not Custom Agents, not every model. |
| At the limit | **Prompted fallback:** "Notion Agent will ask you to switch to your workspace's Notion credits." Switches back automatically when the allowance resets. |
| Controls | The user can toggle ChatGPT funding off without disconnecting. Admins only see connect and disconnect events in audit logs. There is no documented way for admins to disable it. |
| Data | Notion's subprocessor terms apply. Chats do not appear in ChatGPT history. The connection is "used only for billing and usage attribution". |
| Commercial angle | Gating it to Business makes it an upsell lever. It offsets metered Custom Agent credits. |

### Competitor Analysis Table

| Competitor | Has Feature? | Key Capabilities | Pricing Tier | UX Approach | Unique Differentiator |
|---|---|---|---|---|---|
| **Notion** | Yes, ChatGPT plan (SIWC) | Notion Agent on the user's GPT plan, prompted fallback, auto switch-back | Business, Enterprise | Settings > Notion AI > Usage, plus "Subscriptions" in the model picker. The picker shows who pays. | First mainstream SIWC partner. Per member. Audit log events. |
| **Publer** (social) | **Yes, API key BYOK** | The user's OpenAI key powers AI Assist text and images | Free and Professional need a key. Business ($21) includes unlimited AI. | Integration Settings > OpenAI > paste key (also on mobile) | **Only social media competitor with BYO.** Uses it as a plan gate. |
| **Zed** | Yes, ChatGPT sign-in | Chat on the user's OpenAI subscription, no API key | Any | "Use an existing subscription" provider list | Also accepts Claude, Copilot and Cursor subscriptions via agents |
| **Raycast** | Yes, both | ChatGPT or Claude subscription via the local Codex CLI or Claude Code. BYOK for OpenAI, Anthropic and Google. | Paid Raycast plan | Settings > AI > Models & Providers | Prompts never touch Raycast servers |
| **JetBrains AI** | Yes, both | Codex via ChatGPT sign-in. BYOK for chat and agents. | BYOK works without an AI subscription | "Bring Your Own API Key" on the chat start screen | Three billing paths in one chat |
| **Devin, Warp, Amp, Vercel, Kilo Code** | Yes, SIWC | Coding agents on the user's ChatGPT plan | Varies | "Continue with ChatGPT" | OpenAI launch partners |
| **Cursor** | BYOK only | OpenAI key for local chat and agent | Any | Paste key | OpenAI reportedly ending Cursor's model access (12 Nov cutoff), showing partner terms can change suddenly |
| **Zapier / Make / n8n** | BYOK | OpenAI modules on the user's key | Any | Credential field | Docs state outright that ChatGPT Plus does not give API credits |
| **TypingMind** | BYOK | Multi-provider chat on user keys | One-time license | Paste keys | Whole business is BYOK |
| **Canva, Airtable, HubSpot** | Sign-in only | Identity via ChatGPT, no plan usage | n/a | Login button | n/a |
| **Buffer** | No | Built-in AI Assistant. MCP and OAuth so ChatGPT and Claude can drive Buffer. | All | n/a | Inbound only |
| **Hootsuite** | No | OwlyWriter and OwlyGPT folded into the "Wisdom" agent | All | n/a | Bundled AI |
| **Sprout Social** | No | AI Assist. Trellis with 100 credits per seat per month. | All | n/a | Credit-metered agent |
| **Later** | No | AI Caption Writer with 5 to 100 credits per tier | Starter+ | n/a | Tight credits |
| **Agorapulse** | No | AI captions, alt text, brand tone. MCP for ChatGPT and Claude. | All | n/a | Inbound MCP |
| **Vista Social / Metricool** | No | Built-in AI. First-party MCP servers. | Varies | n/a | Inbound MCP |
| **Loomly, Sendible, SocialBee, Predis, Ocoya** | No evidence | Built-in AI generators | Varies | n/a | n/a |

### Common Patterns

1. **Two tiers:** subscription OAuth (new, gated, mostly coding tools) and API key BYOK (mature, everywhere).
2. **The model choice decides who pays,** and the picker says so before the user sends.
3. **Fallback is prompted, not silent.** The user never gets silently switched to the app's billing, or the reverse.
4. **The settings live under AI > Usage or Providers,** with Connect, Disconnect and a funding toggle.
5. **Per user, not per workspace.** Usage follows the person who connected.
6. **Plan gating goes both ways:**
   - Notion makes it a Business+ upsell.
   - Publer flips it: lower tiers bring their own key, the top tier bundles AI.
7. **Social media tools go "inbound"** (MCP so ChatGPT can drive them) rather than letting users power in-app AI. ContentStudio already has the inbound side, so this would be the reverse direction.

### Differentiators available to ContentStudio

- **First social media tool whose AI runs on the user's ChatGPT plan,** if OpenAI accepts us.
- **A better BYOK than Publer's bare paste-key field:**
  - a funding label in the model picker
  - a clear limit-reached choice
  - an admin control
  - mobile honoring the connection
- **A two-way story:** ContentStudio inside ChatGPT (MCP) and ChatGPT inside ContentStudio.

### User Expectations

**Table stakes**
- One-click connect and disconnect in Settings, with a visible status (account or masked key, last used).
- A model picker that shows which account pays.
- Non-technical messages for an invalid key or a reached limit, with a choice to continue on ContentStudio credits.
- No silent charges.
- The key is never shown in full again.
- It works in AI chat on web.

**Delighters**
- Automatic switch-back when the ChatGPT allowance resets.
- Usage visible per source.
- The Flutter AI chat honors the same connection.
- An admin toggle to allow or block member connections, plus audit events. Notion lacks the toggle.
- Coverage beyond chat, such as composer captions.

---

## Part B: Codebase Analysis

Paths are abbreviated: BE = `contentstudio-backend/`, FE = `contentstudio-frontend/src/`, AG = `contentstudio-ai-agents/src/`, FL = `contentstudio-flutter/lib/`.

### Existing Related Code

**AI chat, end to end**

Frontend:
- Chat UI: `FE/modules/AI-tools/` (`AIChatMain.vue`, `AIChatModal.vue`, `AIChatWidget.vue`, `ChatHeader.vue`, `BotChatTemplate.vue`).
- Composables: `useAIChatStream.ts`, `useAiChatEngine.ts`.
- Stream reducers: `utils/reduceStreamFrame.ts`, `utils/reduceStreamEvent.ts`. They handle `usage_error`.
- Store: `FE/stores/core/useAIChatStore.ts`.
- Endpoint: `ai/chatWithStreaming` (`FE/config/api-utils.ts:215`).
- `ModelProviderSelect.vue` already exists, but only for image and video models.

Backend:
- `POST ai/chatWithStreaming` goes to `AIController::processAgnoAgent` (`BE/app/Http/Controllers/AI/AIController.php:1191`). In order, it:
  1. Checks credits (`PlanHelper::checkAvailableAICredits`).
  2. Enforces a hard minimum of 5 text credits (around line 1240, SSE `usage_error` / `insufficient_text_credits`).
  3. Builds inputs in `AiChatHelper::buildWorkflowInputs` (`BE/app/Helpers/Ai/AiChatHelper.php:521`).
  4. Proxies SSE through `AgnoAgentServices('stream/generate')`.

ai-agents:
- `POST /stream/generate` (`AG/api/routers/streaming_router.py:3011`).
- `StreamRequest.model_choice.text` exists, **but chat ignores it**. It is only copied into metadata and used by `ai_post_library.py:446`.
- The Team is a global singleton in `AG/orchestration/team.py` `get_content_team()`:
  - coordinator `claude-haiku-4-5`
  - fallback `gpt-5`
  - follow-ups `claude-haiku-4-5`
- Members come from `AG/orchestration/agents.py:2058`: primary `claude-sonnet-4-6`, fallback `gpt-5`.
- The moderation pre-hook (`AG/orchestration/guardrails.py`) uses ContentStudio's OpenAI key.

Model factory:
- `AG/utils/model_registry.py`. `TEXT_MODELS` at line 18 includes `gpt-5`, `gpt-5-mini`, `gpt-5.4-nano`, `gpt-4o`.
- `_create_model` (line 2921) and `resolve_model` (line 3031) take **no `api_key` parameter**. Keys come only from env (`AG/utils/config.py:16-23`).

**Other AI features and what a user's OpenAI account could cover**

| Feature | Provider today | Coverable by user's OpenAI? |
|---|---|---|
| AI chat text (Team, members, carousel writer) | Claude primary, GPT-5 fallback | Yes, with a model switch to GPT |
| Composer captions (`AG/agents/content/caption_writer.py`, own model builder) | Claude | Yes, with a model switch (P2 candidate) |
| Prompt enhancers, brand describe or decide | gpt-5.4-nano / mini | Yes |
| Evergreen variations, listening mention analyzer | OpenAI | Yes (not user-facing chat) |
| AI insights / analytics (`AG/agents/analytics/base.py`) | Gemini, then Claude | Only with a model switch |
| Images, including gpt-image models | **fal** (`model_registry.py:153+`) | **No**, stays on CS credits |
| Video | fal, Veo, BytePlus | **No** |
| Moderation, audio transcription | CS OpenAI key | No, stays on CS |
| Skills | Inherit the Team's models | Follow chat |

**BYOK precedents:** there is no LLM BYOK code in any repo. There's a dormant FE "YouTube API key" setting (`FE/api/workspace.ts:82`) with no BE route behind it, so it's dead code and not reusable.

### Reusable Components / Services

- **The Canva integration is the OAuth template:**
  - PKCE auth URL with encrypted state (`CanvaController::getAuthorizationUrl`).
  - `oauth.callback` middleware (`BE/app/Http/Middleware/OAuthCallbackMiddleware.php`).
  - Popup HTML response.
  - Disconnect revokes the token.
  - Routes in `BE/routes/web/integrations.php:147,240`.
- **Encrypted token model:**
  - `BE/app/Models/Integrations/CanvaIntegration.php` encrypts tokens in mutators via `SocialHelper::encryptToken` (`BE/app/Libraries/Publish/Helper/SocialHelper.php:83-110`).
  - Refresh-on-expiry and refresh-on-401: `BE/app/Services/Canva/CanvaService.php:40-95`.
  - Reconnect-required state: `CanvaRepo::markInvalid` sets `validity`, `invalid_tries` and `validity_error` (`BE/app/Repository/Integrations/CanvaRepo.php:105`).
- **The FE integrations page** can host it:
  - `FE/modules/integration/components/sections/ConnectedAppsSection.vue` (DataTable with connect, reconnect, disconnect).
  - `useConnectedApps.ts` (rows with `allowed` and `lockedReason` from `canAccess`).
  - `OAuthConsent.vue`.
- **Plan gating:**
  - A plan feature key via migration (pattern: `canva_integration`).
  - `SubscriptionLimits::getSubscriptionLimits` (`BE/app/Libraries/Settings/SubscriptionLimits.php:291`).
  - FE `useFeatures().canAccess()` (`FE/modules/billing/composables/useFeatures.ts:95`) with the `locked ? showUpgradeModal()` pattern from `SettingSidebar.vue`.
- **Rollout flag:** the `feature.flag:<name>` middleware (`BE/app/Http/Middleware/FeatureFlagMiddleware.php`).

### Integration Points

**BE, `processAgnoAgent`:**
- Load the user's AI connection.
- Skip the 5-text-credit minimum when it's active.
- Pass a decrypted credential plus a provider and model override to ai-agents per request.
- Mark the run `byo` so all three billing paths charge zero text credits but still charge images:
  - in-stream `AiChatHelper::deductAiCredits` (line 1508) via `AiChatUsage::claim`
  - the Kafka backstop `ChatUsageEventHandler`
  - cancel billing

**ai-agents:**
- Add an `api_key` override to `_create_model` / `resolve_model`.
- Build the Team per request, or apply model overrides per request, instead of reusing the singleton.
- Decide what happens to the Claude coordinator and follow-ups and the server GPT-5 fallback, so the user's account really covers the whole turn.
- Skip text credit validators (`AG/utils/credit_validation/validation.py`, `streaming_router.py:1925,1966-1990`) when BYO is active.
- Map OpenAI 401, 429 and `insufficient_quota` to new SSE error codes.

**FE:**
- A new settings page or row.
- A funding label in the chat model picker.
- In `ChatHeader.vue`, `AiCreditsChip` changes to "Using your OpenAI account".
- New error codes in `reduceStreamFrame.ts` / `reduceStreamEvent.ts`.

**Settings placement options** (`FE/modules/setting/config/routes.ts`, `SettingSidebar.vue`):
- A new "AI Setup" route next to Integrations in the Workspace group.
- A row in Integrations > Connected apps, like Canva.
- The Account group.

### Technical Considerations

- **Data:** a new Mongo collection (e.g. `ai_provider_connections`) modelled on `CanvaIntegration` with these fields: `user_id`, `workspace_id`, `provider`, `auth_type` (`api_key` | `oauth`), encrypted secret(s), `expires_at`, `selected_model`, `validity`, `validity_error`, `last_validated_at`. Plus a plan feature key via migration.
- **Security:**
  - Never return the key to the FE, only a mask.
  - Validate it on save with a cheap `GET /v1/models`.
  - Pass it BE to ai-agents per request over the existing authenticated channel.
  - **Keep it out of:**
    - Langfuse traces (`streaming_router.py:1299+` records input and metadata)
    - `meta_highlights` logging
    - the Redis Agno session DB (`cs_agno` prefix)
    - Kafka `ChatUsageReportedEvent`
- **Credits are account-level.** `WorkspaceRepo::updateCaptionGenerationCreditLimit` updates every workspace the super admin owns. `WalletService` is the X pay-per-use wallet, **not** AI metering. Only the AI plan counters are affected here.
- **Quality:** chat prompts and tool routing are tuned for Claude Sonnet and Haiku. Moving a user's turn to GPT needs an eval pass. `AG/agents/base.py:160-175` already handles OpenAI's temperature and `max_completion_tokens` quirks. SIWC's no-`temperature` rule would need checking against it.
- **Health:** mark a connection invalid on a 401, Canva `markInvalid` style. Optionally notify "reconnect required".
- **Developer surfaces:** nothing API-facing changes unless model or funding selection is exposed through the public API, CLI or MCP. Likely N/A for v1.

### Mobile Impact

- Flutter calls the same `/ai/chatWithStreaming` (`FL/features/ai_assistant/data/ai_assistant_endpoints.dart:9`, `ai_assistant_service.dart:129`).
- It shows a credits pill and info sheet from `limits` (`presentation/ai_assistant_view.dart:392-444`, `domain/entities/ai_limits.dart`).
- It maps `usage_error` in `data/ai_stream_event_mapper.dart:49,143,170`.
- Routing happens server-side, so mobile chat keeps working unchanged. New error codes would fall through as generic errors, though.
- A single `[Flutter]` story should:
  - handle the new errors with a "manage this on the web" message
  - relabel or hide the text-credits pill when the BE reports BYO
  - optionally show a "Using your OpenAI account" badge
- **No connection setup on mobile in v1.**

---

## Risks and Open Questions

1. **SIWC access (blocking for the CEO's exact ask).** It needs OpenAI's acceptance into a limited commercial trial. **Action: submit the interest form now.** Ask OpenAI about tool calling, image generation, models, rate limits and data terms for plan usage.
2. **Tool calling on plan usage is unconfirmed.** Without it, ChatGPT-plan chat could only do plain text: no post creation, no analytics reads.
3. **Provider switch.** Our chat is Claude-first. A BYO user's chat runs on GPT, with quality and behaviour differences. It needs an eval pass and possibly prompt tuning.
4. **Mixed billing.** Images and video stay on ContentStudio credits, so the UI must make that clear inside a BYO chat.
5. **Who connects, and whose account pays.**
   - Per user: each member uses their own account. This matches Notion and how chat history is keyed.
   - Per workspace: the owner's key is shared by everyone.
   This is a PO decision.
6. **Plan gating and cannibalization.**
   - Notion-style: gate it to higher plans as an upsell.
   - Publer-style: BYO on lower plans, bundled AI on higher.
   This is a PO decision.
7. **BYOK policy.** OpenAI historically discouraged users sharing keys with apps. That guidance was later softened, and the practice is widespread. Treat it as accepted practice, not explicitly blessed.
8. **Revocation blind spot (SIWC).** There is no disconnect webhook. Health must come from request errors, and our own disconnect must call OpenAI's revocation endpoint.
9. **Privacy and DPA.** Prompts would go to OpenAI under the user's account terms as well as ours. The privacy policy and subprocessor list may need updates.
10. **Messaging.** Don't market "use your ChatGPT subscription" until SIWC is approved. Phase 1 is "use your own OpenAI account / API key".

## Recommended Approach (revised after PO direction)

1. **Apply to OpenAI now.** Submit the SIWC interest form, ask for the plan-usage scope, and confirm:
   - tool calling on plan usage
   - the model list
   - rate limits
   - data terms
   - which callback URLs are needed for white-label domains
2. **Author the epic around the Notion-style flow:**
   - connect ChatGPT plan from Settings and from the chat model picker
   - OpenAI consent
   - the plan's models listed in the picker
   - orchestrator, assistant and agents run on the plan
   - captions, images and video stay on ContentStudio credits
   - prompted fallback at the limit
3. **Only the connect step is blocked on OpenAI's client ID.** The rest can be built in parallel and tested behind a feature flag:
   - per-request model and credential routing in ai-agents
   - the metering split
   - picker grouping
   - error states
4. **API key BYOK stays a contingency.** Use it only if OpenAI declines or stalls. It isn't in this epic unless the PO asks for it.

### Original recommendation (superseded)

1. **Now:** submit OpenAI's SIWC interest form. This is a business action, not a story.
2. **Phase 1 (this epic):**
   - **Feature:** "Use your own OpenAI account" with an API key, for **AI chat text**.
   - **Settings:** paste key, validate it, store it encrypted, show it masked, pick a default GPT model, a funding toggle, disconnect.
   - **Chat:** the model picker labels the funding source.
   - **Limits and errors:** at a limit or error, a prompted "Continue with ContentStudio credits".
   - **Billing:** images and video stay on credits.
   - **Admin control:** allow or block.
   - **Mobile:** honors it with labels and error states.
3. **Phase 2 (after OpenAI acceptance):** add "Continue with ChatGPT" (SIWC OAuth) as a second connection method in the same section. It reuses the picker, error model and fallback, plus automatic switch-back when the allowance resets.
4. **Later candidates:** composer caption generation on the user's account. Other providers' keys (Anthropic, Gemini), API key only.

## Sources

**OpenAI:**
- developers.openai.com/siwc (+ `/quickstart`, `/request-client-id`, `/website`, `/token-sharing-open-source/token-reference`, `/token-sharing-open-source/errors-and-recovery`)
- developers.openai.com/cookbook/articles/sign-in-with-chatgpt
- learn.chatgpt.com/docs/sign-in-with-chatgpt
- help.openai.com/en/articles/20001542, help.openai.com/en/articles/20001410

**Coverage of the OpenAI launch:**
- workos.com/blog/sign-in-with-chatgpt-plan-usage-scope
- thenewstack.io/sign-in-with-chatgpt
- notebookcheck.net (signing out does not stop usage)
- securitybrief.com.au (16 partners)

**Notion:**
- notion.com/help/use-your-chatgpt-plan-with-notion-agent

**Anthropic and Google:**
- winbuzzer.com/2026/02/19/anthropic-bans-claude-subscription-oauth-in-third-party-apps-xcxwbn
- discuss.ai.google.dev/t/175577

**Other apps:**
- zed.dev/docs/ai/use-an-existing-subscription
- manual.raycast.com/ai/bring-your-own-subscription
- jetbrains.com/help/ai-assistant/bring-your-own-key-byok.html
- docs.n8n.io/integrations/builtin/credentials/openai
- help.zapier.com/hc/en-us/articles/14860148802829
- digitalapplied.com/blog/openai-ends-cursor-model-access-november-cutoff

**Social media competitors:**
- publer.com/help/en/article/how-to-connect-my-openai-account-1dmawuf
- buffer.com/ai-assistant
- support.vistasocial.com MCP article
- support.agorapulse.com open API article

**BYOK policy:**
- community.openai.com/t/14538
- community.openai.com/t/446168

**Caveats:**
- Some help.openai.com pages returned 403. Their details come from learn.chatgpt.com and secondary coverage.
- Tool calling and image generation for plan usage could not be verified from a verbatim page.
- Publer's help page names models "GPT-6 Astra" and "GPT Image 2.5 Flare". These were not verified.
