import { expect } from "@playwright/test";
import axios from "axios";
import { spawn } from "child_process";
import dotenv from "dotenv";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import { execAsync, spawnAndMatch } from "../utils/spawn";
import { test as setup } from "./base";
import { ADMIN_AUTH_FILE } from "./constants";
import { cliLogin } from "./fixtures/cli";

setup(
  "should start services, log-in a user, and deploy a project",
  async ({ page }) => {
    const timeout = 180_000;
    setup.setTimeout(timeout);

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
        timeoutMs: timeout,
      },
    );

    // Load environment variables from our root `.env` file
    const __dirname = path.dirname(fileURLToPath(import.meta.url));
    dotenv.config({ path: path.resolve(__dirname, "../../../.env") });

    // Check that the required environment variables are set
    // The above `rill devtool` command pulls the `.env` file with these values.
    if (
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
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
      fs.writeFileSync("rill-devtool-pid.txt", child.pid.toString());
    } else {
      throw new Error("Failed to get pid of child process");
    }

    // Wait for the admin service to be ready
    await expect
      .poll(() => isServiceReady("http://localhost:8080/v1/ping"), {
        timeout: 30_000,
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

    // Pull the repositories to be used for testing
    await execAsync(
      "git clone https://github.com/rilldata/rill-examples.git tests/setup/git/repos/rill-examples",
    );

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

    // Save auth cookies to file
    // Subsequent tests can use these cookies to authenticate, rather than logging in again.
    await page.context().storageState({ path: ADMIN_AUTH_FILE });

    // Create an organization named "e2e"
    await cliLogin(page);
    const { stdout: orgCreateStdout } = await execAsync("rill org create e2e");
    expect(orgCreateStdout).toContain("Created organization");

    // Go to the organization's page
    await page.goto("/e2e");
    await expect(page.getByRole("heading", { name: "e2e" })).toBeVisible();

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
        "true",
      ],
      /https?:\/\/[^\s]+/,
    );

    // Navigate to the GitHub auth URL
    // (In a fresh browser, this would typically trigger a log-in to GitHub, but we've bootstrapped the Playwright browser with GitHub auth cookies.
    // See the `save-github-cookies` project in `playwright.config.ts` for details.)
    const url = match[0];
    await page.goto(url);
    await page.waitForURL("/-/github/connect/success");

    // Wait for the deployment to complete (TODO: replace this with a better check)
    await page.waitForTimeout(10000);

    // Expect to see the successful deployment
    await page.goto("/e2e/openrtb");
    await expect(page.getByText("Your trial expires in 30 days")).toBeVisible(); // Billing banner
    await expect(page.getByText("e2e")).toBeVisible(); // Organization breadcrumb
    await expect(page.getByText("Free trial")).toBeVisible(); // Billing status
    await expect(page.getByText("openrtb")).toBeVisible(); // Project breadcrumb

    // Check that the dashboards are listed
    await expect(
      page.getByRole("link", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: "Programmatic Ads Bids" }),
    ).toBeVisible();

    // Wait for the first dashboard to be ready
    await expect
      .poll(
        async () => {
          await page.reload();
          const listing = page.getByRole("link", {
            name: "Programmatic Ads Auction auction_explore",
          });
          return listing.textContent();
        },
        { intervals: Array(24).fill(5_000), timeout: 120_000 },
      )
      .toContain("Last refreshed");
  },
);

async function isServiceReady(url: string): Promise<boolean> {
  try {
    const response = await axios.get(url);
    return response.status === 200;
  } catch {
    return false;
  }
}
