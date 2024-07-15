import { createAdminServiceGetGithubUserStatus } from "@rilldata/web-admin/client";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { get } from "svelte/store";

export class GithubConnection {
  public readonly userStatus = createAdminServiceGetGithubUserStatus({
    query: {
      refetchOnWindowFocus: true,
    },
  });

  private connecting: boolean;

  public constructor(private readonly onReconnect: () => void) {}

  public async check() {
    this.connecting = false;
    await waitUntil(() => !get(this.userStatus).isLoading);
    const userStatus = get(this.userStatus).data;
    if (userStatus.hasAccess) {
      this.onReconnect();
      return;
    }

    this.connecting = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  public focused() {
    if (this.connecting) this.onReconnect();
    this.connecting = false;
  }
}
