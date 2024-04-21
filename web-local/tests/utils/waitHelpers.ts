import { expect } from "@playwright/test";
import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";
import { getEntityLink } from "./commonHelpers";

export async function waitForFileEntry(
  page: Page,
  filePath: string,
  navigated: boolean,
) {
  const [, fileName] = splitFolderAndName(filePath);
  await page.getByLabel(`${fileName} Nav Entry`).waitFor();
  if (navigated) {
    await page.waitForURL(new RegExp(`/files${filePath}`));
  }
}

export async function entityNotPresent(page: Page, name: string) {
  await asyncWait(100);
  await expect(getEntityLink(page, name)).toBeHidden();
}
