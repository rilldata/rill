import { expect } from "@playwright/test";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";

function kebabCase(str: string): string {
  return str.toLowerCase().replace(/\s+/g, "-");
}

test.describe("dimension and measure selectors", () => {
  test.use({ project: "AdBids" });

  test.beforeEach(async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();
  });

  async function escape(page) {
    await page.keyboard.press("Escape");
    await page.getByRole("menu").waitFor({ state: "hidden" });
  }

  // Click the eye icon to hide a shown item
  async function toggleVisibleMeasureItem(page, itemName: string) {
    const itemId = itemName.toLowerCase().replace(/\s+/g, "_");
    await page
      .getByTestId("shown-section")
      .getByTestId(`visible-measures-${itemId}`)
      .getByTestId("toggle-visibility-button")
      .click();
  }

  // Click the eye icon to show a hidden item
  async function toggleHiddenMeasureItem(page, itemName: string) {
    const itemId = itemName.toLowerCase().replace(/\s+/g, "_");
    await page
      .getByTestId("hidden-section")
      .getByTestId(`hidden-measures-${itemId}`)
      .getByTestId("toggle-visibility-button")
      .click();
  }

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

  // FIXME
  // test.skip("dimension selector flow", async ({ page }) => {
  //   const dimensionsButton = page.getByRole("button", {
  //     name: "Choose dimensions to display",
  //   });

  //   // Test individual dimension toggling
  //   await dimensionsButton.click();
  //   // Show "Domain" from hidden section first
  //   await toggleHiddenItem(page, "Domain");
  //   // Now we can hide "Publisher" since we have another dimension shown
  //   await toggleVisibleItem(page, "Publisher");
  //   await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
  //   await escape(page);

  //   await expect(page.getByText("Publisher")).not.toBeVisible();
  //   await expect(page.getByText("Domain")).toBeVisible();

  //   await dimensionsButton.click();
  //   // Show "Publisher" from hidden section first
  //   await toggleHiddenItem(page, "Publisher");
  //   // Now we can hide "Domain" since we have another dimension shown
  //   await toggleVisibleItem(page, "Domain");
  //   await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
  //   await escape(page);

  //   await expect(page.getByText("Publisher")).toBeVisible();
  //   await expect(page.getByText("Domain")).not.toBeVisible();

  //   // Test "Hide all" and "Show all" functionality
  //   await dimensionsButton.click();
  //   // Click "Hide all" button in the shown section header
  //   await page.getByRole("button", { name: "Hide all" }).click();
  //   await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
  //   await escape(page);

  //   await expect(page.getByText("Publisher")).not.toBeVisible();
  //   await expect(page.getByText("Domain")).toBeVisible();

  //   await dimensionsButton.click();
  //   // Click "Show all" button in the hidden section header
  //   await page.getByRole("button", { name: "Show all" }).click();
  //   await expect(dimensionsButton).toHaveText("All Dimensions");
  //   await escape(page);

  //   await expect(page.getByText("Publisher")).toBeVisible();
  //   await expect(page.getByText("Domain")).toBeVisible();
  // });
});
