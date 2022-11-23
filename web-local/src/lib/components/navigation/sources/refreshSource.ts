import type { RuntimeState } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
import { sourceUpdated } from "@rilldata/web-local/lib/redux-store/source/source-apis";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-local/lib/util/file-upload";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function refreshSource(
  connector: string,
  tableName: string,
  runtimeState: RuntimeState,
  refreshSource: UseMutationResult,
  createSource: UseMutationResult
) {
  if (connector === "file") {
    const files = await openFileUploadDialog(false);
    if (!files.length) return Promise.reject();

    overlay.set({ title: `Importing ${tableName}` });
    const filePath = await uploadFile(
      `${config.database.runtimeUrl}/v1/repos/${runtimeState.repoId}/objects/file`,
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
      await createSource.mutateAsync({
        instanceId: runtimeState.instanceId,
        data: {
          repoId: runtimeState.repoId,
          instanceId: runtimeState.instanceId,
          path: `sources/${tableName}.yaml`,
          blob: yaml,
          create: true,
          strict: true,
        },
      });
    }
  } else {
    overlay.set({ title: `Importing ${tableName}` });
    await refreshSource.mutateAsync({
      instanceId: runtimeState.instanceId,
      name: tableName,
    });
  }
  return sourceUpdated(tableName);
}
