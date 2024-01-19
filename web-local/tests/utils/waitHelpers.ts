import { expect } from "@playwright/test";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";
import { getEntityLink, TestEntityType } from "./commonHelpers";

export async function waitForEntity(
  page: Page,
  type: TestEntityType,
  name: string,
  navigated: boolean,
) {
  await getEntityLink(page, name).waitFor();
  if (navigated) {
    await page.waitForURL(new RegExp(`/${type}/${name}`));
  }
}

export async function entityNotPresent(page: Page, name: string) {
  await asyncWait(100);
  await expect(getEntityLink(page, name)).toBeHidden();
}
