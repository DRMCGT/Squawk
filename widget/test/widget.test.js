import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mountSquawk } from "../src/index.js";

function mockStorage(initial = {}) {
  const map = new Map(Object.entries(initial));
  return {
    getItem: (k) => (map.has(k) ? map.get(k) : null),
    setItem: (k, v) => map.set(k, String(v)),
    removeItem: (k) => map.delete(k),
  };
}

let current = null;
let confirmMock;

// mountSquawk is a module-level singleton, so tests must unmount the previous
// widget before mounting a fresh one with its own storage.
function mount(options) {
  if (current) current.unmount();
  current = mountSquawk(options);
  return current;
}

beforeEach(() => {
  confirmMock = vi.fn(() => true);
  vi.stubGlobal("confirm", confirmMock);
  URL.createObjectURL = vi.fn(() => "blob:mock-url");
  URL.revokeObjectURL = vi.fn();
});

afterEach(() => {
  if (current) current.unmount();
  current = null;
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  document.body.innerHTML = "";
});

function findWidget() {
  const host = document.querySelector("[data-squawk-widget]");
  expect(host).toBeTruthy();
  return host.shadowRoot;
}

function findPanel(root) {
  return root.querySelector(".panel");
}

function readBlob(blob) {
  return new Promise((resolve, reject) => {
    const fr = new FileReader();
    fr.onload = () => resolve(fr.result);
    fr.onerror = reject;
    fr.readAsText(blob);
  });
}

function stubAnchorDownload() {
  const anchor = document.createElement("a");
  anchor.click = vi.fn();
  vi.spyOn(document, "createElement").mockImplementation((tag) =>
    tag === "a" ? anchor : document.createElement(tag)
  );
  return anchor;
}

describe("mountSquawk", () => {
  it("renders a floating button with an icon and a zero badge", () => {
    const widget = mount({ storage: mockStorage() });
    const root = findWidget();
    const fab = root.querySelector(".fab");
    expect(fab).toBeTruthy();
    expect(fab.querySelector(".icon svg")).toBeTruthy();
    expect(fab.textContent).toContain("Squawk");
    expect(root.querySelector(".fab .badge").textContent).toBe("0");
    expect(widget.getFlags()).toEqual([]);
  });

  it("is idempotent", () => {
    const a = mount({ storage: mockStorage() });
    const b = mountSquawk({ storage: mockStorage() });
    expect(b).toBe(a);
    expect(document.querySelectorAll("[data-squawk-widget]")).toHaveLength(1);
  });

  it("unmount removes the host and allows remounting", () => {
    const widget = mount({ storage: mockStorage() });
    widget.unmount();
    expect(document.querySelector("[data-squawk-widget]")).toBeNull();
    const again = mount({ storage: mockStorage() });
    expect(again).not.toBe(widget);
    expect(document.querySelectorAll("[data-squawk-widget]")).toHaveLength(1);
  });

  it("restores flags already in storage", () => {
    const existing = [
      {
        id: "F001",
        note: "existing bug",
        url: "/",
        timestamp: "2026-09-12T20:00:00Z",
      },
    ];
    const storage = mockStorage({ "squawk:flags": JSON.stringify(existing) });
    const widget = mount({ storage });
    const root = findWidget();
    expect(widget.getFlags()).toEqual(existing);
    expect(root.querySelector(".fab .badge").textContent).toBe("1");
  });
});

describe("flagging", () => {
  it("add() writes a flag with id, note, url, screenLabel, and ISO timestamp", () => {
    const storage = mockStorage();
    const widget = mount({ storage });
    const flag = widget.add("save button unresponsive");

    expect(flag.id).toBe("F001");
    expect(flag.note).toBe("save button unresponsive");
    expect(flag.url).toBe("/");
    expect(flag.screenLabel).toBe("/");
    expect(new Date(flag.timestamp).toISOString()).toBe(flag.timestamp);
    expect(JSON.parse(storage.getItem("squawk:flags"))).toEqual([flag]);
  });

  it("increments ids across adds", () => {
    const widget = mount({ storage: mockStorage() });
    widget.add("first bug");
    const second = widget.add("second bug");
    expect(second.id).toBe("F002");
  });

  it("records the current page path including query and hash", () => {
    const original = globalThis.location;
    Object.defineProperty(globalThis, "location", {
      value: new URL("http://localhost:3000/dashboard?tab=reports#top"),
      configurable: true,
    });
    try {
      const widget = mount({ storage: mockStorage() });
      const flag = widget.add("broken chart");
      expect(flag.url).toBe("/dashboard?tab=reports#top");
    } finally {
      Object.defineProperty(globalThis, "location", { value: original, configurable: true });
    }
  });

  it("saves through the form when Save is clicked", () => {
    const storage = mockStorage();
    mount({ storage });
    const root = findWidget();
    const panel = findPanel(root);
    const textarea = panel.querySelector("textarea");
    const save = panel.querySelector(".save");

    expect(save.disabled).toBe(true);

    textarea.value = "form is broken";
    textarea.dispatchEvent(new Event("input", { bubbles: true }));
    expect(save.disabled).toBe(false);

    save.click();
    expect(JSON.parse(storage.getItem("squawk:flags"))[0].note).toBe("form is broken");
    expect(textarea.value).toBe("");
    expect(save.disabled).toBe(true);
  });

  it("updates the badge and list after saving", () => {
    const widget = mount({ storage: mockStorage() });
    const root = findWidget();
    widget.add("first");
    widget.add("second");

    expect(root.querySelector(".fab .badge").textContent).toBe("2");
    const items = root.querySelectorAll(".list li:not(.empty)");
    expect(items).toHaveLength(2);
    expect(items[0].querySelector(".t").textContent).toBe("first");
    expect(root.querySelector(".status").textContent).toBe("2 bugs flagged");
  });
});

describe("export", () => {
  it("exports a timestamped bugs.md download with the collected flags", async () => {
    const widget = mount({ storage: mockStorage() });
    widget.add("save button unresponsive");

    const create = vi.fn(() => "blob:mock-url");
    URL.createObjectURL = create;
    const anchor = stubAnchorDownload();

    widget.exportMarkdown();

    expect(create).toHaveBeenCalledTimes(1);
    expect(anchor.download).toMatch(/^bugs-\d{8}-\d{6}-\d{3}\.md$/);

    const md = await readBlob(create.mock.calls[0][0]);
    expect(md).toContain("# Squawk Bug Report");
    expect(md).toContain("## F001 — save button unresponsive");
    expect(md).toContain("- Page: /");
    expect(md).toContain("save button unresponsive");
    expect(anchor.click).toHaveBeenCalledTimes(1);
  });

  it("export button is disabled when there are no flags", () => {
    mount({ storage: mockStorage() });
    const root = findWidget();
    const panel = findPanel(root);
    const exportBtn = panel.querySelectorAll(".actions button")[1];
    expect(exportBtn.textContent).toBe("Export .md");
    expect(exportBtn.disabled).toBe(true);
  });

  it("supports a custom export filename", () => {
    const widget = mount({ storage: mockStorage(), filename: "qa-notes.md" });
    widget.add("a bug");
    const anchor = stubAnchorDownload();
    widget.exportMarkdown();
    expect(anchor.download).toBe("qa-notes.md");
  });
});

describe("clear", () => {
  it("clears flags and storage after confirmation", () => {
    const storage = mockStorage();
    const widget = mount({ storage });
    widget.add("one");
    widget.add("two");
    const root = findWidget();

    confirmMock.mockReturnValue(true);
    findPanel(root).querySelector(".danger").click();

    expect(widget.getFlags()).toEqual([]);
    expect(storage.getItem("squawk:flags")).toBeNull();
    expect(root.querySelector(".fab .badge").textContent).toBe("0");
    expect(confirmMock).toHaveBeenCalledTimes(1);
  });

  it("does nothing when confirmation is declined", () => {
    const storage = mockStorage();
    const widget = mount({ storage });
    widget.add("one");

    confirmMock.mockReturnValue(false);
    findPanel(findWidget()).querySelector(".danger").click();

    expect(widget.getFlags()).toHaveLength(1);
    expect(JSON.parse(storage.getItem("squawk:flags"))).toHaveLength(1);
  });
});

describe("element picker", () => {
  function appButton(label = "Click me") {
    const btn = document.createElement("button");
    btn.className = "btn text-green";
    btn.textContent = label;
    document.body.appendChild(btn);
    return btn;
  }

  function saveNote(note) {
    const root = findWidget();
    const panel = findPanel(root);
    const textarea = panel.querySelector("textarea");
    textarea.value = note;
    textarea.dispatchEvent(new Event("input", { bubbles: true }));
    panel.querySelector(".save").click();
    return textarea;
  }

  function forceTargetEvent(type, target, props = {}) {
    const ev = new MouseEvent(type, { bubbles: true, cancelable: true, ...props });
    Object.defineProperty(ev, "target", { value: target, configurable: true });
    document.dispatchEvent(ev);
    return ev;
  }

  it("arming attaches picker listeners without any overlay div", () => {
    const w = mount({ storage: mockStorage() });
    w.arm();
    const panel = findPanel(findWidget());
    expect(panel.querySelector(".pick").textContent).toBe("Cancel pick");
    expect(panel.querySelector(".pick").classList.contains("active")).toBe(true);
    expect(document.querySelector("[data-squawk-overlay]")).toBeNull();
    w.disarm();
    expect(panel.querySelector(".pick").textContent).toBe("Pick element");
  });

  it("highlights the hovered element with a positioned box", () => {
    const btn = appButton();
    vi.spyOn(btn, "getBoundingClientRect").mockReturnValue({
      left: 10,
      top: 20,
      width: 100,
      height: 40,
      right: 110,
      bottom: 60,
      x: 10,
      y: 20,
      toJSON() {},
    });
    const w = mount({ storage: mockStorage() });
    w.arm();
    btn.dispatchEvent(new MouseEvent("mousemove", { bubbles: true }));

    const box = document.querySelector("[data-squawk-highlight]");
    expect(box).toBeTruthy();
    expect(box.style.display).toBe("block");
    expect(box.style.left).toBe("10px");
    expect(box.style.top).toBe("20px");
    expect(box.style.width).toBe("100px");
    expect(box.style.height).toBe("40px");
  });

  it("ignores events targeting the widget itself", () => {
    const w = mount({ storage: mockStorage() });
    w.arm();
    const fab = findWidget().querySelector(".fab");
    forceTargetEvent("mousemove", fab);
    expect(document.querySelector("[data-squawk-highlight]").style.display).toBe("none");
  });

  it("capture-phase click selects the element and suppresses the app handler", () => {
    const btn = appButton();
    const appHandler = vi.fn();
    btn.addEventListener("click", appHandler);
    const w = mount({ storage: mockStorage() });
    w.arm();

    btn.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));

    expect(appHandler).not.toHaveBeenCalled();
    const panel = findPanel(findWidget());
    expect(panel.classList.contains("anchored")).toBe(true);
    expect(document.querySelector("[data-squawk-highlight]").style.display).toBe("block");
  });

  it("removes capture listeners after a selection", () => {
    const btn = appButton();
    const appHandler = vi.fn();
    btn.addEventListener("click", appHandler);
    const w = mount({ storage: mockStorage() });
    w.arm();
    btn.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    expect(appHandler).not.toHaveBeenCalled();

    btn.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    expect(appHandler).toHaveBeenCalledTimes(1);
  });

  it("saving a picked flag records selector, tagName, textPreview, and screenLabel", () => {
    document.body.innerHTML = "<h1>Reports</h1>";
    const storage = mockStorage();
    const w = mount({ storage });
    const btn = appButton("Menos de 6 meses");
    w.arm();

    btn.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    saveNote("wrong default selected");

    const stored = JSON.parse(storage.getItem("squawk:flags"))[0];
    expect(stored.screenLabel).toBe("Reports");
    expect(stored.element).toEqual({
      selector: "button.btn.text-green",
      tagName: "button",
      textPreview: "Menos de 6 meses",
    });
    expect(stored).not.toHaveProperty("position");
  });

  it("add() without an element keeps flags element-free", () => {
    const storage = mockStorage();
    const w = mount({ storage });
    const flag = w.add("plain note");
    expect(flag).not.toHaveProperty("element");
    expect(JSON.parse(storage.getItem("squawk:flags"))[0]).not.toHaveProperty("element");
  });

  it("Escape cancels a pending selection", () => {
    const btn = appButton();
    const storage = mockStorage();
    const w = mount({ storage });
    w.arm();
    btn.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));

    expect(document.querySelector("[data-squawk-highlight]").style.display).toBe("none");
    saveNote("no element note");
    const stored = JSON.parse(storage.getItem("squawk:flags"))[0];
    expect(stored).not.toHaveProperty("element");
  });

  it("unmount removes the highlight box", () => {
    const btn = appButton();
    const w = mount({ storage: mockStorage() });
    w.arm();
    btn.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    w.unmount();
    expect(document.querySelector("[data-squawk-highlight]")).toBeNull();
  });
});

describe("filename field", () => {
  function filenameInput() {
    return findPanel(findWidget()).querySelector(".file input");
  }

  it("pre-fills the field with a timestamped name", () => {
    mount({ storage: mockStorage() });
    expect(filenameInput().value).toMatch(/^bugs-\d{8}-\d{6}-\d{3}\.md$/);
  });

  it("export uses the current field value", () => {
    const w = mount({ storage: mockStorage() });
    w.add("a bug");
    const input = filenameInput();
    input.value = "qa-notes.md";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    const anchor = stubAnchorDownload();
    w.exportMarkdown();
    expect(anchor.download).toBe("qa-notes.md");
  });

  it("mount filename option pre-fills and is used", () => {
    const w = mount({ storage: mockStorage(), filename: "fixed.md" });
    w.add("a bug");
    expect(filenameInput().value).toBe("fixed.md");
    const anchor = stubAnchorDownload();
    w.exportMarkdown();
    expect(anchor.download).toBe("fixed.md");
  });

  it("falls back to a timestamped name when the field is empty", () => {
    const w = mount({ storage: mockStorage() });
    w.add("a bug");
    const input = filenameInput();
    input.value = "";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    const anchor = stubAnchorDownload();
    w.exportMarkdown();
    expect(anchor.download).toMatch(/^bugs-\d{8}-\d{6}-\d{3}\.md$/);
  });

  it("untouched repeated exports never reuse the same name", () => {
    let t = 1700000000000;
    vi.spyOn(Date, "now").mockImplementation(() => (t += 1000));
    const w = mount({ storage: mockStorage() });
    w.add("a bug");
    const anchor = stubAnchorDownload();
    w.exportMarkdown();
    const first = anchor.download;
    w.exportMarkdown();
    const second = anchor.download;
    expect(first).toMatch(/^bugs-\d{8}-\d{6}-\d{3}\.md$/);
    expect(second).toMatch(/^bugs-\d{8}-\d{6}-\d{3}\.md$/);
    expect(first).not.toBe(second);
  });
});

describe("panel", () => {
  it("toggles open and closed", () => {
    const widget = mount({ storage: mockStorage() });
    const root = findWidget();
    const panel = findPanel(root);

    expect(panel.classList.contains("open")).toBe(false);
    widget.open();
    expect(panel.classList.contains("open")).toBe(true);
    widget.close();
    expect(panel.classList.contains("open")).toBe(false);
  });
});