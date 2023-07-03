<script lang="ts">
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";

  import type { EditorView } from "@codemirror/basic-setup";
  import EditorContainer from "@rilldata/web-common/components/editor/EditorContainer.svelte";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFileAndReconcile,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  import { debounceDocUpdateAnnotation } from "@rilldata/web-common/components/editor/annotations";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getMetricsDefErrors } from "../../utils";
  import {
    createPlaceholder,
    createPlaceholderElement,
  } from "./create-placeholder";
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

  /** note: this codemirror plugin does actually utilize tanstack query, and the
   * instantiation of the underlying svelte component that defines the placeholder
   * must be instantiated in the component.
   */
  const placeholderElement = createPlaceholderElement(metricsDefName);
  $: if (view) {
    placeholderElement.setEditorView(view);
  }

  const placeholder = createPlaceholder(placeholderElement.DOMElement);

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
    // We otherwise debounce to 200ms to prevent a lot of reconciliation thrashing.
    debounce(
      () => {
        callReconcileAndUpdateYaml(content);
      },
      immediateReconcileFromContentDeletion
        ? 0
        : debounceAnnotation !== undefined
        ? debounceAnnotation
        : 200
    );

    // immediately set the line statuses to be empty if the content is empty.
    if (!content?.length) {
      setLineStatuses([], view);
    }
  }

  $: runtimeErrors = getMetricsDefErrors($fileArtifactsStore, metricsDefName);

  $: lineBasedRuntimeErrors = mapRuntimeErrorsToLines(runtimeErrors, yaml);
  /** display the main error (the first in this array) at the bottom */
  $: mainError = [...lineBasedRuntimeErrors, ...(runtimeErrors || [])]?.at(0);
  let view: EditorView;

  /** If the errors change, run the following transaction.
   * Given that we are debouncing the core edit,
   */
  $: if (view) setLineStatuses(lineBasedRuntimeErrors, view);
</script>

<EditorContainer error={Boolean(yaml?.length) ? mainError : undefined}>
  <YAMLEditor
    bind:this={editor}
    content={yaml}
    bind:view
    on:update={updateMetrics}
    extensions={[placeholder]}
  />
</EditorContainer>
