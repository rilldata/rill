import { expect } from "@playwright/test";
import { addCanvasWidget } from "@rilldata/web-common/tests/utils/canvas-helpers";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas kpi grid measure selector", () => {
  test.use({ project: "AdBids" });

  test("dropdown reflects selection state after toggling", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");
    await addCanvasWidget(page, "KPI");

    await page.getByLabel("Add measure fields").click();

    const item = page.getByRole("menuitemcheckbox", {
      name: "Sum of Bid Price",
    });
    await expect(item).toHaveAttribute("aria-checked", "true");
    // The menu stays open in multi-select mode and must reflect deselection...
    await item.click();
    await expect(item).toHaveAttribute("aria-checked", "false");
    // ...and reselection.
    await item.click();
    await expect(item).toHaveAttribute("aria-checked", "true");
  });
});
