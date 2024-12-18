import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Explores", () => {
  test("should have data", async ({ project: _, page }) => {
    // Wait for reconciliation to complete
    // (But, really, the dashboard link should be disabled until it's been reconciled)
    await page.waitForTimeout(5000);

    // Navigate to the explore
    await page
      .getByRole("link", { name: "Programmatic Ads Auction" })
      .first()
      .click();

    // Check the Big Number
    await expect(
      page.getByRole("button", { name: "Requests 635M" }),
    ).toBeVisible();
  });
});
