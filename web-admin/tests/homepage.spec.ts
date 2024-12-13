import { expect } from "@playwright/test";
import { test } from "./setup/base";

test("Authenticated user sees the homepage", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Hi qa@rilldata.com!")).toBeVisible();
});

test("Unauthenticated user gets redirected to login", async ({ anonPage }) => {
  await anonPage.goto("/");
  await expect(anonPage.getByText("Log in to Rill")).toBeVisible();
});
