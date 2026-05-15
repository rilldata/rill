import { expect, type Page } from "@playwright/test";

export async function assertUrlParams(page: Page, expectedParams: string) {
  await expect
    .poll(() => new URL(page.url()).searchParams.toString(), {
      timeout: 10000,
      intervals: new Array(5).fill(2000),
    })
    .toMatch(new URLSearchParams(expectedParams).toString());
}
