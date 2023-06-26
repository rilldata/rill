<script lang="ts">
  import type { EditorView } from "@codemirror/basic-setup";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-common/layout/app-store";
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
  import { getSyntaxErrors, mapRuntimeErrorsToLines } from "./editor/errors";
  import MetricsInspector from "./inspector/MetricsInspector.svelte";

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

  /** keep track of the IMMEDIATE client-side YAML changes. */
  let intermediateYAML = yaml;

  function updateMetrics(event) {
    const { content } = event.detail;
    intermediateYAML = content;
    callReconcileAndUpdateYaml(content);
  }

  /** handle errors */

  $: path = Object.keys($fileArtifactsStore?.entities)?.find((key) => {
    return key.endsWith(`${metricsDefName}.yaml`);
  });

  $: runtimeErrors = $fileArtifactsStore?.entities?.[path]?.errors;
  $: lineBasedRuntimeErrors = mapRuntimeErrorsToLines(runtimeErrors, yaml);
  $: clientSideSyntaxErrors = getSyntaxErrors(yaml);
  $: lineErrors = [
    ...clientSideSyntaxErrors,
    ...(lineBasedRuntimeErrors || []),
  ];
  /** display the main error (the first in this array) at the bottom */
  $: mainError = [...lineErrors, ...(runtimeErrors || [])]?.at(0);

  let view: EditorView;

  /** if the errors change, let's run this transaction. */
  $: if (view) setLineStatuses(lineErrors)(view);
</script>

<WorkspaceContainer inspector={true} assetID={`${metricsDefName}-config`}>
  <MetricsWorkspaceHeader slot="header" {metricsDefName} {yaml} />
  <MetricsEditor
    slot="body"
    bind:view
    on:update={updateMetrics}
    {yaml}
    {metricsDefName}
    error={mainError}
  />
  <MetricsInspector slot="inspector" yaml={intermediateYAML} />
</WorkspaceContainer>
