import type { V1PutFileAndReconcileResponse } from "@rilldata/web-common/runtime-client";
import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-local/lib/util/file-upload";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function refreshSource(
  connector: string,
  sourceName: string,
  instanceId: string,
  refreshSource: UseMutationResult,
  createSource: UseMutationResult<V1PutFileAndReconcileResponse>
) {
  if (connector !== "file") {
    overlay.set({ title: `Importing ${sourceName}` });
    await refreshSource.mutateAsync({
      data: {
        instanceId,
        path: `sources/${sourceName}.yaml`,
      },
    });
    return;
  }

  // different logic for the file connector

  const files = await openFileUploadDialog(false);
  if (!files.length) return Promise.reject();

  overlay.set({ title: `Importing ${sourceName}` });
  const filePath = await uploadFile(
    `${config.database.runtimeUrl}/v1/instances/${instanceId}/files/upload`,
    files[0]
  );
  if (filePath) {
    const yaml = compileCreateSourceYAML(
      {
        sourceName: sourceName,
        path: filePath,
      },
      "file"
    );
    const resp = await createSource.mutateAsync({
      data: {
        instanceId,
        path: `sources/${sourceName}.yaml`,
        blob: yaml,
        create: true,
        strict: true,
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  }
}
