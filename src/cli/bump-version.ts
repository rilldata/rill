import { readFileSync, writeFileSync } from "fs";
import semverInc from "semver/functions/inc";
import { execSync } from "node:child_process";

const BUMP_TYPES = {
  major: 1,
  minor: 1,
  patch: 1,
};
const PACKAGE_JSON_FILE = "./package.json";
const PACKAGE_LOCK_JSON_FILE = "./package-lock.json";
const execSyncToStdout = (cmd) => execSync(cmd, { stdio: "inherit" });

const bumpType = process.argv[2];
if (!(bumpType in BUMP_TYPES)) {
  console.log(
    `Invalid version bump type: ${bumpType}. Can be one of ${Object.keys(
      BUMP_TYPES
    )}`
  );
  process.exit(1);
}

const packageJsonString = readFileSync(PACKAGE_JSON_FILE).toString();
const currentVersion = JSON.parse(packageJsonString).version;
const newVersion = semverInc(currentVersion, bumpType);

console.log(`Bumping version from ${currentVersion} to ${newVersion}`);
writeFileSync(
  PACKAGE_JSON_FILE,
  packageJsonString.replace(
    `"version": "${currentVersion}",`,
    `"version": "${newVersion}",`
  )
);
console.log("Regenerating `package-lock.json`");
execSyncToStdout(`npm install`);

const branch = "release";
console.log(`Pushing to ${branch}`);
execSyncToStdout(`git checkout ${branch}`);
execSyncToStdout(`git add ${PACKAGE_JSON_FILE} ${PACKAGE_LOCK_JSON_FILE}`);
execSyncToStdout(
  `git commit -m "Bump version: v${currentVersion} -> v${newVersion}"`
);
execSyncToStdout(`git push origin ${branch}`);

console.log(`Creating tag ${newVersion}`);
execSyncToStdout(`git tag -m "Release: v${newVersion}" v${newVersion}`);
execSyncToStdout(`git push --tags`);

console.log("Trying to create a github release");
execSyncToStdout(
  `gh release create v${newVersion} --notes "Release: v${newVersion}" -t "Release: v${newVersion}"`
);
