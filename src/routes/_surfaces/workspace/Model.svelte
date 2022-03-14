<script lang="ts">
    import { getContext } from "svelte";
    import { slide } from "svelte/transition";
    import type { ApplicationStore } from "$lib/app-store";
    import { dataModelerService } from "$lib/app-store";
    import { cubicOut as easing } from "svelte/easing";
    import Editor from "$lib/components/Editor.svelte";

    import PreviewTable from "$lib/components/table/PreviewTable.svelte";
    import type {
        PersistentModelEntity
    } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
    import type {
        DerivedModelEntity
    } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
    import { DerivedModelStore, PersistentModelStore } from "$lib/modelStores";
    import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
    import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
    import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

    const store = getContext("rill:app:store") as ApplicationStore;
const queryHighlight = getContext("rill:app:query-highlight");
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;
const derivedModelStore = getContext('rill:app:derived-model-store') as DerivedModelStore;

let errorLineNumber;

let showPreview = true;

let currentModel: PersistentModelEntity;
$: currentModel = ($store?.activeEntity && $persistentModelStore?.entities) ?
    $persistentModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
let currentDerivedModel: DerivedModelEntity;
$: currentDerivedModel = ($store?.activeEntity && $derivedModelStore?.entities) ?
    $derivedModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
let error: string;

async function handleModelQueryUpdate(query: string) {
    dataModelerService.dispatch('setActiveAsset', [EntityType.Model, currentModel.id]);
    const response = await dataModelerService.dispatch(
        'updateModelQuery', [currentModel.id, query ]);
    if (response?.status !== ActionStatus.Failure) {
        error = "";
        return;
    }

    const errorMessage = response.messages[0];
    if (errorMessage.errorType === ActionErrorType.QueryCancelled) {
        error = "";
        return;
    }
    error = errorMessage.message;
}

</script>

<div class="editor-pane">
  <div>
  {#if $store && $persistentModelStore?.entities && $derivedModelStore?.entities && currentModel}
    <div class="input-body p-6 pt-0 overflow-auto">
      {#key currentModel?.id}
        <Editor 
          content={currentModel.query}
          name={currentModel.name}
          selections={$queryHighlight}
          errorLineNumber={ currentModel.id === $store.activeEntity.id ? errorLineNumber : undefined }
          on:down={() => { dataModelerService.dispatch('moveModelDown', [currentModel.id]); }}
          on:up={() => { dataModelerService.dispatch('moveModelUp', [currentModel.id]); }}
          on:delete={() => { dataModelerService.dispatch('deleteModel', [currentModel.id]); }}
          on:receive-focus={() => {
              dataModelerService.dispatch('setActiveAsset', [EntityType.Model, currentModel.id]);
              error = "";
          }}
          on:release-focus={() => {
            //dataModelerService.dispatch('releaseActiveQueryFocus', [{ id: q.id }]);
          }}
          on:model-profile={() => {
            //dataModelerService.dispatch('computeModelProfile', [{ id: currentQuery.id }]);
          }}
          on:rename={(evt) => {
            dataModelerService.dispatch('updateModelName', [currentModel.id, evt.detail]);
          }}
          on:write={(evt) => {
              handleModelQueryUpdate(evt.detail.content);
          }}
      />
    {/key}
    {#if error}
      <div transition:slide={{ duration: 200, easing }} 
        class="error p-4 m-4 rounded-lg shadow-md"
      >
        {error}
      </div>
    {/if}
    </div>
  {/if}
</div>
<!-- Show the model output preview -->
{#if currentModel}
  <div style:font-size="12px" class="overflow-hidden h-full grid p-5" style:grid-template-rows="max-content auto">
    <button on:click={() => { showPreview = !showPreview }}>{#if showPreview}hide{:else}show{/if} preview</button>
    <div class="rounded overflow-auto border border border-gray-300 {!showPreview && 'hidden'}"
    >
      {#if currentDerivedModel?.preview && currentDerivedModel?.profile}
        <PreviewTable rows={currentDerivedModel.preview} columnNames={currentDerivedModel.profile} />
      {:else}
        <div class="p-5  grid items-center justify-center italic">no columns selected</div>
      {/if}
    </div>
  </div>
  {/if}
</div>
<style>

.editor-pane {
  display: grid;
  grid-template-rows: auto 400px;
  height: calc(100vh - var(--header-height));
}
.error {
  background-color: var(--error-bg);
  color: var(--error-text);
  font-size: 13px;
  align-self: bottom;
}
</style>
