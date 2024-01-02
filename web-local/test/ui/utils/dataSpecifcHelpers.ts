import type { Page } from "playwright";
import { updateCodeEditor } from "./commonHelpers";
import { assertLeaderboards } from "./dashboardHelpers";
import { waitForProfiling, wrapRetryAssertion } from "./helpers";
import { createModel } from "./modelHelpers";
import { uploadFile, waitForSource } from "./sourceHelpers";

export async function waitForAdBids(page: Page, name: string) {
  return waitForSource(page, name, ["publisher", "domain", "timestamp"]);
}

export async function waitForAdImpressions(page: Page, name: string) {
  return waitForSource(page, name, ["city", "country"]);
}

export async function createAdBidsModel(page: Page) {
  await Promise.all([
    waitForAdBids(page, "AdBids"),
    uploadFile(page, "AdBids.csv"),
  ]);

  await createModel(page, "AdBids_model");

  await page.waitForTimeout(200);
  await waitForProfiling(page, "AdBids_model", [
    "publisher",
    "domain",
    "timestamp",
  ]);

  await page.waitForTimeout(200);
  await updateCodeEditor(page, `select * from "AdBids"`);
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
