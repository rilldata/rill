import { createDashboardFromModel } from "../utils/dashboardHelpers";
import { createAdBidsModel } from "../utils/dataSpecifcHelpers";
import { test, expect } from "@playwright/test";
import { startRuntimeForEachTest } from "../utils/startRuntimeForEachTest";

test.describe("sorting flow", () => {
  startRuntimeForEachTest();

  test("Dashboard runthrough - sorting", async ({ page }) => {
    test.setTimeout(10000);

    await page.goto("/");
    // disable animations
    await page.addStyleTag({
      content: `
        *, *::before, *::after {
          animation-duration: 0s !important;
          transition-duration: 0s !important;
        }
      `,
    });
    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");

    // test.setTimeout(5000);

    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Publisher # null 32.9k Facebook 19.3k Google 18.8k Yahoo 18.6k Microsoft 10.4k",
      })
      .getByRole("button", { name: "Toggle sort leaderboards by value" })
      .click();
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Domain # news.google.com 12.9k sports.yahoo.com 12.9k instagram.com 13.1k news.y",
      })
      .getByRole("button", { name: "Toggle sort leaderboards by value" })
      .click();
    await page.getByRole("button", { name: "Select a context column" }).click();
    await page.getByRole("menuitem", { name: "Percent of total" }).click();
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Publisher # % null 32.9k 32% Facebook 19.3k 19% Google 18.8k 18% Yahoo 18.6k 18%",
      })
      .getByRole("button", {
        name: "Toggle sort leaderboards by context column",
      })
      .click();
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
    await page.getByRole("button", { name: "Select time range" }).click();
    await page.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    await page.getByRole("button", { name: "Select a context column" }).click();

    debugger;
    // await test.pause();

    await page.getByRole("menuitem", { name: "Absolute change" }).click();
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Publisher # Microsoft 5.7k 2.9k Facebook 6.9k 2.7k null 8.7k -1.6k Yahoo 2.5k -4",
      })
      .getByRole("button", { name: "Toggle sort leaderboards by value" })
      .click();
    await page
      .locator("svelte-virtual-list-row")
      .filter({
        hasText:
          "Domain # facebook.com 8.5k 4.5k msn.com 8.5k 4.5k news.google.com 1.9k -360 goog",
      })
      .getByRole("button", {
        name: "Toggle sort leaderboards by context column",
      })
      .click();
    await page.locator("#svelte").getByText("Publisher").click();
    await page.locator("button:nth-child(4)").first().click();
  });
});
