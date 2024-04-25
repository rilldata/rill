import { expect } from "@playwright/test";
import { updateCodeEditor, wrapRetryAssertion } from "../utils/commonHelpers";
import {
  assertLeaderboards,
  createDashboardFromModel,
  createDashboardFromSource,
  interactWithComparisonMenu,
  interactWithTimeRangeMenu,
  metricsViewRequestFilterMatcher,
  updateAndWaitForDashboard,
  waitForComparisonTopLists,
  waitForTimeSeries,
  type RequestMatcher,
} from "../utils/dashboardHelpers";
import {
  assertAdBidsDashboard,
  createAdBidsModel,
} from "../utils/dataSpecifcHelpers";
import { createSource } from "../utils/sourceHelpers";
import { test } from "../utils/test";
import { waitForFileNavEntry } from "../utils/waitHelpers";

test.describe("dashboard", () => {
  test("Autogenerate dashboard from source", async ({ page }) => {
    await createSource(page, "AdBids.csv", "/sources/AdBids.yaml");
    await createDashboardFromSource(page, "/sources/AdBids.yaml");
    await waitForFileNavEntry(page, `/dashboards/AdBids_dashboard.yaml`, true);
    await page.getByRole("button", { name: "Preview" }).click();
    await assertAdBidsDashboard(page);
  });

  test("Autogenerate dashboard from model", async ({ page }) => {
    await createAdBidsModel(page);
    await Promise.all([
      waitForFileNavEntry(
        page,
        `/dashboards/AdBids_model_dashboard.yaml`,
        true,
      ),
      createDashboardFromModel(page, "/models/AdBids_model.sql"),
    ]);
    await Promise.all([
      waitForTimeSeries(page, "AdBids_model_dashboard"),
      waitForComparisonTopLists(page, "AdBids_model_dashboard", ["domain"]),
      page.getByRole("button", { name: "Preview" }).click(),
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
      page.getByRole("button", { name: "Facebook 19.3K" }).click(),
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
    await createDashboardFromModel(page, "/models/AdBids_model.sql");
    await page.getByRole("button", { name: "Preview" }).click();

    // Check the total records are 100k
    await expect(page.getByText("Total records 100.0k")).toBeVisible();

    // Check the row viewer accordion is visible
    await expect(page.getByText("Model Data 100k of 100k rows")).toBeVisible();

    // Change the metric trend granularity

    const timeGrainSelector = page.getByRole("button", {
      name: "Select a time grain",
    });
    await timeGrainSelector.click();
    await page.getByRole("menuitem", { name: "day" }).click();

    // Change the time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    // Change time zone to UTC
    await page.getByLabel("Timezone selector").click();
    await page.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }).click();
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }),
    ).not.toBeVisible();

    await interactWithComparisonMenu(page, "No comparison", (l) =>
      l.getByRole("menuitem", { name: "Time" }).click(),
    );

    // Check that the total records are 272 and have comparisons
    await expect(page.getByText("272 -23 -8%")).toBeVisible();

    // Check the row viewer accordion is updated
    await expect(page.getByText("Model Data 272 of 100k rows")).toBeVisible();

    // Check row viewer is collapsed by looking for the cell value "7029", which should be in the table
    await expect(page.getByRole("cell", { name: "7029" })).not.toBeVisible();

    // Expand row viewer and check data is there
    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    await expect(page.getByRole("cell", { name: "7029" })).toBeVisible();

    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    // Check row viewer is collapsed
    await expect(page.getByRole("cell", { name: "7029" })).not.toBeVisible();

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

    kind: metrics_view
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
    await updateAndWaitForDashboard(page, changeDisplayNameDoc);

    // Remove timestamp column
    // await page.getByLabel("Remove timestamp column").click();

    await page.getByRole("button", { name: "Preview" }).click();

    // Assert that name changed
    await expect(
      page.getByRole("link", { name: "AdBids_model_dashboard_rename" }),
    ).toBeVisible();

    // Assert that no time dimension specified
    await expect(page.getByText("No time dimension specified")).toBeVisible();

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Add timestamp column back

    const addBackTimestampColumnDoc = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    kind: metrics_view
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
    await updateAndWaitForDashboard(page, addBackTimestampColumnDoc);

    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Assert that time dimension is now week
    await expect(timeGrainSelector).toHaveText("Metric trends by week");

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    const deleteOnlyMeasureDoc = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    kind: metrics_view
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
    // Check warning message appears, Preview is disabled
    await expect(
      page.getByText("must define at least one measure"),
    ).toBeVisible();

    await expect(page.getByRole("button", { name: "Preview" })).toBeDisabled();

    // Add back the total rows measure for
    const docWithIncompleteMeasure = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    kind: metrics_view
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
    await expect(page.getByRole("button", { name: "Preview" })).toBeDisabled();

    const docWithCompleteMeasure = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

kind: metrics_view
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
    await expect(page.getByRole("button", { name: "Preview" })).toBeEnabled();

    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Check Avg Bid Price
    await expect(page.getByText("Avg Bid Price $3.01")).toBeVisible();

    // Change the leaderboard metric
    await page
      .getByRole("button", { name: "Select a measure to filter by" })
      .click();
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
    await page.getByText("msn.com").click();

    await expect(page.getByText(" Total rows 43.7k")).toBeVisible();

    await page.getByRole("cell", { name: "Total rows" }).locator("div").click();

    await page.getByRole("button", { name: "Total rows", exact: true }).click();
    await page.getByRole("menuitem", { name: "Avg Bid Price" }).click();

    await expect(page.getByText(" Avg Bid Price $3.02")).toBeVisible();

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    });

    await page.getByRole("button", { name: "Domain name" }).nth(1).click();
    await page.getByRole("menuitem", { name: "Time" }).click();

    await expect(page.getByText("~0%")).toBeVisible();

    await page.getByRole("button", { name: "Edit Metrics" }).click();

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
