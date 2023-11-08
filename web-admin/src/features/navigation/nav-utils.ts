import type { Page } from "@sveltejs/kit";

export function isProjectPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]" ||
    page.route.id === "/[organization]/[project]/-/reports" ||
    page.route.id === "/[organization]/[project]/-/logs"
  );
}
export function isDashboardPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/[dashboard]";
}
