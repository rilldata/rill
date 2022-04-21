import {readFileSync, writeFileSync} from "fs";
import semverInc from "semver/functions/inc";

const BUMP_TYPES = {
    "major": 1,
    "minor": 1,
    "patch": 1,
};
const PACKAGE_JSON_FILE = "./package.json";

const packageJsonString = readFileSync(PACKAGE_JSON_FILE).toString();
const currentVersion = JSON.parse(packageJsonString).version;
const bumpType = process.argv[2];

if (!(bumpType in BUMP_TYPES)) {
    console.log(`Invalid version bump type: ${bumpType}. Can be one of ${Object.keys(BUMP_TYPES)}`);
    process.exit(1);
}

const newVersion = semverInc(currentVersion, bumpType);

console.log(`Bumping version from ${currentVersion} to ${newVersion}`);
writeFileSync(PACKAGE_JSON_FILE, packageJsonString
    .replace(`"version": "${currentVersion}",`, `"version": "${newVersion}",`));
