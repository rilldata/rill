<script lang="ts">
  import SectionDragHandle from "./SectionDragHandle.svelte";
  import { layout } from "$lib/application-state-stores/layout-store";
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  import {
    createReadableStoreWithSelector,
    store,
  } from "$lib/redux-store/store-root";
  import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { getContext } from "svelte";
  import {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type { ProfileColumn } from "$lib/types";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { MeasuresColumns } from "$lib/components/metrics-definition/MeasuresColumns";
  import { DimensionColumns } from "$lib/components/metrics-definition/DimensionColumns";
  import {
    fetchManyMeasuresApi,
    createMeasuresApi,
    updateMeasuresApi,
    manyMeasuresSelector,
  } from "$lib/redux-store/measure-definition-slice";
  import type { Readable } from "svelte/store";
  import {
    fetchManyDimensionsApi,
    manyDimensionsSelector,
  } from "$lib/redux-store/dimension-definition-slice";
  import {
    createDimensionsApi,
    updateDimensionsApi,
  } from "$lib/redux-store/dimension-definition-slice.js";
  import {
    singleMetricsDefSelector,
    updateMetricsDefsApi,
  } from "$lib/redux-store/metrics-definition-slice";

  export let metricsDefId;

  let innerHeight;
  const tableContainerDivClass =
    "rounded border border-gray-200 border-2  overflow-auto ";

  let selectedMetricsDef: Readable<MetricsDefinitionEntity>;

  let measures: Readable<Array<MeasureDefinitionEntity>>;
  let dimensions: Readable<Array<DimensionDefinitionEntity>>;
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
    selectedMetricsDef = createReadableStoreWithSelector(
      singleMetricsDefSelector(metricsDefId)
    );
    measures = createReadableStoreWithSelector(
      manyMeasuresSelector(metricsDefId)
    );
    dimensions = createReadableStoreWithSelector(
      manyDimensionsSelector(metricsDefId)
    );
  }

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let derivedModelColumns: Array<ProfileColumn>;
  $: if ($selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    derivedModelColumns = $derivedModelStore?.entities.find(
      (model) => model.id === $selectedMetricsDef.sourceModelId
    ).profile;
  } else {
    derivedModelColumns = [];
  }

  function updateMetricsDefinitionHandler(
    metricsDef: Partial<MetricsDefinitionEntity>
  ) {
    store.dispatch(
      updateMetricsDefsApi({
        id: metricsDefId,
        changes: metricsDef,
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
      <div style:height="40px">
        <select
          class="pl-1 mb-2"
          value={$selectedMetricsDef?.sourceModelId}
          on:change={(evt) => {
            updateMetricsDefinitionHandler({ sourceModelId: evt.target.value });
          }}
        >
          {#each $persistentModelStore?.entities as entity}
            <option value={entity.id}>{entity.name}</option>
          {/each}
        </select>
      </div>
      <div style:height="40px">
        <select
          class="pl-1 mb-2"
          value={$selectedMetricsDef?.timeDimension}
          on:change={(evt) => {
            updateMetricsDefinitionHandler({ timeDimension: evt.target.value });
          }}
        >
          {#each derivedModelColumns as column}
            <option value={column.name}>{column.name}</option>
          {/each}
        </select>
      </div>
    </div>
    <div style:flex="1" class={tableContainerDivClass}>
      <PreviewTable
        rows={$measures ?? []}
        columnNames={MeasuresColumns}
        on:change={(evt) => {
          store.dispatch(
            updateMeasuresApi({
              id: $measures[evt.detail.index].id,
              changes: {
                [evt.detail.name]: evt.detail.value,
              },
            })
          );
        }}
        on:add={() => {
          store.dispatch(createMeasuresApi({ metricsDefId }));
        }}
      />
    </div>
  </div>

  <SectionDragHandle />

  <div style:height="{$layout.modelPreviewHeight}px" class="p-6 ">
    <div class={tableContainerDivClass + " h-full"}>
      <PreviewTable
        rows={$dimensions ?? []}
        columnNames={DimensionColumns}
        on:change={(evt) => {
          store.dispatch(
            updateDimensionsApi({
              id: $dimensions[evt.detail.index].id,
              changes: {
                [evt.detail.name]: evt.detail.value,
              },
            })
          );
        }}
        on:add={() => {
          store.dispatch(createDimensionsApi({ metricsDefId }));
        }}
      />
    </div>
  </div>
</div>
