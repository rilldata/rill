import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  RpcStatus,
  runtimeServiceGetCatalogEntry,
  runtimeServicePutFileAndReconcile,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import {
  MutationFunction,
  useMutation,
  UseMutationOptions,
} from "@sveltestack/svelte-query";
import { get } from "svelte/store";
import { runtime } from "../../runtime-client/runtime-store";
import {
  addQuickMetricsToDashboardYAML,
  initBlankDashboardYAML,
} from "../metrics-views/metrics-internal-store";

export interface CreateDashboardFromSourceRequest {
  sourceName: string;
  newModelName: string;
  newDashboardName: string;
}

export interface CreateDashboardFromSourceResponse {
  affectedPaths?: string[];
  errors?: V1ReconcileError[];
}

export const useCreateDashboardFromSource = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    { data: CreateDashboardFromSourceRequest }
  > = async (props) => {
    const { data } = props ?? {};

    const instanceId = get(runtime).instanceId;

    // first, create model from source

    await runtimeServicePutFileAndReconcile({
      instanceId: instanceId,
      path: getFilePathFromNameAndType(data.newModelName, EntityType.Model),
      blob: `select * from ${data.sourceName}`,
    });

    // second, create dashboard from model

    const model = await runtimeServiceGetCatalogEntry(data.newModelName);
    const blankDashboardYAML = initBlankDashboardYAML(data.newDashboardName);
    const fullDashboardYAML = addQuickMetricsToDashboardYAML(
      blankDashboardYAML,
      model.entry.model
    );

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: instanceId,
      path: getFilePathFromNameAndType(
        data.newDashboardName,
        EntityType.MetricsDefinition
      ),
      blob: fullDashboardYAML,
      create: true,
      createOnly: true,
      strict: false,
    });

    return {
      affectedPaths: response?.affectedPaths,
      errors: response?.errors,
    };
  };

  return useMutation<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >(mutationFn, mutationOptions);
};
