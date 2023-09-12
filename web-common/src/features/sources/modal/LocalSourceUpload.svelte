<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/modal/file-upload";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModelNames } from "../../models/selectors";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileCreateSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$sourceNames?.data, $modelNames?.data],
      $runtime.instanceId,
      false
    );
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
          "local_file"
        );

        await createSource(runtimeInstanceId, tableName, yaml);
      } catch (err) {
        // TODO: file write errors
      }

      overlay.set(null);
      dispatch("close");

      // Navigate to source page
      goto(`/source/${tableName}`);
    }
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV, JSON or Parquet file
  </Button>
</div>
