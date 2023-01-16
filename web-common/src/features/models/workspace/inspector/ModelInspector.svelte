<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { modelIsEmpty } from "../../utils/model-is-empty";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: emptyModel = modelIsEmpty($runtimeStore?.instanceId, modelName);
</script>

{#if !$emptyModel?.data}
  <div>
    {#key modelName}
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
  <div class="px-4 py-24 italic ui-copy-disabled text-center">
    Model is empty.
  </div>
{/if}
