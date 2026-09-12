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

function stubViewport({ innerWidth, innerHeight, scrollX, scrollY, scrollWidth, scrollHeight }) {
  const orig = {
    innerWidth: window.innerWidth,
    innerHeight: window.innerHeight,
    scrollX: window.scrollX,
    scrollY: window.scrollY,
    scrollWidth: document.documentElement.scrollWidth,
    scrollHeight: document.documentElement.scrollHeight,
  };
  Object.defineProperty(window, "innerWidth", { value: innerWidth, configurable: true });
  Object.defineProperty(window, "innerHeight", { value: innerHeight, configurable: true });
  Object.defineProperty(window, "scrollX", { value: scrollX, configurable: true });
  Object.defineProperty(window, "scrollY", { value: scrollY, configurable: true });
  Object.defineProperty(document.documentElement, "scrollWidth", { value: scrollWidth, configurable: true });
  Object.defineProperty(document.documentElement, "scrollHeight", { value: scrollHeight, configurable: true });
  return () => {
    Object.defineProperty(window, "innerWidth", { value: orig.innerWidth, configurable: true });
    Object.defineProperty(window, "innerHeight", { value: orig.innerHeight, configurable: true });
    Object.defineProperty(window, "scrollX", { value: orig.scrollX, configurable: true });
    Object.defineProperty(window, "scrollY", { value: orig.scrollY, configurable: true });
    Object.defineProperty(document.documentElement, "scrollWidth", { value: orig.scrollWidth, configurable: true });
    Object.defineProperty(document.documentElement, "scrollHeight", { value: orig.scrollHeight, configurable: true });
  };
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
  it("add() writes a flag with id, note, url, and ISO timestamp", () => {
    const storage = mockStorage();
    const widget = mount({ storage });
    const flag = widget.add("save button unresponsive");

    expect(flag.id).toBe("F001");
    expect(flag.note).toBe("save button unresponsive");
    expect(flag.url).toBe("/");
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

describe("pin to locate", () => {
  let restore;

  beforeEach(() => {
    restore = stubViewport({
      innerWidth: 800,
      innerHeight: 600,
      scrollX: 50,
      scrollY: 200,
      scrollWidth: 2000,
      scrollHeight: 1000,
    });
  });

  afterEach(() => {
    restore();
  });

  function clickOverlay(clientX, clientY) {
    const overlay = document.querySelector("[data-squawk-overlay]");
    expect(overlay).toBeTruthy();
    overlay.dispatchEvent(
      new MouseEvent("click", { bubbles: true, cancelable: true, clientX, clientY })
    );
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

  it("add() persists a position passed in", () => {
    const storage = mockStorage();
    const widget = mount({ storage });
    const position = { xPercent: 0.25, yPercent: 0.5, viewportWidth: 800, viewportHeight: 600 };
    const flag = widget.add("broken chart", position);
    expect(flag.position).toEqual(position);
    expect(JSON.parse(storage.getItem("squawk:flags"))[0].position).toEqual(position);
  });

  it("add() without a position keeps flags position-free", () => {
    const storage = mockStorage();
    const widget = mount({ storage });
    const flag = widget.add("plain note");
    expect("position" in flag).toBe(false);
    expect(JSON.parse(storage.getItem("squawk:flags"))[0]).not.toHaveProperty("position");
  });

  it("arming shows a full-page overlay and Escape disarms it", () => {
    const widget = mount({ storage: mockStorage() });
    widget.arm();
    expect(document.querySelector("[data-squawk-overlay]")).toBeTruthy();

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(document.querySelector("[data-squawk-overlay]")).toBeNull();
    expect(widget.getFlags()).toEqual([]);
  });

  it("clicking the overlay places a pin, anchors the panel, and the click does not reach the page", () => {
    const pageClicks = vi.fn();
    document.addEventListener("click", pageClicks);
    try {
      const storage = mockStorage();
      const w = mount({ storage });
      w.arm();

      clickOverlay(400, 100);

      expect(document.querySelector("[data-squawk-overlay]")).toBeNull();
      expect(document.querySelectorAll("[data-squawk-pin]")).toHaveLength(1);

      const panel = findPanel(findWidget());
      expect(panel.classList.contains("anchored")).toBe(true);
      expect(panel.style.left).toBe("416px");
      expect(panel.style.top).toBe("116px");

      expect(pageClicks).not.toHaveBeenCalled();
    } finally {
      document.removeEventListener("click", pageClicks);
    }
  });

  it("saving a pinned flag records the click position with document-relative percentages", () => {
    const storage = mockStorage();
    const w = mount({ storage });
    w.arm();

    clickOverlay(400, 100);
    saveNote("save button unresponsive");

    const stored = JSON.parse(storage.getItem("squawk:flags"))[0];
    expect(stored.id).toBe("F001");
    expect(stored.position).toEqual({
      xPercent: 0.225,
      yPercent: 0.3,
      viewportWidth: 800,
      viewportHeight: 600,
    });
    expect(stored.position.xPercent).toBeCloseTo((400 + 50) / 2000, 5);
    expect(stored.position.yPercent).toBeCloseTo((100 + 200) / 1000, 5);
  });

  it("keeps the pin visible after save and cancels an unsaved pin when the panel closes", () => {
    const storage = mockStorage();
    const w = mount({ storage });
    w.arm();
    clickOverlay(400, 100);
    saveNote("kept");
    expect(document.querySelectorAll("[data-squawk-pin]")).toHaveLength(1);

    w.arm();
    clickOverlay(500, 200);
    expect(document.querySelectorAll("[data-squawk-pin]")).toHaveLength(2);
    findPanel(findWidget()).querySelector(".close").click();
    expect(document.querySelectorAll("[data-squawk-pin]")).toHaveLength(1);
  });

  it("clear removes live pins and the overlay", () => {
    const storage = mockStorage();
    const w = mount({ storage });
    w.add("one");
    w.arm();
    clickOverlay(400, 100);
    saveNote("pinned");
    confirmMock.mockReturnValue(true);
    findPanel(findWidget()).querySelector(".danger").click();

    expect(document.querySelectorAll("[data-squawk-pin]")).toHaveLength(0);
    expect(document.querySelector("[data-squawk-overlay]")).toBeNull();
  });

  it("unmount removes pins and overlay", () => {
    const w = mount({ storage: mockStorage() });
    w.arm();
    clickOverlay(400, 100);
    w.unmount();
    expect(document.querySelector("[data-squawk-overlay]")).toBeNull();
    expect(document.querySelector("[data-squawk-pins]")).toBeNull();
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