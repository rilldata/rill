import type { Page } from "playwright";
import { waitForSource } from "./sourceHelpers";

export async function waitForAdBids(page: Page, name: string) {
  return waitForSource(page, name, ["publisher", "domain", "timestamp"]);
}

export async function waitForAdImpressions(page: Page, name: string) {
  return waitForSource(page, name, ["city", "country"]);
}
