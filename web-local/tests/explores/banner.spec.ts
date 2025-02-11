import { expect } from "@playwright/test";
import { test } from "../utils/test";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";
import { useDashboardFlowTestSetup } from "web-local/tests/explores/dashboard-flow-test-setup";

test.describe("banner in explore preview", () => {
    useDashboardFlowTestSetup();

    test("Banner should not be visible in explore preview", async ({ page }) => {

        await page.getByRole("button", { name: "Preview" }).click();
        await page.waitForTimeout(1000);

        const banner = page.locator('.app-banner');
        await expect(banner).toBeHidden();
    });

    test("Banner should be visible in explore preview", async ({ page }) => {
        const watcher = new ResourceWatcher(page);

        const exploreWithBanner = `
  type: explore
  metrics_view: "AdBids_model_metrics"
  banner: "This is a banner text"
  measures: "*"
  dimensions: "*"
      `;

        await page.getByLabel("code").click();
        await watcher.updateAndWaitForDashboard(exploreWithBanner);
        await page.getByRole("button", { name: "Preview" }).click();
        await page.waitForTimeout(500);

        const banner = page.locator('.app-banner');
        await expect(banner).toBeVisible();
        await expect(banner).toHaveText(/This is a banner text/);
    });

});