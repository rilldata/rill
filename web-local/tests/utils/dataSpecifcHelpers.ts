import type { Page } from "playwright";
import {
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "./commonHelpers";
import { assertLeaderboards } from "./dashboardHelpers";
import { createModel } from "./modelHelpers";
import { uploadFile, waitForSource } from "./sourceHelpers";

export async function createAdBidsModel(page: Page) {
  await Promise.all([
    waitForSource(page, "sources/AdBids.yaml", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    uploadFile(page, "AdBids.csv"),
  ]);

  await createModel(page, "AdBids_model");
  await Promise.all([
    waitForProfiling(page, "AdBids_model", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    updateCodeEditor(page, `select * from "AdBids"`),
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
    ]),
  );
  // TODO: how do we assert timeseries?
}
