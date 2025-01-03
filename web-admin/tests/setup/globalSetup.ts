import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { exec } from "child_process";
import { spawn } from "node:child_process";
import path from "path";
import { fileURLToPath } from "url";
import { promisify } from "util";

const execAsync = promisify(exec);

const skipGlobalSetup = Boolean(process.env.E2E_SKIP_GLOBAL_SETUP);
const timeout = 120_000;

export default async function globalSetup() {
  if (skipGlobalSetup) return;

  // Get the repository root directory, the only place from which `rill devtool` is allowed to be run
  const currentDir = path.dirname(fileURLToPath(import.meta.url));
  const repoRoot = path.resolve(currentDir, "../../../");

  // Start the cloud services (except for the UI, which is run by Playwright)
  const cloudProcess = spawn(
    "rill",
    ["devtool", "start", "e2e", "--reset", "--except", "ui"],
    {
      stdio: "pipe",
      cwd: repoRoot,
    },
  );

  // Capture output
  let logBuffer = "";
  cloudProcess.stdout?.on("data", (data) => {
    logBuffer += data.toString();
    console.log(data.toString());
  });

  cloudProcess.stderr?.on("data", (data) => {
    logBuffer += data.toString();
    console.error(data.toString());
  });

  // Wait for services to be ready
  const ready = await waitUntil(() => {
    return logBuffer.includes("All services ready");
  }, timeout);
  if (!ready) {
    throw new Error("Cloud services did not start in time");
  }

  // Pull the repositories to be used for testing
  await execAsync(
    "git clone https://github.com/rilldata/rill-examples.git tests/setup/git/repos/rill-examples",
  );
}
