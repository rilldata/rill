import { expect } from "@playwright/test";
import { test } from "./setup/base";
import { uploadFile } from "./utils/sourceHelpers";

test.describe("Connector Table Menu Visibility", () => {
  test.use({ project: "Blank" });

  test("should show correct menu items for OLAP connector", async ({
    page,
  }) => {
    await uploadFile(page, "AdBids.csv");

    await page.getByText("View this source").click();

    await page.locator(".database-schema-entry-header").first().click();
    await page.locator(".table-entry-header").first().click();
    await page.locator(".table-entry-header").first().hover();
    await page.getByTestId("more-actions-context-button").click();

    // Wait for the menu to open
    await page.getByRole("menu").waitFor({ state: "visible" });

    await expect(
      page.getByRole("menuitem", { name: "Create model" }),
    ).toBeVisible({ timeout: 10000 });
    await expect(
      page.getByRole("menuitem", { name: "Generate metrics" }),
    ).toBeVisible({ timeout: 10000 });
    await expect(
      page.getByRole("menuitem", { name: "Generate dashboard" }),
    ).toBeVisible({ timeout: 10000 });
  });
});
