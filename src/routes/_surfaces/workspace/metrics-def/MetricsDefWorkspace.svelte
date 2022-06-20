<script lang="ts">
  import SectionDragHandle from "./SectionDragHandle.svelte";
  import { layout } from "$lib/application-state-stores/layout-store";
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import type {
    DimensionDefinition,
    MeasureDefinition,
    MetricsDefinitionEntity,
  } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import {
    setDimensions,
    setMeasures,
  } from "$lib/redux-store/metrics-definition/metrics-definition-slice";
  import { getContext } from "svelte";
  import {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import { MetricsDimensionClient } from "$lib/components/metrics-definition/MetricsDimensionClient.js";
  import { MetricsMeasureClient } from "$lib/components/metrics-definition/MetricsMeasureClient";
  import { MetricsDefinitionClient } from "$lib/components/metrics-definition/MetricsDefinitionClient.js";
  import type { ProfileColumn } from "$lib/types";
  import EditableTableCell from "$lib/components/table/EditableTableCell.svelte";
  import type { ColumnConfig } from "$lib/components/table/pinnableUtils";

  export let metricsDefId;

  let innerHeight;
  const tableContainerDivClass =
    "rounded border border-gray-200 border-2  overflow-auto ";

  const MeasuresColumns: Array<ColumnConfig> = [
    "label",
    "sqlName",
    "expression",
    "description",
  ].map((col) => ({
    name: col,
    type: "VARCHAR",
    renderer: EditableTableCell,
  }));
  MeasuresColumns[1].validation = (row: MeasureDefinition) =>
    row.sqlNameIsValid;
  MeasuresColumns[2].validation = (row: MeasureDefinition) =>
    row.expressionIsValid;

  const DimensionColumns: Array<ColumnConfig> = [
    "sqlName",
    "dimensionColumn",
    "description",
  ].map((col) => ({
    name: col,
    type: "VARCHAR",
    renderer: EditableTableCell,
  }));
  DimensionColumns[0].validation = (row: DimensionDefinition) =>
    row.sqlNameIsValid;
  DimensionColumns[1].validation = (row: DimensionDefinition) =>
    row.dimensionIsValid;

  let selectedMetricsDef: MetricsDefinitionEntity;
  $: if ($reduxReadable?.metricsDefinition?.entities[metricsDefId]?.id) {
    selectedMetricsDef =
      $reduxReadable?.metricsDefinition?.entities[metricsDefId];
    if (!("measures" in selectedMetricsDef)) {
      MetricsMeasureClient.instance
        .getAll(selectedMetricsDef.id)
        .then((measures) =>
          store.dispatch(setMeasures(selectedMetricsDef.id, measures))
        );
    }
    if (!("dimensions" in selectedMetricsDef)) {
      MetricsDimensionClient.instance
        .getAll(selectedMetricsDef.id)
        .then((dimensions) =>
          store.dispatch(setDimensions(selectedMetricsDef.id, dimensions))
        );
    }
  }

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
            MetricsDefinitionClient.instance.updateMetricsDefinitionModel(
              metricsDefId,
              evt.target.value
            );
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
            MetricsDefinitionClient.instance.updateMetricsDefinitionTimestamp(
              metricsDefId,
              evt.target.value
            );
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
        rows={selectedMetricsDef?.measures ?? []}
        columnNames={MeasuresColumns}
        on:change={(evt) => {
          MetricsMeasureClient.instance.updateField(
            selectedMetricsDef?.measures[evt.detail.index].id,
            evt.detail.name,
            evt.detail.value,
            selectedMetricsDef?.id
          );
        }}
        on:add={() => {
          MetricsMeasureClient.instance.create(selectedMetricsDef?.id);
        }}
      />
    </div>
  </div>

  <SectionDragHandle />

  <div style:height="{$layout.modelPreviewHeight}px" class="p-6 ">
    <div class={tableContainerDivClass + " h-full"}>
      <PreviewTable
        rows={selectedMetricsDef?.dimensions ?? []}
        columnNames={DimensionColumns}
        on:change={(evt) => {
          MetricsDimensionClient.instance.updateField(
            selectedMetricsDef?.dimensions[evt.detail.index].id,
            evt.detail.name,
            evt.detail.value,
            selectedMetricsDef?.id
          );
        }}
        on:add={() => {
          MetricsDimensionClient.instance.create(selectedMetricsDef?.id);
        }}
      />
    </div>
  </div>
</div>
