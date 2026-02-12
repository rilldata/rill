import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Time grain derivation from URL", () => {
  test.use({ project: "AdBids" });

  // Helper to run a time grain derivation test
  async function testGrainDerivation(
    page: Page,
    timeRange: string,
    expectedGrain: string,
  ) {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?tr=${encodeURIComponent(timeRange)}`,
    );

    // Wait for the time grain selector to appear and contain the expected grain
    const timeGrainSelector = page.getByLabel("Select aggregation grain");
    await expect(timeGrainSelector).toContainText(`by ${expectedGrain}`, {
      timeout: 10000,
    });

    // Check the grain parameter in the URL
    const url = new URL(page.url());
    const grain = url.searchParams.get("grain");
    expect(grain).toBe(expectedGrain);
  }

  const cases: [string, string, string][] = [
    // Basic ISO duration tests
    ["P7D (7 days)", "P7D", "day"],
    ["PT6H (6 hours)", "PT6H", "hour"],
    ["P4W (4 weeks)", "P4W", "week"],

    // Rill time syntax tests with "as of latest" format
    ["365d as of latest/h+1h", "365d as of latest/h+1h", "day"],
    ["12M as of latest/m+1m", "12M as of latest/m+1m", "day"],
    ["365M as of latest/d+1d", "365M as of latest/d+1d", "month"],

    // Edge cases around bucket limits (1500 max)
    ["24h as of latest/h", "24h as of latest/h", "hour"],
    ["90d as of latest/d", "90d as of latest/d", "day"],
    ["2y as of latest/M", "2y as of latest/M", "month"],
    ["5y as of latest/d", "5y as of latest/d", "week"],

    // Period-to-date tests
    ["week-to-date", "rill-WTD", "day"],
    ["month-to-date", "rill-MTD", "day"],

    // Snap grain should influence derived grain
    ["7d as of latest/h", "7d as of latest/h", "hour"],
    ["52w as of latest/w", "52w as of latest/w", "week"],

    // Bucket limit boundary tests (1500 max buckets)
    ["24h as of latest/m (1440 buckets)", "24h as of latest/m", "minute"],
    ["26h as of latest/m (1560 buckets)", "26h as of latest/m", "hour"],
    ["62d as of latest/h (1488 buckets)", "62d as of latest/h", "hour"],
    ["64d as of latest/h (1536 buckets)", "64d as of latest/h", "day"],

    // Very large intervals
    ["10y as of latest/M", "10y as of latest/M", "month"],
    ["6y as of latest/d", "6y as of latest/d", "week"],

    // Quarter grain tests
    ["3y as of latest/Q", "3y as of latest/Q", "quarter"],
    ["20y as of latest/M", "20y as of latest/M", "month"],
  ];

  test("derives correct grain for all time ranges", async ({ page }) => {
    for (const [label, timeRange, expectedGrain] of cases) {
      await test.step(`${label} → ${expectedGrain}`, async () => {
        await testGrainDerivation(page, timeRange, expectedGrain);
      });
    }
  });
});
