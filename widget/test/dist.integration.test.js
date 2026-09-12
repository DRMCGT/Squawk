import { existsSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { JSDOM } from "jsdom";
import { describe, expect, it } from "vitest";

const here = dirname(fileURLToPath(import.meta.url));
const bundle = join(here, "..", "dist", "squawk.js");
const hasBundle = existsSync(bundle);

function readBlob(win, blob) {
  return new Promise((resolve, reject) => {
    const fr = new win.FileReader();
    fr.onload = () => resolve(fr.result);
    fr.onerror = reject;
    fr.readAsText(blob);
  });
}

describe.skipIf(!hasBundle)("built plain-script bundle", () => {
  it("drives the full element-picker flag -> localStorage -> export loop via the Squawk global", async () => {
    const dom = new JSDOM(
      '<!doctype html><html><head><title>Cillus</title></head><body><h1>Reports</h1><button class="btn text-green">Menos de 6 meses</button></body></html>',
      {
        runScripts: "outside-only",
        url: "http://localhost:3000/dashboard?tab=reports",
      }
    );
    const win = dom.window;

    const created = [];
    win.URL.createObjectURL = (blob) => {
      created.push(blob);
      return "blob:mock-url";
    };
    win.URL.revokeObjectURL = () => {};
    win.confirm = () => true;
    win.HTMLAnchorElement.prototype.click = function () {
      // record the download name instead of navigating
    };

    win.eval(readFileSync(bundle, "utf8"));

    expect(typeof win.Squawk).toBe("object");
    expect(typeof win.Squawk.mount).toBe("function");
    expect(typeof win.Squawk.mountSquawk).toBe("function");

    const widget = win.Squawk.mount();

    // plain note
    const f1 = widget.add("save button unresponsive");
    expect(f1.id).toBe("F001");
    expect(f1.url).toBe("/dashboard?tab=reports");
    expect(f1.screenLabel).toBe("Reports");

    // element picker flow
    const btn = win.document.querySelector("button");
    widget.arm();
    btn.dispatchEvent(new win.MouseEvent("click", { bubbles: true, cancelable: true }));
    const shadow = win.document.querySelector("[data-squawk-widget]").shadowRoot;
    const textarea = shadow.querySelector("textarea");
    textarea.value = "wrong default selected";
    textarea.dispatchEvent(new win.Event("input", { bubbles: true }));
    shadow.querySelector(".save").click();

    const stored = JSON.parse(win.localStorage.getItem("squawk:flags"));
    expect(stored).toHaveLength(2);
    expect(stored[1].element).toEqual({
      selector: "button.btn.text-green",
      tagName: "button",
      textPreview: "Menos de 6 meses",
    });

    // custom export filename
    const fileInput = shadow.querySelector(".file input");
    fileInput.value = "qa-cillus.md";
    fileInput.dispatchEvent(new win.Event("input", { bubbles: true }));

    widget.exportMarkdown();
    expect(created).toHaveLength(1);
    const md = await readBlob(win, created[0]);

    expect(md).toContain("# Squawk Bug Report");
    expect(md).toContain("## F001 — save button unresponsive");
    expect(md).toContain("- Screen: Reports");
    expect(md).toContain("- Page: /dashboard?tab=reports");
    expect(md).toContain("## F002 — wrong default selected");
    expect(md).toContain('- Element: `button.btn.text-green` — "Menos de 6 meses"');
    expect(md).toContain("- Time: ");
    expect(md).toContain("save button unresponsive");

    widget.unmount();
    expect(win.document.querySelector("[data-squawk-widget]")).toBeNull();
  });

  it("clear wipes localStorage after confirm", () => {
    const dom = new JSDOM("<!doctype html><html><body></body></html>", {
      runScripts: "outside-only",
      url: "http://localhost:3000/",
    });
    const win = dom.window;
    win.confirm = () => true;
    win.URL.createObjectURL = () => "blob:mock-url";
    win.URL.revokeObjectURL = () => {};

    win.eval(readFileSync(bundle, "utf8"));
    const widget = win.Squawk.mount();
    widget.add("one");
    expect(win.localStorage.getItem("squawk:flags")).toBeTruthy();

    widget.clear();
    expect(win.localStorage.getItem("squawk:flags")).toBeNull();
    expect(widget.getFlags()).toEqual([]);
  });
});