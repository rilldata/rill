<script lang="ts">
  import {
    getDuplicateNameChecker,
    getIncrementedNameGetter,
  } from "@rilldata/web-local/common/utils/duplicateNameUtils";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { compileCreateSourceSql } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { uploadFilesWithDialog } from "@rilldata/web-local/lib/util/file-upload";
  import { createEventDispatcher, getContext } from "svelte";
  import { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores.js";
  import { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores.js";
  import { useRuntimeServiceMigrateSingle } from "web-common/src/runtime-client";

  const dispatch = createEventDispatcher();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

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
    dispatch("close");
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleUpload} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
