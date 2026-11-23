import { expect } from "@playwright/test";
import { addCanvasWidget } from "@rilldata/web-common/tests/utils/canvas-helpers";
import { test } from "../setup/base";

test.describe("canvas", () => {
  test.use({ project: "AdBids" });

  test("can add widgets via primary button", async ({ page }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();

    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

    await addCanvasWidget(page, "KPI");

    await expect(
      page.getByRole("heading", { name: "Sum of Bid Price" }),
    ).toBeVisible();

    await addCanvasWidget(page, "Text/Markdown");

    await expect(
      page.getByRole("heading", { name: "H1 Markdown Text" }),
    ).toBeVisible();
  });

  test("can add widgets via column divider", async ({ page }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();

    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

    await addCanvasWidget(page, "KPI");

    await expect(
      page.getByRole("heading", { name: "Sum of Bid Price" }),
    ).toBeVisible();

    await page
      .getByRole("button", { name: "Resize row 1 column 1" })
      .hover({ force: true });

    await page
      .getByRole("button", {
        name: "Insert widget in row 1 at column 2",
        exact: true,
      })
      .click();

    await page.getByRole("menuitem", { name: "Text/Markdown" }).click();

    await expect(
      page.getByRole("heading", { name: "H1 Markdown Text" }),
    ).toBeVisible();
  });
});
