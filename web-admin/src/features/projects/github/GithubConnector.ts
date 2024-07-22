import {
  createAdminServiceGetGithubUserStatus,
  getAdminServiceGetGithubUserStatusQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { get } from "svelte/store";

export class GithubConnector {
  public readonly userStatus = createAdminServiceGetGithubUserStatus();

  private connecting: boolean;

  public constructor(
    private readonly onConnect: () => void,
    private readonly onFailure: () => void,
  ) {}

  public async checkConnection() {
    await waitUntil(() => !get(this.userStatus).isLoading);
    const userStatus = get(this.userStatus).data;
    if (userStatus?.hasAccess) {
      return this.onConnect();
    }

    this.connecting = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  public async refetchStatus() {
    if (!this.connecting) return;
    await queryClient.refetchQueries(
      getAdminServiceGetGithubUserStatusQueryKey(),
    );

    await waitUntil(() => !get(this.userStatus).isLoading);
    const userStatus = get(this.userStatus).data;
    if (userStatus?.hasAccess) {
      this.connecting = false;
      this.onConnect();
    } else {
      this.onFailure();
    }
  }
}
