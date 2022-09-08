<script lang="ts">
  import type { DerivedModelStore } from "$lib/application-state-stores/model-stores";
  import { store } from "$lib/redux-store/store-root";
  import { getContext, onMount } from "svelte";

  import { initDimensionColumns } from "$lib/components/metrics-definition/DimensionColumns";
  import { initMeasuresColumns } from "$lib/components/metrics-definition/MeasuresColumns";
  import MetricsDefinitionGenerateButton from "$lib/components/metrics-definition/MetricsDefinitionGenerateButton.svelte";
  import {
    createDimensionsApi,
    deleteDimensionsApi,
    updateDimensionsWrapperApi,
  } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import {
    createMeasuresApi,
    deleteMeasuresApi,
    updateMeasuresWrapperApi,
    validateMeasureExpressionApi,
  } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";

  import { MetricsSourceSelectionError } from "$common/errors/ErrorMessages";
  import { Callout } from "$lib/components/callout";
  import LayoutManager from "$lib/components/metrics-definition/MetricsDesignerLayoutManager.svelte";
  import type { SelectorOption } from "$lib/components/table-editable/ColumnConfig";
  import { CATEGORICALS } from "$lib/duckdb-data-types";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { bootstrapMetricsDefinition } from "$lib/redux-store/metrics-definition/bootstrapMetricsDefinition";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { queryClient } from "$lib/svelte-query/globalQueryClient";
  import { invalidateMetricsView } from "$lib/svelte-query/queries/metrics-view";
  import MetricsDefEntityTable from "./MetricsDefEntityTable.svelte";

  export let metricsDefId;

  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: dimensions = getDimensionsByMetricsId(metricsDefId);
  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  function handleCreateMeasure() {
    store.dispatch(createMeasuresApi({ metricsDefId }));
  }
  function handleUpdateMeasure(index, name, value) {
    store.dispatch(
      updateMeasuresWrapperApi({
        id: $measures[index].id,
        changes: { [name]: value },
      })
    );
  }
  function handleDeleteMeasure(evt) {
    store.dispatch(deleteMeasuresApi(evt.detail));
    invalidateMetricsView(queryClient, metricsDefId);
  }
  function handleMeasureExpressionValidation(index, name, value) {
    store.dispatch(
      validateMeasureExpressionApi({
        metricsDefId: metricsDefId,
        measureId: $measures[index].id,
        expression: value,
      })
    );
  }

  function handleCreateDimension() {
    store.dispatch(createDimensionsApi({ metricsDefId }));
  }
  function handleUpdateDimension(index, name, value) {
    store.dispatch(
      updateDimensionsWrapperApi({
        id: $dimensions[index].id,
        changes: {
          [name]: value,
        },
      })
    );
  }
  function handleDeleteDimension(evt) {
    store.dispatch(deleteDimensionsApi(evt.detail));
    invalidateMetricsView(queryClient, metricsDefId);
  }

  // FIXME: the only data that is needed from the derived model store is the data types of the
  // columns in this model. I need to make this available in the redux store.
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let validDimensionSelectorOption: SelectorOption[] = [];
  $: if ($selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    const selectedMetricsDefModelProfile =
      $derivedModelStore?.entities.find(
        (model) => model.id === $selectedMetricsDef.sourceModelId
      )?.profile ?? [];
    validDimensionSelectorOption = selectedMetricsDefModelProfile
      .filter((column) => CATEGORICALS.has(column.type))
      .map((column) => ({ label: column.name, value: column.name }));
  } else {
    validDimensionSelectorOption = [];
  }

  $: MeasuresColumns = initMeasuresColumns(
    handleUpdateMeasure,
    handleMeasureExpressionValidation
  );
  $: DimensionColumns = initDimensionColumns(
    handleUpdateDimension,
    validDimensionSelectorOption
  );

  $: metricsSourceSelectionError = $selectedMetricsDef
    ? MetricsSourceSelectionError($selectedMetricsDef)
    : "";

  onMount(() => {
    store.dispatch(bootstrapMetricsDefinition(metricsDefId));
  });
</script>

<div
  class="editor-pane bg-gray-100 p-6 pt-0 flex flex-col"
  style:height="calc(100vh - var(--header-height))"
>
  <div class="flex-none flex flex-row">
    <div>
      <MetricsDefModelSelector {metricsDefId} />
      <MetricsDefTimeColumnSelector {metricsDefId} />
    </div>
    <div class="self-center pl-10">
      {#if metricsSourceSelectionError}
        <Callout level="error">
          {metricsSourceSelectionError}
        </Callout>
      {:else}
        <MetricsDefinitionGenerateButton {metricsDefId} />
      {/if}
    </div>
  </div>

  <div
    style="display: flex; flex-direction:column; overflow:hidden;"
    class="flex-1"
  >
    <LayoutManager let:topResizeCallback let:bottomResizeCallback>
      <MetricsDefEntityTable
        slot="top-item"
        resizeCallback={topResizeCallback}
        label={"Measures"}
        addEntityHandler={handleCreateMeasure}
        updateEntityHandler={handleUpdateMeasure}
        deleteEntityHandler={handleDeleteMeasure}
        rows={$measures ?? []}
        columnNames={MeasuresColumns}
        tooltipText={"add a new measure"}
        addButtonId={"add-measure-button"}
      />

      <MetricsDefEntityTable
        slot="bottom-item"
        resizeCallback={bottomResizeCallback}
        label={"Dimensions"}
        addEntityHandler={handleCreateDimension}
        updateEntityHandler={handleUpdateDimension}
        deleteEntityHandler={handleDeleteDimension}
        rows={$dimensions ?? []}
        columnNames={DimensionColumns}
        tooltipText={"add a new dimension"}
        addButtonId={"add-dimension-button"}
      />
    </LayoutManager>
  </div>
</div>
