# Squawk — Session Log

Working log of Squawk sessions: what was built, decided, and is up next.

## Session 1 — CLI MVP + installer (2026-09-11)

**Goal:** working Android-only CLI, buildable in days. **Status: done.**

**Built (Go, module `github.com/DRMCGT/Squawk`, Go 1.23):**
- Commands: `init`, `capture` (`--note`/stdin), `watch`, `report`
  (markdown/json), `devices`, `version`. Session pointer
  `.squawk/current-session`.
- Capture: `adb exec-out screencap -p` + `adb logcat -d -t N`, PNG-validated,
  atomic session writes. Report: 4-backtick logcat fence,
  `Squawk 00N (no note)` fallback, relative image paths. Version via
  `-ldflags`.
- Tests: unit + CLI integration via `testdata/fake-adb` (no emulator needed).
- Deps pinned for Go 1.23: cobra v1.8.1, x/term v0.23.0, x/sys v0.24.0.
  Toolchain at `/home/lamb/opt/go` (not on default PATH).

**Distribution:**
- `install.sh` (Linux/macOS, `~/.local/bin`, SHA256-verified) + `make release`
  tarballs + checksums. README leads with one-liner.
- Pushed to `github.com/DRMCGT/Squawk`, tag `v0.1.0`, release assets
  published, repo made public. One-liner verified live
  (`squawk version v0.1.0`).

**Machine setup:** `~/Android/Sdk/platform-tools` added to PATH in `~/.bashrc`
(adb 37.0.1). AVD `Medium_Phone` exists but emulator intentionally not used.

**Decisions (browser backend, agreed 2026-09-11):**
- Attach to user-launched Chrome (`--remote-debugging-port=9222`), not
  Squawk-managed Chrome.
- Log scope: console + uncaught exceptions (no network-request logging yet).

## Session 2 — Browser backend (2026-09-11)

**Goal:** same CLI (`init`/`capture`/`watch`/`report`/`devices`) working against
a Chrome on localhost, mirroring the Android flow. **Status:** in progress.

- [x] Write `Session.md` (this file).
- [ ] Confirm the GitHub PAT used in Session 1 is revoked (was pasted in chat).
- [x] Backend abstraction: `capture.Backend` interface; adapt `adb.Client`;
      switch `capture.Service` off `*adb.Client` (Android stays green).
- [x] New `internal/browser` package (gorilla/websocket CDP client): tab
      listing via `/json/list`, tab screenshot with PNG validation, console +
      `Runtime.exceptionThrown` log buffer honoring `--log-lines`.
- [x] Session + report: `Backend` (`android`/`browser`) + `Target` (serial or
      page URL) + `Endpoint` (CDP URL); report header shows backend + target.
      Keep `screenshot.png`/`logcat.txt` artifact names.
- [x] CLI: `init --backend browser [--cdp ...] [--tab ...]`; `capture`/`watch`
      inherit backend from session; `devices` lists tabs for browser sessions.
- [x] Tests + docs: fake CDP server unit tests (no Chrome needed); README
      Chrome launch line.
- [x] Verify against a real Chrome on localhost (E2E): headless Chrome +
      file:// page; `init --backend browser --tab index.html`, `capture`,
      `report` all produced a valid PNG + console/exception excerpt.
- [x] Ship: commit → push `main` → tag `v0.2.0` → confirm release assets →
      re-run the install one-liner as proof.
- [ ] Stretch: `init` auto-detects `~/Android/Sdk/platform-tools` before
      failing on missing adb (noted as follow-up in Session 1).

**Acceptance:** `squawk init --backend browser` against a real Chrome on
localhost, `squawk capture --note ...`, and a `report.md` with screenshot +
console excerpt — mirroring the Session 1 demo, no code changes to the Android
path.

## Session 3 — Browser backend field-test + v0.1 polish (2026-09-11/12)

**Goal:** use Squawk on a real project (CillusApp in Chrome), then polish the
CLI for the OSS release. **Status: done.**

**Field test (browser backend, real Chrome + Capacitor app on localhost:8080):**
- 5 captures: valid PNGs + console excerpts each; `report.md` handed to a
  coding agent — the intended loop works with no emulator.
- Findings (accepted): browser logs are a 2s live window (re-captures without
  a page reload replay the same page-load lines); `watch` is type-the-note
  (two captures recorded the typed command text as the note).

**Shipped in between:** banner + watch slash commands as `v0.1.1` (fix
`vv0.1.1` double-v: `v0.1.2`), both pushed + released; install one-liner
re-verified. PAT used for pushes was pasted in chat again — revoke it.

**Session 4 — watch TUI addendum (2026-09-12, supersedes plain polish path):**
- `squawk watch` is now a full-screen Bubble Tea + Lip Gloss TUI: block-letter
  SQUAWK logo, session/backend/target/count status line, rounded boxed input
  with placeholder, dim footer hint bar, async `tea.Cmd` captures with inline
  "Capturing…"/"Writing report…" states and fading confirmations. Same command
  set and parsing behavior as the plain loop; everything else (`init`,
  `capture`, `report`, `devices`, `version`) stays plain text.
- When stdin/stdout are not TTYs (piped/scripted), `watch` falls back to the
  original plain loop — existing integration tests unchanged.
- Deps: bubbletea v1.2.4 + lipgloss v1.0.0 (x/sys bumped to v0.27.0; go.mod
  stays on Go 1.23 — newer charmbracelet releases require 1.24).
- Fixed a real bubbletea input quirk in tests + live PTY: a lone Space arrives
  as `tea.KeySpace`, not `KeyRunes` — must be handled or spaces vanish from
  notes.
- Verified over a real PTY (slow-adb wrapper): "Capturing…" renders while a
  0.8s capture runs; notes with spaces persisted; report + quit paths clean.

## Session 5 — Web widget MVP (2026-09-12, supersedes the Go-only CLI for the web)

**Goal:** an embeddable, zero-backend bug-flagging widget for web apps
replacing the CLI flow for the browser iteration (no screenshots, no adb, no
CDP, no server). **Status: done.**

**Built (`widget/`, standalone npm package, `squawk-widget` v0.1.0):**
- Vanilla JS + Shadow DOM (no framework, no external runtime deps).
- Floating "Squawk" button bottom-right → panel with note textarea + Save,
  running flag count, recent-flag list, "Export .md", and "Clear" (confirm-gated).
- Flags persist under `localStorage["squawk:flags"]` as `{id, note, url, timestamp}`;
  ids are `F001`-style (max-suffix + 1). In-memory fallback if localStorage is
  unavailable. Malformed/foreign payloads are dropped on load.
- Export: Blob + `<a download="bugs.md">`, format per spec (first ~8 words of
  the note become the heading; full note text preserved verbatim).
- Builds (esbuild): ESM (`squawk.esm.js`), CJS (`squawk.cjs.js`), and plain-script
  IIFE (`squawk.js`, exposes `window.Squawk` with `.mount()`). Minified IIFE was
  8.4KB at v0 (10KB target); grew to ~12.4KB with pin-to-locate, gate revised
  to 13KB (see Session 5 addendum below). Gated by `npm run check:size`.
- Tests: 42 vitest+jsdom tests (store, markdown, download, widget interactions,
  plus an integration test that loads the built IIFE and drives the full
  flag → localStorage → export loop).
- Docs: `widget/README.md` (usage, API, export format) + root README section.
- CI: `.github/workflows/test.yml` now also builds + tests the widget.

**Session 5 addendum — pre-launch polish (2026-09-12):**
- FAB now shows a paper-airplane inline SVG (no icon library dependency).
- Export filename is timestamped (`bugs-YYYYMMDD-HHmmss-<ms>.md`), shared with
  the `Exported:` line and filesystem-safe; repeat exports never overwrite.
- **Pin-to-locate:** `Pin location` arms a transparent full-page overlay that
  intercepts the next click (never reaches the app), drops a numbered pin at
  the spot, and anchors the note form next to it. Flags record `position`
  (`xPercent`/`yPercent` relative to the full document + viewport size) and the
  export gains a `- Position: …% from left, …% from top (viewport WxH)` line.
  Pins are live-only (not redrawn across reloads); optional `position` keeps
  plain page-level flags unchanged.
- Bundle grew to ~12.4KB minified (~4.7KB gzipped); size gate revised to 13KB
  to fit the added pin feature. Tests: 59 total (was 42).

## Session 6 — Element picker + named export (2026-09-12)

**Goal:** replace raw click coordinates with a DevTools-style element picker
and let the tester name the export file. **Status: done.**

- **Element picker** replaces the overlay/pin approach: `Pick element` arms
  capture-phase `document` listeners for `mousemove` (blue highlight box tracks
  the hovered DOM element via `getBoundingClientRect`, no overlay div) and
  `click` (intercepts before app handlers, `preventDefault` +
  `stopImmediatePropagation`, selects `event.target`). Note form anchors near
  the selected element; Esc/Cancel clears the selection.
- **Selector generation** (`src/selector.js`, no deps): prefer the element's
  `#id` (or nearest ancestor id), else a short tag + class path with
  `:nth-child()` only when siblings are ambiguous; `textPreview` = first ~50
  chars of text, whitespace-collapsed. `escapeCss` with a fallback.
- **Data model:** `position` → `element { selector, tagName, textPreview }` plus
  required `screenLabel` (first visible h1/h2/h3, else `document.title`, else
  url). Old stored flags (position / no screenLabel) still load and export.
- **Named export:** inline **Filename** field (not `prompt()`) pre-filled with
  the timestamped name; export uses the field's current value; untouched
  exports refresh the field so repeats never overwrite; `filename` mount option
  pre-fills it as a fixed override.
- **Markdown:** `- Screen:` and `- Element: \`sel\` — "preview"` lines;
  `- Position:` still rendered for legacy flags.
- Bundle ~14.5KB minified (~5KB gzipped); size gate revised to 15KB. Tests:
  89 total (was 59), including a new `selector.test.js`.

**Decisions:**
- No spatial/pixel pins in v0 — plain per-page note list only; add pin
  coordinates later only if usage shows they're missing.
- Package name pending (likely under the Squawk brand); placeholder is
  `squawk-widget`.
- Manual mounting (no auto-mount side effects); ESM consumers gate on
  `NODE_ENV === 'development'`.

**Remaining / parked:**
- Go/CLI work is parked, not deleted (still the Android + screenshot path).
- Advanced cloud/SaaS phase (cloud sandbox, LLM-driven review) tracked
  separately, not part of this MVP.