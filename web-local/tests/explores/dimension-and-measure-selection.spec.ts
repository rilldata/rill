import { expect } from "@playwright/test";
import { gotoNavEntry } from "../utils/waitHelpers";
import { test } from "../setup/base";

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

  // Click the eye icon to hide/show items
  async function toggleItemVisibility(
    page,
    itemName: string,
    sectionName: "Shown" | "Hidden",
  ) {
    await page
      .getByRole("button", { name: itemName })
      .locator(`.${sectionName.toLowerCase()}-section`)
      .getByRole("button", { name: "Toggle visibility" })
      .click();
  }

  test("measure selector flow", async ({ page }) => {
    const measuresButton = page.getByRole("button", {
      name: "Choose measures to display",
    });

    // Test individual measure toggling
    await measuresButton.click();
    // Show "Total records" from hidden section first
    await toggleItemVisibility(page, "Total records", "Hidden");
    // Now we can hide "Sum of Bid Price" since we have another measure shown
    await toggleItemVisibility(page, "Sum of Bid Price", "Shown");
    await escape(page);
    await expect(measuresButton).toHaveText("1 of 2 Measures");

    await expect(page.getByText("Sum of Bid Price 301k")).not.toBeVisible();
    await expect(page.getByText("Total records 100k")).toBeVisible();

    await measuresButton.click();
    // Show "Sum of Bid Price" from hidden section first
    await toggleItemVisibility(page, "Sum of Bid Price", "Hidden");
    // Now we can hide "Total records" since we have another measure shown
    await toggleItemVisibility(page, "Total records", "Shown");
    await expect(measuresButton).toHaveText("1 of 2 Measures");
    await escape(page);

    await expect(page.getByText("Sum of Bid Price 301k")).toBeVisible();
    await expect(page.getByText("Total records 100k")).not.toBeVisible();

    // Test "Hide all" and "Show all" functionality
    await measuresButton.click();
    // Click "Hide all" button in the shown section header
    await page.getByRole("button", { name: "Hide all" }).click();
    await expect(measuresButton).toHaveText("1 of 2 Measures");
    await escape(page);

    await expect(page.getByText("Sum of Bid Price 301k")).not.toBeVisible();
    await expect(page.getByText("Total records 100k")).toBeVisible();

    await measuresButton.click();
    // Click "Show all" button in the hidden section header
    await page.getByRole("button", { name: "Show all" }).click();
    await expect(measuresButton).toHaveText("All Measures");
    await escape(page);

    await expect(page.getByText("Sum of Bid Price 301k")).toBeVisible();
    await expect(page.getByText("Total records 100k")).toBeVisible();
  });

  test("dimension selector flow", async ({ page }) => {
    const dimensionsButton = page.getByRole("button", {
      name: "Choose dimensions to display",
    });

    // Test individual dimension toggling
    await dimensionsButton.click();
    // Show "Domain" from hidden section first
    await toggleItemVisibility(page, "Domain", "Hidden");
    // Now we can hide "Publisher" since we have another dimension shown
    await toggleItemVisibility(page, "Publisher", "Shown");
    await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
    await escape(page);

    await expect(page.getByText("Publisher")).not.toBeVisible();
    await expect(page.getByText("Domain")).toBeVisible();

    await dimensionsButton.click();
    // Show "Publisher" from hidden section first
    await toggleItemVisibility(page, "Publisher", "Hidden");
    // Now we can hide "Domain" since we have another dimension shown
    await toggleItemVisibility(page, "Domain", "Shown");
    await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
    await escape(page);

    await expect(page.getByText("Publisher")).toBeVisible();
    await expect(page.getByText("Domain")).not.toBeVisible();

    // Test "Hide all" and "Show all" functionality
    await dimensionsButton.click();
    // Click "Hide all" button in the shown section header
    await page.getByRole("button", { name: "Hide all" }).click();
    await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
    await escape(page);

    await expect(page.getByText("Publisher")).not.toBeVisible();
    await expect(page.getByText("Domain")).toBeVisible();

    await dimensionsButton.click();
    // Click "Show all" button in the hidden section header
    await page.getByRole("button", { name: "Show all" }).click();
    await expect(dimensionsButton).toHaveText("All Dimensions");
    await escape(page);

    await expect(page.getByText("Publisher")).toBeVisible();
    await expect(page.getByText("Domain")).toBeVisible();
  });
});
