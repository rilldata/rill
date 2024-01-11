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
  toName: string,
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
    page.waitForResponse(/rename/),
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
    page.waitForResponse(
      (response) =>
        response.url().includes(name) &&
        response.request().method() === "DELETE",
    ),
    clickMenuButton(page, "Delete"),
  ]);
}

export async function updateCodeEditor(page: Page, code: string) {
  await page.locator(".cm-line").first().click();
  if (process.platform === "darwin") {
    await page.keyboard.press("Meta+A");
  } else {
    await page.keyboard.press("Control+A");
  }
  await page.keyboard.press("Delete");
  await page.keyboard.insertText(code);
}

export async function waitForValidResource(
  page: Page,
  name: string,
  kind: string,
) {
  await page.waitForResponse(async (response) => {
    if (
      !response
        .url()
        .includes(
          `/v1/instances/default/resource?name.kind=${kind}&name.name=${name}`,
        )
    )
      return false;
    try {
      const resp = JSON.parse((await response.body()).toString());
      return resp.resource?.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE";
    } catch (err) {
      return false;
    }
  });
}
