import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions.ts";
import { test } from "./setup/base";

test.describe.serial("Reports", () => {
  test("Should create report", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb/explore/auction_explore");

    // Enter dimension table "App Site Name"
    await adminPage.getByText("App Site Domain").click();

    // Now and then clicking "App Site Domain" results in a tooltip being shown for a column in the dimension table.
    // This tooltip blocks the export button causing the test to fail.
    // So hover over "select all" to get rid of this tooltip.
    await adminPage.getByText("Select all").hover();

    // Open scheduled report dialog
    await adminPage.getByLabel("Export dimension table data").click();
    await adminPage
      .getByRole("menuitem", { name: "Create scheduled report..." })
      .click();

    const reportForm = adminPage.locator("form#scheduled-report-form");

    // Set the name
    await reportForm.getByTitle("Report title").fill("Report for last 14 days");

    // Set as a daily report
    await reportForm.getByLabel("Frequency").click();
    await adminPage.getByRole("option", { name: "Daily" }).click();
    // Set to run at 10:00 pm
    await reportForm.getByLabel("Time", { exact: true }).click();
    await adminPage.getByRole("option", { name: "10:00 PM" }).click();

    // Select "Last 14 Days" as time range
    await interactWithTimeRangeMenu(reportForm, async () => {
      await reportForm.getByRole("menuitem", { name: "Last 14 Days" }).click();
    });
    // Enable time comparison
    await reportForm.getByLabel("Toggle time comparison").click();

    // Create the report
    await adminPage.getByLabel("Create report").click();

    // Notification is shown
    await expect(adminPage.getByLabel("Notification")).toHaveText(
      "Report created Go to scheduled reports",
    );
    // Clicking "Go to scheduled reports" takes us to the reports page
    await adminPage
      .getByRole("link", { name: "Go to scheduled reports" })
      .click();

    // Go to the newly created report
    await adminPage
      .getByRole("link", {
        name: "Report for last 14 days",
      })
      .click();

    // Assert that report is created with correct fields
    // Assert report name
    await expect(adminPage.getByLabel("Report name")).toHaveText(
      "Report for last 14 days",
    );
    // Assert report dashboard
    await expect(adminPage.getByLabel("Report dashboard name")).toHaveText(
      "Dashboard Programmatic Ads Auction",
    );
    // Assert report schedule
    await expect(adminPage.getByLabel("Report schedule")).toHaveText(
      /Repeats\s+At 10:00 PM, every day/m,
    );
  });

  test("Should edit report", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb/-/reports");

    await adminPage
      .getByRole("link", {
        name: "Report for last 14 days",
      })
      .click();

    // Update the report
    await adminPage.getByLabel("Report context menu").click();
    await adminPage.getByRole("menuitem", { name: "Edit Report" }).click();

    const reportForm = adminPage.locator("form#scheduled-report-form");

    // Set as a monthly report
    await reportForm.getByLabel("Frequency").click();
    await adminPage.getByRole("option", { name: "Monthly" }).click();

    // Select "Last 4 Weeks" as time range
    await interactWithTimeRangeMenu(reportForm, async () => {
      await reportForm.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    });

    // Add "Ad Size" filter
    await reportForm.getByLabel("Add filter button").click();
    await reportForm.getByRole("menuitem", { name: "Ad Size" }).click();
    // Add filters for 1024x768, 120x600, 160x600
    await reportForm.getByRole("menuitem", { name: "1024x768" }).click();
    await reportForm.getByRole("menuitem", { name: "120x600" }).click();
    await reportForm.getByRole("menuitem", { name: "160x600" }).click();
    await reportForm.getByLabel("Open ad_size filter").click();

    // Save the report
    await adminPage.getByLabel("Save report").click();

    // Notification is shown
    await expect(adminPage.getByLabel("Notification")).toHaveText(
      "Report edited",
    );

    // Assert that report is updated with correct schedule
    await expect(adminPage.getByLabel("Report schedule")).toHaveText(
      /Repeats\s+At 10:00 PM, on the 1st of each month/m,
    );
  });

  test("Should delete report", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb/-/reports");

    await adminPage.goto("/e2e/openrtb/-/reports");

    await adminPage
      .getByRole("link", {
        name: "Report for last 14 days",
      })
      .click();

    // Delete the report
    await adminPage.getByLabel("Report context menu").click();
    await adminPage.getByRole("menuitem", { name: "Delete Report" }).click();

    // Back to listing page without any alerts
    await expect(
      adminPage.getByText("You don't have any reports yet"),
    ).toBeVisible();
  });
});
