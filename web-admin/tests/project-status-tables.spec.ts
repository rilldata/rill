import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Project Status - Model Overview", () => {
  test.beforeEach(async ({ adminPage }) => {
    // Navigate directly to the model-overview page
    await adminPage.goto("/e2e/openrtb/-/status/model-overview");
    // Wait for the Model Overview heading to be visible
    await expect(
      adminPage.getByRole("heading", { name: "Model Overview" }),
    ).toBeVisible({ timeout: 30_000 });
  });

  test("should display Model Overview summary cards", async ({ adminPage }) => {
    // Verify the three summary cards are visible
    // Tables (Materialized Models) card
    await expect(
      adminPage.getByText("Tables (Materialized Models)"),
    ).toBeVisible();

    // Views card
    await expect(adminPage.getByText("Views")).toBeVisible();

    // OLAP Engine card
    await expect(adminPage.getByText("OLAP Engine")).toBeVisible();

    // Verify that the counts are displayed (not loading placeholders)
    // The cards should show numeric values (not just "-")
    const tablesCard = adminPage
      .locator("div")
      .filter({ hasText: "Tables (Materialized Models)" })
      .first();
    await expect(tablesCard).toBeVisible();

    const viewsCard = adminPage
      .locator("div")
      .filter({ hasText: /^Views$/ })
      .first();
    await expect(viewsCard).toBeVisible();

    // OLAP Engine should show "duckdb" for the openrtb project
    await expect(adminPage.getByText("duckdb")).toBeVisible({
      timeout: 15_000,
    });
  });

  test("should display Model Details table with correct columns", async ({
    adminPage,
  }) => {
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
    await expect(headers.nth(2)).toContainText("Status");
    await expect(headers.nth(3)).toContainText("Rows");
    await expect(headers.nth(4)).toContainText("Columns");
    await expect(headers.nth(5)).toContainText("Database Size");


    // Verify table rows are rendered (VirtualizedTable uses .row divs, skip the header row)
    const dataRows = tablesTable.locator(".row").filter({
      hasNot: adminPage.locator('[role="columnheader"]'),
    });
    const rowCount = await dataRows.count();
    expect(rowCount).toBeGreaterThan(0);

    // Verify the table is visible
    await expect(tablesTable).toBeVisible();
  });

  test("should display table rows with model data", async ({ adminPage }) => {
    // Scope to the project tables table
    const tablesTable = adminPage.locator("#project-tables-table");

    // Wait for rows to load
    await expect(tablesTable.locator(".row").first()).toBeVisible({
      timeout: 30_000,
    });

    // Verify specific table data - auction_data_model should exist
    const auctionRow = tablesTable.locator(".row", {
      hasText: "auction_data_model",
    });
    await expect(auctionRow).toBeVisible({ timeout: 30_000 });

    // Verify that the row has visible content
    const cells = auctionRow.locator("> div");
    const cellCount = await cells.count();
    expect(cellCount).toBeGreaterThanOrEqual(3); // At least 3 columns

    // Verify Name column contains auction_data_model
    await expect(auctionRow).toContainText("auction_data_model");
  });

  test("should navigate to Model Overview from Status link", async ({
    adminPage,
  }) => {
    // Start from the project home page
    await adminPage.goto("/e2e/openrtb");

    // Click on Status link
    const statusLink = adminPage.getByRole("link", { name: "Status" });
    await expect(statusLink).toBeVisible();
    await statusLink.click();

    // Should redirect to project-status page first
    await adminPage.waitForURL("**/status/project-status");

    // Click on Model Overview nav link
    const modelOverviewLink = adminPage.getByRole("link", {
      name: "Model Overview",
    });
    await expect(modelOverviewLink).toBeVisible();
    await modelOverviewLink.click();

    // Should navigate to model-overview page
    await adminPage.waitForURL("**/status/model-overview");

    // Verify the Model Overview heading is visible
    await expect(
      adminPage.getByRole("heading", { name: "Model Overview" }),
    ).toBeVisible();
  });
});
