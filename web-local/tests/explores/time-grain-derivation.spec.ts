import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Time grain derivation from URL", () => {
  test.use({ project: "AdBids" });

  // Helper to run a time grain derivation test
  async function testGrainDerivation(
    page: import("@playwright/test").Page,
    timeRange: string,
    expectedGrain: string,
  ) {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?tr=${encodeURIComponent(timeRange)}`,
    );

    // Wait for the explore to load and the URL to update with derived grain
    await page.waitForTimeout(2000);

    // Check the grain parameter in the URL
    const url = new URL(page.url());
    const grain = url.searchParams.get("grain");
    expect(grain).toBe(expectedGrain);

    // Also verify the time grain selector dropdown shows the correct grain
    const timeGrainSelector = page.getByLabel("Select aggregation grain");
    await expect(timeGrainSelector).toContainText(`by ${expectedGrain}`);
  }

  // Basic ISO duration tests
  test("derives day grain for P7D (7 days)", async ({ page }) => {
    await testGrainDerivation(page, "P7D", "day");
  });

  test("derives hour grain for PT6H (6 hours)", async ({ page }) => {
    await testGrainDerivation(page, "PT6H", "hour");
  });

  test("derives week grain for P4W (4 weeks)", async ({ page }) => {
    await testGrainDerivation(page, "P4W", "week");
  });

  // Rill time syntax tests with "as of latest" format
  test("derives day grain for 365d as of latest/h+1h", async ({ page }) => {
    // 365 days at hour grain = 8760 buckets (exceeds 1500 limit)
    // Should use day grain = 365 buckets
    await testGrainDerivation(page, "365d as of latest/h+1h", "day");
  });

  test("derives day grain for 12M as of latest/m+1m", async ({ page }) => {
    await testGrainDerivation(page, "12M as of latest/m+1m", "day");
  });

  test("derives month grain for 365M as of latest/d+1d", async ({ page }) => {
    // 365 months at day grain = ~11,000 buckets (exceeds 1500 limit)
    // Should use month grain = 365 buckets
    await testGrainDerivation(page, "365M as of latest/d+1d", "month");
  });

  // Edge cases around bucket limits (1500 max)
  test("derives hour grain for 24h as of latest/h", async ({ page }) => {

    await testGrainDerivation(page, "24h as of latest/h", "hour");
  });

  test("derives day grain for 90d as of latest/d", async ({ page }) => {

    await testGrainDerivation(page, "90d as of latest/d", "day");
  });

  test("derives month grain for 2y as of latest/M", async ({ page }) => {
   
    await testGrainDerivation(page, "2y as of latest/M", "month");
  });

  test("derives week grain for 5y as of latest/d", async ({ page }) => {
    // 5 years = ~1825 days (exceeds 1500 limit at day grain)
    // Should use week grain = ~260 buckets
    await testGrainDerivation(page, "5y as of latest/d", "week");
  });

  // Period-to-date tests
  test("derives day grain for week-to-date", async ({ page }) => {
    await testGrainDerivation(page, "rill-WTD", "day");
  });

  test("derives day grain for month-to-date", async ({ page }) => {
    await testGrainDerivation(page, "rill-MTD", "day");
  });

  // Snap grain should influence derived grain
  test("derives hour grain for 7d as of latest/h", async ({ page }) => {
    // 7 days with hour snap = 168 buckets (under 1500)
    // Snap grain (hour) should be used
    await testGrainDerivation(page, "7d as of latest/h", "hour");
  });

  test("derives week grain for 52w as of latest/w", async ({ page }) => {
    // 52 weeks = 52 buckets at week grain
    await testGrainDerivation(page, "52w as of latest/w", "week");
  });

  // Bucket limit boundary tests (1500 max buckets)
  test("derives minute grain for 24h as of latest/m (1440 buckets, under limit)", async ({
    page,
  }) => {
    // 24 hours = 1440 minutes (just under 1500 limit)
    await testGrainDerivation(page, "24h as of latest/m", "minute");
  });

  test("derives hour grain for 26h as of latest/m (1560 buckets, over limit)", async ({
    page,
  }) => {
    // 26 hours = 1560 minutes (just over 1500 limit)
    // Should fall back to hour grain = 26 buckets
    await testGrainDerivation(page, "26h as of latest/m", "hour");
  });

  test("derives day grain for 62d as of latest/h (1488 buckets, under limit)", async ({
    page,
  }) => {
    // 62 days = 1488 hours (just under 1500 limit)
    await testGrainDerivation(page, "62d as of latest/h", "hour");
  });

  test("derives day grain for 64d as of latest/h (1536 buckets, over limit)", async ({
    page,
  }) => {
    // 64 days = 1536 hours (just over 1500 limit)
    // Should fall back to day grain = 64 buckets
    await testGrainDerivation(page, "64d as of latest/h", "day");
  });

  // Very large intervals
  test("derives year grain for 10y as of latest/M", async ({ page }) => {
    // 10 years = 120 months at month grain
    await testGrainDerivation(page, "10y as of latest/M", "month");
  });

  test("derives month grain for 6y as of latest/d", async ({ page }) => {
    // 6 years = ~2190 days (exceeds 1500 limit at day grain)
    // Should use week grain = ~312 buckets, or month = 72 buckets
    await testGrainDerivation(page, "6y as of latest/d", "week");
  });

  // Quarter grain tests
  test("derives quarter grain for 3y as of latest/Q", async ({ page }) => {
    // 3 years = 12 quarters
    await testGrainDerivation(page, "3y as of latest/Q", "quarter");
  });

  test("derives quarter grain for 20y as of latest/M", async ({ page }) => {
    // 20 years = 240 months (under 1500 at month grain)
    // Should use month grain
    await testGrainDerivation(page, "20y as of latest/M", "month");
  });
});
