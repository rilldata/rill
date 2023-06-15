import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-common/features/sources/add-source/file-upload";
import { compileCreateSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import { overlay } from "@rilldata/web-common/layout/overlay-store";
import type {
  V1PutFileAndReconcileResponse,
  V1RefreshAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
import type {
  CreateBaseMutationResult,
  QueryClient,
} from "@tanstack/svelte-query";

export async function refreshAndReconcile(
  sourceName: string,
  instanceId: string,
  refreshSource: CreateBaseMutationResult<V1RefreshAndReconcileResponse>,
  queryClient: QueryClient,
  path: string,
  displayName = undefined
) {
  overlay.set({ title: `Importing ${displayName || sourceName}` });
  const resp = await refreshSource.mutateAsync({
    data: {
      instanceId,
      path,
    },
  });
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  return resp;
}

export async function refreshSource(
  connector: string,
  sourceName: string,
  instanceId: string,
  refreshSource: CreateBaseMutationResult<V1RefreshAndReconcileResponse>,
  createSource: CreateBaseMutationResult<V1PutFileAndReconcileResponse>,
  queryClient: QueryClient,
  displayName = undefined
) {
  const artifactPath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  if (connector !== "local_file") {
    return refreshAndReconcile(
      sourceName,
      instanceId,
      refreshSource,
      queryClient,
      artifactPath,
      displayName
    );
  }

  // different logic for the file connector

  const files = await openFileUploadDialog(false);
  if (!files.length) return Promise.reject();

  overlay.set({ title: `Importing ${sourceName}` });
  const filePath = await uploadFile(instanceId, files[0]);
  if (filePath === null) {
    return Promise.reject();
  }
  const yaml = compileCreateSourceYAML(
    {
      sourceName,
      path: filePath,
    },
    "local_file"
  );
  const resp = await createSource.mutateAsync({
    data: {
      instanceId,
      path: artifactPath,
      blob: yaml,
      syncFromUrl: true,
      strict: true,
    },
  });
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  return resp;
}
