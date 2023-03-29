import { expect as playwrightExpect } from "@playwright/test";
import { asyncWait } from "@rilldata/web-local/lib/util/waitUtils";
import type { Page } from "playwright";
import { getEntityLink, TestEntityType } from "./helpers";

export async function waitForEntity(
  page: Page,
  type: TestEntityType,
  name: string,
  navigated: boolean
) {
  await getEntityLink(page, name).waitFor();
  if (navigated) {
    await page.waitForURL(new RegExp(`/${type}/${name}`));
  }
}

export async function entityNotPresent(page: Page, name: string) {
  await asyncWait(100);
  await playwrightExpect(getEntityLink(page, name)).toBeHidden();
}
