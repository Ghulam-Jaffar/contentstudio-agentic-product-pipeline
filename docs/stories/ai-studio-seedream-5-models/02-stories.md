# Stories: Add Seedream 5 Pro and Lite to AI Studio

Two stories: register the models in the AI backend, then expose them in the AI Studio model selector.

---

## [BE] Register Seedream 5 Pro and Seedream 5 Lite image models

### Description

Add Seedream 5 Pro and Seedream 5 Lite as selectable image-generation models in the AI backend so AI Studio can offer them.

### Acceptance criteria

- [ ] Seedream 5 Pro and Seedream 5 Lite are registered in the image model registry and callable through the existing image generation path.
- [ ] Each model returns generated images successfully for a standard text-to-image request.
- [ ] Each model is credit-metered using the existing image credit rules.
- [ ] Invalid or unsupported parameters for these models return a clear, structured error.
- [ ] The models appear in the backend list of available image models consumed by the frontend.

### Impact on existing data

None. Additive model registration.

### Impact on other products

Backend (AI generation) is web-only surface. Mobile and Chrome extension: N/A.

### Dependencies

None. Pairs with the frontend selector story below.

### Global quality and compliance checklist

- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (error/status messages localized or fallback handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Image model registry: `contentstudio-ai-agents/src/agents/image/models_registry.py`, `src/tools/dedicated/models.py`, provider layer `src/agents/image/providers/fal.py`. Add the Seedream 5 Pro and Lite identifiers here (confirm the exact provider slugs).
- Follow how the currently registered image models are declared and credit-metered.

---

## [FE] Show Seedream 5 Pro and Lite in the AI Studio model selector

### Description

Let users pick Seedream 5 Pro or Seedream 5 Lite when generating images in AI Studio.

### Workflow

1. User opens image generation in AI Studio and opens the model selector.
2. User sees "Seedream 5 Pro" and "Seedream 5 Lite" alongside the existing models.
3. User selects one and generates as usual.

### Acceptance criteria

- [ ] The model selector lists "Seedream 5 Pro" and "Seedream 5 Lite".
- [ ] Selecting either model generates images using that model.
- [ ] Model names are shown exactly as "Seedream 5 Pro" and "Seedream 5 Lite".
- [ ] Any per-model options or limits surfaced by the backend are reflected in the UI.

### UI copy

- Model option labels: "Seedream 5 Pro", "Seedream 5 Lite"

### Impact on existing data

None.

### Impact on other products

Web only (AI generation). Mobile and Chrome extension: N/A.

### Dependencies

Depends on: **[BE] Register Seedream 5 Pro and Seedream 5 Lite image models**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Model selector components: `contentstudio-frontend/src/modules/AI-tools/components/ModelProviderSelect.vue`, `contentstudio-frontend/src/modules/publisher/ai-content-library/components/form/ImageModelControls.vue`. If the model list is driven by the backend list endpoint, the new models may appear automatically; verify labels render correctly.
