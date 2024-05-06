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
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileCreateSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(files, $runtime.instanceId, false);
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        // If project is uninitialized, initialize an empty project
        if (!$isProjectInitialized.data) {
          $unpackEmptyProject.mutate({
            instanceId: $runtime.instanceId,
            data: {
              title: EMPTY_PROJECT_TITLE,
            },
          });
        }

        const yaml = compileCreateSourceYAML(
          {
            sourceName: tableName,
            path: filePath,
          },
          "local_file",
        );

        await createSource(runtimeInstanceId, tableName, yaml);
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
