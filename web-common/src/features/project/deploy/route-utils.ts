import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  DeployingDashboardUrlParam,
  getDeployingDashboard,
} from "@rilldata/web-common/features/project/deploy/utils.ts";
import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
import { createLocalServiceGitStatus } from "@rilldata/web-common/runtime-client/local-service";
import type { Page } from "@sveltejs/kit";
import { derived, readable } from "svelte/store";
import { getLocalGitRepoStatus } from "../selectors";
import { page } from "$app/stores";
import { featureFlags } from "../../feature-flags";
import type { V1ResourceName } from "@rilldata/web-common/runtime-client";

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

export function getCreateProjectRoute(orgName: string, useGit = false) {
  const useGitParam = useGit ? "&use_git=true" : "";
  return `/deploy/project/create?org=${orgName}${useGitParam}`;
}

export function getUpdateProjectRoute(
  page: Page,
  currentResource: V1ResourceName | undefined,
  orgName: string,
  projectName: string,
  createManagedRepo = false,
) {
  const deployUrl = new URL(page.url);
  deployUrl.pathname = "/deploy/project/update";
  deployUrl.search = "";

  deployUrl.searchParams.set("org", orgName);
  deployUrl.searchParams.set("project", projectName);
  if (createManagedRepo) {
    deployUrl.searchParams.set("create_managed_repo", "true");
  }

  if (
    currentResource?.kind === ResourceKind.Explore ||
    currentResource?.kind === ResourceKind.Canvas
  ) {
    deployUrl.searchParams.set(
      DeployingDashboardUrlParam,
      currentResource.name!,
    );
  } else if (isDashboardVizRoute(page)) {
    deployUrl.searchParams.set(DeployingDashboardUrlParam, page.params.name);
  }

  return deployUrl.toString();
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

export function getDeployingPageUrl(frontendUrl: string, isInvite: boolean) {
  const url = new URL(frontendUrl);
  const deployingDashboard = getDeployingDashboard();
  if (deployingDashboard) {
    url.searchParams.set(DeployingDashboardUrlParam, deployingDashboard);
  }
  if (isInvite) {
    url.pathname += "/-/invite";
  } else {
    url.pathname += "/-/deploying";
  }
  const projectInviteUrlWithSessionId = addPosthogSessionIdToUrl(
    url.toString(),
  );
  return projectInviteUrlWithSessionId;
}

/**
 * Returns a readable for a route getter based on org name.
 * 1. If the project is not a github repo then returns {@link getCreateProjectRoute} that starts the deploy.
 * 2. If the project is a github repo then returns {@link getDeployUsingGithubRoute} that prompts the user for using either github/deploy.
 */
export function getDeployOrGithubRouteGetter() {
  return derived(
    [createLocalServiceGitStatus(), featureFlags.legacyArchiveDeploy],
    ([$gitStatus, legacyArchiveDeploy]) => {
      const hasLocalGitRepo = Boolean(
        $gitStatus.data?.githubUrl && !$gitStatus.data?.managedGit,
      );
      // For E2E we cannot use github just yet. So do not show the git path for that case.
      const shouldUseGit = !legacyArchiveDeploy && hasLocalGitRepo;
      return {
        isLoading: $gitStatus.isPending,
        getter: shouldUseGit
          ? getDeployUsingGithubRoute
          : getCreateProjectRoute,
      };
    },
  );
}

/**
 * Returns a readable for the deploy route for the open project.
 * 1. If the project is not a github repo it returns the create project route (/developers/deploy/project/create) directly.
 *    Right now this is just a safeguard. We check upfront for git repo
 * 2. If the project is a github repo and we already have access to the repo then,
 *    it returns the create project route with github option (/developers/deploy/project/create?use_git=true).
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
