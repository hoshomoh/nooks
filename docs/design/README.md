# Design

The design canvas is the source of truth for every screen. It lives in Claude Design:

- Project: `e1148b43-1103-4b0e-9507-20c38d1db5b8`
- <https://claude.ai/design/p/e1148b43-1103-4b0e-9507-20c38d1db5b8>

## Artboards

| File | Covers |
| --- | --- |
| `Nook-Foundations.dc.html` | Palette, type, spacing, components, page furniture, overlays, words, identity, states |
| `Nook-List-View.dc.html` | List, Today, Upcoming, calendar, empty states, side sheet, full-screen note, list menu |
| `Nook-Auth-Onboarding.dc.html` | First run, sign in, temporary password, join approved, reset flow, settings, tokens, members, groups, public list, about |
| `Nook-Sharing-Team.dc.html` | Share dialog, specific people, presence, activity panel |
| `Nook-Public-Access.dc.html` | Public list with no account, sign-in prompt, ask to join, request sent |
| `Nook-Print.dc.html` | A4 print sheets — single column and two-up |

`support.js` and `doc-page.js` in that project are the canvas runtime (`<sc-for>`, `<sc-if>`, `DCLogic`,
and the A4 pagination element). They are not app code and nothing here ports them.

## Pulling a file locally

Cached copies of two artboards sit in this directory. To refresh one, or pull any of the others:

```
DesignSync get_file projectId=e1148b43-1103-4b0e-9507-20c38d1db5b8 path=Nook-List-View.dc.html
```

Read an artboard bottom-up: the `<script data-dc-script>` block at the end holds the palette and all the
screen data, and the markup above it is the layout that consumes it.

## Rules that came out of the canvas

These are decisions, not suggestions. They are enforced in `apps/web/src/index.css` and in review.

- Four meaning colours, one job each: accent/shared, done/presence, offline, overdue/error. Nothing is
  coloured for decoration.
- A list row is 44px minimum, always: 6px inner padding, 14px between checkbox and label, metadata
  right-aligned, truncation in the label. A row with five pieces of metadata is the height of a bare one.
- Radii mean things: 4 checkbox, 6 row, 7 control, 10 surface.
- Three rule weights mean different things: 2px separates sections of a page, 1px `--nooks-line`
  separates blocks inside a section, 1px `--nooks-hair` separates rows.
- 16px is the floor for anything a Member wrote. Print steps up, never down.
- Empty states are dashed, left-aligned, never centred, and carry no illustration.
- Dialogs are 560px with a bordered footer: consequence text left, cancel then confirm right.
- Focus is a 2px accent ring at 2px offset. Hit areas are 44px even when the control is 17px.
