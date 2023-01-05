<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore?.instanceId,
    modelName
  );

  $: getFile = useRuntimeServiceGetFile(
    $runtimeStore?.instanceId,
    `/models/${modelName}.sql`
  );

  // $: emptyModel = modelIsEmpty($runtimeStore?.instanceId, modelName);

  $: fileHasSQL = !$getModel?.isError && $getFile?.data?.blob?.length > 0;
</script>

{#if fileHasSQL === true}
  <div>
    {#key modelName + fileHasSQL}
      <div use:listenToNodeResize>
        <ModelInspectorHeader
          {modelName}
          containerWidth={$observedNode?.clientWidth}
        />
        <hr />
        <ModelInspectorModelProfile {modelName} />
      </div>
    {/key}
  </div>
{:else}
  <div class="px-4 py-2">Model is empty.</div>
{/if}
