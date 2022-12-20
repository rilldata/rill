import type {
  V1PutFileAndReconcileResponse,
  V1ReconcileResponse,
  V1RefreshAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { humanReadableErrorMessage } from "@rilldata/web-local/lib/components/navigation/sources/errors";
import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
import { notifications } from "@rilldata/web-local/lib/components/notifications";
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
    showNotificationForSourceRefresh(resp, connector, sourceName, artifactPath);
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
  showNotificationForSourceRefresh(
    resp,
    "local_file",
    sourceName,
    artifactPath
  );
  return resp;
}

function showNotificationForSourceRefresh(
  resp: V1ReconcileResponse,
  connector: string,
  name: string,
  artifactPath: string
) {
  const error = resp.errors.find((err) => err.filePath === artifactPath);
  if (error) {
    notifications.send({
      message: `Failed to refresh source ${name}`,
      detail: humanReadableErrorMessage(connector, 3, error.message),
      options: {
        persisted: true,
      },
    });
  } else {
    notifications.send({
      message: `Refreshed source ${name}.`,
    });
  }
}
