import {
  adminServiceListGithubUserRepos,
  createAdminServiceGetGithubUserStatus,
  createAdminServiceListGithubUserRepos,
  getAdminServiceListGithubUserReposQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { derived, get } from "svelte/store";

export class GithubReposFetcher {
  public readonly userStatus = createAdminServiceGetGithubUserStatus();
  public readonly userRepos = derived([this.userStatus], ([userStatus], set) =>
    createAdminServiceListGithubUserRepos({
      query: {
        enabled: !!userStatus.data?.hasAccess,
        queryClient,
      },
    }).subscribe(set),
  ) as ReturnType<
    typeof createAdminServiceListGithubUserRepos<
      Awaited<ReturnType<typeof adminServiceListGithubUserRepos>>,
      RpcStatus
    >
  >;

  public readonly status = derived(
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

  private promptingUser: boolean;

  public constructor() {}

  public async promptUser() {
    await waitUntil(() => !get(this.userStatus).isLoading);
    const userStatus = get(this.userStatus).data;

    this.promptingUser = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  /**
   * Called when page is back in focus.
   * If it happens just after selecting new list of repos (triggered in {@link promptUser}),
   * then we refetch the repos query to get the latest list.
   */
  public async handlePageFocus() {
    if (!this.promptingUser) return;
    this.promptingUser = false;
    await queryClient.refetchQueries(
      getAdminServiceListGithubUserReposQueryKey(),
    );
  }
}
