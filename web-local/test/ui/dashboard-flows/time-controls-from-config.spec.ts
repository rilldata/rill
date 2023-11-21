import { expect, test } from "@playwright/test";
import { useDashboardFlowTestSetup } from "web-local/test/ui/dashboard-flows/dashboard-flow-test-setup";
import { updateCodeEditor } from "web-local/test/ui/utils/commonHelpers";
import {
  interactWithTimeRangeMenu,
  waitForDashboard,
} from "web-local/test/ui/utils/dashboardHelpers";
import { startRuntimeForEachTest } from "web-local/test/ui/utils/startRuntimeForEachTest";

test.describe("time controls settings from dashboard config", () => {
  startRuntimeForEachTest();
  // dashboard test setup
  useDashboardFlowTestSetup();

  test("default_time_range", async ({ page }) => {
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set a time range that is one of the supported presets
    await updateCodeEditor(page, getDashboardYaml(`default_time_range: "P4W"`));
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Time range has changed
    await expect(page.getByText("Last 4 Weeks")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total rows 26.7k")).toBeVisible();
    await expect(page.getByText("Facebook 7.0k")).toBeVisible();
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set a time range that is one of the period to date preset
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "rill-WTD"`)
    );
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Time range has changed
    await expect(page.getByText("Week to Date")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total rows 3.7k")).toBeVisible();
    await expect(page.getByText("Facebook 948")).toBeVisible();

    // Select a different time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 7 Days" }).click();
    });
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 7 Days" })
    ).not.toBeVisible();
    // Data has changed
    await expect(page.getByText("Total rows 7.9k")).toBeVisible();
    await expect(page.getByText("Facebook 2.0k")).toBeVisible();

    // Last 2 weeks is still available in the menu
    // Select a different time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 2 Weeks" }).click();
    });
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 2 Weeks" })
    ).not.toBeVisible();
    // Data has changed
    await expect(page.getByText("Total rows 11.2k")).toBeVisible();
    await expect(page.getByText("Facebook 2.9k")).toBeVisible();

    // Set a time range that is not one of the supported presets
    await updateCodeEditor(page, getDashboardYaml(`default_time_range: "P2W"`));
    await waitForDashboard(page);
    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Time range has changed
    await expect(page.getByText("Last 2 Weeks")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total rows 11.2k")).toBeVisible();
    await expect(page.getByText("Facebook 2.9k")).toBeVisible();
  });

  test("default_comparison", async ({ page }) => {
    await page.getByRole("button", { name: "Edit metrics" }).click();

    // Set comparison to time
    await updateCodeEditor(
      page,
      getDashboardYaml(`default_time_range: "P4W"
default_comparison:
  mode: time
`)
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
`)
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
`)
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
available_time_ranges:
  - PT6H
  - range: P5D
    comparison_offsets:
      - rill-PP
      - rill-PW
  - P4W
  - rill-WTD
  - rill-MTD
`)
    );
  });
});

function getDashboardYaml(defaults: string) {
  return `
# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

title: "AdBids_model_dashboard_rename"
model: "AdBids_model"
smallest_time_grain: "week"
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
