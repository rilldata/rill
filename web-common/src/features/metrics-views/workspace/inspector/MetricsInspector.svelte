<script lang="ts">
  import {
    useModel,
    useModels,
  } from "@rilldata/web-common/features/models/selectors";
  import {
    V1Resource,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    getModelOutOfPossiblyMalformedYAML,
    getTableOutOfPossiblyMalformedYAML,
  } from "../../utils";
  import WorkspaceInspector from "@rilldata/web-common/features/sources/inspector/WorkspaceInspector.svelte";

  export let filePath: string;

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, {
    path: filePath,
  });
  $: yaml = $fileQuery.data?.blob || "";

  // get file.
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml)?.replace(/"/g, "");
  $: tableName = getTableOutOfPossiblyMalformedYAML(yaml)?.replace(/"/g, "");

  // check to see if this model name exists.
  $: modelQuery = useModel($runtime.instanceId, modelName ?? "");

  $: allModels = useModels($runtime.instanceId);

  let isValidModel = false;
  $: if ($allModels?.data?.entries) {
    isValidModel = $allModels?.data.some(
      (model) => model?.meta?.name?.name === modelName,
    );
  }

  let entry: V1Resource;
  // refresh entry value only if the data has changed
  $: entry = $modelQuery?.data || entry;
</script>

{#if modelName && !$modelQuery?.isError && isValidModel && entry}
  <WorkspaceInspector
    hasErrors={false}
    showReferences={false}
    sourceIsReconciling={false}
    hasUnsavedChanges={false}
    showSummaryTitle
    model={entry?.model}
    tableName={modelName}
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
