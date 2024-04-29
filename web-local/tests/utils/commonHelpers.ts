import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import type {
  V1GetResourceResponse,
  V1ListResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import type { Page } from "playwright";

export enum TestEntityType {
  Source = "source",
  Model = "model",
  Dashboard = "dashboard",
}

export async function openFileNavEntryContextMenu(
  page: Page,
  filePath: string,
) {
  const entityLocator = getFileNavEntry(page, filePath);
  await entityLocator.hover();
  await entityLocator.getByLabel(`${filePath} actions menu trigger`).click();
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

export function getFileNavEntry(page: Page, filePath: string) {
  return page.getByLabel(`${filePath} Nav Entry`);
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

export async function goToFile(page: Page, filePath: string) {
  const link = page.locator(`a[id="${filePath}-nav-link"]`);
  await link.click();
}

export async function renameFileUsingMenu(
  page: Page,
  filePath: string,
  toFileName: string,
) {
  // open context menu and click rename
  await openFileNavEntryContextMenu(page, filePath);
  await clickMenuButton(page, "Rename...");

  // wait for rename modal to open
  await page
    .locator("#rill-portal h1", {
      hasText: "Rename",
    })
    .waitFor();

  // type new fileName and submit
  await page.locator("#rill-portal input").fill(toFileName);
  await Promise.all([
    page.waitForResponse(/rename/),
    clickModalButton(page, "Change Name"),
  ]);
}

export async function renameFileUsingTitle(page: Page, toName: string) {
  await page.locator("#model-title-input").fill(toName);
  await page.keyboard.press("Enter");
}

export async function deleteFile(page: Page, filePath: string) {
  // open context menu and click rename
  await openFileNavEntryContextMenu(page, filePath);
  await Promise.all([
    page.waitForResponse(
      (response) =>
        response.url().includes(filePath) &&
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
  await page.keyboard.insertText(code);
}

export async function waitForValidResource(
  page: Page,
  name: string,
  kind: string,
) {
  await page.waitForResponse(async (response) => {
    const responseUrl = response.url();
    const getResourceRequest = responseUrl.includes(
      `/v1/instances/default/resource?name.kind=${kind}&name.name=${name}`,
    );

    const listResourceRequest = responseUrl.includes(
      `/v1/instances/default/resource?name.kind=${kind}`,
    );

    if (getResourceRequest) {
      try {
        const resp = (await response.json()) as V1GetResourceResponse;
        return resp.resource?.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE";
      } catch (err) {
        return false;
      }
    } else if (listResourceRequest) {
      try {
        const resp = (await response.json()) as V1ListResourcesResponse;
        return (
          resp.resources?.find((r) => r.meta?.name === name)?.meta
            ?.reconcileStatus === "RECONCILE_STATUS_IDLE"
        );
      } catch (err) {
        return false;
      }
    }
    return false;
  });
}
