<script lang="ts">
  import { goto } from "$app/navigation";
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileCreateSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { uploadTableFiles } from "./file-upload";

  export let showDropOverlay: boolean;

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceNames = useSourceFileNames(runtimeInstanceId);
  $: modelNames = useModelFileNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const uploadedFiles = uploadTableFiles(
      Array.from(e?.dataTransfer?.files),
      [$sourceNames?.data, $modelNames?.data],
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
      } catch (err) {
        console.error(err);
      }

      // Navigate to source page
      goto(`/source/${tableName}`);
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
