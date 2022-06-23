<script lang="ts">
  import SectionDragHandle from "./SectionDragHandle.svelte";
  import { layout } from "$lib/application-state-stores/layout-store";
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
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
  import { metricsDefinitionsApi } from "$lib/redux-store/metricsDefinitionsApi";
  import { measuresApi } from "$lib/redux-store/measuresApi";
  import { dimensionsApi } from "$lib/redux-store/dimensionsApi";

  const {
    endpoints: { getOneMetricsDefinition, updateMetricsDefinition },
  } = metricsDefinitionsApi;
  const {
    endpoints: { getAllMeasures, createMeasure, updateMeasure },
  } = measuresApi;
  const {
    endpoints: { getAllDimensions, createDimension, updateDimension },
  } = dimensionsApi;

  export let metricsDefId;

  let innerHeight;
  const tableContainerDivClass =
    "rounded border border-gray-200 border-2  overflow-auto ";

  let selectedMetricsDef: MetricsDefinitionEntity;
  $: ({ data: selectedMetricsDef } =
    getOneMetricsDefinition.select(metricsDefId)($reduxReadable));

  let measures: Array<MeasureDefinitionEntity>;
  let dimensions: Array<DimensionDefinitionEntity>;
  $: if (metricsDefId) {
    store.dispatch(getOneMetricsDefinition.initiate(metricsDefId));
    store.dispatch(getAllMeasures.initiate(metricsDefId));
    store.dispatch(getAllDimensions.initiate(metricsDefId));
  }
  $: if (metricsDefId)
    ({ data: measures } = getAllMeasures.select(metricsDefId)($reduxReadable));
  $: if (metricsDefId)
    ({ data: dimensions } =
      getAllDimensions.select(metricsDefId)($reduxReadable));

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let derivedModelColumns: Array<ProfileColumn>;
  $: if (selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    derivedModelColumns = $derivedModelStore?.entities.find(
      (model) => model.id === selectedMetricsDef.sourceModelId
    ).profile;
  } else {
    derivedModelColumns = [];
  }

  function updateMetricsDefinitionHandler(
    metricsDef: Partial<MetricsDefinitionEntity>
  ) {
    store.dispatch(
      updateMetricsDefinition.initiate({
        id: metricsDefId,
        metricsDef,
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
          value={selectedMetricsDef?.sourceModelId}
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
          value={selectedMetricsDef?.timeDimension}
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
        rows={measures ?? []}
        columnNames={MeasuresColumns}
        on:change={(evt) => {
          store.dispatch(
            updateMeasure.initiate({
              id: measures[evt.detail.index].id,
              measure: {
                [evt.detail.name]: evt.detail.value,
              },
            })
          );
        }}
        on:add={() => {
          store.dispatch(createMeasure.initiate(selectedMetricsDef?.id));
        }}
      />
    </div>
  </div>

  <SectionDragHandle />

  <div style:height="{$layout.modelPreviewHeight}px" class="p-6 ">
    <div class={tableContainerDivClass + " h-full"}>
      <PreviewTable
        rows={dimensions ?? []}
        columnNames={DimensionColumns}
        on:change={(evt) => {
          store.dispatch(
            updateDimension.initiate({
              id: dimensions[evt.detail.index].id,
              dimension: {
                [evt.detail.name]: evt.detail.value,
              },
            })
          );
        }}
        on:add={() => {
          store.dispatch(createDimension.initiate(selectedMetricsDef?.id));
        }}
      />
    </div>
  </div>
</div>
