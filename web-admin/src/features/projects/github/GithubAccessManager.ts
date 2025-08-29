import {
  createAdminServiceGetGithubUserStatus,
  getAdminServiceGetGithubUserStatusQueryKey,
  getAdminServiceListGithubUserReposQueryKey,
} from "@rilldata/web-admin/client";
import { PopupWindow } from "@rilldata/web-common/lib/openPopupWindow.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { get, writable } from "svelte/store";

export class GithubAccessManager {
  public readonly githubConnectionFailed = writable(false);
  public userStatus = createAdminServiceGetGithubUserStatus(
    undefined,
    queryClient,
  );

  private userPromptWindow = new PopupWindow();

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

  /**
   * Ensures github access.
   * Opens the grantAccessUrl page to achieve this.
   */
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
    await queryClient.refetchQueries({
      queryKey: getAdminServiceGetGithubUserStatusQueryKey(),
    });
    await waitUntil(() => !get(this.userStatus).isFetching);

    if (!get(this.userStatus).data?.hasAccess) {
      this.githubConnectionFailed.set(true);
      return;
    }
    this.githubConnectionFailed.set(false);

    await queryClient.resetQueries({
      queryKey: getAdminServiceListGithubUserReposQueryKey(),
    });
  }
}
