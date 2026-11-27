import { expect } from "@playwright/test";
import { test } from "./setup/base";

const ADMIN_URL = "http://localhost:8080";

async function currentUserStatus(token: string) {
  const res = await fetch(`${ADMIN_URL}/v1/users/current`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.status;
}

test.describe("Personal access tokens", () => {
  test("creates a token with an expiration and revokes it", async ({
    page,
  }) => {
    const name = `E2E token ${Date.now()}`;

    // Open the page from the avatar menu
    await page.goto("/");
    await page.locator("header [data-dropdown-menu-trigger]").last().click();
    await page.getByRole("menuitem", { name: "Access tokens" }).click();
    await expect(page).toHaveURL(/\/-\/account\/tokens$/);

    // Create a token that expires in 7 days
    await page.getByRole("button", { name: "New token" }).first().click();
    const dialog = page.getByRole("dialog");
    await dialog.locator("#access-token-name").fill(name);
    const expiration = dialog.locator("#access-token-expiration");
    await expiration.click();
    await page.getByRole("menuitem", { name: "7 days" }).click();
    await expect(expiration).toHaveText("7 days");
    await dialog.getByRole("button", { name: "Create token" }).click();

    // The token is shown once and authenticates as the user
    await expect(dialog.getByText("Copy your new token")).toBeVisible();
    const token = (await dialog.locator("code").textContent())?.trim() ?? "";
    expect(token).toMatch(/^rill_usr_/);
    expect(await currentUserStatus(token)).toBe(200);
    await dialog.getByRole("button", { name: "Done" }).click();

    // The list shows it as expiring soon
    const row = page.locator(".row", { hasText: name });
    await expect(row).toBeVisible();
    await expect(row.getByText("Expires soon")).toBeVisible();

    // Revoking removes it and the token stops working
    await row.getByRole("button", { name: "Revoke" }).click();
    await page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Revoke token" })
      .click();
    await expect(row).toHaveCount(0);
    // The admin service caches token validation for 10 seconds.
    await expect
      .poll(() => currentUserStatus(token), { timeout: 20_000 })
      .toBe(401);
  });
});
