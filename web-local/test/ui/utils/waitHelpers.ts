import { expect as playwrightExpect } from "@playwright/test";
import { asyncWait } from "@rilldata/web-local/common/utils/waitUtils";
import type { Page } from "playwright";
import { getEntityLink, TestEntityType } from "./helpers";

export async function waitForEntity(
  page: Page,
  type: TestEntityType,
  name: string,
  navigated: boolean
) {
  await page.locator(getEntityLink(page, type, name)).waitFor();
  if (navigated) {
    await page.waitForURL(new RegExp(`/${type}/${name}`));
  }
}

export async function entityNotPresent(
  page: Page,
  type: TestEntityType,
  name: string
) {
  await asyncWait(100);
  await playwrightExpect(
    page.locator(getEntityLink(page, type, name))
  ).toBeHidden();
}
