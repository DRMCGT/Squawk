import { readFileSync, statSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const file = join(here, "..", "dist", "squawk.min.js");

const bytes = statSync(file).size;
const limit = 15 * 1024;

console.log(`squawk.min.js: ${bytes} bytes (limit ${limit})`);
if (bytes > limit) {
  console.error(`FAIL: minified widget exceeds ${limit} bytes`);
  process.exit(1);
}
console.log("size check passed");