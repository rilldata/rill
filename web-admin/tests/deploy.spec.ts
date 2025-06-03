import { execAsync } from "@rilldata/web-common/tests/utils/spawn";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { test, TestTempDirectory } from "./setup/base";
import { expect } from "@playwright/test";
import { join } from "node:path";

test.describe("Deploy journey", () => {
  test.use({
    cliHome: join(TestTempDirectory, "deploy_home"),
    rillDevProject: "adbids_lite",
  });

  test.afterAll(async () => {
    await execAsync(
      // We need to set the home to get the correct creds
      `HOME=${join(TestTempDirectory, "deploy_home")} rill org delete e2e-viewer --interactive=false`,
    );

    // Wait for the organization to be deleted
    // This includes deleting the org from Orb and Stripe, which we'd like to do to keep those environments clean.
    await expect
      .poll(async () => await isOrgDeleted("e2e-viewer"), {
        intervals: [1_000],
        timeout: 15_000,
      })
      .toBeTruthy();
  });

  // Note: This uses the viewer account to avoid conflicts with the admin account that would already have an org and project.

  test("Should create new org and deploy", async ({ rillDevPage }) => {
    // Load environment variables from our root `.env` file
    // We need this here again since this would be a separate process compared to setup.
    const __dirname = path.dirname(fileURLToPath(import.meta.url));
    dotenv.config({ path: path.resolve(__dirname, "../../.env") });

    // Check that the required environment variables are set.
    if (
      !process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_PASSWORD
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    // Start waiting for popup before clicking Deploy.
    const popupPromise = rillDevPage.waitForEvent("popup");

    await rillDevPage.getByRole("button", { name: "Deploy" }).click();
    // 1st time deploy modal is opened.
    await expect(
      rillDevPage.getByText("Deploy this project for free"),
    ).toBeVisible();
    // Hit continue to start deployment
    await rillDevPage.getByRole("button", { name: "Continue" }).click();

    // Await for the popped up deploy page after hitting deploy
    const deployPage = await popupPromise;

    // Login 1st to start deploy.

    // Fill in the email
    const emailInput = deployPage.locator('input[name="username"]');
    await emailInput.waitFor({ state: "visible" });
    await emailInput.click();
    await emailInput.fill(process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_EMAIL);

    // Click the continue button
    await deployPage
      .locator('button[type="submit"][data-action-button-primary="true"]', {
        hasText: "Continue",
      })
      .click();

    // Fill in the password
    const passwordInput = deployPage.locator('input[name="password"]');
    await passwordInput.waitFor({ state: "visible" });
    await passwordInput.click();
    await passwordInput.fill(
      process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_PASSWORD,
    );

    // Click the continue button
    await deployPage
      .locator('button[type="submit"][data-action-button-primary="true"]', {
        hasText: "Continue",
      })
      .click();

    // Deploy should continue after logging in

    // Deploy loader should show up.
    await expect(
      deployPage.getByText("Hang tight! We're deploying your project..."),
    ).toBeVisible();

    // Deploy is a success and invite page is opened. This can take a while, so it has increased timeout.
    await expect
      .poll(
        async () => {
          return deployPage
            .getByText("Invite teammates to your project")
            .isVisible();
        },
        { intervals: Array(6).fill(5_000), timeout: 30_000 },
      )
      .toBeTruthy();

    // Skip invite and continue to status page.
    await deployPage.getByRole("button", { name: "Skip for now" }).click();

    // Project status page is opened.
    await expect(deployPage.getByLabel("Container title")).toHaveText(
      "Project status",
    );

    // Check that the dashboards are listed
    await expect(deployPage.getByText("AdBids_metrics_explore")).toBeVisible();

    // TODO: verify reconciliation when we expand deploy tests
  });
});
