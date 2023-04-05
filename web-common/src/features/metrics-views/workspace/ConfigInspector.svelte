<script lang="ts">
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ModelInspectorModelProfile from "../../models/workspace/inspector/ModelInspectorModelProfile.svelte";
  import { getModelOutOfPossiblyMalformedYAML } from "../utils";

  export let yaml: string;

  // get file.
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml).replace(/"/g, "");

  // check to see if this model name exists.
  //$: modelExists = $fileArtifactsStore.has(modelName);
  $: modelQuery = useRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );

  $: models = useRuntimeServiceListCatalogEntries($runtime.instanceId, {
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
  <ModelInspectorModelProfile {modelName} />
{:else if modelName !== undefined}
  Model {modelName} not found.
{:else}
  Let's get started. add <code>model: MODEL_NAME</code> to connect a Model.
{/if}
