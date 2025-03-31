import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { test } from "../setup/base";

test.describe("canvas time filters", () => {
  test.use({ project: "AdBids" });

  test.describe.serial("KPI component", () => {
    test("can add kpi component", async ({ page }) => {
      await page.getByLabel("Add asset").waitFor({ state: "visible" });
      await page.getByLabel("Add asset").click();

      await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

      await page
        .getByRole("button", { name: "Add widget" })
        .waitFor({ state: "visible" });
      await page.getByRole("button", { name: "Add widget" }).click();

      await page.getByRole("menuitem", { name: "KPI" }).click();

      await expect(
        page.getByRole("heading", { name: "Sum of Bid Price" }),
      ).toBeVisible();
    });

    test("can update global time filter", async ({ page }) => {
      await interactWithTimeRangeMenu(page, async () => {
        await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
      });

      // Check that the total records are 272 and have comparisons
      await expect(page.getByText("272 -23 -8%")).toBeVisible();
    });
  });
});
