import { expect } from "@playwright/test";
import { test } from "./setup/test";

test("Unauthenticated user can see the login page", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Log in to Rill")).toBeVisible();
});
