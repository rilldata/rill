<script lang="ts">
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    V1PutFileAndReconcileResponse,
    V1ReconcileError,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/lib/temp/errors/ErrorMessages";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createInternalRepresentation } from "../../../application-state-stores/metrics-internal-store";
  import { CATEGORICALS } from "../../../duckdb-data-types";
  import { getFilePathFromNameAndType } from "../../../util/entity-mappers";
  import { Callout } from "../../callout";
  import { initDimensionColumns } from "../../metrics-definition/DimensionColumns";
  import { initMeasuresColumns } from "../../metrics-definition/MeasuresColumns";
  import MetricsDefinitionGenerateButton from "../../metrics-definition/MetricsDefinitionGenerateButton.svelte";
  import LayoutManager from "../../metrics-definition/MetricsDesignerLayoutManager.svelte";
  import type { SelectorOption } from "../../table-editable/ColumnConfig";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import MetricsDefEntityTable from "./MetricsDefEntityTable.svelte";
  import MetricsDefModelSelector from "./MetricsDefModelSelector.svelte";
  import MetricsDefTimeColumnSelector from "./MetricsDefTimeColumnSelector.svelte";
  import MetricsDefWorkspaceHeader from "./MetricsDefWorkspaceHeader.svelte";

  // the runtime yaml string
  export let yaml: string;
  export let metricsDefName: string;
  export let nonStandardError;

  const queryClient = useQueryClient();

  $: instanceId = $runtimeStore.instanceId;

  const switchToMetrics = async (metricsDefName: string) => {
    if (!metricsDefName) return;

    appStore.setActiveEntity(metricsDefName, EntityType.MetricsDefinition);
  };

  $: switchToMetrics(metricsDefName);

  const metricMigrate = useRuntimeServicePutFileAndReconcile();
  async function callPutAndMigrate(internalYamlString) {
    const filePath = getFilePathFromNameAndType(
      metricsDefName,
      EntityType.MetricsDefinition
    );
    const resp = (await $metricMigrate.mutateAsync({
      data: {
        instanceId,
        path: filePath,
        blob: internalYamlString,
        create: false,
      },
    })) as V1PutFileAndReconcileResponse;
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    invalidateAfterReconcile(queryClient, $runtimeStore.instanceId, resp);
  }

  // create initial internal representation
  let metricsInternalRep = createInternalRepresentation(
    yaml,
    callPutAndMigrate
  );

  function updateInternalRep() {
    metricsInternalRep = createInternalRepresentation(yaml, callPutAndMigrate);
    if (errors) $metricsInternalRep.updateErrors(errors);
  }

  // reset internal representation in case of deviation from runtime YAML
  $: if (yaml !== $metricsInternalRep.internalYAML) {
    updateInternalRep();
  }

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  $: modelName = $metricsInternalRep.getMetricKey("model");
  $: getModel = useRuntimeServiceGetCatalogEntry(instanceId, modelName);
  $: model = $getModel.data?.entry?.model;

  function handleCreateMeasure() {
    $metricsInternalRep.addNewMeasure();
  }
  function handleUpdateMeasure(index, name, value) {
    $metricsInternalRep.updateMeasure(index, name, value);
  }

  function handleDeleteMeasure(evt) {
    $metricsInternalRep.deleteMeasure(evt.detail);
  }
  function handleMeasureExpressionValidation(_index, _name, _value) {
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
      .filter((column) => CATEGORICALS.has(column.type.code as string))
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

  let errors: Array<V1ReconcileError>;
  $: errors =
    $fileArtifactsStore.entities[
      getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
    ]?.errors;

  $: metricsSourceSelectionError = nonStandardError
    ? nonStandardError
    : MetricsSourceSelectionError(errors);
</script>

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
          <MetricsDefTimeColumnSelector
            selectedModel={model}
            {metricsInternalRep}
          />
        </div>
        <div class="self-center pl-10">
          {#if metricsSourceSelectionError}
            <Callout level="error">
              {metricsSourceSelectionError}
            </Callout>
          {:else}
            <MetricsDefinitionGenerateButton
              handlePutAndMigrate={callPutAndMigrate}
              selectedModel={model}
              {metricsInternalRep}
            />
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
            tooltipText={"Add a new measure"}
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
            tooltipText={"Add a new dimension"}
            addButtonId={"add-dimension-button"}
          />
        </LayoutManager>
      </div>
    </div>
  </div>
</WorkspaceContainer>
