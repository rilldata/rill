import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Test Connection", () => {
  test.use({ project: "Blank" });

  test("GCS connector - HMAC", async ({ page }) => {
    // Skip test if environment variables are not set
    if (
      !process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY ||
      !process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_HMAC_KEY or RILL_RUNTIME_GCS_TEST_HMAC_SECRET environment variable is not set",
      );
    }

    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select GCS connector
    await page.locator("#gcs").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="gcs"]');

    // Select HMAC keys authentication method
    await page.getByRole("radio", { name: "HMAC keys" }).click();

    // Fill in invalid HMAC credentials
    await page
      .getByRole("textbox", { name: "Access Key ID" })
      .fill("invalid-key-id");
    await page
      .getByRole("textbox", { name: "Secret Access Key" })
      .fill("invalid-secret");

    // Click the "Test and Connect" button to test the connection
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" })
      .click();

    // Wait for error container to appear
    await expect(page.locator(".error-container")).toBeVisible();
  });

  test("GCS connector - step transition from connector to model", async ({
    page,
  }) => {
    // Skip test if environment variables are not set
    if (
      !process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY ||
      !process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_HMAC_KEY or RILL_RUNTIME_GCS_TEST_HMAC_SECRET environment variable is not set",
      );
    }

    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select GCS connector
    await page.locator("#gcs").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="gcs"]');

    // Verify we're in step 1 (connector configuration)
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();

    // Select HMAC keys authentication method
    await page.getByRole("radio", { name: "HMAC keys" }).click();

    // Fill in valid HMAC credentials
    await page
      .getByRole("textbox", { name: "Access Key ID" })
      .fill(process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY!);
    await page
      .getByRole("textbox", { name: "Secret Access Key" })
      .fill(process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET!);

    // Click the "Test and Connect" button to transition to step 2
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" })
      .click();

    // Wait for step 2 (model configuration) to appear
    await expect(page.getByText("Model preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Add data" }),
    ).toBeVisible();

    // Verify step 2 form fields are present
    await expect(page.getByRole("textbox", { name: "GS URI" })).toBeVisible();
    await expect(
      page.getByRole("textbox", { name: "Source name" }),
    ).toBeVisible();

    // Test back button functionality
    await page.getByRole("button", { name: "Back" }).click();

    // Verify we're back in step 1
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();

    // Verify HMAC fields are still filled
    await expect(
      page.getByRole("textbox", { name: "Access Key ID" }),
    ).toHaveValue(process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY!);
    await expect(
      page.getByRole("textbox", { name: "Secret Access Key" }),
    ).toHaveValue(process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET!);
  });
});
