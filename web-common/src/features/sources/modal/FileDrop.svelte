<script lang="ts">
  import { invalidate } from "$app/navigation";
  import { navigateToFile } from "@rilldata/web-common/features/workspaces/edit-routing";
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceUnpackEmptyMutation } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { uploadTableFiles } from "./file-upload";

  export let showDropOverlay: boolean;

  const runtimeClient = useRuntimeClient();

  const unpackEmptyProject =
    createRuntimeServiceUnpackEmptyMutation(runtimeClient);

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const files = e?.dataTransfer?.files;

    // no-op if no files are dropped
    if (files === undefined) return;

    const uploadedFiles = uploadTableFiles(Array.from(files), runtimeClient);

    const initialized = await isProjectInitialized(runtimeClient);
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        // If project is uninitialized, initialize an empty project
        if (!initialized) {
          $unpackEmptyProject.mutate({
            displayName: EMPTY_PROJECT_TITLE,
            olap: "duckdb", // Explicitly set DuckDB as OLAP for local file uploads
          });

          // Race condition: invalidate("app:init") must be called before we navigate to
          // `/files/${newFilePath}`. invalidate("app:init") is also called in the
          // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
          await invalidate("app:init");
        }

        const yaml = compileLocalFileSourceYAML(filePath);
        await createSource(runtimeClient, tableName, yaml);
        const newFilePath = getFilePathFromNameAndType(
          tableName,
          EntityType.Table,
        );
        await navigateToFile(newFilePath);
      } catch (err) {
        console.error(err);
      }
    }
  };
</script>

<Overlay bg="rgba(0,0,0,.6)">
  <div
    class="w-screen h-screen grid place-content-center"
    ondragenter={(e) => {
      e.preventDefault();
      e.stopPropagation();
    }}
    ondragleave={(e) => {
      e.preventDefault();
      e.stopPropagation();
    }}
    ondragover={(e) => {
      e.preventDefault();
      e.stopPropagation();
    }}
    ondrag={(e) => {
      e.preventDefault();
      e.stopPropagation();
    }}
    ondrop={(e) => {
      e.preventDefault();
      e.stopPropagation();
      handleSourceDrop(e);
    }}
    onmouseup={(e) => {
      e.preventDefault();
      e.stopPropagation();
      showDropOverlay = false;
    }}
    role="presentation"
  >
    <div
      class="grid place-content-center grid-gap-2 text-white m-auto p-6 break-all text-3xl"
    >
      <span class="place-content-center">
        drop your files to add new source
      </span>
    </div>
  </div>
</Overlay>
