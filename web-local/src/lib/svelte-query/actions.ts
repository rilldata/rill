import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import { EntityType } from "@rilldata/web-common/lib/entity";
import {
  RpcStatus,
  runtimeServiceGetCatalogEntry,
  runtimeServicePutFileAndReconcile,
  useRuntimeServiceListFiles,
  V1DeleteFileAndReconcileResponse,
  V1ReconcileError,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import type { ActiveEntity } from "@rilldata/web-local/lib/application-state-stores/app-store";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import {
  invalidateAfterReconcile,
  removeModelQueries,
} from "@rilldata/web-local/lib/svelte-query/invalidation";
import {
  getFilePathFromNameAndType,
  getLabel,
  getNameFromFile,
  getRouteFromName,
} from "@rilldata/web-local/lib/util/entity-mappers";
import { getNextEntityName } from "@rilldata/web-local/lib/util/getNextEntityId";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import {
  MutationFunction,
  useMutation,
  UseMutationOptions,
} from "@sveltestack/svelte-query";
import {
  addQuickMetricsToDashboardYAML,
  initBlankDashboardYAML,
} from "../application-state-stores/metrics-internal-store";

export function useAllNames(instanceId: string) {
  return useRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        select: (data) =>
          data.paths?.map((path) => getNameFromFile(path)) ?? [],
      },
    }
  );
}

export function isDuplicateName(name: string, names: Array<string>) {
  return names.findIndex((n) => n.toLowerCase() === name.toLowerCase()) >= 0;
}

export async function renameFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  fromName: string,
  toName: string,
  type: EntityType,
  renameMutation: UseMutationResult<V1RenameFileAndReconcileResponse>
) {
  const resp = await renameMutation.mutateAsync({
    data: {
      instanceId,
      fromPath: getFilePathFromNameAndType(fromName, type),
      toPath: getFilePathFromNameAndType(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });

  removeModelQueries(queryClient, instanceId, fromName);
  invalidateAfterReconcile(queryClient, instanceId, resp);
}

export async function deleteFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  name: string,
  type: EntityType,
  deleteMutation: UseMutationResult<V1DeleteFileAndReconcileResponse>,
  activeEntity: ActiveEntity,
  names: Array<string>,
  showNotification = true
) {
  try {
    const resp = await deleteMutation.mutateAsync({
      data: {
        instanceId,
        path: getFilePathFromNameAndType(name, type),
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });
    }

    if (type == EntityType.Model)
      removeModelQueries(queryClient, instanceId, name);

    invalidateAfterReconcile(queryClient, instanceId, resp);
    if (activeEntity?.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }
  } catch (err) {
    console.error(err);
  }
}

export interface CreateDashboardFromSourceRequest {
  instanceId: string;
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

    // first, create model from source

    await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: getFilePathFromNameAndType(data.newModelName, EntityType.Model),
      blob: `select * from ${data.sourceName}`,
    });

    // second, create dashboard from model

    const model = await runtimeServiceGetCatalogEntry(
      data.instanceId,
      data.newModelName
    );
    const blankDashboardYAML = initBlankDashboardYAML(data.newDashboardName);
    const fullDashboardYAML = addQuickMetricsToDashboardYAML(
      blankDashboardYAML,
      model.entry.model
    );

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
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
