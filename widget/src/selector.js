export function escapeCss(ident) {
  if (typeof CSS !== "undefined" && typeof CSS.escape === "function") {
    return CSS.escape(ident);
  }
  return String(ident).replace(/[^a-zA-Z0-9_-]/g, (c) => `\\${c}`);
}

// First ~max chars of an element's text content, whitespace-collapsed/trimmed.
export function textPreview(el, max = 50) {
  const text = (el.textContent || "").replace(/\s+/g, " ").trim();
  return text.length > max ? text.slice(0, max) + "…" : text;
}

function usableClasses(node) {
  return Array.from(node.classList || []).filter((c) => /^[a-zA-Z0-9_-]+$/.test(c)).slice(0, 2);
}

function segmentFor(node) {
  let seg = node.tagName.toLowerCase();
  const classes = usableClasses(node);
  if (classes.length) seg += "." + classes.join(".");
  return seg;
}

function sameShape(a, b) {
  if (a.tagName !== b.tagName) return false;
  const ca = usableClasses(a);
  const cb = usableClasses(b);
  if (ca.length !== cb.length) return false;
  return ca.every((c) => cb.includes(c));
}

function childIndex(node) {
  let i = 0;
  for (let sib = node.parentNode.firstElementChild; sib; sib = sib.nextElementSibling) {
    i++;
    if (sib === node) return i;
  }
  return i;
}

// A simplified "Copy selector": prefer #id (own or nearest ancestor), else a
// short tag+class path with :nth-child() only when siblings would be
// ambiguous. Stops at body/html.
export function buildSelector(el, { maxDepth = 6 } = {}) {
  if (!el || el.nodeType !== 1) return "";
  if (el === document.body) return "body";
  if (el === document.documentElement) return "html";

  const segments = [];
  let node = el;
  while (
    node &&
    node.nodeType === 1 &&
    node !== document.body &&
    node !== document.documentElement &&
    segments.length < maxDepth
  ) {
    if (node.id) {
      segments.unshift("#" + escapeCss(node.id));
      break;
    }
    let seg = segmentFor(node);
    const parent = node.parentNode;
    if (parent && parent.nodeType === 1) {
      let same = 0;
      for (let sib = parent.firstElementChild; sib; sib = sib.nextElementSibling) {
        if (sib !== node && sameShape(sib, node)) same++;
      }
      if (same > 0) seg += ":nth-child(" + childIndex(node) + ")";
    }
    segments.unshift(seg);
    node = parent;
  }
  return segments.join(" > ");
}

export function describeElement(el) {
  if (!el || el.nodeType !== 1) return null;
  return {
    selector: buildSelector(el),
    tagName: el.tagName.toLowerCase(),
    textPreview: textPreview(el),
  };
}