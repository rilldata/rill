import { test as base } from "@playwright/test";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { spawn } from "node:child_process";
import treeKill from "tree-kill";

// Global setup
base.beforeAll(async () => {
  console.log("Starting cloud services...");
  const cloudProcess = spawn("rill", ["devtool", "start", "e2e", "--reset"], {
    stdio: "pipe",
  });

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

  // Wait for services
  await waitUntil(() => {
    return logBuffer.includes("All services ready");
  }, 1000);

  process.env.CLOUD_PID = cloudProcess.pid?.toString();
});

base.afterAll(() => {
  console.log("Stopping cloud services...");
  const pid = process.env.CLOUD_PID;
  if (pid) {
    treeKill(parseInt(pid));
  }
});

export const test = base;
