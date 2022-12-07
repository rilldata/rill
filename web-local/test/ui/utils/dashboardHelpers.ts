import type { Page } from "playwright";
import { clickMenuButton, openEntityMenu, TestEntityType } from "./helpers";

export async function createDashboardFromSource(page: Page, source: string) {
  await openEntityMenu(page, TestEntityType.Source, source);
  await clickMenuButton(page, "autogenerate dashboard");
}
