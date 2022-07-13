<script lang="ts">
  import { store } from "$lib/redux-store/store-root";
  import { getContext } from "svelte";
  import type { DerivedModelStore } from "$lib/application-state-stores/model-stores";

  import { initMeasuresColumns } from "$lib/components/metrics-definition/MeasuresColumns";
  import { initDimensionColumns } from "$lib/components/metrics-definition/DimensionColumns";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";
  import {
    deleteDimensionsApi,
    fetchManyDimensionsApi,
  } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import {
    createDimensionsApi,
    updateDimensionsApi,
  } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import {
    deleteMeasuresApi,
    fetchManyMeasuresApi,
  } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import {
    createMeasuresApi,
    updateMeasuresApi,
  } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import MetricsDefinitionGenerateButton from "$lib/components/metrics-definition/MetricsDefinitionGenerateButton.svelte";

  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import MetricsDefEntityTable from "./MetricsDefEntityTable.svelte";
  import { CATEGORICALS } from "$lib/duckdb-data-types";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";

  export let metricsDefId;

  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: dimensions = getDimensionsByMetricsId(metricsDefId);
  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  // FIXME: this pattern of calling the `fetch*API` from components should
  // be replaced by a call within a thunk fetches the relevant data at the
  // time the active metricsDefId is set in the redux store. (Currently, the
  // active metricsDefId is not available in the redux store, but it sh0uld be)
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  function handleCreateMeasure() {
    store.dispatch(createMeasuresApi({ metricsDefId }));
  }
  function handleUpdateMeasure(index, name, value) {
    store.dispatch(
      updateMeasuresApi({
        id: $measures[index].id,
        changes: { [name]: value },
      })
    );
  }
  function handleDeleteMeasure(evt) {
    store.dispatch(deleteMeasuresApi(evt.detail));
  }

  function handleCreateDimension() {
    store.dispatch(createDimensionsApi({ metricsDefId }));
  }
  function handleUpdateDimension(index, name, value) {
    store.dispatch(
      updateDimensionsApi({
        id: $dimensions[index].id,
        changes: {
          [name]: value,
        },
      })
    );
  }
  function handleDeleteDimension(evt) {
    store.dispatch(deleteDimensionsApi(evt.detail));
  }

  // FIXME: the only data that is needed from the derived model store is the data types of the
  // columns in this model. This data should be available in the redux store.
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let validDimensions: string[] = [];
  $: if ($selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    const selectedMetricsDefModelProfile = $derivedModelStore?.entities.find(
      (model) => model.id === $selectedMetricsDef.sourceModelId
    ).profile;
    validDimensions = selectedMetricsDefModelProfile
      .filter((column) => CATEGORICALS.has(column.type))
      .map((column) => column.name);
  } else {
    validDimensions = [];
  }

  $: MeasuresColumns = initMeasuresColumns(handleUpdateMeasure);
  $: DimensionColumns = initDimensionColumns(
    handleUpdateDimension,
    validDimensions
  );
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
      <MetricsDefinitionGenerateButton {metricsDefId} />
    </div>
  </div>

  <div
    style="display: flex; flex-direction:column; overflow:hidden;"
    class="flex-1"
  >
    <MetricsDefEntityTable
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
      label={"Dimensions"}
      addEntityHandler={handleCreateDimension}
      updateEntityHandler={handleUpdateDimension}
      deleteEntityHandler={handleDeleteDimension}
      rows={$dimensions ?? []}
      columnNames={DimensionColumns}
      tooltipText={"add a new dimension"}
      addButtonId={"add-dimension-button"}
    />
  </div>
</div>
