import { expect } from "@playwright/test";
import { spawnAndMatch } from "@rilldata/web-common/tests/utils/spawn";
import { spawn } from "child_process";
import dotenv from "dotenv";
import { openSync } from "fs";
import { mkdir } from "fs/promises";
import path from "path";
import { fileURLToPath } from "url";
import { writeFileEnsuringDir } from "@rilldata/web-common/tests/utils/fs";
import { test as setup } from "./base";
import {
  RILL_DEV_STORAGE_STATE,
  RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE,
} from "../constants";
import { isServiceReady } from "@rilldata/web-common/tests/utils/is-service-ready";

setup.describe("global setup", () => {
  setup.describe.configure({
    mode: "serial",
    timeout: 180_000,
  });

  setup("should start services", async () => {
    // Get the repository root directory, the only place from which `rill devtool` is allowed to be run
    const currentDir = path.dirname(fileURLToPath(import.meta.url));
    const repoRoot = path.resolve(currentDir, "../../../");

    // Start the cloud dependencies via Docker
    // This will block until the services are ready
    await spawnAndMatch(
      "rill",
      ["devtool", "start", "other", "--reset", "--only", "deps"],
      /All services ready/,
      {
        cwd: repoRoot,
        timeoutMs: 60_000,
      },
    );

    // Load environment variables from our root `.env` file
    const __dirname = path.dirname(fileURLToPath(import.meta.url));
    dotenv.config({ path: path.resolve(__dirname, "../../../.env") });

    // Check that the required environment variables are set
    // The above `rill devtool` command pulls the `.env` file with these values.
    // Fail quickly if any of these are missing.
    if (
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    // Setup a log file to capture the output of the admin and runtime services
    await mkdir("playwright/logs", { recursive: true });
    const logPath = path.resolve("playwright/logs/admin-runtime.log");
    const logFd = openSync(logPath, "w");

    // Start the admin and runtime services in a detached background process.
    // A detached process ensures they are not cleaned up when this setup project completes.
    // However, we need to be sure to clean-up the processes manually in the teardown project.
    const child = spawn(
      "rill",
      ["devtool", "start", "other", "--only", "admin,runtime"],
      {
        detached: true,
        stdio: ["ignore", logFd, logFd],
        cwd: repoRoot,
      },
    );
    child.unref();

    // Write the pid to a file, so I can kill it later
    if (child.pid) {
      writeFileEnsuringDir(
        RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE,
        child.pid.toString(),
      );
    } else {
      throw new Error("Failed to get pid of child process");
    }

    // Wait for the admin service to be ready
    await expect
      .poll(() => isServiceReady("http://localhost:8080/v1/ping"), {
        timeout: 45_000,
      })
      .toBeTruthy();
    console.log("Admin service ready");

    // Wait for the runtime service to be ready
    await expect
      .poll(() => isServiceReady("http://localhost:8081/v1/ping"), {
        timeout: 20_000,
      })
      .toBeTruthy();
    console.log("Runtime service ready");
  });

  setup("should log in with the admin account", async ({ page }) => {
    // Again, check that the required environment variables are set. This is for type-safety.
    if (
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    // Log in with the admin account
    await page.goto("/");

    // Fill in the email
    const emailInput = page.locator('input[name="username"]');
    await emailInput.waitFor({ state: "visible" });
    await emailInput.click();
    await emailInput.fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL);

    // Click the continue button
    await page
      .locator('button[type="submit"][data-action-button-primary="true"]', {
        hasText: "Continue",
      })
      .click();

    // Fill in the password
    const passwordInput = page.locator('input[name="password"]');
    await passwordInput.waitFor({ state: "visible" });
    await passwordInput.click();
    await passwordInput.fill(
      process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD,
    );

    // Click the continue button
    await page
      .locator('button[type="submit"][data-action-button-primary="true"]', {
        hasText: "Continue",
      })
      .click();

    await page.waitForURL("/");

    // Save the admin's Rill auth cookies to file.
    // Subsequent tests can seed their browser with this state, instead of needing to go through the log-in flow again.
    await page.context().storageState({ path: RILL_DEV_STORAGE_STATE });
  });
});
