import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas global time filters", () => {
  test.use({ project: "AdBids" });

  // TODO: Fix test with latest time related changes
  test.skip("global time filters run through", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByRole("button", { name: "Preview" }).click();

    await page.waitForTimeout(1000);

    // Change global time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    await expect(page.getByText("Total records 272")).toBeVisible();

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Week to date" }).click();
    });

    await expect(page.getByText("Total records 3,435")).toBeVisible();

    const timeGrainSelector = page.getByRole("button", {
      name: "Select reference time and grain",
    });
    await timeGrainSelector.click();
    await page.getByRole("menuitem", { name: "day" }).click();

    await page.getByLabel("Toggle time comparison").click();

    await expect(page.getByText("Total records 3,435 +52 +2%")).toBeVisible();

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Today" }).click();
    });

    await timeGrainSelector.click();
    await page.getByRole("menuitem", { name: "hour" }).click();

    await expect(page.getByText("Total records 1,122")).toBeVisible();

    const timeZoneSelector = page.getByRole("button", {
      name: "Timezone selector",
    });
    await timeZoneSelector.click();

    await page.getByRole("textbox", { name: "Search" }).click();
    await page.getByRole("textbox", { name: "Search" }).fill("IST");
    await page.getByText("Asia/Calcutta").click();

    await expect(page.getByText("Total records 251")).toBeVisible();
  });
});
