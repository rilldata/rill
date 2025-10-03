<script lang="ts">
  import { goto, invalidate } from "$app/navigation";
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { uploadTableFiles } from "./file-upload";

  export let showDropOverlay: boolean;

  $: ({ instanceId } = $runtime);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const files = e?.dataTransfer?.files;

    // no-op if no files are dropped
    if (files === undefined) return;

    const uploadedFiles = uploadTableFiles(Array.from(files), instanceId);

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
    }
  };
</script>

<Overlay bg="rgba(0,0,0,.6)">
  <div
    class="w-screen h-screen grid place-content-center"
    on:dragenter|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:drag|preventDefault|stopPropagation
    on:drop|preventDefault|stopPropagation={handleSourceDrop}
    on:mouseup|preventDefault|stopPropagation={() => {
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
