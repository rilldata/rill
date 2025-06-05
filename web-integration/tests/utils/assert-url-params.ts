import { expect, type Page } from "@playwright/test";

export function assertUrlParams(page: Page, expectedParams: string) {
  expect(new URL(page.url()).searchParams.toString()).toMatch(
    new URLSearchParams(expectedParams).toString(),
  );
}
