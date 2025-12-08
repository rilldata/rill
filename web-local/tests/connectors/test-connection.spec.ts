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

  test("S3 connector - step transition from connector to model", async ({
    page,
  }) => {
    // Skip test if environment variables are not set
    if (
      !process.env.RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID ||
      !process.env.RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID or RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY environment variable is not set",
      );
    }

    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select S3 connector
    await page.locator("#s3").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="s3"]');

    // Verify we're in step 1 (connector configuration)
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "Skip" })).toBeVisible();

    // Fill in valid AWS credentials
    await page
      .getByRole("textbox", { name: "AWS access key ID" })
      .fill(process.env.RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID!);
    await page
      .getByRole("textbox", { name: "AWS secret access key" })
      .fill(process.env.RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY!);

    // Test Skip button - should transition to step 2 without testing connection
    await page.getByRole("button", { name: "Skip" }).click();

    // Wait for step 2 (model configuration) to appear
    await expect(page.getByText("Model preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Add data" }),
    ).toBeVisible();

    // Verify step 2 form fields are present
    await expect(page.getByRole("textbox", { name: "S3 URI" })).toBeVisible();
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
    await expect(page.getByRole("button", { name: "Skip" })).toBeVisible();

    // Verify credentials are still filled
    await expect(
      page.getByRole("textbox", { name: "AWS access key ID" }),
    ).toHaveValue(process.env.RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID!);
    await expect(
      page.getByRole("textbox", { name: "AWS secret access key" }),
    ).toHaveValue(process.env.RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY!);

    // Now test "Test and Connect" button - should transition to step 2 after successful connection test
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" })
      .click();

    // Wait for step 2 (model configuration) to appear after successful connection
    await expect(page.getByText("Model preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Add data" }),
    ).toBeVisible();

    // Verify step 2 form fields are present
    await expect(page.getByRole("textbox", { name: "S3 URI" })).toBeVisible();
    await expect(
      page.getByRole("textbox", { name: "Source name" }),
    ).toBeVisible();

    // Test back button again
    await page.getByRole("button", { name: "Back" }).click();

    // Verify we're back in step 1 again
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();

    // Verify credentials are still preserved
    await expect(
      page.getByRole("textbox", { name: "AWS access key ID" }),
    ).toHaveValue(process.env.RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID!);
    await expect(
      page.getByRole("textbox", { name: "AWS secret access key" }),
    ).toHaveValue(process.env.RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY!);
  });
});
