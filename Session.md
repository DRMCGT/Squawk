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