<script lang="ts">
  import {
    getRuntimeServiceListCatalogEntriesQueryKey,
    RuntimeServiceListCatalogEntriesType,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import { useSourceNames } from "@rilldata/web-local/lib/svelte-query/sources";
  import { uploadTableFiles } from "../../util/file-upload";
  import { createSource } from "../navigation/sources/createSource";
  import Overlay from "./Overlay.svelte";

  export let showDropOverlay: boolean;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSourceMutation = useRuntimeServicePutFileAndReconcile();

  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const uploadedFiles = uploadTableFiles(
      Array.from(e?.dataTransfer?.files),
      [sourceNames, modelNames],
      $runtimeStore
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
          runtimeInstanceId,
          tableName,
          yaml,
          $createSourceMutation
        );
      } catch (err) {
        console.error(err);
      }
    }
    return queryClient.invalidateQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(runtimeInstanceId, {
        type: RuntimeServiceListCatalogEntriesType.OBJECT_TYPE_SOURCE,
      })
    );
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
