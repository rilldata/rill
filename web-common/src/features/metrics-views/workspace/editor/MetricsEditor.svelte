<script lang="ts">
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";

  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFileAndReconcile,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import MetricsEditorContainer from "./MetricsEditorContainer.svelte";

  import { skipDebounceAnnotation } from "@rilldata/web-common/components/editor/annotations";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    fileArtifactsStore,
    getFileArtifactReconciliationErrors,
  } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createPlaceholder } from "./create-placeholder";
  import { mapRuntimeErrorsToLines } from "./errors";

  export let metricsDefName: string;

  let editor: YAMLEditor;

  const queryClient = useQueryClient();
  const debounce = createDebouncer();

  // get the yaml blob from the file.
  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  );
  $: yaml = $fileQuery.data?.blob || "";

  const placeholderElements = createPlaceholder(metricsDefName);

  const placeholderElement = placeholderElements.component;
  $: if (view) {
    placeholderElement.setEditorView(view);
  }
  const placeholder = placeholderElements.extension;

  $: instanceId = $runtime.instanceId;

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

  /** update this configuration file.
   * To do so, we'll track whether or not the view update has a transaction with
   * a debounceDocUpdateAnnotation. If so, we will use this to update the actual debounce
   * time.
   */
  function updateMetrics(event) {
    const { content, viewUpdate } = event.detail;
    // immediately reconcile if the user deletes all the content.
    let immediateReconcileFromContentDeletion = !content?.length;

    // check to see if this transaction has a debounce annotation.
    // This will be dispatched in change transactions with the debounceDocUpdateAnnotation
    // added to it.
    const debounceTransaction = viewUpdate.transactions.find(
      (transaction) =>
        transaction.annotation(skipDebounceAnnotation) !== undefined
    );

    // get the annotation.
    const debounceAnnotation = debounceTransaction?.annotation(
      skipDebounceAnnotation
    );
    // We will skip the debounce if the user deletes all the content or there is a skipDebounceAnnotation.
    // See Placeholder.svelte for usage of this annotation.
    // We otherwise debounce to 200ms to prevent unneeded reconciliation thrashing.
    const debounceMS =
      immediateReconcileFromContentDeletion || debounceAnnotation ? 0 : 200;
    debounce(() => {
      callReconcileAndUpdateYaml(content);
    }, debounceMS);

    // immediately set the line statuses to be empty if the content is empty.
    if (!content?.length) {
      setLineStatuses([], view);
    }
  }

  $: runtimeErrors = getFileArtifactReconciliationErrors(
    $fileArtifactsStore,
    `${metricsDefName}.yaml`
  );

  $: lineBasedRuntimeErrors = mapRuntimeErrorsToLines(runtimeErrors, yaml);
  /** display the main error (the first in this array) at the bottom */
  $: mainError = [...lineBasedRuntimeErrors, ...(runtimeErrors || [])]?.at(0);
  let view: EditorView;

  /** If the errors change, run the following transaction.
   * Given that we are debouncing the core edit,
   */
  $: if (view) setLineStatuses(lineBasedRuntimeErrors, view);
</script>

<MetricsEditorContainer error={yaml?.length ? mainError : undefined}>
  <YAMLEditor
    bind:this={editor}
    content={yaml}
    bind:view
    on:update={updateMetrics}
    extensions={[placeholder]}
  />
</MetricsEditorContainer>
