<script lang="ts">
  import { goto } from "$app/navigation";
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileCreateSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { uploadTableFiles } from "./file-upload";

  export let showDropOverlay: boolean;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceNames = useSourceFileNames(runtimeInstanceId);
  $: modelNames = useModelFileNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const files = e?.dataTransfer?.files;

    // no-op if no files are dropped
    if (files === undefined) return;

    const uploadedFiles = uploadTableFiles(
      Array.from(files),
      [$sourceNames?.data ?? [], $modelNames?.data ?? []],
      $runtime.instanceId
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
        checkSourceImported(
          queryClient,
          tableName,
          getFilePathFromNameAndType(tableName, EntityType.Table)
        );
        goto(`/source/${tableName}`);
      } catch (err) {
        console.error(err);
      }
    }
  };
</script>

<Overlay bg="rgba(0,0,0,.6)">
  <div
    role="presentation"
    class="w-screen h-screen grid place-content-center"
    on:dragenter|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:drag|preventDefault|stopPropagation
    on:drop|preventDefault|stopPropagation={handleSourceDrop}
    on:mouseup|preventDefault|stopPropagation={() => {
      showDropOverlay = false;
    }}
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
