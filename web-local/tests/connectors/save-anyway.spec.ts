import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Save connector feature", () => {
  test.use({ project: "Blank" });

  test("Save button saves connector and redirects to connector file", async ({
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

    // Save button should be visible on the connector step
    const saveButton = page.getByRole("button", { name: "Save", exact: true });
    await expect(saveButton).toBeVisible();

    // Click "Save" button to save connector without testing connection
    await saveButton.click();

    // Wait for navigation to connector file, then for the editor to appear
    await expect(page).toHaveURL(/.*\/files\/connectors\/.*\.yaml/, {
      // Allow extra time for the file to be written and navigation to occur on slower CI
      timeout: 6_000,
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
