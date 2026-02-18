import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Project Status - Resource Refresh (openrtb)", () => {
  // Increase timeout for tests that interact with virtualized tables
  test.setTimeout(60_000);

  test.beforeEach(async ({ adminPage }) => {
    // Navigate to the project status resources page
    await adminPage.goto("/e2e/openrtb/-/status/resources");
    // Wait for the resources heading to load
    await expect(
      adminPage.getByRole("heading", { name: "Resources" }),
    ).toBeVisible();
    // Wait for a specific model name to appear (indicates data is loaded)
    // Note: VirtualizedTable uses div.row elements, not role="row"
    await expect(adminPage.getByText("auction_data_model")).toBeVisible({
      timeout: 60_000,
    });
  });

  test("should show Full Refresh option for models", async ({ adminPage }) => {
    // Find the auction_data_model row and click its actions menu
    // The row contains both the model name and a "Model" badge
    const modelRow = adminPage.locator(".row").filter({
      hasText: "auction_data_model",
    });
    // Target the dropdown menu trigger specifically (rows may have multiple buttons)
    await modelRow.locator("[data-melt-dropdown-menu-trigger]").click();

    // Verify "Full Refresh" is visible
    await expect(
      adminPage.getByRole("menuitem", { name: "Full Refresh" }),
    ).toBeVisible();
  });

  test("should show Full Refresh option for sources", async ({ adminPage }) => {
    // Wait for the source row to be visible before interacting
    // Look for bids_data_raw source which should be in the openrtb test project
    await expect(adminPage.getByText("bids_data_raw")).toBeVisible({
      timeout: 60_000,
    });

    // Find a source row and click its actions menu
    const sourceRow = adminPage.locator(".row").filter({
      hasText: "bids_data_raw",
    });
    // Target the dropdown menu trigger specifically (rows may have multiple buttons)
    await sourceRow.locator("[data-melt-dropdown-menu-trigger]").click();

    // Verify "Full Refresh" is visible
    await expect(
      adminPage.getByRole("menuitem", { name: "Full Refresh" }),
    ).toBeVisible();

    // Incremental Refresh should NOT be visible for sources
    await expect(
      adminPage.getByRole("menuitem", { name: "Incremental Refresh" }),
    ).not.toBeVisible();
  });

  test("should not show Incremental Refresh for non-incremental models", async ({
    adminPage,
  }) => {
    // Find the auction_data_model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "auction_data_model",
    });
    // Target the dropdown menu trigger specifically (rows may have multiple buttons)
    await modelRow.locator("[data-melt-dropdown-menu-trigger]").click();

    // Verify "Full Refresh" is visible
    await expect(
      adminPage.getByRole("menuitem", { name: "Full Refresh" }),
    ).toBeVisible();

    // For non-incremental models, "Incremental Refresh" should not be visible
    await expect(
      adminPage.getByRole("menuitem", { name: "Incremental Refresh" }),
    ).not.toBeVisible();
  });

  test("should not show Refresh Errored Partitions for models without errors", async ({
    adminPage,
  }) => {
    // Find the auction_data_model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "auction_data_model",
    });
    // Target the dropdown menu trigger specifically (rows may have multiple buttons)
    await modelRow.locator("[data-melt-dropdown-menu-trigger]").click();

    // "Refresh Errored Partitions" should not be visible for models without errored partitions
    await expect(
      adminPage.getByRole("menuitem", { name: "Refresh Errored Partitions" }),
    ).not.toBeVisible();
  });

  test("should show correct dialog for Full Refresh", async ({ adminPage }) => {
    // Find the auction_data_model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "auction_data_model",
    });
    // Target the dropdown menu trigger specifically (rows may have multiple buttons)
    await modelRow.locator("[data-melt-dropdown-menu-trigger]").click();

    // Click "Full Refresh"
    await adminPage.getByRole("menuitem", { name: "Full Refresh" }).click();

    // Verify the dialog shows "Full Refresh" in the title
    await expect(adminPage.getByRole("alertdialog")).toBeVisible();
    await expect(
      adminPage.getByRole("heading", { name: /Full Refresh/ }),
    ).toBeVisible();

    // Verify the warning message about full refresh
    await expect(
      adminPage.getByText(/Warning.*will re-ingest ALL data from scratch/),
    ).toBeVisible();

    // Close dialog by clicking cancel
    await adminPage.getByRole("button", { name: "Cancel" }).click();
    await expect(adminPage.getByRole("alertdialog")).not.toBeVisible();
  });
});
