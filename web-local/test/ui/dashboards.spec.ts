import { describe, it } from "@jest/globals";
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

  it.only("Autogenerate dashboard from model", async () => {
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
});
