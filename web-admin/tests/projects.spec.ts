import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Projects", () => {
  test("admins should see the admin-only pages", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb");
    await expect(adminPage.getByRole("link", { name: "Status" })).toBeVisible();
    await expect(
      adminPage.getByRole("link", { name: "Settings" }),
    ).toBeVisible();
  });
});
