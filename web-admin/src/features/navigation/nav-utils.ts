import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Page } from "@sveltejs/kit";

export function isOrganizationPage(page: Page): boolean {
  return page.route.id === "/[organization]";
}

export function withinOrganization(page: Page): boolean {
  return !!page.route?.id?.startsWith("/[organization]");
}

export function isProjectPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]" ||
    page.route.id === "/[organization]/[project]/-/reports" ||
    page.route.id === "/[organization]/[project]/-/alerts" ||
    page.route.id === "/[organization]/[project]/-/status" ||
    page.route.id === "/[organization]/[project]/-/settings" ||
    page.route.id === "/[organization]/[project]/-/settings/public-urls" ||
    !!page.route?.id?.startsWith("/[organization]/[project]/-/request-access")
  );
}

export function withinProject(page: Page): boolean {
  return !!page.route?.id?.startsWith("/[organization]/[project]");
}

export function isMetricsExplorerPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]/explore/[dashboard]" ||
    page.route.id === "/-/embed"
  );
}

export function isCanvasDashboardPage(page: Page): boolean {
  // TODO: Change the route to canvas
  return page.route.id === "/[organization]/[project]/-/dashboards/[dashboard]";
}

/**
 * Returns true if the page is any kind of dashboard page (either a Metrics Explorer or a Custom Dashboard).
 */
export function isAnyDashboardPage(page: Page): boolean {
  return isMetricsExplorerPage(page) || isCanvasDashboardPage(page);
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

export function isPublicURLPage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/share/[token]";
}

export function isProjectRequestAccessPage(page: Page): boolean {
  return !!page.route.id?.startsWith(
    "/[organization]/[project]/-/request-access",
  );
}

export function isProjectInvitePage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/invite";
}

export function getScreenNameFromPage(page: Page): MetricsEventScreenName {
  switch (true) {
    case isOrganizationPage(page):
      return MetricsEventScreenName.Organization;
    case isProjectPage(page):
      return MetricsEventScreenName.Project;
    case isMetricsExplorerPage(page):
      return MetricsEventScreenName.Dashboard;
    case isCanvasDashboardPage(page):
      return MetricsEventScreenName.Canvas;
    case isReportPage(page):
      return MetricsEventScreenName.Report;
    case isAlertPage(page):
      return MetricsEventScreenName.Alert;
    case isReportExportPage(page):
      return MetricsEventScreenName.ReportExport;
  }
  return MetricsEventScreenName.Unknown;
}
