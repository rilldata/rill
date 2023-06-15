<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-common/layout/app-store";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServicePutFileAndReconcile,
    V1PutFileAndReconcileResponse,
    V1ReconcileError,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/lib/temp/errors/ErrorMessages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onMount, setContext } from "svelte";
  import { writable } from "svelte/store";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import { createResizeListenerActionFactory } from "../../../lib/actions/create-resize-listener-factory";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { initDimensionColumns } from "../DimensionColumns";
  import { initMeasuresColumns } from "../MeasuresColumns";
  import { createInternalRepresentation } from "../metrics-internal-store";
  import ConfigParameters from "./config-parameters/ConfigParameters.svelte";
  import MetricsEntityTable from "./MetricsEntityTable.svelte";
  import LayoutManager from "./MetricsLayoutManager.svelte";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";

  // the runtime yaml string
  export let yaml: string;
  export let metricsDefName: string;
  export let nonStandardError;

  // this store is used to store errors that are not related to the reconciliation/runtime
  // used to prevent the user from going to the dashboard.
  // Ultimately, the runtime should be catching the different errors we encounter with regards to
  // mismatches between the fields. For now, this is a very simple to use solution.
  let configurationErrorStore = writable({
    defaultTimeRange: null,
    smallestTimeGrain: null,
    model: null,
    timeColumn: null,
  });
  setContext("rill:metrics-config:errors", configurationErrorStore);

  $: dashboardConfig = createRuntimeServiceGetCatalogEntry(
    instanceId,
    metricsDefName
  );

  const queryClient = useQueryClient();
  const { listenToNodeResize } = createResizeListenerActionFactory();

  $: instanceId = $runtime.instanceId;

  const switchToMetrics = async (metricsDefName: string) => {
    if (!metricsDefName) return;

    appStore.setActiveEntity(metricsDefName, EntityType.MetricsDefinition);
  };

  $: switchToMetrics(metricsDefName);

  const metricMigrate = createRuntimeServicePutFileAndReconcile();
  async function callReconcileAndUpdateYaml(internalYamlString) {
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

    invalidateAfterReconcile(queryClient, $runtime.instanceId, resp);
  }

  // create initial internal representation
  let metricsInternalRep = createInternalRepresentation(
    yaml,
    callReconcileAndUpdateYaml
  );

  onMount(() => {
    // Reconcile on mount
    callReconcileAndUpdateYaml(yaml);
  });

  async function updateInternalRep() {
    const isDifferent =
      $metricsInternalRep && yaml !== $metricsInternalRep.internalYAML;

    metricsInternalRep = createInternalRepresentation(
      yaml,
      callReconcileAndUpdateYaml
    );

    if (isDifferent) {
      $metricsInternalRep.regenerateInternalYAML(true);
    }

    if (errors) $metricsInternalRep.updateErrors(errors);
  }

  // reset internal representation in case of deviation from runtime YAML
  $: if (yaml) {
    updateInternalRep();
  }

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  $: modelName = $metricsInternalRep.getMetricKey("model");
  $: getModel = createRuntimeServiceGetCatalogEntry(instanceId, modelName);
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

  let validDimensionSelectorOption = [];
  $: if (model) {
    const selectedMetricsDefModelProfile = model?.schema?.fields ?? [];
    validDimensionSelectorOption = selectedMetricsDefModelProfile.map(
      (column) => ({ label: column.name, value: column.name })
    );
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

<WorkspaceContainer assetID={`${metricsDefName}-config`} inspector={false}>
  <MetricsWorkspaceHeader {metricsDefName} {metricsInternalRep} slot="header" />

  <div slot="body" use:listenToNodeResize>
    <div
      class="editor-pane bg-gray-100 p-6 flex flex-col"
      style:height="calc(100vh - var(--header-height))"
    >
      <ConfigParameters
        {metricsInternalRep}
        {metricsSourceSelectionError}
        {model}
        updateRuntime={callReconcileAndUpdateYaml}
      />

      <div
        class="flex-1"
        style="display: flex; flex-direction:column; overflow:hidden;"
      >
        <LayoutManager let:bottomResizeCallback let:topResizeCallback>
          <MetricsEntityTable
            addButtonId={"add-measure-button"}
            addLabel="Add measure"
            addEntityHandler={handleCreateMeasure}
            columnNames={MeasuresColumns}
            deleteEntityHandler={handleDeleteMeasure}
            label={"Measures"}
            resizeCallback={topResizeCallback}
            rows={measures ?? []}
            slot="top-item"
            tooltipText={"Add a new measure"}
            updateEntityHandler={handleUpdateMeasure}
          />

          <MetricsEntityTable
            addButtonId={"add-dimension-button"}
            addLabel="Add dimension"
            addEntityHandler={handleCreateDimension}
            columnNames={DimensionColumns}
            deleteEntityHandler={handleDeleteDimension}
            label={"Dimensions"}
            resizeCallback={bottomResizeCallback}
            rows={dimensions ?? []}
            slot="bottom-item"
            tooltipText={"Add a new dimension"}
            updateEntityHandler={handleUpdateDimension}
          />
        </LayoutManager>
      </div>
    </div>
  </div>
</WorkspaceContainer>
