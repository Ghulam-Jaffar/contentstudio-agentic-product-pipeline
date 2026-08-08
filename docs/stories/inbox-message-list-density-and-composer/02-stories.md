# Story: Social Inbox message list density and reply composer

Raised by the CTO from the Inbox, annotated *"There is a lot of empty space being consumed in messages list"*, with the reply composer highlighted alongside.

One design story, as requested. The companion `[FE]` build story is the obvious follow-on and is not written here.

---

## [Design] Rework the Social Inbox message list density and reply composer

### Description

The Social Inbox message list gives each conversation a hard 90px floor for what is usually a name, one line of preview text and two small badges. On a full-height screen that is roughly seven conversations, so anyone triaging a busy inbox scrolls constantly to see work that would comfortably fit. The reply composer has the opposite problem: it opens as a two-row box and the reply wraps into a narrow column while horizontal space sits unused beside the action buttons.

This story is the design pass. It decides how dense the list should be, what each row actually needs to show, and how the composer should use the space it has, then hands the frontend team a spec precise enough to build from.

It is a visual and interaction redesign only. Which conversations appear, how they are filtered, sorted, assigned or replied to does not change.

### Workflow

N/A. Design deliverable.

### Acceptance criteria

**Message list**

- [ ] A target row height is specified, along with the number of conversations that become visible at the viewport heights the app supports. The target is a measurable improvement on the current 90px floor and the improvement is stated as a number, not as "tighter".
- [ ] Every element a row can display is accounted for and given a place or explicitly dropped: selection checkbox, avatar, platform badge, sender name, timestamp, message preview, unread indicator, conversation-type icon, assignee avatar, and tags.
- [ ] The design states which of those elements are always visible, which appear only on hover or selection, and which move into the row's overflow, so the second row currently occupied by two small badges is justified or removed.
- [ ] Truncation rules are specified for the sender name and the message preview, based on available width rather than a fixed pixel cap, so a wide list column shows more text than a narrow one.
- [ ] The unread state is specified and remains obvious at the new density, without relying solely on a full-row background tint.
- [ ] The selected state is specified and remains distinguishable from both unread and hover at the new density.
- [ ] The hover state is specified.
- [ ] Multi-select is specified: how the checkbox behaves at the new density, and how a row reads when selected for a bulk action versus opened for reading.
- [ ] Tag display is specified, including what happens when a conversation carries more tags than fit.
- [ ] The loading skeleton is specified at the new row height so the list does not shift when real rows replace placeholders.
- [ ] The empty state and the end-of-list state for the message list are specified.
- [ ] A recommendation is made on whether density should be user-selectable (for example comfortable and compact) or a single tuned default, with a reason. A single well-tuned default is an acceptable and preferred answer if the designer judges it sufficient.

**Reply composer**

- [ ] The composer's resting height, maximum height and scroll-past-maximum behaviour are specified.
- [ ] The horizontal relationship between the text input and the action row is specified, so a multi-line reply uses the full width of the pane rather than wrapping into a narrow column beside the buttons.
- [ ] Every action in the composer is accounted for: emoji, attachment, saved replies, AI assist, send, and the send split control. Each is either kept, moved, or collapsed behind an overflow.
- [ ] Behaviour at narrow pane widths is specified, including which actions collapse first.
- [ ] Note mode is specified as a distinct visual state, replacing the current one-off border and background treatment.
- [ ] The auto-reply draft state is specified, replacing the two different dashed-border treatments currently used for the compact and full composer.
- [ ] The compact composer variant is specified alongside the full one, since both exist and currently differ in more than size.
- [ ] Attachment chip placement is specified, including behaviour with several attachments and with a long filename.
- [ ] The sending and disabled states are specified, including how the composer reads while a reply is in flight.

**Both**

- [ ] All colours are specified as theme tokens. The design explicitly replaces the hardcoded values currently in use, including the selected-row tint, the row hover, the tag colours, the note-mode border and background, the auto-reply draft borders, the toolbar icon backgrounds and the AI action's gradient.
- [ ] The design is verified against a white-label palette, not only the default brand colour.
- [ ] Behaviour is specified at every supported width, including when the folder sidebar is collapsed and when the conversation detail pane is open beside the list.
- [ ] The design names the design-library components to reuse and flags anything unavailable as a component gap.
- [ ] Redlines cover spacing, type scale and icon sizing at the new density, so the build does not need a second design round.
- [ ] The design does not change which conversations appear, how they are filtered or sorted, or how replying works. Anything that looks functionally wrong is raised as a separate finding rather than redesigned here.

### Mock-ups

This story produces them.

### Impact on existing data

None. Presentation only.

### Impact on other products

- Web app only. The mobile apps have their own inbox screens and are unaffected.
- White-label domains render the Inbox, and this is the story that makes the list and composer theme correctly rather than shipping fixed colours.

### Dependencies

- Related to **[FE] Adapt the shared post previews for published posts and adopt them in Inbox**, which changes the post preview inside the same conversation pane. The two touch neighbouring areas, so the designer should be aware of that work rather than designing the pane twice.
- The companion `[FE]` build story is not yet written and should follow this one.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — names, previews and action labels must tolerate longer translated strings at the new density
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Where the space goes today.** `contentstudio-frontend/src/modules/inbox-revamp/components/InboxListItem.vue` line 3 sets the row to `px-[13px] py-4 ... min-h-[90px]`. Inside it, the content column is `flex flex-col gap-3` (12px between the name/preview block and the badge row) and the name/preview block is `gap-[6px]`. The avatar is `w-8 h-8` with a 20px platform badge overlapping it. Roughly 40px of content in a 90px row.

**Fixed dimensions that should probably become fluid.** `max-w-[150px]` truncates the sender name regardless of column width, and the content column is `w-58`.

**The skeleton must move with the row.** `InboxListing.vue` line 324 renders placeholders at `height="90px"`. If the row height changes and this does not, the list will jump when data lands.

**Hardcoded colours to replace with tokens.** In `InboxListItem.vue`: `bg-[#02B2FF1A]` (selected, line 6), `hover:bg-[#FBFBFB]` (line 9), and the tag colours `bg-[#FD67271A]`, `bg-[#9747FF1A]`, `bg-[#FFF7DE]` (lines 227, 232, 237). In `MessageComposer.vue`: `border-[#FFE48B]` and `bg-[#FFFCF3]` (note mode, line 5), `border-[#F953C6]/50` and `border-[#93B5E1]` (auto-reply draft, lines 8 and 9), `bg-[#E6F0FC]` and `text-[#75797C]` (toolbar icons, lines 116 to 130), `bg-[#FF00B1]` with `hover:bg-[#FF68D1]` (line 204) and the inline `linear-gradient(202deg, #FF68D1 14.46%, #FF00B1 96.91%)` (line 150), plus `color="#3A4562E5"` on the repeated Check icons (lines 163 to 184).

**Composer sizing today.** `MessageComposer.vue` uses a `<textarea>` with `:rows="miniBox ? 1 : 2"` and `resize-none` (lines 33 to 44), with a second textarea for note mode at lines 67 to 77. Growth is handled by an `autoResize` handler rather than CSS, and there is no defined maximum height.

**A latent bug for the build story, not for design.** The composer resolves its textarea with `document.querySelector('textarea')` in four places (lines 647, 657, 711, 892). That returns the first textarea in the *document*, not the component's, and the component renders two. Emoji insertion, cursor positioning and height reset can therefore act on the wrong element. Worth fixing with a template ref while the file is open, and worth the designer knowing that emoji-insert behaviour may currently be unreliable if they test it.

**Scale note.** `MessageComposer.vue` is 985 lines and `InboxListItem.vue` is 508. Per the repo's own guidance, the build story should stay scoped to the redesign rather than becoming a full decomposition pass.
