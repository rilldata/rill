import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { waitForReconciliation } from "../utils/wait-for-reconciliation.ts";

test.describe("Adaptive time grain derivation from URL", () => {
  test.use({ project: "AdBids" });

  // Loads the explore with only a time range and checks the grain the adaptive mode picks.
  async function testGrainDerivation(
    page: Page,
    timeRange: string,
    expectedGrain: string,
  ) {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    // Wait for all resources to reconcile
    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?tr=${encodeURIComponent(timeRange)}`,
    );

    // Wait for the time grain selector to appear and contain the expected grain
    const timeGrainSelector = page.getByLabel("Select aggregation grain");
    await expect(timeGrainSelector).toContainText(`by ${expectedGrain}`);

    // An adaptive grain is derived from the range, so it is never written to the URL.
    const url = new URL(page.url());
    expect(url.searchParams.get("grain")).toBeNull();
  }

  // The range's own precision is kept when it gives 6 to 400 points.
  // Otherwise the allowed grain closest to 30 points wins, preferring the coarser one on a tie.
  // The data ends on 2022-03-30, so "latest" resolves relative to that.
  const cases: [string, string, string][] = [
    // ISO durations keep their precision when it gives enough points
    ["P7D (7 daily points)", "P7D", "day"],
    ["PT6H (6 hourly points)", "PT6H", "hour"],
    ["P4W (4 weekly points is too few)", "P4W", "day"],

    // Snap grains below the 1500 bucket cap are re-sized to the nearest usable grain
    ["365d as of latest/h+1h", "365d as of latest/h+1h", "week"],
    ["12M as of latest/m+1m", "12M as of latest/m+1m", "week"],
    ["365M as of latest/d+1d", "365M as of latest/d+1d", "year"],

    // Precision kept within 6..400 points
    ["24h as of latest/h", "24h as of latest/h", "hour"],
    ["90d as of latest/d", "90d as of latest/d", "day"],
    ["2y as of latest/M", "2y as of latest/M", "month"],
    ["5y as of latest/d (1826 daily points)", "5y as of latest/d", "quarter"],

    // Period-to-date ranges are sized by their elapsed duration
    ["week-to-date (3 days)", "rill-WTD", "hour"],
    ["month-to-date (30 days)", "rill-MTD", "day"],

    // Snap grain sets the precision
    ["7d as of latest/h", "7d as of latest/h", "hour"],
    ["52w as of latest/w", "52w as of latest/w", "week"],

    // Too many points at the range precision
    ["24h as of latest/m (1440 points)", "24h as of latest/m", "hour"],
    ["26h as of latest/m (1560 points)", "26h as of latest/m", "hour"],
    ["62d as of latest/h (1488 points)", "62d as of latest/h", "day"],
    ["64d as of latest/h (1536 points)", "64d as of latest/h", "day"],

    // Very large intervals
    ["10y as of latest/M", "10y as of latest/M", "month"],
    ["6y as of latest/d", "6y as of latest/d", "quarter"],

    // Quarter grain tests
    ["3y as of latest/Q", "3y as of latest/Q", "quarter"],
    ["20y as of latest/M", "20y as of latest/M", "month"],
  ];

  test("derives correct grain for all time ranges", async ({ page }) => {
    // Each case is a full page load, and there are 22 of them, so the default
    // 30s budget leaves no headroom on CI hardware.
    test.setTimeout(90_000);

    for (const [label, timeRange, expectedGrain] of cases) {
      await test.step(`${label} → ${expectedGrain}`, async () => {
        await testGrainDerivation(page, timeRange, expectedGrain);
      });
    }
  });
});
