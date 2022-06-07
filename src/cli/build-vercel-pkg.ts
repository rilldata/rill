import "../moduleAlias";
import os from "node:os";
import { execSync } from "node:child_process";

const NodeJSVersion = "node16";
const CLIPath = "dist/cli/data-modeler-cli.js";

const nodePlatform = os.platform();
const cpuArch = os.arch();

// NodeJS platform to vercel/pkg's platform map
const PlatformMap: {
  [platform in NodeJS.Platform]?: string;
} = {
  darwin: "macos",
  linux: "linux",
  win32: "win",
};

const CPUArchAllowList = {
  x64: true,
  arm64: true,
};

if (!(nodePlatform in PlatformMap) || !(cpuArch in CPUArchAllowList)) {
  console.error(`${nodePlatform} and ${cpuArch} is not supported`);
  process.exit(1);
}
const platform = PlatformMap[nodePlatform];
const binaryPath = `rilldata/rill-${platform}-${cpuArch}`;

execSync(
  `npx pkg -c package.json --compress GZip ` +
    `-t ${NodeJSVersion}-${platform}-${cpuArch} ` +
    `-o ${binaryPath} ${CLIPath}`,
  { stdio: "inherit" }
);
console.log(`Generated binary for ${platform} ${cpuArch}`);
console.log(execSync(`ls -ltr rilldata`).toString());
