import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { execAsync, spawnAndMatch } from "../utils/spawn";
import { test } from "./base";
import { ADMIN_AUTH_FILE } from "./constants";
import { cliLogin, cliLogout } from "./fixtures/cli";

let page: Page;

test.describe("Bootstrap", () => {
  test.beforeAll(async ({ browser }) => {
    page = await browser.newPage();
  });

  test("should authenticate and deploy a project", async () => {
    expect(page).toBeDefined();

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

    await cliLogin(page);
    const { stdout: orgCreateStdout } = await execAsync("rill org create e2e");
    expect(orgCreateStdout).toContain("Created organization");
    // Go to the organization's page
    await page.goto("/e2e");
    await expect(page.getByRole("heading", { name: "e2e" })).toBeVisible();
    // Deploy the project
    const { match } = await spawnAndMatch(
      "rill",
      [
        "deploy",
        "--path",
        "tests/setup/git/repos/rill-examples",
        "--subpath",
        "rill-openrtb-prog-ads",
        "--project",
        "openrtb",
        "--github",
        "true",
      ],
      /https?:\/\/[^\s]+/,
    );
    // Manually navigate to the auth URL
    const url = match[0];
    await page.goto(url);
    // Expect to see the successful deployment
    await page.goto("/e2e/openrtb");
    await expect(page.getByText("Your trial expires in 30 days")).toBeVisible(); // Billing banner
    await expect(page.getByText("e2e")).toBeVisible(); // Organization breadcrumb
    await expect(page.getByText("Free trial")).toBeVisible(); // Billing status
    await expect(page.getByText("openrtb")).toBeVisible(); // Project breadcrumb
    await expect(
      page.getByRole("link", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible(); // Link to dashboard
    await expect(
      page.getByRole("link", { name: "Programmatic Ads Bids" }),
    ).toBeVisible(); // Link to dashboard
  });
});

test.afterAll(async () => {
  // Delete the organization
  await cliLogin(page);
  await execAsync("rill org delete e2e --force");
  await cliLogout();
  await page.close();

  // Delete the project's Rill Cloud metadata
  await execAsync(
    "rm -rf tests/setup/git/repos/rill-examples/rill-openrtb-prog-ads/.rillcloud",
  );
});
