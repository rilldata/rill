<script lang="ts">
  import {
    getDuplicateNameChecker,
    getIncrementedNameGetter,
  } from "@rilldata/web-local/common/utils/duplicateNameUtils";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { uploadFilesWithDialog } from "@rilldata/web-local/lib/util/file-upload";
  import { createEventDispatcher, getContext } from "svelte";
  import { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores.js";
  import { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores.js";

  const dispatch = createEventDispatcher();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  async function handleUpload() {
    await uploadFilesWithDialog(
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
        )
    );
    dispatch("close");
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleUpload} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
