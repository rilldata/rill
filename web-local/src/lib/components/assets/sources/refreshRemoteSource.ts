import type { RuntimeState } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { compileCreateSourceSql } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-local/lib/util/file-upload";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function refreshRemoteSource(
  connector: string,
  tableName: string,
  runtimeState: RuntimeState,
  refreshSource: UseMutationResult,
  createSource: UseMutationResult
) {
  try {
    if (connector === "file") {
      const files = await openFileUploadDialog(false);
      if (!files.length) return;
      const filePath = await uploadFile(
        `${config.database.runtimeUrl}/v1/repos/${runtimeState.repoId}/objects/file`,
        files[0]
      );
      if (filePath) {
        const sql = compileCreateSourceSql(
          {
            sourceName: tableName,
            path: filePath,
          },
          "file"
        );
        await createSource.mutateAsync({
          instanceId: runtimeState.instanceId,
          data: { sql, createOrReplace: true },
        });
      }
    } else {
      await refreshSource.mutateAsync({
        instanceId: runtimeState.instanceId,
        name: tableName,
      });
    }
  } catch (err) {
    console.error(err);
  }
}
