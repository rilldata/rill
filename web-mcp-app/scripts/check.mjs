// Runs svelte-check over the widget and fails only on diagnostics in this workspace.
// The TypeScript program also pulls in the web-common modules the widget imports, and
// web-common carries a few known type errors that the repo tracks separately
// (scripts/tsc-with-whitelist.sh); those are listed as context, not as failures.
// A run that does not complete (svelte-check missing, crashed or killed) fails the check:
// svelte-check exits 1 for any error, upstream ones included, so its status alone cannot tell.
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("..", import.meta.url));
const result = spawnSync(
  "npx",
  ["svelte-check", "--tsconfig", "./tsconfig.json", "--output", "machine"],
  { cwd: root, encoding: "utf8", maxBuffer: 64 * 1024 * 1024 },
);
const output = `${result.stdout ?? ""}${result.stderr ?? ""}`;

/** @returns {never} */
function fail(message) {
  console.error(`svelte-check: ${message}`);
  const tail = output.trim().split("\n").slice(-20).join("\n");
  if (tail) console.error(tail);
  process.exit(1);
}

if (result.error) fail(`could not run svelte-check: ${result.error.message}`);
if (result.signal) fail(`svelte-check was killed by ${result.signal}`);
const completed = output.match(/^\d+ COMPLETED (\d+) FILES (\d+) ERRORS/m);
if (!completed) {
  fail(`svelte-check did not complete (exit status ${result.status})`);
}
const [, files, errors] = completed;
if (Number(files) === 0) fail("svelte-check checked no files");

const own = [];
const upstream = [];
for (const line of output.split("\n")) {
  const match = line.match(/^\d+ (ERROR|WARNING) "([^"]+)" (\d+:\d+) (".*")$/);
  if (!match) continue;
  const [, level, file, position, message] = match;
  const entry = `${level} ${file}:${position} ${JSON.parse(message)}`;
  (file.startsWith("../") ? upstream : own).push(entry);
}
// Exit 1 without a listed error means it failed in a way the diagnostics do not show.
if (result.status !== 0 && Number(errors) === 0) {
  fail(`svelte-check exited with status ${result.status} and no errors`);
}

if (upstream.length) {
  console.log(
    `${upstream.length} diagnostic(s) in imported web-common files (not failing this check):`,
  );
  for (const entry of upstream) console.log(`  ${entry}`);
}
if (own.length) {
  console.error(`${own.length} diagnostic(s) in web-mcp-app:`);
  for (const entry of own) console.error(`  ${entry}`);
  process.exit(1);
}
console.log(`svelte-check: ${files} files, no diagnostics in web-mcp-app`);
