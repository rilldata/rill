import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

// AdBids_grains_canvas renders a line chart on AdBids_grains_metrics, which declares
// `time_grains: [10m, hour, day, week]`. The canvas filter bar offers those grains that fit the
// 24-hour default range and the chart re-queries with the 10-minute multiple once it is picked.
test.describe("canvas time grains", () => {
  test.use({ project: "AdBids" });

  test("the filter bar offers only the declared grains", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_grains_canvas.yaml");

    const grainSelector = page.getByLabel("Select aggregation grain");
    await expect(grainSelector).toContainText("by hour");

    await grainSelector.click();
    const menu = page.getByRole("menu");
    for (const grain of ["10 minutes", "hour", "day"]) {
      await expect(
        menu.getByRole("menuitemcheckbox", { name: grain, exact: true }),
      ).toBeVisible();
    }
    for (const grain of ["minute", "week", "month"]) {
      await expect(
        menu.getByRole("menuitemcheckbox", { name: grain, exact: true }),
      ).toHaveCount(0);
    }

    const chartRequest = page.waitForRequest(
      (request) =>
        request.url().includes("QueryService/MetricsViewAggregation") &&
        (request.postData() ?? "").includes("TIME_GRAIN_10_MINUTE"),
    );
    await menu
      .getByRole("menuitemcheckbox", { name: "10 minutes", exact: true })
      .click();
    await chartRequest;

    await expect(grainSelector).toContainText("by 10 minutes");
    expect(new URL(page.url()).searchParams.get("grain")).toBe("10m");
  });
});
