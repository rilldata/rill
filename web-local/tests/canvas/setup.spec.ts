import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("canvas", () => {
  test.use({ project: "AdBids" });

  test("can add widgets via primary button", async ({ page }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();

    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();

    await page.getByRole("menuitem", { name: "KPI" }).click();

    await expect(
      page.getByRole("heading", { name: "Sum of Bid Price" }),
    ).toBeVisible();

    await page.getByRole("button", { name: "Add widget" }).click();
    await page.getByRole("menuitem", { name: "Text" }).click();

    await expect(
      page.getByRole("heading", { name: "H1 Markdown Text" }),
    ).toBeVisible();
  });

  test("can add widgets via column divider", async ({ page }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();

    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();

    await page.getByRole("menuitem", { name: "KPI" }).click();

    await expect(
      page.getByRole("heading", { name: "Sum of Bid Price" }),
    ).toBeVisible();

    await page
      .getByRole("button", {
        name: "Insert widget in row 1 at column 2",
        exact: true,
      })
      .click();

    await page.getByRole("menuitem", { name: "KPI" }).click();
  });
});
