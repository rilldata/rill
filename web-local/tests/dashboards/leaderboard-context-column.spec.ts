import { expect } from "@playwright/test";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";
import { clickMenuButton } from "../utils/commonHelpers";
import { interactWithTimeRangeMenu } from "../utils/dashboardHelpers";
import { test } from "../utils/test";
import { useDashboardFlowTestSetup } from "./dashboard-flow-test-setup";

test.describe("leaderboard context column", () => {
  useDashboardFlowTestSetup();

  test("Leaderboard context column", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    // reset metrics, and add a metric with `valid_percent_of_total: true`
    const metricsWithValidPercentOfTotal = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

  type: metrics_view
  title: "AdBids_model_dashboard"
  model: "AdBids_model"
  default_time_range: ""
  smallest_time_grain: ""
  timeseries: "timestamp"
  measures:
    - label: Total rows
      expression: count(*)
      name: total_rows
      description: Total number of records present
    - label: Total Bid Price
      expression: sum(bid_price)
      name: total_bid_price
      format_preset: currency_usd
      valid_percent_of_total: true
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
    await watcher.updateAndWaitForDashboard(metricsWithValidPercentOfTotal);

    async function clickMenuItem(itemName: string, wait = true) {
      await clickMenuButton(page, itemName);
      if (wait) {
        await page.getByRole("menu").waitFor({ state: "hidden" });
      }
    }

    const measuresButton = page.getByRole("button", {
      name: "Select a measure to filter by",
    });

    async function escape() {
      await page.keyboard.press("Escape");
      await page.getByRole("menu").waitFor({ state: "hidden" });
    }

    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // make sure "All time" is selected to clear any time comparison
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    const deltaPercentColumn = page.getByLabel(
      "Toggle sort leaderboards by percent change",
    );
    const percentOfTotalColumn = page.getByLabel(
      "Toggle sort leaderboards by percent of total",
    );
    const deltaAbsoluteColumn = page.getByLabel(
      "Toggle sort leaderboards by absolute change",
    );

    // Delta columns not visible since there is no time comparison
    await expect(deltaPercentColumn).not.toBeVisible();
    await expect(deltaAbsoluteColumn).not.toBeVisible();

    // Percent of total column is not visible since `valid_percent_of_total` is not set for the measure "total rows"
    await expect(percentOfTotalColumn).not.toBeVisible();

    /**
     * SUBFLOW: check correct behavior when a time comparison
     * is activated, but there is no valid_percent_of_total
     */

    // Select a time range, that supports comparisons
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });
    // enable comparisons which should automatically enable a time comparison (including context column)
    await page.getByRole("button", { name: "Comparing" }).click();

    // This regex matches a line that:
    // - starts with "Facebook"
    // - has two white space separated sets of characters (the number and the percent change)
    // - ends with a percent sign literal
    // e.g. "Facebook 68.9k -24k -12%".
    // This will detect both percent change and percent of total
    const comparisonColumnRegex = /Facebook\s*\S*\s*\S*\s*\S*%/;

    // Check that time comparison context column is visible with correct value now that there is a time comparison
    await expect(page.getByText(comparisonColumnRegex)).toBeVisible();

    // Delta columns visible
    await expect(deltaPercentColumn).toBeVisible();
    await expect(deltaAbsoluteColumn).toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - the context column is turned back on,
     * - there is no valid_percent_of_total,
     * - and then time comparison is turned off
     */

    await clickMenuItem("Percent change");
    await expect(page.getByText(comparisonColumnRegex)).toBeVisible();

    // click back to "All time" to clear the time comparison
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    // Check that time comparison context column is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();
    await expect(page.getByText("Facebook 19.3k")).toBeVisible();

    // Check that the "percent change" menuitem is disabled
    await expect(deltaPercentColumn).not.toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - switching to a measure with valid_percent_of_total
     * - but no time comparison enabled
     */

    // Switch to measure "total bid price"
    await measuresButton.click();
    await clickMenuItem("Total Bid Price", false);
    await escape();
    await expect(measuresButton).toHaveText("Showing Total Bid Price");

    // Check that the "percent of total" menuitem is enabled
    await expect(percentOfTotalColumn).not.toBeVisible();
    // Check that the "percent change" menuitem is disabled
    await expect(deltaPercentColumn).not.toBeVisible();

    await escape();

    // Check that the percent of total is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - measure with valid_percent_of_total
     * - no time comparison enabled
     * - percent of total context column is turned on
     */

    // check that the percent of total is visible
    await expect(page.getByText("Facebook $57.8k 19%")).toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - measure with valid_percent_of_total
     * - no time comparison enabled
     * - percent of total context column is turned on
     * - and then time comparison is enabled
     */

    // Add a time comparison
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 6 Hours" }),
    ).not.toBeVisible();
    // check that the percent of total remains visible,
    // with updated value for the time comparison
    await expect(page.getByText("Facebook $229.26 29%")).toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - switch context column to percent change
     * - and then switch back to percent of total
     */

    // Need to manually enable comparison since we disabled it
    await page.getByRole("button", { name: "Comparing" }).click();

    await clickMenuItem("Percent change");

    // check that the percent change is visible+correct
    await expect(page.getByText("Facebook $229.26 4%")).toBeVisible();

    await clickMenuItem("Percent of total");

    // check that the percent of total is visible+correct
    await expect(page.getByText("Facebook $229.26 29%")).toBeVisible();

    /**
     * Go back to measure without valid_percent_of_total
     * while percent of total context column is enabled.
     * Make sure the context column is hidden,
     * and the menuitems have the correct enabled/disabled state.
     */

    // Switch to measure "total rows" (no valid_percent_of_total)
    await measuresButton.click();
    await clickMenuItem("Total Rows");
    await expect(measuresButton).toHaveText("Showing Total rows");
    // check that the context column is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();

    // open the context column menu

    // check that the "percent of total" menuitem is disabled
    await expect(
      page.getByRole("menuitem", { name: "Percent of total" }),
    ).toBeDisabled();
  });
});
