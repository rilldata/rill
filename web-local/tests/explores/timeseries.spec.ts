import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import {
  interceptTimeseriesResponse,
  type TimeSeriesValue,
} from "../utils/dataSpecifcHelpers";
import { gotoNavEntry } from "../utils/waitHelpers";
import {
  interactWithTimeRangeMenu,
  setDashboardTimezone,
} from "@rilldata/web-common/tests/utils/explore-interactions";
import { DateTime } from "luxon";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { formatGrainBucket } from "@rilldata/web-common/lib/time/ranges/formatter";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";

const HOVER_STEP_PX = 5;

// The app sets Settings.defaultLocale = "en" globally.
const CHART_LOCALE = "en";

interface TimeRangeTestCase {
  menuItem: string;
  expectedDataPoints: number;
  grain: V1TimeGrain;
}

const TIME_RANGE_TEST_CASES: TimeRangeTestCase[] = [
  {
    menuItem: "Last 7 days",
    expectedDataPoints: 7,
    grain: V1TimeGrain.TIME_GRAIN_DAY,
  },
  {
    menuItem: "Last 24 hours",
    expectedDataPoints: 24,
    grain: V1TimeGrain.TIME_GRAIN_HOUR,
  },
];

/**
 * Hovers across a chart and verifies that each unique tooltip date and value
 * matches the corresponding data point from the API response.
 */
async function verifyChartTooltipData(
  page: Page,
  apiData: { data: TimeSeriesValue[] },
  grain: V1TimeGrain,
  measureName: string,
  dashboardTZ: string,
) {
  const chart = page.getByLabel(`Measure Chart for ${measureName}`).first();
  const box = await chart.boundingBox();
  if (!box) throw new Error("Chart bounding box not found");

  const centerY = box.y + box.height / 2;
  let verifiedPoints = 0;
  let lastDateText: string | undefined;

  const expectedPoints = apiData.data.length;

  for (let x = box.x; x < box.x + box.width; x += HOVER_STEP_PX) {
    await page.mouse.move(x, centerY);

    const dateLabel = page.getByLabel(/primary time label/).first();
    if (!(await dateLabel.isVisible())) continue;

    const dateText = await dateLabel.textContent();
    if (!dateText || dateText === lastDateText) continue;

    lastDateText = dateText;

    const point = apiData.data[verifiedPoints];
    const dateTime = DateTime.fromISO(point.ts, {
      zone: dashboardTZ,
      locale: CHART_LOCALE,
    });
    const pattern = formatGrainBucket(dateTime, grain);

    // Verify the date label
    await expect(dateLabel).toHaveText(pattern, { timeout: 2000 });

    // Verify the measure value
    const valueLabel = page.getByLabel("main value").first();
    await expect(valueLabel).toBeVisible({ timeout: 2000 });
    const valueText = await valueLabel.textContent();
    expect(valueText).toBeTruthy();

    const expectedValue = point.records[measureName];
    if (expectedValue !== null && expectedValue !== undefined) {
      const formatter = createMeasureValueFormatter({
        formatPreset: "humanize",
      });
      expect(valueText!.trim()).toBe(formatter(expectedValue));
    }

    verifiedPoints++;

    if (verifiedPoints >= expectedPoints) break;
  }
}

// ---------------------------------------------------------------------------
// Group 1: Rendering coverage — test dashboard TZ variations with a fixed
// system TZ. Covers the actual label formatting and value display logic.
// ---------------------------------------------------------------------------
const DASHBOARD_TIMEZONES = ["UTC", "America/Los_Angeles"];

test.describe("timeseries charts (rendering)", () => {
  test.use({ project: "AdBids", timezoneId: "UTC", locale: "en-US" });

  for (const dashboardTZ of DASHBOARD_TIMEZONES) {
    test(`chart data matches API response (dashboard: ${dashboardTZ})`, async ({
      page,
    }) => {
      await page.getByLabel("/dashboards").click();
      await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

      await expect(
        page.getByRole("button", { name: /Total records/ }).first(),
      ).toBeVisible({ timeout: 5000 });

      await page.getByRole("button", { name: "Preview" }).click();

      await expect(
        page.getByRole("button", { name: /Total records/ }).first(),
      ).toBeVisible({ timeout: 5000 });

      await setDashboardTimezone(page, dashboardTZ);

      for (const testCase of TIME_RANGE_TEST_CASES) {
        await interactWithTimeRangeMenu(page, async () => {
          await page.getByRole("menuitem", { name: testCase.menuItem }).click();
        });

        await page
          .getByRole("button", {
            name: "Select aggregation grain",
          })
          .click();

        const timeseriesPromise = interceptTimeseriesResponse(page);
        await page
          .getByRole("menuitem", {
            name: V1TimeGrainToDateTimeUnit[testCase.grain],
            exact: true,
          })
          .click();

        // Wait for chart to update with new data
        await page.waitForTimeout(500);

        const apiData = await timeseriesPromise;
        expect(apiData.data.length).toBe(testCase.expectedDataPoints);

        await verifyChartTooltipData(
          page,
          apiData,
          testCase.grain,
          "total_records",
          dashboardTZ,
        );
      }
    });
  }
});

// ---------------------------------------------------------------------------
// Group 2: System TZ independence — prove that a non-UTC system timezone
// does not affect chart rendering. Uses Europe/Prague (de-DE locale) as the
// system TZ with LA dashboard TZ to maximize the mismatch.
// ---------------------------------------------------------------------------
test.describe("timeseries charts system TZ independence", () => {
  test.use({ project: "AdBids", timezoneId: "Europe/Prague", locale: "de-DE" });

  test("chart data matches UTC baseline (dashboard: America/Los_Angeles)", async ({
    page,
  }) => {
    const dashboardTZ = "America/Los_Angeles";

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

    await expect(
      page.getByRole("button", { name: /Total records/ }).first(),
    ).toBeVisible({ timeout: 5000 });

    await page.getByRole("button", { name: "Preview" }).click();

    await expect(
      page.getByRole("button", { name: /Total records/ }).first(),
    ).toBeVisible({ timeout: 5000 });

    await setDashboardTimezone(page, dashboardTZ);

    for (const testCase of TIME_RANGE_TEST_CASES) {
      await interactWithTimeRangeMenu(page, async () => {
        await page.getByRole("menuitem", { name: testCase.menuItem }).click();
      });

      await page
        .getByRole("button", {
          name: "Select aggregation grain",
        })
        .click();

      const timeseriesPromise = interceptTimeseriesResponse(page);
      await page
        .getByRole("menuitem", {
          name: V1TimeGrainToDateTimeUnit[testCase.grain],
          exact: true,
        })
        .click();

      await page.waitForTimeout(500);

      const apiData = await timeseriesPromise;
      expect(apiData.data.length).toBe(testCase.expectedDataPoints);

      await verifyChartTooltipData(
        page,
        apiData,
        testCase.grain,
        "total_records",
        dashboardTZ,
      );
    }
  });
});
