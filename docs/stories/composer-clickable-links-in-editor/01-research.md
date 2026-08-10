# Research — Clickable links in the composer editor

User request from Camilo via Frill, 7 days ago. Tagged **#Improvement**, **#Integrations**, status **Planned**.

> Make links in the copy clickable directly from the text editor so they open in a new tab.
>
> It's useful because you can verify the links you're going to post much faster and more efficiently before scheduling them.

Source: https://contentstudio.frill.co/dashboard/inbox/ideas?idea=make-links-in-the-copy-clickable-directly-from-the-text-editor-so-they-open-in-a-new-tab

## Backlog check

Nothing covers this. `consistent-post-previews-composer-planner` and `composer-v2-refactor` mention links only in passing, in the context of link *previews* rather than opening them.

## Current state

The composer editor is **TipTap 3**, wrapped in `contentstudio-frontend/src/components/UI/TextEditor/CstSocialEditor.vue`. Used by `modules/composer/components/EditorBox/EditorBox.vue` (main composer) and `EditorBox/LiteEditorBox.vue`, so a fix in the wrapper reaches both.

### Links are already found and already look like links

`components/UI/TextEditor/editorHighlights.ts` applies a ProseMirror decoration:

```
const PATTERNS = [
  { re: URL_RE, className: 'tt-link' },
  { re: HASHTAG_RE, className: 'tt-hashtag' },
]
```

with `Decoration.inline(pos + range.from, pos + range.to, { class: range.className })`. Styled in `CstSocialEditor.vue`:

```
.cst-social-editor-content .tt-link { color: #216fdb; }
```

Its own comment explains the history: the old `@contentstudio/social-editor` bundle applied this via linkifyjs decorations, the TipTap migration dropped it, and these decorations restore *"develop's exact treatment: links are blue text"*.

So a URL already renders blue. It just is not interactive, because a decoration is a styling span, not an anchor.

### The Link extension is deliberately off

`CstSocialEditor.vue` configures StarterKit with formatting stripped, including:

```
StarterKit.configure({ ..., link: false, ... })
```

The comment above it: *"Formatting stripped to plain text — social posts have no rich formatting."*

**That decision should stand.** Enabling the Link extension would introduce anchor marks into the document, which changes what the editor serialises and risks the plain-text contract the whole composer depends on. The right approach is click handling over the decoration that already exists, not a new mark.

### There is no click handling today

`editorProps` implements `handleKeyDown`, `handlePaste`, `handleDrop` and `handleDOMEvents` (drag events). There is no `handleClick`. Nothing responds to clicking a URL.

## The constraint that shapes the whole thing

**This is an editable field.** A plain click has to keep placing the caret. If clicking a URL navigated instead, a user could not put their cursor inside or beside a link to edit it, which is a worse problem than the one being solved.

So "clickable" has to mean a deliberate gesture. The two conventions, both worth having:

- **Modifier + click** (Cmd on macOS, Ctrl on Windows) — what VS Code, Notion and Linear use. Precise, no new UI, but undiscoverable on its own.
- **A hover affordance** — a small popover or inline action on the link, showing the URL with an open control. Discoverable, and it can show the *whole* URL, which matters when a long link is truncated in a narrow editor.

## The finding that matters most for the stated goal

Camilo's reason is verification: *"verify the links you're going to post"*.

The composer transforms URLs before posting. Both features exist today:

- **Shorten** — `locales/en/composer.json` has `shorten` and `invalid_url` ("Please enter a valid URL to shorten.")
- **Add UTM** — `EditorBox/EditorOptions/UtmDropdown.vue`, with `add_utm` copy in two places

So the URL typed in the editor is not necessarily the URL that ships. If a user opens the typed link and it works, they have verified the typed link, not the posted one. With UTM appended that is usually harmless. With a shortener in play the posted link is a different URL entirely.

That does not block the request, but the story has to take a position on it rather than leave the user with false confidence. Options: open the URL as typed and say so; open the final URL where it can be resolved; or show both in the affordance.

## Security considerations

The editor contains whatever a user typed, so the URL is untrusted input.

- Only `http` and `https` should ever open. `javascript:`, `data:`, `file:` and similar must be refused. `URL_RE` should be checked for what it currently admits.
- Opening in a new tab needs `noopener` so the opened page cannot reach back into the app via `window.opener`.

## Adjacent defect worth fixing while in the file

Both decoration styles are hardcoded hex, so neither follows a white-label customer's brand colour:

```
.cst-social-editor-content .tt-link { color: #216fdb; }
.cst-social-editor-content .tt-hashtag { background-color: rgba(22, 120, 247, 0.2); }
```

The repo's rules require brandable colour to come from semantic tokens.

## Files involved

- `contentstudio-frontend/src/components/UI/TextEditor/CstSocialEditor.vue` — the editor wrapper, `editorProps`, the decoration styles
- `contentstudio-frontend/src/components/UI/TextEditor/editorHighlights.ts` — where the `tt-link` ranges are produced
- `contentstudio-frontend/src/components/UI/TextEditor/linkDetect.ts` — `extractLinks`, `URL_RE`, and the existing link-preview reporting
- `contentstudio-frontend/src/modules/composer/components/EditorBox/EditorBox.vue`, `LiteEditorBox.vue` — the two consuming surfaces
- `contentstudio-frontend/src/locales/*/composer.json` — new strings

## Mobile

No mobile story. The composer exists in the apps, but hover and modifier keys do not, so the interaction does not translate. If this is wanted on mobile it needs its own gesture (long press, or a tappable chip) and should be raised separately once the web behaviour is settled.
