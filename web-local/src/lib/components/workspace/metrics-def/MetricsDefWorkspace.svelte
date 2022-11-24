<script lang="ts">
  import type { DerivedModelStore } from "../../../application-state-stores/model-stores";
  import { Callout } from "../../callout";

  import { getContext } from "svelte";
  import { CATEGORICALS } from "../../../duckdb-data-types";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
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
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { createInternalRepresentation } from "./metrics-internal-store";
  // import { $metricsresentation } from "./$metricsresentation";

  export let metricsDefName;
  export let nonStandardError;

  // the runtime yaml string
  export let yaml;

  // the local copy of the yaml string
  // let $metrics = new $metricsresentation(yaml);

  let metrics = createInternalRepresentation(yaml);

  // reset internal representation in case of deviation from runtime YAML
  $: if (yaml !== $metrics.internalYAML) {
    metrics = createInternalRepresentation(yaml);
  }

  $: repoId = $runtimeStore.repoId;
  $: instanceId = $runtimeStore.instanceId;

  const metricMigrate = useRuntimeServicePutFileAndMigrate();
  function callPutAndMigrate() {
    $metricMigrate.mutate({
      data: {
        repoId,
        instanceId,
        path: `dashboards/${metricsDefName}.yaml`,
        blob: $metrics.internalYAML,
        create: false,
      },
    });
  }

  $: measures = $metrics.getMeasures();

  $: console.log("measures", measures);
  $: dimensions = $metrics.getDimensions();

  $: model_path = $metrics.getMetricKey("model_path");
  $: getModel = useRuntimeServiceGetCatalogEntry(instanceId, model_path);
  $: model = $getModel.data?.object?.model;

  function handleCreateMeasure() {
    $metrics.addNewMeasure();
    callPutAndMigrate();
  }
  function handleUpdateMeasure(index, name, value) {
    $metrics.updateMeasure(index, name, value);
    callPutAndMigrate();
  }

  function handleDeleteMeasure(evt) {
    $metrics.deleteMeasure(evt.detail);
    callPutAndMigrate();

    // invalidateMetricsView(queryClient, metricsDefId);
  }
  function handleMeasureExpressionValidation(index, name, value) {
    // store.dispatch(
    //   validateMeasureExpressionApi({
    //     metricsDefId: metricsDefId,
    //     measureId: $measures[index].id,
    //     expression: value,
    //   })
    // );
  }

  function handleCreateDimension() {
    $metrics.addNewDimension();
    callPutAndMigrate();
  }
  function handleUpdateDimension(index, name, value) {
    $metrics.updateDimension(index, name, value);
    callPutAndMigrate();
  }
  function handleDeleteDimension(evt) {
    $metrics.deleteDimension(evt.detail);
    callPutAndMigrate();
    // invalidateMetricsView(queryClient, metricsDefId);
  }

  // FIXME: the only data that is needed from the derived model store is the data types of the
  // columns in this model. I need to make this available in the redux store.
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let validDimensionSelectorOption: SelectorOption[] = [];
  $: if (model) {
    const selectedMetricsDefModelProfile = model?.schema?.fields ?? [];
    validDimensionSelectorOption = selectedMetricsDefModelProfile
      .filter((column) => CATEGORICALS.has(column.type as string))
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

  // $: metricsSourceSelectionError = $selectedMetricsDef
  //   ? MetricsSourceSelectionError($selectedMetricsDef)
  //   : nonStandardError
  //   ? nonStandardError
  //   : "";

  $: metricsSourceSelectionError = nonStandardError ? nonStandardError : "";
</script>

{#if measures && dimensions}
  <WorkspaceContainer inspector={false} assetID={`${metricsDefName}-config`}>
    <div slot="body">
      <MetricsDefWorkspaceHeader
        {metricsDefName}
        metricsInternalRep={metrics}
      />

      <div
        class="editor-pane bg-gray-100 p-6 pt-0 flex flex-col"
        style:height="calc(100vh - var(--header-height))"
      >
        <div class="flex-none flex flex-row">
          <div>
            <MetricsDefModelSelector metricsInternalRep={metrics} />
            <MetricsDefTimeColumnSelector metricsInternalRep={metrics} />
          </div>
          <div class="self-center pl-10">
            {#if metricsSourceSelectionError}
              <Callout level="error">
                {metricsSourceSelectionError}
              </Callout>
            {:else}
              <!-- <MetricsDefinitionGenerateButton {metricsDefId} /> -->
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
              rows={measures ?? []}
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
              rows={dimensions ?? []}
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
