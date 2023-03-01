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
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";

export async function refreshAndReconcile(
  sourceName: string,
  refreshSource: UseMutationResult<V1RefreshAndReconcileResponse>,
  queryClient: QueryClient,
  path: string,
  displayName = undefined
) {
  overlay.set({ title: `Importing ${displayName || sourceName}` });
  const resp = await refreshSource.mutateAsync({
    data: {
      instanceId: get(runtime).instanceId,
      path,
    },
  });
  invalidateAfterReconcile(queryClient, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  return resp;
}

export async function refreshSource(
  connector: string,
  sourceName: string,
  refreshSource: UseMutationResult<V1RefreshAndReconcileResponse>,
  createSource: UseMutationResult<V1PutFileAndReconcileResponse>,
  queryClient: QueryClient,
  displayName = undefined
) {
  const artifactPath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  if (connector !== "local_file") {
    return refreshAndReconcile(
      sourceName,
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
  const filePath = await uploadFile(files[0]);
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
      instanceId: get(runtime).instanceId,
      path: artifactPath,
      blob: yaml,
      create: true,
      strict: true,
    },
  });
  invalidateAfterReconcile(queryClient, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  return resp;
}
