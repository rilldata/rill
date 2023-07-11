import { describe, it } from "@jest/globals";
import { expect as playwrightExpect } from "@playwright/test";
import {
  RequestMatcher,
  assertLeaderboards,
  clickOnFilter,
  createDashboardFromModel,
  createDashboardFromSource,
  metricsViewRequestFilterMatcher,
  updateMetricsInput,
  waitForTimeSeries,
  waitForTopLists,
} from "./utils/dashboardHelpers";
import {
  assertAdBidsDashboard,
  createAdBidsModel,
} from "./utils/dataSpecifcHelpers";
import { TestEntityType, wrapRetryAssertion } from "./utils/helpers";
import { useRegisteredServer } from "./utils/serverConfigs";
import { createOrReplaceSource } from "./utils/sourceHelpers";
import { waitForEntity } from "./utils/waitHelpers";

describe("dashboards", () => {
  const testBrowser = useRegisteredServer("dashboards");

  it("Autogenerate dashboard from source", async () => {
    const { page } = testBrowser;

    await createOrReplaceSource(page, "AdBids.csv", "AdBids");
    await createDashboardFromSource(page, "AdBids");
    await waitForEntity(
      page,
      TestEntityType.Dashboard,
      "AdBids_dashboard",
      true
    );
    await assertAdBidsDashboard(page);
  });

  it("Autogenerate dashboard from model", async () => {
    const { page } = testBrowser;

    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");
    await Promise.all([
      waitForEntity(
        page,
        TestEntityType.Dashboard,
        "AdBids_model_dashboard",
        true
      ),
      waitForTimeSeries(page, "AdBids_model_dashboard"),
      waitForTopLists(page, "AdBids_model_dashboard", ["domain"]),
    ]);
    await assertAdBidsDashboard(page);

    // metrics view filter matcher to select just publisher=Facebook since we click on it
    const domainFilterMatcher: RequestMatcher = (response) =>
      metricsViewRequestFilterMatcher(
        response,
        [{ label: "publisher", values: ["Facebook"] }],
        []
      );
    await Promise.all([
      waitForTimeSeries(page, "AdBids_model_dashboard", domainFilterMatcher),
      waitForTopLists(
        page,
        "AdBids_model_dashboard",
        ["domain"],
        domainFilterMatcher
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
      ])
    );
  });

  it("should run through the dashboard", async () => {
    const { page } = testBrowser;
    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");
    await waitForEntity(
      page,
      TestEntityType.Dashboard,
      "AdBids_model_dashboard",
      true
    );

    // Check the total records are 100k
    await playwrightExpect(
      page.getByText("Total records 100.0k")
    ).toBeVisible();

    // Check the row viewer accordion is visible
    await playwrightExpect(
      page.getByText("Model Data 100k of 100k rows")
    ).toBeVisible();

    // Change the metric trend granularity
    await page.getByRole("button", { name: "Metric trends by day" }).click();
    await page.getByRole("menuitem", { name: "day" }).click();

    // Change the time range
    await page.getByLabel("Select time range").click();
    await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    // Check that the total records are 272 and have comparisons
    await playwrightExpect(page.getByText("272 -23 -7%")).toBeVisible();

    // Check the row viewer accordion is updated
    await playwrightExpect(
      page.getByText("Model Data 272 of 100k rows")
    ).toBeVisible();

    // Check row viewer is collapsed by looking for the cell value "7029", which should be in the table
    await playwrightExpect(
      page.getByRole("button", { name: "7029" })
    ).not.toBeVisible();

    // Expand row viewer and check data is there
    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    await playwrightExpect(
      page.getByRole("button", { name: "7029" })
    ).toBeVisible();

    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    // Check row viewer is collapsed
    await playwrightExpect(
      page.getByRole("button", { name: "7029" })
    ).not.toBeVisible();

    // Download the data as CSV
    // Start waiting for download before clicking. Note no await.
    const downloadCSVPromise = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export model data" }).click();
    await page.getByText("Export as CSV").click();
    const downloadCSV = await downloadCSVPromise;
    await downloadCSV.path();
    const csvRegex = /^AdBids_model_filtered_.*\.csv$/;
    playwrightExpect(csvRegex.test(downloadCSV.suggestedFilename())).toBe(true);

    // Download the data as XLSX
    // Start waiting for download before clicking. Note no await.
    const downloadXLSXPromise = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export model data" }).click();
    await page.getByText("Export as XLSX").click();
    const downloadXLSX = await downloadXLSXPromise;
    await downloadXLSX.path();
    const xlsxRegex = /^AdBids_model_filtered_.*\.xlsx$/;
    playwrightExpect(xlsxRegex.test(downloadXLSX.suggestedFilename())).toBe(
      true
    );

    // Turn off comparison
    await page
      .getByRole("button", { name: "Comparing to last period" })
      .click();
    await page
      .getByLabel("Time comparison selector")
      .getByRole("menuitem", { name: "no comparison" })
      .click();

    // Check number
    await playwrightExpect(
      page.getByText("272", { exact: true })
    ).toBeVisible();

    // Add comparison back
    await page.getByRole("button", { name: "no comparison" }).click();
    await page
      .getByLabel("Time comparison selector")
      .getByRole("menuitem", { name: "last period" })
      .click();

    /*
      There is a bug where if you programmatically click the Time Range Selector button right after clicking the "Last Period" menu item,
      the comparison menu closes, the time range menu opens, and then the comparison menu opens again. You can reproduce with a script like this in console
      after opening up comparison menu when "no comparison" is selected:
      (() => {
        document.evaluate("//button[contains(., 'last period')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext().click();
        document.querySelector('[aria-label="Select time range"]').click();
      })()

      For now, we will wait for the menu to disappear before clicking the next menu
     */
    await playwrightExpect(
      page.getByLabel("Time comparison selector")
    ).not.toBeVisible();

    // Switch to a custom time range
    await page.getByLabel("Select time range").click();

    const timeRangeMenu = page.getByRole("menu", {
      name: "Time range selector",
    });

    await timeRangeMenu.getByRole("menuitem", { name: "Custom range" }).click();
    await timeRangeMenu.getByLabel("Start date").fill("2022-02-01");
    await timeRangeMenu.getByLabel("Start date").blur();
    await timeRangeMenu.getByRole("button", { name: "Apply" }).click();

    // Check number
    await playwrightExpect(page.getByText("Total records 65.1k")).toBeVisible();

    // Flip back to All Time
    await page.getByLabel("Select time range").click();
    await page.getByRole("menuitem", { name: "All Time" }).click();

    // Check number
    await playwrightExpect(
      page.getByText("Total records 100.0k", { exact: true })
    ).toBeVisible();

    // Filter to Facebook via leaderboard
    await page.getByRole("button", { name: "Facebook 19.3k" }).click();

    // Change filter to excluded
    await page.getByText("Publisher Facebook").click();
    await page.getByRole("button", { name: "Exclude" }).click();
    await page.getByText("Exclude Publisher Facebook").click();

    // Check number
    await playwrightExpect(
      page.getByText("Total records 80.7k", { exact: true })
    ).toBeVisible();

    // Clear the filter from filter bar
    await page.getByLabel("View filter").getByLabel("Remove").click();

    // Apply a different filter
    await page.getByRole("button", { name: "google.com 15.1k" }).click();

    // Check number
    await playwrightExpect(
      page.getByText("Total records 15.1k", { exact: true })
    ).toBeVisible();

    // Clear all filters button
    await page.getByRole("button", { name: "Clear filters" }).click();

    // Check number
    await playwrightExpect(
      page.getByText("Total records 100.0k", { exact: true })
    ).toBeVisible();

    // Check no filters label
    await playwrightExpect(
      page.getByText("No filters selected", { exact: true })
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
    await updateMetricsInput(page, changeDisplayNameDoc);

    // Remove timestamp column
    // await page.getByLabel("Remove timestamp column").click();

    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    // Assert that name changed
    await playwrightExpect(
      page.getByText("AdBids_model_dashboard_rename")
    ).toBeVisible();

    // Assert that no time dimension specified
    await playwrightExpect(
      page.getByText("No time dimension specified")
    ).toBeVisible();

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
    await updateMetricsInput(page, addBackTimestampColumnDoc);

    // Go to dashboard
    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    // Assert that time dimension is now week
    await playwrightExpect(
      page.getByRole("button", { name: "Metric trends by week" })
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
    await updateMetricsInput(page, deleteOnlyMeasureDoc);
    // Check warning message appears, Go to Dashboard is disabled
    await playwrightExpect(
      page.getByText("at least one measure should be present")
    ).toBeVisible();

    await playwrightExpect(
      page.getByRole("button", { name: "Go to dashboard" })
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

    await updateMetricsInput(page, docWithIncompleteMeasure);
    await playwrightExpect(
      page.getByRole("button", { name: "Go to dashboard" })
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

    await updateMetricsInput(page, docWithCompleteMeasure);
    await playwrightExpect(
      page.getByRole("button", { name: "Go to dashboard" })
    ).toBeEnabled();

    // Go to dashboard
    await page.getByRole("button", { name: "Go to dashboard" }).click();

    // Check Avg Bid Price
    await playwrightExpect(page.getByText("Avg Bid Price $3.01")).toBeVisible();

    // Change the leaderboard metric
    await page.getByRole("button", { name: "Total rows" }).click();
    await page.getByRole("menuitem", { name: "Avg Bid Price" }).click();

    // Check domain and sample value in leaderboard
    await playwrightExpect(page.getByText("Domain Name")).toBeVisible();
    await playwrightExpect(page.getByText("facebook.com $3.13")).toBeVisible();

    // Open the Publisher details table
    await page
      .getByLabel("Open dimension details")
      .filter({ hasText: "Publisher" })
      .click();

    // Check that table is shown
    await playwrightExpect(
      page.getByRole("table", { name: "Dimension table" })
    ).toBeVisible();

    // Check for a table value
    // Can do better table checking in the future when table is refactored to use proper row setup
    // For now, just check the dimensions
    await playwrightExpect(
      page
        .getByRole("button", { name: "Filter dimension value" })
        .filter({ hasText: "Microsoft" })
    ).toBeVisible();

    // TODO when table is better formatted
    //    Change sort direction
    //    Check new sort direction worked in first table row
    //    Change sort column and check

    // Click a table value to filter
    await page
      .getByRole("button", { name: "Filter dimension value" })
      .filter({ hasText: "Microsoft" })
      .click();

    // Check that filter was applied
    await playwrightExpect(
      page.getByLabel("View filter").getByText("Publisher Microsoft")
    ).toBeVisible();

    // go back to the leaderboards.
    await page.getByText("All dimensions").click();
    // clear all filters
    await page.getByText("Clear filters").click();

    await page.getByRole("button", { name: "Edit metrics" }).click();

    /** walk through empty metrics def  */
    await runThroughEmptyMetricsFlows(page);

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

/**
 * This flow assumes you start on a metrics page, and ends on the metrics page.
 * It will (1) delete any content
 * (2) add a skeleton YAML file
 * (3) scaffold in a metrics def from a model
 * (4) verify that the scaffolding works by looking at the dashboard.
 */
async function runThroughEmptyMetricsFlows(page) {
  await updateMetricsInput(page, "");

  // the inspector should be empty.
  await playwrightExpect(
    await page.getByText("Let's get started.")
  ).toBeVisible();

  // skeleton should result in an empty skeleton YAML file
  await page.getByText("start with a skeleton").click();

  // check to see that the placeholder is gone by looking for the button
  // that was once there.
  await wrapRetryAssertion(async () => {
    await playwrightExpect(
      await page.getByText("start with a skeleton")
    ).toBeHidden();
  });

  // the  button should be disabled.
  await playwrightExpect(
    await page.getByRole("button", { name: "Go to dashboard" })
  ).toBeDisabled();

  // the inspector should be empty.
  await playwrightExpect(
    await page.getByText("Model not defined.")
  ).toBeVisible();

  // now let's scaffold things in
  await updateMetricsInput(page, "");

  await wrapRetryAssertion(async () => {
    await playwrightExpect(
      await page.getByText("metrics configuration from an existing model")
    ).toBeVisible();
  });

  // select the first menu item.
  await page.getByText("metrics configuration from an existing model").click();
  await page.getByRole("menuitem").getByText("AdBids_model").click();

  // let's check the inspector.
  await playwrightExpect(await page.getByText("Model summary")).toBeVisible();
  await playwrightExpect(await page.getByText("Model columns")).toBeVisible();

  // go to teh dashboard and make sure the metrics and dimensions are there.

  await page.getByRole("button", { name: "Go to dashboard" }).click();

  // check to see metrics make sense.
  await playwrightExpect(
    await page.getByText("Total Records 100.0k")
  ).toBeVisible();

  // double-check that leaderboards make sense.
  await playwrightExpect(
    await page.getByRole("button", { name: "google.com 15.1k" })
  ).toBeVisible();

  // go back to the metrics page.
  await page.getByRole("button", { name: "Edit metrics" }).click();
}
