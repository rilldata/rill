import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import {
  ResourceKind,
  createSchemaForTable,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { generateDashboardYAMLForModel } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
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
import { derived } from "svelte/store";

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
    const dashboardYAML = generateDashboardYAMLForModel(
      data.modelName,
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
