import type { Page } from "playwright";
import {
  clickMenuButton,
  clickModalButton,
  getEntityLink,
  openEntityMenu,
} from "./helpers";

export async function gotoEntity(page: Page, name: string) {
  await getEntityLink(page, name).click();
}

export async function renameEntityUsingMenu(
  page: Page,
  name: string,
  toName: string
) {
  // open context menu and click rename
  await openEntityMenu(page, name);
  await clickMenuButton(page, "Rename...");

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

export async function deleteEntity(page: Page, name: string) {
  // open context menu and click rename
  await openEntityMenu(page, name);
  await Promise.all([
    page.waitForResponse(/delete-and-reconcile/),
    clickMenuButton(page, "Delete"),
  ]);
}
