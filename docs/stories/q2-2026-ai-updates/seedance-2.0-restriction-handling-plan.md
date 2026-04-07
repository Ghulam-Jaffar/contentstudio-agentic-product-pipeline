# Seedance 2.0 — fal.ai Restriction Handling Plan

## Background

fal.ai granted us conditional access to Seedance 2.0 (ByteDance's trending video generation model) with three compliance requirements. Here's how we handle each:

---

## Restriction 1: Geographic — No US or Japan users

**How we handle it:**

- Seedance 2.0 appears in the model list for all users — no frontend filtering
- When a user selects Seedance 2.0 and submits a generation request, the backend checks the user's country via Cloudflare headers (`X-COUNTRY-NAME`) which are already available on every request
- If the user is in the US or Japan, the request is rejected with a clear message: *"Seedance 2.0 is not available in your region. Please try a different video model."*
- No extra infrastructure needed — Cloudflare already resolves country from IP on every request

**Why this approach:**

- Zero frontend complexity — model stays visible, backend enforces the rule
- Works for all user types (trial, team members, any plan) since it's IP-based
- Industry-standard approach for geo-compliance

---

## Restriction 2: B2B Verification — Only serve businesses

**How we handle it:**

- ContentStudio is a B2B SaaS product by nature — it's a social media management platform used by businesses, agencies, and marketing teams
- Even single-user Standard plan subscribers are using it for business purposes (managing social media accounts, scheduling posts, running campaigns)
- This is a self-certification argument — no additional verification logic needed on our end
- If fal.ai ever asks, we can confirm that our entire user base consists of business users

**Additionally, trial users are blocked from Seedance 2.0.** Only paid plan users can access it. This prevents sign-up-and-churn abuse and strengthens the B2B compliance position — every Seedance 2.0 user is a paying business customer.

**Implementation:**
- **Frontend:** Seedance 2.0 appears in the model list but is **disabled** for trial users with a tooltip: *"Seedance 2.0 is available on paid plans. Upgrade to unlock."* The user's trial status is already known on the frontend.
- **Backend (safety net):** Also checks if user is on a trial subscription (`trial-*` prefix). If so, rejects with: *"Seedance 2.0 is available on paid plans only. Upgrade to access this model."*

---

## Restriction 3: End User Identification — Pass `end_user_id` + ability to restrict

**How we handle it:**

- Pass `workspace_id:user_id` as the `end_user_id` parameter in every Seedance 2.0 API call to fal.ai
- Maintain a simple blocklist (DB config or collection) — if fal.ai ever requests us to restrict a specific end user, we add their ID to the blocklist
- Before making any Seedance 2.0 fal.ai call, check the blocklist — if blocked, reject with a support-contact message

**Minimal code change** — just add the parameter to the API call and a blocklist check.

---

## Summary

| Restriction | Approach | Effort |
|---|---|---|
| Geo (no US/Japan) | Backend checks Cloudflare country header on Seedance 2.0 requests, rejects with user-friendly message | Low |
| B2B verification | Self-certification + block trial users from Seedance 2.0 | Low |
| End user ID | Pass `workspace_id:user_id` in fal.ai calls + simple blocklist | Low |

**Total additional effort beyond standard model integration:** ~1 day of backend work for the geo-check and blocklist mechanism. No frontend changes beyond adding the model to the existing hardcoded list.
