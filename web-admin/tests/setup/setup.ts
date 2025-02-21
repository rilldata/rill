import { chromium, expect } from "@playwright/test";
import axios from "axios";
import { spawn } from "child_process";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { writeFileEnsuringDir } from "../utils/fs";
import {
  execAsync,
  spawnAndMatch,
} from "@rilldata/web-common/tests/utils/spawn";
import type { StorageState } from "../utils/storage-state";
import { test as setup } from "./base";
import {
  ADMIN_STORAGE_STATE,
  RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE,
} from "./constants";
import { cliLogin } from "./fixtures/cli";

setup.describe("global setup", () => {
  setup.describe.configure({
    mode: "serial",
    timeout: 240_000,
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
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD ||
      !process.env.RILL_DEVTOOL_E2E_GITHUB_STORAGE_STATE_JSON
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    // Start the admin and runtime services in a detached background process.
    // A detached process ensures they are not cleaned up when this setup project completes.
    // However, we need to be sure to clean-up the processes manually in the teardown project.
    const child = spawn(
      "rill",
      ["devtool", "start", "e2e", "--only", "admin,runtime"],
      {
        detached: true,
        stdio: "ignore",
        cwd: repoRoot,
      },
    );
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

  setup("should log in with the admin account", async () => {
    // Again, check that the required environment variables are set. This is for type-safety.
    if (
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD ||
      !process.env.RILL_DEVTOOL_E2E_GITHUB_STORAGE_STATE_JSON
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    // Launch a Chromium browser with an authenticated GitHub session
    const browser = await chromium.launch();
    const context = await browser.newContext({
      storageState: JSON.parse(
        process.env.RILL_DEVTOOL_E2E_GITHUB_STORAGE_STATE_JSON,
      ) as StorageState,
    });
    const page = await context.newPage();

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

    // Save the admin's Rill auth cookies to file. The resultant file will include both the GitHub and Rill auth cookies.
    // Subsequent tests can seed their browser with this state, instead of needing to go through the log-in flow again.
    await page.context().storageState({ path: ADMIN_STORAGE_STATE });
  });

  setup("should create an organization", async ({ adminPage }) => {
    // Create an organization named "e2e"
    await cliLogin(adminPage);
    const { stdout: orgCreateStdout } = await execAsync("rill org create e2e");
    expect(orgCreateStdout).toContain("Created organization");

    // Go to the organization's page
    await adminPage.goto("/e2e");
    await expect(adminPage.getByRole("heading", { name: "e2e" })).toBeVisible();
  });

  setup("should deploy the OpenRTB project", async ({ adminPage }) => {
    // Pull the repositories to be used for testing
    const examplesRepoPath = "tests/setup/git/repos/rill-examples";
    await execAsync(
      `rm -rf ${examplesRepoPath} && git clone https://github.com/rilldata/rill-examples.git ${examplesRepoPath}`,
    );

    // Deploy the OpenRTB project
    const { match } = await spawnAndMatch(
      "rill",
      [
        "deploy",
        "--path",
        "tests/setup/git/repos/rill-examples",
        "--subpath",
        "rill-openrtb-prog-ads",
        "--project",
        "openrtb",
        "--github",
        "--interactive=false",
      ],
      /https?:\/\/[^\s]+/,
    );

    // Navigate to the GitHub auth URL
    // (In a fresh browser, this would typically trigger a log-in to GitHub, but we've bootstrapped the Playwright browser with an authenticated GitHub session.
    // See the `setup-github-session` project in `playwright.config.ts` for details.)
    const url = match[0];
    await adminPage.goto(url);
    await adminPage.waitForURL("/-/github/connect/success");

    // Wait for the deployment to complete
    // TODO: Replace this with a better check. Maybe we could modify `spawnAndMatch` to match an array of regexes.
    await adminPage.waitForTimeout(10000);

    // Expect to see the successful deployment
    await adminPage.goto("/e2e/openrtb");
    await expect(
      adminPage.getByText("Your trial expires in 30 days"),
    ).toBeVisible(); // Billing banner
    await expect(adminPage.getByText("e2e")).toBeVisible(); // Organization breadcrumb
    await expect(adminPage.getByText("Free trial")).toBeVisible(); // Billing status
    await expect(adminPage.getByText("openrtb")).toBeVisible(); // Project breadcrumb

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
        { intervals: Array(24).fill(5_000), timeout: 180_000 },
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
