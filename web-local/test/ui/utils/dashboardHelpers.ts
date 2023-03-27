import type { Page } from "playwright";
import { clickMenuButton, openEntityMenu } from "./helpers";

export async function createDashboardFromSource(page: Page, source: string) {
  await openEntityMenu(page, source);
  await clickMenuButton(page, "Autogenerate Dashboard");
}
