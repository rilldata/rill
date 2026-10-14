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
    // DimensionFilter content is portaled to document.body, so use page-level scope
    await page.getByPlaceholder("Search").click();
    await page.getByPlaceholder("Search").fill("disney");
    await page
      .getByRole("menuitemcheckbox", { name: "Disney" })
      .first()
      .click();
    await page.keyboard.press("Escape"); // Close the dimension filter popover

    // Change the time grain to hour
    // (Tests that non-default state propagates to the public URL)
    await page.getByLabel("Select reference time and grain").click();
    await page.getByRole("menuitemcheckbox", { name: "hour" }).click();

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

test.describe.serial("Password-protected public URLs", () => {
  let publicUrl: string;
  let password: string;

  test("should be able to create a password-protected public URL", async ({
    page,
  }) => {
    await page.goto("/e2e/openrtb/explore/auction_explore");

    await page.getByRole("button", { name: "Share" }).click();
    await page.getByRole("tab", { name: "Create public URL" }).click();
    await page
      .getByPlaceholder("Label this URL")
      .fill("Test Public URL - Password");

    // Turning on the switch shows an empty password field, which the button below it fills in
    await page.getByLabel("Require password").click();
    const passwordInput = page.getByPlaceholder("Password");
    await expect(passwordInput).toHaveValue("");
    const createButton = page.getByRole("button", {
      name: "Create",
      exact: true,
    });
    await expect(createButton).toBeDisabled();
    await page
      .getByRole("button", { name: "Generate and copy secure password" })
      .click();
    await expect(passwordInput).not.toHaveValue("");
    await createButton.click();

    // The URL and the password are both shown once after creation
    const copyButton = page.getByRole("button", { name: /Copy Public URL/ });
    await copyButton.waitFor();
    const url = await copyButton.getAttribute("data-public-url");
    const createdPassword = await page
      .locator("[data-public-url-password]")
      .getAttribute("data-public-url-password");
    if (!url || !createdPassword) {
      throw new Error("Could not find the public URL or its password");
    }
    expect(url).toContain("e2e/openrtb/-/share/rill_mgc_");
    expect(createdPassword).toHaveLength(16);

    publicUrl = url;
    password = createdPassword;
  });

  test("anon must enter the password to view the public URL", async ({
    anonPage,
  }) => {
    test.skip(!publicUrl, "Public URL not set");

    await anonPage.goto(publicUrl);

    // The dashboard is hidden behind a password prompt
    await expect(
      anonPage.getByText("This dashboard is password protected"),
    ).toBeVisible();
    await expect(
      anonPage.getByRole("button", { name: /Requests/ }),
    ).toBeHidden();

    // A wrong password is rejected
    await anonPage.getByPlaceholder("Password").fill("not the password");
    await anonPage.getByRole("button", { name: "Continue" }).click();
    await expect(anonPage.getByText("Incorrect password")).toBeVisible();

    // The right password shows the dashboard
    await anonPage.getByPlaceholder("Password").fill(password);
    await anonPage.getByRole("button", { name: "Continue" }).click();
    await expect(
      anonPage.getByRole("button", { name: /Requests/ }),
    ).toBeVisible();
    await expect(
      anonPage.getByText("This dashboard is password protected"),
    ).toBeHidden();
  });

  test("admin can remove the password from the settings page", async ({
    page,
    anonPage,
  }) => {
    test.skip(!publicUrl, "Public URL not set");

    await page.goto("/e2e/openrtb/-/settings/public-urls");
    const row = page.locator(".row", { hasText: "Test Public URL - Password" });
    await expect(row.getByLabel("Password protected")).toBeVisible();
    await row.getByRole("button").click();
    await page.getByRole("menuitem", { name: "Edit" }).click();

    const dialog = page.getByRole("dialog");
    await expect(
      dialog.getByText("A password is set", { exact: false }),
    ).toBeVisible();
    await dialog.getByLabel("Remove password").click();
    await dialog.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Public URL updated")).toBeVisible();
    await expect(row.getByLabel("Password protected")).toBeHidden();

    // The URL now opens without a password
    await anonPage.goto(publicUrl);
    await expect(
      anonPage.getByRole("button", { name: /Requests/ }),
    ).toBeVisible();
    await expect(
      anonPage.getByText("This dashboard is password protected"),
    ).toBeHidden();
  });
});
