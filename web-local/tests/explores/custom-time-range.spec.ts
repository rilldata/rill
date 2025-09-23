import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";

test.describe("custom timerange in Explore", () => {
  test.use({ project: "AdBids" });

  test("Custom time range should be visible by default", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

    await page.getByRole("button", { name: "Preview" }).click();
    await page.waitForTimeout(1000);

    await page.getByLabel("Select time range").click();

    await expect(page.getByRole("menuitem", { name: "Custom" })).toBeVisible();
  });

  test("Custom time range should not be visible when toggled off", async ({
    page,
  }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

    const watcher = new ResourceWatcher(page);

    const exploreWithBanner = `
  type: explore
  metrics_view: "AdBids_metrics"
  allow_custom_time_range: false
  measures: "*"
  dimensions: "*"
      `;

    await page.getByRole("button", { name: "switch to code editor" }).click();
    await watcher.updateAndWaitForDashboard(exploreWithBanner);
    await page.getByRole("button", { name: "Preview" }).click();
    await page.waitForTimeout(500);

    await page.getByLabel("Select time range").click();
    await expect(page.getByRole("menuitem", { name: "Custom" })).toBeHidden();
  });
});
