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

export class GithubRepoUpdater {
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
          isLoading: true,
          error: undefined,
        };
      }

      return {
        isLoading: false,
        error: userStatus.error ?? userRepos.error,
      };
    },
  );

  private connecting: boolean;

  public constructor() {}

  public async check() {
    this.connecting = false;
    await waitUntil(() => !get(this.userStatus).isLoading);
    const userStatus = get(this.userStatus).data;

    this.connecting = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  public async focused() {
    if (!this.connecting) return;
    await queryClient.refetchQueries(
      getAdminServiceListGithubUserReposQueryKey(),
    );
    this.connecting = false;
  }
}
