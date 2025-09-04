import { page } from "$app/stores";
import { getLocalGitRepoStatus } from "@rilldata/web-common/features/project/selectors.ts";
import { createLocalServiceGitStatus } from "@rilldata/web-common/runtime-client/local-service.ts";
import { derived } from "svelte/store";

export function getDeployRoute(pageUrl: URL) {
  const deployUrl = new URL(pageUrl);
  deployUrl.pathname = "/deploy";
  deployUrl.search = "";
  return deployUrl.toString();
}

export function getCreateProjectRoute(orgName: string, useGit = false) {
  const useGitParam = useGit ? "&use_git=true" : "";
  return `/deploy/project/create?org=${orgName}${useGitParam}`;
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

export function getDeployUsingGithubRoute(orgName: string) {
  return `/deploy/project/github?org=${orgName}`;
}

export function getCreateOrganizationRoute() {
  return `/deploy/organization/create`;
}

export function getSelectOrganizationRoute() {
  return `/deploy/organization/select`;
}

/**
 * Returns a readable for a route getter based on org name.
 * 1. If the project is not a github repo then returns {@link getCreateProjectRoute} that starts the deploy.
 * 2. If the project is a github repo then returns {@link getDeployUsingGithubRoute} that prompts the user for using either github/deploy.
 */
export function getDeployOrGithubRouteGetter() {
  return derived(createLocalServiceGitStatus(), ($gitStatus) => {
    const hasLocalGitRepo = Boolean(
      $gitStatus.data?.githubUrl && !$gitStatus.data?.managedGit,
    );
    return {
      isLoading: $gitStatus.isPending,
      getter: hasLocalGitRepo
        ? getDeployUsingGithubRoute
        : getCreateProjectRoute,
    };
  });
}

/**
 * Returns a readable for the deploy route for the open project.
 * 1. If the project is not a github repo it returns the create project route (/deploy/project/create) directly.
 *    Right now this is just a safeguard. We check upfront for git repo
 * 2. If the project is a github repo and we already have access to the repo then,
 *    it returns the create project route with github option (/deploy/project/create?use_git=true).
 * 3. If the project is a github repo and we do not have access to the repo then it returns github access route (<admin_server>/github/connect)
 */
export function getDeployRouteForProject(orgName: string) {
  return derived(
    [createLocalServiceGitStatus(), getLocalGitRepoStatus(), page],
    ([$gitStatus, $localGitRepoStatus, $page]) => {
      if ($gitStatus.isPending) return "";

      const hasLocalGitRepo = Boolean(
        $gitStatus.data?.githubUrl && !$gitStatus.data?.managedGit,
      );
      // Use the rill-managed deploy method if the project folder is not connected to git.
      if (!hasLocalGitRepo) return getCreateProjectRoute(orgName);
      const deployRoute = getCreateProjectRoute(orgName, true);

      if ($localGitRepoStatus.isPending) return "";

      // Already have access so deploy using git
      const hasRepoAccess = Boolean($localGitRepoStatus.data?.hasAccess);
      if (hasRepoAccess) return deployRoute;

      // type safety
      if (!$localGitRepoStatus.data?.grantAccessUrl)
        return getCreateProjectRoute(orgName);

      // Build the grant access url if we do not have access to the current repo.

      const deployUrl = new URL($page.url);
      deployUrl.pathname = deployRoute;
      deployUrl.search = "";

      const connectUrl = new URL($localGitRepoStatus.data.grantAccessUrl);
      connectUrl.searchParams.set("remote", $gitStatus.data?.githubUrl ?? "");
      connectUrl.searchParams.set("redirect", deployUrl.toString());
      return connectUrl.toString();
    },
  );
}
