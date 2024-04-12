import { expect } from "@playwright/test";
import {
  extractFileName,
  splitFolderAndName,
} from "@rilldata/web-common/features/sources/extract-file-name";
import type { Page } from "playwright";
import { renameEntityUsingTitle } from "./commonHelpers";
import { waitForFileEntry } from "./waitHelpers";

export async function createModel(page: Page, filePath: string) {
  const [folder, fileName] = splitFolderAndName(filePath);
  const name = extractFileName(fileName);

  // add asset button
  await page.getByLabel("Add Asset").click();
  // add model menu item
  await page.getByLabel("Add Model").click();

  await waitForFileEntry(page, `${folder}/model.sql`, true);
  await renameEntityUsingTitle(page, name);
  await waitForFileEntry(page, filePath, true);
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
