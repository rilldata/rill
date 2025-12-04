import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Save Anyway feature", () => {
  test.use({ project: "Blank" });

  test("Save Anyway button appears and redirects to connector file", async ({
    page,
  }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select ClickHouse connector
    await page.locator("#clickhouse").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="clickhouse"]');

    // Fill in connection details with invalid values
    await page.getByRole("textbox", { name: "Host" }).fill("asd");
    await page.getByRole("textbox", { name: "Password" }).fill("asd");

    // Click "Test and Connect" - this should fail connection test and show "Save Anyway" button
    await page
      .getByRole("button", { name: /^(Test and Connect|Connect)$/ })
      .click();

    // Wait for "Save Anyway" button to appear
    await expect(
      page.getByRole("button", { name: "Save Anyway" }),
    ).toBeVisible();

    // Click "Save Anyway" button
    await page.getByRole("button", { name: "Save Anyway" }).click();

    // Wait for navigation to connector file, then for the editor to appear
    await expect(page).toHaveURL(/.*\/files\/connectors\/.*\.yaml/, {
      timeout: 5000,
    });
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toBeVisible({ timeout: 5000 });

    await expect(codeEditor).toBeVisible();
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: clickhouse");
  });
});
