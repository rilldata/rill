<script lang="ts">
  import { Callout } from "@rilldata/web-common/components/callout";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { CATEGORICALS } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    V1PutFileAndReconcileResponse,
    V1ReconcileError,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { SelectorOption } from "@rilldata/web-local/lib/components/table-editable/ColumnConfig";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/lib/temp/errors/ErrorMessages";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import { initDimensionColumns } from "../DimensionColumns";
  import { initMeasuresColumns } from "../MeasuresColumns";
  import { createInternalRepresentation } from "../metrics-internal-store";
  import MetricsAvailableTimeGrains from "./MetricsAvailableTimeGrains.svelte";
  import MetricsDefaultTimeGrainSelector from "./MetricsDefaultTimeGrainSelector.svelte";
  import MetricsDefaultTimeRange from "./MetricsDefaultTimeRange.svelte";
  import MetricsDisplayNameInput from "./MetricsDisplayNameInput.svelte";
  import MetricsEntityTable from "./MetricsEntityTable.svelte";
  import MetricsGenerateButton from "./MetricsGenerateButton.svelte";
  import LayoutManager from "./MetricsLayoutManager.svelte";
  import MetricsModelSelector from "./MetricsModelSelector.svelte";
  import MetricsTimeColumnSelector from "./MetricsTimeColumnSelector.svelte";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";

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
  <MetricsWorkspaceHeader slot="header" {metricsDefName} {metricsInternalRep} />

  <div slot="body">
    <div
      class="editor-pane bg-gray-100 p-6 flex flex-col"
      style:height="calc(100vh - var(--header-height))"
    >
      <div class="flex-none flex flex-row">
        <div>
          <MetricsDisplayNameInput {metricsInternalRep} />
          <MetricsModelSelector {metricsInternalRep} />
          <MetricsTimeColumnSelector
            selectedModel={model}
            {metricsInternalRep}
          />
        </div>
        <div class="pl-10">
          <MetricsDefaultTimeRange selectedModel={model} {metricsInternalRep} />
          <MetricsDefaultTimeGrainSelector
            selectedModel={model}
            {metricsInternalRep}
          />
          <MetricsAvailableTimeGrains
            selectedModel={model}
            {metricsInternalRep}
          />
        </div>

        <div class="ml-auto">
          {#if metricsSourceSelectionError}
            <Callout level="error">
              {metricsSourceSelectionError}
            </Callout>
          {:else}
            <div>
              <MetricsGenerateButton
                handlePutAndMigrate={callPutAndMigrate}
                selectedModel={model}
                {metricsInternalRep}
              />
            </div>
          {/if}
        </div>
      </div>

      <div
        style="display: flex; flex-direction:column; overflow:hidden;"
        class="flex-1"
      >
        <LayoutManager let:topResizeCallback let:bottomResizeCallback>
          <MetricsEntityTable
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

          <MetricsEntityTable
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
