// src/store.js
var STORAGE_KEY = "squawk:flags";
function memoryStorage() {
  const map = /* @__PURE__ */ new Map();
  return {
    getItem(key) {
      return map.has(key) ? map.get(key) : null;
    },
    setItem(key, value) {
      map.set(key, String(value));
    },
    removeItem(key) {
      map.delete(key);
    }
  };
}
function resolveBacking(storage) {
  if (storage) return storage;
  try {
    const probe = "__squawk_probe__";
    globalThis.localStorage.setItem(probe, "1");
    globalThis.localStorage.removeItem(probe);
    return globalThis.localStorage;
  } catch {
    return memoryStorage();
  }
}
function isValidPosition(p) {
  return p != null && typeof p === "object" && typeof p.xPercent === "number" && Number.isFinite(p.xPercent) && typeof p.yPercent === "number" && Number.isFinite(p.yPercent) && typeof p.viewportWidth === "number" && typeof p.viewportHeight === "number";
}
function isValidFlag(f) {
  return f != null && typeof f === "object" && typeof f.id === "string" && f.id !== "" && typeof f.note === "string" && typeof f.url === "string" && typeof f.timestamp === "string" && (f.position == null || isValidPosition(f.position));
}
function createStore({ storage = null, storageKey = STORAGE_KEY } = {}) {
  const backing = resolveBacking(storage);
  return {
    load() {
      try {
        const raw = backing.getItem(storageKey);
        if (!raw) return [];
        const parsed = JSON.parse(raw);
        return Array.isArray(parsed) ? parsed.filter(isValidFlag) : [];
      } catch {
        return [];
      }
    },
    save(flags) {
      backing.setItem(storageKey, JSON.stringify(flags));
    },
    clear() {
      backing.removeItem(storageKey);
    }
  };
}
function nextFlagId(flags) {
  let max = 0;
  for (const f of flags) {
    const m = /^F(\d+)$/.exec(f.id);
    if (m) max = Math.max(max, Number(m[1]));
  }
  return "F" + String(max + 1).padStart(3, "0");
}

// src/markdown.js
function noteTitle(note) {
  const words = String(note).trim().split(/\s+/).filter(Boolean);
  return words.slice(0, 8).join(" ");
}
function pad(n, width = 2) {
  return String(n).padStart(width, "0");
}
function formatExportStamp(date) {
  const d = new Date(date);
  return d.getUTCFullYear() + pad(d.getUTCMonth() + 1) + pad(d.getUTCDate()) + "-" + pad(d.getUTCHours()) + pad(d.getUTCMinutes()) + pad(d.getUTCSeconds()) + "-" + pad(d.getUTCMilliseconds(), 3);
}
function exportFilename(date) {
  return `bugs-${formatExportStamp(date)}.md`;
}
function formatPosition(p) {
  const x = Math.round(p.xPercent * 100);
  const y = Math.round(p.yPercent * 100);
  return `- Position: ${x}% from left, ${y}% from top (viewport ${p.viewportWidth}x${p.viewportHeight})`;
}
function toMarkdown(flags, exportedAt = /* @__PURE__ */ new Date()) {
  const lines = [
    "# Squawk Bug Report",
    `Exported: ${exportedAt.toISOString()}`,
    ""
  ];
  if (!flags.length) {
    lines.push("_No bugs flagged._");
    return lines.join("\n") + "\n";
  }
  for (const f of flags) {
    const title = noteTitle(f.note) || "(no note)";
    lines.push("", `## ${f.id} \u2014 ${title}`);
    lines.push(`- Page: ${f.url}`);
    lines.push(`- Time: ${f.timestamp}`);
    if (f.position) lines.push(formatPosition(f.position));
    lines.push("");
    if (f.note) lines.push(f.note);
  }
  return lines.join("\n") + "\n";
}

// src/download.js
function downloadMarkdown(markdown, filename = "bugs.md") {
  const blob = new Blob([markdown], { type: "text/markdown;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.rel = "noopener";
  a.style.display = "none";
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1e3);
}

// src/widget.js
var STYLES = ":host{all:initial}.squawk{position:fixed;right:16px;bottom:16px;z-index:2147483000;font-family:system-ui,-apple-system,'Segoe UI',Roboto,sans-serif;font-size:13px;line-height:1.4;color:#1a1a1a}.fab{display:flex;align-items:center;gap:8px;height:44px;padding:0 16px;border:0;border-radius:22px;background:#111;color:#fff;font:inherit;font-weight:600;cursor:pointer;box-shadow:0 4px 16px rgba(0,0,0,.25)}.fab:hover{background:#000}.fab .icon{display:inline-flex}.badge{display:inline-flex;align-items:center;justify-content:center;min-width:18px;height:18px;padding:0 5px;border-radius:9px;background:#e5484d;color:#fff;font-size:11px;font-weight:700}.badge.zero{background:#6b7280}.panel{position:fixed;right:16px;bottom:68px;width:320px;max-width:calc(100vw - 32px);background:#fff;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 8px 32px rgba(0,0,0,.18);overflow:hidden;display:none}.panel.open{display:block}.panel.anchored{right:auto;bottom:auto}.head{display:flex;align-items:center;justify-content:space-between;padding:12px 16px;border-bottom:1px solid #e5e7eb}.head h1{margin:0;font-size:14px;font-weight:700}.close{border:0;background:none;cursor:pointer;font-size:16px;color:#6b7280;padding:2px 6px}.body{padding:16px}textarea{width:100%;box-sizing:border-box;resize:vertical;min-height:64px;border:1px solid #d1d5db;border-radius:8px;padding:8px 10px;font:inherit}textarea:focus{outline:2px solid #111;outline-offset:-1px}.save{display:block;width:100%;margin-top:8px;padding:9px 0;border:0;border-radius:8px;background:#111;color:#fff;font:inherit;font-weight:600;cursor:pointer}.save:disabled{opacity:.5}.status{margin:12px 0 8px;color:#6b7280}.list{margin:0 0 12px;padding:0;list-style:none;max-height:180px;overflow-y:auto;border:1px solid #f3f4f6;border-radius:8px}.list li{padding:8px 10px;border-bottom:1px solid #f3f4f6}.list li:last-child{border-bottom:0}.list .t{font-weight:600;word-break:break-word}.list .meta{color:#6b7280;font-size:11px;margin-top:2px}.empty{color:#9ca3af}.actions{display:flex;gap:8px}.actions button{flex:1;padding:8px 0;border:1px solid #d1d5db;border-radius:8px;background:#fff;font:inherit;font-weight:600;cursor:pointer}.actions .active{background:#111;color:#fff;border-color:#111}.actions .active:hover{background:#000}.actions .danger{color:#b91c1c;border-color:#fecaca}.actions .danger:hover{background:#fef2f2}.actions button:disabled{opacity:.5}";
function pageUrl() {
  const loc = globalThis.location;
  return (loc.pathname || "/") + (loc.search || "") + (loc.hash || "");
}
function clamp(v, min, max) {
  return Math.min(max, Math.max(min, v));
}
function paperPlaneIcon() {
  const span = document.createElement("span");
  span.className = "icon";
  span.innerHTML = '<svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>';
  return span;
}
function createPin(number) {
  const pin = document.createElement("div");
  pin.setAttribute("data-squawk-pin", "");
  pin.style.cssText = "position:absolute;transform:translate(-50%,-50%);width:22px;height:22px;border-radius:50%;background:#e5484d;color:#fff;font:700 12px system-ui,sans-serif;display:flex;align-items:center;justify-content:center;box-shadow:0 1px 4px rgba(0,0,0,.35);pointer-events:none";
  pin.textContent = String(number);
  return pin;
}
function createWidget({ store, filename } = {}) {
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
  close.textContent = "\xD7";
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
  const actions = document.createElement("div");
  actions.className = "actions";
  const pinBtn = document.createElement("button");
  pinBtn.type = "button";
  pinBtn.className = "pin";
  pinBtn.textContent = "Pin location";
  pinBtn.setAttribute("title", "Click a spot on the page to flag it precisely");
  const exportBtn = document.createElement("button");
  exportBtn.type = "button";
  exportBtn.textContent = "Export .md";
  exportBtn.disabled = true;
  const clearBtn = document.createElement("button");
  clearBtn.type = "button";
  clearBtn.className = "danger";
  clearBtn.textContent = "Clear";
  clearBtn.disabled = true;
  actions.append(pinBtn, exportBtn, clearBtn);
  body.append(textarea, save, status, list, actions);
  panel.append(head, body);
  root.append(fab, panel);
  document.body.appendChild(host);
  const pinLayer = document.createElement("div");
  pinLayer.setAttribute("data-squawk-pins", "");
  pinLayer.style.cssText = "position:absolute;top:0;left:0;width:0;height:0;z-index:2147482500;";
  document.body.appendChild(pinLayer);
  let flags = store.load();
  let armed = false;
  let overlay = null;
  let pendingPosition = null;
  let pendingPin = null;
  let livePins = [];
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
      const loc = f.position ? `${f.id} \xB7 ${f.url} \xB7 ${Math.round(f.position.xPercent * 100)}% / ${Math.round(f.position.yPercent * 100)}%` : `${f.id} \xB7 ${f.url}`;
      meta.textContent = `${loc} \xB7 ${new Date(f.timestamp).toLocaleString()}`;
      li.append(t, meta);
      list.appendChild(li);
    }
  }
  function add(note, position) {
    const flag = {
      id: nextFlagId(flags),
      note: String(note ?? "").trim(),
      url: pageUrl(),
      timestamp: (/* @__PURE__ */ new Date()).toISOString()
    };
    if (position) flag.position = position;
    flags = flags.concat(flag);
    store.save(flags);
    render();
    return flag;
  }
  function saveFromInput() {
    if (!textarea.value.trim()) return;
    add(textarea.value, pendingPosition || void 0);
    textarea.value = "";
    save.disabled = true;
    pendingPosition = null;
    pendingPin = null;
  }
  function exportMarkdown() {
    if (!flags.length) return;
    const exportedAt = /* @__PURE__ */ new Date();
    const name = filename || exportFilename(exportedAt);
    downloadMarkdown(toMarkdown(flags, exportedAt), name);
  }
  function clearLivePins() {
    for (const pin of livePins) pin.remove();
    livePins = [];
    pendingPin = null;
    pendingPosition = null;
  }
  function clear() {
    if (!flags.length) return;
    const message = `Clear all ${flags.length} flagged bug${flags.length === 1 ? "" : "s"}? This cannot be undone.`;
    if (!globalThis.confirm(message)) return;
    flags = [];
    store.clear();
    disarm();
    clearLivePins();
    render();
  }
  function setOpen(open) {
    panel.classList.toggle("open", open);
    if (!open) {
      cancelPendingPin();
      resetPanelPosition();
    }
  }
  function resetPanelPosition() {
    panel.style.left = "";
    panel.style.top = "";
    panel.classList.remove("anchored");
  }
  function cancelPendingPin() {
    if (pendingPin) {
      pendingPin.remove();
      livePins = livePins.filter((p) => p !== pendingPin);
      pendingPin = null;
    }
    pendingPosition = null;
  }
  function onEscape(e) {
    if (e.key !== "Escape") return;
    if (armed) {
      disarm();
      return;
    }
    setOpen(false);
  }
  function computePosition(clientX, clientY) {
    const doc = document.documentElement;
    const docW = doc.scrollWidth;
    const docH = doc.scrollHeight;
    const pageX = clientX + window.scrollX;
    const pageY = clientY + window.scrollY;
    return {
      xPercent: docW > 0 ? clamp(pageX / docW, 0, 1) : 0,
      yPercent: docH > 0 ? clamp(pageY / docH, 0, 1) : 0,
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight
    };
  }
  function openPanelNear(clientX, clientY) {
    setOpen(true);
    const vw = window.innerWidth || 800;
    const vh = window.innerHeight || 600;
    const pw = 320;
    const ph = panel.offsetHeight || 300;
    let left = clientX + 16;
    let top = clientY - ph - 24;
    if (left + pw > vw - 8) left = clientX - pw - 16;
    if (top < 8) top = clientY + 16;
    panel.style.left = `${clamp(left, 8, Math.max(8, vw - pw - 8))}px`;
    panel.style.top = `${clamp(top, 8, Math.max(8, vh - ph - 8))}px`;
    panel.classList.add("anchored");
    textarea.focus();
  }
  function placeAt(clientX, clientY) {
    pendingPosition = computePosition(clientX, clientY);
    const number = parseInt(nextFlagId(flags).slice(1), 10);
    const marker = createPin(number);
    pinLayer.appendChild(marker);
    const rect = pinLayer.getBoundingClientRect();
    marker.style.left = `${clientX - rect.left}px`;
    marker.style.top = `${clientY - rect.top}px`;
    pendingPin = marker;
    livePins.push(marker);
    disarm();
    openPanelNear(clientX, clientY);
  }
  function arm() {
    if (armed) return;
    setOpen(false);
    armed = true;
    pinBtn.textContent = "Cancel pin";
    pinBtn.classList.add("active");
    overlay = document.createElement("div");
    overlay.setAttribute("data-squawk-overlay", "");
    overlay.setAttribute("aria-label", "Squawk location picker: click anywhere on the page");
    overlay.style.cssText = "position:fixed;inset:0;z-index:2147482000;background:transparent;cursor:crosshair;";
    const hint = document.createElement("div");
    hint.textContent = "Click anywhere to pin \xB7 Esc to cancel";
    hint.style.cssText = "position:fixed;top:16px;left:50%;transform:translateX(-50%);background:#111;color:#fff;padding:6px 12px;border-radius:14px;font:600 12px system-ui,sans-serif;box-shadow:0 2px 8px rgba(0,0,0,.3);pointer-events:none;";
    overlay.appendChild(hint);
    overlay.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation();
      placeAt(e.clientX, e.clientY);
    });
    document.body.appendChild(overlay);
  }
  function disarm() {
    if (!armed) return;
    armed = false;
    overlay.remove();
    overlay = null;
    pinBtn.textContent = "Pin location";
    pinBtn.classList.remove("active");
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
  pinBtn.addEventListener("click", () => armed ? disarm() : arm());
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
      pinLayer.remove();
      host.remove();
    }
  };
}

// src/index.js
var version = "0.1.0";
var active = null;
function mountSquawk(options = {}) {
  if (active) return active;
  const store = createStore({
    storage: options.storage,
    storageKey: options.storageKey
  });
  const widget = createWidget({ store, filename: options.filename });
  active = {
    ...widget,
    unmount() {
      widget.unmount();
      active = null;
    }
  };
  return active;
}
export {
  STORAGE_KEY,
  createStore,
  mountSquawk,
  noteTitle,
  toMarkdown,
  version
};
