import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  RpcStatus,
  runtimeServiceListFiles,
  runtimeServicePutFileAndReconcile,
  V1DeleteFileAndReconcileResponse,
  V1PutFileAndReconcileRequest,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getNextEntityName } from "@rilldata/web-local/common/utils/getNextEntityId";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import notifications from "@rilldata/web-local/lib/components/notifications";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import {
  getFileFromName,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-local/lib/util/entity-mappers";
import type { UseMutationResult } from "@sveltestack/svelte-query";
import {
  MutationFunction,
  useMutation,
  UseMutationOptions,
} from "@sveltestack/svelte-query";
import { getName } from "../../common/utils/incrementName";

export async function renameFileArtifact(
  instanceId: string,
  fromName: string,
  toName: string,
  type: EntityType,
  renameMutation: UseMutationResult<V1RenameFileAndReconcileResponse>
) {
  const resp = await renameMutation.mutateAsync({
    data: {
      instanceId,
      fromPath: getFileFromName(fromName, type),
      toPath: getFileFromName(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  await dataModelerService.dispatch("renameEntity", [type, fromName, toName]);
  goto(getRouteFromName(toName, type), {
    replaceState: true,
  });
  notifications.send({
    message: `renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
  await queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(instanceId)
  );
}

export async function deleteFileArtifact(
  instanceId: string,
  name: string,
  type: EntityType,
  deleteMutation: UseMutationResult<V1DeleteFileAndReconcileResponse>,
  activeEntity: ActiveEntity,
  names: Array<string>
) {
  try {
    const resp = await deleteMutation.mutateAsync({
      data: {
        instanceId,
        path: getFileFromName(name, type),
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
    if (activeEntity.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }
    // Temporary until nodejs is removed
    await dataModelerService.dispatch("deleteEntity", [type, name]);

    // TODO: update all entities based on affected path
    return queryClient.invalidateQueries(
      getRuntimeServiceListFilesQueryKey(instanceId)
    );
  } catch (err) {
    console.error(err);
  }
}

// Option 1: vanilla function
// TODO: pass mutations into here (or call the mutationFns directly)
export async function createDashboardFromSource(
  instanceId: string,
  sourceName: string
) {
  // TODO: filter results for names in the right format
  const existingModelNames = await runtimeServiceListFiles(instanceId, {
    glob: "models/*.sql",
  });

  const newModelName = getName(`${sourceName}_model`, existingModelNames.paths);

  // create model from source
  await $createFileMutation.mutateAsync({
    data: {
      instanceId: instanceId,
      path: `models/${newModelName}.sql`,
      blob: `select * from ${sourceName}`,
      create: true,
      createOnly: true,
      strict: true,
    },
  });

  // TODO: filter results for names in the right format
  const existingDashboardNames = await runtimeServiceListFiles(instanceId, {
    glob: "dashboards/*.yaml",
  });

  const newDashboardName = getName(
    `${newModelName}_dashboard`,
    existingDashboardNames.paths
  );

  // create dashboard from model
  await $createFileMutation.mutateAsync({
    data: {
      instanceId: instanceId,
      path: `dashboards/${newDashboardName}.yaml`,
      blob: metricsTemplate, // TODO: compile a real yaml file
      create: true,
      createOnly: true,
      strict: false,
    },
  });

  return newDashboardName;
}

export interface CreateDashboardFromSourceRequest {
  instanceId?: string;
  sourceName?: string;
}

// Option 2: Custom hook
export const useCreateDashboardFromSource = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof runtimeServicePutFileAndReconcile>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof runtimeServicePutFileAndReconcile>>,
    { data: CreateDashboardFromSourceRequest }
  > = async (props) => {
    const { data } = props ?? {};

    // TODO: filter results for names in the right format
    const existingModelNames = await runtimeServiceListFiles(data.instanceId, {
      glob: "models/*.sql",
    });

    const newModelName = getName(
      `${data.sourceName}_model`,
      existingModelNames.paths
    );

    await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: `models/${newModelName}.sql`,
      blob: `select * from ${data.sourceName}`,
      create: true,
      createOnly: true,
      strict: true,
    });

    // TODO: filter results for names in the right format
    const existingDashboardNames = await runtimeServiceListFiles(
      data.instanceId,
      { glob: "dashboards/*.yaml" }
    );

    const newDashboardName = getName(
      `${newModelName}_dashboard`,
      existingDashboardNames.paths
    );

    // compose the request for the dashboard file
    return runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: `dashboards/${newDashboardName}.yaml`,
      blob: metricsTemplate, // TODO: compile a real yaml file
      create: true,
      createOnly: true,
      strict: false,
    });
  };

  return useMutation<
    Awaited<ReturnType<typeof runtimeServicePutFileAndReconcile>>,
    TError,
    { data: V1PutFileAndReconcileRequest },
    TContext
  >(mutationFn, mutationOptions);
};
