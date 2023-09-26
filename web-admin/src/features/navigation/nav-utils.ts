import type { Page } from "@sveltejs/kit";

export function isDashboardPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/[dashboard]";
}
