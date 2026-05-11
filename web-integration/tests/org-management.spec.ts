import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { createLocalFileSource } from "@rilldata/web-common/tests/utils/source-helpers.ts";
import {
  ClickHouseTestContainer,
  enterClickhouseCredentials,
  selectAdBidsAndSubmit,
} from "@rilldata/web-common/tests/utils/clickhouse.ts";

test.describe.serial("Org management flow", () => {
  const welcomeUserOrg = "e2e-welcome-user-org";
  const welcomeUserProject = "e2e-welcome-project";

  const clickhouse = new ClickHouseTestContainer();

  test.beforeAll(async () => {
    await clickhouse.start();
    await clickhouse.seedAdBids();
  });

  test.afterAll(async () => {
    await clickhouse.stop();
  });

  // Load environment variables from our root `.env` file
  const __dirname = path.dirname(fileURLToPath(import.meta.url));
  dotenv.config({ path: path.resolve(__dirname, "../../.env") });

  test("Should login with viewer account and enter the welcome flow", async ({
    anonPage,
  }) => {
    // Check that the required environment variables are set. This is for type-safety.
    if (
      !process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_PASSWORD
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    // Log in with the admin account
    await anonPage.goto("/");

    await loginUser(
      anonPage,
      process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_EMAIL,
      process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_PASSWORD,
    );

    await expect(anonPage.getByText("Pick your color mode")).toBeVisible();

    // Visiting the root redirects back to welcome
    await anonPage.goto("/");
    await expect(anonPage.getByText("Pick your color mode")).toBeVisible();

    // Select dark theme
    await anonPage.getByLabel("Select dark theme").click();
    // Dark theme is set
    await expect(anonPage.locator(".dark")).toBeVisible();

    // Continue to org creation page
    await anonPage.getByRole("button", { name: "Continue" }).click();

    // Dark theme is persisted across pages
    await expect(anonPage.locator(".dark")).toBeVisible();

    // Org creation page is opened.
    await expect(anonPage.getByText("Create an organization")).toBeVisible();

    // Visiting the root redirects back to welcome
    await anonPage.goto("/");
    await expect(anonPage.getByText("Create an organization")).toBeVisible();

    // Set the feature flag for create project and refresh
    await anonPage.evaluate(
      `localStorage.setItem("rill:welcome:enabled", "true");`,
    );
    await anonPage.reload();
    // Create org form is still available
    await expect(anonPage.getByText("Create an organization")).toBeVisible();

    // Update the org name
    await anonPage.getByLabel("URL").fill(welcomeUserOrg);
    // Click the Continue button
    await anonPage.getByRole("button", { name: "Continue" }).click();

    // Project creation page is opened.
    await expect(anonPage.getByText("Create your first project")).toBeVisible();

    // Update the project name
    await anonPage.getByLabel("Name").fill(welcomeUserProject);
    // Click the Create project button
    await anonPage.getByRole("button", { name: "Create project" }).click();

    const ProjectEditRootPath = `/${welcomeUserOrg}/${welcomeUserProject}/@develop/-/edit`;

    // Create a clickhouse connector that has secrets
    await anonPage.getByLabel("Connect to clickhouse").click();
    await enterClickhouseCredentials(anonPage, clickhouse);
    // Submit the form and select adbids
    await anonPage.getByRole("button", { name: "Test and Connect" }).click();
    await selectAdBidsAndSubmit(anonPage, false);

    // Assert user is on edit path
    expect(anonPage.url()).toContain(ProjectEditRootPath);

    // Upload a local file to create a source
    await createLocalFileSource(
      anonPage,
      "AdImpressions.tsv",
      "/models/AdImpressions.yaml",
    );

    // Start waiting for popup before clicking Publish.
    const popupPromise = anonPage.waitForEvent("popup");

    // Publish
    await anonPage.getByRole("button", { name: "Publish" }).click();

    // Await for the popped up publish page after hitting Publish
    const publishPage = await popupPromise;

    // Skip invite and continue to status page.
    await publishPage.getByRole("button", { name: "Skip for now" }).click();

    // Project homepage is opened on since we are not deploying from any dashboard.
    await expect(publishPage.getByText("Welcome to")).toBeVisible({
      timeout: 120_000,
    });
  });

  test("Should delete the org created in welcome flow", async ({
    anonPage,
  }) => {
    // Check that the required environment variables are set. This is for type-safety.
    if (
      !process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_PASSWORD
    ) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    await loginUser(
      anonPage,
      process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_EMAIL,
      process.env.RILL_DEVTOOL_E2E_VIEWER_ACCOUNT_PASSWORD,
    );

    // Start the delete org process
    await anonPage.goto(`/${welcomeUserOrg}/-/settings`);
    const deleteOrgButton = anonPage.getByLabel("Delete organization");
    await deleteOrgButton.scrollIntoViewIfNeeded();
    await deleteOrgButton.click();
    // Confirm the deletion
    await anonPage.getByTitle("confirmation").fill(`delete ${welcomeUserOrg}`);
    // Click the delete button
    await anonPage.getByRole("button", { name: "Delete", exact: true }).click();
    // Wait for delete to complete
    await expect(anonPage.getByLabel("Notification")).toHaveText(
      "Deleted organization",
    );

    await anonPage.goto(`/${welcomeUserOrg}`);

    // Wait for async jobs to finish deleting org.
    await expect
      .poll(
        async () => {
          await anonPage.reload();
          try {
            // Wait for "Organization not found" to show up after a reload.
            await expect(
              anonPage.getByText("Organization not found"),
            ).toBeVisible();
            return true;
          } catch {
            return false;
          }
        },
        { timeout: 120_000, intervals: Array(12).fill(10_000) },
      )
      .toBeTruthy();
  });
});

async function loginUser(page: Page, email: string, password: string) {
  // Log in with the admin account
  await page.goto("/");

  // Fill in the email
  const emailInput = page.locator('input[name="username"]');
  await emailInput.waitFor({ state: "visible" });
  await emailInput.click();
  await emailInput.fill(email);

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
  await passwordInput.fill(password);

  // Click the continue button
  await page
    .locator('button[type="submit"][data-action-button-primary="true"]', {
      hasText: "Continue",
    })
    .click();
}
