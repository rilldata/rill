import {
  createAdminServiceConnectProjectToGithub,
  createAdminServiceUpdateProject,
} from "@rilldata/web-admin/client";
import { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors";
import { derived, get, writable } from "svelte/store";

export class GithubConnectionUpdater {
  public readonly showOverwriteConfirmation = writable(false);
  public isCreate: boolean;

  private readonly updateProjectMutation = createAdminServiceUpdateProject();
  private readonly connectToGithubMutation =
    createAdminServiceConnectProjectToGithub();

  public readonly status = derived(
    [this.updateProjectMutation, this.connectToGithubMutation],
    ([updateProjectMutation, connectToGithubMutation]) => {
      if (
        updateProjectMutation.isLoading ||
        connectToGithubMutation.isLoading
      ) {
        return {
          isFetching: true,
          error: undefined,
        };
      }

      return {
        isFetching: false,
        error: updateProjectMutation.error ?? connectToGithubMutation.error,
      };
    },
  );

  public async update({
    organization,
    project,
    githubUrl,
    subpath,
    branch,
    force,
  }: {
    organization: string;
    project: string;
    githubUrl: string;
    subpath: string;
    branch: string;
    force: boolean;
  }) {
    try {
      if (this.isCreate) {
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
      } else {
        await get(this.updateProjectMutation).mutateAsync({
          organizationName: organization,
          name: project,
          data: {
            githubUrl,
            subpath,
            prodBranch: branch,
          },
        });
      }
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
