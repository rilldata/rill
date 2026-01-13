#!/usr/bin/env node

const { execSync } = require("node:child_process");
const path = require("node:path");
const fs = require("node:fs");

const project = process.argv[2];

const DEFAULT_PROJECT = "dev-project";
const testLocation = path.resolve("./web-common/tests/projects");

const projectPaths = {
  adbids: path.join(testLocation, "AdBids"),
  openrtb: path.join(testLocation, "openrtb"),
  adimpressions: path.join(testLocation, "AdImpressions"),
  blank: path.join(testLocation, "blank"),
};


const projectDir = project
  ? projectPaths[project]
  : DEFAULT_PROJECT;


if (project && !projectDir) {
  console.error("Unknown project:", project);
  console.error("Valid options: adbids | openrtb | adimpressions | blank");
  process.exit(1);
}

console.log(`Starting runtime`);
console.log(`Using project dir: ${projectDir}`);

execSync(
  `go run cli/main.go start "${projectDir}" --no-ui --allowed-origins "*"`,
  { stdio: "inherit" }
);
