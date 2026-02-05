import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { gotoNavEntry } from "../utils/waitHelpers";
import {
  interactWithTimeRangeMenu,
  setDashboardTimezone,
} from "@rilldata/web-common/tests/utils/explore-interactions";
import { formatGrainBucket } from "@rilldata/web-common/lib/time/ranges/formatter";
import { DateTime } from "luxon";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";

// Annotation timestamps as they'll be serialized from DuckDB (UTC).
// All annotations that may be visible at day grain in "Last 7 days":
const DAY_ANNOTATION_TIMES = [
  "2022-03-24T00:00:00Z", // Point A
  "2022-03-25T00:00:00Z", // Range E (start)
  "2022-03-26T00:00:00Z", // Point B
  "2022-03-27T00:00:00Z", // Point C
  "2022-03-28T00:00:00Z", // Point D
  "2022-03-30T06:00:00Z", // Hour E (snaps to Mar 30 at day grain in UTC)
  "2022-03-30T14:00:00Z", // Hour F (same day bucket as E in UTC, but splits to Mar 30 in LA)
];

// Hour-level annotations visible in "Last 24 hours":
const HOUR_ANNOTATION_TIMES = [
  "2022-03-30T06:00:00Z", // Hour E
  "2022-03-30T14:00:00Z", // Hour F
];

// The app sets Settings.defaultLocale = "en" globally, so chart labels are
// always English regardless of the browser's navigator.language.
const CHART_LOCALE = "en";

function expectedDates(
  times: string[],
  grain: V1TimeGrain,
  dashboardTZ: string,
): Set<string> {
  const dates = times.map((ts) =>
    formatGrainBucket(
      DateTime.fromISO(ts, { zone: dashboardTZ, locale: CHART_LOCALE }),
      grain,
    ),
  );
  return new Set(dates);
}

async function setupDashboard(page: Page) {
  await page.getByLabel("/dashboards").click();
  await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

  await expect(
    page.getByRole("button", { name: /Total records/ }).first(),
  ).toBeVisible({ timeout: 10_000 });

  await page.getByRole("button", { name: "Preview" }).click();

  await expect(
    page.getByRole("button", { name: /Total records/ }).first(),
  ).toBeVisible({ timeout: 10_000 });
}

async function selectGrain(page: Page, grain: string) {
  const grainSelector = page.getByLabel("Select aggregation grain");
  await grainSelector.click();
  await page.getByRole("menuitem", { name: grain, exact: true }).click();
  await expect(grainSelector).toContainText(grain, { timeout: 5000 });
}

/**
 * For each diamond marker on the chart, hover at its x-position to trigger
 * the date readout, then verify the displayed date is in the expected set.
 */
async function verifyDiamondDates(
  page: Page,
  chart: ReturnType<Page["getByLabel"]>,
  expected: Set<string>,
) {
  const diamonds = chart.locator('rect[aria-label="annotation marker"]');
  const dateReadout = page
    .getByLabel("total_records primary time label")
    .first();

  // Wait for the chart to settle by retrying until the first diamond's
  // date is in the expected set (handles time-range / grain re-render).
  await expect(async () => {
    await expect(diamonds.first()).toBeVisible();
    const box = await chart.boundingBox();
    expect(box).toBeTruthy();
    const centerY = box!.y + box!.height / 2;
    const dBox = await diamonds.first().boundingBox();
    expect(dBox).toBeTruthy();
    await page.mouse.move(dBox!.x + dBox!.width / 2, centerY);
    await expect(dateReadout).toBeVisible();
    const text = (await dateReadout.textContent())?.trim();
    expect(text).toBeTruthy();
    expect([...expected]).toContain(text!);
  }).toPass({ timeout: 10_000 });

  // Chart is now in the correct state — verify all diamonds.
  const count = await diamonds.count();
  expect(count).toBeGreaterThanOrEqual(1);

  const box = await chart.boundingBox();
  if (!box) throw new Error("Chart bounding box not found");
  const centerY = box.y + box.height / 2;

  const matchedDates: string[] = [];

  for (let i = 0; i < count; i++) {
    const dBox = await diamonds.nth(i).boundingBox();
    if (!dBox) continue;

    await page.mouse.move(dBox.x + dBox.width / 2, centerY);
    await expect(dateReadout).toBeVisible({ timeout: 2000 });

    const dateText = (await dateReadout.textContent())?.trim();
    expect(dateText).toBeTruthy();
    expect(
      [...expected],
      `Diamond ${i} date "${dateText}" not in expected set`,
    ).toContain(dateText!);
    matchedDates.push(dateText!);
  }

  expect(matchedDates.length).toBeGreaterThanOrEqual(1);
  expect(matchedDates.length).toBe(count);
}

// ---------------------------------------------------------------------------
// Group 1: Rendering coverage — test all dashboard TZs with a fixed system TZ.
// This covers the actual rendering logic (different offsets shift annotations
// to different day/hour buckets).
// ---------------------------------------------------------------------------
const DASHBOARD_TIMEZONES = ["UTC", "Asia/Kolkata", "America/Los_Angeles"];

test.describe("annotations (rendering)", () => {
  test.use({ project: "AdBids", timezoneId: "UTC", locale: "en-US" });

  for (const dashboardTZ of DASHBOARD_TIMEZONES) {
    test.describe(`dashboard: ${dashboardTZ}`, () => {
      test("day-level annotations placed at correct dates", async ({
        page,
      }) => {
        await setupDashboard(page);
        await setDashboardTimezone(page, dashboardTZ);

        await interactWithTimeRangeMenu(page, async () => {
          await page.getByRole("menuitem", { name: "Last 7 days" }).click();
        });
        await selectGrain(page, "day");

        const chart = page
          .getByLabel("Measure Chart for total_records")
          .first();
        await expect(chart).toBeVisible({ timeout: 10_000 });

        const expected = expectedDates(
          DAY_ANNOTATION_TIMES,
          V1TimeGrain.TIME_GRAIN_DAY,
          dashboardTZ,
        );
        await verifyDiamondDates(page, chart, expected);
      });

      test("hour-level annotations placed at correct times", async ({
        page,
      }) => {
        await setupDashboard(page);
        await setDashboardTimezone(page, dashboardTZ);

        await interactWithTimeRangeMenu(page, async () => {
          await page.getByRole("menuitem", { name: "Last 24 hours" }).click();
        });
        await selectGrain(page, "hour");

        const chart = page
          .getByLabel("Measure Chart for total_records")
          .first();
        await expect(chart).toBeVisible({ timeout: 10_000 });

        const expected = expectedDates(
          HOUR_ANNOTATION_TIMES,
          V1TimeGrain.TIME_GRAIN_HOUR,
          dashboardTZ,
        );
        await verifyDiamondDates(page, chart, expected);
      });

      test("popover shows on hover over annotation diamond", async ({
        page,
      }) => {
        await setupDashboard(page);
        await setDashboardTimezone(page, dashboardTZ);

        await interactWithTimeRangeMenu(page, async () => {
          await page.getByRole("menuitem", { name: "Last 7 days" }).click();
        });
        await selectGrain(page, "day");

        const chart = page
          .getByLabel("Measure Chart for total_records")
          .first();
        await expect(chart).toBeVisible({ timeout: 10_000 });

        const diamonds = chart.locator('rect[aria-label="annotation marker"]');
        await expect(diamonds.first()).toBeVisible({ timeout: 10_000 });

        // Hover directly at each diamond's position near the bottom of
        // the chart until a popover appears.
        const box = await chart.boundingBox();
        if (!box) throw new Error("Chart bounding box not found");
        const hoverY = box.y + box.height - 8;

        let popoverFound = false;
        const count = await diamonds.count();

        for (let i = 0; i < count; i++) {
          const dBox = await diamonds.nth(i).boundingBox();
          if (!dBox) continue;

          await page.mouse.move(dBox.x + dBox.width / 2, hoverY);
          await page.waitForTimeout(100);

          const popover = page
            .locator('[role="menu"]')
            .filter({ hasText: /annotation/i });

          if (
            (await popover.count()) > 0 &&
            (await popover.first().isVisible())
          ) {
            popoverFound = true;
            const text = await popover.first().textContent();
            expect(text).toMatch(
              /Point annotation|Range annotation|Hour annotation/,
            );
            break;
          }
        }

        expect(popoverFound).toBe(true);
      });

      test("no diamond markers on bid_price_sum chart", async ({ page }) => {
        await setupDashboard(page);
        await setDashboardTimezone(page, dashboardTZ);

        await interactWithTimeRangeMenu(page, async () => {
          await page.getByRole("menuitem", { name: "Last 7 days" }).click();
        });
        await selectGrain(page, "day");

        const chart = page
          .getByLabel("Measure Chart for bid_price_sum")
          .first();
        await expect(chart).toBeVisible({ timeout: 10_000 });

        // Wait for the total_records chart to settle first (ensures data loaded).
        const trChart = page
          .getByLabel("Measure Chart for total_records")
          .first();
        const trDiamonds = trChart.locator(
          'rect[aria-label="annotation marker"]',
        );
        await expect(trDiamonds.first()).toBeVisible({ timeout: 10_000 });

        const diamonds = chart.locator('rect[aria-label="annotation marker"]');
        await expect(diamonds).toHaveCount(0);
      });
    });
  }
});

// ---------------------------------------------------------------------------
// Group 2: System TZ independence — prove that changing the browser's system
// timezone does not affect annotation placement. We run the day-level test
// (most sensitive to TZ shifts) with non-UTC system TZs against the LA
// dashboard TZ (largest offset from UTC).
// ---------------------------------------------------------------------------
const INDEPENDENCE_CONFIGS = [
  { systemTZ: "America/New_York", locale: "en-US" },
  { systemTZ: "Europe/Prague", locale: "de-DE" },
];

for (const sys of INDEPENDENCE_CONFIGS) {
  test.describe(`annotations system TZ independence (system: ${sys.systemTZ})`, () => {
    test.use({
      project: "AdBids",
      timezoneId: sys.systemTZ,
      locale: sys.locale,
    });

    test("day-level annotations match UTC baseline (dashboard: America/Los_Angeles)", async ({
      page,
    }) => {
      const dashboardTZ = "America/Los_Angeles";

      await setupDashboard(page);
      await setDashboardTimezone(page, dashboardTZ);

      await interactWithTimeRangeMenu(page, async () => {
        await page.getByRole("menuitem", { name: "Last 7 days" }).click();
      });
      await selectGrain(page, "day");

      const chart = page.getByLabel("Measure Chart for total_records").first();
      await expect(chart).toBeVisible({ timeout: 10_000 });

      const expected = expectedDates(
        DAY_ANNOTATION_TIMES,
        V1TimeGrain.TIME_GRAIN_DAY,
        dashboardTZ,
      );
      await verifyDiamondDates(page, chart, expected);
    });
  });
}
