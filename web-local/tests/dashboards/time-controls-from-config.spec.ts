import { expect } from "@playwright/test";
import { useDashboardFlowTestSetup } from "web-local/tests/dashboards/dashboard-flow-test-setup";
import { updateCodeEditor } from "web-local/tests/utils/commonHelpers";
import {
  interactWithTimeRangeMenu,
  waitForDashboard,
} from "web-local/tests/utils/dashboardHelpers";
import { test } from "../utils/test";

test.describe("time controls settings from dashboard config", () => {
  // dashboard test setup
  useDashboardFlowTestSetup();

  test("default_time_range", async ({ page }) => {
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set a time range that is one of the supported presets
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P4W"
default_comparison:
  mode: time
`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Time range has changed
    await expect(page.getByText("Last 4 Weeks")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total rows 26.7k -4.7k -15%")).toBeVisible();
    await expect(page.getByText("Facebook 7.0k 67%")).toBeVisible();
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set a time range that is one of the period to date preset
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "rill-WTD"
default_comparison:
  mode: time
`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Time range has changed
    await expect(page.getByText("Week to Date")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total rows 3.4k +156 5%")).toBeVisible();
    await expect(page.getByText("Facebook 889 4%")).toBeVisible();

    // Select a different time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 7 Days" }).click();
    });
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 7 Days" }),
    ).not.toBeVisible();
    // Data has changed
    await expect(page.getByText("Total rows 7.9k -15 ~0%")).toBeVisible();
    await expect(page.getByText("Facebook 2.0k -2%")).toBeVisible();
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set a time range that is not one of the supported presets
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P2W"
default_comparison:
  mode: time
`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Time range has changed
    await expect(page.getByText("Last 2 Weeks")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total rows 11.2k -4.4k -28%")).toBeVisible();
    await expect(page.getByText("Facebook 2.9k -29%")).toBeVisible();
  });

  test("default_comparison", async ({ page }) => {
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set comparison to time
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P4W"
default_comparison:
  mode: time
`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();
    // Comparison is selected
    await expect(page.getByText("Comparing by Time")).toBeVisible();
    // Go back to metrics editor
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set comparison to dimension
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P4W"
default_comparison:
  mode: dimension
  dimension: publisher
`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();
    // Comparison is selected
    await expect(page.getByText("Comparing by Publisher")).toBeVisible();
    // Go back to metrics editor
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set comparison to none
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P4W"
default_comparison:
  mode: none
`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();
    // No Comparison
    await expect(page.getByText("No Comparison")).toBeVisible();
  });

  test("available_time_ranges", async ({ page }) => {
    await page.getByRole("button", { name: "Edit metrics" }).click();
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P4W"
default_comparison:
  mode: time
available_time_ranges:
  - PT6H
  - range: P5D
    comparison_offsets:
      - rill-PP
      - rill-PW
  - P4W
  - rill-WTD
  - rill-MTD`),
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Open the time range menu
    await page.getByLabel("Select time range").click();
    // Assert the options available
    await Promise.all(
      [
        "All Time",
        "Last 6 Hours",
        "Last 5 Days",
        "Last 4 Weeks",
        "Week To Date",
        "Month To Date",
      ].map((label) =>
        expect(page.getByRole("menuitem", { name: label })).toBeVisible(),
      ),
    );
    // Select Last 6 hours
    await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    // Wait for time range menu to close
    await expect(
      page.getByRole("menu", { name: "Time range selector" }),
    ).not.toBeVisible();
    // Assert data has changed
    await expect(page.getByText("Total rows 272 -23 -8%")).toBeVisible();
    await expect(page.getByText("Facebook 68 -4%")).toBeVisible();

    // Open the time comparison
    await page.getByLabel("Select time comparison option").click();
    // Assert the options available
    await Promise.all(
      [
        "Previous Period",
        "Previous day",
        "Previous week",
        "Previous month",
      ].map((label) =>
        expect(page.getByRole("menuitem", { name: label })).toBeVisible(),
      ),
    );
    // Select Previous week
    await page.getByRole("menuitem", { name: "Previous week" }).click();
    // Wait for time range menu to close
    await expect(
      page.getByRole("menu", { name: "Time comparison selector" }),
    ).not.toBeVisible();
    // Assert data has changed
    await expect(page.getByText("Total rows 272 -18 -6%")).toBeVisible();
    await expect(page.getByText("Facebook 68 -26%")).toBeVisible();

    // Select Last 5 days
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 5 Days" }).click();
    });
    // Assert data has changed
    await expect(page.getByText("Total rows 5.6k +16 ~0%")).toBeVisible();
    await page.pause();
    await expect(page.getByText("Facebook 1.5k -2%")).toBeVisible();

    // Open the time comparison
    await page.getByLabel("Select time comparison option").click();
    // Assert the options available
    await Promise.all(
      ["Previous Period", "Previous week"].map((label) =>
        expect(page.getByRole("menuitem", { name: label })).toBeVisible(),
      ),
    );
    // Select Last 6 hours
    await page.getByRole("menuitem", { name: "Previous Period" }).click();
    // Wait for time range menu to close
    await expect(
      page.getByRole("menu", { name: "Time comparison selector" }),
    ).not.toBeVisible();
    // Assert data has changed
    await expect(page.getByText("Total rows 5.6k -23 ~0%")).toBeVisible();
    await expect(page.getByText("Facebook 1.5k ~0%")).toBeVisible();
  });
});

function getDashboardYaml(defaults: string) {
  return `
# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

title: "AdBids_model_dashboard_rename"
model: "AdBids_model"
timeseries: "timestamp"
${defaults}
measures:
  - label: Total rows
    expression: count(*)
    name: total_rows
    description: Total number of records present
dimensions:
  - name: publisher
    label: Publisher
    column: publisher
    description: ""
  - name: domain
    label: Domain Name
    column: domain
    description: ""
  `;
}
