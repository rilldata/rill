import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas leaderboards", () => {
  test.use({ project: "AdBids" });

  test("add leaderboard component", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");
    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();

    await page.getByRole("menuitem", { name: "Leaderboard" }).click();

    await page.getByLabel("Add measure fields").click();

    await page.getByRole("menuitem", { name: "Sum of Bid Price" }).click();
    await page
      .getByLabel("domain leaderboard")
      .locator("button")
      .filter({ hasText: "Sum of Bid Price" })
      .click();
  });
});
