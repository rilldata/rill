import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Page } from "@sveltejs/kit";

// TODO: update all methods to use partial Page based on what is needed, so that it can be called in loader functions.

export function isOrganizationPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]" ||
    !!page.route?.id?.startsWith("/[organization]/-/users") ||
    !!page.route?.id?.startsWith("/[organization]/-/settings")
  );
}

export function withinOrganization({ route }: Pick<Page, "route">): boolean {
  return !!route?.id?.startsWith("/[organization]");
}

export function isProjectCreatePage(page: Page): boolean {
  return page.route.id === "/[organization]/-/create-project";
}

export function isProjectPage(page: Page): boolean {
  const routeId = page.route?.id;
  if (!routeId) return false;
  return (
    routeId === "/[organization]/[project]" ||
    (routeId.startsWith("/[organization]/[project]/-/") &&
      !routeId.startsWith("/[organization]/[project]/-/invite") &&
      !routeId.startsWith("/[organization]/[project]/-/share") &&
      !routeId.startsWith("/[organization]/[project]/-/edit"))
  );
}

export function withinProject(page: Page): boolean {
  return !!page.route?.id?.startsWith("/[organization]/[project]");
}

export function isMetricsExplorerPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]/explore/[dashboard]" ||
    page.route.id ===
      "/[organization]/[project]/-/share/[token]/explore/[dashboard]" ||
    page.route.id === "/-/embed"
  );
}

export function isCanvasDashboardPage(page: Page): boolean {
  return (
    page.route.id === "/[organization]/[project]/canvas/[dashboard]" ||
    page.route.id ===
      "/[organization]/[project]/-/share/[token]/canvas/[dashboard]"
  );
}

export function isPersonalFilePage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/personal/[name]";
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
  if (!page.route.id) return false;

  return (
    page.route.id.startsWith("/[organization]/[project]/-/share/[token]") ||
    isPublicReportPage(page) ||
    isPublicAlertPage(page)
  );
}

export function isPublicReportPage(page: Page): boolean {
  return (
    !!page.route.id?.startsWith(
      "/[organization]/[project]/-/reports/[report]",
    ) && page.url.searchParams.has("token")
  );
}

export function isPublicAlertPage(page: Page): boolean {
  return (
    !!page.route.id?.startsWith("/[organization]/[project]/-/alerts/[alert]") &&
    page.url.searchParams.has("token")
  );
}

export function isEditPage({ route }: Pick<Page, "route">): boolean {
  return !!route?.id?.startsWith("/[organization]/[project]/-/edit");
}

/**
 * True when the page is the explore or canvas preview inside Cloud Rill
 * Developer (`/-/edit/(viz)/{explore,canvas}/[name]`). `isMetricsExplorerPage`
 * and `isCanvasDashboardPage` only match production routes, so this is the
 * editor-side equivalent for surfaces that need to swap chat affordances.
 */
export function isEditDashboardPreviewPage({
  route,
}: Pick<Page, "route">): boolean {
  return !!route?.id?.startsWith("/[organization]/[project]/-/edit/(viz)/");
}

export function isProjectRequestAccessPage(page: Page): boolean {
  return !!page.route.id?.startsWith(
    "/[organization]/[project]/-/request-access",
  );
}

export function isProjectInvitePage(page: Page): boolean {
  return page.route.id === "/[organization]/[project]/-/invite";
}

export function isBillingUpgradePage(page: Page): boolean {
  return page.route.id === "/[organization]/-/upgrade-callback";
}

export function isWelcomePage({ route }: Pick<Page, "route">): boolean {
  return !!route.id?.startsWith("/-/welcome");
}

export function isProjectWelcomePage({ route }: Pick<Page, "route">): boolean {
  return !!route.id?.startsWith("/[organization]/[project]/-/edit/welcome");
}

export function isAuthPage({ route }: Pick<Page, "route">): boolean {
  return !!route.id?.startsWith("/-/auth");
}

/**
 * Returns true if the page is a page that is part of the onboarding flow.
 * Project invite page, org/project welcome page, and project create page are all onboarding pages as of now.
 * @param page
 */
export function isOnboardingPage(page: Page): boolean {
  return (
    isProjectInvitePage(page) ||
    isWelcomePage(page) ||
    isProjectWelcomePage(page) ||
    isProjectCreatePage(page)
  );
}

// `isProjectPage` matches every `/{org}/{project}/-/*` route bar a few exceptions, so it subsumes
// reports and alerts. It has to be tested last, or the specific screens below it are never reached.
export function getScreenNameFromPage(page: Page): MetricsEventScreenName {
  switch (true) {
    case isOrganizationPage(page):
      return MetricsEventScreenName.Organization;
    case isMetricsExplorerPage(page):
      return MetricsEventScreenName.Dashboard;
    case isCanvasDashboardPage(page):
      return MetricsEventScreenName.Canvas;
    case isReportExportPage(page):
      return MetricsEventScreenName.ReportExport;
    case isReportPage(page):
      return MetricsEventScreenName.Report;
    case isAlertPage(page):
      return MetricsEventScreenName.Alert;
    case isProjectPage(page):
      return MetricsEventScreenName.Project;
  }
  return MetricsEventScreenName.Unknown;
}

/**
 * Identifies the resource a page is showing, for telemetry.
 * Returns empty strings on pages that aren't showing a named resource, in which case consumers fall
 * back to parsing the resource out of the page URL.
 *
 * The type values are also produced by that URL fallback, which lives in the `rill_ui_telemetry_model`
 * model of the rill-cloud-metrics project. Keep the two vocabularies in step: they land in the same
 * column, so a value only used by one of them reads as two different resource types over time.
 */
export function getResourceFromPage(page: Page): {
  type: string;
  name: string;
} {
  switch (true) {
    case isMetricsExplorerPage(page):
      return { type: "explore", name: page.params.dashboard ?? "" };
    case isCanvasDashboardPage(page):
      return { type: "canvas", name: page.params.dashboard ?? "" };
    case isReportPage(page):
      return { type: "report", name: page.params.report ?? "" };
    case isAlertPage(page):
      return { type: "alert", name: page.params.alert ?? "" };
  }
  return { type: "", name: "" };
}
