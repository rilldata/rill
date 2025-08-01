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
  const deployThirdOrgName = "e2e-org-third";
  const sharedProjectDir = makeTempDir("adbids");

  test.use({
    cliHomeDir,
    project: "AdBids",
  });

  test.describe.configure({
    mode: "serial",
    timeout: 180_000,
  });

  test.afterAll(async () => {
    const allOrgNames = [
      deployOrgName,
      deploySecondOrgName,
      deployThirdOrgName,
    ];
    await Promise.all(
      allOrgNames.map(async (orgName) =>
        execAsync(
          // We need to set the home to get the correct creds
          `HOME=${cliHomeDir} rill org delete ${orgName} --interactive=false`,
        ),
      ),
    );

    // Wait for the organization to be deleted
    // This includes deleting the org from Orb and Stripe, which we'd like to do to keep those environments clean.
    await expect
      .poll(
        async () =>
          (
            await Promise.all(
              allOrgNames.map(async (orgName) =>
                isOrgDeleted(orgName, cliHomeDir),
              ),
            )
          ).every(Boolean),
        {
          intervals: [1_000],
          timeout: 15_000,
        },
      )
      .toBeTruthy();
  });

  test.describe("Update flow", () => {
    test.use({
      projectDir: sharedProjectDir,
    });

    test("Should create new org and deploy", async ({ rillDevPage }) => {
      // Start waiting for popup before clicking Deploy.
      const popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      await login(rillDevPage);

      // Deploy should continue after logging in

      // 1st time deploy modal is opened after the login
      await expect(
        rillDevPage.getByText("Deploy this project for free"),
      ).toBeVisible();
      // Hit continue to start deployment
      await rillDevPage.getByRole("button", { name: "Continue" }).click();

      // Await for the popped up deploy page after hitting deploy
      const deployPage = await popupPromise;

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

      await assertAndSkipInvite(deployPage);

      // Project status page is opened.
      await expect(deployPage.getByLabel("Container title")).toHaveText(
        "Project status",
      );

      // Check that the dashboards are listed
      await expect(
        deployPage.getByText("AdBids_metrics_explore"),
      ).toBeVisible();

      // Org title is correct
      await expect(
        deployPage.getByLabel("Breadcrumb navigation, level 0"),
      ).toHaveText(
        /E2E Test Org.*/, // Trial pill is not always present because of race condition.
      );
    });

    test("Should deploy to another project from the same local project folder", async ({
      rillDevPage,
    }) => {
      await ensureLogout(rillDevPage);

      // Start waiting for popup before clicking Deploy.
      const popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      await login(rillDevPage);

      // Deploy should continue after logging in
      await expect(
        rillDevPage.getByText("Push local changes to Rill Cloud?"),
      ).toBeVisible();

      // Deploy to another project
      await rillDevPage
        .getByRole("button", { name: "Deploy to another project" })
        .click();

      // Await for the popped up deploy page after hitting deploy
      const deployPage = await popupPromise;

      // Org selection page is opened.
      await expect(
        deployPage.getByText("Select an organization"),
      ).toBeVisible();

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
      await deployPage
        .getByRole("button", { name: "Deploy as a new project" })
        .click();

      await assertAndSkipInvite(deployPage);

      // Project status page is opened.
      await expect(deployPage.getByLabel("Container title")).toHaveText(
        "Project status",
      );

      // Org title is correct
      await expect(
        deployPage.getByLabel("Breadcrumb navigation, level 0"),
      ).toHaveText(
        /e2e-org-second.*/, // Trial pill is not always present because of race condition.
      );

      // Check that the dashboards are listed
      await expect(
        deployPage.getByText("AdBids_metrics_explore"),
      ).toBeVisible();
    });

    test("Should be able to redeploy to different projects with same name", async ({
      rillDevPage,
    }) => {
      await ensureLogout(rillDevPage);

      // Update the title of dashboard
      await rillDevPage.goto(
        rillDevPage.url() + "/files/dashboards/AdBids_metrics_explore.yaml",
      );
      await rillDevPage
        .getByTitle("Display name")
        .fill("Adbids dashboard edited first org");

      // Start waiting for popup before clicking Deploy.
      let popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      await login(rillDevPage);

      // Deploy should continue after logging in
      await expect(
        rillDevPage.getByText("Push local changes to Rill Cloud?"),
      ).toBeVisible();

      // Select the 1st org's project
      await rillDevPage
        .getByRole("button", { name: `${deployOrgName}/adbids` })
        .click();

      // Hit update
      await rillDevPage.getByRole("button", { name: "Update" }).click();

      // Await for the popped up deploy page after hitting deploy
      let deployPage = await popupPromise;

      await ensureProjectRedeployed(deployPage);

      await ensureDashboardTitle(
        deployPage,
        "Adbids dashboard edited first org",
      );

      // Close the deploy page which is the rill cloud page and focus the rill dev page.
      await deployPage.close();
      await rillDevPage.bringToFront();

      // Update the title of dashboard again
      await rillDevPage.goto(
        rillDevPage.url() + "/files/dashboards/AdBids_metrics_explore.yaml",
      );
      await rillDevPage
        .getByTitle("Display name")
        .fill("Adbids dashboard edited second org");

      // Start waiting for popup before clicking Deploy.
      popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      // Select the 2nd org's project
      await rillDevPage
        .getByRole("button", { name: `${deploySecondOrgName}/adbids` })
        .click();

      // Hit update
      await rillDevPage.getByRole("button", { name: "Update" }).click();

      // Await for the popped up deploy page after hitting deploy
      deployPage = await popupPromise;

      await ensureProjectRedeployed(deployPage);

      await ensureDashboardTitle(
        deployPage,
        "Adbids dashboard edited second org",
      );
    });
  });

  test("Should create a third org and deploy for a fresh project", async ({
    rillDevPage,
  }) => {
    await ensureLogout(rillDevPage);

    // Start waiting for popup before clicking Deploy.
    const popupPromise = rillDevPage.waitForEvent("popup");

    await rillDevPage.getByRole("button", { name: "Deploy" }).click();

    // Await for the popped up deploy page after hitting deploy
    const deployPage = await popupPromise;

    await login(deployPage);

    // Org selection page is opened.
    await expect(deployPage.getByText("Select an organization")).toBeVisible();

    // Create a new org from the dropdown.
    await deployPage.getByLabel("Select organization").click();
    await deployPage.getByText("+ Create organization").click();

    // Enter the org display name
    await deployPage
      .getByLabel("Organization display name")
      .fill(deployThirdOrgName);
    await deployPage.getByLabel("Create new org").click();

    // Notification is shown for org creation
    await expect(deployPage.getByLabel("Notification")).toHaveText(
      `Created organization ${deployThirdOrgName}`,
    );

    // Click the "Deploy as a new project" button to deploy
    await deployPage
      .getByRole("button", { name: "Deploy as a new project" })
      .click();

    await assertAndSkipInvite(deployPage);

    // Project status page is opened.
    await expect(deployPage.getByLabel("Container title")).toHaveText(
      "Project status",
    );

    // Org title is correct
    await expect(
      deployPage.getByLabel("Breadcrumb navigation, level 0"),
    ).toHaveText(
      /e2e-org-third.*/, // Trial pill is not always present because of race condition.
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

async function assertAndSkipInvite(page: Page) {
  // Deploy is a success and invite page is opened. This can take a while, so it has increased timeout.
  await expect
    .poll(
      async () => {
        return page.getByText("Invite teammates to your project").isVisible();
      },
      { intervals: Array(6).fill(5_000), timeout: 30_000 },
    )
    .toBeTruthy();

  // Skip invite and continue to status page.
  await page.getByRole("button", { name: "Skip for now" }).click();
}

async function ensureLogout(page: Page) {
  await page.getByLabel("Avatar", { exact: true }).click();

  try {
    await page.getByText("Logout").click();
  } catch {
    // nothing to do if already logged out
  }
}

async function ensureProjectRedeployed(page: Page) {
  // Dashboard listing page is opened on a re-deploy. This can take a while, so it has increased timeout.
  await expect
    .poll(
      async () => {
        await page.reload();
        const title = page.getByLabel("Container title");
        return title.textContent();
      },
      { intervals: Array(5).fill(20_000), timeout: 120_000 },
    )
    .toEqual("Project dashboards");
}

async function ensureDashboardTitle(page: Page, title: string) {
  // Do a check upfront to avoid a reload.
  try {
    await expect(
      page
        .getByRole("link", {
          name: title,
        })
        .first(),
    ).toBeVisible();
    // Title is already updated so return
    return;
  } catch (e) {
    console.log(e);
    // no-op
  }

  // Check that the dashboard's title has changed. Since there is an async reconcile is involved, we need to refresh and wait.
  await expect
    .poll(
      async () => {
        await page.reload();
        const listing = page
          .getByRole("link", {
            name: title,
          })
          .first();
        return listing.isVisible();
      },
      {
        // Increased timeout for the 1st dashboard to make reconcile.
        intervals: [10_000, 10_000, 20_000, 20_000, 30_000, 30_000],
        timeout: 120_000,
      },
    )
    .toBeTruthy();
}
