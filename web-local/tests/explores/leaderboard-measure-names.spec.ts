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

  test("measure selection", async ({ page }) => {
    // Test single selection
    await page.getByTestId("leaderboard-measure-names-dropdown").click();
    await page.getByRole("menuitem", { name: "Sum of Bid Price" }).click();

    // Reopen dropdown to verify selection
    await page.getByTestId("leaderboard-measure-names-dropdown").click();
    await expect(
      page.getByRole("menuitem", { name: "Sum of Bid Price" }),
    ).toHaveAttribute("aria-checked", "true");

    // Close dropdown and reopen to test multiple selection
    await page.keyboard.press("Escape");
    await page.getByTestId("leaderboard-measure-names-dropdown").click();

    // Verify initial state
    await expect(
      page.getByRole("menuitem", { name: "Sum of Bid Price" }),
    ).toHaveAttribute("aria-checked", "true");
    await expect(
      page.getByRole("menuitem", { name: "Total records" }),
    ).toHaveAttribute("aria-checked", "false");

    // Enable multi-select
    await page.getByTestId("multi-measure-select-switch").click();
    await expect(page.getByTestId("multi-measure-select-switch")).toBeChecked();

    // Select second measure and wait for state update
    await page.getByRole("menuitem", { name: "Total records" }).click();
    await expect(
      page.getByRole("menuitem", { name: "Total records" }),
    ).toHaveAttribute("aria-checked", "true");

    await page.keyboard.press("Escape");
    await expect(
      page.getByTestId("leaderboard-measure-names-dropdown"),
    ).toHaveAttribute("data-leaderboard-measures-count", "2");
  });
});
