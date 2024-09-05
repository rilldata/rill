import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getConnectorServiceOLAPListTablesQueryKey,
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1ReconcileStatus,
  V1ResourceEvent,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import {
  invalidateComponentData,
  invalidateMetricsViewData,
  invalidateProfilingQueries,
} from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import { get } from "svelte/store";
import { connectorExplorerStore } from "../connectors/connector-explorer-store";

export class WatchResourcesClient {
  public readonly client: WatchRequestClient<V1WatchResourcesResponse>;
  private readonly instanceId = get(runtime).instanceId;
  private readonly resourceStateVersions = new Map<string, string>();
  private readonly connectorNames = new Set<string>();
  private readonly softDeletedTableConnectors = new Map<string, string>();

  public constructor() {
    this.client = new WatchRequestClient<V1WatchResourcesResponse>();
    this.client.on("response", (res) => this.handleWatchResourceResponse(res));
    this.client.on("reconnect", () => this.invalidateAllRuntimeQueries());
  }

  private handleWatchResourceResponse(res: V1WatchResourcesResponse) {
    // Log resource status to the browser console during e2e tests. Currently, our e2e tests make assertions
    // based on these logs. However, the e2e tests really should make UI-based assertions.
    if (import.meta.env.VITE_PLAYWRIGHT_TEST) {
      console.log(
        `[${res.resource?.meta?.reconcileStatus}] ${res.name?.kind}/${res.name?.name}`,
      );
    }

    // Type guards
    if (!res?.event || !res?.name || !res?.name?.name || !res?.name?.kind) {
      return;
    }

    // Update the resource in the query cache
    queryClient.setQueryData(
      getRuntimeServiceGetResourceQueryKey(this.instanceId, {
        "name.name": res.name.name,
        "name.kind": res.name.kind,
      }),
      {
        resource: res?.resource,
      },
    );

    // Nothing more to do for the ProjectParser resource
    if ((res.name.kind as ResourceKind) === ResourceKind.ProjectParser) return;

    // Update the file artifacts client-side cache (which maps files to resources)
    switch (res.event) {
      case V1ResourceEvent.RESOURCE_EVENT_WRITE:
        if (res.resource) {
          fileArtifacts.updateArtifacts(res.resource);
        }
        break;
      case V1ResourceEvent.RESOURCE_EVENT_DELETE:
        fileArtifacts.deleteResource(res.name);
        break;
    }

    switch (res.event) {
      case V1ResourceEvent.RESOURCE_EVENT_WRITE: {
        // Type guards
        if (
          !res?.resource ||
          !res?.resource?.meta ||
          !res?.resource?.meta?.reconcileStatus ||
          !res?.resource?.meta?.stateVersion
        ) {
          return;
        }

        // Proceed to query invalidations only when the resource has finished reconciling
        // We know the resource has finished reconciling when:
        //   1) the reconcileStatus is IDLE
        //   2) the state version has been incremented
        if (
          res.resource.meta.reconcileStatus !==
            V1ReconcileStatus.RECONCILE_STATUS_IDLE ||
          this.resourceStateVersions.get(res.name.name) ===
            res.resource.meta.stateVersion
        )
          return;

        // Update our client-side memory of the resource's latest state version
        this.resourceStateVersions.set(
          res.name.name,
          res.resource.meta.stateVersion,
        );

        // Refetch `ListResources` queries
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, undefined),
        );
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, {
            kind: res.name.kind,
          }),
        );

        switch (res.name.kind as ResourceKind) {
          case ResourceKind.Connector:
            // Invalidate the list of connectors
            void queryClient.invalidateQueries(
              getRuntimeServiceAnalyzeConnectorsQueryKey(this.instanceId),
            );

            // Invalidate the connector's list of tables
            void queryClient.invalidateQueries(
              getConnectorServiceOLAPListTablesQueryKey({
                instanceId: this.instanceId,
                connector: res.name.name,
              }),
            );

            // Done
            return;

          case ResourceKind.Source:
          case ResourceKind.Model: {
            // TODO: differentiate between a Source's sourceConnector and sinkConnector
            // TODO: differentiate between a Model's inputConnector, stageConnector, and outputConnector
            const connectorName =
              (res.name.kind as ResourceKind) === ResourceKind.Source
                ? res.resource.source?.spec?.sinkConnector
                : res.resource.model?.spec?.outputConnector;

            // The following invalidations are only needed if the Source/Model has a defined connector
            if (!connectorName) return;

            // If the connector is new, invalidate the list of connectors
            // (This is needed because Sources and Models can implicitly create Connectors)
            if (!this.connectorNames.has(connectorName)) {
              this.connectorNames.add(connectorName);
              void queryClient.invalidateQueries(
                getRuntimeServiceAnalyzeConnectorsQueryKey(this.instanceId),
              );
            }

            // Invalidate the connector's list of tables
            void queryClient.invalidateQueries(
              getConnectorServiceOLAPListTablesQueryKey({
                instanceId: this.instanceId,
                connector: connectorName,
              }),
            );

            // Note: Sources/Models that fail to ingest will not have a table name
            const tableName =
              (res.name.kind as ResourceKind) === ResourceKind.Source
                ? res.resource.source?.state?.table
                : res.resource.model?.state?.resultTable;

            // The following invalidations are only needed if the Source/Model has an active table
            if (!tableName) return;

            // Invalidate profiling queries
            const failed = !!res.resource.meta?.reconcileError;
            void invalidateProfilingQueries(queryClient, tableName, failed);

            // Record the connector name for soft deleted tables, so we can invalidate the
            // connector's list of tables once the table is hard deleted
            const isSoftDelete = !!res.resource.meta?.deletedOn;
            if (isSoftDelete) {
              this.softDeletedTableConnectors.set(tableName, connectorName);
            }

            // Done
            return;
          }

          case ResourceKind.MetricsView: {
            const failed = !!res.resource.meta?.reconcileError;
            void invalidateMetricsViewData(queryClient, res.name.name, failed);

            // Done
            return;
          }

          case ResourceKind.Component: {
            const failed = !!res.resource.meta?.reconcileError;
            void invalidateComponentData(queryClient, res.name.name, failed);

            // Done
            return;
          }

          default:
            // No specific handling for the given resource kind
            return;
        }
      }

      /**
       * Note: Resource "deletes" occur in two stages:
       * 1. A `WRITE` event marks the resource for deletion by setting the `deletedOn` property ("soft" delete).
       * 2. A `DELETE` event signals that the resource has actually been deleted ("hard" delete).
       */
      case V1ResourceEvent.RESOURCE_EVENT_DELETE:
        // Remove the resource from the resource versions map
        this.resourceStateVersions.delete(res.name.name);

        // Refetch `ListResources` queries
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, undefined),
        );
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, {
            kind: res.name.kind,
          }),
        );

        switch (res.name.kind as ResourceKind) {
          case ResourceKind.Connector:
            // Invalidate the list of connectors
            void queryClient.invalidateQueries(
              getRuntimeServiceAnalyzeConnectorsQueryKey(this.instanceId),
            );

            // Remove the connector's state from the connector explorer store
            connectorExplorerStore.deleteItem(res.name.name);

            // Done
            return;

          case ResourceKind.Source:
          case ResourceKind.Model: {
            // Get the connector name
            const connectorName = this.softDeletedTableConnectors.get(
              res.name.name,
            );

            // Invalidate the connector's list of tables
            void queryClient.invalidateQueries(
              getConnectorServiceOLAPListTablesQueryKey({
                instanceId: this.instanceId,
                connector: connectorName,
              }),
            );

            // Remove the soft-deleted table from our record
            this.softDeletedTableConnectors.delete(res.name.name);

            // Done
            return;
          }

          default:
            // No specific handling for the given resource kind
            return;
        }
    }
  }

  private invalidateAllRuntimeQueries() {
    return queryClient.invalidateQueries({
      predicate: (query) =>
        query.queryHash.includes(`v1/instances/${get(runtime).instanceId}`),
    });
  }
}
