import { getDeployingName } from "@rilldata/web-common/features/project/deploy/utils.ts";
import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
import type { Page } from "@sveltejs/kit";

export function getDeployRoute(page: Page) {
  const deployUrl = new URL(page.url);
  deployUrl.pathname = "/deploy";
  deployUrl.search = "";
  if (page.params.name) {
    deployUrl.searchParams.set("deploying_name", page.params.name);
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
  const dashboardName = getDeployingName();
  if (dashboardName) {
    url.searchParams.set("deploying_name", dashboardName);
  }
  url.pathname += "/-/invite";
  const projectInviteUrlWithSessionId = addPosthogSessionIdToUrl(
    url.toString(),
  );
  return projectInviteUrlWithSessionId;
}
