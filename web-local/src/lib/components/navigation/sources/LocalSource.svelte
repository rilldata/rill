<script lang="ts">
  import { useRuntimeServicePutFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import { useSourceNames } from "@rilldata/web-local/lib/svelte-query/sources";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-local/lib/util/file-upload";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);

  const createSourceMutation = useRuntimeServicePutFileAndReconcile();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$sourceNames?.data, $modelNames?.data],
      $runtimeStore.instanceId
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
          queryClient,
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
