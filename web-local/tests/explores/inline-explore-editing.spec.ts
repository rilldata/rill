import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { updateCodeEditor } from "../utils/commonHelpers";
import {
  assertAdBidsDashboard,
  createAdBidsModel,
} from "../utils/dataSpecifcHelpers";
import { createExploreFromModel } from "../utils/exploreHelpers";
import { waitForReconciliation } from "../utils/wait-for-reconciliation";
import { gotoNavEntry } from "../utils/waitHelpers";

test.describe("inline explore editing", () => {
  test.use({ project: "Blank" });

  test("edit the inline explore from the metrics view workspace", async ({
    page,
  }) => {
    test.setTimeout(60_000);

    // Generate a metrics view, which defines its explore dashboard inline
    await createAdBidsModel(page);
    await createExploreFromModel(page);

    // Switch to the explore view via the three-way editor toggle
    await page
      .getByRole("button", { name: "switch to explore editor" })
      .click();

    // The visual explore editor renders next to the live dashboard preview;
    // give the explore resource time to reconcile on slow machines.
    await expect(page.getByText("Edit dashboard")).toBeVisible({
      timeout: 15_000,
    });
    await assertAdBidsDashboard(page);

    // Restrict measures and dimensions to a subset via the sidebar
    await page.getByRole("button", { name: "Subset" }).first().click();
    await page.getByRole("button", { name: "Subset" }).nth(1).click();

    // The edits land under the `explore:` key of the metrics view file
    await page.getByRole("button", { name: "switch to code editor" }).click();
    const editor = page.getByRole("textbox", { name: "codemirror editor" });
    const editorText = () => editor.textContent();
    await expect.poll(editorText, { timeout: 10_000 }).toContain("explore:");
    await expect.poll(editorText).toContain("type: metrics_view");

    // The explore block sits at the end of the file, which can fall outside
    // CodeMirror's virtualized viewport; page down to the end to render it.
    await editor.click();
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press("PageDown");
    }
    await expect.poll(editorText, { timeout: 10_000 }).toContain("- publisher");
    await expect
      .poll(editorText, { timeout: 10_000 })
      .toContain("- total_records");
  });

  test("header CTA offers canvas generation", async ({ page }) => {
    test.setTimeout(60_000);

    await createAdBidsModel(page);
    await createExploreFromModel(page);

    // The inline explore is reachable via the Preview button and the editor toggle,
    // so the CTA only deals in canvases. With none in the project yet, it offers
    // canvas generation, but neither "Go to dashboard" nor another explore.
    await expect(
      page.getByRole("button", { name: "Generate Canvas Dashboard" }),
    ).toBeVisible({ timeout: 10_000 });
    await expect(
      page.getByRole("button", { name: "Create Explore Dashboard" }),
    ).toHaveCount(0);
    await expect(page.getByRole("link", { name: /Go to/ })).toHaveCount(0);
  });
});

test.describe("inline explore go-to-dashboard CTA", () => {
  test.use({ project: "AdBids" });

  test("lists the canvas dashboards of the metrics view", async ({ page }) => {
    test.setTimeout(60_000);

    // Define the explore inline in the existing metrics view, which the project's
    // canvas dashboard is built on. The nav tree renders expanded, so the entry
    // is directly reachable.
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");
    await page.getByRole("button", { name: "switch to code editor" }).click();
    await updateCodeEditor(page, METRICS_VIEW_WITH_INLINE_EXPLORE);

    // The dropdown lists canvases by their validSpec display name, which is only
    // available once the canvas has reconciled against the rewritten metrics view.
    await waitForReconciliation(page);

    // The header CTA links to the metrics view's canvas dashboards
    await expect(
      page.getByRole("link", { name: "Go to canvas dashboard" }),
    ).toBeVisible({ timeout: 10_000 });
    await page.getByRole("button", { name: "Create resource menu" }).click();
    await expect(
      page.getByRole("menuitem", { name: "Adbids Canvas Dashboard" }),
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Generate Canvas Dashboard" }),
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Create Explore Dashboard" }),
    ).toHaveCount(0);

    // The project's standalone explore is not listed; the CTA only deals in canvases
    await expect(
      page.getByRole("menuitem", { name: "Adbids dashboard" }),
    ).toHaveCount(0);

    await page
      .getByRole("menuitem", { name: "Adbids Canvas Dashboard" })
      .click();
    await page.waitForURL("**/files/dashboards/AdBids_metrics_canvas.yaml");
  });
});

const METRICS_VIEW_WITH_INLINE_EXPLORE = `version: 1
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

explore:
  display_name: Adbids Inline Explore
`;
