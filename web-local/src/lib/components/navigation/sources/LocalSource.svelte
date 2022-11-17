<script lang="ts">
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServicePutFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores.js";
  import type { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores.js";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
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
  $: repoId = $runtimeStore.repoId;

  const createSource = useRuntimeServicePutFileAndMigrate();

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
        const yaml = compileCreateSourceYAML(
          {
            sourceName: tableName,
            path: filePath,
          },
          "file"
        );
        $createSource.mutate({
          data: {
            repoId,
            instanceId: runtimeInstanceId,
            path: `sources/${tableName}.yaml`,
            blob: yaml,
            create: true,
            createOnly: true,
            strict: true,
          },
        });
      } catch (err) {
        console.error(err);
      }
    }
    dispatch("close");
    return queryClient.invalidateQueries(
      getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
    );
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
