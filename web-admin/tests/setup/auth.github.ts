import { test as setup } from "./base";
import { GITHUB_AUTH_FILE } from "./constants";

setup("should authenticate to GitHub", async ({ page }) => {
  if (
    !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
    !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
  ) {
    throw new Error(
      "Missing required environment variables for authentication",
    );
  }

  // Log-in to GitHub
  await page.goto("https://github.com/login");
  await page.getByLabel("Username or email address").click();
  await page
    .getByLabel("Username or email address")
    .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL);
  await page.getByLabel("Password").click();
  await page
    .getByLabel("Password")
    .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();

  // Run in `debug` mode and pause here to do 2FA
  await page.pause();
  await page.waitForURL("https://github.com");

  // Save the auth cookies
  await page.context().storageState({ path: GITHUB_AUTH_FILE });
});
