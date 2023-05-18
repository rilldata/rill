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
    await page
      .getByRole("button", { name: "All Time January 1 - March 30, 2022" })
      .click();
    await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();

    // Change the leaderboard metric to avg bid price
    await page.getByRole("button", { name: "Total records" }).click();

    return;
    await page.getByRole("menuitem", { name: "Avg Bid Price" }).click();

    // Check leaderboard entry for comparison
    await playwrightExpect(
      page.getByText("Facebook 3.02 → 3.37 11%", { exact: true })
    ).toBeVisible();
    // await page
    //   .getByRole("button", { name: "Facebook 3.02 → 3.37 11%" })
    //   .click();
    // await page
    //   .getByRole("button", { name: "Publisher Facebook" })
    //   .getByRole("button")
    //   .click();

    // // Open the leaderboard details
    // await page.getByRole("button", { name: "Publisher" }).click();

    // // TODO, add more here

    // // Go back
    // await page.getByRole("button", { name: "All Dimensions" }).click();
  });
});
