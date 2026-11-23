import { addCanvasWidget } from "@rilldata/web-common/tests/utils/canvas-helpers";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas leaderboards", () => {
  test.use({ project: "AdBids" });

  test("add leaderboard component", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");
    await addCanvasWidget(page, "Leaderboard");

    await page.getByLabel("Add measure fields").click();

    await page
      .getByRole("menuitemcheckbox", { name: "Sum of Bid Price" })
      .click();
    await page
      .getByLabel("domain leaderboard")
      .locator("button")
      .filter({ hasText: "Sum of Bid Price" })
      .click();
  });
});
