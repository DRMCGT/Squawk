# Squawk Widget

A lightweight, embeddable bug-flagging widget for testing a web app locally.
While you browse, you squawk issues the moment you spot them: type a short
note, hit save, and the flag is stored in the browser's `localStorage`. When
you're done, one button exports everything as a `bugs.md` file — ready to hand
to a coding agent.

No backend, no network calls, no accounts, no screenshots. Everything lives in
`localStorage` until you export it.

> **Package name pending** — shipped here as `squawk-widget`; likely to move
> under the Squawk brand before publishing.

## Install

```bash
npm install squawk-widget
```

## Usage (bundler)

Import and mount it, ideally only in development:

```js
import { mountSquawk } from "squawk-widget";

if (process.env.NODE_ENV === "development") {
  mountSquawk();
}
```

Vite/other `import.meta.env` setups:

```js
import { mountSquawk } from "squawk-widget";

if (import.meta.env?.DEV) {
  mountSquawk();
}
```

## Usage (no bundler)

Load the plain-script build and mount manually:

```html
<script src="squawk.js"></script>
<script>
  Squawk.mount();
</script>
```

A floating "Squawk" button (paper-airplane icon) appears bottom-right. Click it
to open the note form. Save a note and the count increments. The panel shows
everything you've flagged this session, plus **Pin location**, **Export .md**
and **Clear** (empties `localStorage` after a confirmation).

### Precise pinning

To flag an exact spot on the page:

1. Click **Pin location** in the panel — a transparent crosshair overlay covers
   the page, so your next click is intercepted and never reaches the app.
2. Click the exact spot. A numbered pin appears there and the note form opens
   anchored next to it.
3. Save the note. The flag records the click position relative to the full
   document; the pin stays visible for this page session.
4. Press **Esc** or click **Cancel pin** to back out.

Pins are live-only: they show at the moment of flagging and are not re-drawn
when the page is later reloaded.

## API

`mountSquawk(options?)` mounts the widget once and returns a handle. Calling it
again returns the same instance. Options:

| Option      | Default             | Description                                   |
| ----------- | ------------------- | --------------------------------------------- |
| `storage`   | `localStorage`      | Custom Storage-like object (testing, sandbox) |
| `storageKey`| `"squawk:flags"`    | localStorage key                              |
| `filename`  | generated           | Export filename override (default is a timestamped `bugs-*.md`) |

Handle methods:

- `getFlags()` — the current flags
- `add(note, position?)` — programmatically flag a bug; `position` (optional)
  pins it to a spot. Returns the created flag
- `arm()` / `disarm()` — turn precise-pin mode on/off
- `open()` / `close()` / `toggle()` — panel visibility
- `exportMarkdown()` — trigger the timestamped `bugs-*.md` download
- `clear()` — clear all flags after a confirmation
- `unmount()` — remove the widget from the DOM

Also exported: `createStore`, `toMarkdown`, `noteTitle`, `exportFilename`,
`STORAGE_KEY`.

## Data model

Flags are stored as JSON under `squawk:flags`:

```ts
interface Flag {
  id: string;        // e.g. "F001"
  note: string;
  url: string;        // page path (+ query/hash) where it was flagged
  timestamp: string;  // ISO 8601
  position?: {
    xPercent: number;   // 0–1 relative to full page width
    yPercent: number;   // 0–1 relative to full document height
    viewportWidth: number;
    viewportHeight: number;
  };
}
```

`position` is optional — plain page-level notes keep working without it. If
`localStorage` is unavailable (private mode, sandboxed iframe), the widget
falls back to an in-memory store so it still works, but flags won't survive a
reload.

## Export format

```markdown
# Squawk Bug Report
Exported: <ISO timestamp>

## F001 — save button unresponsive
- Page: /dashboard
- Time: 2026-09-12T21:03:00Z
- Position: 42% from left, 18% from top (viewport 1440x900)

<full note text>
```

Headings use the first ~8 words of the note. The full note text is preserved
verbatim. The `Position` line is included only for pinned flags.

Exports download as `bugs-YYYYMMDD-HHmmss-<ms>.md` (filesystem-safe, no colons
or slashes). The filename and the `Exported:` line come from the same instant,
and the milliseconds component means repeat exports never overwrite each other.
Use the `filename` option to force a fixed name. Export is a browser download
via Blob + `<a download>` — no server round-trip.

## Development

```bash
npm install
npm run build   # ESM + CJS + plain-script (IIFE) bundles into dist/
npm test        # vitest + jsdom (no browser needed)
npm run check   # build + test + bundle-size gate (<13KB minified)
```

Try it against the demo page:

```bash
cd widget
npx esbuild demo/index.html --servedir=. --outdir=dist-demo
# open http://localhost:8000/demo/index.html
```

## Out of scope (v0)

- Screenshot capture and spatial/pixel annotation
- Any server, account, or network calls
- Browser extension distribution
- Automated bug detection — flagging is manual by design

See the repo root for the CLI product (Android / Chrome CDP capture), which
remains separate from this widget.