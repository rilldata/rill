import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  connectorServiceOLAPGetTable,
  RpcStatus,
  runtimeServicePutFile,
  V1PutFileResponse,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  createMutation,
  CreateMutationOptions,
  MutationFunction,
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
    Awaited<Promise<V1PutFileResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<Promise<V1PutFileResponse>>,
    { data: CreateDashboardFromSourceRequest }
  > = async (props) => {
    const { data } = props ?? {};
    const sourceName = data.sourceResource?.meta?.name?.name;
    if (!sourceName) throw new Error("Source name is missing");

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

    return runtimeServicePutFile(
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
  };

  return createMutation<
    Awaited<Promise<V1PutFileResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >(mutationFn, mutationOptions);
};
