import { expect, Page, test } from "@playwright/test";
import {
  TestEntityType,
  updateCodeEditor,
  wrapRetryAssertion,
} from "../utils/commonHelpers";
import {
  assertLeaderboards,
  clickOnFilter,
  createDashboardFromModel,
  createDashboardFromSource,
  interactWithComparisonMenu,
  interactWithTimeRangeMenu,
  metricsViewRequestFilterMatcher,
  RequestMatcher,
  waitForComparisonTopLists,
  waitForDashboard,
  waitForTimeSeries,
} from "../utils/dashboardHelpers";
import {
  assertAdBidsDashboard,
  createAdBidsModel,
} from "../utils/dataSpecifcHelpers";
import { createOrReplaceSource } from "../utils/sourceHelpers";
import { startRuntimeForEachTest } from "../utils/startRuntimeForEachTest";
import { waitForEntity } from "../utils/waitHelpers";

test.describe("dashboard", () => {
  startRuntimeForEachTest();

  test("Autogenerate dashboard from source", async ({ page }) => {
    await page.goto("/");

    await createOrReplaceSource(page, "AdBids.csv", "AdBids");
    await createDashboardFromSource(page, "AdBids");
    await waitForEntity(
      page,
      TestEntityType.Dashboard,
      "AdBids_dashboard",
      true,
    );
    await assertAdBidsDashboard(page);
  });

  test("Autogenerate dashboard from model", async ({ page }) => {
    await page.goto("/");

    await createAdBidsModel(page);
    await Promise.all([
      waitForEntity(
        page,
        TestEntityType.Dashboard,
        "AdBids_model_dashboard",
        true,
      ),
      waitForTimeSeries(page, "AdBids_model_dashboard"),
      waitForComparisonTopLists(page, "AdBids_model_dashboard", ["domain"]),
      createDashboardFromModel(page, "AdBids_model"),
    ]);
    await assertAdBidsDashboard(page);

    // metrics view filter matcher to select just publisher=Facebook since we click on it
    const domainFilterMatcher: RequestMatcher = (response) =>
      metricsViewRequestFilterMatcher(
        response,
        [{ label: "publisher", values: ["Facebook"] }],
        [],
      );
    await Promise.all([
      waitForTimeSeries(page, "AdBids_model_dashboard", domainFilterMatcher),
      waitForComparisonTopLists(
        page,
        "AdBids_model_dashboard",
        ["domain"],
        domainFilterMatcher,
      ),
      // click on publisher=Facebook leaderboard value
      clickOnFilter(page, "Publisher", "Facebook"),
    ]);
    await wrapRetryAssertion(() =>
      assertLeaderboards(page, [
        {
          label: "Publisher",
          values: ["null", "Facebook", "Google", "Yahoo", "Microsoft"],
        },
        {
          label: "Domain",
          values: ["facebook.com", "instagram.com"],
        },
      ]),
    );
  });

  test("Dashboard runthrough", async ({ page }) => {
    // Enable to get logs in CI
    // page.on("console", async (msg) => {
    //   console.log(msg.text());
    // });
    // page.on("pageerror", (exception) => {
    //   console.log(
    //     `Uncaught exception: "${exception.message}"\n${exception.stack}`
    //   );
    // });

    test.setTimeout(60000);
    await page.goto("/");
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

    // Check the total records are 100k
    await expect(page.getByText("Total records 100.0k")).toBeVisible();

    // Check the row viewer accordion is visible
    await expect(page.getByText("Model Data 100k of 100k rows")).toBeVisible();

    // Change the metric trend granularity
    await page.getByRole("button", { name: "Metric trends by day" }).click();
    await page.getByRole("menuitem", { name: "day" }).click();

    // Change the time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    // Change time zone to UTC
    await page.getByLabel("Timezone selector").click();
    await page
      .getByRole("menuitem", { name: "UTC GMT +00:00 Etc/UTC" })
      .click();
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "UTC GMT +00:00 Etc/UTC" }),
    ).not.toBeVisible();

    await interactWithComparisonMenu(page, "No comparison", (l) =>
      l.getByRole("menuitem", { name: "Time" }).click(),
    );

    // Check that the total records are 272 and have comparisons
    await expect(page.getByText("272 -23 -7%")).toBeVisible();

    // Check the row viewer accordion is updated
    await expect(page.getByText("Model Data 272 of 100k rows")).toBeVisible();

    // Check row viewer is collapsed by looking for the cell value "7029", which should be in the table
    await expect(page.getByRole("button", { name: "7029" })).not.toBeVisible();

    // Expand row viewer and check data is there
    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    await expect(page.getByRole("button", { name: "7029" })).toBeVisible();

    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    // Check row viewer is collapsed
    await expect(page.getByRole("button", { name: "7029" })).not.toBeVisible();

    // Download the data as CSV
    // Start waiting for download before clicking. Note no await.
    const downloadCSVPromise = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export model data" }).click();
    await page.getByText("Export as CSV").click();
    const downloadCSV = await downloadCSVPromise;
    await downloadCSV.saveAs("temp/" + downloadCSV.suggestedFilename());
    const csvRegex = /^AdBids_model_filtered_.*\.csv$/;
    expect(csvRegex.test(downloadCSV.suggestedFilename())).toBe(true);

    // Download the data as XLSX
    // Start waiting for download before clicking. Note no await.
    const downloadXLSXPromise = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export model data" }).click();
    await page.getByText("Export as XLSX").click();
    const downloadXLSX = await downloadXLSXPromise;
    await downloadXLSX.saveAs("temp/" + downloadXLSX.suggestedFilename());
    const xlsxRegex = /^AdBids_model_filtered_.*\.xlsx$/;
    expect(xlsxRegex.test(downloadXLSX.suggestedFilename())).toBe(true);

    // Download the data as Parquet
    // Start waiting for download before clicking. Note no await.
    const downloadParquetPromise = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export model data" }).click();
    await page.getByText("Export as Parquet").click();
    const downloadParquet = await downloadParquetPromise;
    await downloadParquet.saveAs("temp/" + downloadParquet.suggestedFilename());

    const parquetRegex = /^AdBids_model_filtered_.*\.parquet$/;
    expect(parquetRegex.test(downloadParquet.suggestedFilename())).toBe(true);

    // Turn off comparison
    await interactWithComparisonMenu(page, "Comparing by Time", (l) =>
      l.getByRole("menuitem", { name: "No comparison" }).click(),
    );

    // Check number
    await expect(page.getByText("272", { exact: true })).toBeVisible();

    // Add comparison back
    await interactWithComparisonMenu(page, "No comparison", (l) =>
      l.getByRole("menuitem", { name: "Time" }).click(),
    );

    /*
      There is a bug where if you programmatically click the Time Range Selector button right after clicking the "Previous Period" menu item,
      the comparison menu closes, the time range menu opens, and then the comparison menu opens again. You can reproduce with a script like this in console
      after opening up comparison menu when "no comparison" is selected:
      (() => {
        document.evaluate("//button[contains(., 'previous period')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext().click();
        document.querySelector('[aria-label="Select time range"]').click();
      })()

      For now, we will wait for the menu to disappear before clicking the next menu
     */
    await expect(page.getByLabel("Comparison selector")).not.toBeVisible();

    // Switch to a custom time range
    await interactWithTimeRangeMenu(page, async () => {
      const timeRangeMenu = page.getByRole("menu", {
        name: "Time range selector",
      });

      await timeRangeMenu
        .getByRole("menuitem", { name: "Custom range" })
        .click();
      await timeRangeMenu.getByLabel("Start date").fill("2022-02-01");
      await timeRangeMenu.getByLabel("Start date").blur();
      await timeRangeMenu.getByRole("button", { name: "Apply" }).click();
    });

    // Check number
    await expect(page.getByText("Total records 65.1k")).toBeVisible();

    // Flip back to All Time
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    // Check number
    await expect(
      page.getByText("Total records 100.0k", { exact: true }),
    ).toBeVisible();

    // Filter to Facebook via leaderboard
    await page.getByRole("button", { name: "Facebook 19.3k" }).click();

    await page.waitForSelector("text=Publisher Facebook");

    // Change filter to excluded
    await page.getByText("Publisher Facebook").click();
    await page.getByRole("button", { name: "Exclude" }).click();
    await page.getByText("Exclude Publisher Facebook").click();

    // Check number
    await expect(
      page.getByText("Total records 80.7k", { exact: true }),
    ).toBeVisible();

    // Clear the filter from filter bar
    await page.getByLabel("View filter").getByLabel("Remove").click();

    // Apply a different filter
    await page.getByRole("button", { name: "google.com 15.1k" }).click();

    // Check number
    await expect(
      page.getByText("Total records 15.1k", { exact: true }),
    ).toBeVisible();

    // Clear all filters button
    await page.getByRole("button", { name: "Clear filters" }).click();

    // Check number
    await expect(
      page.getByText("Total records 100.0k", { exact: true }),
    ).toBeVisible();

    // Check no filters label
    await expect(
      page.getByText("No filters selected", { exact: true }),
    ).toBeVisible();

    // TODO
    //    Change time range to last 6 hours
    //    Check that the data is updated for last 6 hours
    //    Change time range back to all time

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Get the dashboard name field and change it

    const changeDisplayNameDoc = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    title: "AdBids_model_dashboard_rename"
    model: "AdBids_model"
    default_time_range: ""
    smallest_time_grain: ""
    measures:
      - label: Total records
        expression: count(*)
        name: total_records
        description: Total number of records present
        format_preset: humanize
    dimensions:
      - name: publisher
        label: Publisher
        column: publisher
        description: ""
      - name: domain
        label: Domain
        column: domain
        description: ""

        `;
    await updateCodeEditor(page, changeDisplayNameDoc);
    await waitForDashboard(page);

    // Remove timestamp column
    // await page.getByLabel("Remove timestamp column").click();

    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    // Assert that name changed
    await expect(page.getByText("AdBids_model_dashboard_rename")).toBeVisible();

    // Assert that no time dimension specified
    await expect(page.getByText("No time dimension specified")).toBeVisible();

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Add timestamp column back

    const addBackTimestampColumnDoc = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    title: "AdBids_model_dashboard_rename"
    model: "AdBids_model"
    default_time_range: ""
    smallest_time_grain: "week"
    timeseries: "timestamp"
    measures:
      - label: Total records
        expression: count(*)
        name: total_records
        description: Total number of records present
        format_preset: humanize
    dimensions:
      - name: publisher
        label: Publisher
        column: publisher
        description: ""
      - name: domain
        label: Domain
        column: domain
        description: ""

        `;
    await updateCodeEditor(page, addBackTimestampColumnDoc);
    await waitForDashboard(page);

    // Go to dashboard
    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    // Assert that time dimension is now week
    await expect(
      page.getByRole("button", { name: "Metric trends by week" }),
    ).toBeVisible();

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    const deleteOnlyMeasureDoc = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    title: "AdBids_model_dashboard_rename"
    model: "AdBids_model"
    default_time_range: ""
    smallest_time_grain: "week"
    timeseries: "timestamp"
    measures: []
    dimensions:
      - name: publisher
        label: Publisher
        column: publisher
        description: ""
      - name: domain
        label: Domain
        column: domain
        description: ""

        `;
    await updateCodeEditor(page, deleteOnlyMeasureDoc);
    // Check warning message appears, Go to Dashboard is disabled
    await expect(
      page.getByText("must define at least one measure"),
    ).toBeVisible();

    await expect(
      page.getByRole("button", { name: "Go to dashboard" }),
    ).toBeDisabled();

    // Add back the total rows measure for
    const docWithIncompleteMeasure = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    title: "AdBids_model_dashboard_rename"
    model: "AdBids_model"
    default_time_range: ""
    smallest_time_grain: "week"
    timeseries: "timestamp"
    measures:
      - label: Avg Bid Price
    dimensions:
      - name: publisher
        label: Publisher
        column: publisher
        description: ""
      - name: domain
        label: Domain
        column: domain
        description: ""
        
        `;

    await updateCodeEditor(page, docWithIncompleteMeasure);
    await expect(
      page.getByRole("button", { name: "Go to dashboard" }),
    ).toBeDisabled();

    const docWithCompleteMeasure = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

title: "AdBids_model_dashboard_rename"
model: "AdBids_model"
default_time_range: ""
smallest_time_grain: "week"
timeseries: "timestamp"
measures:
  - label: Total rows
    expression: count(*)
    name: total_rows
    description: Total number of records present
  - label: Avg Bid Price
    expression: avg(bid_price)
    name: avg_bid_price
    format_preset: currency_usd
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

    await updateCodeEditor(page, docWithCompleteMeasure);
    await expect(
      page.getByRole("button", { name: "Go to dashboard" }),
    ).toBeEnabled();

    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Check Avg Bid Price
    await expect(page.getByText("Avg Bid Price $3.01")).toBeVisible();

    // Change the leaderboard metric
    await page.getByRole("button", { name: "Total rows", exact: true }).click();
    await page.getByRole("menuitem", { name: "Avg Bid Price" }).click();

    // Check domain and sample value in leaderboard
    await expect(page.getByText("Domain Name")).toBeVisible();
    await expect(page.getByText("facebook.com $3.13")).toBeVisible();

    // Open the Publisher details table
    await page
      .getByLabel("Open dimension details")
      .filter({ hasText: "Publisher" })
      .click();

    // Check that table is shown
    await expect(
      page.getByRole("table", { name: "Dimension table" }),
    ).toBeVisible();

    // Check for a table value
    // Can do better table checking in the future when table is refactored to use proper row setup
    // For now, just check the dimensions
    await expect(
      page.locator("button").filter({ hasText: /^Microsoft$/ }),
    ).toBeVisible();

    // TODO when table is better formatted
    //    Change sort direction
    //    Check new sort direction worked in first table row
    //    Change sort column and check

    // Click a table value to filter
    await page
      .locator("button")
      .filter({ hasText: /^Microsoft$/ })
      .click();

    // Check that filter was applied
    await expect(
      page.getByLabel("View filter").getByText("Publisher Microsoft"),
    ).toBeVisible();

    // go back to the leaderboards.
    await page.getByText("All dimensions").click();
    // clear all filters
    await page.getByText("Clear filters").click();

    // run through TDD table view
    await page.getByText("Total rows 100.0k").click();

    await expect(
      page.getByText("No comparison dimension selected"),
    ).toBeVisible();

    await page.getByRole("button", { name: "No comparison" }).nth(1).click();
    await page.getByRole("menuitem", { name: "Domain Name" }).click();

    await page.getByText("google.com", { exact: true }).click();
    await page.getByText("instagram.com").click();
    await page.getByText("news.google.com").click();

    await expect(page.getByText(" Total rows 41.1k")).toBeVisible();

    await page.getByRole("cell", { name: "Total rows" }).locator("div").click();

    await page.getByRole("button", { name: "Total rows", exact: true }).click();
    await page.getByRole("menuitem", { name: "Avg Bid Price" }).click();

    await expect(page.getByText(" Avg Bid Price $2.93")).toBeVisible();

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    });

    await page.getByRole("button", { name: "Domain name" }).nth(1).click();
    await page.getByRole("menuitem", { name: "Time" }).click();

    await expect(page.getByText("-52.2%")).toBeVisible();

    await page.getByRole("button", { name: "Edit metrics" }).click();
    /** walk through empty metrics def  */
    await runThroughEmptyMetricsFlows(page);

    await runThroughLeaderboardContextColumnFlows(page);

    // go back to the dashboard

    // TODO
    //    Check that details table can exclude
    //    Add search criteria
    //    Check that table got search
    //    Clear search
    //    Change the sort column to total rows
    //    Go back to leaderboard
    //    Check that selected metric is total rows
    //    Change the leaderboard metric to avg bid price
    //    await page.getByRole("button", { name: "Total records" }).click();
  });
});

async function runThroughLeaderboardContextColumnFlows(page: Page) {
  // NOTE: this flow picks up from the end of runThroughEmptyMetricsFlows,
  // at which point we are in the metrics editor

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
  await expect(page.getByText("Facebook $229.26 28%")).toBeVisible();

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

  // Go back to metrics editor
  await page.getByRole("button", { name: "Edit metrics" }).click();
}

async function runThroughEmptyMetricsFlows(page: Page) {
  await updateCodeEditor(page, "");

  // the inspector should be empty.
  await expect(await page.getByText("Let's get started.")).toBeVisible();

  // skeleton should result in an empty skeleton YAML file
  await page.getByText("start with a skeleton").click();

  // check to see that the placeholder is gone by looking for the button
  // that was once there.
  await wrapRetryAssertion(async () => {
    await expect(await page.getByText("start with a skeleton")).toBeHidden();
  });

  // the  button should be disabled.
  await expect(
    await page.getByRole("button", { name: "Go to dashboard" }),
  ).toBeDisabled();

  // the inspector should be empty.
  await expect(await page.getByText("Model not defined.")).toBeVisible();

  // now let's scaffold things in
  await updateCodeEditor(page, "");

  await wrapRetryAssertion(async () => {
    await expect(
      await page.getByText("metrics configuration from an existing model"),
    ).toBeVisible();
  });

  // select the first menu item.
  await page.getByText("metrics configuration from an existing model").click();
  await page.getByRole("menuitem").getByText("AdBids_model").click();
  // Wait for menu to close
  await expect(
    page.getByRole("menuitem", { name: "Last 6 Hours" }),
  ).not.toBeVisible();

  // let's check the inspector.
  await expect(await page.getByText("Model summary")).toBeVisible();
  await expect(await page.getByText("Model columns")).toBeVisible();

  // go to teh dashboard and make sure the metrics and dimensions are there.

  await page.getByRole("button", { name: "Go to dashboard" }).click();

  // check to see metrics make sense.
  await expect(await page.getByText("Total Records 100.0k")).toBeVisible();

  // double-check that leaderboards make sense.
  await expect(
    await page.getByRole("button", { name: "google.com 15.1k" }),
  ).toBeVisible();

  // go back to the metrics page.
  await page.getByRole("button", { name: "Edit metrics" }).click();
}
