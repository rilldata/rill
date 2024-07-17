import {
  createAdminServiceGetGithubUserStatus,
  getAdminServiceGetGithubUserStatusQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { get } from "svelte/store";

export class GithubConnection {
  public readonly userStatus = createAdminServiceGetGithubUserStatus();

  private connecting: boolean;

  public constructor(private readonly onReconnect: () => void) {}

  public async check() {
    this.connecting = false;
    await waitUntil(() => !get(this.userStatus).isLoading);
    const userStatus = get(this.userStatus).data;
    if (userStatus.hasAccess) {
      return this.onReconnect();
    }

    this.connecting = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  public async focused() {
    if (!this.connecting) return;
    await queryClient.refetchQueries(
      getAdminServiceGetGithubUserStatusQueryKey(),
    );
    if (this.connecting) this.onReconnect();
    this.connecting = false;
  }
}
