<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import { compileCreateSourceSql } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
  import { getContext } from "svelte";
  import { useRuntimeServiceMigrateSingle } from "web-common/src/runtime-client";
  import Overlay from "./Overlay.svelte";
  import { uploadTableFiles } from "../../util/file-upload";

  export let showDropOverlay: boolean;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const uploadedFiles = uploadTableFiles(
      Array.from(e?.dataTransfer?.files),
      [$persistentModelStore.entities, $persistentTableStore.entities],
      $runtimeStore
    );
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        const sql = compileCreateSourceSql(
          {
            sourceName: tableName,
            path: filePath,
          },
          "file"
        );
        await $createSource.mutateAsync({
          instanceId: runtimeInstanceId,
          data: { sql, createOrReplace: true },
        });
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
