import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { expectProjectFile, putProjectFile } from "../utils/canvasBlockHelpers";
import { waitForReconciliation } from "../utils/wait-for-reconciliation";
import { gotoNavEntry } from "../utils/waitHelpers";

const CANVAS = "/dashboards/controls_canvas.yaml";

// A KPI row that opts out of controls, and a container whose Metric control swaps the measure
// of its chart. The chart also has a control of its own that swaps its dimension.
const CANVAS_YAML = `type: canvas
display_name: Controls canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: AdBids_metrics
          measures:
            - total_records
        ignore_controls: true
        width: 12
    height: 128px
  - name: story
    controls:
      - measure: total_records
        label: Metric
        options: [total_records, bid_price_sum]
      - filter: domain
        multiple: true
    rows:
      - items:
          - name: by_publisher
            controls:
              - dimension: publisher
                label: By
                options: [publisher, domain]
            bar_chart:
              metrics_view: AdBids_metrics
              x:
                field: publisher
                type: nominal
                limit: 20
              y:
                field: total_records
                type: quantitative
            width: 12
        height: 320px
`;

// One chart without controls, for adding one in the editor.
const EDIT_CANVAS_YAML = `type: canvas
display_name: Controls canvas
rows:
  - items:
      - bar_chart:
          metrics_view: AdBids_metrics
          x:
            field: publisher
            type: nominal
            limit: 20
          y:
            field: total_records
            type: quantitative
        width: 12
    height: 320px
`;

// Viewer controls are behind the canvas_controls flag.
const RILL_YAML = `compiler: rillv1
display_name: Controls e2e
features:
  chat: false
  canvas_controls: true
`;

test.describe("canvas viewer controls", () => {
  test.use({ project: "AdBids" });

  test("keeps controls on their defaults, without bars, when the flag is off", async ({
    page,
  }) => {
    await putProjectFile(page, CANVAS, CANVAS_YAML);
    await waitForReconciliation(page);
    await gotoNavEntry(page, CANVAS);

    await expect(page.getByText("Total records per Publisher")).toBeVisible();
    await expect(page.getByRole("group", { name: "Metric" })).toHaveCount(0);
    await expect(page.getByRole("group", { name: "By" })).toHaveCount(0);
  });

  test("swaps fields for the widgets under a control and keeps the choice in the URL", async ({
    page,
  }) => {
    await putProjectFile(page, "/rill.yaml", RILL_YAML);
    await putProjectFile(page, CANVAS, CANVAS_YAML);
    await waitForReconciliation(page);
    await page.reload();
    await gotoNavEntry(page, CANVAS);
    await page.getByRole("button", { name: "Preview" }).click();

    const metric = page.getByRole("group", { name: "Metric" });
    const by = page.getByRole("group", { name: "By" });
    await expect(page.getByText("Total records per Publisher")).toBeVisible();

    // The container's control changes the chart inside it.
    await metric.getByRole("button", { name: "Sum of Bid Price" }).click();
    await expect(
      page.getByText("Sum of Bid Price per Publisher"),
    ).toBeVisible();
    await expect(page).toHaveURL(/c\.story\.total_records=bid_price_sum/);
    // The KPI ignores controls, so it still shows the measure in its YAML.
    await expect(
      page.getByRole("button", { name: "total_records KPI data" }),
    ).toBeVisible();

    // The chart's own control is addressed by the chart's name.
    await by.getByRole("button", { name: "Domain" }).click();
    await expect(page.getByText("Sum of Bid Price per Domain")).toBeVisible();
    await expect(page).toHaveURL(/c\.by_publisher\.publisher=domain/);

    // Pointing at the container's control marks the chart it changes, and not the KPI.
    const chart = page.locator("article.component-card", {
      hasText: "per Domain",
    });
    const kpi = page.locator("article.component-card", {
      has: page.getByRole("button", { name: "total_records KPI data" }),
    });
    await metric.hover();
    await expect(chart).toHaveClass(/control-highlight/);
    await expect(kpi).not.toHaveClass(/control-highlight/);
    await page.mouse.move(0, 0);
    await expect(chart).not.toHaveClass(/control-highlight/);

    // A link with control values restores them.
    await page.reload();
    await expect(page.getByText("Sum of Bid Price per Domain")).toBeVisible();

    // Resetting goes back to what the YAML says and clears the URL.
    await page.getByRole("button", { name: "Reset controls" }).click();
    await expect(page.getByText("Total records per Publisher")).toBeVisible();
    await expect(page).not.toHaveURL(/c\./);
  });

  test("moves controls that don't fit on a line into a menu", async ({
    page,
  }) => {
    await putProjectFile(page, "/rill.yaml", RILL_YAML);
    await putProjectFile(
      page,
      CANVAS,
      CANVAS_YAML.replace("rows:", "max_width: 300\nrows:"),
    );
    await waitForReconciliation(page);
    await page.reload();
    await gotoNavEntry(page, CANVAS);
    await page.getByRole("button", { name: "Preview" }).click();

    // The container has two controls, which need more than the canvas's width.
    await expect(page.getByRole("group", { name: "Metric" })).toHaveCount(0);
    await page.getByRole("button", { name: "Adjust (2)" }).click();
    await page
      .getByRole("group", { name: "Metric" })
      .getByRole("button", { name: "Sum of Bid Price" })
      .click();
    await expect(
      page.getByText("Sum of Bid Price per Publisher"),
    ).toBeVisible();
  });

  test("adds, edits and removes a control in the visual editor", async ({
    page,
  }) => {
    await putProjectFile(page, "/rill.yaml", RILL_YAML);
    await putProjectFile(page, CANVAS, EDIT_CANVAS_YAML);
    await waitForReconciliation(page);
    await page.reload();
    await gotoNavEntry(page, CANVAS);

    // Offer the chart's measure to viewers from the link under its field.
    await page.getByText("Total records per Publisher").click();
    await page
      .getByRole("button", { name: "Let viewers choose" })
      .last()
      .click();
    const dialog = page.getByRole("dialog");
    await expect(
      dialog.getByText("Pick at least one more option"),
    ).toBeVisible();
    await dialog.getByLabel("Sum of Bid Price").click();
    await dialog.getByRole("button", { name: "Add control" }).click();

    // The chart gets a name, which addresses its controls, and the control before its component.
    await expectProjectFile(page, CANVAS, [
      "name: bar_chart\n",
      "controls:\n          - measure: total_records\n            options:\n              - total_records\n              - bid_price_sum\n",
    ]);
    // The control works in the editor's canvas.
    const control = page.getByRole("group", { name: "total_records" });
    await control.getByRole("button", { name: "Sum of Bid Price" }).click();
    await expect(
      page.getByText("Sum of Bid Price per Publisher"),
    ).toBeVisible();

    // Label it.
    await page.getByRole("button", { name: "Edit control" }).click();
    await dialog.locator("#canvas-control-label").fill("Metric");
    await dialog.getByRole("button", { name: "Save control" }).click();
    await expectProjectFile(page, CANVAS, ["label: Metric"]);
    await expect(page.getByRole("group", { name: "Metric" })).toBeVisible();

    // Removing the last control removes the list.
    await page.getByRole("button", { name: "Remove control" }).click();
    await expectProjectFile(page, CANVAS, ["name: bar_chart"], ["controls:"]);
    await expect(page.getByRole("group", { name: "Metric" })).toHaveCount(0);
  });
});
