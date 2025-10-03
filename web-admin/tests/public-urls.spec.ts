import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe.serial("Public URLs", () => {
  let publicUrl: string;

  test("should be able to create a public URL", async ({ page }) => {
    await page.goto("/e2e/openrtb/explore/auction_explore");

    // (Tests that the filtered column is hidden in the public URL)
    // Add a filter on "Pub Name"
    await page.getByLabel("Add Filter Button").click();
    await page.getByRole("menuitem", { name: "Pub Name" }).click();
    await page.getByLabel("Pub Name").getByPlaceholder("Search").click();
    await page.getByLabel("Pub Name").getByPlaceholder("Search").fill("disney");
    await page.getByRole("menuitem", { name: "Disney" }).first().click();
    await page.getByLabel("pub_name filter", { exact: true }).click(); // Hides the popover

    // Change the time grain to hour
    // (Tests that non-default state propagates to the public URL)
    await page.getByLabel("Select reference time and grain").click();
    await page.getByRole("menuitem", { name: "hour" }).click();

    // Check the Big Number
    await expect(
      page.getByRole("button", { name: "Requests 87,000" }),
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
    await page.getByRole("button", { name: "Create", exact: true }).click();

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
    test.skip(!publicUrl, "Public URL not set");

    await page.goto(publicUrl);

    // Check that the "Limited view" banner is visible
    await expect(
      page.getByText(
        "Limited view. For full access and features, visit the original dashboard.",
      ),
    ).toBeVisible();

    // Check that the Big Number reflects the filtered data
    await expect(
      page.getByRole("button", { name: "Requests 87,000" }),
    ).toBeVisible();

    // Check that the original dashboard state is reflected in the public URL
    await expect(
      page.getByLabel("Select reference time and grain").getByText("Hour"),
    ).toBeVisible();
  });

  test("anon should be able to view a public URL", async ({ anonPage }) => {
    test.skip(!publicUrl, "Public URL not set");

    await anonPage.goto(publicUrl);

    // Check that the "Limited view" banner is NOT visible
    await expect(
      anonPage.getByText(
        "Limited view. For full access and features, visit the original dashboard.",
      ),
    ).toBeHidden();

    // Check the Big Number reflects the filtered data
    await expect(
      anonPage.getByRole("button", { name: "Requests 87,000" }),
    ).toBeVisible();

    // Check that the filtered column is hidden
    await anonPage.getByLabel("Choose dimensions to display").click();
    await anonPage.getByPlaceholder("Search").fill("pub name");
    await expect(
      anonPage.getByText("No matching dimensions shown"),
    ).toBeVisible();
  });
});
