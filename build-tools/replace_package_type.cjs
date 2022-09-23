// eslint-disable-next-line @typescript-eslint/no-var-requires
const {readFileSync, writeFileSync} = require("fs");

/**
 * Temporary script to replace module type on windows
 */

function replacePackageJson(packageJsonPath) {
  let packageJson = readFileSync(packageJsonPath).toString();
  packageJson = packageJson.replace(
    new RegExp(`"type":\\s*"${process.argv[2]}",`, "g"),
    `"type": "${process.argv[3]}",`
  )
  writeFileSync(packageJsonPath, packageJson);
}

replacePackageJson("package.json");
replacePackageJson("web-local/package.json");
