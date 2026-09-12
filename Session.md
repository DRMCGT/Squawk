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