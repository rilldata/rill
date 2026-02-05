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

    // Wait for the form to load and Host field to be visible
    // (Host is visible when connection type is "cloud" which is the default)
    await page.waitForSelector('form[id*="clickhouse"]');
    const hostField = page.getByRole("textbox", { name: "Host" });
    await expect(hostField).toBeVisible();

    // Fill in connection details with invalid values
    await hostField.fill("asd");
    await page.getByRole("textbox", { name: "Password" }).fill("asd");

    // Click "Test and Connect" - this should fail connection test and show "Save Anyway" button
    // Scope to dialog and use force:true to handle viewport overflow in CI environments.
    const submitButton = page
      .getByRole("dialog")
      .getByRole("button", { name: /^(Test and Connect|Connect)$/ });
    await submitButton.click({ force: true });

    // Wait for "Save Anyway" button to appear
    const saveAnywayButton = page
      .getByRole("dialog")
      .getByRole("button", { name: "Save Anyway" });
    await expect(saveAnywayButton).toBeVisible();

    // Click "Save Anyway" button (force:true for same viewport overflow reason)
    await saveAnywayButton.click({ force: true });

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
