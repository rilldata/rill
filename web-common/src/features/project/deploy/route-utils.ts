export function getDeployRoute(pageUrl: URL) {
  const deployUrl = new URL(pageUrl);
  deployUrl.pathname = "/deploy";
  deployUrl.search = "";
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

export function getGithubAccessUrl(grantAccessUrl: string, pageUrl: URL) {
  const redirectUrl = new URL(pageUrl);
  redirectUrl.searchParams.set("mode", "github");

  const githubAccessUrl = new URL(grantAccessUrl);
  githubAccessUrl.searchParams.set("redirect", redirectUrl.toString());
  return githubAccessUrl.toString();
}
