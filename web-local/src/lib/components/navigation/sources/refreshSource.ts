import type {
  V1PutFileAndReconcileResponse,
  V1RefreshAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
import { EntityType } from "@rilldata/web-local/lib/temp/entity";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-local/lib/util/file-upload";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import { invalidateAfterReconcile } from "../../../svelte-query/invalidation";
import { getFilePathFromNameAndType } from "../../../util/entity-mappers";

export async function refreshSource(
  connector: string,
  sourceName: string,
  instanceId: string,
  refreshSource: UseMutationResult<V1RefreshAndReconcileResponse>,
  createSource: UseMutationResult<V1PutFileAndReconcileResponse>,
  queryClient: QueryClient
) {
  if (connector !== "local_file") {
    overlay.set({ title: `Importing ${sourceName}` });
    const resp = await refreshSource.mutateAsync({
      data: {
        instanceId,
        path: getFilePathFromNameAndType(sourceName, EntityType.Table),
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
      path: getFilePathFromNameAndType(sourceName, EntityType.Table),
      blob: yaml,
      create: true,
      strict: true,
    },
  });
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  return resp;
}
