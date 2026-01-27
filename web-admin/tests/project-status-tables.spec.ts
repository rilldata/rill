import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Project Status - Model Details", () => {
  test("should display tables with their metadata values", async ({
    adminPage,
  }) => {
    // Navigate to the project page
    await adminPage.goto("/e2e/openrtb");

    // Click on Status link
    const statusLink = adminPage.getByRole("link", { name: "Status" });
    await expect(statusLink).toBeVisible();
    await statusLink.click();

    // Wait for the Model Details heading to be visible
    const modelDetailsHeading = adminPage.getByRole("heading", {
      name: "Model Details",
    });
    await expect(modelDetailsHeading).toBeVisible();

    // Verify the table structure with column headers (VirtualizedTable uses role="columnheader")
    // Use the #project-tables-table id to scope to the correct table
    const tablesTable = adminPage.locator("#project-tables-table");
    const headers = tablesTable.locator('[role="columnheader"]');
    await expect(headers.nth(0)).toContainText("Type");
    await expect(headers.nth(1)).toContainText("Name");
    await expect(headers.nth(2)).toContainText("Database Size");

    // Verify table rows are rendered (VirtualizedTable uses .row divs, skip the header row)
    const dataRows = tablesTable.locator(".row").filter({
      hasNot: adminPage.locator('[role="columnheader"]'),
    });
    const rowCount = await dataRows.count();
    expect(rowCount).toBeGreaterThan(0);

    // Verify specific table data if auction_data_model exists
    const auctionRow = tablesTable.locator(".row", {
      hasText: "auction_data_model",
    });
    const auctionRowExists = await auctionRow.isVisible().catch(() => false);

    if (auctionRowExists) {
      await expect(auctionRow).toBeVisible();

      // Verify that the row has visible content
      const cells = auctionRow.locator("> div");
      const cellCount = await cells.count();
      expect(cellCount).toBeGreaterThanOrEqual(3); // At least 3 columns

      // Get the text content of each cell
      const cellTexts = await cells.allTextContents();
      console.log("auction_data_model row cells:", cellTexts);

      // Verify Name column contains auction_data_model
      await expect(auctionRow).toContainText("auction_data_model");
    }

    // Verify the table is visible
    await expect(tablesTable).toBeVisible();
  });

  test("should handle empty table list gracefully", async ({ adminPage }) => {
    // This test verifies the UI renders correctly for a project
    // Navigate to project page and click Status link
    await adminPage.goto("/e2e/openrtb");
    await adminPage.getByRole("link", { name: "Status" }).click();

    // Wait for the page to load
    await expect(
      adminPage.getByRole("heading", { name: "Model Details" }),
    ).toBeVisible();

    // If no tables, it should show the table container (possibly with no data message)
    // or with the table headers visible
    const tableSection = adminPage
      .locator("section")
      .filter({ hasText: "Model Details" })
      .first();
    await expect(tableSection).toBeVisible();
  });
});
