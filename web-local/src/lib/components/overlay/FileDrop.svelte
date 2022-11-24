<script lang="ts">
  import {
    getRuntimeServiceListCatalogEntriesQueryKey,
    RuntimeServiceListCatalogEntriesType,
    useRuntimeServicePutFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import type { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import { compileCreateSourceYAML } from "@rilldata/web-local/lib/components/navigation/sources/sourceUtils";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { getContext } from "svelte";
  import { uploadTableFiles } from "../../util/file-upload";
  import Overlay from "./Overlay.svelte";

  export let showDropOverlay: boolean;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServicePutFileAndMigrate();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const uploadedFiles = uploadTableFiles(
      Array.from(e?.dataTransfer?.files),
      [$persistentModelStore.entities, $persistentTableStore.entities],
      $runtimeStore,
      persistentTableStore
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
        await $createSource.mutateAsync({
          data: {
            instanceId: runtimeInstanceId,
            path: `sources/${tableName}.yaml`,
            blob: yaml,
            create: true,
            createOnly: true,
            strict: true,
          },
        });
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
