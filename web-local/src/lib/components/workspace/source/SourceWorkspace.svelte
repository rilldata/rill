<script lang="ts">
  import type { DerivedTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import { getContext } from "svelte";
  import { EntityType } from "../../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "../../../application-state-stores/application-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "../../../application-state-stores/table-stores";
  import PreviewTable from "../../preview-table/PreviewTable.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceId: string;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  let currentSource: PersistentTableEntity;
  let currentDerivedSource: DerivedTableEntity;

  const switchToSource = async (sourceId: string) => {
    if (!sourceId) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Table,
      sourceId,
    ]);

    currentSource = $persistentTableStore?.entities
      ? $persistentTableStore.entities.find((q) => q.id === sourceId)
      : undefined;

    currentDerivedSource = $derivedTableStore?.entities
      ? $derivedTableStore.entities.find((q) => q.id === sourceId)
      : undefined;
  };

  $: switchToSource(sourceId);

  /** check to see if we need to perform a migration.
   * We will deprecate this in a few versions from 0.8.
   */

  let profiling = false;
  $: if (currentDerivedSource && !profiling) {
    const previewRowCount = currentDerivedSource?.preview?.length;
    /** migration point from 0.7 ~ upgrade active source to have more rows in preview */
    if (previewRowCount === 1 || currentDerivedSource?.cardinality !== 1) {
      profiling = true;
      dataModelerService.dispatch("refreshPreview", [
        currentSource.id,
        currentSource.tableName,
      ]);
    }
  }
</script>

<div
  class="grid pb-6"
  style:grid-template-rows="max-content auto"
  style:height="100vh"
>
  <SourceWorkspaceHeader id={currentSource?.id} />
  <div
    style:overflow="auto"
    style:height="100%"
    class="m-3 border border-gray-300 rounded"
  >
    {#if currentDerivedSource}
      {#key currentDerivedSource.id}
        <PreviewTable
          rows={currentDerivedSource?.preview}
          columnNames={currentDerivedSource?.profile}
        />
      {/key}
    {/if}
  </div>
</div>
