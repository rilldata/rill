import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Save Anyway feature", () => {
  test.use({ project: "Blank" });

  test("Save Anyway button appears and works for invalid connector", async ({
    page,
  }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select ClickHouse connector
    await page.locator("#clickhouse").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="clickhouse"]');

    // Fill in connection details
    await page.getByRole("textbox", { name: "Host" }).fill("asd");
    await page.getByRole("textbox", { name: "Password" }).fill("asd");

    // Click "Test and Connect" - this should fail connection test and show "Save Anyway" button
    await page.getByRole("button", { name: "Test and Connect" }).click();

    // Wait for "Connecting..." state to appear
    await expect(
      page.getByRole("button", { name: "Connecting..." }),
    ).toBeVisible();

    // Wait for "Save Anyway" button to appear
    await expect(
      page.getByRole("button", { name: "Save Anyway" }),
    ).toBeVisible();

    // Click "Save Anyway" button
    await page.getByRole("button", { name: "Save Anyway" }).click();

    // Wait for navigation to the new connector file (should succeed despite connection failure)
    await page.waitForURL(`**/files/connectors/clickhouse.yaml`);

    // Verify the connector file was created
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");

    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: clickhouse");
  });
});
