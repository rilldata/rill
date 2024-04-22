<script lang="ts">
  import { useModel } from "@rilldata/web-common/features/models/selectors";
  import WorkspaceInspector from "@rilldata/web-common/features/sources/inspector/WorkspaceInspector.svelte";
  import {
    V1Resource,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import {
    getModelOutOfPossiblyMalformedYAML,
    getTableOutOfPossiblyMalformedYAML,
  } from "../../utils";

  export let filePath: string;

  // Get model/table name from YAML
  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);
  $: yaml = $fileQuery.data?.blob || "";
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml)?.replace(/"/g, "");
  $: tableName = getTableOutOfPossiblyMalformedYAML(yaml)?.replace(/"/g, "");

  let modelQuery: CreateQueryResult<V1Resource>;
  $: if (modelName) {
    modelQuery = useModel($runtime.instanceId, modelName);
  }
</script>

{#if modelName && $modelQuery.data && $modelQuery.data?.model?.state?.connector}
  <WorkspaceInspector
    connector={$modelQuery.data?.model?.state?.connector}
    tableName={modelName}
    model={$modelQuery.data?.model}
    hasErrors={false}
    showReferences={false}
    sourceIsReconciling={false}
    hasUnsavedChanges={false}
    showSummaryTitle
  />
{:else}
  <div
    class="px-4 py-24 italic ui-copy-disabled text-center w-full"
    style:text-wrap="balance"
  >
    {#if !yaml?.length}
      <p>Let's get started.</p>
    {:else if modelName !== undefined}
      <div>
        <p>Model not defined.</p>
        <p>
          Set a model with <code>model: MODEL_NAME</code> to connect your metrics
          to a model.
        </p>
      </div>
    {:else if tableName !== undefined}
      <div>
        <p>Table not defined.</p>
        <p>
          Set a table with <code>table: TABLE_NAME</code> to connect your metrics
          to a table.
        </p>
      </div>
    {/if}
  </div>
{/if}
