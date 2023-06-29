<script lang="ts">
  import type { EditorView } from "@codemirror/basic-setup";
  import { debounceDocUpdateAnnotation } from "@rilldata/web-common/components/editor/annotations";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-common/layout/app-store";
  import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
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
  import { mapRuntimeErrorsToLines } from "./editor/errors";
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
  // update the intermediate yaml if the source yaml itself changes.
  $: intermediateYAML = yaml;

  const debounce = createDebouncer();

  /** update this configuration file.
   * To do so, we'll track whether or not the view update has a transaction with
   * a debounceDocUpdateAnnotation. If so, we will use this to update the actual debounce
   * time.
   */
  function updateMetrics(event) {
    const { content, viewUpdate } = event.detail;
    intermediateYAML = content;

    // check to see if this transaction has a debounce annotation.
    // This will be dispatched in change transactions with the debounceDocUpdateAnnotation
    // added to it.
    const debounceTransaction = viewUpdate.transactions.find(
      (transaction) =>
        transaction.annotation(debounceDocUpdateAnnotation) !== undefined
    );

    // get the annotation.
    const debounceAnnotation = debounceTransaction?.annotation(
      debounceDocUpdateAnnotation
    );
    // If there is no debounce annotation, we'll use the default debounce time.
    // Otherwise, we'll use the debounce based on the annotation.
    // This annotation comes from a CodeMirror editor update transaction.
    // Most likely, if debounceAnnotation is not undefined, it's because
    // the user took an action to explicitly update the editor contents
    // that didn't look like regular text editing (in this case,
    // probably Placeholder.svelte).
    //
    // We otherwise debounce to 300ms to prevent a lot of reconciliation thrashing.
    debounce(
      () => {
        callReconcileAndUpdateYaml(content);
      },
      debounceAnnotation !== undefined ? debounceAnnotation : 300
    );

    // immediately set the line statuses to be empty if the content is empty.
    if (!content?.length) {
      setLineStatuses([], false)(view);
    }
  }

  /** handle errors */

  $: path = Object.keys($fileArtifactsStore?.entities)?.find((key) => {
    return key.endsWith(`${metricsDefName}.yaml`);
  });

  $: runtimeErrors = $fileArtifactsStore?.entities?.[path]?.errors;
  $: lineBasedRuntimeErrors = mapRuntimeErrorsToLines(runtimeErrors, yaml);
  /** display the main error (the first in this array) at the bottom */
  $: mainError = [...lineBasedRuntimeErrors, ...(runtimeErrors || [])]?.at(0);

  let view: EditorView;

  /** if the errors change, run the following transaction. */
  $: if (view) setLineStatuses(lineBasedRuntimeErrors)(view);
</script>

<WorkspaceContainer inspector={true} assetID={`${metricsDefName}-config`}>
  <MetricsWorkspaceHeader
    slot="header"
    {metricsDefName}
    {yaml}
    error={mainError}
  />
  <MetricsEditor
    slot="body"
    bind:view
    {yaml}
    on:update={updateMetrics}
    {metricsDefName}
    error={mainError}
  />
  <MetricsInspector slot="inspector" yaml={intermediateYAML} />
</WorkspaceContainer>
