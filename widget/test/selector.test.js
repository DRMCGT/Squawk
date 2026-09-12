import { afterEach, describe, expect, it } from "vitest";
import { buildSelector, describeElement, escapeCss, textPreview } from "../src/selector.js";

afterEach(() => {
  document.body.innerHTML = "";
});

describe("buildSelector", () => {
  it("prefers a single id", () => {
    document.body.innerHTML = '<div id="app"><button id="save-btn">Save</button></div>';
    const btn = document.querySelector("#save-btn");
    expect(buildSelector(btn)).toBe("#save-btn");
  });

  it("builds a tag + class path", () => {
    document.body.innerHTML =
      '<div class="card"><section class="body"><p class="title">hi</p></section></div>';
    const p = document.querySelector("p");
    expect(buildSelector(p)).toBe("div.card > section.body > p.title");
  });

  it("uses only the first two usable classes", () => {
    document.body.innerHTML = '<div class="a b c"></div>';
    const div = document.querySelector("div");
    expect(buildSelector(div)).toBe("div.a.b");
  });

  it("disambiguates siblings with :nth-child", () => {
    document.body.innerHTML = "<ul><li>a</li><li>b</li><li>c</li></ul>";
    const li = document.querySelectorAll("li")[1];
    expect(buildSelector(li)).toBe("ul > li:nth-child(2)");
  });

  it("does not add :nth-child for unique siblings", () => {
    document.body.innerHTML = "<ul><li>a</li><p>x</p></ul>";
    const li = document.querySelector("li");
    expect(buildSelector(li)).toBe("ul > li");
  });

  it("uses a nearest ancestor id to root the path", () => {
    document.body.innerHTML = '<div id="app"><section><button>Save</button></section></div>';
    const btn = document.querySelector("button");
    expect(buildSelector(btn)).toBe("#app > section > button");
  });

  it("handles body and html roots", () => {
    expect(buildSelector(document.body)).toBe("body");
    expect(buildSelector(document.documentElement)).toBe("html");
  });

  it("escapes ids with special characters", () => {
    document.body.innerHTML = '<div id="panel:1"></div>';
    const div = document.querySelector("div");
    expect(buildSelector(div)).toBe("#" + escapeCss("panel:1"));
  });

  it("returns an empty string for non-elements", () => {
    expect(buildSelector(null)).toBe("");
    expect(buildSelector(document.createTextNode("x"))).toBe("");
  });
});

describe("textPreview", () => {
  it("collapses whitespace and trims", () => {
    const el = document.createElement("p");
    el.textContent = "  hello \n\n world  ";
    expect(textPreview(el)).toBe("hello world");
  });

  it("truncates long text to 50 chars with an ellipsis", () => {
    const el = document.createElement("p");
    el.textContent = "x".repeat(80);
    expect(textPreview(el)).toBe("x".repeat(50) + "…");
  });

  it("returns empty string for empty content", () => {
    const el = document.createElement("p");
    expect(textPreview(el)).toBe("");
  });
});

describe("describeElement", () => {
  it("returns selector, tagName and textPreview", () => {
    document.body.innerHTML =
      '<div id="app"><button class="btn text-green">Menos de 6 meses</button></div>';
    const btn = document.querySelector("button");
    expect(describeElement(btn)).toEqual({
      selector: "#app > button.btn.text-green",
      tagName: "button",
      textPreview: "Menos de 6 meses",
    });
  });

  it("returns null for non-elements", () => {
    expect(describeElement(null)).toBeNull();
  });
});

describe("escapeCss", () => {
  it("escapes special characters", () => {
    expect(escapeCss("a:b")).toBe("a\\:b");
    expect(escapeCss("a.b")).toBe("a\\.b");
  });

  it("leaves safe identifiers untouched", () => {
    expect(escapeCss("save-btn_1")).toBe("save-btn_1");
  });
});