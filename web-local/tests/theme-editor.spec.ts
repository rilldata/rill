import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Theme editor", () => {
  test.use({ project: "Blank" });

  test("Can create theme and view content", async ({ page }) => {
    // Create a new theme via Add -> More -> Theme
    await page.getByLabel("Add Asset").click();
    await page.getByRole("menuitem", { name: "More" }).hover();
    await page.getByRole("menuitem", { name: "Theme" }).click();

    // Check file exists in nav
    await expect(
      page.getByRole("link", { name: "theme.yaml" }).first(),
    ).toBeVisible();

    // Navigate to file
    await page.getByRole("link", { name: "theme.yaml" }).first().click();
    await page.waitForURL(/\/files\/themes\/theme\.yaml/);

    // Check file heading is shown
    await expect(
      page.getByRole("heading", { name: "theme.yaml" }),
    ).toBeVisible();

    // Check theme content is visible in editor
    await expect(page.getByText("type: theme")).toBeVisible();
    await expect(page.getByText("light:")).toBeVisible();

    // TODO: Test code/visual toggle and preview inspector once ThemeWorkspace
    // loading is fixed. Currently, newly created theme files don't load the
    // ThemeWorkspace component (which has the code/visual toggle and preview
    // inspector) - they fall back to the generic editor. This appears to be a
    // timing issue where resourceName isn't set before workspace selection.
  });
});
