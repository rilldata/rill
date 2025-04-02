import { expect } from "@playwright/test";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";

async function escape(page) {
  await page.keyboard.press("Escape");
  await page.getByRole("menu").waitFor({ state: "hidden" });
}

async function toggleVisibleMeasureItem(page, itemName: string) {
  const itemId = itemName.toLowerCase().replace(/\s+/g, "_");
  await page
    .getByTestId("shown-section")
    .getByTestId(`visible-measures-${itemId}`)
    .getByTestId("toggle-visibility-button")
    .click();
}

// async function toggleHiddenMeasureItem(page, itemName: string) {
//   const itemId = itemName.toLowerCase().replace(/\s+/g, "_");
//   await page
//     .getByTestId("hidden-section")
//     .getByTestId(`hidden-measures-${itemId}`)
//     .getByTestId("toggle-visibility-button")
//     .click();
// }

async function toggleVisibleDimensionItem(page, itemName: string) {
  const itemId = itemName.toLowerCase().replace(/\s+/g, "_");
  await page
    .getByTestId("shown-section")
    .getByTestId(`visible-dimensions-${itemId}`)
    .getByTestId("toggle-visibility-button")
    .click();
}

// async function toggleHiddenDimensionItem(page, itemName: string) {
//   const itemId = itemName.toLowerCase().replace(/\s+/g, "_");
//   await page
//     .getByTestId("hidden-section")
//     .getByTestId(`hidden-dimensions-${itemId}`)
//     .getByTestId("toggle-visibility-button")
//     .click();
// }

test.describe("dimension and measure selectors", () => {
  test.use({ project: "AdBids" });

  test.beforeEach(async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();
  });

  test("measure selector flow", async ({ page }) => {
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
