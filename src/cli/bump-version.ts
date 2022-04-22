import {readFileSync, writeFileSync} from "fs";
import semverInc from "semver/functions/inc";
import {execSync} from "node:child_process";

const BUMP_TYPES = {
    "major": 1,
    "minor": 1,
    "patch": 1,
};
const PACKAGE_JSON_FILE = "./package.json";
const PACKAGE_LOCK_JSON_FILE = "./package-lock.json";
const execSyncToStdout = cmd => execSync(cmd, { stdio: "inherit" });

const packageJsonString = readFileSync(PACKAGE_JSON_FILE).toString();
const currentVersion = JSON.parse(packageJsonString).version;
const bumpType = process.argv[2];

if (!(bumpType in BUMP_TYPES)) {
    console.log(`Invalid version bump type: ${bumpType}. Can be one of ${Object.keys(BUMP_TYPES)}`);
    process.exit(1);
}

console.log("Pulling latest changes from main branch");
execSyncToStdout("git checkout main");
execSyncToStdout("git pull");

const newVersion = semverInc(currentVersion, bumpType);

console.log(`Bumping version from ${currentVersion} to ${newVersion}`);
writeFileSync(PACKAGE_JSON_FILE, packageJsonString
    .replace(`"version": "${currentVersion}",`, `"version": "${newVersion}",`));
writeFileSync(PACKAGE_LOCK_JSON_FILE, packageJsonString
    .replace(`"version": "${currentVersion}",`, `"version": "${newVersion}",`));

console.log("Creating new branch");
const branchName = `release-candidate-${newVersion}`;
execSyncToStdout(`git checkout -b ${branchName}`);
execSyncToStdout(`git add ${PACKAGE_JSON_FILE} ${PACKAGE_LOCK_JSON_FILE}`);
execSyncToStdout(`git push origin ${branchName}`);
