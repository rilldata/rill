import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";

test.describe("banner in explore preview", () => {
  test.use({ project: "AdBids" });

  test("Banner should not be visible in explore preview", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

    await page.getByRole("button", { name: "Preview" }).click();
    await page.waitForTimeout(1000);

    const banner = page.locator(".app-banner");
    await expect(banner).toBeHidden();
  });

  test("Banner should be visible in explore preview", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

    const watcher = new ResourceWatcher(page);

    const exploreWithBanner = `
  type: explore
  metrics_view: "AdBids_metrics"
  banner: "This is a banner text"
  measures: "*"
  dimensions: "*"
      `;

    await page.getByRole("button", { name: "switch to code editor" }).click();
    await watcher.updateAndWaitForDashboard(exploreWithBanner);
    await page.getByRole("button", { name: "Preview" }).click();
    await page.waitForTimeout(500);

    const banner = page.locator(".app-banner");
    await expect(banner).toBeVisible();
    await expect(banner).toHaveText(/This is a banner text/);
  });
});
