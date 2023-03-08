<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { useModelFileIsEmpty } from "../../selectors";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: emptyModel = useModelFileIsEmpty($runtime?.instanceId, modelName);
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
