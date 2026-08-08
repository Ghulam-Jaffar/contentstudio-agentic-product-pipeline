# Research — Social Inbox message list density and reply composer

Raised by the CTO with a screenshot of the Inbox, annotated: *"Can we improve the UI and UX for the social inbox. There is a lot of empty space being consumed in messages list."* The screenshot additionally highlights the reply composer in the conversation pane.

Two distinct complaints in one screenshot:

1. **The message list wastes vertical space**, so far fewer conversations fit on screen than should.
2. **The reply composer is cramped**, with the reply text wrapping into a narrow box while horizontal space sits unused beside it.

## Backlog check

No existing story covers Inbox visual density or the composer. The nearest neighbours are all functional rather than visual: `inbox-stability-hardening`, `inbox-module-architecture-sync-findings`, `inbox-webhook-events`, `meta-inbox-webhook-coverage-review`, and the network-integration features. `post-previews-inbox-analytics` touches the Inbox conversation pane but only to standardise the *post preview* component, not the list or the composer. This is new work.

## Fault 1: the message list row has a hard 90px floor

`contentstudio-frontend/src/modules/inbox-revamp/components/InboxListItem.vue`, line 3:

```
class="px-[13px] py-4 bottom-border-only cursor-pointer select-none flex gap-[14px] group min-h-[90px]"
```

Row anatomy, top to bottom:

- `py-4` — 16px padding top and bottom, 32px before any content
- `min-h-[90px]` — a hard floor regardless of content
- The content column is `flex flex-col gap-3` — a 12px gap between the name/preview block and the row of status badges beneath it
- Within that block, `gap-[6px]` between the name line and the preview line
- The avatar is `w-8 h-8` (32px) with a 20px platform badge overlapping its corner

So a row carrying a 32px avatar, one name line and one truncated preview line has roughly 40px of real content and occupies at least 90px. The screenshot shows about seven conversations in a full-height viewport.

Two more fixed dimensions constrain the layout rather than adapt to it:

- `max-w-[150px]` on the name, so names truncate early even when the column is wide
- `w-58` on the content column

The skeleton matches the floor: `InboxListing.vue` line 324 renders placeholder rows at `height="90px"`, so any change to row height has to move in step or the loading state will jump.

### Theming defects in the same component

Hardcoded colours that cannot follow a white-label customer's brand:

- `bg-[#02B2FF1A]` — the active/selected row (line 6)
- `hover:bg-[#FBFBFB]` — row hover (line 9)
- `bg-[#FD67271A]`, `bg-[#9747FF1A]`, `bg-[#FFF7DE]` — tag colours (lines 227, 232, 237)

## Fault 2: the reply composer

`contentstudio-frontend/src/modules/inbox-revamp/components/MessageComposer.vue`, 985 lines.

- The input is a `<textarea>` with `:rows="miniBox ? 1 : 2"` and `resize-none` (lines 33 to 44, and a second textarea at lines 67 to 77 for the note variant). It grows via an `autoResize` handler rather than by CSS.
- There is no visible maximum height, so a long reply can grow the composer without a defined ceiling, and no minimum beyond two rows, so it starts small in a pane that has room.
- The action row carries five icon buttons plus a split send control, all inline with the input, which is what squeezes the text column in the screenshot.

### A latent bug worth fixing while the composer is open

The component reaches for its textarea with a **global document query** in four places (lines 647, 657, 711, 892):

```
const textarea = document.querySelector('textarea')
```

`document.querySelector` returns the *first* textarea in the document, not this component's. The component renders two textareas (reply and note), and other textareas exist elsewhere on the page. So emoji insertion, cursor positioning and height reset can operate on the wrong element. This is a real defect independent of the redesign, and the redesign is a sensible moment to replace it with a template ref.

### Theming defects in the composer

Extensive hardcoded colour, including an inline gradient:

- `border-[#FFE48B]`, `bg-[#FFFCF3]` — note mode (line 5)
- `border-[#F953C6]/50`, `border-[#93B5E1]` — auto-reply draft states (lines 8, 9)
- `bg-[#E6F0FC]`, `text-[#75797C]` — toolbar icon buttons (lines 116 to 130)
- `bg-[#FF00B1]`, `hover:bg-[#FF68D1]`, and `style="background: linear-gradient(202deg, #FF68D1 14.46%, #FF00B1 96.91%)"` — the AI action (lines 204, 150)
- `color="#3A4562E5"` on repeated `Check` icons (lines 163 to 184)

## What the design needs to decide

**Message list**
- A target row height, and what that yields in conversations visible at the supported viewport heights.
- Which elements earn their place in a row: avatar, platform badge, name, timestamp, preview, unread indicator, conversation-type icon, assignee avatar, tags, checkbox. The screenshot shows a second row used almost entirely by two small badges.
- Whether the status badges belong on their own row, inline with the timestamp, or on hover.
- Whether density should be selectable (comfortable vs compact) or a single tuned default.
- How the selected and unread states read once rows are tighter, since the current treatment relies on a full-row background tint.

**Composer**
- Resting height, maximum height, and scroll behaviour past the maximum.
- How the text column and the action row share horizontal space, so a reply of a few lines does not wrap into a narrow column.
- Which actions stay visible and which collapse behind an overflow at narrow widths.
- How note mode and auto-reply draft mode read without three different hardcoded border colours.
- Where attachment chips sit relative to the input.

## Files involved

- `contentstudio-frontend/src/modules/inbox-revamp/components/InboxListItem.vue`
- `contentstudio-frontend/src/modules/inbox-revamp/components/InboxListing.vue`
- `contentstudio-frontend/src/modules/inbox-revamp/components/MessageComposer.vue`
- `contentstudio-frontend/src/modules/inbox-revamp/components/ConversationHeader.vue`, `ChatView.vue`, `CommentBlock.vue`, `PostView.vue` — the surrounding conversation pane
- `contentstudio-frontend/src/modules/inbox-revamp/views/InboxView.vue`

## Mobile

The mobile apps have their own inbox screens and are unaffected. This is the web Inbox.
