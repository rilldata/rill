import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Page } from "@sveltejs/kit";

export function isOrganizationPage(page: Page): boolean {
  return page.route.id === "/[organization]";
}

export function isProjectPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]" ||
    page.route.id === "/[organization]/[project]/-/reports" ||
    page.route.id === "/[organization]/[project]/-/alerts" ||
    page.route.id === "/[organization]/[project]/-/status"
  );
}

export function isMetricsExplorerPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]/[dashboard]" ||
    page.route.id === "/-/embed"
  );
}

export function isCustomDashboardPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/dashboards/[dashboard]";
}

/**
 * Returns true if the page is any kind of dashboard page (either a Metrics Explorer or a Custom Dashboard).
 */
export function isAnyDashboardPage(page: Page): boolean {
  return isMetricsExplorerPage(page) || isCustomDashboardPage(page);
}

export function isReportPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/reports/[report]";
}

export function isAlertPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/alerts/[alert]";
}

export function isReportExportPage(page: Page): boolean {
  return (
    page.route.id ===
    "/[organization]/[project]/[dashboard]/-/reports/[report]/export"
  );
}

export function getScreenNameFromPage(page: Page): MetricsEventScreenName {
  switch (true) {
    case isOrganizationPage(page):
      return MetricsEventScreenName.Organization;
    case isProjectPage(page):
      return MetricsEventScreenName.Project;
    case isMetricsExplorerPage(page):
      return MetricsEventScreenName.Dashboard;
    case isReportPage(page):
      return MetricsEventScreenName.Report;
    case isAlertPage(page):
      return MetricsEventScreenName.Alert;
    case isReportExportPage(page):
      return MetricsEventScreenName.ReportExport;
  }
  return MetricsEventScreenName.Unknown;
}
