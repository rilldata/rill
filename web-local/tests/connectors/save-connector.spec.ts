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

    // Wait for the form to load and Host field to be visible
    // (Host is visible when connection type is "cloud" which is the default)
    await page.waitForSelector('form[id*="clickhouse"]');
    const hostField = page.getByRole("textbox", { name: "Host" });
    await expect(hostField).toBeVisible();

    // Fill in connection details with invalid values
    await hostField.fill("asd");
    await page.getByRole("textbox", { name: "Password" }).fill("asd");

    // Save button should be visible on the connector step
    const saveButton = page.getByRole("button", { name: "Save", exact: true });
    await saveButton.scrollIntoViewIfNeeded();
    await expect(saveButton).toBeVisible();

    // Click "Save" button to save connector without testing connection
    await saveButton.click();

    // Wait for navigation to connector file, then for the editor to appear.
    // Use generous timeouts: invalidate("app:init") re-runs the root layout
    // load, which can briefly unmount and remount the editor component.
    await expect(page).toHaveURL(/.*\/files\/connectors\/.*\.yaml/, {
      timeout: 10_000,
    });
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toBeVisible({ timeout: 10_000 });
    await expect(codeEditor).toContainText("type: connector", {
      timeout: 10_000,
    });
    await expect(codeEditor).toContainText("driver: clickhouse", {
      timeout: 10_000,
    });
  });
});
