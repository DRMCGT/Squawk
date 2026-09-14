import { build } from "esbuild";
import { mkdirSync } from "node:fs";

mkdirSync("dist", { recursive: true });

const common = {
  entryPoints: ["src/index.js"],
  bundle: true,
  target: ["es2020"],
  logLevel: "info",
};

const iifeFooter = { js: "Squawk.mount = Squawk.mountSquawk;" };

const targets = [
  { ...common, format: "esm", outfile: "dist/squawk.esm.js" },
  { ...common, format: "esm", outfile: "dist/squawk.esm.min.js", minify: true },
  { ...common, format: "iife", globalName: "Squawk", outfile: "dist/squawk.js", footer: iifeFooter },
  { ...common, format: "iife", globalName: "Squawk", outfile: "dist/squawk.min.js", minify: true, footer: iifeFooter },
  { ...common, format: "cjs", outfile: "dist/squawk.cjs" },
];

for (const target of targets) {
  await build(target);
}

console.log("widget build complete");