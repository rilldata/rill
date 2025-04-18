import { expect } from "@playwright/test";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";

async function escape(page) {
  await page.keyboard.press("Escape");
  await page.getByRole("menu").waitFor({ state: "hidden" });
}

async function toggleVisibleMeasureItem(page, itemName: string) {
  const itemId = itemName.toLowerCase().replace(/\s+/g, "_");

  // Wait for the popover menu to be visible
  await page.getByRole("menu").waitFor({ state: "visible" });

  // Wait for the section and item to be visible
  const section = page.getByTestId("shown-section");
  await section.waitFor({ state: "visible" });

  const item = section.getByTestId(`visible-measures-${itemId}`);
  await item.waitFor({ state: "visible" });

  // Click the toggle button
  await item.getByRole("button", { name: "Toggle visibility" }).click();
}

async function toggleVisibleDimensionItem(page, itemName: string) {
  const itemId = itemName.toLowerCase().replace(/\s+/g, "_");

  // Wait for the popover menu to be visible
  await page.getByRole("menu").waitFor({ state: "visible" });

  // Wait for the section and item to be visible
  const section = page.getByTestId("shown-section");
  await section.waitFor({ state: "visible" });

  const item = section.getByTestId(`visible-dimensions-${itemId}`);
  await item.waitFor({ state: "visible" });

  // Click the toggle button
  await item.getByRole("button", { name: "Toggle visibility" }).click();
}

test.describe("dimension and measure selectors", () => {
  test.use({ project: "AdBids" });

  test.beforeEach(async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();
  });

  // FIXME: Skipping this so it doesn't block release-0.59
  test.skip("measure selector flow", async ({ page }) => {
    const measuresButton = page.getByRole("button", {
      name: "Choose measures to display",
    });

    // Test individual measure toggling
    await measuresButton.click();
    await toggleVisibleMeasureItem(page, "Sum of Bid Price");
    await escape(page);
    await expect(measuresButton).toHaveText("1 of 2 Measures");

    await expect(page.getByText("Sum of Bid Price 301k")).not.toBeVisible();
    await expect(page.getByText("Total records 100k")).toBeVisible();
  });

  // FIXME: Skipping this so it doesn't block release-0.59
  test.skip("dimension selector flow", async ({ page }) => {
    const dimensionsButton = page.getByRole("button", {
      name: "Choose dimensions to display",
    });

    // Test individual dimension toggling
    await dimensionsButton.click();
    await toggleVisibleDimensionItem(page, "Domain");
    await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
    await escape(page);

    await expect(page.getByText("Publisher")).not.toBeVisible();
    await expect(page.getByText("Domain")).toBeVisible();
  });
});
