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

const CANVAS = "/dashboards/story_canvas.yaml";
const SHIFT_CANVAS = "/dashboards/shift_canvas.yaml";

// A free header row, a container holding one row, and a loose free row.
const CANVAS_YAML = `type: canvas
display_name: Story canvas
rows:
  - items:
      - markdown:
          content: Header row
        width: 12
    height: 40px
  - name: story
    rows:
      - items:
          - kpi_grid:
              metrics_view: AdBids_metrics
              measures:
                - total_records
            width: 12
        height: 128px
  - items:
      - markdown:
          content: Loose row
        width: 12
    height: 40px
`;

// A free row, an unnamed tab group with two tabs, and an unnamed container.
const SHIFT_YAML = `type: canvas
display_name: Shift canvas
rows:
  - items:
      - markdown:
          content: First row
        width: 12
    height: 40px
  - tabs:
      - label: Overview
        rows: []
      - label: Details
        rows: []
  - rows:
      - items:
          - markdown:
              content: Story row
            width: 12
        height: 40px
`;

// The editor's container affordances are behind the canvas_containers flag.
const RILL_YAML = `compiler: rillv1
display_name: Containers e2e
features:
  chat: false
  canvas_containers: true
`;

test.describe("canvas containers", () => {
  test.use({ project: "AdBids" });

  test("renders a container's rows as plain rows outside the editor", async ({
    page,
  }) => {
    await putProjectFile(page, CANVAS, CANVAS_YAML);
    await waitForReconciliation(page);
    await gotoNavEntry(page, CANVAS);

    // The editor frames the container and names it in a chip.
    const chip = blockRegion(page, 1).locator(".chip");
    await expect(chip).toContainText("Container");
    await expect(chip).toContainText("story");

    // Viewers get its rows like free rows, without the frame or the chip.
    await page.getByRole("button", { name: "Preview" }).click();
    await expect(page.locator("#canvas-row-story-0")).toBeVisible();
    await expect(page.locator(".chip")).toHaveCount(0);
    await expect(page.locator(".block-region")).toHaveCount(0);
  });

  test("groups rows into containers and moves, conditions, ungroups and deletes them", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 1600 });
    await putProjectFile(page, "/rill.yaml", RILL_YAML);
    await putProjectFile(page, CANVAS, CANVAS_YAML);
    await waitForReconciliation(page);
    await page.reload();
    await gotoNavEntry(page, CANVAS);
    await expect(page.getByText("Header row")).toBeVisible();

    // Wrap the header row in a container from the widget menu.
    const header = page.locator("article.component-card", {
      hasText: "Header row",
    });
    await header.hover();
    await header.getByRole("button", { name: "Component menu" }).click();
    await page.getByRole("menuitem", { name: "Wrap row in container" }).click();
    await expectProjectFile(page, CANVAS, [
      "rows:\n  - name: container-1\n    rows:",
      "content: Header row",
      "  - name: story",
    ]);

    // Add an empty container at the end and drag the loose widget into it.
    await addFromFloatingMenu(page, "Container");
    await expectProjectFile(page, CANVAS, [
      "content: Loose row",
      "  - name: container-2\n    rows: []",
    ]);
    await dragTo(
      page,
      page.locator("article.component-card", { hasText: "Loose row" }),
      page.locator("#block-add-3"),
      "center",
    );
    await expectProjectFile(
      page,
      CANVAS,
      [
        "  - name: story",
        "  - name: container-2\n    rows:\n      - items:",
        "content: Loose row",
      ],
      ["rows: []"],
    );

    // Move the story container above the header container by its grip.
    await dragTo(
      page,
      blockRegion(page, 1).locator(".chip .grip"),
      page.locator("#block-row-0"),
    );
    await expectProjectFile(page, CANVAS, [
      "rows:\n  - name: story",
      "content: Header row",
      "content: Loose row",
    ]);

    // Rename the story in its inspector, which a click on the chip opens.
    await blockRegion(page, 0).locator(".chip .label").click();
    const name = page.getByLabel("Container name");
    await expect(name).toHaveValue("story");
    await name.fill("revenue_story");
    await name.press("Enter");
    await expectProjectFile(page, CANVAS, [
      "rows:\n  - name: revenue_story\n    rows:",
    ]);
    await expect(blockRegion(page, 0).locator(".chip")).toContainText(
      "revenue_story",
    );

    // Show it to some viewers only.
    const condition = page.locator("#canvas-block-visibility-condition");
    await condition.fill("{{ .user.admin }}");
    await condition.press("Enter");
    await expectProjectFile(page, CANVAS, [
      '  - name: revenue_story\n    if: "{{ .user.admin }}"\n    rows:',
    ]);

    // Ungroup it: its row keeps the condition.
    await page.getByRole("button", { name: "Ungroup" }).click();
    await expectProjectFile(
      page,
      CANVAS,
      ['rows:\n  - if: "{{ .user.admin }}"\n    items:', "kpi_grid:"],
      ["name: revenue_story"],
    );

    // Delete the header container from its chip menu; it has content, so a dialog asks.
    const headerChip = blockRegion(page, 1).locator(".chip");
    await headerChip.getByRole("button", { name: "Container options" }).click();
    await page.getByRole("menuitem", { name: "Delete container" }).click();
    await page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Delete" })
      .click();
    await expectProjectFile(
      page,
      CANVAS,
      ["kpi_grid:", "content: Loose row"],
      ["Header row"],
    );
  });

  test("keeps the selection and a tab group's active tab when blocks shift", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 1600 });
    await putProjectFile(page, "/rill.yaml", RILL_YAML);
    await putProjectFile(page, SHIFT_CANVAS, SHIFT_YAML);
    await waitForReconciliation(page);
    await page.reload();
    await gotoNavEntry(page, SHIFT_CANVAS);
    await expect(page.getByText("Story row")).toBeVisible();

    // Show the group's second tab, then move the container above the group. The editor
    // names the blocks before it moves them, so each keeps its key and its state.
    await blockRegion(page, 1).getByRole("tab", { name: "Details" }).click();
    await dragTo(
      page,
      blockRegion(page, 2).locator(".chip .grip"),
      page.locator("#block-row-1"),
    );
    await expectProjectFile(page, SHIFT_CANVAS, [
      "content: First row",
      "  - name: container-2\n    rows:",
      "  - name: group-1\n    tabs:",
    ]);
    await expect(
      blockRegion(page, 2).getByRole("tab", { name: "Details" }),
    ).toHaveAttribute("aria-selected", "true");

    // Select the container, then insert another container above it, from the add menu
    // below the first row.
    await blockRegion(page, 1).locator(".chip .label").click();
    const name = page.getByLabel("Container name");
    await expect(name).toHaveValue("container-2");
    // The add button shows only while the pointer is over the row's edge, so open its menu
    // from the keyboard.
    await page
      .locator("#canvas-row-0")
      .getByRole("button", { name: "Insert widget in row 2", exact: true })
      .focus();
    await page.keyboard.press("Enter");
    await page
      .getByRole("menuitem", { name: "Container", exact: true })
      .click();
    await expectProjectFile(page, SHIFT_CANVAS, [
      "  - name: container-1\n    rows: []",
      "  - name: container-2\n    rows:",
    ]);

    // The selection stays on the container it was on, now one row down.
    await expect(blockRegion(page, 2).locator(".chip.selected")).toBeVisible();
    await expect(name).toHaveValue("container-2");
  });
});
