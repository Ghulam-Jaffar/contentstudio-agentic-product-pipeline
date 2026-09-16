# Story: Add GPT Image 2.5 models

**Date:** 2026-09-16
**Stories:** 1

---

## Story 1

### Title

**[Full Stack] Add the GPT Image 2.5 models everywhere images are generated or edited**

### Description

As someone generating images in ContentStudio, I want OpenAI's GPT Image 2.5 models available, so that I can pick between a fast everyday model and a slower high-fidelity one depending on what the work needs.

GPT Image 2.5 comes in two variants, each doing both generation from a prompt and editing of an existing image, so there are four models in total:

| Variant | What it is for |
|---|---|
| **Flare, text to image** | The everyday default. Fast, high quality, natural lighting and rich textures, handles complex layouts, and can produce **transparent backgrounds** |
| **Flare, edit** | Precise editing that changes only what was asked, keeping the subject, composition and background intact, and keeping reference subjects recognisable across styles and successive edits |
| **Sunburst, text to image** | Premium visual work. Extra fidelity on intricate detail, **in exchange for longer generation times** |
| **Sunburst, edit** | The tightest editing control, with edits scoped precisely to the instruction and subject and composition preserved across many rounds of revision |

They need to be available **everywhere a model can be chosen or used**, not only in the AI Studio picker. That means the model lists in AI Studio, the image tools, the assistant, the composer, and the public surfaces: the API, the MCP tools and the CLI.

---

### Workflow

1. A user opens any place in ContentStudio where they generate an image and opens the model picker.
2. Under OpenAI they see the two new models named clearly enough to tell apart, with a short line explaining when to reach for each.
3. They pick **Flare** and generate. It returns quickly, at the quality expected of an everyday default.
4. They pick **Sunburst** for a piece that needs finer detail. Before generating, they can see that it takes longer, so the wait is expected rather than alarming.
5. They upload an existing image and ask for a change. The edit variant of whichever model they chose is used, and only what they asked for changes.
6. They repeat the edit several times in a row, and the subject and composition stay consistent through every round.
7. Their choice of model is remembered for next time, as it is with the existing models.
8. A developer generating images through the API, the MCP tools or the CLI can select the same models by name and gets the same results.

---

### Acceptance criteria

**The models are available**

- [ ] All four models are available: Flare text to image, Flare edit, Sunburst text to image, Sunburst edit
- [ ] They are grouped under **OpenAI** as the provider, alongside the existing OpenAI models
- [ ] Generating from a prompt works on both variants
- [ ] Editing an existing image works on both variants, and asking for an edit uses the edit variant of the model the user selected rather than sending an edit to the generation model
- [ ] The aspect ratios each model supports are offered, and selecting a model updates the ratios available to match what it actually supports
- [ ] The number of images that can be produced in one go is set per model according to what each supports
- [ ] **Flare's transparent background support is available to the user**, since that is a real capability of the model and is the sort of thing people pick a model for
- [ ] The user's model choice is remembered between sessions, the same way it is for existing models

**Everywhere a model can be chosen**

- [ ] The models appear in the AI Studio model picker
- [ ] They are available in the AI Studio image tools that let a model be chosen
- [ ] They are available when the assistant generates or edits an image in chat
- [ ] They are available wherever the composer generates images
- [ ] They are available anywhere else in the product that lists image models, so no surface is left behind on an older list

**Public surfaces**

- [ ] The models can be selected by name through the public API
- [ ] The models can be selected through the MCP tools
- [ ] The models can be selected through the CLI
- [ ] The names used across the API, MCP and CLI are identical to each other, so a script written against one works against the others
- [ ] The public documentation lists the new models alongside the existing ones, including which do generation and which do editing

**Naming and description**

- [ ] Each model has a name a user can tell apart at a glance, distinguishing the fast variant from the high fidelity one
- [ ] Each has a one-line description saying when to reach for it, not a restatement of its name
- [ ] **Sunburst's longer generation time is stated before the user commits to it**, not discovered while waiting
- [ ] Names and descriptions are translated across all supported locales in the same change

**Cost and limits**

- [ ] Each model has its credit cost set, and the cost is shown to the user before they generate, as it is for existing models
- [ ] Usage is metered and deducted the same way as existing image models, so the balance and usage figures stay correct
- [ ] A user without enough credits is told before generating, using the existing treatment

**Behaving properly when things go wrong**

- [ ] A generation that fails on either model surfaces a clear message and does not consume credits for work that produced nothing
- [ ] A request for an aspect ratio or option a model does not support is prevented rather than failing after the user waits
- [ ] Existing models are unaffected, and anyone mid-work on one is not disturbed

---

### UI copy

**Model names**

> GPT Image 2.5 Flare
> GPT Image 2.5 Sunburst

**Model descriptions in the picker**

> **Flare:** Fast, high quality images with natural lighting and rich textures. Handles complex layouts and can give you a transparent background. Good for most things.
> **Sunburst:** Extra detail for premium work, at the cost of a longer wait. Reach for it when the fine detail matters.

**Generation time**

> Sunburst carries the existing longer-generation indicator so the wait is expected. Copy follows whatever the existing models use for this, rather than introducing a new pattern.

**Editing**

> No separate name is shown for the edit variants. A user picks Flare or Sunburst and the right one is used depending on whether they are generating or editing, which is how the existing models already behave.

**Error state**

> Follows the existing image generation error treatment. No new copy.

---

### Mock-ups

N/A. The models appear in the existing pickers and lists using the existing treatment. No new interface.

---

### Impact on existing data

None. No stored data changes. Images already generated with other models are unaffected, and a user's saved model preference continues to work.

---

### Impact on other products

- **Public API, MCP and CLI:** additive. Existing integrations keep working, and the new names become selectable. Their documentation needs the new entries.
- **Mobile app:** AI image generation is web only, so no mobile work. Worth confirming nothing in the app lists image models for display purposes.
- **Chrome extension:** no impact.
- **Billing and usage:** new credit costs enter the usage figures. Worth a note to support before release so they can answer questions about the new models' cost.

---

### Open before build

1. **Credit cost for each model.** Sunburst is the premium, slower option, so it presumably costs more than Flare, but the actual numbers are a product and pricing decision rather than something to infer.
2. **Does Flare become the default?** The provider positions it as the default for most applications. Whether ContentStudio's default image model changes is a separate decision from making it available, and changing it affects everyone who has never picked a model.
3. **How transparent backgrounds are offered.** It is a genuine Flare capability. Whether it appears as an option in the picker, is inferred from the prompt, or is left for a later change needs deciding, because it is the one capability here that is not just "another model in the list".
4. **Whether any existing model is retired** as a result. Adding four models to a list that already has several makes the picker longer, and it is worth asking whether anything is now redundant.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
