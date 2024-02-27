import { expect } from "@playwright/test";
import { updateCodeEditor } from "../utils/commonHelpers";
import {
  createDashboardFromModel,
  interactWithComparisonMenu,
  interactWithTimeRangeMenu,
  waitForDashboard,
} from "../utils/dashboardHelpers";
import { createAdBidsModel } from "../utils/dataSpecifcHelpers";
import { test } from "../utils/test";

test.describe("leaderboard context column", () => {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(60000);

    // disable animations
    await page.addStyleTag({
      content: `
        *, *::before, *::after {
          animation-duration: 0s !important;
          transition-duration: 0s !important;
        }
      `,
    });
    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Close the navigation sidebar to give the code editor more space
    await page.locator(".contents > .fixed").click();
  });

  test("Leaderboard context column", async ({ page }) => {
    /*
     * SUBFLOW: setup state for the leaderboard context column tests
     */

    // reset metrics, and add a metric with `valid_percent_of_total: true`
    const metricsWithValidPercentOfTotal = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

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
    await updateCodeEditor(page, metricsWithValidPercentOfTotal);
    await waitForDashboard(page);

    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // make sure "All time" is selected to clear any time comparison
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    /**
     * SUBFLOW:
     * check menu items are disabled/enabled as expected
     * when there is no time comparison
     */

    // click to open the context column menu
    await page.getByLabel("Select a context column").click();
    // Check "percent change" menuitem disabled since there is no time comparison
    await expect(
      page.getByRole("menuitem", { name: "Percent change" }),
    ).toBeDisabled();
    // Check "percent of total" item is disabled since `valid_percent_of_total` is not set for the measure "total rows"
    await expect(
      page.getByRole("menuitem", { name: "Percent of total" }),
    ).toBeDisabled();
    // click to close the context column menu
    await page.getByLabel("Select a context column").click();

    /**
     * SUBFLOW: check correct behavior when a time comparison
     * is activated, but there is no valid_percent_of_total
     */

    // Select a time range, that supports comparisons
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });
    // enable comparisons which should automatically enable a time comparison (including context column)
    await interactWithComparisonMenu(page, "No comparison", (l) =>
      l.getByRole("menuitem", { name: "Time" }).click(),
    );

    // check that the "select a context column" button now reads "with Percent change"
    await expect(
      page.getByLabel("Select a context column"),
      // getByRole("button", { name: "with Percent change" })
    ).toContainText("Percent change");

    // This regex matches a line that:
    // - starts with "Facebook"
    // - has two white space separated sets of characters (the number and the percent change)
    // - ends with a percent sign literal
    // e.g. "Facebook 68.9k -12%".
    // This will detect both percent change and percent of total
    const comparisonColumnRegex = /Facebook\s*\S*\s*\S*%/;

    // Check that time comparison context column is visible with correct value now that there is a time comparison
    await expect(page.getByText(comparisonColumnRegex)).toBeVisible();

    // click to open the context column menu
    await page.getByLabel("Select a context column").click();
    // Check that the "percent change" menuitem is enabled
    await expect(
      page.getByRole("menuitem", { name: "Percent change" }),
    ).toBeEnabled();
    // check that the "percent of total" menuitem is still disabled
    await expect(
      page.getByRole("menuitem", { name: "Percent of total" }),
    ).toBeDisabled();
    // click to close the context column menu
    await page.getByLabel("Select a context column").click();

    /**
     * SUBFLOW: check correct behavior when
     * - a time comparison is activated,
     * - there is no valid_percent_of_total,
     * - and then the context column is turned off
     */

    // turn off the context column
    await page.getByLabel("Select a context column").click();
    await page.getByRole("menuitem", { name: "No context column" }).click();
    // Check that time comparison context column is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();
    await expect(page.getByText("Facebook 68")).toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - the context column is turned back on,
     * - there is no valid_percent_of_total,
     * - and then time comparison is turned off
     */

    // turn on the context column
    await page.getByLabel("Select a context column").click();
    await page.getByRole("menuitem", { name: "Percent change" }).click();
    await expect(page.getByText(comparisonColumnRegex)).toBeVisible();

    // click back to "All time" to clear the time comparison
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });
    await interactWithComparisonMenu(page, "Comparing by Time", (l) =>
      l.getByRole("menuitem", { name: "No Comparison" }).click(),
    );

    // Check that time comparison context column is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();
    await expect(page.getByText("Facebook 19.3k")).toBeVisible();
    // Check that the "percent change" menuitem is disabled
    await page.getByLabel("Select a context column").click();
    await expect(
      page.getByRole("menuitem", { name: "Percent change" }),
    ).toBeDisabled();

    /**
     * SUBFLOW: check correct behavior when
     * - switching to a measure with valid_percent_of_total
     * - but no time comparison enabled
     */

    // Switch to measure "total bid price"
    await page.getByRole("button", { name: "Total rows", exact: true }).click();
    await page.getByRole("menuitem", { name: "Total Bid Price" }).click();
    await page.getByRole("button", { name: "Total Bid Price" }).isVisible();

    // open the context column menu
    await page.getByLabel("Select a context column").click();
    // Check that the "0ercent of total" menuitem is enabled
    await expect(
      page.getByRole("menuitem", { name: "Percent of total" }),
    ).toBeDisabled();
    // Check that the "percent change" menuitem is disabled
    await expect(
      page.getByRole("menuitem", { name: "Percent change" }),
    ).toBeDisabled();
    // close the context column menu
    await page.getByLabel("Select a context column").click();

    // Check that the percent of total is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();

    /**
     * SUBFLOW: check correct behavior when
     * - measure with valid_percent_of_total
     * - no time comparison enabled
     * - percent of total context column is turned on
     */

    // open the context column menu
    await page.getByLabel("Select a context column").click();
    // click on "percent of total" menuitem
    await page.getByRole("menuitem", { name: "Percent of total" }).click();
    //close the context column menu
    await page.getByLabel("Select a context column").click();
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
    await interactWithComparisonMenu(page, "No Comparison", (l) =>
      l.getByRole("menuitem", { name: "Time" }).click(),
    );

    // open the context column menu
    await page.getByLabel("Select a context column").click();
    // click on "percent change" menuitem
    await page.getByRole("menuitem", { name: "Percent change" }).click();
    //close the context column menu
    await page.getByLabel("Select a context column").click();
    // check that the percent change is visible+correct
    await expect(page.getByText("Facebook $229.26 3%")).toBeVisible();

    // open the context column menu
    await page.getByLabel("Select a context column").click();
    // click on "percent of total" menuitem
    await page.getByRole("menuitem", { name: "Percent of total" }).click();
    //close the context column menu
    await page.getByLabel("Select a context column").click();
    // check that the percent of total is visible+correct
    await expect(page.getByText("Facebook $229.26 28%")).toBeVisible();

    /**
     * Go back to measure without valid_percent_of_total
     * while percent of total context column is enabled.
     * Make sure the context column is hidden,
     * and the menuitems have the correct enabled/disabled state.
     */

    // Switch to measure "total rows" (no valid_percent_of_total)
    await page
      .getByRole("button", { name: "Total Bid Price", exact: true })
      .click();
    await page.getByRole("menuitem", { name: "Total rows" }).click();
    // check that the context column is hidden
    await expect(page.getByText(comparisonColumnRegex)).not.toBeVisible();
    // open the context column menu
    await page.getByLabel("Select a context column").click();
    // check that the "percent of total" menuitem is disabled
    await expect(
      page.getByRole("menuitem", { name: "Percent of total" }),
    ).toBeDisabled();
  });
});
