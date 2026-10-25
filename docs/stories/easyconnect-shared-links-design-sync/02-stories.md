# [FE] Redesign the EasyConnect links panel to match Manage Shared Links in Planner and Analytics

## Description

ContentStudio has three places where a user manages public links they have shared: Planner shared links, Analytics shared links and EasyConnect links. The Planner and Analytics panels already share one design, the "Manage Shared Links" panel. The EasyConnect panel still uses the old design, so the same kind of screen looks and behaves differently depending on where the user opens it.

This story brings the EasyConnect links panel in line with the Manage Shared Links design so all three look and work the same. It is a visual and layout sync only. What an EasyConnect link does, and the data behind it, does not change.

Reference screenshots are in the Mock-ups section below.

## Workflow

1. The user goes to Settings > Social Accounts and opens their EasyConnect links.
2. The panel opens with the same header as Planner and Analytics: a title, a count badge, and a one-line subtitle saying how many links are active.
3. A search box sits under the header. The user types part of a link name and the list filters as they type.
4. Each link shows as a card in the same layout as the Planner and Analytics cards: link name with a status badge, an on/off switch, the full link in a row with a Copy button, and a row of small detail chips.
5. The user clicks Copy and gets the same "Link copied" confirmation they get in Planner and Analytics.
6. The user hovers a card and sees Edit and Delete, in the same place and style as Planner and Analytics.
7. The user can still see which social accounts each EasyConnect link has connected, by expanding the card.

## Acceptance criteria

**Header**

- [ ] Title reads "EasyConnect Links" with a count badge showing the total number of links, matching the Planner and Analytics header style
- [ ] Subtitle reads: "Links your clients use to connect their social accounts to this workspace. {count} active."
- [ ] Close (X) button sits top right, same as Planner and Analytics

**Search**

- [ ] Search box with placeholder "Search links" under the header
- [ ] Filters the list by link name as the user types
- [ ] When nothing matches, show "No links match your search."

**Link card** (same structure and spacing as the Planner and Analytics cards)

- [ ] Link name in bold, followed by a status badge: "Active" (green dot) when on, "Disabled" when off
- [ ] On/off switch on the right of the name row, same switch component as Planner and Analytics. Tooltip: "Disable Link" when on, "Enable Link" when off
- [ ] Link row shows a link icon, the full EasyConnect URL (truncated with an ellipsis if too long, full URL in a tooltip on hover) and a "Copy" button with a copy icon
- [ ] Copy shows the toast "Link copied to clipboard"
- [ ] Detail chips row shows: created date (calendar icon, for example "Oct 12, 2026"), a "Password" chip with a lock icon when the link is password protected (this replaces the current lock icon next to the name), and a chip with the number of connected accounts (for example "3 accounts")
- [ ] Edit (pencil) and Delete (bin) actions appear on hover on desktop and are always visible on mobile, same as Planner and Analytics. Tooltips: "Edit" and "Delete"
- [ ] Edit opens the existing EasyConnect edit flow, unchanged
- [ ] Delete keeps the existing confirmation, with the title "Remove EasyConnect Link" and the message "Are you sure you want to remove this EasyConnect link? Clients who have it will no longer be able to connect accounts with it."
- [ ] A disabled link keeps the card but shows the "Disabled" badge and a muted look, same as a disabled link in Planner and Analytics

**Connected accounts**

- [ ] The connected accounts for each link are still available. They show in an expandable section inside the card, styled to fit the new card instead of the old chevron row. Expand label: "Show connected accounts", collapse label: "Hide connected accounts"

**States**

- [ ] Loading: skeleton cards in the same shape as the new card, same as Planner and Analytics
- [ ] Empty (no links yet): "No EasyConnect links yet" with the secondary line "Create a link to let your clients connect their social accounts."
- [ ] Error loading: "We couldn't load your EasyConnect links. Please try again."

**Consistency**

- [ ] Placed side by side, the EasyConnect panel matches the Planner and Analytics panels in header, search, card layout, spacing, typography, colors, icons and badge styles
- [ ] Brand colors come from the theme (primary-cs tokens) so white-label workspaces get their own color, instead of the hardcoded pink link color and black switch in the current EasyConnect panel
- [ ] Reuse the shared card and panel components that Planner and Analytics already use wherever possible, rather than building a third copy of the same design
- [ ] No change to how EasyConnect links are created, what they connect, or their stored data

## Mock-ups

The Planner and Analytics Manage Shared Links panels are the target design. The EasyConnect panel must be updated to match them.

**1. Planner: Manage Shared Links (target design)**

![Planner Manage Shared Links panel, target design](https://api.helpin.ai/api/pm/attachments/a290099c-ada1-4020-bff5-97bd43ad3aa0/content)

**2. Analytics: Manage Shared Links (target design)**

![Analytics Manage Shared Links panel, target design](https://api.helpin.ai/api/pm/attachments/d3fa5c48-7845-439a-b2e7-2fd824aa9558/content)

**3. EasyConnect: current design (to be updated)**

![EasyConnect links panel, current outdated design](https://api.helpin.ai/api/pm/attachments/820f4e64-e2ed-4a64-94f0-2857b66ed578/content)

## Impact on existing data

None. Visual and layout change only. Existing EasyConnect links, their settings, passwords and connected accounts are untouched.

## Impact on other products

- Web app only. The Flutter app, Chrome extension and marketing website are not affected.
- Planner and Analytics shared links panels must keep working exactly as they do now if their shared components are reused here.

## Dependencies

None.

## Global quality checklist

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (any new or changed API is reflected in the public API, CLI, MCP server and automation apps, N/A when nothing API-facing changes)
