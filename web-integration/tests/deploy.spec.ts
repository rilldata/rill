import { expect, type Page } from "@playwright/test";
import { isOrgDeleted } from "@rilldata/web-common/tests/utils/is-org-deleted";
import { makeTempDir } from "@rilldata/web-common/tests/utils/make-temp-dir";
import { execAsync } from "@rilldata/web-common/tests/utils/spawn";
import { RILL_DEV_STORAGE_STATE } from "@rilldata/web-integration/tests/constants.ts";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "./setup/base";

test.describe("Deploy journey", () => {
  const cliHomeDir = makeTempDir("deploy_home");
  const deployFirstOrgName = "e2e-org-first";
  const deploySecondOrgName = "e2e-org-second";
  // Create consistent folders to facilitate running individual tests, not just the full suite.
  const sharedProjectDir = join(tmpdir(), "adbids");

  test.use({
    cliHomeDir,
    project: "AdBids",
    rillDevBrowserState: RILL_DEV_STORAGE_STATE,
  });

  test.describe.configure({
    mode: "serial",
    timeout: 60_000,
  });

  test.afterAll(async () => {
    const allOrgNames = [deployFirstOrgName, deploySecondOrgName];
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

  test.describe("Create/Update flow", () => {
    test.use({
      projectDir: sharedProjectDir,
    });

    test("Should create new org and deploy", async ({ rillDevPage }) => {
      // Goto explore before deploying. This should land the user to explore on cloud on deploy.
      await rillDevPage.goto(
        rillDevPage.url() + "/files/dashboards/AdBids_metrics_explore.yaml",
      );

      // Start waiting for popup before clicking Deploy.
      const popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      // 1st time deploy modal is opened
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
      await deployPage.getByLabel("URL").fill(deployFirstOrgName);
      // Update the display name
      await deployPage
        .getByLabel("Organization display name")
        .fill("E2E Test Org");
      // Org name should not be updated
      await expect(deployPage.getByLabel("URL")).toHaveValue(
        deployFirstOrgName,
      );
      // Click the continue button to deploy
      await deployPage.getByRole("button", { name: "Continue" }).click();

      // Deploy loader should show up.
      await expect(
        deployPage.getByText("Hang tight! We're deploying your project..."),
      ).toBeVisible();

      await assertAndSkipInvite(deployPage);

      // Explore is opened after deploying.
      await expect(
        deployPage.getByLabel("Breadcrumb navigation, level 2"),
      ).toHaveText("Adbids dashboard");

      // Org title is correct
      await expect(
        deployPage.getByLabel("Breadcrumb navigation, level 0"),
      ).toHaveText(
        /E2E Test Org.*/, // Trial pill is not always present because of race condition.
      );
    });

    test("Should be able to redeploy to project", async ({ rillDevPage }) => {
      // Update the title of dashboard
      await rillDevPage.goto(
        rillDevPage.url() + "/files/dashboards/AdBids_metrics_explore.yaml",
      );
      await rillDevPage
        .getByTitle("Display name")
        .fill("Adbids dashboard edited first org");

      // Start waiting for popup before clicking Deploy.
      const popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      await expect(
        rillDevPage.getByText("Push local changes to Rill Cloud?"),
      ).toBeVisible();

      // Select the 1st org's project
      await rillDevPage
        .getByRole("button", { name: `${deployFirstOrgName}/adbids` })
        .click();

      // Hit update
      await rillDevPage.getByRole("button", { name: "Update" }).click();

      // Await for the popped up deploy page after hitting deploy
      const deployPage = await popupPromise;

      await ensureProjectRedeployed(deployPage);

      await ensureDashboardTitle(
        deployPage,
        "Adbids dashboard edited first org",
      );
    });
  });

  test("Should create a second org and deploy for a fresh project", async ({
    rillDevPage,
  }) => {
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

    // Click the "Deploy as a new project" button to deploy
    await deployPage
      .getByRole("button", { name: "Deploy as a new project" })
      .click();

    await assertAndSkipInvite(deployPage);

    // Canvas is opened.
    await expect(
      deployPage.getByLabel("Breadcrumb navigation, level 2"),
    ).toHaveText("Adbids Canvas Dashboard");

    // Org title is correct
    await expect(
      deployPage.getByLabel("Breadcrumb navigation, level 0"),
    ).toHaveText(
      /e2e-org-second.*/, // Trial pill is not always present because of race condition.
    );
  });
});

async function assertAndSkipInvite(page: Page) {
  // Deploy is a success and invite page is opened. This can take a while, so it has increased timeout.
  await expect
    .poll(
      async () => {
        return page.getByText("Invite teammates to your project").isVisible();
      },
      { intervals: Array(6).fill(10_000), timeout: 60_000 },
    )
    .toBeTruthy();

  // Skip invite and continue to status page.
  await page.getByRole("button", { name: "Skip for now" }).click();
}

async function ensureProjectRedeployed(page: Page) {
  // Project homepage is opened on a re-deploy. This can take a while, so it has increased timeout.
  await expect(page.getByText("Welcome to")).toBeVisible({
    timeout: 120_000,
  });
}

async function ensureDashboardTitle(page: Page, title: string) {
  await expect(
    page.getByRole("link", {
      name: title,
    }),
  ).toBeVisible({ timeout: 60_000 });
}
