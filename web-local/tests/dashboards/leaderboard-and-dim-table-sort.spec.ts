import { useDashboardFlowTestSetup } from "web-local/tests/dashboards/dashboard-flow-test-setup";
import { test, expect, Locator } from "@playwright/test";
import { startRuntimeForEachTest } from "../utils/startRuntimeForEachTest";

async function assertAAboveB(locA: Locator, locB: Locator) {
  const topA = await locA.boundingBox().then((box) => box?.y);
  const topB = await locB.boundingBox().then((box) => box?.y);

  await expect(topA).toBeDefined();
  await expect(topB).toBeDefined();

  // Safety: topB is defined
  await expect(topA).toBeLessThan(topB as number);
}

test.describe("leaderboard and dimension table sorting", () => {
  startRuntimeForEachTest();
  useDashboardFlowTestSetup();

  test("leaderboard and dimension table sorting", async ({ page }) => {
    /**
     * LEADERBOARD
     */
    await assertAAboveB(
      page.getByRole("button", { name: "null 32.9k" }),
      page.getByRole("button", { name: "Microsoft 10.4k" }),
    );

    // toggle sort by value
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Publisher # null 32.9k Facebook 19.3k Google 18.8k Yahoo 18.6k Microsoft 10.4k",
      })
      .getByRole("button", { name: "Toggle sort leaderboards by value" })
      .click();

    await assertAAboveB(
      page.getByRole("button", { name: "Microsoft 10.4k" }),
      page.getByRole("button", { name: "null 32.9k" }),
    );

    // add pct of total context column
    await page.getByRole("button", { name: "Select a context column" }).click();
    await page.getByRole("menuitem", { name: "Percent of total" }).click();

    await assertAAboveB(
      page.getByRole("button", { name: "Microsoft 10.4k 10%" }),
      page.getByRole("button", { name: "null 32.9k 32%" }),
    );

    //toggle sort by pct of total
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Domain # % news.google.com 12.9k 12% sports.yahoo.com 12.9k 12% instagram.com 13",
      })
      .getByRole("button", {
        name: "Toggle sort leaderboards by context column",
      })
      .click();

    await assertAAboveB(
      page.getByRole("button", { name: "facebook.com 15.6k 15%" }),
      page.getByRole("button", { name: "news.google.com 12.9k 12%" }),
    );

    // add time comparison and select Pct change
    await page.getByRole("button", { name: "No comparison" }).click();
    await page.getByRole("menuitem", { name: "Time" }).click();

    await page.getByRole("button", { name: "Select time range" }).click();
    await page.getByRole("menuitem", { name: "Last 24 Hours" }).click();

    // need a slight delay for the time range to update
    // and the "Pct change" option to be available
    // in the context column dropdown
    await page.waitForTimeout(1000);

    await page.getByRole("button", { name: "Select a context column" }).click();
    await page.getByRole("menuitem", { name: "Percent change" }).click();

    // need a slight delay for the rankings to update
    await page.waitForTimeout(1000);

    // Broader selectors using RegEx to account for some Playwright runs triggering the display
    // of the starting value on hover
    await assertAAboveB(
      page.getByRole("button", { name: /^Google/ }),
      page.getByRole("button", { name: /^Facebook/ }),
    );

    // toggle sort by pct change
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Publisher # % Facebook 283 -4% Microsoft 237 ~0% null 383 0% Yahoo 103 3% Google",
      })
      .getByRole("button", {
        name: "Toggle sort leaderboards by context column",
      });

    await assertAAboveB(
      page.getByRole("button", { name: "news.yahoo.com 89 15%" }),
      page.getByRole("button", { name: "sports.yahoo.com 67 -27%" }),
    );

    // select absolute change
    await page.getByRole("button", { name: "Select a context column" }).click();
    await page.getByRole("menuitem", { name: "Absolute change" }).click();

    await assertAAboveB(
      page.getByRole("button", { name: "Google 116 5" }),
      page.getByRole("button", { name: "Facebook 283 -14" }),
    );

    // toggle sort by absolute change
    await page
      .getByRole("button", {
        name: "Toggle sort leaderboards by context column",
      })
      .nth(1)
      .click();

    await assertAAboveB(
      page.getByRole("button", { name: "Facebook 283 -14" }),
      page.getByRole("button", { name: "Google 116 5" }),
    );

    /**
     * DIMENSION TABLE
     */
    // click publisher leaderboard header to enter dimension table for that dim
    await page.locator("#svelte").getByText("Publisher").click();

    // click publisher column header to sort by publisher
    await page.getByRole("button", { name: "Publisher" }).click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^Yahoo$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^Microsoft$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by total records
    await page
      .getByRole("table", { name: "Dimension table" })
      .getByRole("button")
      .filter({ hasText: "Total records" })
      .click();
    await assertAAboveB(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^383$/ }),
      page
        .locator("div")
        .filter({ hasText: /^283$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by absolute change once to sort by absolute change descending
    await page.locator(".w-full > button:nth-child(2)").click();
    await assertAAboveB(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^2$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^-14$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );
    // sort by absolute change TWICE to sort by absolute change ascending
    await page.locator(".w-full > button:nth-child(2)").click();
    await assertAAboveB(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^-14$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^2$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by pct change ONCE to sort by pct change descending
    await page.locator("button:nth-child(3)").first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^4%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^3%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by pct change TWICE to sort by pct change ascending
    await page.locator("button:nth-child(3)").first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^3%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^4%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by pct of total ONCE to sort by pct of total descending
    await page.locator("button:nth-child(4)").first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^34%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^25%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );
    // sort by pct of total TWICE to sort by pct of total ascending
    await page.locator("button:nth-child(4)").first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^25%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^34%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );
  });
});
