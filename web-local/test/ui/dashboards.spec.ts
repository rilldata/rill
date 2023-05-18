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
import { asyncWait } from "@rilldata/web-local/lib/util/waitUtils";

describe.only("dashboards", () => {
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

  it.only("should run through the dashboard", async () => {
    const { page } = testBrowser;
    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");
    // await Promise.all([
    //   waitForEntity(
    //     page,
    //     TestEntityType.Dashboard,
    //     "AdBids_model_dashboard",
    //     true
    //   ),
    //   waitForTimeSeries(page, "AdBids_model_dashboard"),
    //   waitForTopLists(page, "AdBids_model_dashboard", ["domain"]),
    // ]);

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
    await page.getByRole("menuitem", { name: "hour" }).click();

    // Change the time range
    await page
      .getByRole("button", { name: "All Time January 1 - March 30, 2022" })
      .click();
    await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();

    // Check that the total records are 275 and have comparisons
    await playwrightExpect(page.getByText("275 -12 -4%")).toBeVisible();

    // Turn off comparison
    await page
      .getByRole("button", { name: "Comparing to last period" })
      .click();
    await page.getByRole("menuitem", { name: "no comparison" }).click(); // tooltip is breaking this button rn...

    // Check number
    await playwrightExpect(
      page.getByText("275", { exact: true })
    ).toBeVisible();

    // Add comparison back
    await page.getByRole("button", { name: "no comparison" }).click();
    await page.getByRole("menuitem", { name: "last period" }).click();

    // Switch to a custom time range
    await page
      .getByRole("button", {
        name: "Last 6 Hours March 30, 2022 (5:00PM-10:59PM)",
      })
      .click();

    await page.getByRole("menuitem", { name: "Custom range" }).click();
    await page.getByLabel("Start date").fill("2022-02-01");
    await page.getByRole("button", { name: "Apply" }).click();

    // Check number
    await playwrightExpect(page.getByText("Total records 64.0k")).toBeVisible();

    // Flip back to All Time
    await page
      .getByRole("button", { name: "Custom range February 1 - March 29, 2022" })
      .click();
    await page.getByRole("menuitem", { name: "All Time" }).click();
    await page.getByText("Total records 100.0k").click();

    // Check number
    await playwrightExpect(
      page.getByText("Total records 100.0k", { exact: true })
    ).toBeVisible();

    // Filter to Facebook via leaderboard
    await page.getByRole("button", { name: "Facebook 19.3k" }).click();

    // Change filter to excluded
    await page.getByRole("button", { name: "Publisher Facebook" }).click();
    await page.getByRole("button", { name: "Exclude" }).click();
    await page
      .getByRole("button", { name: "Exclude Publisher Facebook" })
      .click();

    // Check number
    await playwrightExpect(
      page.getByText("Total records 80.7k", { exact: true })
    ).toBeVisible();

    // Clear the filter from filter bar
    await page
      .getByRole("button", { name: "Exclude Publisher Facebook" })
      .getByRole("button")
      .click();

    // Aply a different filter
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

    // Change time range to last 6 hours
    // await page
    //   .getByRole("button", { name: "All Time January 1 - March 30, 2022" })
    //   .click();
    // await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();

    // Check that the data is updated for last 6 hours
    // TODO

    // Change time range back to all time
    // TODO

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Get the dashboard name field and change it
    await page.getByLabel("Display name").fill("AdBids_model_dashboard_rename");
    await page.getByLabel("Display name").blur();
    // TODO: change model?

    // await asyncWait(3000);

    // Remove timestamp column
    await page.getByLabel("Remove timestamp column").click();

    // await asyncWait(3000);

    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    // await asyncWait(100000);

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

    // Check Quick Metrics button visible

    // Add Avg Bid Price metric, first without a definition

    // Check that Go to Dashboard is disabled

    // Add a definition and pick a format

    // Remove all dimensions

    // Check that Go to Dashboard is disabled

    // Add Published, Domain back to dashboard

    // Go to dashboard

    // Check Avg Bid Price

    // Change the leaderboard metric

    // Check domain and sample value in leaderboard

    // Open the Publisher details table

    // Check the first table row?

    // Change sort direction

    // Check new sort direction worked in first table row

    // Change the sort column to total rows

    // Check that first table row again

    // Click a table value to filter

    // Check that filter was applied

    // Check that details table can exclude

    // Add search criteria

    // Check that table got search

    // Clear search

    // Go back to leaderboard

    // Check that selected metric is total rows

    // Change the leaderboard metric to avg bid price
    // await page.getByRole("button", { name: "Total records" }).click();
  });
});
