import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Canvases", () => {
  test("should have data", async ({ page }) => {
    await page.goto("/e2e/openrtb");

    // Navigate to the explore
    await page
      .getByRole("link", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    // Check the KPI Grid data
    await expect(
      page.getByText("Advertising Spend Overall $447k"),
    ).toBeVisible();
  });
});
