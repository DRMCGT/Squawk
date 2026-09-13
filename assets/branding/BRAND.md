# Squawk Brand Guidelines

This document defines the official visual identity, color system, typography recommendations, logo assets, and usage guidelines for **Squawk**.

---

## 1. Brand Concept & Story

The name **Squawk** comes from aviation: an aircraft's transponder broadcasts a 4-digit squawk code to Air Traffic Control to communicate identity and status (e.g. Mode C altitude reporting or emergency status).

In developer QA:
- **The Tester** acts as the pilot signalling a status change when a bug is spotted.
- **Squawk** acts as the transponder, broadcasting precise telemetry (screenshots, logs, DOM elements, notes).
- **The AI Coding Agent** acts as Air Traffic Control / Maintenance ground crew, receiving clean context to execute an immediate fix.

### Core Impression
- **Developer-focused**: Clean, precise, dark-mode first, technical.
- **Fast & Action-Oriented**: Energetic electric cyan signal waves with amber alert accents.
- **Reliable & Approachable**: High-contrast, legible at 16px favicons or terminal ASCII banners.

---

## 2. Color System

Squawk uses a cohesive palette crafted for high visibility across GitHub light/dark themes, terminal UIs, and web widgets.

| Color Name | Hex Code | Purpose / Usage |
| :--- | :--- | :--- |
| **Radar Blue** (Primary) | `#0284c7` | Core brand identity, primary wing vector fill |
| **Electric Signal** (Accent) | `#38bdf8` | Outer signal arcs, glowing highlights, terminal cyan |
| **Transponder Amber** (Alert) | `#f59e0b` | Bug beacon dot, active capture state, warnings |
| **Amber Flame** (Highlight) | `#fbbf24` | High-contrast amber highlight, glowing core |
| **Deep Space** (Dark BG) | `#0b0f19` | Dark mode background, terminal background |
| **Slate Dark** (Panel BG) | `#1e293b` | Dark card backgrounds, widget panel UI |
| **Slate Light** (Light Text) | `#f8fafc` | Primary text on dark backgrounds |
| **Steel Dark** (Light BG Text)| `#0f172a` | Primary text on light backgrounds |

---

## 3. Typography Recommendations

Squawk uses clean, freely available open-source fonts for high legibility across platforms:

- **Headings & Logos**: Inter, Outfit, or standard system sans-serif (`system-ui`, `-apple-system`, `BlinkMacSystemFont`, `'Segoe UI'`, `Roboto`).
- **Code & Terminal UI**: JetBrains Mono, Fira Code, or system monospace (`ui-monospace`, `SFMono-Regular`, `Menlo`, `Consolas`).

---

## 4. Logo System & Assets

All official vector logo assets live under `assets/branding/`:

- [`squawk-logo.svg`](squawk-logo.svg) — Primary horizontal logo (Mark + Wordmark + Subtitle)
- [`squawk-mark.svg`](squawk-mark.svg) — Compact icon mark (Signal Arcs + Wing + Beacon)
- [`squawk-logo-dark.svg`](squawk-logo-dark.svg) — Optimized for dark backgrounds (e.g., GitHub dark mode)
- [`squawk-logo-light.svg`](squawk-logo-light.svg) — Optimized for light backgrounds (e.g., GitHub light mode)
- [`squawk-mark-monochrome.svg`](squawk-mark-monochrome.svg) — Single-color version (`currentColor` stroke/fill)
- [`favicon.svg`](favicon.svg) — Standalone icon in dark card container, crisp down to 16x16
- [`social-preview.svg`](social-preview.svg) — 1280x640 GitHub social preview card

### Minimum Sizes & Clear Space
- **Primary Horizontal Logo**: Minimum width `140px` (or `35px` height).
- **Compact Mark**: Minimum size `16px × 16px`.
- **Clear Space**: Maintain a minimum clear padding equal to half the height of the transponder chevron around all edges of the mark.

---

## 5. Terminal ASCII Treatment

For terminal applications (such as `squawk watch` TUI), use the official ASCII block banner:

```text
  ___  FE  _   _   ___  _    _ _  __
 / __|/ _ \| | | |/ _ \| |  | | |/ /
 \__ \ (_) | |_| | (_) | |__| | ' < 
 |___/\__\_\\___/ \__\_|____|_|_|\_\
  [ SIGNAL BUG  •  HAND TO AI ]
```

---

## 6. Usage & Misuse Guidance

### Recommended Usage
- Use `squawk-logo.svg` or `squawk-logo-dark.svg` at the top of repository documentation.
- Use `squawk-mark.svg` or `favicon.svg` for browser icons, web widget headers, and avatars.
- Ensure proper contrast when displaying over custom background colors.

### Misuse Guidance
- ❌ **Do not stretch or distort** the mark or typography ratio.
- ❌ **Do not replace brand colors** with arbitrary uncurated colors.
- ❌ **Do not add heavy drop shadows** or complex 3D bevel effects to vector assets.
- ❌ **Do not place dark logo versions** on dark background elements without checking contrast.

---

## 7. PNG Export Instructions

If raster PNG versions are needed (e.g. for platforms that do not support SVG previews), convert the SVGs using `inkscape`, `rsvg-convert`, or `imagemagick`:

```bash
# Using rsvg-convert
rsvg-convert -w 1280 -h 640 assets/branding/social-preview.svg -o assets/branding/social-preview.png
rsvg-convert -w 512 -h 512 assets/branding/favicon.svg -o assets/branding/favicon.png

# Using Inkscape
inkscape assets/branding/social-preview.svg --export-filename=assets/branding/social-preview.png -w 1280 -h 640
```
