import { describe, it } from "@jest/globals";
import { expect as playwrightExpect } from "@playwright/test";
import {
  assertAdBidsDashboard,
  createAdBidsModel,
} from "./utils/dataSpecifcHelpers";
import {
  assertLeaderboards,
  clickOnFilter,
  createDashboardFromModel,
  createDashboardFromSource,
  metricsViewRequestFilterMatcher,
  RequestMatcher,
  waitForTimeSeries,
  waitForTopLists,
} from "./utils/dashboardHelpers";
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

    // Change the metric trend granularity
    await page.getByRole("button", { name: "Metric trends by day" }).click();
    await page.getByRole("menuitem", { name: "day" }).click();

    // Change the time range
    await page.getByLabel("Select time range").click();
    await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();

    // Check that the total records are 275 and have comparisons
    await playwrightExpect(page.getByText("275 -12 -4%")).toBeVisible();

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
      page.getByText("275", { exact: true })
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
    await playwrightExpect(page.getByText("Total records 64.0k")).toBeVisible();

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
    await page.getByLabel("Display name").fill("AdBids_model_dashboard_rename");
    await page.getByLabel("Display name").blur();

    // Remove timestamp column
    await page.getByLabel("Remove timestamp column").click();

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
    await page.getByRole("button", { name: "Select a time column" }).click();
    await page.getByRole("menuitem", { name: "timestamp" }).click();

    // Change smallest grain
    await page
      .getByRole("button", { name: "Change smallest time grain" })
      .click();
    await page.getByRole("menuitem", { name: "week" }).click();

    // Go to dashboard
    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    // Assert that time dimension is now week
    await playwrightExpect(
      page.getByRole("button", { name: "Metric trends by week" })
    ).toBeVisible();

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Delete the only measure
    const measuresTable = await page.getByRole("table", { name: "Measures" });
    const firstRow = await measuresTable.getByRole("row").nth(1);
    await firstRow.hover();
    await firstRow.getByRole("button", { name: "More" }).click();
    await page.getByRole("menuitem", { name: "Delete row" }).click();

    // Check warning message appears, Go to Dashboard is disabled
    await playwrightExpect(
      page.getByText("at least one measure should be present")
    ).toBeVisible();

    await playwrightExpect(
      page.getByRole("button", { name: "Go to dashboard" })
    ).toBeDisabled();

    // Add total rows measure back
    await page.getByRole("button", { name: "Add measure" }).click();

    await measuresTable
      .getByRole("row")
      .nth(1)
      .getByRole("textbox", { name: "Measure label" })
      .fill("Total rows");

    await measuresTable
      .getByRole("row")
      .nth(1)
      .getByRole("textbox", { name: "Measure expression" })
      .fill("count(*)");

    // Check Quick Metrics button visible
    await playwrightExpect(page.getByText("Quick Metrics")).toBeVisible();

    // Add Avg Bid Price metric, first without a definition
    await page.getByRole("button", { name: "Add measure" }).click();
    await measuresTable
      .getByRole("row")
      .nth(2)
      .getByRole("textbox", { name: "Measure label" })
      .fill("Avg Bid Price");

    // Click Go to Dashboard and get redirected back to metrics config
    await page.getByRole("button", { name: "Go to dashboard" }).click();
    await playwrightExpect(measuresTable).toBeVisible();

    // Add a definition and pick a format
    await measuresTable
      .getByRole("row")
      .nth(2)
      .getByRole("textbox", { name: "Measure expression" })
      .fill("avg(bid_price)");

    await measuresTable
      .getByRole("row")
      .nth(2)
      .getByLabel("Measure number formatting")
      .selectOption("Currency (USD)");

    // Remove all dimensions
    const dimensionsTable = await page.getByRole("table", {
      name: "Dimensions",
    });
    await dimensionsTable.getByRole("row").nth(1).hover();
    await dimensionsTable
      .getByRole("row")
      .nth(1)
      .getByRole("button", { name: "More" })
      .click();
    await page.getByRole("menuitem", { name: "Delete row" }).click();
    await dimensionsTable.getByRole("row").nth(1).hover();
    await dimensionsTable
      .getByRole("row")
      .nth(1)
      .getByRole("button", { name: "More" })
      .click();
    await page.getByRole("menuitem", { name: "Delete row" }).click();

    // Check that Go to Dashboard is disabled
    await playwrightExpect(
      page.getByRole("button", { name: "Go to dashboard" })
    ).toBeDisabled();

    // Add Published, Domain back to dashboard
    await page.getByRole("button", { name: "Add dimension" }).click();
    await dimensionsTable
      .getByRole("row")
      .nth(1)
      .getByRole("textbox", { name: "Dimension label" })
      .fill("Publisher");
    await dimensionsTable
      .getByRole("row")
      .nth(1)
      .getByLabel("Dimension column")
      .selectOption("publisher");
    await page.getByRole("button", { name: "Add dimension" }).click();
    await dimensionsTable
      .getByRole("row")
      .nth(2)
      .getByRole("textbox", { name: "Dimension label" })
      .fill("Domain Name");
    await dimensionsTable
      .getByRole("row")
      .nth(2)
      .getByLabel("Dimension column")
      .selectOption("domain");

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
