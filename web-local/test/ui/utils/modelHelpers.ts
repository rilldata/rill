import { expect } from "@jest/globals";
import { asyncWait } from "@rilldata/web-local/lib/util/waitUtils";
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
  await openEntityMenu(page, source);
  await clickMenuButton(page, "Create New Model");
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
