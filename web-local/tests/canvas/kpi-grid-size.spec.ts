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

// The same two KPIs in a small grid and in a grid of the default size.
const TWO_GRIDS_CANVAS_YAML = `type: canvas
display_name: "KPI size canvas"
defaults:
  time_range: PT24H
  comparison_mode: time
rows:
  - items:
      - kpi_grid:
          metrics_view: AdBids_metrics
          measures:
            - total_records
            - bid_price_sum
          sparkline: none
          size: small
        width: 12
    height: 128px
  - items:
      - kpi_grid:
          metrics_view: AdBids_metrics
          measures:
            - total_records
            - bid_price_sum
          sparkline: none
        width: 12
    height: 128px
`;

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

// Width in pixels of the widget holding a KPI card:
// the query container that decides whether the grid stacks its cards.
function widgetWidth(card: Locator) {
  return card.evaluate((element) => {
    let node: Element | null = element;
    while (
      node &&
      getComputedStyle(node).containerName !== "component-container"
    ) {
      node = node.parentElement;
    }
    if (!node) throw new Error("KPI card is not inside a canvas widget");
    return node.getBoundingClientRect().width;
  });
}

// The number of columns a grid's cards are laid out in.
function columnCount(cards: Locator) {
  return cards.evaluateAll(
    (elements) =>
      new Set(
        elements.map((element) =>
          Math.round(element.getBoundingClientRect().left),
        ),
      ).size,
  );
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

  test("a small grid keeps its cards side by side in a narrow widget", async ({
    page,
  }) => {
    await gotoNavEntry(page, CANVAS_FILE);
    await setCanvasYaml(page, TWO_GRIDS_CANVAS_YAML);
    await page.getByRole("button", { name: "Preview" }).click();
    // On a phone-width viewport the canvas stays at its minimum width,
    // which is wide enough for two cards of either size
    // but under the 440px at which a grid of any other size stacks its cards.
    await page.setViewportSize({ width: 400, height: 900 });

    const smallCards = page.locator('.wrapper[data-size="small"]');
    const mediumCards = page.locator('.wrapper[data-size="medium"]');
    await expect(smallCards).toHaveCount(2, { timeout: 60_000 });
    await expect(mediumCards).toHaveCount(2);
    expect(await widgetWidth(smallCards.first())).toBeLessThan(440);

    await expect.poll(() => columnCount(smallCards)).toBe(2);
    await expect.poll(() => columnCount(mediumCards)).toBe(1);
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
