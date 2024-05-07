import { expect, type Locator } from "@playwright/test";
import { useDashboardFlowTestSetup } from "web-local/tests/dashboards/dashboard-flow-test-setup";
import { test } from "../utils/test";

async function assertAAboveB(locA: Locator, locB: Locator) {
  const topA = await locA.boundingBox().then((box) => box?.y);
  const topB = await locB.boundingBox().then((box) => box?.y);

  expect(topA).toBeDefined();
  expect(topB).toBeDefined();

  // Safety: topB is defined
  expect(topA).toBeLessThan(topB as number);
}

test.describe("leaderboard and dimension table sorting", () => {
  useDashboardFlowTestSetup();

  test("leaderboard and dimension table sorting", async ({ page }) => {
    await page.getByRole("button", { name: "Preview" }).click();

    /**
     * LEADERBOARD
     */
    await assertAAboveB(
      page.getByRole("button", { name: "null 32.9k" }),
      page.getByRole("button", { name: "Microsoft 10.4k" }),
    );

    await page
      .getByLabel("publisher leaderboard")
      .getByLabel("Toggle sort leaderboards by value")
      .click();

    await assertAAboveB(
      page.getByRole("button", { name: "Microsoft 10.4k" }),
      page.getByRole("button", { name: "null 32.9k" }),
    );

    const timeRangeMenu = page.getByRole("button", {
      name: "Select time range",
    });
    const contextColumnMenu = page.getByRole("button", {
      name: "Select a context column",
    });

    async function openTimeRangeMenu() {
      await timeRangeMenu.click();
      await page
        .getByRole("menu", { name: "Select time range" })
        .waitFor({ state: "visible" });
    }

    async function openContextColumnMenu() {
      await contextColumnMenu.click();
      await page.getByRole("menu").waitFor({ state: "visible" });
    }

    // add pct of total context column
    await openContextColumnMenu();
    await page.getByRole("menuitem", { name: "Percent of total" }).click();

    await assertAAboveB(
      page.getByRole("button", { name: "Microsoft 10.4k 10%" }),
      page.getByRole("button", { name: "null 32.9k 33%" }),
    );

    //toggle sort by pct of total
    await page
      .getByLabel("publisher leaderboard")
      .getByLabel("Toggle sort leaderboards by context column")
      .click();

    await assertAAboveB(
      page.getByRole("button", { name: "facebook.com 15.6k 16%" }),
      page.getByRole("button", { name: "news.google.com 12.9k 13%" }),
    );

    // add time comparison and select Pct change
    await page
      .getByRole("button", { name: "No comparison", exact: true })
      .click();
    await page.getByRole("menuitem", { name: "Time" }).click();
    await page.keyboard.press("Escape");

    await openTimeRangeMenu();
    await page.getByRole("menuitem", { name: "Last 24 Hours" }).click();

    // need a slight delay for the time range to update
    // and the "Pct change" option to be available
    // in the context column dropdown
    await page.waitForTimeout(1000);

    await openContextColumnMenu();
    await page.getByRole("menuitem", { name: "Percent change" }).click();

    // need a slight delay for the rankings to update
    await page.waitForTimeout(1000);

    // Broader selectors using RegEx to account for some Playwright runs triggering the display
    // of the starting value on hover
    await assertAAboveB(
      page.getByRole("button", { name: /^Google/ }),
      page.getByRole("button", { name: /^Facebook/ }),
    );

    await assertAAboveB(
      page.getByRole("button", { name: "news.yahoo.com 89 16%" }),
      page.getByRole("button", { name: "sports.yahoo.com 67 -27%" }),
    );

    // select absolute change
    await openContextColumnMenu();
    await page.getByRole("menuitem", { name: "Absolute change" }).click();

    await assertAAboveB(
      page.getByRole("button", { name: "Google 116 5" }),
      page.getByRole("button", { name: "Facebook 283 -14" }),
    );

    // toggle sort by absolute change
    await page
      .getByLabel("publisher leaderboard")
      .getByRole("button", {
        name: "Toggle sort leaderboards by context column",
      })
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

    // await page.waitForTimeout(60000);

    // sort by pct change ONCE to sort by pct change descending
    // await page.locator(".w-full > button:nth-child(3)").first().click();
    await page.getByRole("button", { name: "%" }).first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^5%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^3%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by pct change TWICE to sort by pct change ascending
    await page.getByRole("button", { name: "%" }).first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^3%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^5%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by pct of total ONCE to sort by pct of total descending
    await page.getByRole("button", { name: "%" }).nth(1).click();
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
    await page.getByRole("button", { name: "%" }).nth(1).click();
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
