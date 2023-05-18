<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-common/layout/app-store";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import {
    V1PutFileAndReconcileResponse,
    createRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";
  import MetricsEditor from "./editor/MetricsEditor.svelte";
  import ConfigInspector from "./inspector/ConfigInspector.svelte";
  // the runtime yaml string
  export let yaml: string;
  export let metricsDefName: string;

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

  function updateYAML(event) {
    const { content } = event.detail;
    callReconcileAndUpdateYaml(content);
  }

  // let validDimensionSelectorOption = [];
  // $: if (model) {
  //   const selectedMetricsDefModelProfile = model?.schema?.fields ?? [];
  //   validDimensionSelectorOption = selectedMetricsDefModelProfile.map(
  //     (column) => ({ label: column.name, value: column.name })
  //   );
  // } else {
  //   validDimensionSelectorOption = [];
  // }

  // $: MeasuresColumns = initMeasuresColumns(
  //   handleUpdateMeasure,
  //   handleMeasureExpressionValidation
  // );
  // $: DimensionColumns = initDimensionColumns(
  //   handleUpdateDimension,
  //   validDimensionSelectorOption
  // );

  // let errors: Array<V1ReconcileError>;
  // $: errors =
  //   $fileArtifactsStore.entities[
  //     getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  //   ]?.errors;

  // $: metricsSourceSelectionError = nonStandardError
  //   ? nonStandardError
  //   : MetricsSourceSelectionError(errors);
</script>

<WorkspaceContainer inspector={true} assetID={`${metricsDefName}-config`}>
  <MetricsWorkspaceHeader slot="header" {metricsDefName} {yaml} />
  <div slot="body" use:listenToNodeResize>
    <div
      class="editor-pane bg-gray-100 p-6 grid  content-stretch"
      style:height="calc(100vh - var(--header-height))"
    >
      <MetricsEditor on:update={updateYAML} {yaml} {metricsDefName} />
      <!-- {#each [...mappedErrors, ...mappedSyntaxErrors] as error}
        <div>
          {JSON.stringify(error)}
        </div>
      {/each} -->
    </div>
  </div>
  <ConfigInspector slot="inspector" {metricsDefName} {yaml} />
</WorkspaceContainer>
