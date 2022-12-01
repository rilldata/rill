<script lang="ts">
  import { useRuntimeServicePutFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-local/lib/util/file-upload";
  import { createEventDispatcher } from "svelte";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();

  $: runtimeInstanceId = $runtimeStore.instanceId;

  const createSourceMutation = useRuntimeServicePutFileAndReconcile();

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
        // TODO: errors
        await createSource(
          runtimeInstanceId,
          tableName,
          yaml,
          $createSourceMutation
        );
      } catch (err) {
        console.error(err);
      }
      overlay.set(null);
      dispatch("close");
    }
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
