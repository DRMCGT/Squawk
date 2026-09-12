import { createStore, STORAGE_KEY } from "./store.js";
import { createWidget } from "./widget.js";
import { toMarkdown, noteTitle } from "./markdown.js";

export const version = "0.1.0";

let active = null;

// Mounts the widget once and returns a handle for programmatic control.
// Options:
//   storage     - custom Storage-like object (defaults to localStorage)
//   storageKey  - localStorage key (default "squawk:flags")
//   filename    - exported report filename override; by default a timestamped
//                 name is generated (e.g. bugs-20260912-140740-123.md)
export function mountSquawk(options = {}) {
  if (active) return active;
  const store = createStore({
    storage: options.storage,
    storageKey: options.storageKey,
  });
  const widget = createWidget({ store, filename: options.filename });
  active = {
    ...widget,
    unmount() {
      widget.unmount();
      active = null;
    },
  };
  return active;
}

export { createStore, STORAGE_KEY, toMarkdown, noteTitle };