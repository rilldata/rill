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
});
