<script lang="ts">
  import {
    getDuplicateNameChecker,
    getIncrementedNameGetter,
  } from "@rilldata/web-local/common/utils/duplicateNameUtils";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import { compileCreateSourceSql } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
  import { getContext } from "svelte";
  import { useRuntimeServiceMigrateSingle } from "web-common/src/runtime-client";
  import Overlay from "./Overlay.svelte";
  import { onSourceDrop } from "../../util/file-upload";

  export let showDropOverlay: boolean;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

  const handleSourceDrop = (e: DragEvent) => {
    showDropOverlay = false;
    onSourceDrop(
      e,
      (name) =>
        getDuplicateNameChecker(
          name,
          $persistentModelStore.entities,
          $persistentTableStore.entities
        ),
      (name) =>
        getIncrementedNameGetter(
          name,
          $persistentModelStore.entities,
          $persistentTableStore.entities
        ),
      async (tableName, filePath) => {
        return new Promise((resolve, reject) => {
          const sql = compileCreateSourceSql(
            {
              sourceName: tableName,
              path: filePath,
            },
            "file"
          );
          $createSource.mutate(
            {
              instanceId: runtimeInstanceId,
              data: { sql, createOrReplace: true },
            },
            {
              onSuccess: resolve,
              onError: reject,
            }
          );
        });
      }
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
