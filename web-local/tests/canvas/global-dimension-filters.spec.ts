import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas global dimension filters", () => {
  test.use({ project: "AdBids" });

  // TODO: Fix test with latest filter related changes
  test.skip("global dimension filters run through", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByRole("button", { name: "Preview" }).click();

    await page.waitForTimeout(1000);

    await expect(page.getByText("Total records 1,122")).toBeVisible();

    await page.getByRole("button", { name: "Add filter button" }).click();
    await page.getByRole("menuitem", { name: "Publisher" }).click();

    await page.getByLabel("publisher results").getByText("Facebook").click();

    await expect(page.getByText("Total records 283")).toBeVisible();

    // Change filter to excluded
    await page.getByLabel("Include exclude toggle").click();
    await page.getByText("Exclude Publisher Facebook").click();

    await expect(page.getByText("Total records 839")).toBeVisible();

    await page.getByRole("button", { name: "Clear filters" }).click();

    await expect(page.getByText("Total records 1,122")).toBeVisible();
  });
});
