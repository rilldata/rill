<script lang="ts">
  import { goto, invalidate } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/modal/file-upload";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();

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
              olap: "duckdb", // Explicitly set DuckDB as OLAP for local file uploads
            },
          });

          // Race condition: invalidate("init") must be called before we navigate to
          // `/files/${newFilePath}`. invalidate("init") is also called in the
          // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
          await invalidate("init");
        }

        const yaml = compileLocalFileSourceYAML(filePath);
        await createSource(instanceId, tableName, yaml);
        const newFilePath = getFilePathFromNameAndType(
          tableName,
          EntityType.Table,
        );
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
  <Button onClick={handleOpenFileDialog} type="primary"
    >Upload a CSV, JSON or Parquet file
  </Button>
</div>
<div class="flex p-6">
  <div class="grow" />
  <Button onClick={() => dispatch("back")} type="secondary">Back</Button>
</div>
