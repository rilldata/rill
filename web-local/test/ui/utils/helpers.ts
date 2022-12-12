import type { Page } from "playwright";

export enum TestEntityType {
  Source = "source",
  Model = "model",
  Dashboard = "dashboard",
}

export async function openEntityMenu(
  page: Page,
  type: TestEntityType,
  name: string
) {
  const entityLocator = page.locator(getEntityLink(page, type, name));
  await entityLocator.hover();
  await page
    // get the navigation entry for the entity
    .locator(".navigation-entry-title", {
      has: entityLocator,
    })
    .locator("div.contents div.contents button")
    .click();
}

export async function clickModalButton(page: Page, text: string) {
  return page
    .locator(".portal button", {
      hasText: text,
    })
    .click();
}

export async function clickMenuButton(page: Page, text: string) {
  await page
    .locator(".portal button[role='menuitem'] div.text-left div", {
      hasText: new RegExp(text),
    })
    .click();
}

export async function waitForProfiling(
  page: Page,
  name: string,
  columns: Array<string>
) {
  return Promise.all(
    [
      page.waitForResponse(
        new RegExp(`/queries/columns-profile/tables/${name}`)
      ),
      columns.map((column) =>
        page.waitForResponse(
          new RegExp(`/queries/null-count/tables/${name}?column_name=${column}`)
        )
      ),
    ].flat()
  );
}

export function getEntityLink(
  page: Page,
  type: TestEntityType,
  name: string
): string {
  return `a[href='/${type}/${name}']`;
}
