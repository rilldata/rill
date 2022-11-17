<script lang="ts">
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages";
  import type { DerivedModelStore } from "../../../application-state-stores/model-stores";
  import { Callout } from "../../callout";

  import { getContext, onMount } from "svelte";
  import { CATEGORICALS } from "../../../duckdb-data-types";
  import {
    createDimensionsApi,
    deleteDimensionsApi,
    updateDimensionsWrapperApi,
  } from "../../../redux-store/dimension-definition/dimension-definition-apis";
  import { getDimensionsByMetricsId } from "../../../redux-store/dimension-definition/dimension-definition-readables";
  import {
    createMeasuresApi,
    deleteMeasuresApi,
    updateMeasuresWrapperApi,
    validateMeasureExpressionApi,
  } from "../../../redux-store/measure-definition/measure-definition-apis";
  import { getMeasuresByMetricsId } from "../../../redux-store/measure-definition/measure-definition-readables";
  import { bootstrapMetricsDefinition } from "../../../redux-store/metrics-definition/bootstrapMetricsDefinition";
  import { getMetricsDefReadableById } from "../../../redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "../../../redux-store/store-root";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import { invalidateMetricsView } from "../../../svelte-query/queries/metrics-views/invalidation";
  import { initDimensionColumns } from "../../metrics-definition/DimensionColumns";
  import { initMeasuresColumns } from "../../metrics-definition/MeasuresColumns";
  import MetricsDefinitionGenerateButton from "../../metrics-definition/MetricsDefinitionGenerateButton.svelte";
  import LayoutManager from "../../metrics-definition/MetricsDesignerLayoutManager.svelte";
  import type { SelectorOption } from "../../table-editable/ColumnConfig";
  import MetricsDefEntityTable from "./MetricsDefEntityTable.svelte";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";

  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import MetricsDefWorkspaceHeader from "./MetricsDefWorkspaceHeader.svelte";
  import { useRuntimeServicePutFileAndMigrate } from "@rilldata/web-common/runtime-client";
  export let metricsDefId;
  export let nonStandardError;

  // the runtime yaml string
  export let yaml;

  // the local copy of the yaml string
  let internalYAML = yaml;

  const metricQuery = useRuntimeServicePutFileAndMigrate();

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
    : nonStandardError
    ? nonStandardError
    : "";

  onMount(() => {
    store.dispatch(bootstrapMetricsDefinition(metricsDefId));
  });
</script>

{#if $selectedMetricsDef}
  <WorkspaceContainer inspector={false} assetID={`${metricsDefId}-config`}>
    <div slot="body">
      <MetricsDefWorkspaceHeader {metricsDefId} />

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
    </div>
  </WorkspaceContainer>
{/if}
