<script lang="ts">
  import { getContext } from "svelte";
  import { EntityType } from "../../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "../../../application-state-stores/application-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "../../../application-state-stores/table-stores";
  import PreviewTable from "../../preview-table/PreviewTable.svelte";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  $: currentSource = $persistentTableStore?.entities
    ? $persistentTableStore.entities.find((q) => q.tableName === sourceName)
    : undefined;
  $: currentDerivedSource = $derivedTableStore?.entities
    ? $derivedTableStore.entities.find((q) => q.id === currentSource.id)
    : undefined;

  const switchToSource = async (sourceID: string) => {
    if (!sourceID) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Table,
      sourceID,
    ]);
  };

  $: switchToSource(currentSource.id);

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

<!-- for now, we will key the entire element on the sourceId. -->
{#key currentSource.id}
  <WorkspaceContainer assetID={sourceName}>
    <div
      slot="body"
      class="grid pb-6"
      style:grid-template-rows="max-content auto"
      style:height="100vh"
    >
      <SourceWorkspaceHeader id={currentSource.id} />
      <div
        style:overflow="auto"
        style:height="100%"
        class="m-6 mt-0 border border-gray-300 rounded"
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
    <SourceInspector sourceID={currentSource.id} slot="inspector" />
  </WorkspaceContainer>
{/key}
