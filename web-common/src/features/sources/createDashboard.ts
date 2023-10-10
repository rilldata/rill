import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
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
    if (!sourceName) throw new Error("Source name is missing");
    if (
      !data.sourceResource.source.state.connector ||
      !data.sourceResource.source.state.table
    )
      throw new Error("Source is not ready");

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
