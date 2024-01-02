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
  columns: Array<string>
) {
  await page.waitForResponse(
    new RegExp(`/queries/columns-profile/tables/${name}`)
  );

  for (const column of columns) {
    await page.waitForResponse(
      new RegExp(`/queries/null-count/tables/${name}\\?columnName=${column}`)
    );
  }
}

export function getEntityLink(page: Page, name: string) {
  return page.getByRole("link", { name, exact: true });
}

/**
 * Runs an assertion multiple times until a timeout.
 * Throws the last thrown error by the assertion.
 */
export async function wrapRetryAssertion(
  assertion: () => Promise<void>,
  timeout = 1000,
  interval = 100
) {
  let lastError: Error;
  await asyncWaitUntil(
    async () => {
      try {
        await assertion();
        lastError = undefined;
        return true;
      } catch (err) {
        lastError = err;
        return false;
      }
    },
    timeout,
    interval
  );
  if (lastError) throw lastError;
}
