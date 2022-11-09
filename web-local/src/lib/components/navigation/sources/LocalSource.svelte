<script lang="ts">
  import {
    getRuntimeServiceListCatalogObjectsQueryKey,
    RuntimeServiceListCatalogObjectsType,
    useRuntimeServiceMigrateSingle,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores.js";
  import { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores.js";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { compileCreateSourceSql } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-local/lib/util/file-upload";
  import { createEventDispatcher, getContext } from "svelte";

  const dispatch = createEventDispatcher();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$persistentModelStore.entities, $persistentTableStore.entities],
      $runtimeStore
    );
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        const sql = compileCreateSourceSql(
          {
            sourceName: tableName,
            path: filePath,
          },
          "file"
        );
        await $createSource.mutateAsync({
          instanceId: runtimeInstanceId,
          data: { sql, createOrReplace: true },
        });
      } catch (err) {
        console.error(err);
      }
    }
    dispatch("close");
    return queryClient.invalidateQueries(
      getRuntimeServiceListCatalogObjectsQueryKey(runtimeInstanceId, {
        type: RuntimeServiceListCatalogObjectsType.TYPE_SOURCE,
      })
    );
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
