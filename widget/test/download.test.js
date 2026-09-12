import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { downloadMarkdown } from "../src/download.js";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
  document.body.innerHTML = "";
});

it("triggers a Blob download with the given filename", () => {
  const create = vi.fn(() => "blob:mock-url");
  const revoke = vi.fn();
  URL.createObjectURL = create;
  URL.revokeObjectURL = revoke;

  const anchor = document.createElement("a");
  anchor.click = vi.fn();
  const remove = vi.spyOn(anchor, "remove");
  const appendSpy = vi.spyOn(document.body, "appendChild");
  vi.spyOn(document, "createElement").mockImplementation((tag) =>
    tag === "a" ? anchor : document.createElement(tag)
  );

  downloadMarkdown("# Squawk Bug Report\n", "bugs.md");

  expect(create).toHaveBeenCalledTimes(1);
  const blob = create.mock.calls[0][0];
  expect(blob).toBeInstanceOf(Blob);
  expect(blob.type).toBe("text/markdown;charset=utf-8");

  expect(anchor.href).toBe("blob:mock-url");
  expect(anchor.download).toBe("bugs.md");
  expect(anchor.style.display).toBe("none");

  expect(appendSpy).toHaveBeenCalledWith(anchor);
  expect(anchor.click).toHaveBeenCalledTimes(1);

  vi.advanceTimersByTime(1000);
  expect(remove).toHaveBeenCalledTimes(1);
  expect(revoke).toHaveBeenCalledTimes(1);
  expect(revoke).toHaveBeenCalledWith("blob:mock-url");
  expect(document.body.contains(anchor)).toBe(false);
});

it("defaults to bugs.md as the filename", () => {
  URL.createObjectURL = vi.fn(() => "blob:mock-url");
  URL.revokeObjectURL = vi.fn();
  const anchor = document.createElement("a");
  anchor.click = vi.fn();
  vi.spyOn(document, "createElement").mockImplementation((tag) =>
    tag === "a" ? anchor : document.createElement(tag)
  );

  downloadMarkdown("x");

  expect(anchor.download).toBe("bugs.md");
});