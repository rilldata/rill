import type { Page } from "@sveltejs/kit";

export function isOrganizationPage(page: Page): boolean {
  return page.route.id === "/[organization]";
}

export function isProjectPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]" ||
    page.route.id === "/[organization]/[project]/-/reports" ||
    page.route.id === "/[organization]/[project]/-/status"
  );
}
export function isDashboardPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/[dashboard]";
}

export function isReportPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/reports/[report]";
}
