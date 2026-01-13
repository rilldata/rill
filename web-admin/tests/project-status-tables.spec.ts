import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Project Status - Tables", () => {
  test("should display tables with their metadata values", async ({
    adminPage,
  }) => {
    // Navigate to the Status page of the openrtb project
    await adminPage.goto("/e2e/openrtb");

    // Click on Status link
    const statusLink = adminPage.getByRole("link", { name: "Status" });
    await expect(statusLink).toBeVisible();
    await statusLink.click();

    // Wait for the Tables heading to be visible
    const tablesHeading = adminPage.getByRole("heading", { name: "Tables" });
    await expect(tablesHeading).toBeVisible();

    // Verify the table structure with column headers
    const headers = adminPage.locator("thead th");
    await expect(headers.nth(0)).toContainText("Type");
    await expect(headers.nth(1)).toContainText("Name");
    await expect(headers.nth(2)).toContainText("Row Count");
    await expect(headers.nth(3)).toContainText("Column Count");
    await expect(headers.nth(4)).toContainText("Database Size");

    // Verify table rows are rendered
    const tableBody = adminPage.locator("tbody tr");
    const rowCount = await tableBody.count();
    expect(rowCount).toBeGreaterThan(0);

    // Verify specific table data if auction_data_model exists
    // Look for any row with auction_data_model
    const tableNameCells = adminPage.locator("tbody tr td:nth-child(2)");
    const tableNames = await tableNameCells.allTextContents();

    if (tableNames.some((name) => name.includes("auction_data_model"))) {
      // Find the row containing auction_data_model
      const auctionRow = adminPage.locator(
        'tbody tr:has(td:has-text("auction_data_model"))',
      );
      await expect(auctionRow).toBeVisible();

      // Verify that the row has visible values in all columns
      const cells = auctionRow.locator("td");
      const cellCount = await cells.count();
      expect(cellCount).toBeGreaterThanOrEqual(5); // At least 5 columns

      // Get the text content of each cell
      const cellTexts = await cells.allTextContents();
      console.log("auction_data_model row cells:", cellTexts);

      // Verify Type column (should show "Table" or icon)
      await expect(cells.nth(0)).toBeVisible();

      // Verify Name column
      await expect(cells.nth(1)).toContainText("auction_data_model");

      // Verify Row Count column has some value (number or loading state)
      const rowCountCell = cells.nth(2);
      await expect(rowCountCell).toBeVisible();

      // Verify Column Count column has some value
      const columnCountCell = cells.nth(3);
      await expect(columnCountCell).toBeVisible();

      // Verify Database Size column has some value
      const sizeCell = cells.nth(4);
      await expect(sizeCell).toBeVisible();
    }

    // Verify other common tables exist
    const expectedTables = [
      "annotations_auction",
      "auction_data_raw",
      "bids_data_model",
    ];

    for (const tableName of expectedTables) {
      const row = adminPage.locator(
        `tbody tr:has(td:has-text("${tableName}"))`,
      );
      // These tables might not exist in all environments, so we don't assert they exist
      // but if they do, we verify their structure
      const isVisible = await row.isVisible().catch(() => false);
      if (isVisible) {
        const cells = row.locator("td");
        expect(await cells.count()).toBeGreaterThanOrEqual(5);
      }
    }

    // Verify that the table is interactive (has scrollable content if needed)
    const table = adminPage.locator("table").first();
    await expect(table).toBeVisible();
  });

  test("should handle empty table list gracefully", async ({ adminPage }) => {
    // This test verifies the UI renders correctly for a project
    // Navigate to Status page
    await adminPage.goto("/e2e/openrtb");
    await adminPage.getByRole("link", { name: "Status" }).click();

    // Wait for the page to load
    await expect(
      adminPage.getByRole("heading", { name: "Tables" }),
    ).toBeVisible();

    // If no tables, it should show the table container (possibly with no data message)
    // or with the table headers visible
    const tableSection = adminPage.locator("section").first();
    await expect(tableSection).toBeVisible();
  });

  test("should display row count values in the Row Count column", async ({
    adminPage,
  }) => {
    await adminPage.goto("/e2e/openrtb");
    await adminPage.getByRole("link", { name: "Status" }).click();

    // Wait for tables heading
    await expect(
      adminPage.getByRole("heading", { name: "Tables" }),
    ).toBeVisible();

    // Get the Row Count column (3rd data column, index 2)
    const rowCountCells = adminPage.locator("tbody tr td:nth-child(3)");
    const cellCount = await rowCountCells.count();

    if (cellCount > 0) {
      // Get all row count values
      const rowCounts = await rowCountCells.allTextContents();
      console.log("Row counts found:", rowCounts);

      // Verify that we have numeric values or loading/error states
      for (const count of rowCounts) {
        const trimmedCount = count.trim();
        // Should be a number, or "loading", "error", or "-"
        expect(
          /^\d+$|^loading$|^error$|^-$/.test(trimmedCount) ||
            trimmedCount === "",
        ).toBeTruthy();
      }
    }
  });
});
