import { makeTempDir } from "@rilldata/web-common/tests/utils/make-temp-dir";
import { execAsync } from "@rilldata/web-common/tests/utils/spawn";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "./setup/base";
import { expect, type Page } from "@playwright/test";
import { isOrgDeleted } from "@rilldata/web-common/tests/utils/is-org-deleted";

test.describe("Deploy journey", () => {
  const cliHomeDir = makeTempDir("deploy_home");
  const deployFirstOrgName = "e2e-org-first";
  const deploySecondOrgName = "e2e-org-second";
  const deployThirdOrgName = "e2e-org-third";
  // Create consistent folders to facilitate running individual tests, not just the full suite.
  const sharedFirstProjectDir = join(tmpdir(), "adbids");
  const sharedSecondProjectDir = join(tmpdir(), "adimpressions");

  test.use({
    cliHomeDir,
    project: "AdBids",
  });

  test.describe.configure({
    mode: "serial",
    timeout: 60_000,
  });

  test.afterAll(async () => {
    const allOrgNames = [
      deployFirstOrgName,
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

  test.describe("Create/Update flow", () => {
    test.use({
      projectDir: sharedFirstProjectDir,
    });

    test("Should create new org and deploy", async ({ rillDevPage }) => {
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
      // Start waiting for popup before clicking Deploy.
      const popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      // Update popup is opened
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

  test.describe("Overwrite flow", () => {
    test.use({
      project: "AdImpressions",
      projectDir: sharedSecondProjectDir,
    });

    test("Should overwrite a different project of a different org", async ({
      rillDevPage,
    }) => {
      // Start waiting for popup before clicking Deploy.
      const popupPromise = rillDevPage.waitForEvent("popup");

      await rillDevPage.getByRole("button", { name: "Deploy" }).click();

      // Await for the popped up deploy page after hitting deploy
      const deployPage = await popupPromise;

      // Org selection page is opened.
      await expect(
        deployPage.getByText("Select an organization"),
      ).toBeVisible();

      // Create a new org from the dropdown.
      await deployPage.getByLabel("Select organization").click();
      await deployPage.getByText(deployFirstOrgName).click();

      // Click the "Or overwrite an existing project" button to select a project to overwrite
      await deployPage
        .getByRole("button", { name: "Or overwrite an existing project" })
        .click();

      // Select the adbids project to overwrite
      await deployPage
        .getByRole("button", { name: `${deployFirstOrgName}/adbids` })
        .click();

      // Hit "Update selected project" to continue to deploy
      await deployPage
        .getByRole("button", { name: "Update selected project" })
        .click();

      // Overwrite configuration shows up.
      await expect(
        deployPage.getByText(
          "Are you sure you want to overwrite this project?",
        ),
      ).toBeVisible();
      // Confirm overwrite
      await deployPage.getByRole("button", { name: "Yes, overwrite" }).click();

      // Deploy loader should show up.
      await expect(
        deployPage.getByText("Hang tight! We're deploying your project..."),
      ).toBeVisible();

      await ensureProjectRedeployed(deployPage);

      await ensureDashboardTitle(deployPage, "AdImpressions dashboard");
    });
  });
});

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
  } catch {
    // no-op
  }

  // Check that the dashboard's title has changed. Since there is an async reconcile is involved, we need to refresh and wait.
  await expect
    .poll(
      async () => {
        await page.reload();
        await page.waitForTimeout(1000); // Wait for page to fully load
        const listing = page
          .getByRole("link", {
            name: title,
          })
          .first();
        return listing.isVisible();
      },
      {
        // Increased timeout for the 1st dashboard to make reconcile.
        intervals: [10_000, 20_000, 30_000],
        timeout: 60_000,
      },
    )
    .toBeTruthy();
}
