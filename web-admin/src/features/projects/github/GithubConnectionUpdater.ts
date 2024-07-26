import {
  createAdminServiceConnectProjectToGithub,
  getAdminServiceGetGithubUserStatusQueryKey,
  getAdminServiceGetProjectQueryKey,
} from "@rilldata/web-admin/client";
import { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";
import { get, writable } from "svelte/store";

export class GithubConnectionUpdater {
  public readonly showOverwriteConfirmation = writable(false);
  public readonly connectToGithubMutation =
    createAdminServiceConnectProjectToGithub();

  public async update({
    organization,
    project,
    githubUrl,
    subpath,
    branch,
    force,
    instanceId,
  }: {
    organization: string;
    project: string;
    githubUrl: string;
    subpath: string;
    branch: string;
    force: boolean;
    instanceId: string;
  }) {
    try {
      await get(this.connectToGithubMutation).mutateAsync({
        organization,
        project,
        data: {
          repo: githubUrl,
          subpath,
          branch,
          force,
        },
      });
      void queryClient.refetchQueries(
        getAdminServiceGetProjectQueryKey(organization, project),
        {
          // avoid refetching createAdminServiceGetProjectWithBearerToken
          exact: true,
        },
      );
      void queryClient.refetchQueries(
        getAdminServiceGetGithubUserStatusQueryKey(),
      );
      void invalidateRuntimeQueries(queryClient, instanceId);
    } catch (e) {
      if (force) {
        throw e;
      }
      const err = extractGithubConnectError(e);
      if (err.notEmpty) {
        this.showOverwriteConfirmation.set(true);
        return false;
      } else {
        throw e;
      }
    }

    return true;
  }
}
