<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Callout } from "@rilldata/web-common/components/callout";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/add-source/file-upload";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import {
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { useModelNames } from "../../models/selectors";
  import { compileCreateSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { hasDuckDBUnicodeError, niceDuckdbUnicodeError } from "./errors";

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);

  const createSourceMutation = useRuntimeServicePutFileAndReconcile();
  const deleteSource = useRuntimeServiceDeleteFileAndReconcile();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  const handleDeleteSource = async (tableName: string) => {
    await deleteFileArtifact(
      queryClient,
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $deleteSource,
      $appStore.activeEntity,
      $sourceNames.data,
      false
    );
  };

  let errors;

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$sourceNames?.data, $modelNames?.data],
      $runtimeStore.instanceId,
      false
    );
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        const yaml = compileCreateSourceYAML(
          {
            sourceName: tableName,
            path: filePath,
          },
          "local_file"
        );
        // TODO: errors
        errors = await createSource(
          queryClient,
          runtimeInstanceId,
          tableName,
          yaml,
          $createSourceMutation
        );
      } catch (err) {
        // no-op
      }
      overlay.set(null);
      if (!errors?.length) {
        dispatch("close");
      } else {
        // if the upload didn't work, delete the source file.
        handleDeleteSource(tableName);
      }
    }
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV or Parquet file
  </Button>
  {#if errors?.length}
    <div transition:slide={{ duration: LIST_SLIDE_DURATION * 2 }}>
      <Callout level="error">
        <ul style:max-width="400px">
          {#each errors as error}
            <li>
              {hasDuckDBUnicodeError(error.message)
                ? niceDuckdbUnicodeError(error.message)
                : error.message}
            </li>
          {/each}
        </ul>
      </Callout>
    </div>
  {/if}
</div>
