import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  DeployingDashboardUrlParam,
  getDeployingDashboard,
} from "@rilldata/web-common/features/project/deploy/utils.ts";
import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
import type { Page } from "@sveltejs/kit";
import { derived, readable } from "svelte/store";

/**
 * Returns a {@link Readable} with a route to deploy.
 *
 * Adds a "deploying_dashboard" param based on the active page.
 * 1. If the user is on the editor page, gets the dashboard name from resourceName defined in {@link FileArtifact}. Since this is async we need a store derived from resourceName.
 * 2. If the user is on the visualization page, sets the dashboard name from route param.
 */
export function getDeployRoute(page: Page) {
  const deployUrl = new URL(page.url);
  deployUrl.pathname = "/deploy";
  deployUrl.search = "";

  if (isEditRoute(page)) {
    return getDeployRouteFromEditor(page, deployUrl);
  }

  if (isDashboardVizRoute(page)) {
    deployUrl.searchParams.set(DeployingDashboardUrlParam, page.params.name);
  }

  return readable(deployUrl.toString());
}

export function getCreateProjectRoute(orgName: string) {
  return `/deploy/project/create?org=${orgName}`;
}

export function getUpdateProjectRoute(
  orgName: string,
  projectName: string,
  createManagedRepo = false,
) {
  const createManagedRepoParam = createManagedRepo
    ? "&create_managed_repo=true"
    : "";
  return `/deploy/project/update?org=${orgName}&project=${projectName}${createManagedRepoParam}`;
}

export function getSelectProjectRoute() {
  return `/deploy/project/select`;
}

export function getOverwriteProjectRoute(orgName: string) {
  return `/deploy/project/overwrite?org=${orgName}`;
}

export function getCreateOrganizationRoute() {
  return `/deploy/organization/create`;
}

export function getSelectOrganizationRoute() {
  return `/deploy/organization/select`;
}

export function getDeployLandingPage(frontendUrl: string) {
  const url = new URL(frontendUrl);
  url.searchParams.set("deploying", "true");
  const deployingDashboard = getDeployingDashboard();
  if (deployingDashboard) {
    url.searchParams.set("deploying_dashboard", deployingDashboard);
  }
  url.pathname += "/-/invite";
  const projectInviteUrlWithSessionId = addPosthogSessionIdToUrl(
    url.toString(),
  );
  return projectInviteUrlWithSessionId;
}

function getDeployRouteFromEditor(page: Page, deployUrl: URL) {
  // Fetch the FileArtifact for the path and return a store derived from resourceName.
  const filePath = page.url.pathname.replace("/files", "");
  const curFile = fileArtifacts.getFileArtifact(filePath);
  return derived(curFile.resourceName, (curResource) => {
    const isDashboardResource =
      curResource?.kind === ResourceKind.Explore ||
      curResource?.kind === ResourceKind.Canvas;
    if (isDashboardResource && curResource.name) {
      deployUrl.searchParams.set(DeployingDashboardUrlParam, curResource.name);
    }

    return deployUrl.toString();
  });
}

function isDashboardVizRoute(page: Page) {
  return (
    page.route.id === "/(viz)/explore/[name]" ||
    page.route.id === "/(viz)/canvas/[name]"
  );
}

function isEditRoute(page: Page) {
  return page.route.id?.startsWith("/(application)/(workspace)/files");
}
