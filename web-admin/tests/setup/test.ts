import { test as base } from "@playwright/test";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { spawn } from "node:child_process";
import path from "path";
import treeKill from "tree-kill";
import { fileURLToPath } from "url";

const skipGlobalSetup = process.env.E2E_SKIP_GLOBAL_SETUP === "true";

// Global setup
base.beforeAll(async () => {
  if (skipGlobalSetup) return;

  const timeout = 120_000;
  base.setTimeout(timeout);

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

  process.env.CLOUD_PID = cloudProcess.pid?.toString();
});

base.afterAll(() => {
  const pid = process.env.CLOUD_PID;
  if (pid) {
    treeKill(parseInt(pid));
  }
});

export const test = base;
