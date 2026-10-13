// Runs svelte-check over the widget and fails only on diagnostics in this workspace.
// The TypeScript program also pulls in the web-common modules the widget imports, and
// web-common carries a few known type errors that the repo tracks separately
// (scripts/tsc-with-whitelist.sh); those are listed as context, not as failures.
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("..", import.meta.url));
const result = spawnSync(
  "npx",
  ["svelte-check", "--tsconfig", "./tsconfig.json", "--output", "machine"],
  { cwd: root, encoding: "utf8", maxBuffer: 64 * 1024 * 1024 },
);
const output = `${result.stdout ?? ""}${result.stderr ?? ""}`;

const own = [];
const upstream = [];
for (const line of output.split("\n")) {
  const match = line.match(/^\d+ (ERROR|WARNING) "([^"]+)" (\d+:\d+) (".*")$/);
  if (!match) continue;
  const [, level, file, position, message] = match;
  const entry = `${level} ${file}:${position} ${JSON.parse(message)}`;
  (file.startsWith("../") ? upstream : own).push(entry);
}

if (upstream.length) {
  console.log(`${upstream.length} diagnostic(s) in imported web-common files (not failing this check):`);
  for (const entry of upstream) console.log(`  ${entry}`);
}
if (own.length) {
  console.error(`${own.length} diagnostic(s) in web-mcp-app:`);
  for (const entry of own) console.error(`  ${entry}`);
  process.exit(1);
}
const completed = output.match(/COMPLETED (\d+) FILES/);
console.log(`svelte-check: ${completed ? completed[1] : "?"} files, no diagnostics in web-mcp-app`);
