import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";

export enum TestEntityType {
  Source = "source",
  Model = "model",
  Dashboard = "dashboard",
}

export async function openEntityMenu(page: Page, name: string) {
  const entityLocator = getEntityLink(page, name);
  await entityLocator.hover();
  await entityLocator.locator("button").last().click();
}

export async function clickModalButton(page: Page, text: string) {
  return page.getByText(text).click();
}

export async function clickMenuButton(page: Page, text: string) {
  await page.getByRole("menuitem", { name: text }).click();
}

export async function waitForProfiling(
  page: Page,
  name: string,
  columns: Array<string>,
) {
  return Promise.all(
    [
      page.waitForResponse(
        new RegExp(`/queries/columns-profile/tables/${name}`),
      ),
      columns.map((column) =>
        page.waitForResponse(
          new RegExp(
            `/queries/null-count/tables/${name}\\?columnName=${column}`,
          ),
        ),
      ),
    ].flat(),
  );
}

export function getEntityLink(page: Page, name: string) {
  return page.getByRole("listitem", { name, exact: true });
}

/**
 * Runs an assertion multiple times until a timeout.
 * Throws the last thrown error by the assertion.
 */
export async function wrapRetryAssertion(
  assertion: () => Promise<void>,
  timeout = 1000,
  interval = 100,
) {
  let lastError: Error | undefined | string = undefined;
  await asyncWaitUntil(
    async () => {
      try {
        await assertion();
        lastError = undefined;
        return true;
      } catch (err) {
        if (err instanceof Error) lastError = err;
        else lastError = JSON.stringify(err);
        return false;
      }
    },
    timeout,
    interval,
  );
  if (lastError) throw lastError;
}

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
    .locator("#rill-portal h1", {
      hasText: "Rename",
    })
    .waitFor();

  // type new name and submit
  await page.locator("#rill-portal input").fill(toName);
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
