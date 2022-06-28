<script lang="ts">
  import SectionDragHandle from "./SectionDragHandle.svelte";
  import { layout } from "$lib/application-state-stores/layout-store";
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import { MeasuresColumns } from "$lib/components/metrics-definition/MeasuresColumns";
  import { DimensionColumns } from "$lib/components/metrics-definition/DimensionColumns";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import {
    createDimensionsApi,
    updateDimensionsApi,
  } from "$lib/redux-store/dimension-definition/dimension-definition-apis.js";
  import { selectDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { selectMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-selectors";
  import {
    createMeasuresApi,
    updateMeasuresApi,
  } from "$lib/redux-store/measure-definition/measure-definition-apis.js";
  import MetricsDefinitionGenerateButton from "$lib/components/metrics-definition/MetricsDefinitionGenerateButton.svelte";

  export let metricsDefId;

  let innerHeight;
  const tableContainerDivClass =
    "rounded border border-gray-200 border-2  overflow-auto ";

  $: measures = selectMeasuresByMetricsId(metricsDefId)($reduxReadable);
  $: dimensions = selectDimensionsByMetricsId(metricsDefId)($reduxReadable);

  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  function handleCreateMeasure() {
    store.dispatch(createMeasuresApi({ metricsDefId }));
  }
  function handleUpdateDimension(evt) {
    store.dispatch(
      updateMeasuresApi({
        id: measures[evt.detail.index].id,
        changes: {
          [evt.detail.name]: evt.detail.value,
        },
      })
    );
  }
  function handleCreateDimension() {
    store.dispatch(createDimensionsApi({ metricsDefId }));
  }
  function handleUpdateMeasure(evt) {
    store.dispatch(
      updateDimensionsApi({
        id: dimensions[evt.detail.index].id,
        changes: {
          [evt.detail.name]: evt.detail.value,
        },
      })
    );
  }
</script>

<svelte:window bind:innerHeight />

<div class="editor-pane bg-gray-100">
  <div
    style:height="calc({innerHeight}px - {$layout.modelPreviewHeight}px -
    var(--header-height))"
    style="display: flex; flex-flow: column;"
    class="p-6 pt-0"
  >
    <div>
      <MetricsDefModelSelector {metricsDefId} />
      <MetricsDefTimeColumnSelector {metricsDefId} />
      <MetricsDefinitionGenerateButton {metricsDefId} />
    </div>
    <div style:flex="1" class={tableContainerDivClass}>
      <PreviewTable
        rows={measures ?? []}
        columnNames={MeasuresColumns}
        on:change={handleUpdateMeasure}
        on:add={handleCreateMeasure}
      />
    </div>
  </div>

  <SectionDragHandle />

  <div style:height="{$layout.modelPreviewHeight}px" class="p-6 ">
    <div class={tableContainerDivClass + " h-full"}>
      <PreviewTable
        rows={dimensions ?? []}
        columnNames={DimensionColumns}
        on:change={handleUpdateDimension}
        on:add={handleCreateDimension}
      />
    </div>
  </div>
</div>
