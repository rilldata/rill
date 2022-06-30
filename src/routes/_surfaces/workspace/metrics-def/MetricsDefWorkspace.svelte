<script lang="ts">
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  import { store } from "$lib/redux-store/store-root";
  import { MeasuresColumns } from "$lib/components/metrics-definition/MeasuresColumns";
  import { DimensionColumns } from "$lib/components/metrics-definition/DimensionColumns";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import {
    createDimensionsApi,
    updateDimensionsApi,
  } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import {
    createMeasuresApi,
    updateMeasuresApi,
  } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import MetricsDefinitionGenerateButton from "$lib/components/metrics-definition/MetricsDefinitionGenerateButton.svelte";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";

  export let metricsDefId;

  let innerHeight;

  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  function handleCreateMeasure() {
    store.dispatch(createMeasuresApi({ metricsDefId }));
  }
  function handleUpdateMeasure(evt) {
    store.dispatch(
      updateMeasuresApi({
        id: $measures[evt.detail.index].id,
        changes: {
          [evt.detail.name]: evt.detail.value,
        },
      })
    );
  }

  function handleCreateDimension() {
    store.dispatch(createDimensionsApi({ metricsDefId }));
  }
  function handleUpdateDimension(evt) {
    store.dispatch(
      updateDimensionsApi({
        id: $dimensions[evt.detail.index].id,
        changes: {
          [evt.detail.name]: evt.detail.value,
        },
      })
    );
  }

  const tableContainerDivClass =
    "rounded border border-gray-200 border-2 overflow-auto flex-1";
  const h4Class =
    "text-ellipsis overflow-hidden whitespace-nowrap text-gray-400 font-bold uppercase pt-6 pb-2 flex-none";
</script>

<svelte:window bind:innerHeight />

<div
  class="editor-pane bg-gray-100 p-6 pt-0 flex-col-container"
  style:height="calc(100vh - var(--header-height))"
>
  <div class="flex-none">
    <MetricsDefModelSelector {metricsDefId} />
    <MetricsDefTimeColumnSelector {metricsDefId} />
    <MetricsDefinitionGenerateButton {metricsDefId} />
  </div>

  <div
    style="display: flex; flex-direction:column; overflow-y:hidden;"
    class="flex-1"
  >
    <div class="metrics-def-section">
      <h4 class={h4Class}>Measures</h4>
      <div class={tableContainerDivClass}>
        <PreviewTable
          tableConfig={{ enableAdd: true }}
          rows={$measures ?? []}
          columnNames={MeasuresColumns}
          on:change={handleUpdateMeasure}
          on:add={handleCreateMeasure}
        />
      </div>
    </div>

    <div class="metrics-def-section">
      <h4 class={h4Class}>Dimensions</h4>
      <div class={tableContainerDivClass}>
        <PreviewTable
          tableConfig={{ enableAdd: true }}
          rows={$dimensions ?? []}
          columnNames={DimensionColumns}
          on:change={handleUpdateDimension}
          on:add={handleCreateDimension}
        />
      </div>
    </div>
  </div>
</div>

<style>
  .flex-col-container {
    display: flex;
    flex-direction: column;
  }

  .metrics-def-section {
    max-height: 50%;
    overflow-y: hidden;
    display: flex;
    flex-direction: column;
  }
</style>
