import { createStore, nextFlagId } from "./store.js";
import { toMarkdown, exportFilename } from "./markdown.js";
import { downloadMarkdown } from "./download.js";
import { describeElement } from "./selector.js";

const STYLES =
  ":host{all:initial}" +
  ".squawk{position:fixed;right:16px;bottom:16px;z-index:2147483000;font-family:system-ui,-apple-system,'Segoe UI',Roboto,sans-serif;font-size:13px;line-height:1.4;color:#1a1a1a}" +
  ".fab{display:flex;align-items:center;gap:8px;height:44px;padding:0 16px;border:0;border-radius:22px;background:#111;color:#fff;font:inherit;font-weight:600;cursor:pointer;box-shadow:0 4px 16px rgba(0,0,0,.25)}" +
  ".fab:hover{background:#000}.fab .icon{display:inline-flex}" +
  ".badge{display:inline-flex;align-items:center;justify-content:center;min-width:18px;height:18px;padding:0 5px;border-radius:9px;background:#e5484d;color:#fff;font-size:11px;font-weight:700}.badge.zero{background:#6b7280}" +
  ".panel{position:fixed;right:16px;bottom:68px;width:320px;max-width:calc(100vw - 32px);background:#fff;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 8px 32px rgba(0,0,0,.18);overflow:hidden;display:none}.panel.open{display:block}.panel.anchored{right:auto;bottom:auto}" +
  ".head{display:flex;align-items:center;justify-content:space-between;padding:12px 16px;border-bottom:1px solid #e5e7eb}.head h1{margin:0;font-size:14px;font-weight:700}" +
  ".close{border:0;background:none;cursor:pointer;font-size:16px;color:#6b7280;padding:2px 6px}" +
  ".body{padding:16px}textarea{width:100%;box-sizing:border-box;resize:vertical;min-height:64px;border:1px solid #d1d5db;border-radius:8px;padding:8px 10px;font:inherit}textarea:focus{outline:2px solid #111;outline-offset:-1px}" +
  ".save{display:block;width:100%;margin-top:8px;padding:9px 0;border:0;border-radius:8px;background:#111;color:#fff;font:inherit;font-weight:600;cursor:pointer}.save:disabled{opacity:.5}" +
  ".status{margin:12px 0 8px;color:#6b7280}" +
  ".list{margin:0 0 12px;padding:0;list-style:none;max-height:180px;overflow-y:auto;border:1px solid #f3f4f6;border-radius:8px}" +
  ".list li{padding:8px 10px;border-bottom:1px solid #f3f4f6}.list li:last-child{border-bottom:0}.list .t{font-weight:600;word-break:break-word}.list .meta{color:#6b7280;font-size:11px;margin-top:2px}.empty{color:#9ca3af}" +
  ".file{display:flex;align-items:center;gap:6px;margin-bottom:8px}.file label{color:#6b7280;font-size:11px;flex:0 0 auto}.file input{flex:1;min-width:0;box-sizing:border-box;border:1px solid #d1d5db;border-radius:6px;padding:5px 8px;font:inherit;font-size:12px}" +
  ".actions{display:flex;gap:8px}.actions button{flex:1;padding:8px 0;border:1px solid #d1d5db;border-radius:8px;background:#fff;font:inherit;font-weight:600;cursor:pointer}" +
  ".actions .active{background:#111;color:#fff;border-color:#111}.actions .active:hover{background:#000}" +
  ".actions .danger{color:#b91c1c;border-color:#fecaca}.actions .danger:hover{background:#fef2f2}.actions button:disabled{opacity:.5}";

function pageUrl() {
  const loc = globalThis.location;
  return (loc.pathname || "/") + (loc.search || "") + (loc.hash || "");
}

function clamp(v, min, max) {
  return Math.min(max, Math.max(min, v));
}

// A small paper-airplane inline SVG, matching the aviation branding. No icon
// library dependency — this is the only icon the widget ships. The string is
// static (no user data), so innerHTML is safe here.
function paperPlaneIcon() {
  const span = document.createElement("span");
  span.className = "icon";
  span.innerHTML =
    '<svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">' +
    '<path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>';
  return span;
}

function isVisible(el) {
  const style = getComputedStyle(el);
  return style.display !== "none" && style.visibility !== "hidden";
}

// The screen name for a flag: first visible heading, then document.title, then
// the page path.
function screenLabel() {
  const headings = document.querySelectorAll("h1, h2, h3");
  for (const h of headings) {
    if (!isVisible(h)) continue;
    const text = (h.textContent || "").replace(/\s+/g, " ").trim();
    if (text) return text;
  }
  const title = (document.title || "").trim();
  if (title) return title;
  return pageUrl();
}

export function createWidget({ store, filename } = {}) {
  const host = document.createElement("div");
  host.setAttribute("data-squawk-widget", "");

  const shadow = host.attachShadow({ mode: "open" });
  const style = document.createElement("style");
  style.textContent = STYLES;
  shadow.appendChild(style);

  const root = document.createElement("div");
  root.className = "squawk";
  shadow.appendChild(root);

  const fab = document.createElement("button");
  fab.type = "button";
  fab.className = "fab";
  fab.setAttribute("aria-label", "Toggle Squawk");
  const icon = paperPlaneIcon();
  const fabLabel = document.createElement("span");
  fabLabel.textContent = "Squawk";
  const badge = document.createElement("span");
  badge.className = "badge";
  fab.append(icon, fabLabel, badge);

  const panel = document.createElement("section");
  panel.className = "panel";
  panel.setAttribute("aria-label", "Squawk panel");

  const head = document.createElement("div");
  head.className = "head";
  const heading = document.createElement("h1");
  heading.textContent = "Squawk";
  const close = document.createElement("button");
  close.type = "button";
  close.className = "close";
  close.setAttribute("aria-label", "Close");
  close.textContent = "×";
  head.append(heading, close);

  const body = document.createElement("div");
  body.className = "body";
  const textarea = document.createElement("textarea");
  textarea.placeholder = "What did you spot?";
  textarea.rows = 3;
  const save = document.createElement("button");
  save.type = "button";
  save.className = "save";
  save.textContent = "Save";
  save.disabled = true;
  const status = document.createElement("p");
  status.className = "status";
  const list = document.createElement("ul");
  list.className = "list";

  const fileRow = document.createElement("div");
  fileRow.className = "file";
  const fileLabel = document.createElement("label");
  fileLabel.textContent = "Filename";
  fileLabel.setAttribute("for", "squawk-filename");
  const filenameInput = document.createElement("input");
  filenameInput.type = "text";
  filenameInput.id = "squawk-filename";
  filenameInput.setAttribute("aria-label", "Export filename");
  fileRow.append(fileLabel, filenameInput);

  const actions = document.createElement("div");
  actions.className = "actions";
  const pickBtn = document.createElement("button");
  pickBtn.type = "button";
  pickBtn.className = "pick";
  pickBtn.textContent = "Pick element";
  pickBtn.setAttribute("title", "Hover to highlight an element, click to select it");
  const exportBtn = document.createElement("button");
  exportBtn.type = "button";
  exportBtn.textContent = "Export .md";
  exportBtn.disabled = true;
  const clearBtn = document.createElement("button");
  clearBtn.type = "button";
  clearBtn.className = "danger";
  clearBtn.textContent = "Clear";
  clearBtn.disabled = true;
  actions.append(pickBtn, exportBtn, clearBtn);

  body.append(textarea, save, status, list, fileRow, actions);
  panel.append(head, body);
  root.append(fab, panel);
  document.body.appendChild(host);

  const hoverBox = document.createElement("div");
  hoverBox.setAttribute("data-squawk-highlight", "");
  hoverBox.style.cssText =
    "position:fixed;z-index:2147482500;display:none;pointer-events:none;box-sizing:border-box;" +
    "background:rgba(66,133,244,.18);border:2px solid #4285f4;";
  document.body.appendChild(hoverBox);

  let flags = store.load();
  let armed = false;
  let pendingElement = null;
  let pendingElRef = null;
  let highlightEl = null;
  let filenameDirty = false;

  filenameInput.value = filename || exportFilename(new Date());
  filenameDirty = Boolean(filename);

  function isWidgetEvent(event) {
    const t = event.target;
    if (!t) return false;
    if (t === host || t === shadow || host.contains(t) || shadow.contains(t)) return true;
    if (typeof event.composedPath === "function") return event.composedPath().includes(host);
    return false;
  }

  function hideHighlight() {
    hoverBox.style.display = "none";
    highlightEl = null;
  }

  function showHighlight(el) {
    const r = el.getBoundingClientRect();
    hoverBox.style.display = "block";
    hoverBox.style.left = `${r.left}px`;
    hoverBox.style.top = `${r.top}px`;
    hoverBox.style.width = `${r.width}px`;
    hoverBox.style.height = `${r.height}px`;
  }

  function onMove(event) {
    if (isWidgetEvent(event)) {
      hideHighlight();
      return;
    }
    const el = event.target;
    if (!el || el.nodeType !== 1) {
      hideHighlight();
      return;
    }
    highlightEl = el;
    showHighlight(el);
  }

  function onPick(event) {
    if (isWidgetEvent(event)) return;
    event.preventDefault();
    event.stopPropagation();
    event.stopImmediatePropagation();
    const el = event.target;
    if (!el || el.nodeType !== 1) return;
    pendingElRef = el;
    pendingElement = describeElement(el);
    disarm();
    showHighlight(el);
    openPanelNearEl(el);
  }

  function render() {
    const n = flags.length;
    badge.textContent = String(n);
    badge.classList.toggle("zero", n === 0);
    status.textContent = n === 1 ? "1 bug flagged" : `${n} bugs flagged`;
    exportBtn.disabled = n === 0;
    clearBtn.disabled = n === 0;

    list.textContent = "";
    if (n === 0) {
      const li = document.createElement("li");
      li.className = "empty";
      li.textContent = "No bugs flagged yet.";
      list.appendChild(li);
      return;
    }
    for (const f of flags) {
      const li = document.createElement("li");
      const t = document.createElement("div");
      t.className = "t";
      t.textContent = f.note || "(no note)";
      const meta = document.createElement("div");
      meta.className = "meta";
      const loc = f.element
        ? `${f.id} · ${f.element.selector}`
        : f.position
          ? `${f.id} · ${Math.round(f.position.xPercent * 100)}% / ${Math.round(f.position.yPercent * 100)}%`
          : `${f.id} · ${f.url}`;
      meta.textContent = `${loc} · ${new Date(f.timestamp).toLocaleString()}`;
      li.append(t, meta);
      list.appendChild(li);
    }
  }

  function add(note, element) {
    const flag = {
      id: nextFlagId(flags),
      note: String(note ?? "").trim(),
      url: pageUrl(),
      screenLabel: screenLabel(),
      timestamp: new Date().toISOString(),
    };
    if (element) flag.element = element;
    flags = flags.concat(flag);
    store.save(flags);
    render();
    return flag;
  }

  function saveFromInput() {
    if (!textarea.value.trim()) return;
    add(textarea.value, pendingElement || undefined);
    textarea.value = "";
    save.disabled = true;
    cancelSelection();
  }

  function exportMarkdown() {
    if (!flags.length) return;
    const exportedAt = new Date();
    let name = filenameInput.value.trim();
    if (!name) {
      name = exportFilename(exportedAt);
      filenameInput.value = name;
      filenameDirty = false;
    }
    downloadMarkdown(toMarkdown(flags, exportedAt), name);
    if (!filenameDirty) filenameInput.value = exportFilename(new Date());
  }

  function clear() {
    if (!flags.length) return;
    const message = `Clear all ${flags.length} flagged bug${flags.length === 1 ? "" : "s"}? This cannot be undone.`;
    if (!globalThis.confirm(message)) return;
    flags = [];
    store.clear();
    disarm();
    cancelSelection();
    render();
  }

  function setOpen(open) {
    panel.classList.toggle("open", open);
    if (!open) {
      cancelSelection();
      resetPanelPosition();
    }
  }

  function resetPanelPosition() {
    panel.style.left = "";
    panel.style.top = "";
    panel.classList.remove("anchored");
  }

  function cancelSelection() {
    pendingElement = null;
    pendingElRef = null;
    hideHighlight();
  }

  function onEscape(e) {
    if (e.key !== "Escape") return;
    if (armed) {
      disarm();
      cancelSelection();
      return;
    }
    setOpen(false);
  }

  function openPanelNearEl(el) {
    const r = el.getBoundingClientRect();
    setOpen(true);
    const vw = window.innerWidth || 800;
    const vh = window.innerHeight || 600;
    const pw = 320;
    const ph = panel.offsetHeight || 300;
    let left = r.left;
    let top = r.top - ph - 8;
    if (top < 8) top = r.bottom + 8;
    panel.style.left = `${clamp(left, 8, Math.max(8, vw - pw - 8))}px`;
    panel.style.top = `${clamp(top, 8, Math.max(8, vh - ph - 8))}px`;
    panel.classList.add("anchored");
    textarea.focus();
  }

  function arm() {
    if (armed) return;
    setOpen(false);
    armed = true;
    pickBtn.textContent = "Cancel pick";
    pickBtn.classList.add("active");
    document.addEventListener("mousemove", onMove, { capture: true });
    document.addEventListener("click", onPick, { capture: true });
  }

  function disarm() {
    if (!armed) return;
    armed = false;
    document.removeEventListener("mousemove", onMove, { capture: true });
    document.removeEventListener("click", onPick, { capture: true });
    pickBtn.textContent = "Pick element";
    pickBtn.classList.remove("active");
  }

  fab.addEventListener("click", () => {
    if (panel.classList.contains("open")) {
      setOpen(false);
    } else {
      setOpen(true);
      resetPanelPosition();
      textarea.focus();
    }
  });
  close.addEventListener("click", () => setOpen(false));
  save.addEventListener("click", saveFromInput);
  pickBtn.addEventListener("click", () => (armed ? disarm() : arm()));
  filenameInput.addEventListener("input", () => {
    filenameDirty = true;
  });
  textarea.addEventListener("input", () => {
    save.disabled = !textarea.value.trim();
  });
  textarea.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      saveFromInput();
    }
  });
  document.addEventListener("keydown", onEscape);
  exportBtn.addEventListener("click", exportMarkdown);
  clearBtn.addEventListener("click", clear);

  render();

  return {
    host,
    getFlags() {
      return flags.slice();
    },
    add,
    arm,
    disarm,
    open() {
      setOpen(true);
      resetPanelPosition();
    },
    close() {
      setOpen(false);
    },
    toggle() {
      if (panel.classList.contains("open")) {
        setOpen(false);
      } else {
        setOpen(true);
        resetPanelPosition();
      }
    },
    exportMarkdown,
    clear,
    unmount() {
      setOpen(false);
      disarm();
      document.removeEventListener("keydown", onEscape);
      hoverBox.remove();
      host.remove();
    },
  };
}