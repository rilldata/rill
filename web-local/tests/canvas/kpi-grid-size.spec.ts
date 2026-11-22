import { expect, type Locator, type Page } from "@playwright/test";
import { updateCodeEditor } from "web-local/tests/utils/commonHelpers";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { validateYamlContents } from "web-local/tests/utils/yamlHelpers";
import { test } from "../setup/base";

const CANVAS_FILE = "/dashboards/AdBids_metrics_canvas.yaml";

// The fixture's metrics view plus a measure with a long formatted value:
// the canvas's 24 hours hold 1,122 records, so it shows "1,246,542".
const METRICS_YAML = `version: 1
type: metrics_view

display_name: Adbids
table: AdBids_model
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
    type: time
  - name: offset_timestamp
    display_name: Offset Timestamp
    column: offset_timestamp
    type: time

measures:
  - name: total_records
    display_name: Total records
    expression: COUNT(*)
    format_preset: humanize
  - name: bid_price_sum
    display_name: Sum of Bid Price
    expression: SUM(bid_price)
    format_preset: humanize
  - name: scaled_records
    display_name: Scaled records
    expression: COUNT(*) * 1111
    format_d3: ",.0f"
`;

function canvasYaml(size: string) {
  return `type: canvas
display_name: "KPI size canvas"
defaults:
  time_range: PT24H
  comparison_mode: time
rows:
  - items:
      - kpi_grid:
          metrics_view: AdBids_metrics
          measures:
            - scaled_records
          comparison:
            - delta
            - percent_change
          sparkline: right
          size: ${size}
        width: 12
    height: 128px
`;
}

// Replaces the canvas YAML through the code editor and returns to the visual editor.
async function setCanvasYaml(page: Page, yaml: string) {
  await page.getByRole("button", { name: "Switch to code editor" }).click();
  await updateCodeEditor(page, yaml);
  await page.getByRole("button", { name: "Switch to visual editor" }).click();
}

// How far the value runs past either edge of its column, in pixels; the
// column hides its overflow, so anything above zero is clipped.
function valueOverflow(data: Locator) {
  return data.evaluate((column) => {
    const value = column.querySelector(".big-number span");
    if (!value) throw new Error("KPI value not rendered");
    const columnBox = column.getBoundingClientRect();
    const valueBox = value.getBoundingClientRect();
    return Math.max(
      valueBox.right - columnBox.right,
      columnBox.left - valueBox.left,
    );
  });
}

test.describe("canvas kpi grid size", () => {
  test.use({ project: "AdBids" });
  // Each test starts its own runtime and ingests the project before the
  // canvas renders, which can take longer than the default 30s locally.
  test.describe.configure({ timeout: 180_000 });

  test("a larger size keeps a long value whole beside the sparkline", async ({
    page,
  }) => {
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");
    await page.getByRole("button", { name: "switch to code editor" }).click();
    const saved = page.waitForResponse(
      "**/rill.runtime.v1.RuntimeService/PutFile",
    );
    await updateCodeEditor(page, METRICS_YAML);
    await saved;

    await gotoNavEntry(page, CANVAS_FILE);
    const data = page.getByRole("button", { name: "scaled_records KPI data" });
    const card = page.locator(".wrapper[data-size]", { has: data });

    for (const [size, fontSize] of [
      ["medium", "30px"],
      ["large", "44px"],
      ["xlarge", "64px"],
    ]) {
      await setCanvasYaml(page, canvasYaml(size));
      await expect(card).toHaveAttribute("data-size", size, {
        timeout: 60_000,
      });
      await expect(data.locator(".big-number")).toHaveText("1,246,542", {
        timeout: 60_000,
      });
      // The size reaches the card through the grid and the provider.
      await expect(data.locator(".big-number")).toHaveCSS(
        "font-size",
        fontSize,
      );
      await expect(card.locator(".sparkline-wrapper")).toBeVisible();
      expect(await valueOverflow(data), `value clipped at ${size}`).toBe(0);
    }
  });

  test("the Size switcher writes larger sizes and drops the key for medium", async ({
    page,
  }) => {
    await gotoNavEntry(page, CANVAS_FILE);
    const value = page
      .getByRole("button", { name: "total_records KPI data" })
      .locator(".big-number");
    await expect(value).toHaveCSS("font-size", "30px", { timeout: 60_000 });

    await page.locator("#AdBids_metrics_canvas--component-0-0").click();
    const sizeSwitcher = page
      .locator(".component-param")
      .filter({ has: page.getByText("Size", { exact: true }) });

    await sizeSwitcher.getByRole("button", { name: "L", exact: true }).click();
    await expect(value).toHaveCSS("font-size", "44px");
    await page.getByRole("button", { name: "Switch to code editor" }).click();
    await validateYamlContents(page, ["size: large"]);

    await page.getByRole("button", { name: "Switch to visual editor" }).click();
    await page.locator("#AdBids_metrics_canvas--component-0-0").click();
    await sizeSwitcher.getByRole("button", { name: "M", exact: true }).click();
    await expect(value).toHaveCSS("font-size", "30px");
    await page.getByRole("button", { name: "Switch to code editor" }).click();
    await validateYamlContents(
      page,
      [
        "metrics_view: AdBids_metrics",
        "- total_records",
        "- bid_price_sum",
        "- percent_change",
      ],
      ["size:"],
    );
  });
});
