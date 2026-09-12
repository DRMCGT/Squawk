export const STORAGE_KEY = "squawk:flags";

function memoryStorage() {
  const map = new Map();
  return {
    getItem(key) {
      return map.has(key) ? map.get(key) : null;
    },
    setItem(key, value) {
      map.set(key, String(value));
    },
    removeItem(key) {
      map.delete(key);
    },
  };
}

// Use the real localStorage when available; fall back to an in-memory store so
// the widget still works in restrictive contexts (private mode, sandboxed
// iframes) even though flags will not survive a reload there.
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
  return (
    p != null &&
    typeof p === "object" &&
    typeof p.xPercent === "number" &&
    Number.isFinite(p.xPercent) &&
    typeof p.yPercent === "number" &&
    Number.isFinite(p.yPercent) &&
    typeof p.viewportWidth === "number" &&
    typeof p.viewportHeight === "number"
  );
}

function isValidFlag(f) {
  return (
    f != null &&
    typeof f === "object" &&
    typeof f.id === "string" &&
    f.id !== "" &&
    typeof f.note === "string" &&
    typeof f.url === "string" &&
    typeof f.timestamp === "string" &&
    (f.position == null || isValidPosition(f.position))
  );
}

export function createStore({ storage = null, storageKey = STORAGE_KEY } = {}) {
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
    },
  };
}

// "F001" -> "F002", skipping over gaps by taking the max numeric suffix.
export function nextFlagId(flags) {
  let max = 0;
  for (const f of flags) {
    const m = /^F(\d+)$/.exec(f.id);
    if (m) max = Math.max(max, Number(m[1]));
  }
  return "F" + String(max + 1).padStart(3, "0");
}