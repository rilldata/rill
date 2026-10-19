import { expect, type Page } from "@playwright/test";
import { updateCodeEditor } from "web-local/tests/utils/commonHelpers";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { validateYamlContents } from "web-local/tests/utils/yamlHelpers";
import { test } from "../setup/base";

const CANVAS_FILE = "/dashboards/AdBids_metrics_canvas.yaml";

// Replaces the canvas YAML through the code editor and returns to the visual editor.
async function setCanvasYaml(page: Page, yaml: string) {
  await page.getByRole("button", { name: "Switch to code editor" }).click();
  await updateCodeEditor(page, yaml);
  await page.getByRole("button", { name: "Switch to visual editor" }).click();
}

// `compare` turns on the time comparison, which adds a Δ and a Δ% column
// after every measure of a table.
function canvasYaml(
  component: string,
  { width = 12, compare = false }: { width?: number; compare?: boolean } = {},
) {
  return `type: canvas
display_name: "Table canvas"
defaults:
  time_range: PT24H${compare ? "\n  comparison_mode: time" : ""}
rows:
  - items:
${component
  .split("\n")
  .map((line) => `      ${line}`)
  .join("\n")}
        width: ${width}
    height: 400px
`;
}

// Data rows of the flat table: the spacer rows have no cells and the totals
// row is labelled "Total".
function dataRows(page: Page) {
  return page
    .locator("table tbody tr")
    .filter({ has: page.locator("td") })
    .filter({ hasNotText: "Total" });
}

const horizontalOverflow = (page: Page) =>
  page
    .locator(".table-wrapper")
    .evaluate((el) => el.scrollWidth - el.clientWidth);

test.describe("canvas table column config", () => {
  test.use({ project: "AdBids" });
  // Each test starts its own runtime and ingests the project before the
  // canvas renders, which can take longer than the default 30s locally.
  test.describe.configure({ timeout: 180_000 });

  test("renders configured widths, labels, formats, wrapping and sort", async ({
    page,
  }) => {
    await gotoNavEntry(page, CANVAS_FILE);
    await setCanvasYaml(
      page,
      canvasYaml(
        `- table:
    metrics_view: AdBids_metrics
    title: Table component
    wrap_headers: true
    sort_by: total_records
    sort_dir: desc
    columns:
      - name: publisher
        width: 250
        label: Publisher name
        wrap: true
      - domain
      - name: total_records
        format_d3: ",.2f"
        align: center
        delta:
          width: 90
      - bid_price_sum`,
        { compare: true },
      ),
    );

    // Label override in the table header (the first render waits for the
    // project to reconcile). Scoped to the header so the wait cannot be
    // satisfied by the same text elsewhere while the previous table is up.
    await expect(page.locator("thead").getByText("Publisher name")).toBeVisible(
      { timeout: 60_000 },
    );
    // Configured width
    await expect(page.locator("table colgroup col").first()).toHaveCSS(
      "width",
      "250px",
    );
    // With the comparison on, total_records is followed by its Δ column,
    // which takes the width from `delta`.
    await expect(page.locator("table colgroup col").nth(3)).toHaveCSS(
      "width",
      "90px",
    );
    // Number format override on the totals row (1,122 records in the fixture)
    await expect(page.getByText("1,122.00")).toBeVisible();
    // Wrapped cells and headers keep a uniform, taller height (2 lines)
    await expect(dataRows(page).first().locator("td").first()).toHaveCSS(
      "height",
      "40px",
    );
    await expect(page.locator("thead button.header-cell").first()).toHaveCSS(
      "height",
      "46px",
    );
    // Initial sort: descending on total_records (third column)
    const first = await dataRows(page).first().locator("td").nth(2).innerText();
    const second = await dataRows(page).nth(1).locator("td").nth(2).innerText();
    expect(parseFloat(first.replace(/,/g, ""))).toBeGreaterThanOrEqual(
      parseFloat(second.replace(/,/g, "")),
    );
  });

  test("fit_to_width removes the horizontal overflow", async ({ page }) => {
    await page.setViewportSize({ width: 1800, height: 900 });
    await gotoNavEntry(page, CANVAS_FILE);

    // A third-width item next to a markdown block is narrower than the
    // estimated column widths (a lone item would stretch to the full row).
    const table = (fit: boolean) => `type: canvas
display_name: "Table canvas"
defaults:
  time_range: PT24H
rows:
  - items:
      - markdown:
          content: Filler
        width: 8
      - table:
          metrics_view: AdBids_metrics
          fit_to_width: ${fit}
          time_filters: tr=PT24H
          columns:
            - publisher
            - domain
            - total_records
            - bid_price_sum
        width: 4
    height: 400px
`;

    await setCanvasYaml(page, table(false));
    await expect(page.locator("table colgroup col")).toHaveCount(4, {
      timeout: 60_000,
    });
    await expect.poll(() => horizontalOverflow(page)).toBeGreaterThan(0);

    await setCanvasYaml(page, table(true));
    await expect(page.locator("table colgroup col")).toHaveCount(4, {
      timeout: 15_000,
    });
    await expect.poll(() => horizontalOverflow(page)).toBeLessThanOrEqual(0);
  });

  test("pivot measure widths apply under every column group", async ({
    page,
  }) => {
    await gotoNavEntry(page, CANVAS_FILE);
    await setCanvasYaml(
      page,
      canvasYaml(`- pivot:
    metrics_view: AdBids_metrics
    sort_by: total_records
    time_filters: tr=PT24H
    row_dimensions:
      - name: publisher
        width: 220
    col_dimensions:
      - domain
    measures:
      - name: total_records
        width: 120`),
    );

    const cols = page.locator("table colgroup col");
    // Row header, then one measure column per domain value (plus totals). The
    // widget's own time filter without a comparison keeps the delta columns off.
    await expect(cols.first()).toHaveCSS("width", "220px", {
      timeout: 60_000,
    });
    await expect(cols.nth(1)).toHaveCSS("width", "120px");
    await expect(cols.nth(2)).toHaveCSS("width", "120px");
  });

  test("the editor persists column drags, header sorts and the fit switch", async ({
    page,
  }) => {
    await gotoNavEntry(page, CANVAS_FILE);
    // The first render waits for the project to ingest and reconcile.
    await expect(page.getByText("Table component")).toBeVisible({
      timeout: 60_000,
    });
    // Components load lazily once scrolled into view; the table is the last row.
    const tableCard = page.locator("#AdBids_metrics_canvas--component-3-0");
    await tableCard.scrollIntoViewIfNeeded();
    await expect(dataRows(page).first()).toBeVisible({ timeout: 90_000 });

    // Drag the first column's edge 80px to the right.
    // The handle spans the table's full virtual height, most of which is
    // clipped by the scroll container, so press near its top.
    const handle = page.locator(".table-wrapper button.EW").first();
    const box = await handle.boundingBox();
    if (!box) throw new Error("column resize handle not found");
    const x = box.x + box.width / 2;
    const y = box.y + 30;
    await page.mouse.move(x, y);
    await page.mouse.down();
    await page.mouse.move(x + 40, y, { steps: 4 });
    await page.mouse.move(x + 80, y, { steps: 4 });
    await page.mouse.up();

    // Drag the edge of the Δ column of total_records (publisher, domain,
    // total_records, Δ, ...) 30px; its width is stored on the measure's `delta`.
    const deltaHandle = page.locator(".table-wrapper button.EW").nth(3);
    const deltaBox = await deltaHandle.boundingBox();
    if (!deltaBox) throw new Error("delta column resize handle not found");
    const dx = deltaBox.x + deltaBox.width / 2;
    const dy = deltaBox.y + 30;
    await page.mouse.move(dx, dy);
    await page.mouse.down();
    await page.mouse.move(dx + 15, dy, { steps: 3 });
    await page.mouse.move(dx + 30, dy, { steps: 3 });
    await page.mouse.up();

    // Clicking a header sorts and persists the sort.
    await page
      .locator("thead")
      .getByRole("button", { name: "Domain", exact: true })
      .click();
    // The fixture compares with the previous period, so the measure columns
    // carry delta headers: publisher, domain, total_records, Δ, Δ%, ...
    // Clicking the Δ header persists a comparison sort.
    await page.locator("thead th").nth(3).locator("button.header-cell").click();

    // The fit switch in the sidebar of the selected table component.
    await tableCard.click();
    await page
      .locator(".component-param", { hasText: "Fit columns to width" })
      .getByRole("switch")
      .click();

    await page.getByRole("button", { name: "Switch to code editor" }).click();
    // CodeMirror only renders the visible lines and the table block is at the
    // end of the file, so scroll the editor to the end before reading its text.
    const editor = page.getByRole("textbox", { name: "codemirror editor" });
    await editor.waitFor({ state: "visible" });
    await page
      .locator(".cm-scroller")
      .first()
      .evaluate((el) => el.scrollTo({ top: el.scrollHeight }));
    await validateYamlContents(page, [
      "- name: publisher",
      "width:",
      "- name: total_records",
      "delta:",
      "sort_by: total_records",
      "sort_comparison: delta",
      "sort_dir:",
      "fit_to_width: true",
    ]);
  });
});
