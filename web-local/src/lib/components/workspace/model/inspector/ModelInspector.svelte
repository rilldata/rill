<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import StickToHeaderDivider from "@rilldata/web-local/lib/components/panel/StickToHeaderDivider.svelte";
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
    <StickToHeaderDivider />
    <ModelInspectorModelProfile {modelName} />
  </div>
{/key}
