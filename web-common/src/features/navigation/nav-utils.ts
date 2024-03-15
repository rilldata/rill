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

export function isDashboardPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]/[dashboard]" ||
    page.route.id === "/-/embed" ||
    page.route.id === "/(application)/dashboard/[name]"
  );
}

export function isMetricsDefinitionPage(page: Page): boolean {
  return page.route.id === "/(application)/dashboard/[name]/edit";
}

export function isSourcePage(page: Page): boolean {
  return page.route.id === "/(application)/source/[name]";
}

export function isModelPage(page: Page): boolean {
  return page.route.id === "/(application)/model/[name]";
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
    case isDashboardPage(page):
      return MetricsEventScreenName.Dashboard;
    case isMetricsDefinitionPage(page):
      return MetricsEventScreenName.MetricsDefinition;
    case isSourcePage(page):
      return MetricsEventScreenName.Source;
    case isModelPage(page):
      return MetricsEventScreenName.Model;
    case isReportPage(page):
      return MetricsEventScreenName.Report;
    case isAlertPage(page):
      return MetricsEventScreenName.Alert;
    case isReportExportPage(page):
      return MetricsEventScreenName.ReportExport;
  }
  return MetricsEventScreenName.Unknown;
}
