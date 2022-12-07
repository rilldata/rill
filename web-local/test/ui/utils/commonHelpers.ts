import type { Page } from "playwright";
import {
  clickMenuButton,
  clickModalButton,
  getEntityLink,
  openEntityMenu,
  TestEntityType,
} from "./helpers";

export async function gotoEntity(
  page: Page,
  type: TestEntityType,
  name: string
) {
  await page.locator(getEntityLink(page, type, name)).click();
}

export async function renameEntityUsingMenu(
  page: Page,
  type: TestEntityType,
  name: string,
  toName: string
) {
  // open context menu and click rename
  await openEntityMenu(page, type, name);
  await clickMenuButton(page, "rename");

  // wait for rename modal to open
  await page
    .locator(".portal h1", {
      hasText: "Rename",
    })
    .waitFor();

  // type new name and submit
  await page.locator(".portal input").fill(toName);
  await Promise.all([
    page.waitForResponse(/rename-and-reconcile/),
    clickModalButton(page, "Change Name"),
  ]);
}

export async function renameEntityUsingTitle(page: Page, toName: string) {
  await page.locator("#model-title-input").fill(toName);
  await page.keyboard.press("Enter");
}

export async function deleteEntity(
  page: Page,
  type: TestEntityType,
  name: string
) {
  // open context menu and click rename
  await openEntityMenu(page, type, name);
  await Promise.all([
    page.waitForResponse(/delete-and-reconcile/),
    clickMenuButton(page, "delete"),
  ]);
}
