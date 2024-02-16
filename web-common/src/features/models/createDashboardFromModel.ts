import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import {
  ResourceKind,
  createSchemaForTable,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { generateDashboardYAMLForTable } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { overlay } from "@rilldata/web-common/layout/overlay-store";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import type { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import {
  RpcStatus,
  V1ReconcileStatus,
  V1StructType,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import {
  CreateMutationOptions,
  MutationFunction,
  QueryClient,
  createMutation,
  useQueryClient,
} from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export interface CreateDashboardFromModelRequest {
  instanceId: string;
  modelName: string;
  schema: V1StructType;
  newDashboardName: string;
}

export const useCreateDashboardFromModel = <
  TError = RpcStatus,
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<Promise<void>>,
    TError,
    { data: CreateDashboardFromModelRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};
  const queryClient = mutationOptions?.queryClient ?? useQueryClient();

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: CreateDashboardFromModelRequest }
  > = async (props) => {
    const { data } = props ?? {};

    // create dashboard from model
    const isModel = true;
    const dashboardYAML = generateDashboardYAMLForTable(
      data.modelName,
      isModel,
      data.schema,
      data.newDashboardName,
    );

    await runtimeServicePutFile(
      data.instanceId,
      getFileAPIPathFromNameAndType(
        data.newDashboardName,
        EntityType.MetricsDefinition,
      ),
      {
        blob: dashboardYAML,
        create: true,
        createOnly: true,
      },
    );
    await waitForResource(
      queryClient,
      data.instanceId,
      getFilePathFromNameAndType(
        data.newDashboardName,
        EntityType.MetricsDefinition,
      ),
    );
  };

  return createMutation<
    Awaited<Promise<void>>,
    TError,
    { data: CreateDashboardFromModelRequest },
    TContext
  >(mutationFn, mutationOptions);
};

export function useModelSchemaIsReady(
  queryClient: QueryClient,
  instanceId: string,
  modelName: string,
) {
  return derived(
    [
      useResource(
        instanceId,
        modelName,
        ResourceKind.Model,
        undefined,
        queryClient,
      ),
      createSchemaForTable(
        instanceId,
        modelName,
        ResourceKind.Model,
        queryClient,
      ),
    ],
    ([model, schema]) => {
      return (
        !model.isFetching &&
        !!model.data &&
        model.data.meta.reconcileStatus ===
          V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
        !schema.isFetching &&
        !!schema.data
      );
    },
  );
}

/**
 * Wrapper function that takes care of UI side effects on top of creating a dashboard from model.
 */
export function useCreateDashboardFromModelUIAction(
  instanceId: string,
  modelName: string,
  queryClient: QueryClient,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
) {
  const createDashboardFromModelMutation = useCreateDashboardFromModel();
  const dashboardNames = useDashboardFileNames(instanceId);
  const schemaForModel = createSchemaForTable(
    instanceId,
    modelName,
    ResourceKind.Model,
    queryClient,
  );

  return async () => {
    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      get(dashboardNames).data,
    );

    try {
      await get(createDashboardFromModelMutation).mutateAsync({
        data: {
          instanceId,
          modelName,
          schema: get(schemaForModel).data?.schema,
          newDashboardName,
        },
      });
      goto(`/dashboard/${newDashboardName}`);
      behaviourEvent.fireNavigationEvent(
        newDashboardName,
        behaviourEventMedium,
        metricsEventSpace,
        get(appScreen)?.type,
        MetricsEventScreenName.Dashboard,
      );
    } catch (err) {
      notifications.send({
        message: "Failed to create a dashboard for " + modelName,
        detail: err.response?.data?.message ?? err.message,
      });
    }
    overlay.set(null);
  };
}
