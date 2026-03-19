# Available UI Components

> **Last updated:** 2026-03-18
>
> Single source of truth for what UI components exist in the ContentStudio design system. The `/feature` and `/story` pipelines read this file before writing stories to ensure they only reference available components.
>
> **To update:** Edit this file when components are added/removed from `@contentstudio/ui` or `src/components/UI/`. The package is `@contentstudio/ui` (currently v0.2.21).

---

## @contentstudio/ui — Design System Package

These are the **preferred** components. They are theme-aware, white-label compatible, and should be used over the legacy `Cst*` equivalents when both exist.

### Button
| Component | Status | Notes |
|---|---|---|
| `Button` | Available | Primary button with variants (primary, secondary, ghost, etc.). Use props for variants — don't override with Tailwind. |
| `SplitButton` | Available | Button with dropdown for secondary actions |

### Dropdown / Select
| Component | Status | Notes |
|---|---|---|
| `Dropdown` | Available | Dropdown menu container |
| `DropdownItem` | Available | Individual item inside a Dropdown |

### List
| Component | Status | Notes |
|---|---|---|
| `ListItem` | Available | Standard list item component |

### Checkbox
| Component | Status | Notes |
|---|---|---|
| `Checkbox` | Available | Standard checkbox |

### Badge
| Component | Status | Notes |
|---|---|---|
| `Badge` | Available | Status/count badge |

### Tabs / Segmented Control
| Component | Status | Notes |
|---|---|---|
| `Tabs` | Available | Tab container |
| `SegmentedControl` | Available | Segmented toggle control (use instead of custom tab-style toggles) |

### Radio
| Component | Status | Notes |
|---|---|---|
| `Radio` | Available | Radio button input |

### Input
| Component | Status | Notes |
|---|---|---|
| `TextInput` | Available | Text input field |
| `SearchInput` | Available | Search input with icon |
| `Textarea` | Available | Multi-line text input |

### Avatar
| Component | Status | Notes |
|---|---|---|
| `Avatar` | Available | User/account avatar display |

### Progress / Loader
| Component | Status | Notes |
|---|---|---|
| `Progress` | Available | Progress bar/indicator |
| `Loader` | Available | Loading spinner/indicator |

### Modal / Dialog
| Component | Status | Notes |
|---|---|---|
| `Modal` | Available | Modal dialog. Also has `ModalPlugin` (programmatic open/close) and `ModalDirective`. |
| `Dialog` | Available | Dialog component (alternative to Modal) |

### Collapsible
| Component | Status | Notes |
|---|---|---|
| `Collapsible` | Available | Expandable/collapsible content section |

### Other
| Component | Status | Notes |
|---|---|---|
| `Icon` | Available | Renders design system icons |
| `Switch` | Available | Toggle switch for boolean settings |
| `Alert` | Available | Inline alert component |
| `Breadcrumbs` | Available | Navigation breadcrumbs |
| `Pagination` | Available | Page navigation |
| `ActionIcon` | Available | Icon-only clickable action (like an icon button) |
| `ThemeProvider` | Available | Wraps content with theme CSS variables (infrastructure, not used in stories) |

---

## Legacy Cst* Components (src/components/UI/)

Auto-registered via `unplugin-vue-components` — no imports needed. **Use `@contentstudio/ui` equivalents when available.** These exist for backward compatibility but new stories should prefer the design system package.

| Component | Category | @contentstudio/ui equivalent |
|---|---|---|
| `CstButton` | Button | `Button` |
| `CstInputFields` | Input | `TextInput` |
| `CstFloatingLabelInput` | Input | — (no equivalent) |
| `CstTextArea` | Text area | `Textarea` |
| `CstDropdown` | Dropdown | `Dropdown` |
| `CstDropdownItem` | Dropdown | `DropdownItem` |
| `CstTagsDropdown` | Dropdown | — (no equivalent, use for tag selection) |
| `CstLabelsDropdown` | Dropdown | — (no equivalent, use for label selection) |
| `CstSimpleCheckbox` | Checkbox | `Checkbox` |
| `CstCardCheckbox` | Checkbox | — (no equivalent, card-style) |
| `CstAccountCheckBox` | Checkbox | — (no equivalent, social account selection) |
| `CstRadio` | Radio | `Radio` |
| `CstSwitch` | Switch | `Switch` |
| `CstIconSwitch` | Switch | — (no equivalent, icon toggle) |
| `CstAlert` | Alert | `Alert` |
| `CstBanner` | Alert | — (full-width banner) |
| `CstToast` | Feedback | — (toast notifications) |
| `CstTab` / `CstTabs` | Tabs | `Tabs` |
| `CstCollapsible` | Collapsible | `Collapsible` |
| `CstDrawer` | Layout | — (slide-out drawer) |
| `CstSidebar` | Layout | — (sidebar panel) |
| `CstPopup` | Popup | — (popover/tooltip) |
| `CstConfirmationPopup` | Dialog | `Dialog` / `Modal` |
| `CstEmojiPicker` | Emoji | — (emoji picker) |
| `AccountDropdown` | Dropdown | — (social account selector) |
| `VideoLightBox` | Media | — (video lightbox) |

---

## What's NOT Available Yet

Components that stories frequently need but don't exist. Helps prioritize design system work.

- _Pill / Chip_ — no dedicated pill/chip component (use `Badge` for status indicators, or Tailwind for simple pills)
- _Tooltip_ — no standalone tooltip component in the library (use `CstPopup` or Tailwind-based approach)

---

## Rules for Story Authors

1. **Prefer `@contentstudio/ui` components** over legacy `Cst*` equivalents.
2. **Reference by name.** When specifying UI in FE stories, name the exact component (e.g., "Use the `SegmentedControl` component" not "add a toggle").
3. **Flag gaps.** If a story needs a component not listed here, explicitly note: _"Requires new component: [description]. Not currently in the UI library — needs a [Design] story or library update first."_
4. **Don't invent components.** Never reference a component name that isn't in this catalog without flagging it as new/needed.
5. **Check the "Not Available" section** before assuming a component exists.
