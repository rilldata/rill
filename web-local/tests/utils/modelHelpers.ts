import { expect } from "@playwright/test";
import type { Page } from "playwright";
import { renameFileUsingTitle } from "./commonHelpers";
import { waitForFileNavEntry } from "./waitHelpers";

export async function createModel(page: Page, modelFileName: string) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add model menu item
  await page.getByLabel("Add Model").click();

  // Wait for default model
  await waitForFileNavEntry(page, "/models/model.sql", true);

  // Rename model
  await renameFileUsingTitle(page, modelFileName);
  await waitForFileNavEntry(page, `/models/${modelFileName}`, true);
}

export async function modelHasError(page: Page, hasError: boolean, error = "") {
  const errorLocator = page.locator(".editor-pane .error");
  try {
    await errorLocator.waitFor({
      timeout: 100,
    });
  } catch (err) {
    // assertions not needed
  }

  if (hasError) {
    const actualError = await errorLocator.textContent();
    expect(actualError).toMatch(error);
  } else {
    expect(await errorLocator.count()).toBe(0);
  }
}
