<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-local/lib/components/actions/create-resize-listener-factory";
  import ModelInspectorHeader from "./header/ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
</script>

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
