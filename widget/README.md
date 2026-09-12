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
everything you've flagged this session, plus **Pick element**, **Export .md**
and **Clear** (empties `localStorage` after a confirmation).

### Precise element picking

To flag an exact element on the page (DevTools-style, no code shown):

1. Click **Pick element** in the panel — a blue highlight box tracks the DOM
   element under your cursor as you move the mouse over the real UI.
2. Click the element to select it. Squawk intercepts the click via a
   capture-phase listener, so the app's own handler never fires and you can't
   accidentally submit a form or navigate. The note form opens anchored next to
   the selected element.
3. Save the note. The flag records the element's generated CSS selector, tag,
   and a short text preview.
4. Press **Esc** or click **Cancel pick** to back out.

Selector generation is a simplified "Copy selector": the element's `#id`, else a
short tag + class path with `:nth-child()` only when siblings would be
ambiguous, rooted at the nearest ancestor with an id. Good enough for a human
or an agent to locate the element in the source.

### Named export

Before exporting, the panel shows a **Filename** field pre-filled with a
timestamped name (`bugs-20260912-140740-123.md`). Edit it to choose any name;
Export uses whatever is in the field at that moment. If you leave it untouched,
the field refreshes after each export so repeat exports never overwrite.

## API

`mountSquawk(options?)` mounts the widget once and returns a handle. Calling it
again returns the same instance. Options:

| Option      | Default             | Description                                   |
| ----------- | ------------------- | --------------------------------------------- |
| `storage`   | `localStorage`      | Custom Storage-like object (testing, sandbox) |
| `storageKey`| `"squawk:flags"`    | localStorage key                              |
| `filename`  | generated           | Fixed export filename override (pre-fills the field) |

Handle methods:

- `getFlags()` — the current flags
- `add(note, element?)` — programmatically flag a bug; `element` (optional) is a
  `{ selector, tagName, textPreview }` target. Returns the created flag
- `arm()` / `disarm()` — turn element-picker mode on/off
- `open()` / `close()` / `toggle()` — panel visibility
- `exportMarkdown()` — trigger the download using the current filename field
- `clear()` — clear all flags after a confirmation
- `unmount()` — remove the widget from the DOM

Also exported: `createStore`, `toMarkdown`, `noteTitle`, `exportFilename`,
`describeElement`, `buildSelector`, `STORAGE_KEY`.

## Data model

Flags are stored as JSON under `squawk:flags`:

```ts
interface Flag {
  id: string;        // e.g. "F001"
  note: string;
  url: string;        // page path (+ query/hash) where it was flagged
  screenLabel: string; // first visible heading, else document.title, else url
  timestamp: string;  // ISO 8601
  element?: {
    selector: string;    // generated CSS selector
    tagName: string;
    textPreview: string; // first ~50 chars of the element's text, trimmed
  };
}
```

`element` is optional — plain page-level notes keep working without it. Flags
stored by earlier versions (with a `position` field, or without `screenLabel`)
remain loadable and export fine. If `localStorage` is unavailable (private
mode, sandboxed iframe), the widget falls back to an in-memory store so it
still works, but flags won't survive a reload.

## Export format

```markdown
# Squawk Bug Report
Exported: <ISO timestamp>

## F002 — wrong default selected
- Screen: Reports
- Element: `button.btn.text-green` — "Menos de 6 meses"
- Page: /dashboard
- Time: 2026-09-12T21:03:00Z

<full note text>
```

Headings use the first ~8 words of the note. The full note text is preserved
verbatim. `Element:` is included only for picked flags; `Position:` appears only
for flags stored by older widget versions.

Exports download as `bugs-YYYYMMDD-HHmmss-<ms>.md` (filesystem-safe, no colons
or slashes) unless the filename field is edited. The filename and the
`Exported:` line come from the same instant, and untouched repeat exports never
overwrite each other. Export is a browser download via Blob + `<a download>` —
no server round-trip.

## Development

```bash
npm install
npm run build   # ESM + CJS + plain-script (IIFE) bundles into dist/
npm test        # vitest + jsdom (no browser needed)
npm run check   # build + test + bundle-size gate (<15KB minified)
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