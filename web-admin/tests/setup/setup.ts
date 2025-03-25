import { expect } from "@playwright/test";
import {
  execAsync,
  spawnAndMatch,
} from "@rilldata/web-common/tests/utils/spawn";
import axios from "axios";
import { spawn } from "child_process";
import dotenv from "dotenv";
import { openSync } from "fs";
import { mkdir } from "fs/promises";
import path from "path";
import { fileURLToPath } from "url";
import { writeFileEnsuringDir } from "../utils/fs";
import { test as setup } from "./base";
import {
  ADMIN_STORAGE_STATE,
  RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE,
  RILL_EMBED_SERVICE_TOKEN_FILE,
  RILL_ORG_NAME,
  RILL_PROJECT_NAME,
  RILL_SERVICE_NAME,
} from "./constants";
import { cliLogin } from "./fixtures/cli";

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
      ["devtool", "start", "e2e", "--reset", "--only", "deps"],
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
      ["devtool", "start", "e2e", "--only", "admin,runtime"],
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
    await page.getByRole("button", { name: "Continue with Email" }).click();
    await page.getByPlaceholder("Enter your email address").click();
    await page
      .getByPlaceholder("Enter your email address")
      .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL);
    await page.getByPlaceholder("Enter your email address").press("Tab");
    await page
      .getByPlaceholder("Enter your password")
      .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD);
    await page.getByRole("button", { name: "Continue with Email" }).click();
    await page.waitForURL("/");

    // Save the admin's Rill auth cookies to file.
    // Subsequent tests can seed their browser with this state, instead of needing to go through the log-in flow again.
    await page.context().storageState({ path: ADMIN_STORAGE_STATE });
  });

  setup("should create an organization and service", async ({ adminPage }) => {
    // Create an organization named "e2e"
    await cliLogin(adminPage);
    const { stdout: orgCreateStdout } = await execAsync(
      `rill org create ${RILL_ORG_NAME}`,
    );
    expect(orgCreateStdout).toContain("Created organization");

    // create service and write access token to file
    const { stdout: orgCreateService } = await execAsync(
      `rill service create ${RILL_SERVICE_NAME}`,
    );
    expect(orgCreateService).toContain("Created service");

    const serviceToken = orgCreateService.match(/Access token:\s+(\S+)/);
    writeFileEnsuringDir(RILL_EMBED_SERVICE_TOKEN_FILE, serviceToken![1]);

    // Go to the organization's page
    await adminPage.goto(`/${RILL_ORG_NAME}`);
    await expect(
      adminPage.getByRole("heading", { name: RILL_ORG_NAME }),
    ).toBeVisible();
  });

  setup("should deploy the OpenRTB project", async ({ adminPage }) => {
    // Deploy the OpenRTB project
    const { match } = await spawnAndMatch(
      "rill",
      [
        "deploy",
        "--path",
        "tests/setup/projects/openrtb",
        "--project",
        RILL_PROJECT_NAME,
        "--upload",
        "--interactive=false",
      ],
      /https?:\/\/[^\s]+/,
    );

    // Navigate to the project URL and expect to see the successful deployment
    const url = match[0];
    await adminPage.goto(url);
    await expect(adminPage.getByText(RILL_ORG_NAME)).toBeVisible(); // Organization breadcrumb
    await expect(adminPage.getByText(RILL_PROJECT_NAME)).toBeVisible(); // Project breadcrumb

    // Trial is started in an async job after the 1st deploy. It is not worth the effort to re-fetch the issues list right now.
    // So disabling this for now, we could add a re-fetch to the issues list if users start facing issues.
    // await expect(
    //   adminPage.getByText("Your trial expires in 30 days"),
    // ).toBeVisible(); // Billing banner
    // await expect(adminPage.getByText("Free trial")).toBeVisible(); // Billing status

    // Check that the dashboards are listed
    await expect(
      adminPage.getByRole("link", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible();
    await expect(
      adminPage.getByRole("link", { name: "Programmatic Ads Bids" }),
    ).toBeVisible();

    // Wait for the first dashboard to be ready
    await expect
      .poll(
        async () => {
          await adminPage.reload();
          const listing = adminPage.getByRole("link", {
            name: "Programmatic Ads Auction auction_explore",
          });
          return listing.textContent();
        },
        { intervals: Array(36).fill(5_000), timeout: 180_000 },
      )
      .toContain("Last refreshed");

    await expect
      .poll(
        async () => {
          await adminPage.reload();
          const listing = adminPage.getByRole("link", {
            name: "Programmatic Ads Bids bids_explore",
          });
          return listing.textContent();
        },
        { intervals: Array(12).fill(5_000), timeout: 60_000 },
      )
      .toContain("Last refreshed");
  });
});

async function isServiceReady(url: string): Promise<boolean> {
  try {
    const response = await axios.get(url);
    return response.status === 200;
  } catch {
    return false;
  }
}
