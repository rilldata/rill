import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { waitForReconciliation } from "../utils/wait-for-reconciliation.ts";

// AdBids_grains_metrics declares `time_grains: [10m, hour, day, week]`, so the explore offers
// exactly those grains that fit its 24-hour default range (week does not) and queries with the
// 10-minute multiple once it is picked. The grain rides in the URL as `grain=10m`.
test.describe("metrics view time grains", () => {
  test.use({ project: "AdBids" });

  test("the explore offers only the declared grains and buckets by 10 minutes", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);
    await page.goto(`${baseUrl}/explore/AdBids_grains_metrics_explore`);

    const grainSelector = page.getByLabel("Select aggregation grain");
    // 24 hourly points keep the range's own precision in adaptive mode.
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

    const timeSeriesRequest = page.waitForRequest(
      (request) =>
        request.url().includes("QueryService/MetricsViewTimeSeries") &&
        (request.postData() ?? "").includes(
          '"timeGranularity":"TIME_GRAIN_10_MINUTE"',
        ),
    );
    await menu
      .getByRole("menuitemcheckbox", { name: "10 minutes", exact: true })
      .click();
    await timeSeriesRequest;

    await expect(grainSelector).toContainText("by 10 minutes");
    expect(new URL(page.url()).searchParams.get("grain")).toBe("10m");

    // The grain is restored from the URL.
    await page.goto(`${baseUrl}/explore/AdBids_grains_metrics_explore?grain=10m`);
    await expect(page.getByLabel("Select aggregation grain")).toContainText(
      "by 10 minutes",
    );
  });
});
