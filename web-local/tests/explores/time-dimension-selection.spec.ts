import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { ResourceWatcher } from "../utils/ResourceWatcher";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";
import type { Page } from "@playwright/test";

async function waitForDashboard(page: Page) {
  await expect(page.getByLabel("Select time range")).toBeVisible({
    timeout: 10000,
  });
  await page.waitForTimeout(1000);
}

function getMetricsViewYaml(includeSecondTimeDimension = false) {
  const secondTimeDim = includeSecondTimeDimension
    ? `
  - name: timestamp_offset
    display_name: Timestamp Offset
    expression: "timestamp - INTERVAL '7 days'"
    type: time`
    : "";

  return `
type: metrics_view
display_name: Adbids
table: AdBids
timeseries: timestamp

dimensions:
  - name: publisher
    display_name: Publisher
    column: publisher
  - name: domain
    display_name: Domain
    column: domain
  - name: timestamp
    display_name: Timestamp
    column: timestamp
    type: time${secondTimeDim}

measures:
  - name: total_records
    display_name: Total records
    expression: COUNT(*)
    format_preset: humanize
  - name: bid_price_sum
    display_name: Sum of Bid Price
    expression: SUM(bid_price)
    format_preset: humanize
`;
}

test.describe("time dimension selection", () => {
  test.use({ project: "AdBids" });

  test("time dimensions appear when configured", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    await page.getByLabel("/metrics").click();
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");
    await page.getByRole("button", { name: "switch to code editor" }).click();

    await watcher.updateAndWaitForDashboard(
      getMetricsViewYaml(),
      "AdBids_metrics",
    );

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();

    await waitForDashboard(page);

    await page.getByLabel("Select time range").click();
    await page.waitForTimeout(500);

    const timeZoneButton = page.getByRole("button", { name: /Time zone/ });
    await expect(timeZoneButton).toBeVisible({ timeout: 5000 });

    await page.keyboard.press("Escape");
    await expect(page.getByLabel("Select time range")).toBeVisible();
  });

  test("URL parameters are preserved for time range", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();

    await waitForDashboard(page);

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 7 days" }).click();
    });

    await page.waitForTimeout(1000);

    await expect(page.getByText("Last 7 Days")).toBeVisible();

    const url = new URL(page.url());
    expect(url.searchParams.has("tr") || url.pathname.includes("explore")).toBe(
      true,
    );

    await expect(page.getByLabel("Select time range")).toBeVisible();
  });

  test("dashboard displays correctly with time range selection", async ({
    page,
  }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();

    await waitForDashboard(page);

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 7 days" }).click();
    });

    await expect(
      page.getByRole("menuitem", { name: "Last 7 days" }),
    ).not.toBeVisible();

    await expect(page.getByText("Last 7 Days")).toBeVisible();
    await expect(page.getByText("Publisher")).toBeVisible();
  });

  test("leaderboard data displays correctly", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();

    await waitForDashboard(page);

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });
    await page.waitForTimeout(1000);

    await expect(page.getByText("Publisher")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Facebook", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Google", exact: true }),
    ).toBeVisible();
  });

  test("adding time dimension in metrics view updates dashboard", async ({
    page,
  }) => {
    const watcher = new ResourceWatcher(page);

    await page.getByLabel("/metrics").click();
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");
    await page.getByRole("button", { name: "switch to code editor" }).click();

    await watcher.updateAndWaitForDashboard(
      getMetricsViewYaml(true),
      "AdBids_metrics",
    );

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();
    await waitForDashboard(page);

    await expect(page.getByText("Publisher")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Facebook", exact: true }),
    ).toBeVisible();
  });
});
