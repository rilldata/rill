import { expect } from "@playwright/test";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";

test.describe("leaderboard measure names", () => {
  test.use({ project: "AdBids" });

  test.beforeEach(async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();
  });

  test("single selection", async ({ page }) => {
    await page.getByTestId("leaderboard-measure-names-dropdown").click();

    await page.getByRole("menuitem", { name: "Sum of Bid Price" }).click();

    // Check if the item is checked using aria-checked
    await expect(
      page.getByRole("menuitem", { name: "Sum of Bid Price" }),
    ).toHaveAttribute("aria-checked", "true");
  });

  test("multiple selection", async ({ page }) => {
    await page.getByTestId("leaderboard-measure-names-dropdown").click();

    await page.getByTestId("multi-measure-select-switch").click();
    await expect(page.getByTestId("multi-measure-select-switch")).toBeChecked();

    await page.getByRole("menuitem", { name: "Sum of Bid Price" }).click();
    await page.getByRole("menuitem", { name: "Total records" }).click();

    // Check if the items are checked using aria-checked
    await expect(
      page.getByRole("menuitem", { name: "Sum of Bid Price" }),
    ).toHaveAttribute("aria-checked", "true");
    await expect(
      page.getByRole("menuitem", { name: "Total records" }),
    ).toHaveAttribute("aria-checked", "true");
  });
});
