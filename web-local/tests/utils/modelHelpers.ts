import { expect } from "@playwright/test";
import type { Page } from "playwright";
import { renameFileUsingTitle } from "./commonHelpers";
import { waitForFileNavEntry } from "./waitHelpers";

export async function createModel(page: Page, modelFileName: string) {
  // Click add asset button
  await page.getByLabel("Add Asset").click();
  // Hover the add model option
  await page.getByLabel("Add Model").hover();
  // Click add blank model
  await page.getByLabel("Create blank model").click();

  // Wait for default model
  await waitForFileNavEntry(page, "/models/model.sql", true);

  // Rename model
  await renameFileUsingTitle(page, "model.sql", modelFileName);
  await waitForFileNavEntry(page, `/models/${modelFileName}`, true);
}

export async function modelHasError(page: Page, hasError: boolean, error = "") {
  const errorLocator = page.getByLabel("Model errors");
  try {
    await errorLocator.waitFor({
      timeout: 100,
    });
  } catch {
    // assertions not needed
  }

  if (hasError) {
    const actualError = await errorLocator.textContent();
    expect(actualError).toMatch(error);
  } else {
    expect(await errorLocator.count()).toBe(0);
  }
}
