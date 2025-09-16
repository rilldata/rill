import {
  createAdminServiceGetGithubUserStatus,
  getAdminServiceGetGithubUserStatusQueryKey,
} from "@rilldata/web-admin/client";
import { PopupWindow } from "@rilldata/web-common/lib/openPopupWindow.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
import { get, writable } from "svelte/store";

/**
 * Handles github access. Opens a popup window to prompt user to install rill github app.
 * When the popup is closed `githubConnectionFailed` is set if user did not give access.
 */
export class GithubAccessManager {
  public readonly githubConnectionFailed = writable(false);
  public userStatus = createAdminServiceGetGithubUserStatus(
    undefined,
    queryClient,
  );

  private userPromptWindow = new PopupWindow();

  /**
   * Used to reselect connected orgs.
   * Opens the grantAccessUrl page to achieve this.
   */
  public async reselectOrgs() {
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

    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectStart,
      {
        is_fresh_connection: true,
      },
    );
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
  }
}
