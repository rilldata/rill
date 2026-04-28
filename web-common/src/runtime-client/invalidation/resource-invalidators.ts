import { connectorExplorerStore } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store";
import { isLeafResource } from "@rilldata/web-common/features/entity-management/dag-utils";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { sourceIngestionTracker } from "@rilldata/web-common/features/sources/sources-store";
import {
  getConnectorServiceOLAPListTablesQueryKey,
  getQueryServiceResolveCanvasQueryKey,
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceGetModelPartitionsQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1ReconcileStatus,
  type V1Resource,
  V1ResourceEvent,
  type V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import {
  invalidateComponentData,
  invalidateConnectorQueries,
  invalidateMetricsViewData,
  invalidateProfilingQueries,
} from "@rilldata/web-common/runtime-client/invalidation";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";

export interface ResourceInvalidatorState {
  /** Tracks connector names that have been seen writing, to detect
   *  implicitly-created connectors (e.g. when a Source declares a new one). */
  connectorNames: Set<string>;
}

export function createResourceInvalidatorState(): ResourceInvalidatorState {
  return {
    connectorNames: new Set<string>(),
  };
}

/**
 * Top-level resource-event handler. Looks up the previous cached resource,
 * writes the new one into the cache, and dispatches to per-kind invalidators.
 */
export async function handleResourceEvent(
  event: V1WatchResourcesResponse,
  queryClient: QueryClient,
  runtimeClient: RuntimeClient,
  state: ResourceInvalidatorState,
): Promise<void> {
  if (!event?.event || !event.name?.name || !event.name.kind) {
    return;
  }

  const { instanceId } = runtimeClient;
  const name = event.name.name;
  const kind = event.name.kind;

  const previousResource = queryClient.getQueryData<{
    resource: V1Resource | undefined;
  }>(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      name: { name, kind },
    }),
  )?.resource;

  queryClient.setQueryData(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      name: { name, kind },
    }),
    { resource: event.resource },
  );

  // The ProjectParser is a synthetic resource that does not map to a file.
  if ((kind as ResourceKind) === ResourceKind.ProjectParser) return;

  // Keep the fileArtifacts client-side cache in sync with the server.
  if (event.event === V1ResourceEvent.RESOURCE_EVENT_WRITE && event.resource) {
    fileArtifacts.updateArtifacts(event.resource);
  } else if (event.event === V1ResourceEvent.RESOURCE_EVENT_DELETE) {
    fileArtifacts.deleteResource(event.name);
  }

  switch (event.event) {
    case V1ResourceEvent.RESOURCE_EVENT_WRITE:
      await dispatchWrite(
        event,
        previousResource,
        queryClient,
        instanceId,
        state,
        runtimeClient,
      );
      return;
    case V1ResourceEvent.RESOURCE_EVENT_DELETE:
      dispatchDelete(event, previousResource, queryClient, instanceId);
      return;
  }
}

async function dispatchWrite(
  event: V1WatchResourcesResponse,
  previousResource: V1Resource | undefined,
  queryClient: QueryClient,
  instanceId: string,
  state: ResourceInvalidatorState,
  runtimeClient: RuntimeClient,
): Promise<void> {
  if (
    !event.resource?.meta?.reconcileStatus ||
    !event.resource.meta.stateVersion
  ) {
    return;
  }

  // Mirror the legacy watcher gate:
  //   - resourceVersionUnchanged: duplicate stateVersion (same server view)
  //   - resourceFinishedReconciling: a non-idle -> idle transition
  // If the version advanced and reconcile is still in progress, skip noisy
  // intermediate invalidations.
  const resourceVersionUnchanged =
    event.resource.meta.stateVersion === previousResource?.meta?.stateVersion;
  const resourceFinishedReconciling =
    previousResource?.meta?.reconcileStatus !==
      V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
    event.resource.meta.reconcileStatus ===
      V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  console.log(
    `[(${previousResource?.meta?.specVersion ?? "0"},${previousResource?.meta?.stateVersion ?? "0"})` +
      `(${event.resource?.meta?.specVersion ?? "0"},${event.resource?.meta?.stateVersion ?? "0"})]` +
      ` ${resourceVersionUnchanged}/${resourceFinishedReconciling}=${!resourceVersionUnchanged && !resourceFinishedReconciling}` +
      ` ${event.name?.kind}/${event.name?.name}`,
  );

  if (!resourceVersionUnchanged && !resourceFinishedReconciling) {
    return;
  }

  // Refetch the two `ListResources` views the UI watches (unscoped and
  // scoped-by-kind).
  void queryClient.refetchQueries({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
  });
  void queryClient.refetchQueries({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, {
      kind: event.name!.kind,
    }),
  });

  const kind = event.name!.kind as ResourceKind;
  const resource = event.resource;
  const failed = !!resource.meta?.reconcileError;

  switch (kind) {
    case ResourceKind.Connector:
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId),
      });
      void invalidateConnectorQueries(
        queryClient,
        instanceId,
        event.name!.name!,
      );
      return;

    case ResourceKind.Source:
    case ResourceKind.Model:
      await invalidateForSourceOrModelWrite(
        event,
        previousResource,
        queryClient,
        instanceId,
        state,
        runtimeClient,
      );
      return;

    case ResourceKind.MetricsView:
      void invalidateMetricsViewData(queryClient, event.name!.name!, failed);
      return;

    case ResourceKind.Explore: {
      const metricsView = resource.explore?.state?.validSpec?.metricsView;
      if (metricsView) {
        void invalidateMetricsViewData(queryClient, metricsView, failed);
      }
      void queryClient.refetchQueries({
        queryKey: getRuntimeServiceGetExploreQueryKey(instanceId, {
          name: event.name!.name,
        }),
      });
      return;
    }

    case ResourceKind.Canvas:
      void queryClient.refetchQueries({
        queryKey: getQueryServiceResolveCanvasQueryKey(instanceId, {
          canvas: event.name!.name,
        }),
      });
      return;

    case ResourceKind.Component:
      void invalidateComponentData(queryClient, event.name!.name!, failed);
      return;

    default:
      return;
  }
}

function dispatchDelete(
  event: V1WatchResourcesResponse,
  previousResource: V1Resource | undefined,
  queryClient: QueryClient,
  instanceId: string,
): void {
  // Refetch the two `ListResources` views the UI watches.
  void queryClient.refetchQueries({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
  });
  void queryClient.refetchQueries({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, {
      kind: event.name!.kind,
    }),
  });

  switch (event.name!.kind as ResourceKind) {
    case ResourceKind.Connector:
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId),
      });
      connectorExplorerStore.deleteItem(event.name!.name!);
      return;

    case ResourceKind.Source:
    case ResourceKind.Model:
      invalidateForSourceOrModelDelete(
        event,
        previousResource,
        queryClient,
        instanceId,
      );
      return;

    default:
      return;
  }
}

async function invalidateForSourceOrModelWrite(
  event: V1WatchResourcesResponse,
  previousResource: V1Resource | undefined,
  queryClient: QueryClient,
  instanceId: string,
  state: ResourceInvalidatorState,
  runtimeClient: RuntimeClient,
): Promise<void> {
  const kind = event.name!.kind as ResourceKind;
  const isSource = kind === ResourceKind.Source;

  // TODO: differentiate between a Model's executorConnector and resultConnector.
  const connectorName = isSource
    ? event.resource!.source?.state?.connector
    : event.resource!.model?.state?.resultConnector;
  const previousConnectorName = isSource
    ? previousResource?.source?.state?.connector
    : previousResource?.model?.state?.resultConnector;

  // If the result table changed, invalidate the (old and new) connector's
  // list of tables so the explorer reflects the change.
  const sourceTableChanged =
    event.resource!.source?.state?.table !==
    previousResource?.source?.state?.table;
  const modelResultTableChanged =
    event.resource!.model?.state?.resultTable !==
    previousResource?.model?.state?.resultTable;
  if (sourceTableChanged || modelResultTableChanged) {
    const connectorsToInvalidate = Array.from(
      new Set([connectorName, previousConnectorName].filter(Boolean)),
    );
    for (const connector of connectorsToInvalidate) {
      void queryClient.invalidateQueries({
        queryKey: getConnectorServiceOLAPListTablesQueryKey(instanceId, {
          connector,
        }),
      });
    }
  }

  // Sources and Models can implicitly create Connectors; pick those up.
  if (connectorName && !state.connectorNames.has(connectorName)) {
    state.connectorNames.add(connectorName);
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId),
    });
  }

  // Note: Sources/Models that fail to ingest have no table name.
  const tableName = isSource
    ? event.resource!.source?.state?.table
    : event.resource!.model?.state?.resultTable;

  if (!connectorName || !tableName) return;

  // Show the "Source imported successfully" modal. Guards protect against
  // existing projects that happen to be reconciling anew: first file version,
  // first ingest complete, tracker says it's pending, and the resource is
  // genuinely a leaf.
  const filePath = event.resource!.meta?.filePaths?.[0];
  const isNewSource =
    kind === ResourceKind.Model &&
    filePath !== undefined &&
    sourceIngestionTracker.isPending(filePath) &&
    event.resource!.meta?.specVersion === "1" &&
    event.resource!.meta?.stateVersion === "2" &&
    (await isLeafResource(event.resource!, runtimeClient));
  if (isNewSource && filePath !== undefined) {
    sourceIngestionTracker.trackIngested(filePath);
  }

  if (kind === ResourceKind.Model) {
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGetModelPartitionsQueryKey(instanceId, {
        model: event.name!.name,
      }),
    });
  }

  const failed = !!event.resource!.meta?.reconcileError;
  void invalidateProfilingQueries(queryClient, tableName, failed);
}

function invalidateForSourceOrModelDelete(
  event: V1WatchResourcesResponse,
  previousResource: V1Resource | undefined,
  queryClient: QueryClient,
  instanceId: string,
): void {
  // The connector is no longer on the (now-deleted) resource — pull it from
  // the previous cached version so we can invalidate its tables list.
  const kind = event.name!.kind as ResourceKind;
  const connectorName =
    kind === ResourceKind.Source
      ? previousResource?.source?.state?.connector
      : previousResource?.model?.state?.resultConnector;

  void queryClient.invalidateQueries({
    queryKey: getConnectorServiceOLAPListTablesQueryKey(instanceId, {
      connector: connectorName,
    }),
  });
}
