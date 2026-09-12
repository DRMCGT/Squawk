import { describe, expect, it } from "vitest";
import { createStore, nextFlagId, STORAGE_KEY } from "../src/store.js";

function mockStorage(initial = {}) {
  const map = new Map(Object.entries(initial));
  return {
    getItem: (k) => (map.has(k) ? map.get(k) : null),
    setItem: (k, v) => map.set(k, String(v)),
    removeItem: (k) => map.delete(k),
  };
}

const validFlag = (overrides = {}) => ({
  id: "F001",
  note: "save button unresponsive",
  url: "/dashboard",
  timestamp: "2026-09-12T21:03:00Z",
  ...overrides,
});

describe("createStore", () => {
  it("round-trips flags through storage", () => {
    const storage = mockStorage();
    const store = createStore({ storage });
    const flags = [validFlag()];
    store.save(flags);
    expect(JSON.parse(storage.getItem(STORAGE_KEY))).toEqual(flags);
    expect(store.load()).toEqual(flags);
  });

  it("uses a custom storage key", () => {
    const storage = mockStorage();
    const store = createStore({ storage, storageKey: "app:flags" });
    store.save([validFlag()]);
    expect(storage.getItem("app:flags")).toBeTruthy();
    expect(storage.getItem(STORAGE_KEY)).toBeNull();
  });

  it("returns [] when nothing is stored", () => {
    const store = createStore({ storage: mockStorage() });
    expect(store.load()).toEqual([]);
  });

  it("returns [] for malformed JSON", () => {
    const store = createStore({ storage: mockStorage({ [STORAGE_KEY]: "{oops" }) });
    expect(store.load()).toEqual([]);
  });

  it("returns [] for a non-array payload", () => {
    const store = createStore({ storage: mockStorage({ [STORAGE_KEY]: '{"a":1}' }) });
    expect(store.load()).toEqual([]);
  });

  it("drops invalid entries and keeps valid ones", () => {
    const stored = [validFlag(), { id: 42, note: null }, "nope", { id: "F099" }];
    const store = createStore({
      storage: mockStorage({ [STORAGE_KEY]: JSON.stringify(stored) }),
    });
    expect(store.load()).toEqual([validFlag()]);
  });

  it("round-trips flags with a position", () => {
    const storage = mockStorage();
    const store = createStore({ storage });
    const positioned = {
      ...validFlag(),
      position: {
        xPercent: 0.42,
        yPercent: 0.18,
        viewportWidth: 1440,
        viewportHeight: 900,
      },
    };
    store.save([positioned]);
    expect(store.load()).toEqual([positioned]);
  });

  it("keeps flags without a position", () => {
    const store = createStore({ storage: mockStorage() });
    store.save([validFlag()]);
    expect(store.load()).toEqual([validFlag()]);
  });

  it("drops flags with a malformed position", () => {
    const stored = [
      validFlag(),
      { ...validFlag({ id: "F002" }), position: { xPercent: "oops" } },
      { ...validFlag({ id: "F003" }), position: { viewportWidth: 1 } },
    ];
    const store = createStore({
      storage: mockStorage({ [STORAGE_KEY]: JSON.stringify(stored) }),
    });
    expect(store.load()).toEqual([validFlag()]);
  });

  it("round-trips flags with an element", () => {
    const storage = mockStorage();
    const store = createStore({ storage });
    const withElement = {
      ...validFlag(),
      screenLabel: "Reports",
      element: {
        selector: "button.btn-primary",
        tagName: "button",
        textPreview: "Save changes",
      },
    };
    store.save([withElement]);
    expect(store.load()).toEqual([withElement]);
  });

  it("keeps legacy position flags loadable", () => {
    const legacy = {
      ...validFlag(),
      position: { xPercent: 0.5, yPercent: 0.5, viewportWidth: 800, viewportHeight: 600 },
    };
    const store = createStore({
      storage: mockStorage({ [STORAGE_KEY]: JSON.stringify([legacy]) }),
    });
    expect(store.load()).toEqual([legacy]);
  });

  it("drops flags with a malformed element", () => {
    const stored = [
      validFlag(),
      { ...validFlag({ id: "F002" }), element: { selector: 5 } },
      { ...validFlag({ id: "F003" }), element: { tagName: "button" } },
    ];
    const store = createStore({
      storage: mockStorage({ [STORAGE_KEY]: JSON.stringify(stored) }),
    });
    expect(store.load()).toEqual([validFlag()]);
  });

  it("clear removes the key", () => {
    const storage = mockStorage({ [STORAGE_KEY]: "[]" });
    const store = createStore({ storage });
    store.clear();
    expect(storage.getItem(STORAGE_KEY)).toBeNull();
  });

  it("falls back to in-memory storage when localStorage is unavailable", () => {
    const store = createStore({ storage: null });
    store.save([validFlag()]);
    expect(store.load()).toHaveLength(1);
  });
});

describe("nextFlagId", () => {
  it("starts at F001", () => {
    expect(nextFlagId([])).toBe("F001");
  });

  it("increments from existing ids", () => {
    expect(nextFlagId([{ id: "F001" }, { id: "F002" }])).toBe("F003");
  });

  it("steps over gaps using the max id", () => {
    expect(nextFlagId([{ id: "F001" }, { id: "F007" }])).toBe("F008");
  });

  it("ignores ids that do not match the F pattern", () => {
    expect(nextFlagId([{ id: "foo" }, { id: "X1" }])).toBe("F001");
  });

  it("pads to three digits", () => {
    expect(nextFlagId([{ id: "F099" }])).toBe("F100");
    expect(nextFlagId([{ id: "F999" }])).toBe("F1000");
  });
});