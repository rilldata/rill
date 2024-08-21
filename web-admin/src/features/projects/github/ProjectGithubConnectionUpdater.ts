import {
  createAdminServiceConnectProjectToGithub,
  type ListGithubUserReposResponseRepo,
} from "@rilldata/web-admin/client";
import { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors";
import { invalidateProjectQueries } from "@rilldata/web-admin/features/projects/invalidations";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import { get, writable } from "svelte/store";

/**
 * Wrapper around ConnectProjectToGithub RPC with some extra state moved out of components
 */
export class ProjectGithubConnectionUpdater {
  public readonly showOverwriteConfirmation = writable(false);
  public readonly connectToGithubMutation =
    createAdminServiceConnectProjectToGithub();

  public readonly githubUrl = writable("");
  public readonly subpath = writable("");
  public readonly branch = writable("");

  private readonly isConnected: boolean;
  private defaultBranch = "";

  public constructor(
    private readonly organization: string,
    private readonly project: string,
    private readonly currentGithubUrl: string,
    private readonly currentSubpath: string,
    private readonly currentBranch: string,
  ) {
    this.githubUrl.set(currentGithubUrl);
    this.subpath.set(currentSubpath);
    this.branch.set(currentBranch);
    this.isConnected = !!currentGithubUrl;
  }

  public onSelectedRepoChange(repo: ListGithubUserReposResponseRepo) {
    this.subpath.set("");
    this.branch.set(repo.defaultBranch ?? "");
    this.defaultBranch = repo.defaultBranch ?? "";
  }

  public reset() {
    this.githubUrl.set(this.currentGithubUrl);
    this.subpath.set(this.currentSubpath);
    this.branch.set(this.currentBranch);
  }

  public async update({
    instanceId,
    force,
  }: {
    instanceId: string;
    force: boolean;
  }) {
    const githubUrl = get(this.githubUrl);
    const subpath = get(this.subpath);
    const branch = get(this.branch);
    const hasSubpath = !!subpath;
    const hasNonDefaultBranch = branch !== this.defaultBranch;

    try {
      await get(this.connectToGithubMutation).mutateAsync({
        organization: this.organization,
        project: this.project,
        data: {
          repo: githubUrl,
          subpath,
          branch,
          force,
        },
      });

      behaviourEvent?.fireGithubIntentEvent(
        BehaviourEventAction.GithubConnectSuccess,
        {
          is_overwrite: force,
          has_subpath: hasSubpath,
          has_non_default_branch: hasNonDefaultBranch,
          is_fresh_connection: this.isConnected,
        },
      );
      this.reset();

      void invalidateProjectQueries(
        instanceId,
        this.organization,
        this.project,
      );
    } catch (e) {
      const err = extractGithubConnectError(e);
      if (!force && err.notEmpty) {
        behaviourEvent?.fireGithubIntentEvent(
          BehaviourEventAction.GithubConnectOverwritePrompt,
          {
            has_subpath: hasSubpath,
            has_non_default_branch: hasNonDefaultBranch,
            is_fresh_connection: this.isConnected,
          },
        );
        this.showOverwriteConfirmation.set(true);
        return false;
      } else {
        behaviourEvent?.fireGithubIntentEvent(
          BehaviourEventAction.GithubConnectFailure,
          {
            is_overwrite: force,
            has_subpath: hasSubpath,
            has_non_default_branch: hasNonDefaultBranch,
            is_fresh_connection: this.isConnected,
            failure_error: err.message,
          },
        );
        throw e;
      }
    }

    return true;
  }
}
