import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Multi-step connector wrapper", () => {
  test.use({ project: "Blank" });

  test("GCS connector - renders connector step schema via wrapper", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Choose a multi-step connector (GCS).
    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

    // Connector step should show connector preview and connector CTA.
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page
        .getByRole("dialog")
        .getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();

    // Auth method controls from the connector schema should render.
    const hmacRadio = page.getByRole("radio", { name: "HMAC keys" });
    await expect(hmacRadio).toBeVisible();
    await expect(page.getByRole("radio", { name: "Public" })).toBeVisible();

    // Select HMAC so its fields are rendered.
    await hmacRadio.click();

    // Connector step fields should be present, while source step fields should not yet render.
    await expect(
      page.getByRole("textbox", { name: "Access Key ID" }),
    ).toBeVisible();
    await expect(
      page.getByRole("textbox", { name: "Secret Access Key" }),
    ).toBeVisible();
    await expect(page.getByRole("textbox", { name: "GS URI" })).toHaveCount(0);
  });
});
