import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { throttledRefreshResource } from "@rilldata/web-common/features/entity-management/resource-invalidations";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getConnectorServiceOLAPListTablesQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1ReconcileStatus,
  V1Resource,
  V1ResourceEvent,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import {
  invalidateChartData,
  invalidateMetricsViewData,
  invalidateProfilingQueries,
  invalidationForMetricsViewData,
} from "@rilldata/web-common/runtime-client/invalidation";
import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import { get } from "svelte/store";

const MainResourceKinds: {
  [kind in ResourceKind]?: true;
} = {
  [ResourceKind.Source]: true,
  [ResourceKind.Model]: true,
  [ResourceKind.MetricsView]: true,
  [ResourceKind.Component]: true,
  [ResourceKind.Dashboard]: true,
};
const UsedResourceKinds: {
  [kind in ResourceKind]?: true;
} = {
  [ResourceKind.ProjectParser]: true,
  [ResourceKind.Theme]: true,
  ...MainResourceKinds,
};

export class WatchResourcesClient {
  public readonly client: WatchRequestClient<V1WatchResourcesResponse>;
  private readonly tables = new Map<string, string>();

  public constructor() {
    this.client = new WatchRequestClient<V1WatchResourcesResponse>();
    this.client.on("response", (res) => this.handleWatchResourceResponse(res));
    this.client.on("reconnect", () => this.invalidateAllResources());
  }

  private handleWatchResourceResponse(res: V1WatchResourcesResponse) {
    // only process for the `ResourceKind` present in `UsedResourceKinds`
    if (!res.name?.kind || !res.resource || !UsedResourceKinds[res.name.kind]) {
      if (res.event === V1ResourceEvent.RESOURCE_EVENT_DELETE && res.name) {
        fileArtifacts.resourceDeleted(res.name);
      }
      return;
    }

    const instanceId = get(runtime).instanceId;
    if (
      MainResourceKinds[res.name.kind] &&
      this.shouldSkipResource(instanceId, res.resource)
    ) {
      return;
    }

    // Reconcile does a soft delete 1st by populating deletedOn
    // We then get an event with DELETE after reconcile ends, but without a resource object.
    // So we need to check for deletedOn to be able to use resource.meta, especially the filePaths
    const isSoftDelete = !!res.resource?.meta?.deletedOn;

    if (import.meta.env.VITE_PLAYWRIGHT_TEST) {
      console.log(
        `[${res.resource.meta?.reconcileStatus}] ${res.name.kind}/${res.name.name}`,
      );
    }

    // invalidations will wait until the re-fetched query is completed
    // so, we should not `await` here
    if (!isSoftDelete) {
      void this.invalidateResource(instanceId, res.resource);
    } else {
      this.invalidateRemovedResource(instanceId, res.resource);
    }

    // only re-fetch list queries for kinds in `MainResources`
    if (!MainResourceKinds[res.name.kind]) return;
    return queryClient.refetchQueries(
      // we only use individual kind's queries
      getRuntimeServiceListResourcesQueryKey(instanceId, {
        kind: res.name.kind,
      }),
    );
  }

  private shouldSkipResource(instanceId: string, res: V1Resource) {
    switch (res.meta?.reconcileStatus) {
      case V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED:
        return true;

      case V1ReconcileStatus.RECONCILE_STATUS_PENDING:
      case V1ReconcileStatus.RECONCILE_STATUS_RUNNING:
        void throttledRefreshResource(queryClient, instanceId, res);
        fileArtifacts.updateReconciling(res);
        return true;
    }

    return false;
  }

  private invalidateResource(instanceId: string, resource: V1Resource) {
    if (!resource.meta) return;
    void throttledRefreshResource(queryClient, instanceId, resource);

    const lastStateUpdatedOn = fileArtifacts.getFileArtifact(
      resource.meta?.filePaths?.[0] ?? "",
    ).lastStateUpdatedOn;
    if (
      resource.meta.reconcileStatus !==
        V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
      !lastStateUpdatedOn
    ) {
      // When a resource is created it can send an event with status = IDLE just before it is queued for reconcile.
      // So handle the case when it is 1st queued and status != IDLE
      fileArtifacts.updateLastUpdated(resource);
      return;
    }

    // avoid refreshing for cases where event is sent for a resource that has not changed since we last saw it
    if (
      resource.meta.reconcileStatus !==
        V1ReconcileStatus.RECONCILE_STATUS_IDLE ||
      lastStateUpdatedOn === resource.meta.stateUpdatedOn
    )
      return;

    if (this.shouldInvalidateOLAPTables(resource)) {
      void queryClient.invalidateQueries(
        getConnectorServiceOLAPListTablesQueryKey({
          instanceId: get(runtime).instanceId,
          connector:
            resource.source?.spec?.sinkConnector ??
            resource.model?.spec?.connector ??
            "",
        }),
      );
    }
    fileArtifacts.updateArtifacts(resource);
    this.updateForResource(resource);
    const failed = !!resource.meta.reconcileError;

    const name = resource.meta?.name?.name ?? "";
    let table: string | undefined;
    switch (resource.meta.name?.kind) {
      case ResourceKind.Source:
      case ResourceKind.Model:
        table = resource.source?.state?.table ?? resource.model?.state?.table;
        if (table && resource.meta.name?.name === table)
          // make sure table is populated
          return invalidateProfilingQueries(queryClient, name, failed);
        break;

      case ResourceKind.MetricsView:
        return invalidateMetricsViewData(queryClient, name, failed);

      case ResourceKind.Component:
        return invalidateChartData(queryClient, name, failed);

      case ResourceKind.Dashboard:
      // TODO
    }
  }

  private invalidateRemovedResource(instanceId: string, resource: V1Resource) {
    const name = resource.meta?.name?.name ?? "";
    queryClient.setQueryData(
      getRuntimeServiceGetResourceQueryKey(instanceId, {
        "name.name": name,
        "name.kind": resource.meta?.name?.kind,
      }),
      {
        resource: undefined,
      },
    );
    fileArtifacts.softDeleteResource(resource);
    // cancel queries to make sure any pending requests are cancelled.
    // There could still be some errors because of the race condition between a view/table deleted and we getting the event
    switch (resource?.meta?.name?.kind) {
      case ResourceKind.Source:
      case ResourceKind.Model:
        void queryClient.cancelQueries({
          predicate: (query) => isProfilingQuery(query, name),
        });
        void queryClient.invalidateQueries(
          getConnectorServiceOLAPListTablesQueryKey(),
        );
        break;
      case ResourceKind.MetricsView:
        void queryClient.cancelQueries({
          predicate: (query) => invalidationForMetricsViewData(query, name),
        });
        break;
    }
  }

  private invalidateAllResources() {
    return queryClient.refetchQueries({
      predicate: (query) =>
        query.queryHash.includes(`v1/instances/${get(runtime).instanceId}`),
    });
  }

  private shouldInvalidateOLAPTables(resource: V1Resource) {
    if (
      resource.meta?.name?.kind !== ResourceKind.Source &&
      resource.meta?.name?.kind !== ResourceKind.Model
    ) {
      return false;
    }

    const newTable =
      resource.model?.state?.table ?? resource.source?.state?.table ?? "";
    return resource.meta?.filePaths?.some(
      (f) => this.tables.get(f) !== newTable,
    );
  }

  private updateForResource(resource: V1Resource) {
    if (
      // we only need the data for sources and model right now.
      // ignore the rest of the kinds
      resource.meta?.name?.kind !== ResourceKind.Source &&
      resource.meta?.name?.kind !== ResourceKind.Model
    ) {
      return false;
    }

    const newTable =
      resource.model?.state?.table ?? resource.source?.state?.table ?? "";
    resource.meta?.filePaths?.forEach((filePath) => {
      this.tables.set(filePath, newTable);
    });
  }
}
