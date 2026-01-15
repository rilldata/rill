import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Azure connector form reset", () => {
  test.use({ project: "Blank" });

  test("clears connection string after submit and reopen", async ({ page }) => {
    // Open Add Data modal and pick Azure.
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#azure").click();
    await page.waitForSelector('form[id*="azure"]');

    const connectionString = page.getByRole("textbox", {
      name: "Connection string",
    });
    const submit = page
      .getByRole("dialog")
      .getByRole("button", { name: /(Test and Connect|Continue)/ });

    // Fill a connection string so the CTA enables and submit once.
    await connectionString.fill(
      "DefaultEndpointsProtocol=https;AccountName=test;AccountKey=abc;",
    );
    await expect(submit).toBeEnabled();
    await submit.click();

    // Save Anyway (expected for failed test) to close the modal, then wait for form unmount.
    const saveAnyway = page
      .getByRole("dialog")
      .getByRole("button", { name: "Save Anyway" });
    await expect(saveAnyway).toBeVisible();
    await saveAnyway.click();
    await page.waitForSelector('form[id*="azure"]', { state: "detached" });

    // Re-open and ensure the connection string is cleared.
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#azure").click();
    await page.waitForSelector('form[id*="azure"]');

    await expect(connectionString).toHaveValue("");
  });
});
