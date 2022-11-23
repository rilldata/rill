<script lang="ts">
  import { Callout } from "../../callout";

  import { CATEGORICALS } from "../../../duckdb-data-types";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { initDimensionColumns } from "../../metrics-definition/DimensionColumns";
  import { initMeasuresColumns } from "../../metrics-definition/MeasuresColumns";
  import LayoutManager from "../../metrics-definition/MetricsDesignerLayoutManager.svelte";
  import type { SelectorOption } from "../../table-editable/ColumnConfig";
  import MetricsDefEntityTable from "./MetricsDefEntityTable.svelte";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";

  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import MetricsDefWorkspaceHeader from "./MetricsDefWorkspaceHeader.svelte";
  import {
    getRuntimeServiceGetFileQueryKey,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { createInternalRepresentation } from "./metrics-internal-store";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";

  // the runtime yaml string
  export let yaml: string;
  export let metricsDefName: string;
  export let nonStandardError;

  $: instanceId = $runtimeStore.instanceId;

  const metricMigrate = useRuntimeServicePutFileAndReconcile();
  function callPutAndMigrate(internalYamlString) {
    $metricMigrate.mutate({
      data: {
        instanceId,
        path: `dashboards/${metricsDefName}.yaml`,
        blob: internalYamlString,
        create: false,
      },
    });

    queryClient.invalidateQueries(
      getRuntimeServiceGetFileQueryKey(
        instanceId,
        `dashboards/${metricsDefName}.yaml`
      )
    );
  }

  let metricsInternalRep = createInternalRepresentation(
    yaml,
    callPutAndMigrate
  );

  // reset internal representation in case of deviation from runtime YAML
  // $: if (yaml !== $metrics.internalYAML) {
  //   metrics = createInternalRepresentation(yaml);
  // }

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  $: modelName = $metricsInternalRep.getMetricKey("from");
  $: getModel = useRuntimeServiceGetCatalogEntry(instanceId, modelName);
  $: model = $getModel.data?.entry?.model;

  $: console.log(model);

  function handleCreateMeasure() {
    $metricsInternalRep.addNewMeasure();
  }
  function handleUpdateMeasure(index, name, value) {
    $metricsInternalRep.updateMeasure(index, name, value);
  }

  function handleDeleteMeasure(evt) {
    $metricsInternalRep.deleteMeasure(evt.detail);
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
    $metricsInternalRep.addNewDimension();
  }
  function handleUpdateDimension(index, name, value) {
    $metricsInternalRep.updateDimension(index, name, value);
  }
  function handleDeleteDimension(evt) {
    $metricsInternalRep.deleteDimension(evt.detail);
  }

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
      <MetricsDefWorkspaceHeader {metricsDefName} {metricsInternalRep} />

      <div
        class="editor-pane bg-gray-100 p-6 pt-0 flex flex-col"
        style:height="calc(100vh - var(--header-height))"
      >
        <div class="flex-none flex flex-row">
          <div>
            <MetricsDefModelSelector {metricsInternalRep} />
            <MetricsDefTimeColumnSelector {metricsInternalRep} />
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
