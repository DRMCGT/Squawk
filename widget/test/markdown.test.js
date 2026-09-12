import { describe, expect, it } from "vitest";
import { noteTitle, toMarkdown, formatExportStamp, exportFilename } from "../src/markdown.js";

const flag = (overrides = {}) => ({
  id: "F001",
  note: "save button unresponsive",
  url: "/dashboard",
  timestamp: "2026-09-12T21:03:00Z",
  ...overrides,
});

const position = {
  xPercent: 0.42,
  yPercent: 0.18,
  viewportWidth: 1440,
  viewportHeight: 900,
};

describe("noteTitle", () => {
  it("uses the first eight words of the note", () => {
    const note = "one two three four five six seven eight nine ten";
    expect(noteTitle(note)).toBe("one two three four five six seven eight");
  });

  it("keeps shorter notes intact", () => {
    expect(noteTitle("save button broken")).toBe("save button broken");
  });

  it("returns empty string for empty or whitespace notes", () => {
    expect(noteTitle("")).toBe("");
    expect(noteTitle("   ")).toBe("");
  });
});

describe("formatExportStamp / exportFilename", () => {
  const at = new Date("2026-09-12T14:07:40.123Z");

  it("formats the timestamp without colons or slashes", () => {
    expect(formatExportStamp(at)).toBe("20260912-140740-123");
  });

  it("builds a filesystem-safe timestamped filename", () => {
    expect(exportFilename(at)).toBe("bugs-20260912-140740-123.md");
  });

  it("uses the same timestamp for the filename and the Exported line", () => {
    const md = toMarkdown([flag()], at);
    expect(md).toContain(`Exported: ${at.toISOString()}`);
    expect(exportFilename(at)).toBe("bugs-20260912-140740-123.md");
  });

  it("distinguishes exports in the same second by milliseconds", () => {
    const a = new Date("2026-09-12T14:07:40.001Z");
    const b = new Date("2026-09-12T14:07:40.999Z");
    expect(exportFilename(a)).not.toBe(exportFilename(b));
  });
});

describe("toMarkdown", () => {
  it("renders the report header with an exported timestamp", () => {
    const md = toMarkdown([], new Date("2026-09-12T22:00:00Z"));
    const lines = md.split("\n");
    expect(lines[0]).toBe("# Squawk Bug Report");
    expect(lines[1]).toBe("Exported: 2026-09-12T22:00:00.000Z");
  });

  it("renders an empty-report marker when there are no flags", () => {
    const md = toMarkdown([], new Date("2026-09-12T22:00:00Z"));
    expect(md).toContain("_No bugs flagged._");
  });

  it("renders each flag as a heading with page, time, and full note", () => {
    const md = toMarkdown(
      [flag()],
      new Date("2026-09-12T22:00:00Z")
    );
    expect(md).toContain("## F001 — save button unresponsive");
    expect(md).toContain("- Page: /dashboard");
    expect(md).toContain("- Time: 2026-09-12T21:03:00Z");
    expect(md).toContain("save button unresponsive");
  });

  it("renders the Position line for positioned flags", () => {
    const md = toMarkdown([flag({ position })], new Date("2026-09-12T22:00:00Z"));
    expect(md).toContain("- Page: /dashboard");
    expect(md).toContain("- Time: 2026-09-12T21:03:00Z");
    expect(md).toContain("- Position: 42% from left, 18% from top (viewport 1440x900)");
  });

  it("omits the Position line for flags without a position", () => {
    const md = toMarkdown([flag()], new Date("2026-09-12T22:00:00Z"));
    expect(md).not.toContain("Position:");
  });

  it("keeps full note text beyond the truncated title", () => {
    const note = "a b c d e f g h and the full remainder of the note";
    const md = toMarkdown([flag({ note })], new Date("2026-09-12T22:00:00Z"));
    expect(md).toContain("## F001 — a b c d e f g h");
    expect(md).toContain(note);
  });

  it("falls back to (no note) for empty notes", () => {
    const md = toMarkdown([flag({ note: "" })], new Date("2026-09-12T22:00:00Z"));
    expect(md).toContain("## F001 — (no note)");
  });

  it("preserves flag order", () => {
    const md = toMarkdown(
      [flag({ id: "F001", note: "first" }), flag({ id: "F002", note: "second" })],
      new Date("2026-09-12T22:00:00Z")
    );
    const i1 = md.indexOf("## F001 — first");
    const i2 = md.indexOf("## F002 — second");
    expect(i1).toBeGreaterThan(-1);
    expect(i2).toBeGreaterThan(i1);
  });

  it("ends with a trailing newline", () => {
    const md = toMarkdown([flag()], new Date("2026-09-12T22:00:00Z"));
    expect(md.endsWith("\n")).toBe(true);
  });
});