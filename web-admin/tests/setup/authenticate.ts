import dotenv from "dotenv";
import { test } from "./test";

dotenv.config();

const authFile = "playwright/.auth/user.json";

test("authenticate", async ({ page }) => {
  // Log in with the QA account
  await page.goto("/");
  await page.getByRole("button", { name: "Continue with Email" }).click();
  await page.getByPlaceholder("Enter your email address").click();
  await page
    .getByPlaceholder("Enter your email address")
    .fill(process.env.RILL_STAGE_QA_ACCOUNT_EMAIL);
  await page.getByPlaceholder("Enter your email address").press("Tab");
  await page
    .getByPlaceholder("Enter your password")
    .fill(process.env.RILL_STAGE_QA_ACCOUNT_PASSWORD);
  await page.getByRole("button", { name: "Continue with Email" }).click();

  // The login flow sets cookies in the process of several redirects.
  // Wait for the final URL to ensure that the cookies are actually set.
  await page.waitForURL("/");

  // End of authentication steps.
  await page.context().storageState({ path: authFile });
});
