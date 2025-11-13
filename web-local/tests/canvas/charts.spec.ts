import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas charts", () => {
  test.use({ project: "AdBids" });

  test("switch between charts", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.locator("#AdBids_metrics_canvas--component-1-0 canvas").click();

    await page.locator(".chart-icons").getByLabel("Heatmap").click();

    await page.waitForTimeout(500);

    await page
      .getByLabel("A heatmap chart with embedded data")
      .locator("canvas")
      .click();

    await page.locator(".chart-icons").getByLabel("Donut").click();

    await page.waitForTimeout(500);

    await page
      .getByLabel("A arc chart with embedded")
      .locator("canvas")
      .click();
  });
});
