<script lang="ts">
  import ModelInspectorModelProfile from "@rilldata/web-common/features/models/workspace/inspector/ModelInspectorModelProfile.svelte";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getModelOutOfPossiblyMalformedYAML } from "../../utils";
  import ConfigPreviews from "./ConfigPreviews.svelte";

  export let yaml: string;
  export let metricsDefName: string;

  // get file.
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml).replace(/"/g, "");

  // check to see if this model name exists.
  //$: modelExists = $fileArtifactsStore.has(modelName);
  $: modelQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );

  $: models = createRuntimeServiceListCatalogEntries($runtime.instanceId, {
    type: "OBJECT_TYPE_MODEL",
  });

  let isValidModel = false;
  $: if ($models?.data?.entries) {
    isValidModel = $models?.data.entries.some(
      (model) => model.name === modelName
    );
  }
</script>

{#if modelName && !$modelQuery?.isError && isValidModel}
  <ConfigPreviews {modelName} {metricsDefName} />
  <ModelInspectorModelProfile {modelName} />
{:else if modelName !== undefined}
  Model {modelName} not found.
{:else}
  Let's get started. add <code>model: MODEL_NAME</code> to connect a Model.
{/if}
