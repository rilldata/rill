import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Public URLs", () => {
  let publicUrl: string;

  test("should be able to create a public URL", async ({ page }) => {
    await page.goto("/e2e/openrtb/explore/auction_explore");

    // Add a filter on "Pub Name"
    await page.locator('[id="filter-add-btn"]').click();
    await page.getByRole("menuitem", { name: "Pub Name" }).click();
    await page.getByLabel("Pub Name").getByPlaceholder("Search").click();
    await page.getByLabel("Pub Name").getByPlaceholder("Search").fill("disney");
    await page.getByRole("menuitem", { name: "Disney" }).first().click();
    await page.getByLabel("View filter").first().click(); // Hides the popover

    // Set the time zone to UTC
    await page.getByLabel("Timezone selector").click();
    await page.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }).click();

    // Check the Big Number
    await expect(
      page.getByRole("button", { name: "Requests 10.2M" }),
    ).toBeVisible();

    // Create a Public URL
    await page.getByRole("button", { name: "Share" }).click();
    await page.getByRole("tab", { name: "Create public URL" }).click();
    await page.getByPlaceholder("Label this URL").click();
    await page
      .getByPlaceholder("Label this URL")
      .fill("Test Public URL - Disney");
    await page.getByLabel("Set expiration").click();
    await page.locator('[aria-label="Edit expiration date"]').click(); // Hides the popover
    await page.getByRole("button", { name: "Create" }).click();

    // Wait for the "Copy Public URL" button to appear
    const copyButton = page.getByRole("button", { name: /Copy Public URL/ });
    await copyButton.waitFor();

    // Get the URL from the data attribute
    const url = await copyButton.getAttribute("data-public-url");
    if (!url) {
      throw new Error("Could not find the public URL on the button");
    }
    expect(url).toContain("e2e/openrtb/-/share/rill_mgc_");

    // Save the URL for the subsequent tests
    publicUrl = url;
  });

  test("admin should be able to view a public URL", async ({ page }) => {
    // Skip this test if the publicUrl wasn't set
    test.skip(!publicUrl, "Public URL not created in previous test");

    await page.goto(publicUrl);

    // Check that the "Limited view" banner is visible
    await expect(
      page.getByText(
        "Limited view. For full access and features, visit the original dashboard.",
      ),
    ).toBeVisible();

    // Change to UTC time zone
    await page.getByLabel("Timezone selector").click();
    await page.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }).click();

    // Check that the Big Number reflects the filtered data
    await expect(
      page.getByRole("button", { name: "Requests 10.2M" }),
    ).toBeVisible();
  });

  test("anon should be able to view a public URL", async ({ anonPage }) => {
    // Skip this test if the publicUrl wasn't set
    test.skip(!publicUrl, "Public URL not created in previous test");

    await anonPage.goto(publicUrl);

    // Check that the "Limited view" banner is NOT visible
    await expect(
      anonPage.getByText(
        "Limited view. For full access and features, visit the original dashboard.",
      ),
    ).toBeHidden();

    // Change to UTC time zone
    await anonPage.getByLabel("Timezone selector").click();
    await anonPage
      .getByRole("menuitem", { name: "UTC GMT +00:00 UTC" })
      .click();

    // Check the Big Number reflects the filtered data
    await expect(
      anonPage.getByRole("button", { name: "Requests 10.2M" }),
    ).toBeVisible();

    // Check that the filtered column is hidden
    await anonPage.getByLabel("Choose dimensions to display").click();
    await anonPage.getByPlaceholder("Search").fill("pub name");
    await expect(
      anonPage.getByTestId("searchable-menu-no-results"),
    ).toBeVisible();
  });
});
