<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServicePutFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores.js";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
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

  const createSource = useRuntimeServicePutFileAndMigrate();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$persistentModelStore.entities, $persistentTableStore.entities],
      $runtimeStore,
      // adding to match flow elsewhere
      persistentTableStore
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
        $createSource.mutate(
          {
            data: {
              instanceId: runtimeInstanceId,
              path: `sources/${tableName}.yaml`,
              blob: yaml,
              create: true,
              createOnly: true,
              strict: true,
            },
          },
          {
            onSuccess: async () => {
              dispatch("close");
              goto(`/source/${tableName}`);
              queryClient.invalidateQueries(
                getRuntimeServiceListFilesQueryKey($runtimeStore.instanceId)
              );
            },
            onError: async () => {
              overlay.set(null);
              dispatch("close");
            },
          }
        );
      } catch (err) {
        console.error(err);
      }
    }
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
