import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Organizations", () => {
  test("admins should see the admin-only pages", async ({ adminPage }) => {
    await adminPage.goto("/e2e");
    await expect(adminPage.getByRole("link", { name: "Users" })).toBeVisible();
    await expect(
      adminPage.getByRole("link", { name: "Settings" }),
    ).toBeVisible();
  });
});
