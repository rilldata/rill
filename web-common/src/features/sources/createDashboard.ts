import { goto } from "$app/navigation";
import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
import { useSource } from "@rilldata/web-common/features/sources/selectors";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { overlay } from "@rilldata/web-common/layout/overlay-store";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import {
  connectorServiceOLAPGetTable,
  RpcStatus,
  runtimeServicePutFile,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  createMutation,
  CreateMutationOptions,
  MutationFunction,
  useQueryClient,
} from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { generateDashboardYAMLForModel } from "../metrics-views/metrics-internal-store";

export interface CreateDashboardFromSourceRequest {
  instanceId: string;
  sourceResource: V1Resource;
  newModelName: string;
  newDashboardName: string;
}

export const useCreateDashboardFromSource = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<Promise<void>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};
  const queryClient = mutationOptions?.queryClient ?? useQueryClient();

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: CreateDashboardFromSourceRequest }
  > = async (props) => {
    const { data } = props ?? {};
    const sourceName = data.sourceResource?.meta?.name?.name;
    if (!sourceName)
      throw new Error("Failed to create dashboard: Source name is missing");
    if (
      !data.sourceResource.source.state.connector ||
      !data.sourceResource.source.state.table
    )
      throw new Error("Failed to create dashboard: Source is not ready");

    // first, create model from source

    await runtimeServicePutFile(
      data.instanceId,
      getFileAPIPathFromNameAndType(data.newModelName, EntityType.Model),
      {
        blob: `select * from ${sourceName}`,
        create: true,
        createOnly: true,
      }
    );

    // second, create dashboard from model
    const sourceSchema = await connectorServiceOLAPGetTable({
      instanceId: data.instanceId,
      connector: data.sourceResource.source.state.connector,
      table: data.sourceResource.source.state.table,
    });

    const dashboardYAML = generateDashboardYAMLForModel(
      data.newModelName,
      sourceSchema.schema,
      data.newDashboardName
    );

    await runtimeServicePutFile(
      data.instanceId,
      getFileAPIPathFromNameAndType(
        data.newDashboardName,
        EntityType.MetricsDefinition
      ),
      {
        blob: dashboardYAML,
        create: true,
        createOnly: true,
      }
    );
    await waitForResource(
      queryClient,
      data.instanceId,
      getFilePathFromNameAndType(
        data.newDashboardName,
        EntityType.MetricsDefinition
      )
    );
  };

  return createMutation<
    Awaited<Promise<void>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >(mutationFn, mutationOptions);
};

/**
 * Wrapper function that takes care of UI side effects on top of creating a dashboard from source.
 * TODO: where would this go?
 */
export function useCreateDashboardFromSourceUIAction(
  instanceId: string,
  sourceName: string
) {
  const createDashboardFromSourceMutation = useCreateDashboardFromSource();
  const modelNames = useModelFileNames(instanceId);
  const dashboardNames = useDashboardFileNames(instanceId);
  const sourceQuery = useSource(instanceId, sourceName);

  return async () => {
    overlay.set({
      title: "Creating a dashboard for " + sourceName,
    });
    const newModelName = getName(`${sourceName}_model`, get(modelNames).data);
    const newDashboardName = getName(
      `${sourceName}_dashboard`,
      get(dashboardNames).data
    );

    // Wait for source query to have data
    await waitUntil(() => !!get(sourceQuery).data);
    const sourceResource = get(sourceQuery).data;
    if (sourceResource === undefined) {
      // Note: this should never happen, because we wait for the
      // source query to have data
      console.warn("Failed to create dashboard: Source is not ready");
      return;
    }

    try {
      await get(createDashboardFromSourceMutation).mutateAsync({
        data: {
          instanceId,
          sourceResource,
          newModelName,
          newDashboardName,
        },
      });
      goto(`/dashboard/${newDashboardName}`);
      behaviourEvent.fireNavigationEvent(
        newDashboardName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        get(appScreen)?.type,
        MetricsEventScreenName.Dashboard
      );
      overlay.set(null);
    } catch (err) {
      overlay.set({
        title: "Failed to create a dashboard for " + sourceName,
        message: err.response?.data?.message ?? err.message,
      });
    }
  };
}
