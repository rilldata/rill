import { expect } from "@playwright/test";
import { test } from "./setup/base";

const VALID_GCS_PATH =
  "gs://rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_06.parquet";
const INVALID_GCS_PATH =
  "gs://rilldata-public/github-analytics/Clickhouse/2025/06/commits_2020.parquet";

test.describe("Test Connection", () => {
  test.use({ project: "Blank" });

  test("GCS connector - successful connection test", async ({ page }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select GCS connector
    await page.locator("#gcs").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="gcs"]');

    // Fill in the valid GCS path
    await page.getByRole("textbox", { name: "GS URI" }).fill(VALID_GCS_PATH);

    // Click the "Add data" button to test the connection
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Add data" })
      .click();

    // Wait for success message
    await expect(page.getByText("Source imported successfully")).toBeVisible();
  });

  test("GCS connector - failed connection test", async ({ page }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select GCS connector
    await page.locator("#gcs").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="gcs"]');

    // Fill in the invalid GCS path (file doesn't exist)
    await page.getByRole("textbox", { name: "GS URI" }).fill(INVALID_GCS_PATH);

    // Click the "Add data" button to test the connection
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Add data" })
      .click();

    // Wait for error container to appear
    await expect(page.locator(".error-container")).toBeVisible();
  });
});
