import { page } from "$app/state";
import { getDashboardResourceFromPage } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import {
  createRuntimeServiceListResources,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export class DashboardBookmarkMetadata {
  public inDashboard: boolean;
  public bookmarkUrlParams: URLSearchParams;

  private dashboardResource: ReturnType<typeof getDashboardResourceFromPage>;

  public constructor(runtimeClient: RuntimeClient) {
    this.dashboardResource = $derived(getDashboardResourceFromPage(page));
    this.inDashboard = $derived(!!this.dashboardResource);

    const allResourcesQuery = createRuntimeServiceListResources(
      runtimeClient,
      {},
    );

    let allResources = $state<V1Resource[]>([]);
    const allResourcesUnsub = allResourcesQuery.subscribe(
      (allResourcesResp) => {
        allResources = allResourcesResp.data?.resources ?? [];
      },
    );

    this.bookmarkUrlParams = $derived.by(() => {
      const bookmarkUrlParams = new URLSearchParams(page.url.searchParams);
      if (this.dashboardResource?.kind !== ResourceKind.Explore) {
        return bookmarkUrlParams;
      }

      const exploreSpec = allResources.find(
        (r) => r.meta?.name?.name === this.dashboardResource.name,
      )?.explore?.state?.validSpec;
      if (!exploreSpec?.metricsView) return;

      const metricsViewSpec = allResources.find(
        (r) => r.meta?.name?.name === this.dashboardResource.name,
      )?.metricsView?.state?.validSpec;
    });
  }
}
