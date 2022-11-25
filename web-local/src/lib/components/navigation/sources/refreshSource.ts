import type { V1PutFileAndReconcileResponse } from "@rilldata/web-common/runtime-client";
import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { commonEntitiesStore } from "@rilldata/web-local/lib/application-state-stores/common-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-local/lib/util/file-upload";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function refreshSource(
  connector: string,
  tableName: string,
  instanceId: string,
  refreshSource: UseMutationResult,
  createSource: UseMutationResult<V1PutFileAndReconcileResponse>
) {
  if (connector === "file") {
    const files = await openFileUploadDialog(false);
    if (!files.length) return Promise.reject();

    overlay.set({ title: `Importing ${tableName}` });
    const filePath = await uploadFile(
      `${config.database.runtimeUrl}/v1/instances/${instanceId}/files/upload`,
      files[0]
    );
    if (filePath) {
      const yaml = compileCreateSourceYAML(
        {
          sourceName: tableName,
          path: filePath,
        },
        "file"
      );
      const resp = await createSource.mutateAsync({
        instanceId,
        data: {
          instanceId,
          path: `sources/${tableName}.yaml`,
          blob: yaml,
          create: true,
          strict: true,
        },
      });
      commonEntitiesStore.consolidateMigrateResponse(
        resp.affectedPaths,
        resp.errors
      );
    }
  } else {
    overlay.set({ title: `Importing ${tableName}` });
    await refreshSource.mutateAsync({
      instanceId,
      name: tableName,
    });
  }
}
