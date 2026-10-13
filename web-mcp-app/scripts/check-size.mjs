// Fails the build when the single-file widget exceeds the gzip budget.
// Hosts fetch the resource once per connection and cache it, so this is a
// first-render cost, but it still has to stay comfortably below Claude's
// tool-result limits and mobile WebView comfort.
import { readFileSync } from "node:fs";
import { gzipSync } from "node:zlib";

const BUDGET_BYTES = 600 * 1024;
const file = new URL("../dist/widget.html", import.meta.url);

const raw = readFileSync(file);
const gzipped = gzipSync(raw, { level: 9 }).length;
const kb = (n) => `${(n / 1024).toFixed(1)} KB`;

console.log(
  `dist/widget.html: ${kb(raw.length)} raw, ${kb(gzipped)} gzipped (budget ${kb(BUDGET_BYTES)})`,
);

if (gzipped > BUDGET_BYTES) {
  console.error(
    `widget.html exceeds the ${kb(BUDGET_BYTES)} gzip budget by ${kb(gzipped - BUDGET_BYTES)}`,
  );
  process.exit(1);
}
