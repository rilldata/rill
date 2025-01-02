import { test } from "./setup/base";
import { ADMIN_AUTH_FILE } from "./setup/constants";

test("authenticate the admin account", async ({ page }) => {
  if (
    !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
    !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
  ) {
    throw new Error(
      "Missing required environment variables for authentication",
    );
  }

  // Log in with the admin account
  await page.goto("/");
  await page.getByRole("button", { name: "Continue with Email" }).click();
  await page.getByPlaceholder("Enter your email address").click();
  await page
    .getByPlaceholder("Enter your email address")
    .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL);
  await page.getByPlaceholder("Enter your email address").press("Tab");
  await page
    .getByPlaceholder("Enter your password")
    .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD);
  await page.getByRole("button", { name: "Continue with Email" }).click();

  // The login flow sets cookies in the process of several redirects.
  // Wait for the final URL to ensure that the cookies are actually set.
  await page.waitForURL("/");

  // End of authentication steps.
  await page.context().storageState({ path: ADMIN_AUTH_FILE });
});
