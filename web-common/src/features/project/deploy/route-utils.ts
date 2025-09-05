import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  DeployingDashboardUrlParam,
  getDeployingDashboard,
} from "@rilldata/web-common/features/project/deploy/utils.ts";
import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
import type { Page } from "@sveltejs/kit";
import { get } from "svelte/store";

export function getDeployRoute(page: Page) {
  const deployUrl = new URL(page.url);
  deployUrl.pathname = "/deploy";
  deployUrl.search = "";
  const deployingDashboard =
    getDashboardFromVizRoute(page) ?? getDashboardFromEditorRoute(page);
  if (deployingDashboard) {
    deployUrl.searchParams.set(DeployingDashboardUrlParam, deployingDashboard);
  }
  return deployUrl.toString();
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

function getDashboardFromVizRoute(page: Page) {
  if (
    page.route.id === "/(viz)/explore/[name]" ||
    page.route.id === "/(viz)/canvas/[name]"
  ) {
    return page.params.name;
  }
  return null;
}

function getDashboardFromEditorRoute(page: Page) {
  if (!page.route.id?.startsWith("/(application)/(workspace)/files")) {
    return null;
  }

  const filePath = page.url.pathname.replace("/files", "");
  const curFile = fileArtifacts.getFileArtifact(filePath);
  const curResource = get(curFile.resourceName);
  if (!curResource?.kind || !curResource?.name) return null; // type safety

  const isDashboard =
    curResource.kind === ResourceKind.Explore ||
    curResource.kind === ResourceKind.Canvas;
  return isDashboard ? curResource.name : null;
}
