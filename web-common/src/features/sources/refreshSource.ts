import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-common/features/sources/add-source/file-upload";
import { compileCreateSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import { EntityType } from "@rilldata/web-common/lib/entity";
import type {
  V1PutFileAndReconcileResponse,
  V1RefreshAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";

export async function refreshSource(
  connector: string,
  sourceName: string,
  instanceId: string,
  refreshSource: UseMutationResult<V1RefreshAndReconcileResponse>,
  createSource: UseMutationResult<V1PutFileAndReconcileResponse>,
  queryClient: QueryClient
) {
  const artifactPath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  if (connector !== "local_file") {
    overlay.set({ title: `Importing ${sourceName}` });
    const resp = await refreshSource.mutateAsync({
      data: {
        instanceId,
        path: artifactPath,
      },
    });
    invalidateAfterReconcile(queryClient, instanceId, resp);
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
    return resp;
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
      create: true,
      strict: true,
    },
  });
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  return resp;
}
