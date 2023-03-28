<script lang="ts">
  import { useRuntimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ModelInspector from "../../models/workspace/inspector/ModelInspector.svelte";
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
</script>

{#if modelName && !$modelQuery?.isError}
  <ModelInspector {modelName} />
{:else if modelName !== undefined}
  Model {modelName} not found.
{:else}
  Let's get started. add <code>model: MODEL_NAME</code> to connect a Model.
{/if}
