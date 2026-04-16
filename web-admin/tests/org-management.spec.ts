import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";

test.describe("Org management flow", () => {
  const welcomeUserOrg = "e2e-welcome-user-org";

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

    await anonPage.waitForURL("http://localhost:3000/-/welcome/theme");

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
    // Update the org name
    await anonPage.getByLabel("URL").fill(welcomeUserOrg);
    // Click the continue button to deploy
    await anonPage.getByRole("button", { name: "Continue" }).click();

    await anonPage.waitForURL(`http://localhost:3000/${welcomeUserOrg}`);
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
