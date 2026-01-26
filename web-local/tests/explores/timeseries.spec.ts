import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import {
  interceptTimeseriesResponse,
  type TimeSeriesValue,
} from "../utils/dataSpecifcHelpers";
import { gotoNavEntry } from "../utils/waitHelpers";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { DateTime } from "luxon";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { formatDateTimeByGrain } from "@rilldata/web-common/lib/time/ranges/formatter";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";

const HOVER_STEP_PX = 5;

interface TimeRangeTestCase {
  menuItem: string;
  expectedDataPoints: number;
  grain: V1TimeGrain;
}

const TIME_RANGE_TEST_CASES: TimeRangeTestCase[] = [
  {
    menuItem: "Last 7 days",
    expectedDataPoints: 9,
    grain: V1TimeGrain.TIME_GRAIN_DAY,
  },
  {
    menuItem: "Last 24 hours",
    expectedDataPoints: 26,
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
) {
  const chart = page.getByLabel(`Measure Chart for ${measureName}`).first();
  const box = await chart.boundingBox();
  if (!box) throw new Error("Chart bounding box not found");

  const centerY = box.y + box.height / 2;
  let verifiedPoints = 0;
  let lastDateText: string | undefined;
  // Exclude first and last data points as they're not rendered
  const expectedPoints = apiData.data.length - 2;

  for (let x = box.x; x < box.x + box.width; x += HOVER_STEP_PX) {
    await page.mouse.move(x, centerY);

    const dateLabel = page.getByLabel(/primary time label/).first();
    if (!(await dateLabel.isVisible())) continue;

    const dateText = await dateLabel.textContent();
    if (!dateText || dateText === lastDateText) continue;

    lastDateText = dateText;
    // Skip the first data point (index 0) as chart starts from second point
    const point = apiData.data[verifiedPoints + 1];
    const dateTime = DateTime.fromISO(point.ts, { zone: "UTC" });
    const pattern = formatDateTimeByGrain(dateTime, grain);

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

const TIMEZONES = ["UTC", "Europe/Prague", "Asia/Kolkata"] as const;

for (const timezone of TIMEZONES) {
  test.describe(`timeseries charts (${timezone})`, () => {
    test.use({ project: "AdBids", timezoneId: timezone });

    test("chart data matches API response", async ({ page }) => {
      await page.getByLabel("/dashboards").click();
      await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

      await expect(
        page.getByRole("button", { name: /Total records/ }).first(),
      ).toBeVisible({ timeout: 5000 });

      await page.getByRole("button", { name: "Preview" }).click();

      await expect(
        page.getByRole("button", { name: /Total records/ }).first(),
      ).toBeVisible({ timeout: 5000 });

      for (const testCase of TIME_RANGE_TEST_CASES) {
        const timeseriesPromise = interceptTimeseriesResponse(page);

        await interactWithTimeRangeMenu(page, async () => {
          await page.getByRole("menuitem", { name: testCase.menuItem }).click();
        });

        // Wait for chart to update with new data
        await page.waitForTimeout(500);

        const apiData = await timeseriesPromise;
        expect(apiData.data.length).toBe(testCase.expectedDataPoints);

        await verifyChartTooltipData(
          page,
          apiData,
          testCase.grain,
          "total_records",
        );
      }
    });
  });
}
