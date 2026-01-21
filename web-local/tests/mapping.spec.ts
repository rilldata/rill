import { expect } from "@playwright/test";
import { test } from "./setup/base";
import { waitForFileNavEntry, fileNotPresent } from "./utils/waitHelpers";
import { deleteFile } from "./utils/commonHelpers";

test.describe("mapping workspace", () => {
  test.use({ project: "Blank" });

  test("Create new mapping file from Add menu", async ({ page }) => {
    // Click Add button
    await page.getByRole("button", { name: "Add Asset" }).click();

    // Navigate to More submenu
    await page.getByRole("menuitem", { name: "More" }).click();

    // Click Mapping file
    await page.getByRole("menuitem", { name: "Mapping file" }).click();

    // Should navigate to mapping workspace
    await page.waitForURL(/\/mapping\/new/);

    // Should see the default filename
    await expect(page.locator("#mapping-title-input")).toHaveValue(
      "mapping.csv",
    );

    // Should see Add Column and Add Row buttons
    await expect(
      page.getByRole("button", { name: "Add Column" }),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "Add Row" })).toBeVisible();

    // Should see Save button
    await expect(page.getByRole("button", { name: "Save" })).toBeVisible();

    // Should see default columns
    const columnHeaders = page.locator("thead input");
    await expect(columnHeaders).toHaveCount(2);
  });

  test("Edit cells and columns", async ({ page }) => {
    // Navigate to mapping workspace
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Edit first column name
    const firstColumnInput = page.locator("thead input").first();
    await firstColumnInput.click();
    await firstColumnInput.fill("name");

    // Edit second column name
    const secondColumnInput = page.locator("thead input").nth(1);
    await secondColumnInput.click();
    await secondColumnInput.fill("value");

    // Edit first cell
    const firstCell = page.locator("tbody input").first();
    await firstCell.click();
    await firstCell.fill("test_name");

    // Edit second cell
    const secondCell = page.locator("tbody input").nth(1);
    await secondCell.click();
    await secondCell.fill("test_value");

    // Verify values are set
    await expect(firstColumnInput).toHaveValue("name");
    await expect(secondColumnInput).toHaveValue("value");
    await expect(firstCell).toHaveValue("test_name");
    await expect(secondCell).toHaveValue("test_value");
  });

  test("Add and remove columns", async ({ page }) => {
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Initially 2 columns
    await expect(page.locator("thead input")).toHaveCount(2);

    // Add a column
    await page.getByRole("button", { name: "Add Column" }).click();
    await expect(page.locator("thead input")).toHaveCount(3);

    // Remove a column (click the trash icon in header)
    await page.locator("thead button").first().click();
    await expect(page.locator("thead input")).toHaveCount(2);
  });

  test("Add and remove rows", async ({ page }) => {
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Initially 1 row
    await expect(page.locator("tbody tr")).toHaveCount(1);

    // Add a row
    await page.getByRole("button", { name: "Add Row" }).click();
    await expect(page.locator("tbody tr")).toHaveCount(2);

    // Add another row
    await page.getByRole("button", { name: "Add Row" }).click();
    await expect(page.locator("tbody tr")).toHaveCount(3);

    // Remove a row (click trash icon)
    await page.locator("tbody tr").first().locator("button").click();
    await expect(page.locator("tbody tr")).toHaveCount(2);
  });

  test("Tab creates new row at last cell", async ({ page }) => {
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Initially 1 row
    await expect(page.locator("tbody tr")).toHaveCount(1);

    // Focus the last cell (second column of first row)
    const lastCell = page.locator("tbody input").last();
    await lastCell.click();

    // Press Tab
    await page.keyboard.press("Tab");

    // Should create a new row
    await expect(page.locator("tbody tr")).toHaveCount(2);

    // First cell of new row should be focused
    const newRowFirstCell = page.locator(
      'tbody input[data-row="1"][data-col="0"]',
    );
    await expect(newRowFirstCell).toBeFocused();
  });

  test("Ctrl+A selects all text in cell", async ({ page }) => {
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Fill a cell with text
    const firstCell = page.locator("tbody input").first();
    await firstCell.click();
    await firstCell.fill("test content");

    // Use Ctrl+A (or Cmd+A on Mac)
    const modifier = process.platform === "darwin" ? "Meta" : "Control";
    await page.keyboard.press(`${modifier}+a`);

    // Text should be selected - we can verify by typing and seeing it replaces
    await page.keyboard.type("replaced");
    await expect(firstCell).toHaveValue("replaced");
  });

  test("Save creates CSV and YAML files", async ({ page }) => {
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Edit column names
    const firstColumnInput = page.locator("thead input").first();
    await firstColumnInput.click();
    await firstColumnInput.fill("city");

    const secondColumnInput = page.locator("thead input").nth(1);
    await secondColumnInput.click();
    await secondColumnInput.fill("country");

    // Edit cells
    const firstCell = page.locator("tbody input").first();
    await firstCell.click();
    await firstCell.fill("Paris");

    const secondCell = page.locator("tbody input").nth(1);
    await secondCell.click();
    await secondCell.fill("France");

    // Click Save
    await page.getByRole("button", { name: "Save" }).click();

    // Should navigate to the model file
    await page.waitForURL(/\/files\/models\/mapping\.yaml/);

    // CSV file should exist in nav
    await waitForFileNavEntry(page, "/data/mapping.csv", false);

    // Model file should exist in nav
    await waitForFileNavEntry(page, "/models/mapping.yaml", false);

    // Clean up
    await deleteFile(page, "/data/mapping.csv");
    await deleteFile(page, "/models/mapping.yaml");
  });

  test("Auto-increment filename when files exist", async ({ page }) => {
    // Create first mapping
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    await page.getByRole("button", { name: "Save" }).click();
    await page.waitForURL(/\/files\/models\/mapping\.yaml/);

    // Create second mapping - should auto-increment
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    // Title should show mapping_1.csv
    await expect(page.locator("#mapping-title-input")).toHaveValue(
      "mapping_1.csv",
    );

    await page.getByRole("button", { name: "Save" }).click();
    await page.waitForURL(/\/files\/models\/mapping_1\.yaml/);

    // Verify both files exist
    await waitForFileNavEntry(page, "/data/mapping.csv", false);
    await waitForFileNavEntry(page, "/data/mapping_1.csv", false);
    await waitForFileNavEntry(page, "/models/mapping.yaml", false);
    await waitForFileNavEntry(page, "/models/mapping_1.yaml", false);

    // Clean up
    await deleteFile(page, "/data/mapping.csv");
    await deleteFile(page, "/data/mapping_1.csv");
    await deleteFile(page, "/models/mapping.yaml");
    await deleteFile(page, "/models/mapping_1.yaml");
  });

  test("Edit existing CSV file", async ({ page }) => {
    // First create a mapping file
    await page.goto("/mapping/new");
    await page.waitForURL(/\/mapping\/new/);

    const firstCell = page.locator("tbody input").first();
    await firstCell.click();
    await firstCell.fill("original");

    await page.getByRole("button", { name: "Save" }).click();
    await page.waitForURL(/\/files\/models\/mapping\.yaml/);

    // Navigate to CSV file
    await page.getByLabel("/data/mapping.csv Nav Entry").click();
    await page.waitForURL(/\/files\/data\/mapping\.csv/);

    // Click Edit in Table button
    await page.getByRole("button", { name: "Edit in Table" }).click();
    await page.waitForURL(/\/mapping\/edit\/data\/mapping\.csv/);

    // Verify the data is loaded
    const loadedCell = page.locator("tbody input").first();
    await expect(loadedCell).toHaveValue("original");

    // Edit the cell
    await loadedCell.click();
    await loadedCell.fill("modified");

    // Save
    await page.getByRole("button", { name: "Save" }).click();
    await page.waitForURL(/\/files\/models\/mapping\.yaml/);

    // Go back to CSV and verify via Edit in Table
    await page.getByLabel("/data/mapping.csv Nav Entry").click();
    await page.waitForURL(/\/files\/data\/mapping\.csv/);
    await page.getByRole("button", { name: "Edit in Table" }).click();
    await page.waitForURL(/\/mapping\/edit\/data\/mapping\.csv/);

    await expect(page.locator("tbody input").first()).toHaveValue("modified");

    // Clean up
    await page.goto("/");
    await deleteFile(page, "/data/mapping.csv");
    await deleteFile(page, "/models/mapping.yaml");
  });
});
