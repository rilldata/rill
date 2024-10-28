<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/modal/file-upload";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();

  $: ({ instanceId } = $runtime);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(files, instanceId, false);
    const initialized = await isProjectInitialized(instanceId);
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        // If project is uninitialized, initialize an empty project
        if (!initialized) {
          $unpackEmptyProject.mutate({
            instanceId,
            data: {
              displayName: EMPTY_PROJECT_TITLE,
            },
          });
        }

        const yaml = compileLocalFileSourceYAML(filePath);
        await createSource(instanceId, tableName, yaml);
        const newFilePath = getFilePathFromNameAndType(
          tableName,
          EntityType.Table,
        );
        await checkSourceImported(queryClient, newFilePath);
        await goto(`/files${newFilePath}`);
      } catch (err) {
        console.error(err);
      }

      overlay.set(null);
      dispatch("close");
    }
  }
</script>

<div class="grid place-items-center h-44">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV, JSON or Parquet file
  </Button>
</div>
<div class="flex">
  <div class="grow" />
  <Button on:click={() => dispatch("back")} type="secondary">Back</Button>
</div>
