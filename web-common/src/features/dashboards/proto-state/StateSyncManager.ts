import { goto } from "$app/navigation";
import { page } from "$app/stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export class StateSyncManager {
  private protoState: string;
  private urlState: string;
  private updating = false;

  public constructor(private readonly metricViewName: string) {}

  public handleStateChange(metricsExplorer: MetricsExplorerEntity) {
    const pageUrl = get(page).url;
    let curUrlState = pageUrl.searchParams.get("state");

    if (this.protoState === metricsExplorer.proto) return;
    this.protoState = metricsExplorer.proto;

    if (this.protoState === curUrlState) return;

    // if state didn't change do not call goto. this avoids adding unnecessary urls to history stack
    if (this.protoState !== this.urlState) {
      if (this.protoState === metricsExplorer.defaultProto) {
        goto(`${pageUrl.pathname}`);
      } else {
        goto(`${pageUrl.pathname}?state=${this.protoState}`);
      }
      this.updating = true;
    }
  }

  public handleUrlChange(
    metricsExplorer: MetricsExplorerEntity,
    metricsView: V1MetricsView
  ) {
    const pageUrl = get(page).url;
    let newUrlState = pageUrl.searchParams.get("state");
    if (this.urlState === newUrlState) return;
    if (!newUrlState && metricsExplorer?.defaultProto) {
      newUrlState = metricsExplorer.defaultProto;
    }
    this.urlState = newUrlState;

    // run sync if we didn't change the url through a state change
    // this can happen when url is updated directly by the user
    if (!this.updating && this.urlState && this.urlState !== this.protoState) {
      metricsExplorerStore.syncFromUrl(
        this.metricViewName,
        this.urlState,
        metricsView
      );
    }
    this.updating = false;
  }
}
