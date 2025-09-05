import {
  adminServiceListGithubUserRepos,
  createAdminServiceGetGithubUserStatus,
  createAdminServiceListGithubUserOrgs,
  createAdminServiceListGithubUserRepos,
  getAdminServiceGetGithubUserStatusQueryKey,
  getAdminServiceListGithubUserReposQueryKey,
  getAdminServiceListGithubUserReposQueryOptions,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { PopupWindow } from "@rilldata/web-common/lib/openPopupWindow";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { createQuery } from "@tanstack/svelte-query";
import { getContext, setContext } from "svelte";
import { derived, get, type Readable, writable } from "svelte/store";

export type GithubSelectionType = "new" | "pull" | "push";

/**
 * Contains information about user connection to github and list of repos currently connected.
 */
export class GithubData {
  public readonly selectionType = writable<GithubSelectionType>("new");
  public readonly githubConnectionFailed = writable(false);

  public readonly userStatus = createAdminServiceGetGithubUserStatus(
    undefined,
    queryClient,
  );
  public readonly userOrgs: ReturnType<
    typeof createAdminServiceListGithubUserOrgs
  >;
  public readonly userRepos: ReturnType<
    typeof createAdminServiceListGithubUserRepos
  >;

  private userPromptWindow = new PopupWindow();

  public readonly status: Readable<{
    isFetching: boolean;
    error: RpcStatus | undefined;
  }>;

  public constructor() {
    const userOrgsOpts = derived(
      [this.userStatus, this.selectionType],
      ([userStatus, selectionType]) =>
        getAdminServiceListGithubUserReposQueryOptions({
          query: {
            // do not run it when user gets to status page, only when repo selection is open
            enabled: !!userStatus.data?.hasAccess && selectionType === "new",
          },
        }),
    );
    this.userOrgs = createQuery(userOrgsOpts) as ReturnType<
      typeof createAdminServiceListGithubUserOrgs
    >;

    const userReposOpts = derived(
      [this.userStatus, this.selectionType],
      ([userStatus, selectionType]) =>
        getAdminServiceListGithubUserReposQueryOptions({
          query: {
            // do not run it when user gets to status page, only when repo selection is open
            enabled: !!userStatus.data?.hasAccess && selectionType !== "new",
          },
        }),
    );
    this.userRepos = createQuery(userReposOpts) as ReturnType<
      typeof createAdminServiceListGithubUserRepos
    >;

    this.status = derived(
      [this.userStatus, this.userRepos],
      ([userStatus, userRepos]) => {
        if (userStatus.isFetching || userRepos.isFetching) {
          return {
            isFetching: true,
            error: undefined,
          };
        }

        return {
          isFetching: false,
          error: userStatus.error ?? userRepos.error,
        };
      },
    );
  }

  /**
   * Used to reselect connected repos.
   * Opens the grantAccessUrl page to achieve this.
   */
  public async reselectRepos() {
    await waitUntil(() => !get(this.userStatus).isFetching);
    const userStatus = get(this.userStatus).data;
    if (!userStatus?.grantAccessUrl) {
      return;
    }

    this.userPromptWindow
      .openAndWaitForClose(userStatus.grantAccessUrl + "?remote=autoclose")
      .then(() => this.refetch());
  }

  public async ensureGithubAccess() {
    await waitUntil(() => !get(this.userStatus).isFetching);
    const userStatus = get(this.userStatus).data;
    if (!userStatus || userStatus?.hasAccess || !userStatus.grantAccessUrl) {
      return;
    }

    this.userPromptWindow
      .openAndWaitForClose(userStatus.grantAccessUrl + "?remote=autoclose")
      .then(() => this.refetch());
  }

  private async refetch() {
    const userStatus = get(this.userStatus).data;
    if (!userStatus?.hasAccess) {
      // refetch status if had no access
      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetGithubUserStatusQueryKey(),
      });

      await waitUntil(() => !get(this.userStatus).isFetching);
      if (!get(this.userStatus).data?.hasAccess) {
        this.githubConnectionFailed.set(true);
      } else {
        this.githubConnectionFailed.set(false);
      }
    } else {
      this.githubConnectionFailed.set(false);

      // else refetch the list of repos
      await queryClient.resetQueries({
        queryKey: getAdminServiceListGithubUserReposQueryKey(),
      });
    }
  }
}

export function setGithubData(githubData: GithubData) {
  setContext("rill:app:github", githubData);
}

export function getGithubData() {
  return getContext<GithubData>("rill:app:github");
}
