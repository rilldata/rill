import path from "node:path";
import type { Page } from "playwright";
import { assertLeaderboards } from "./dashboardHelpers";
import { waitForProfiling, wrapRetryAssertion } from "./helpers";
import { createModel, updateModelSql } from "./modelHelpers";
import { waitForSource } from "./sourceHelpers";

export async function waitForAdBids(page: Page, name: string) {
  return waitForSource(page, name, ["publisher", "domain", "timestamp"]);
}

export async function waitForAdImpressions(page: Page, name: string) {
  return waitForSource(page, name, ["city", "country"]);
}

export async function createAdBidsModel(page: Page) {
  await createModel(page, "AdBids_model");
  await Promise.all([
    waitForProfiling(page, "AdBids_model", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    updateModelSql(
      page,
      `from "${path.join(__dirname, "../../data", "AdBids.csv")}"`
    ),
  ]);
}

export async function assertAdBidsDashboard(page: Page) {
  await wrapRetryAssertion(() =>
    assertLeaderboards(page, [
      {
        label: "Publisher",
        values: ["null", "Facebook", "Google", "Yahoo", "Microsoft"],
      },
      {
        label: "Domain",
        values: [
          "facebook.com",
          "msn.com",
          "google.com",
          "news.yahoo.com",
          "instagram.com",
          "sports.yahoo.com",
          "news.google.com",
        ],
      },
    ])
  );
  // TODO: how do we assert timeseries?
}
