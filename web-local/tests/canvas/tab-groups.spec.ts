import { expect } from "@playwright/test";
import { test } from "../setup/base";
import {
  addFromFloatingMenu,
  blockRegion,
  dragTo,
  expectProjectFile,
  putProjectFile,
} from "../utils/canvasBlockHelpers";
import { waitForReconciliation } from "../utils/wait-for-reconciliation";
import { gotoNavEntry } from "../utils/waitHelpers";

const CANVAS = "/dashboards/blocks_canvas.yaml";

// Two free rows, so blocks can be added after them and dragged above them.
const CANVAS_YAML = `type: canvas
display_name: Blocks canvas
rows:
  - items:
      - markdown:
          content: First row
        width: 12
    height: 40px
  - items:
      - kpi_grid:
          metrics_view: AdBids_metrics
          measures:
            - total_records
        width: 12
    height: 128px
`;

const PIVOT_TAB_CANVAS = "/dashboards/pivot_tab_canvas.yaml";

// The pivot sits in the second tab, behind a widget on the same metrics view.
const PIVOT_TAB_CANVAS_YAML = `type: canvas
display_name: Pivot tab canvas
rows:
  - name: views
    tabs:
      - label: Summary
        rows:
          - items:
              - kpi_grid:
                  metrics_view: AdBids_metrics
                  measures:
                    - total_records
                width: 12
            height: 128px
      - label: Breakdown
        rows:
          - items:
              - pivot:
                  metrics_view: AdBids_metrics
                  row_dimensions:
                    - publisher
                  col_dimensions:
                    - domain
                  measures:
                    - total_records
                width: 12
            height: 400px
`;

test.describe("canvas tab groups", () => {
  test.use({ project: "AdBids" });

  test("adds, renames, duplicates, moves and deletes a tab group", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 1600 });
    await putProjectFile(page, CANVAS, CANVAS_YAML);
    await waitForReconciliation(page);
    await gotoNavEntry(page, CANVAS);
    await expect(page.getByText("First row")).toBeVisible();

    // Add a tab group at the end, then a second tab. The editor names every block it adds.
    await addFromFloatingMenu(page, "Tab group");
    await expectProjectFile(page, CANVAS, [
      "  - name: group-1\n    tabs:",
      "label: Tab 1",
    ]);
    const group = blockRegion(page, 2);
    await group.getByTitle("Add tab").click();
    await expectProjectFile(page, CANVAS, ["label: Tab 1", "label: Tab 2"]);

    // Select the group by clicking a tab, and rename it in the inspector.
    await group.getByRole("tab", { name: "Tab 1" }).click();
    const name = page.getByLabel("Tab group name");
    await name.fill("deep_dive");
    await name.press("Enter");
    await expectProjectFile(page, CANVAS, ["  - name: deep_dive", "tabs:"]);

    // Duplicate it: the copy lands right after it, under a new name.
    await page.getByRole("button", { name: "Duplicate tab group" }).click();
    await expectProjectFile(page, CANVAS, [
      "  - name: deep_dive",
      "  - name: group-1\n    tabs:",
    ]);

    // Drag the copy above the first row by its grip.
    await dragTo(
      page,
      blockRegion(page, 3).locator(".grip"),
      page.locator("#canvas-row-0"),
    );
    await expectProjectFile(page, CANVAS, [
      "rows:\n  - name: group-1",
      "content: First row",
      "  - name: deep_dive",
    ]);

    // Delete the copy from the inspector; its tabs are empty, so nothing asks to confirm.
    await blockRegion(page, 0).getByRole("tab", { name: "Tab 1" }).click();
    await page.getByRole("button", { name: "Delete tab group" }).click();
    await expectProjectFile(
      page,
      CANVAS,
      ["rows:\n  - items:", "content: First row", "  - name: deep_dive"],
      ["name: group-1"],
    );
  });

  // A pivot that mounts once the canvas's queries are cached, as in a tab opened later,
  // used to link to an empty Explore pivot, and a viewer's sort was dropped.
  test("links a pivot in a later tab to Explore with its rows, columns and sort", async ({
    page,
  }) => {
    await putProjectFile(page, PIVOT_TAB_CANVAS, PIVOT_TAB_CANVAS_YAML);
    await waitForReconciliation(page);
    await gotoNavEntry(page, PIVOT_TAB_CANVAS);
    await page.getByRole("button", { name: "Preview" }).click();

    await page.getByRole("tab", { name: "Breakdown" }).click();
    const pivot = page.locator("article.component-card", {
      has: page.locator("table"),
    });
    const publisherHeader = pivot
      .locator("thead button.header-cell")
      .filter({ hasText: "Publisher" });
    await expect(publisherHeader).toBeVisible({ timeout: 60_000 });
    await publisherHeader.click();

    await pivot.hover();
    await pivot.getByRole("button", { name: /^Go to / }).click();
    await page.waitForURL(/\/explore\//, { waitUntil: "commit" });
    const params = new URL(page.url()).searchParams;
    expect(params.get("view")).toBe("pivot");
    expect(params.get("rows")).toBe("publisher");
    expect(params.get("cols")).toBe("domain,total_records");
    expect(params.get("sort_by")).toBe("publisher");
  });
});
