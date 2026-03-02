import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";

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

export async function clickMenuButton(
  page: Page,
  text: string,
  role: "menuitem" | "option" = "menuitem",
) {
  await page.getByRole(role, { name: text }).click();
}

export async function waitForProfiling(
  page: Page,
  _name: string,
  columns: Array<string>,
) {
  return Promise.all(
    [
      page.waitForResponse("**/rill.runtime.v1.QueryService/TableColumns"),
      columns.map(() =>
        page.waitForResponse("**/rill.runtime.v1.QueryService/ColumnNullCount"),
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
  await clickMenuButton(page, "Rename");

  // wait for rename modal to open
  await page
    .locator("#rill-portal h2", {
      hasText: "Rename",
    })
    .waitFor();

  // type new fileName and submit
  await page.locator("#rill-portal input").fill(toFileName);
  await Promise.all([
    page.waitForResponse("**/rill.runtime.v1.RuntimeService/RenameFile"),
    clickModalButton(page, "Change Name"),
  ]);
}

export async function renameFileUsingTitle(
  page: Page,
  originalName: string,
  toName: string,
) {
  await page.getByRole("heading", { name: originalName, exact: true }).hover();
  await page.getByRole("button", { name: "File title actions" }).click();

  await page.locator("#model-title-input").fill(toName);
  await page.keyboard.press("Enter");
}

export async function deleteFile(page: Page, filePath: string) {
  // open context menu and click delete
  await openFileNavEntryContextMenu(page, filePath);
  await Promise.all([
    page.waitForResponse("**/rill.runtime.v1.RuntimeService/DeleteFile"),
    clickMenuButton(page, "Delete"),
  ]);
}

export async function updateCodeEditor(page: Page, code: string) {
  // Focus the editor
  const editor = page.getByRole("textbox", { name: "codemirror editor" });
  await editor.waitFor({ state: "visible" });
  await page.keyboard.press("Escape"); // Closes any tooltips
  await editor.click();

  // Select all text
  if (process.platform === "darwin") {
    await page.keyboard.press("Meta+A");
  } else {
    await page.keyboard.press("Control+A");
  }

  // Insert new text and wait for it to settle
  await page.keyboard.insertText(code);
  await page.waitForTimeout(600);
}
