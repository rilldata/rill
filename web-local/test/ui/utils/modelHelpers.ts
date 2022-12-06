import { expect } from "@jest/globals";
import type { Page } from "playwright";
import { renameEntityUsingTitle } from "./commonHelpers";
import { clickMenuButton, openEntityMenu, TestEntityType } from "./helpers";
import { waitForEntity } from "./waitHelpers";

export async function createModel(page: Page, name: string) {
  // add model button
  await page.locator("button#create-model-button").click();
  await waitForEntity(page, TestEntityType.Model, "model", true);
  await renameEntityUsingTitle(page, name);
  await waitForEntity(page, TestEntityType.Model, name, true);
}

export async function createModelFromSource(page: Page, source: string) {
  await openEntityMenu(page, TestEntityType.Source, source);
  await clickMenuButton(page, "create new model");
}

export async function updateModelSql(page: Page, sql: string) {
  await page.locator(".cm-line").first().click();
  if (process.platform === "darwin") {
    await page.keyboard.press("Meta+A");
  } else {
    await page.keyboard.press("Control+A");
  }
  await page.keyboard.press("Delete");
  await page.keyboard.insertText(sql);
}

export async function modelHasError(page: Page, hasError: boolean, error = "") {
  // TODO: better check
  try {
    const errorLocator = page.locator(".editor-pane .error");
    await errorLocator.waitFor({
      timeout: 100,
    });
    expect(hasError).toBeTruthy();
    const actualError = await errorLocator.textContent();
    expect(actualError).toMatch(error);
  } catch (err) {
    expect(hasError).toBeFalsy();
  }
}
