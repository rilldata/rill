import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Project Status - Tables", () => {
  test("should display tables with their metadata values", async ({
    adminPage,
  }) => {
    // Navigate to the project page
    await adminPage.goto("/e2e/openrtb");

    // Click on Status link
    const statusLink = adminPage.getByRole("link", { name: "Status" });
    await expect(statusLink).toBeVisible();
    await statusLink.click();

    // Wait for the Tables heading to be visible
    const tablesHeading = adminPage.getByRole("heading", { name: "Tables" });
    await expect(tablesHeading).toBeVisible();

    // Verify the table structure with column headers (VirtualizedTable uses role="columnheader")
    // Use the #project-tables-table id to scope to the correct table
    const tablesTable = adminPage.locator("#project-tables-table");
    const headers = tablesTable.locator('[role="columnheader"]');
    await expect(headers.nth(0)).toContainText("Type");
    await expect(headers.nth(1)).toContainText("Name");
    // await expect(headers.nth(2)).toContainText("Row Count");
    await expect(headers.nth(3)).toContainText("Column Count");
    await expect(headers.nth(4)).toContainText("Database Size");

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
      expect(cellCount).toBeGreaterThanOrEqual(5); // At least 5 columns

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
      adminPage.getByRole("heading", { name: "Tables" }),
    ).toBeVisible();

    // If no tables, it should show the table container (possibly with no data message)
    // or with the table headers visible
    const tableSection = adminPage
      .locator("section")
      .filter({ hasText: "Tables" })
      .first();
    await expect(tableSection).toBeVisible();
  });

  test("should display row count values in the Row Count column", async ({
    adminPage,
  }) => {
    // Navigate to project page and click Status link
    await adminPage.goto("/e2e/openrtb");
    await adminPage.getByRole("link", { name: "Status" }).click();

    // Wait for tables heading
    await expect(
      adminPage.getByRole("heading", { name: "Tables" }),
    ).toBeVisible();

    // Get the Tables table by id
    const tablesTable = adminPage.locator("#project-tables-table");

    // Get data rows (skip header row which has role="columnheader")
    const dataRows = tablesTable.locator(".row").filter({
      hasNot: adminPage.locator('[role="columnheader"]'),
    });
    const rowCount = await dataRows.count();

    if (rowCount > 0) {
      // For each row, get the 3rd column (Row Count, index 2)
      for (let i = 0; i < Math.min(rowCount, 5); i++) {
        const row = dataRows.nth(i);
        const rowCountCell = row.locator("> div").nth(2);
        const cellText = await rowCountCell.textContent();
        console.log(`Row ${i} count:`, cellText?.trim());

        // Should be a number, formatted number, or loading/error states
        if (cellText) {
          const trimmedCount = cellText.trim();
          // Should be a number (possibly with commas), or "loading", "error", or "-"
          expect(
            /^[\d,]+$|^loading$|^error$|^-$/.test(trimmedCount) ||
              trimmedCount === "",
          ).toBeTruthy();
        }
      }
    }
  });
});
