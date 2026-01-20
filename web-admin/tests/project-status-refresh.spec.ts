import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Project Status - Resource Refresh (openrtb)", () => {
  test.beforeEach(async ({ adminPage }) => {
    // Navigate to the project status page
    await adminPage.goto("/e2e/openrtb/-/status");
    // Wait for the resources table to load with actual data
    await expect(adminPage.getByText("Resources")).toBeVisible();
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
    await modelRow.getByRole("button").click();

    // Verify "Full Refresh" is visible
    await expect(
      adminPage.getByRole("menuitem", { name: "Full Refresh" }),
    ).toBeVisible();
  });

  test("should show Full Refresh option for sources", async ({ adminPage }) => {
    // Find a source row and click its actions menu
    const sourceRow = adminPage.locator(".row").filter({
      hasText: "auction_data_raw",
    });
    await sourceRow.getByRole("button").click();

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
    await modelRow.getByRole("button").click();

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
    await modelRow.getByRole("button").click();

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
    await modelRow.getByRole("button").click();

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

test.describe("Project Status - Incremental Model Refresh (incremental-test)", () => {
  test.beforeEach(async ({ adminPage }) => {
    // Navigate to the incremental-test project status page
    await adminPage.goto("/e2e/incremental-test/-/status");
    // Wait for the resources table to load with actual data
    await expect(adminPage.getByText("Resources")).toBeVisible();
    // Wait for the specific model rows to appear (indicates data is loaded)
    // Note: VirtualizedTable uses div.row elements, not role="row"
    await expect(adminPage.getByText("success_partition")).toBeVisible({
      timeout: 60_000,
    });
  });

  test("should show Incremental Refresh option for incremental models", async ({
    adminPage,
  }) => {
    // Find the success_partition model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "success_partition",
    });
    await modelRow.getByRole("button").click();

    // Verify both "Full Refresh" and "Incremental Refresh" are visible
    await expect(
      adminPage.getByRole("menuitem", { name: "Full Refresh" }),
    ).toBeVisible();
    await expect(
      adminPage.getByRole("menuitem", { name: "Incremental Refresh" }),
    ).toBeVisible();
  });

  test("should show correct dialog for Incremental Refresh", async ({
    adminPage,
  }) => {
    // Find the success_partition model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "success_partition",
    });
    await modelRow.getByRole("button").click();

    // Click "Incremental Refresh"
    await adminPage
      .getByRole("menuitem", { name: "Incremental Refresh" })
      .click();

    // Verify the dialog shows "Incremental Refresh" in the title
    await expect(adminPage.getByRole("alertdialog")).toBeVisible();
    await expect(
      adminPage.getByRole("heading", { name: /Incremental Refresh/ }),
    ).toBeVisible();

    // Verify the message about updating dependent resources
    await expect(
      adminPage.getByText("will update all dependent resources"),
    ).toBeVisible();

    // Close dialog by clicking cancel
    await adminPage.getByRole("button", { name: "Cancel" }).click();
    await expect(adminPage.getByRole("alertdialog")).not.toBeVisible();
  });

  test("should show Refresh Errored Partitions for models with errored partitions", async ({
    adminPage,
  }) => {
    // Find the failed_partition model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "failed_partition",
    });
    await modelRow.getByRole("button").click();

    // Verify "Refresh Errored Partitions" is visible for models with errors
    await expect(
      adminPage.getByRole("menuitem", { name: "Refresh Errored Partitions" }),
    ).toBeVisible();
  });

  test("should show correct dialog for Refresh Errored Partitions", async ({
    adminPage,
  }) => {
    // Find the failed_partition model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "failed_partition",
    });
    await modelRow.getByRole("button").click();

    // Click "Refresh Errored Partitions"
    await adminPage
      .getByRole("menuitem", { name: "Refresh Errored Partitions" })
      .click();

    // Verify the dialog shows "Refresh Errored Partitions" in the title
    await expect(adminPage.getByRole("alertdialog")).toBeVisible();
    await expect(
      adminPage.getByRole("heading", { name: /Refresh Errored Partitions/ }),
    ).toBeVisible();

    // Verify the message about re-running failed partitions
    await expect(
      adminPage.getByText("re-run all partitions that failed"),
    ).toBeVisible();

    // Close dialog by clicking cancel
    await adminPage.getByRole("button", { name: "Cancel" }).click();
    await expect(adminPage.getByRole("alertdialog")).not.toBeVisible();
  });

  test("should not show Refresh Errored Partitions for successful incremental models", async ({
    adminPage,
  }) => {
    // Find the success_partition model row and click its actions menu
    const modelRow = adminPage.locator(".row").filter({
      hasText: "success_partition",
    });
    await modelRow.getByRole("button").click();

    // "Refresh Errored Partitions" should NOT be visible for models without errors
    await expect(
      adminPage.getByRole("menuitem", { name: "Refresh Errored Partitions" }),
    ).not.toBeVisible();
  });
});
