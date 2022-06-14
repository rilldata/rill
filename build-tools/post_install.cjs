// eslint-disable-next-line @typescript-eslint/no-var-requires
const {execSync} = require("node:child_process");
// eslint-disable-next-line @typescript-eslint/no-var-requires
const {existsSync} = require("fs");

/**
 * Wrapper script to switch to either the source .ts or built .js script
 *
 * We check for presence of ts-node-dev and call npm run postinstall:dev
 * Else we call npm run postinstall:prod.
 *
 * We are using npm run because it will take care of calling the binray on different platforms for us
 */

if (existsSync("node_modules/.bin/ts-node-dev")) {
  execSync(
    "npm run postinstall:dev",
    {stdio: "inherit"}
  );
} else {
  execSync(
    "npm run postinstall:prod",
    {stdio: "inherit"}
  );
}
