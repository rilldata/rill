<script lang="ts">
  import { getContext } from "svelte";
  import { EntityType } from "../../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "../../../application-state-stores/application-store";
  import { sourceStore } from "../../../application-state-stores/source-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "../../../application-state-stores/table-stores";
  import { TableSourceType } from "../../../types";
  import { fetchWrapper } from "../../../util/fetchWrapper";
  import Editor from "../../Editor.svelte";
  import PreviewTable from "../../preview-table/PreviewTable.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceId: string;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  $: currentSource = $persistentTableStore?.entities
    ? $persistentTableStore.entities.find((q) => q.id === sourceId)
    : undefined;
  $: currentDerivedSource = $derivedTableStore?.entities
    ? $derivedTableStore.entities.find((q) => q.id === sourceId)
    : undefined;

  const switchToSource = async (sourceId: string) => {
    if (!sourceId) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Table,
      sourceId,
    ]);
  };

  $: switchToSource(sourceId);

  // TODO: this should probably post to an API that writes file to disk
  // const saveSourceArtifactToDisk = async (editorContent: string) => {
  //   if (currentSource) {
  //     await dataModelerService.saveSourceArtifactToDisk(editorContent);
  //   }
  // };

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

  function createOrReplaceTableWithSql(content: any): void {
    fetchWrapper("file/create-table-with-sql", "POST", {
      tableName: currentSource.tableName, // TODO: I should do this by ID, just to avoid all confusion
      sql: content,
    }).catch((error) => {
      console.error(error);
    });
  }

  const updateSqlDraft = (evt) => {
    sourceStore.update((state) => {
      state.entities[currentSource?.id] = {
        id: currentSource?.id,
        sqlDraft: evt.detail.content,
      };
      return state;
    });
  };

  $: sqlDraft =
    $sourceStore.entities[currentSource?.id]?.sqlDraft || currentSource?.sql;
</script>

<div
  class="grid pb-6"
  style:grid-template-rows="max-content auto"
  style:height="100vh"
>
  <SourceWorkspaceHeader id={currentSource?.id} />
  {#if currentSource?.sourceType === TableSourceType.SQL}
    {#key currentSource?.id}
      <Editor
        content={sqlDraft}
        showSubmitButton={true}
        canSubmit={true}
        on:write={(evt) => updateSqlDraft(evt)}
        on:submit-query={(evt) =>
          createOrReplaceTableWithSql(evt.detail.content)}
      />
    {/key}
  {/if}
  <div
    style:overflow="auto"
    style:height="100%"
    class="m-3 border border-gray-300 rounded"
  >
    {#if currentDerivedSource}
      {#key currentDerivedSource.id}
        {#if currentSource?.sqlError}
          <div class="p-4">{currentSource.sqlError}</div>
        {:else if currentDerivedSource.preview}
          <PreviewTable
            rows={currentDerivedSource?.preview}
            columnNames={currentDerivedSource?.profile}
          />
        {/if}
      {/key}
    {/if}
  </div>
</div>
