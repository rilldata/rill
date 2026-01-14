#!/usr/bin/env node

import path from "node:path";
import fs from "node:fs";
import { spawn } from "node:child_process";

const projectArg = process.argv[2]?.toLowerCase();

const DEFAULT_PROJECT = "dev-project";
const testLocation = path.resolve("./web-common/tests/projects");

/** @type {Record<string, string>} */
const projectPaths = {
  adbids: path.join(testLocation, "AdBids"),
  openrtb: path.join(testLocation, "openrtb"),
  adimpressions: path.join(testLocation, "AdImpressions"),
  blank: path.join(testLocation, "Blank"),
};

/** @type {string} */
let projectDir;

if (!projectArg) {
  projectDir = DEFAULT_PROJECT;
} else if (projectPaths[projectArg]) {
  projectDir = projectPaths[projectArg];
} else {
  const resolvedPath = path.resolve(projectArg);

  if (!fs.existsSync(resolvedPath)) {
    console.error("Unknown project or invalid path:", projectArg);
    console.error("Valid options: adbids | openrtb | adimpressions | blank | <path>");
    process.exit(1);
  }

  projectDir = resolvedPath;
}

console.log("Starting runtime");
console.log(`Using project dir: ${projectDir}`);

const child = spawn(
  "go",
  ["run", "cli/main.go", "start", projectDir, "--no-ui", "--allowed-origins", "*"],
  { stdio: "inherit" }
);

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 1);
  }
});