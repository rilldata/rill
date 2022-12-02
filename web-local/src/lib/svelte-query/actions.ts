import { goto } from "$app/navigation";
import {
  RpcStatus,
  runtimeServiceGetCatalogEntry,
  runtimeServiceListFiles,
  runtimeServicePutFileAndReconcile,
  V1DeleteFileAndReconcileResponse,
  V1ReconcileError,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getNextEntityName } from "@rilldata/web-local/common/utils/getNextEntityId";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { notifications } from "@rilldata/web-local/lib/components/notifications";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import {
  getFileFromName,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import {
  MutationFunction,
  useMutation,
  UseMutationOptions,
} from "@sveltestack/svelte-query";
import { getName } from "../../common/utils/incrementName";
import { generateMeasuresAndDimension } from "../application-state-stores/metrics-internal-store";

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
      fromPath: getFileFromName(fromName, type),
      toPath: getFileFromName(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  goto(getRouteFromName(toName, type), {
    replaceState: true,
  });
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
  return invalidateAfterReconcile(queryClient, instanceId, resp);
}

export async function deleteFileArtifact(
  queryClient: QueryClient,
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
    if (activeEntity?.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }

    notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });

    return invalidateAfterReconcile(queryClient, instanceId, resp);
  } catch (err) {
    console.error(err);
  }
}

export interface CreateDashboardFromSourceRequest {
  instanceId?: string;
  sourceName?: string;
}

export interface CreateDashboardFromSourceResponse {
  affectedPaths?: string[];
  errors?: V1ReconcileError[];
  dashboardName: string;
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

    // not ideal that this doesn't come from the useQuery cache
    const existingModelFiles = await runtimeServiceListFiles(data.instanceId, {
      glob: "models/*.sql",
    });
    const existingModelNames = existingModelFiles.paths?.map((path) =>
      path.replace("/models/", "").replace(".sql", "")
    );
    const newModelName = getName(
      `${data.sourceName}_model`,
      existingModelNames
    );

    await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: `models/${newModelName}.sql`,
      blob: `select * from ${data.sourceName}`,
    });

    // second, create dashboard from model

    // not ideal that this doesn't come from the useQuery cache
    const existingDashboardFiles = await runtimeServiceListFiles(
      data.instanceId,
      {
        glob: "dashboards/*.yaml",
      }
    );
    const existingDashboardNames = existingDashboardFiles.paths?.map((path) =>
      path.replace("/dashboards/", "").replace(".yaml", "")
    );
    const newDashboardName = getName(
      `${newModelName}_dashboard`,
      existingDashboardNames
    );

    const model = await runtimeServiceGetCatalogEntry(
      data.instanceId,
      newModelName
    );
    const generatedYAML = generateMeasuresAndDimension(model.entry.model, {
      display_name: `${data.sourceName} dashboard`,
      description: `A dashboard automatically generated from the ${data.sourceName} source.`,
    });

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: `dashboards/${newDashboardName}.yaml`,
      blob: generatedYAML,
      create: true,
      createOnly: true,
      strict: false,
    });

    return {
      affectedPaths: response?.affectedPaths,
      errors: response?.errors,
      dashboardName: newDashboardName,
    };
  };

  return useMutation<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >(mutationFn, mutationOptions);
};
