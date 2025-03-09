import {
  createExploreFromModel,
  createExploreFromSource,
} from "../utils/exploreHelpers";
import { assertLeaderboards } from "../utils/metricsViewHelpers";
import { wrapRetryAssertion } from "../utils/commonHelpers";
import {
  assertAdBidsDashboard,
  createAdBidsModel,
} from "../utils/dataSpecifcHelpers";
import { createSource } from "../utils/sourceHelpers";
import { test } from "../setup/base";

test.describe("explores", () => {
  test.use({ project: "Blank" });

  test("Autogenerate explore from source", async ({ page }) => {
    await createSource(page, "AdBids.csv", "/sources/AdBids.yaml");
    await createExploreFromSource(page);
    // Temporary timeout while the issue is looked into
    await page.waitForTimeout(1000);
    await assertAdBidsDashboard(page);
  });

  test("Autogenerate explore from model", async ({ page }) => {
    await createAdBidsModel(page);
    await createExploreFromModel(page, true);
    await assertAdBidsDashboard(page);

    // click on publisher=Facebook leaderboard value
    await page.getByRole("row", { name: "Facebook 19.3k" }).click();
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
});
