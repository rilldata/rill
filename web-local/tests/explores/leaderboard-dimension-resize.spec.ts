import { expect } from "@playwright/test";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";

test.describe("leaderboard dimension column resize", () => {
  test.use({ project: "AdBids" });

  test.beforeEach(async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();
  });

  test("resizing one leaderboard resizes all and persists", async ({
    page,
  }) => {
    const publisherCol = page
      .getByLabel("publisher leaderboard")
      .locator("col[data-dimension-column]");
    const domainCol = page
      .getByLabel("domain leaderboard")
      .locator("col[data-dimension-column]");

    // Both leaderboards start at the default width
    await expect(publisherCol).toHaveAttribute("style", /width: 164px/);
    await expect(domainCol).toHaveAttribute("style", /width: 164px/);

    // Drag the resize handle on the publisher leaderboard 76px to the right
    const resizer = page
      .getByLabel("publisher leaderboard")
      .locator("th[data-dimension-header] button.EW");
    const box = await resizer.boundingBox();
    if (!box) throw new Error("Resizer not found");
    const startX = box.x + box.width / 2;
    const startY = box.y + box.height / 2;
    await page.mouse.move(startX, startY);
    await page.mouse.down();
    await page.mouse.move(startX + 40, startY, { steps: 4 });
    await page.mouse.move(startX + 76, startY, { steps: 4 });
    await page.mouse.up();

    // Both leaderboards follow the new width
    await expect(publisherCol).toHaveAttribute("style", /width: 240px/);
    await expect(domainCol).toHaveAttribute("style", /width: 240px/);

    // The width is persisted to local storage (debounced)
    await expect
      .poll(() =>
        page.evaluate(() =>
          localStorage.getItem("leaderboard-dimension-column-width"),
        ),
      )
      .toBe("240");

    // The width is restored after a reload
    await page.reload();
    await expect(publisherCol).toHaveAttribute("style", /width: 240px/);
    await expect(domainCol).toHaveAttribute("style", /width: 240px/);
  });
});
