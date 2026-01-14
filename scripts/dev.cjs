#!/usr/bin/env node

const { execSync } = require("node:child_process");
const path = require("node:path");
const fs = require("node:fs");

const projectArg = process.argv[2].toLowerCase();

const DEFAULT_PROJECT = "dev-project";
const testLocation = path.resolve("./web-common/tests/projects");

const projectPaths = {
  adbids: path.join(testLocation, "AdBids"),
  openrtb: path.join(testLocation, "openrtb"),
  adimpressions: path.join(testLocation, "AdImpressions"),
  blank: path.join(testLocation, "Blank"),
};

let projectDir;

if (!projectArg) {
  projectDir = DEFAULT_PROJECT;
} else if (projectPaths[projectArg]) {
  // Named project
  projectDir = projectPaths[projectArg];
} else {
  // Treat argument as explicit path
  const resolvedPath = path.resolve(projectArg);

  if (!fs.existsSync(resolvedPath)) {
    console.error("Unknown project or invalid path:", projectArg);
    console.error("Valid options: adbids | openrtb | adimpressions | blank | <path>");
    process.exit(1);
  }

  projectDir = resolvedPath;
}

console.log(`Starting runtime`);
console.log(`Using project dir: ${projectDir}`);

execSync(
  `go run cli/main.go start "${projectDir}" --no-ui --allowed-origins "*"`,
  { stdio: "inherit" }
);
