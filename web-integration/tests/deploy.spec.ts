import { makeTempDir } from "@rilldata/web-common/tests/utils/make-temp-dir";
import { execAsync } from "@rilldata/web-common/tests/utils/spawn";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { test } from "./setup/base";
import { expect, type Page } from "@playwright/test";
import { isOrgDeleted } from "@rilldata/web-common/tests/utils/is-org-deleted";

test.describe("Deploy journey", () => {
  const cliHomeDir = makeTempDir("deploy_home");
  const deployOrgName = "e2e-org";
  const deploySecondOrgName = "e2e-org-second";

  test.use({
    cliHomeDir,
    project: "AdBids",
  });

  test.afterAll(async () => {
    await execAsync(
      // We need to set the home to get the correct creds
      `HOME=${cliHomeDir} rill org delete ${deployOrgName} --interactive=false`,
    );
    await execAsync(
      // We need to set the home to get the correct creds
      `HOME=${cliHomeDir} rill org delete ${deploySecondOrgName} --interactive=false`,
    );

    // Wait for the organization to be deleted
    // This includes deleting the org from Orb and Stripe, which we'd like to do to keep those environments clean.
    await expect
      .poll(
        async () =>
          (await isOrgDeleted(deployOrgName, cliHomeDir)) &&
          (await isOrgDeleted(deploySecondOrgName, cliHomeDir)),
        {
          intervals: [1_000],
          timeout: 15_000,
        },
      )
      .toBeTruthy();
  });

  test("Should create new org and deploy", async ({ rillDevPage }) => {
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

    await login(deployPage);

    // Deploy should continue after logging in

    // Org creation page is opened.
    await expect(
      deployPage.getByText("Let’s create your first organization"),
    ).toBeVisible();
    // Enter the org display name
    await deployPage.getByLabel("Organization display name").fill("E2E Org");
    // Org name should be auto-generated
    await expect(deployPage.getByLabel("URL")).toHaveValue("E2E-Org");
    // Update the org name directly
    await deployPage.getByLabel("URL").fill(deployOrgName);
    // Update the display name
    await deployPage
      .getByLabel("Organization display name")
      .fill("E2E Test Org");
    // Org name should not be updated
    await expect(deployPage.getByLabel("URL")).toHaveValue(deployOrgName);
    // Click the continue button to deploy
    await deployPage.getByRole("button", { name: "Continue" }).click();

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

    // Org title is correct
    await expect(
      deployPage.getByLabel("Breadcrumb navigation, level 0"),
    ).toHaveText(
      /E2E Test Org.*/, // Trail pill is not always present because of race condition.
    );

    // TODO: verify reconciliation when we expand deploy tests
  });

  test("Should create a second org and deploy", async ({ rillDevPage }) => {
    // Start waiting for popup before clicking Deploy.
    const popupPromise = rillDevPage.waitForEvent("popup");

    await rillDevPage.getByRole("button", { name: "Deploy" }).click();

    // Await for the popped up deploy page after hitting deploy
    const deployPage = await popupPromise;

    // Org selection page is opened.
    await expect(deployPage.getByText("Select an organization")).toBeVisible();

    // Create a new org from the dropdown.
    await deployPage.getByLabel("Select organization").click();
    await deployPage.getByText("+ Create organization").click();

    // Enter the org display name
    await deployPage
      .getByLabel("Organization display name")
      .fill(deploySecondOrgName);
    await deployPage.getByLabel("Create new org").click();

    // Notification is shown for org creation
    await expect(deployPage.getByLabel("Notification")).toHaveText(
      `Created organization ${deploySecondOrgName}`,
    );

    // Click the continue button to deploy
    await deployPage.getByRole("button", { name: "Continue" }).click();

    // Visiting the cloud page needs a login.
    // TODO: reuse context for just this page.
    await login(deployPage);

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

    // Org title is correct
    await expect(
      deployPage.getByLabel("Breadcrumb navigation, level 0"),
    ).toHaveText(
      /e2e-org-second.*/, // Trail pill is not always present because of race condition.
    );

    // Check that the dashboards are listed
    await expect(deployPage.getByText("AdBids_metrics_explore")).toBeVisible();
  });
});

async function login(deployPage: Page) {
  // Load environment variables from our root `.env` file
  // We need this here again since this would be a separate process compared to setup.
  const __dirname = path.dirname(fileURLToPath(import.meta.url));
  dotenv.config({ path: path.resolve(__dirname, "../../.env") });

  // Check that the required environment variables are set.
  if (
    !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
    !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
  ) {
    throw new Error(
      "Missing required environment variables for authentication",
    );
  }

  // Login 1st to start deploy.

  // Fill in the email
  const emailInput = deployPage.locator('input[name="username"]');
  await emailInput.waitFor({ state: "visible" });
  await emailInput.click();
  await emailInput.fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL);

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
  await passwordInput.fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD);

  // Click the continue button
  await deployPage
    .locator('button[type="submit"][data-action-button-primary="true"]', {
      hasText: "Continue",
    })
    .click();
}
